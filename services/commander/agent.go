package commander

import (
	"errors"
	app "twist-commander/app/interface"
	pb "twist-commander/pb"

	"github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
)

type Agent struct {
	app           app.AppImpl
	TransactionID string
	Subscriber    *nats.Subscription
	EventChannel  chan *pb.TransactionEvent
}

func CreateAgent(a app.AppImpl, transactionID string) *Agent {
	return &Agent{
		app:           a,
		TransactionID: transactionID,
		EventChannel:  make(chan *pb.TransactionEvent),
	}
}

func (agent *Agent) OpenEventChannel() error {

	// Listening to queue
	sb := agent.app.GetSignalBus()
	sub, err := sb.Watch("twist.transaction."+agent.TransactionID+".eventEmitted", func(msg *nats.Msg) {

		var event pb.TransactionEvent
		err := proto.Unmarshal(msg.Data, &event)
		if err != nil {
			return
		}

		defer func() {
			if recover() != nil {
				// Do nothing
			}
		}()

		agent.EventChannel <- &event
	})
	if err != nil {
		log.Error("did not connect: ", err)
		return errors.New("Failed to connect to signal server")
	}

	agent.Subscriber = sub

	return nil
}

func (agent *Agent) CloseEventChannel() {
	agent.Subscriber.Unsubscribe()
	close(agent.EventChannel)
}

func (agent *Agent) SendCommand(command string, payload *any.Any) error {

	// Preparing transaction command
	cmd := &pb.TransactionCommand{
		TransactionID: agent.TransactionID,
		Command:       command,
		Payload:       payload,
	}

	// Preparing command packet
	data, err := proto.Marshal(cmd)
	if err != nil {
		return errors.New("Failed to create command")
	}

	// Send command to queue
	sb := agent.app.GetSignalBus()
	err = sb.Emit("twist.transaction."+agent.TransactionID+".cmdReceived", data)
	if err != nil {
		return errors.New("Failed to send command")
	}

	return nil
}
