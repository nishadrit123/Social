package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type NC struct {
	nc *nats.Conn
	js jetstream.JetStream
}

func (nc *NC) SendToChat(subject string, bytePayload []byte) error {
	nc.js.CreateStream(context.Background(), jetstream.StreamConfig{
		Name:     USER_CHAT_STREAM,
		Subjects: []string{"chat.>"},
		Storage:  0,
	})

	err := nc.nc.Publish(subject, bytePayload)
	return err
}

func (nc *NC) GetallChats(subject string) ([]chatPayload, error) {
	var (
		payload      chatPayload
		payloadSlice []chatPayload
	)
	consumer_name := fmt.Sprintf("USER_CHAT_CONSUMER_%v", time.Now().Unix())
	consumer, err := nc.js.CreateConsumer(context.Background(), USER_CHAT_STREAM, jetstream.ConsumerConfig{
		Durable:       consumer_name,
		FilterSubject: subject,
	})
	if err != nil {
		log.Printf("err creating consumer %v", err)
		return payloadSlice, err
	}
	for {
		msg, err := consumer.Next(jetstream.FetchMaxWait(100 * time.Millisecond))
		if err != nil {
			break
		}
		err = json.Unmarshal(msg.Data(), &payload)
		if err != nil {
			log.Printf("Error unmarshaling chat payload, Err: %v", err)
			continue
		}
		payloadSlice = append(payloadSlice, payload)
	}
	nc.js.DeleteConsumer(context.Background(), USER_CHAT_STREAM, consumer_name)

	return payloadSlice, nil
}
