package store

import (
	"context"
	"database/sql"
)

type LikeStore struct {
	db *sql.DB
}

func (s *LikeStore) LikeDislike(ctx context.Context, postID int64, userID int64) (bool, error) {
	var (
		exists bool
		query  string
	)
	err := s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM liked WHERE user_id = $1 AND post_id = $2)", userID, postID).Scan(&exists)
	if err != nil {
		return exists, err
	}

	if !exists { // like the post
		query = `
			INSERT INTO liked (user_id, post_id)
			VALUES ($1, $2) 
		`
	} else { // dislike the post
		query = `
			DELETE FROM liked WHERE user_id = $1 AND post_id = $2 
		`
	}

	_, err = s.db.ExecContext(
		ctx,
		query,
		userID,
		postID,
	)
	return exists, err
}
