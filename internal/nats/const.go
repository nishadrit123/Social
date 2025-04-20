package nats

import "time"

const (
	USER_CHAT_STREAM       = "USER_CHAT_STREAM"
	GROUP_CHAT_STREAM      = "GROUP_CHAT_STREAM"
	USER_CHAT_CONSUMER     = "USER_CHAT_CONSUMER"
	GROUP_CHAT_CONSUMER    = "GROUP_CHAT_CONSUMER"
	USER_SUBJECT_WILDCARD  = "chat.user.>"
	GROUP_SUBJECT_WILDCARD = "chat.group.>"
)

type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
}

type Post struct {
	ID           int64    `json:"id"`
	Content      string   `json:"content"`
	Title        string   `json:"title"`
	UserID       int64    `json:"user_id"`
	User         User     `json:"user"`
	Tags         []string `json:"tags"`
	CreatedAt    string   `json:"created_at"`
	UpdatedAt    string   `json:"updated_at"`
	LikeCount    any      `json:"like_count"`
	CommentCount any      `json:"comment_count"`
	IsPostSaved  bool     `json:"is_post_saved,omitempty"`
	IsPostLiked  bool     `json:"is_post_liked,omitempty"`
}

type chatPayload struct {
	SenderID   int64     `json:"sender_id,omitempty"`
	ReceiverID int64     `json:"receiver_id,omitempty"` // can be both userId or groupID
	Text       string    `json:"text,omitempty"`
	PostID     int64     `json:"post_id,omitempty"`
	ChatPost   *Post     `json:"post,omitempty"`
	Date       time.Time `json:"date,omitempty"`
}
