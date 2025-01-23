package store

import (
	"context"
	"database/sql"
)

type UserStore struct {
	db *sql.DB
}

type User struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"-"`
	CreatedAt string `json:"created_at"`
}

func (s *UserStore) Create(ctx context.Context, tx *sql.Tx, user *User) error {
	query := `
		INSERT INTO users (username, password, email, role_id) VALUES 
	($1, $2, $3, (SELECT id FROM roles WHERE name = $4))
	RETURNING id, created_at
	`

	err := tx.QueryRowContext(
		ctx,
		query,
		user.Username,
		user.Email,
	).Scan(
		&user.ID,
		&user.CreatedAt,
	)
	return err
}
