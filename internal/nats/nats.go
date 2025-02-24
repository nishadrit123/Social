package nats

import (
	"github.com/nats-io/nats.go"
)

type Nats struct {
	NatsConn interface {
		Publish(string) error
	}
}

func NewNatsClient(addr string) (*nats.Conn, error) {
	nc, err := nats.Connect(addr)
	return nc, err
}

func NewNatsConnection(nc *nats.Conn) Nats {
	return Nats{
		NatsConn: &NC{nc: nc},
	}
}
