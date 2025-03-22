package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type NC struct {
	nc *nats.Conn
	js jetstream.JetStream
}

func (nc *NC) SendToChat(subject string, bytePayload []byte, is_user bool) error {
	var (
		stream_name      string
		subject_wildcard string
	)
	if is_user {
		stream_name = USER_CHAT_STREAM
		subject_wildcard = USER_SUBJECT_WILDCARD
	} else {
		stream_name = GROUP_CHAT_STREAM
		subject_wildcard = GROUP_SUBJECT_WILDCARD
	}

	nc.js.CreateStream(context.Background(), jetstream.StreamConfig{
		Name:     stream_name,
		Subjects: []string{subject_wildcard},
		Storage:  0,
	})

	err := nc.nc.Publish(subject, bytePayload)
	return err
}

func (nc *NC) GetallChats(subject string, is_user bool) ([]chatPayload, error) {
	var (
		payload       chatPayload
		payloadSlice  []chatPayload
		consumer_name string
		stream        string
	)
	if is_user {
		consumer_name = fmt.Sprintf("%v_%v", USER_CHAT_CONSUMER, time.Now().Unix())
		stream = USER_CHAT_STREAM
	} else {
		consumer_name = fmt.Sprintf("%v_%v", GROUP_CHAT_CONSUMER, time.Now().Unix())
		stream = GROUP_CHAT_STREAM
	}
	consumer, err := nc.js.CreateConsumer(context.Background(), stream, jetstream.ConsumerConfig{
		Durable:       consumer_name,
		FilterSubject: subject,
	})
	if err != nil {
		log.Printf("err creating consumer %v", err)
		return payloadSlice, err
	}
	for {
		payload = chatPayload{}
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
	nc.js.DeleteConsumer(context.Background(), stream, consumer_name)

	return payloadSlice, nil
}

func (nc *NC) OpenChats(subjects []string, userID int64) [][]int64 {
	var (
		allChatWindows [][]int64 // [[usr]]
		tempUserWindow []int64
	)
	for _, subject := range subjects {
		if strings.Contains(subject, "user") {
			consumer_name := fmt.Sprintf("%v_%v", USER_CHAT_CONSUMER, time.Now().Unix())
			stream := USER_CHAT_STREAM
			users, err := nc.GetChatWindow(subject, consumer_name, stream, userID)
			if err != nil {
				log.Printf("Error fetching users that you have chatted with, Err: %v\n", err)
			}
			tempUserWindow = append(tempUserWindow, users...)
		}
	}
	allChatWindows = append(allChatWindows, removeDuplicates(tempUserWindow))
	return allChatWindows
}

func (nc *NC) GetChatWindow(subject, consumer_name, stream string, userID int64) ([]int64, error) {
	var chatWindowSlice []int64
	consumer, err := nc.js.CreateConsumer(context.Background(), stream, jetstream.ConsumerConfig{
		Durable:       consumer_name,
		FilterSubject: subject,
	})
	if err != nil {
		log.Printf("err creating consumer %v", err)
		return chatWindowSlice, err
	}
	for {
		msg, err := consumer.Next(jetstream.FetchMaxWait(100 * time.Millisecond))
		if err != nil {
			break
		}
		sub := msg.Subject()
		subSplit := strings.Split(sub, ".")
		if subSplit[1] == "user" {
			userIDStr := strconv.Itoa(int(userID))
			if subSplit[2] == userIDStr {
				user, _ := strconv.Atoi(subSplit[3])
				chatWindowSlice = append(chatWindowSlice, int64(user))
			} else if subSplit[3] == userIDStr {
				user, _ := strconv.Atoi(subSplit[2])
				chatWindowSlice = append(chatWindowSlice, int64(user))
			}
		} else if subSplit[1] == "group" {
			group, _ := strconv.Atoi(subSplit[3])
			chatWindowSlice = append(chatWindowSlice, int64(group))
		}
	}
	nc.js.DeleteConsumer(context.Background(), stream, consumer_name)
	return chatWindowSlice, nil
}

func removeDuplicates(intSlice []int64) []int64 {
	allKeys := make(map[int64]bool)
	list := []int64{}
	for _, item := range intSlice {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}
