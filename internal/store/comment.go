package store

import (
	"context"
	"database/sql"
	"errors"
)

type CommentStore struct {
	db *sql.DB
}

type Comment struct {
	ID        int64  `json:"id"`
	Comment   string `json:"comment"`
	UserID    int64  `json:"user_id"`
	PostID    int64  `json:"post_id"`
	CreatedAt string `json:"created_at"`
}

type PostComment struct {
	Comment string `json:"comment"`
	USerID  int64  `json:"user_id"`
}

func (s *CommentStore) Create(ctx context.Context, comment *Comment) error {
	query := `
		INSERT INTO comment (comment, user_id, post_id)
		VALUES ($1, $2, $3) RETURNING id, created_at
	`
	err := s.db.QueryRowContext(
		ctx,
		query,
		comment.Comment,
		comment.UserID,
		comment.PostID,
	).Scan(
		&comment.ID,
		&comment.CreatedAt,
	)
	return err
}

func (s *CommentStore) GetByID(ctx context.Context, id int64) (*Comment, error) {
	query := `
		SELECT id, comment, user_id, post_id
		FROM comment
		WHERE id = $1
	`
	var comment Comment
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&comment.ID,
		&comment.Comment,
		&comment.UserID,
		&comment.PostID,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, errors.New("ErrNotFound")
		default:
			return nil, err
		}
	}

	return &comment, nil
}

func (s *CommentStore) GetByPostID(ctx context.Context, id int64) ([]PostComment, error) {
	query_comment := `
		SELECT comment, user_id
		FROM comment
		WHERE post_id = $1
	`
	var commentslice []PostComment
	rows, err := s.db.QueryContext(ctx, query_comment, id)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, errors.New("ErrNotFound")
		default:
			return nil, err
		}
	}
	for rows.Next() {
		var comment PostComment
		err = rows.Scan(
			&comment.Comment,
			&comment.USerID,
		)
		if err != nil {
			return nil, err
		}
		commentslice = append(commentslice, comment)
	}

	return commentslice, nil
}

func (s *CommentStore) Update(ctx context.Context, commentid int64, comment string) error {
	query := `
		UPDATE comment
		SET comment = $1
		WHERE id = $2
	`
	_, err := s.db.ExecContext(ctx, query, comment, commentid)
	if err != nil {
		return err
	}
	return nil
}

func (s *CommentStore) Delete(ctx context.Context, commentID int64) error {
	query := `DELETE FROM comment WHERE id = $1`

	res, err := s.db.ExecContext(ctx, query, commentID)
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

func (s *CommentStore) DeleteByPostID(ctx context.Context, PostID int64) error {
	query := `DELETE FROM comment WHERE post_id = $1`

	res, err := s.db.ExecContext(ctx, query, PostID)
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
