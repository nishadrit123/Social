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

func (s *GroupStore) IsUserInGroup(ctx context.Context, groupID, userID int64) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1 FROM groups 
			WHERE id = $1 
			AND ($2 = created_by OR $2 = ANY(members))
		);
	`
	var exists bool
	err := s.db.QueryRowContext(ctx, query, groupID, userID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (s *GroupStore) GetGroupInfo(ctx context.Context, groupID int64) (*Group, error) {
	query := `
		SELECT name, created_by, members FROM groups WHERE id = $1;
	`

	var group Group

	err := s.db.QueryRowContext(ctx, query, groupID).Scan(
		&group.Name,
		&group.CreatedBy,
		pq.Array(&group.Members),
	)
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (s *GroupStore) GetGroupsForUser(ctx context.Context, userID int64) ([]int64, error) {
	query := `
		SELECT id FROM groups 
		WHERE ($1 = created_by OR $1 = ANY(members));
	`

	var groupIDs []int64
	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var groupID int64
		if err := rows.Scan(&groupID); err != nil {
			return nil, err
		}
		groupIDs = append(groupIDs, groupID)
	}
	return groupIDs, nil
}
