package signalbus

import (
	nats "github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
)

type SignalBus struct {
	host       string
	clientName string
	client     *nats.Conn
}

func CreateConnector(host string, clientName string) *SignalBus {
	return &SignalBus{
		host:       host,
		clientName: clientName,
	}
}

func (sb *SignalBus) Connect() error {

	log.WithFields(log.Fields{
		"host":       sb.host,
		"clientName": sb.clientName,
	}).Info("Connecting to signal server")

	// Connect to signal server
	nc, err := nats.Connect(sb.host, nats.Name(sb.clientName))
	if err != nil {
		return err
	}

	sb.client = nc

	nc.SetReconnectHandler(func(rcb *nats.Conn) {
		log.Info("Reconnecting to signal server ...")
	})

	return nil
}

func (sb *SignalBus) Close() {
	sb.client.Close()
}

func (sb *SignalBus) Emit(topic string, data []byte) error {

	if err := sb.client.Publish(topic, data); err != nil {
		return err
	}

	return nil
}

func (sb *SignalBus) Watch(topic string, fn func(*nats.Msg)) (*nats.Subscription, error) {

	// Subscribe
	sub, err := sb.client.Subscribe(topic, fn)
	if err != nil {
		return nil, err
	}

	// Add to subscription list

	return sub, nil
}
