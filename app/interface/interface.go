package app

import "github.com/nats-io/nats.go"

type SignalBusImpl interface {
	Emit(string, []byte) error
	Watch(string, func(*nats.Msg)) (*nats.Subscription, error)
}

type AppImpl interface {
	GetSignalBus() SignalBusImpl
}
