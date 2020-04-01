package app

import (
	log "github.com/sirupsen/logrus"
	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"

	app "twist-commander/app/interface"
	pb "twist-commander/pb"
	commander "twist-commander/services/commander"
)

type GRPCServer struct {
	Commander *commander.Service
}

func (a *App) InitGRPCServer(host string) error {
	/*
		// Start to listen on port
		lis, err := net.Listen("tcp", host)
		if err != nil {
			log.Fatal(err)
			return err
		}
	*/
	lis := a.connectionListener.MatchWithWriters(
		cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"),
	)

	log.WithFields(log.Fields{
		"host": host,
	}).Info("Starting gRPC server on " + host)

	// Create gRPC server
	s := grpc.NewServer()

	// Register data source adapter service
	commanderService := commander.CreateService(app.AppImpl(a))
	a.grpcServer.Commander = commanderService
	pb.RegisterCommanderServer(s, commanderService)
	//reflection.Register(s)

	log.WithFields(log.Fields{
		"service": "Commander",
	}).Info("Registered service")

	// Starting server
	if err := s.Serve(lis); err != cmux.ErrListenerClosed {
		log.Fatal(err)
		return err
	}

	return nil
}
