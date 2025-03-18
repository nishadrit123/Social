package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrNotFound          = errors.New("resource not found")
	ErrConflict          = errors.New("resource already exists")
	QueryTimeoutDuration = time.Second * 5
)

type Storage struct {
	Posts interface {
		Create(context.Context, *Post) error
		GetByID(context.Context, int64) (*Post, error)
		Delete(context.Context, int64) error
		Update(context.Context, *Post) error
		GetUserFeed(context.Context, int64, PaginatedFeedQuery) ([]PostWithMetadata, error)
	}
	Users interface {
		Create(context.Context, *sql.Tx, *User) error
		GetByID(context.Context, int64) (*User, error)
		GetPostsByUser(context.Context, int64) ([]Post, error)
		GetByEmail(context.Context, string) (*User, error)
		CreateAndInvite(ctx context.Context, user *User, token string, exp time.Duration) error
		Activate(context.Context, string) error
		Delete(context.Context, int64) error
		SaveUnsavePost(context.Context, int64, int64) (bool, error)
		GetSavedPostsByUser(context.Context, int64) ([]int64, error)
	}
	Comment interface {
		Create(context.Context, *Comment) error
		GetByID(context.Context, int64) (*Comment, error)
		GetByPostID(context.Context, int64) ([]PostComment, error)
		Update(context.Context, int64, string) error
		Delete(context.Context, int64) error
		DeleteByPostID(context.Context, int64) error
	}
	Like interface {
		LikeDislike(context.Context, int64, int64) (bool, error)
		GetUsersLike(context.Context, int64) ([]string, error)
		DeleteByPostID(context.Context, int64) error
	}
	Followers interface {
		Follow(context.Context, int64, int64) error
		Unfollow(context.Context, int64, int64) error
	}
	Roles interface {
		GetByName(context.Context, string) (*Role, error)
	}
	Group interface {
		Create(context.Context, *Group) error
		AddMembers(context.Context, int64, *Group) error
	}
}

func NewStorage(db *sql.DB) Storage {
	return Storage{
		Posts:     &PostStore{db},
		Users:     &UserStore{db},
		Comment:   &CommentStore{db},
		Like:      &LikeStore{db},
		Followers: &FollowerStore{db},
		Roles:     &RoleStore{db},
		Group:     &GroupStore{db},
	}
}

func withTx(db *sql.DB, ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
