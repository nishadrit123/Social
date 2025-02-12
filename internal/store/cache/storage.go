package cache

import (
	"context"

	"github.com/go-redis/redis/v8"
)

type Storage struct {
	Users interface {
		Get(context.Context, int64, string) (any, error)
		Set(context.Context, any, int64, string) error
		UnSet(context.Context, int64, string) error
		Delete(context.Context, int64)
	}
}

func NewRedisStorage(rbd *redis.Client) Storage {
	return Storage{
		Users: &UserStore{rdb: rbd},
	}
}
