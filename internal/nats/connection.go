package nats

import (
	"github.com/nats-io/nats.go"
)

type NC struct {
	nc *nats.Conn
}

func (nc *NC) Publish(x string) error {
	return nil
}
