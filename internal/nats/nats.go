package nats

import (
	"log"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type Nats struct {
	NatsConn interface {
		SendToChat(string, []byte) error
		GetallChats(string) ([]chatPayload, error)
	}
}

func NewNatsClient(addr string) (*nats.Conn, error) {
	nc, err := nats.Connect(addr)
	return nc, err
}

func NewNatsConnection(nc *nats.Conn) Nats {
	js, err := jetstream.New(nc)
	if err != nil {
		log.Fatalf("Err creating JS %v", err)
	}
	return Nats{
		NatsConn: &NC{
			nc: nc,
			js: js,
		},
	}
}
