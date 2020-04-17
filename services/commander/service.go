package commander

import (
	"sync"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/nats-io/nats.go"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/viper"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	app "twist-commander/app/interface"
	pb "twist-commander/pb"
)

type Service struct {
	app       app.AppImpl
	commander *Commander
}

func CreateService(a app.AppImpl) *Service {

	// Preparing service
	service := &Service{
		app:       a,
		commander: CreateCommander(a),
	}

	return service
}

func (service *Service) CreateTransaction(ctx context.Context, in *pb.CreateTransactionRequest) (*pb.CreateTransactionReply, error) {

	mode := in.Mode
	if in.Mode == "" {
		mode = "sync"
	}

	tid := uuid.NewV1()
	transactionID := tid.String()

	// Waiting for response from queue
	wg := sync.WaitGroup{}
	wg.Add(1)

	// Listening to queue
	sb := service.app.GetSignalBus()
	sub, err := sb.Watch("twist.transaction."+transactionID+".eventEmitted", func(msg *nats.Msg) {
		var event pb.TransactionEvent
		err := proto.Unmarshal(msg.Data, &event)
		if err != nil {
			return
		}

		// Got message that transaction was assigned to runner already
		if event.EventName == "Assigned" && event.TransactionID == transactionID {
			wg.Done()
			return
		}
	})
	if err != nil {
		log.Error("did not connect: ", err)
		return &pb.CreateTransactionReply{
			Success: false,
		}, nil
	}
	defer sub.Unsubscribe()

	// Set up a connection to supervisor.
	address := viper.GetString("supervisor.host")
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Error("did not connect: ", err)
		return &pb.CreateTransactionReply{
			Success: false,
		}, nil
	}
	defer conn.Close()

	// Preparing context
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	req := &pb.PrepareTransactionRequest{
		TransactionID: transactionID,
		Mode:          mode,
	}

	// Prepare transaction
	res, err := pb.NewSupervisorClient(conn).PrepareTransaction(ctx, req)
	if err != nil {
		log.Error(err)
		return &pb.CreateTransactionReply{
			Success: false,
		}, nil
	}

	if res.Success == false {
		return &pb.CreateTransactionReply{
			Success: false,
		}, nil
	}

	log.WithFields(log.Fields{
		"mode": mode,
	}).Info("Created transation: ", res.TransactionID)

	// Wait transaction event that is ready
	wg.Wait()

	log.WithFields(log.Fields{
		"transaction": res.TransactionID,
	}).Info("Transaction is ready")

	return &pb.CreateTransactionReply{
		Success:       true,
		TransactionID: res.TransactionID,
	}, nil
}

func (service *Service) ConfirmTransaction(ctx context.Context, in *pb.ConfirmTransactionRequest) (*pb.ConfirmTransactionReply, error) {

	err := service.commander.ConfirmTransaction(in.TransactionID, in)
	if err != nil {
		return &pb.ConfirmTransactionReply{
			Success:       false,
			TransactionID: in.TransactionID,
		}, nil
	}

	return &pb.ConfirmTransactionReply{
		Success:       true,
		TransactionID: in.TransactionID,
	}, nil
}

func (service *Service) RegisterTasks(ctx context.Context, in *pb.RegisterTasksRequest) (*pb.RegisterTasksReply, error) {

	err := service.commander.RegisterTasks(in.TransactionID, in)
	if err != nil {
		return &pb.RegisterTasksReply{
			Success:       false,
			TransactionID: in.TransactionID,
		}, nil
	}

	return &pb.RegisterTasksReply{
		Success:       true,
		TransactionID: in.TransactionID,
	}, nil
}

func (service *Service) CancelTransaction(ctx context.Context, in *pb.CancelTransactionRequest) (*pb.CancelTransactionReply, error) {

	err := service.commander.CancelTransaction(in.TransactionID, in)
	if err != nil {
		return &pb.CancelTransactionReply{
			Success:       false,
			TransactionID: in.TransactionID,
		}, nil
	}

	return &pb.CancelTransactionReply{
		Success:       true,
		TransactionID: in.TransactionID,
	}, nil
}
