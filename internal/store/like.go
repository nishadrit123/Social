package store

import (
	"context"
	"database/sql"
	"errors"
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

func (s *LikeStore) GetUsersLike(ctx context.Context, postID int64) ([]string, error) {
	query := `
		SELECT username
		FROM users
		JOIN liked 
		ON users.id = liked.user_id
		WHERE post_id = $1
	`
	var usernameslice []string
	rows, err := s.db.QueryContext(ctx, query, postID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, errors.New("ErrNotFound")
		default:
			return nil, err
		}
	}
	for rows.Next() {
		var username string
		err = rows.Scan(
			&username,
		)
		if err != nil {
			return nil, err
		}
		usernameslice = append(usernameslice, username)
	}

	return usernameslice, nil
}

func (s *LikeStore) DeleteByPostID(ctx context.Context, postID int64) error {
	query := `DELETE FROM liked WHERE post_id = $1`

	res, err := s.db.ExecContext(ctx, query, postID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}
