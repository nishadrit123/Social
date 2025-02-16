package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"social/internal/store"

	"github.com/go-redis/redis/v8"
)

type UserStore struct {
	rdb *redis.Client
}

func GetRedisKey(key any, keyType string) string {
	var cacheKey string
	if keyType == "user" {
		cacheKey = fmt.Sprintf("user-%d", key)
	} else if keyType == "comment" {
		cacheKey = fmt.Sprintf("post-comment-%d", key)
	} else if keyType == "like" {
		cacheKey = fmt.Sprintf("post-like-%d", key)
	} else if keyType == "following" {
		cacheKey = fmt.Sprintf("user-%d-following", key)
	} else if keyType == "follower" {
		cacheKey = fmt.Sprintf("user-%d-follower", key)
	} else if keyType == "login" {
		cacheKey = fmt.Sprintf("user-%d-login", key)
	} else if keyType == "posts" {
		cacheKey = fmt.Sprintf("user-%d-posts", key)
	} else if keyType == "story" {
		cacheKey = fmt.Sprintf("user-%d-story-*", key)
	}
	return cacheKey
}

func (s *UserStore) Get(ctx context.Context, key int64, arrayVal string, keyType string) (any, error) {
	var (
		data    any
		stories []Story
		story   Story
	)
	cacheKey := GetRedisKey(key, keyType)
	if keyType == "login" {
		return s.rdb.SIsMember(ctx, cacheKey, arrayVal).Val(), nil
	} else if keyType == "story" {
		keys := s.rdb.Keys(ctx, cacheKey).Val()
		for _, key := range keys {
			story_str := s.rdb.Get(ctx, key).Val()
			if err := json.Unmarshal([]byte(story_str), &story); err != nil {
				return nil, err
			}
			stories = append(stories, story)
		}
		return stories, nil
	} else {
		data, _ = s.rdb.Get(ctx, cacheKey).Result()
		if keyType == "user" {
			var user store.User
			if data != "" {
				err := json.Unmarshal([]byte(data.(string)), &user)
				if err != nil {
					return nil, err
				}
			}
			return &user, nil
		}
		return data, nil
	}
}

func (s *UserStore) Set(ctx context.Context, value any, key int64, keyType string) error {
	cacheKey := GetRedisKey(key, keyType)
	if keyType == "login" {
		s.rdb.SAdd(ctx, cacheKey, value)
		return s.rdb.Expire(ctx, cacheKey, time.Hour*24).Err() // 1 day expiration same as jwt expiration
	} else if keyType == "user" {
		user := value.(*store.User)
		json, err := json.Marshal(user)
		if err != nil {
			return err
		}
		return s.rdb.SetEX(ctx, cacheKey, json, UserExpTime).Err()
	} else if keyType == "story" {
		uniqueTimeStamp := time.Now().Unix()
		cacheKey := fmt.Sprintf("user-%d-story-%d", key, uniqueTimeStamp)
		json, err := json.Marshal(value)
		if err != nil {
			return err
		}
		return s.rdb.SetEX(ctx, cacheKey, json, StoryExpTime).Err()
	}
	return s.rdb.Incr(ctx, cacheKey).Err()
}

func (s *UserStore) UnSet(ctx context.Context, key int64, arrayVal, keyType string) error {
	cacheKey := GetRedisKey(key, keyType)
	if keyType == "login" {
		return s.rdb.SRem(ctx, cacheKey, arrayVal).Err()
	}
	return s.rdb.Decr(ctx, cacheKey).Err()
}

func (s *UserStore) Delete(ctx context.Context, key any, keyType string) {
	var cacheKey string
	cacheKey = GetRedisKey(key, keyType)
	if keyType == "story" {
		cacheKey = key.(string)
	}
	s.rdb.Del(ctx, cacheKey)
}
