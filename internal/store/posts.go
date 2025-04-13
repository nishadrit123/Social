package store

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"
)

type PostStore struct {
	db *sql.DB
}

type Post struct {
	ID           int64    `json:"id"`
	Content      string   `json:"content"`
	Title        string   `json:"title"`
	UserID       int64    `json:"user_id"`
	Tags         []string `json:"tags"`
	CreatedAt    string   `json:"created_at"`
	UpdatedAt    string   `json:"updated_at"`
	User         User     `json:"user"`
	LikeCount    any      `json:"like_count"`
	CommentCount any      `json:"comment_count"`
	IsPostSaved  bool     `json:"is_post_saved,omitempty"`
	IsPostLiked  bool     `json:"is_post_liked,omitempty"`
}

type PostWithMetadata struct {
	Post
	CommentsCount int `json:"comments_count"`
}

func (s *PostStore) Create(ctx context.Context, post *Post) error {
	query := `
		INSERT INTO posts (content, title, user_id, tags)
		VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at
	`
	err := s.db.QueryRowContext(
		ctx,
		query,
		post.Content,
		post.Title,
		post.UserID,
		pq.Array(post.Tags),
	).Scan(
		&post.ID,
		&post.CreatedAt,
		&post.UpdatedAt,
	)
	return err
}

func (s *PostStore) GetByID(ctx context.Context, postID, userID int64) (*Post, error) {
	query := `
		SELECT p.id, p.user_id, u.username, p.title, p.content, p.created_at,  p.updated_at, p.tags,
		EXISTS (
        	SELECT 1 FROM liked l WHERE l.post_id = p.id AND l.user_id = $2
    	) as is_liked,
    	EXISTS (
        	SELECT 1 FROM savedpost s WHERE s.savedpost_id = p.id AND s.user_id = $2
    	) as is_saved
		FROM posts as p LEFT JOIN users AS u
		ON p.user_id = u.id
		WHERE p.id = $1
	`

	// ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	// defer cancel()
	var post Post
	err := s.db.QueryRowContext(ctx, query, postID, userID).Scan(
		&post.ID,
		&post.UserID,
		&post.User.Username,
		&post.Title,
		&post.Content,
		&post.CreatedAt,
		&post.UpdatedAt,
		pq.Array(&post.Tags),
		&post.IsPostLiked,
		&post.IsPostSaved,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, errors.New("ErrNotFound")
		default:
			return nil, err
		}
	}

	return &post, nil
}

func (s *PostStore) Delete(ctx context.Context, postID int64) error {
	query := `DELETE FROM posts WHERE id = $1`

	// ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	// defer cancel()

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

func (s *PostStore) Update(ctx context.Context, post *Post) error {
	query := `
		UPDATE posts
		SET title = $1, content = $2
		WHERE id = $3
	`

	// ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	// defer cancel()

	// err := s.db.QueryRowContext(
	// 	ctx,
	// 	query,
	// 	post.Title,
	// 	post.Content,
	// 	post.ID,
	// ).Scan(&post.Version)
	// if err != nil {
	// 	switch {
	// 	case errors.Is(err, sql.ErrNoRows):
	// 		return ErrNotFound
	// 	default:
	// 		return err
	// 	}
	// }

	_, err := s.db.ExecContext(ctx, query, post.Title, post.Content, post.ID)
	if err != nil {
		return err
	}

	return nil
}

func (s *PostStore) GetUserFeed(ctx context.Context, fq PaginatedFeedQuery, userID int64) ([]PostWithMetadata, error) {
	query := `
		SELECT p.id, p.user_id, p.title, p.content, p.created_at, p.tags, u.username,
		EXISTS (
        	SELECT 1 FROM liked l WHERE l.post_id = p.id AND l.user_id = $1
    	) as is_liked,
    	EXISTS (
        	SELECT 1 FROM savedpost s WHERE s.savedpost_id = p.id AND s.user_id = $1
    	) as is_saved
		FROM posts as p LEFT JOIN users as u on p.user_id = u.id
	`
	// query := `
	// 	SELECT id, user_id, title, content, created_at, tags
	// 	FROM posts as p
	// 	ORDER BY p.created_at ` + fq.Sort + `
	// 	LIMIT $1 OFFSET $2
	// `

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	// rows, err := s.db.QueryContext(ctx, query, fq.Limit, fq.Offset, fq.Search, pq.Array(fq.Tags))
	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var feed []PostWithMetadata
	for rows.Next() {
		var p PostWithMetadata
		err := rows.Scan(
			&p.ID,
			&p.UserID,
			&p.Title,
			&p.Content,
			&p.CreatedAt,
			pq.Array(&p.Tags),
			&p.User.Username,
			&p.IsPostLiked,
			&p.IsPostSaved,
		)
		if err != nil {
			return nil, err
		}

		feed = append(feed, p)
	}

	return feed, nil
}
