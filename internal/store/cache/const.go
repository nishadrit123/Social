package cache

import "time"

const (
	UserExpTime  = time.Minute * 5
	StoryExpTime = time.Hour * 24
)

type Story struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}
