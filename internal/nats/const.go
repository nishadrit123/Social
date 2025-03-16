package nats

import "time"

const (
	USER_CHAT_STREAM = "USER_CHAT_STREAM"
)

type chatPayload struct {
	SenderID   int64     `json:"sender_id,omitempty"`
	ReceiverID int64     `json:"receiver_id,omitempty"` // can be both userId or groupID
	Text       string    `json:"text,omitempty"`
	PostID     int64     `json:"post_id,omitempty"`
	Date       time.Time `json:"date,omitempty"`
}
