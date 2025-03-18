package store

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
)

type GroupStore struct {
	db *sql.DB
}

type Group struct {
	ID        int64   `json:"id"`
	Name      string  `json:"name"`
	CreatedBy int64   `json:"created_by"`
	Members   []int64 `json:"members"`
	CreatedAt string  `json:"created_at"`
}

func (s *GroupStore) Create(ctx context.Context, grp *Group) error {
	query := `
		INSERT INTO groups (name, created_by, members)
		VALUES ($1, $2, $3) RETURNING id, created_at
	`
	err := s.db.QueryRowContext(
		ctx,
		query,
		grp.Name,
		grp.CreatedBy,
		pq.Array(grp.Members),
	).Scan(
		&grp.ID,
		&grp.CreatedAt,
	)
	return err
}

func (s *GroupStore) AddMembers(ctx context.Context, grpID int64, grp *Group) error {
	query := `
		UPDATE groups 
		SET members = members || $2
		WHERE id = $1;
	`
	_, err := s.db.ExecContext(ctx, query, grpID, pq.Array(grp.Members))
	return err
}
