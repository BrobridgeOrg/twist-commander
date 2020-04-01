package app

import (
	"net"

	log "github.com/sirupsen/logrus"
	"github.com/soheilhy/cmux"
)

func (a *App) CreateConnectionListener(host string) error {

	// Start to listen on port
	lis, err := net.Listen("tcp", host)
	if err != nil {
		log.Fatal(err)
		return err
	}

	log.WithFields(log.Fields{
		"host": host,
	}).Info("Starting server on " + host)

	m := cmux.New(lis)

	a.connectionListener = m

	return nil
}

func (a *App) Serve() error {
	return a.connectionListener.Serve()
}
