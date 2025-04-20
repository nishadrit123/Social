package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"

	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDuplicateEmail    = errors.New("a user with that email already exists")
	ErrDuplicateUsername = errors.New("a user with that username already exists")
)

type UserStore struct {
	db *sql.DB
}

type User struct {
	ID                 int64    `json:"id"`
	Username           string   `json:"username"`
	Email              string   `json:"email"`
	Password           password `json:"-"`
	CreatedAt          string   `json:"created_at"`
	IsActive           bool     `json:"is_active"`
	RoleID             int64    `json:"role_id"`
	Role               Role     `json:"role"`
	PostCount          any      `json:"post_count"`
	FollowerCount      any      `json:"follower_count"`
	FollowingCount     any      `json:"following_count"`
	IsAlreadyFollowing bool     `json:"is_already_following"`
}

type compactUserPayload struct {
	Id   int64  `json:"userid"`
	Name string `json:"username"`
}

type password struct {
	text *string
	hash []byte
}

func (p *password) Set(text string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(text), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	p.text = &text
	p.hash = hash

	return nil
}

func (p *password) Compare(text string) error {
	return bcrypt.CompareHashAndPassword(p.hash, []byte(text))
}

func (s *UserStore) Create(ctx context.Context, tx *sql.Tx, user *User) error {
	query := `
		INSERT INTO users (username, password, email, role_id) VALUES 
		($1, $2, $3, $4)
		RETURNING id, created_at
	`
	err := tx.QueryRowContext(
		ctx,
		query,
		user.Username,
		user.Password.hash,
		user.Email,
		1,
	).Scan(
		&user.ID,
		&user.CreatedAt,
	)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		case err.Error() == `pq: duplicate key value violates unique constraint "users_username_key"`:
			return ErrDuplicateUsername
		default:
			return err
		}
	}
	return nil
}

func (s *UserStore) GetByID(ctx context.Context, userID, loggedInUserID int64) (*User, error) {
	query := `
		SELECT users.id, username, email, password, created_at, roles.*,
		EXISTS (SELECT 1 from followers where user_id = $2 and following_id = $1) as is_already_following
		FROM users
		JOIN roles ON (users.role_id = roles.id)
		WHERE users.id = $1 AND is_active = true
	`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	user := &User{}
	err := s.db.QueryRowContext(
		ctx,
		query,
		userID,
		loggedInUserID,
	).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password.hash,
		&user.CreatedAt,
		&user.Role.ID,
		&user.Role.Name,
		&user.Role.Level,
		&user.Role.Description,
		&user.IsAlreadyFollowing,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return user, nil
}

func (s *UserStore) GetPostsByUser(ctx context.Context, id, loggedInUserID int64) ([]Post, error) {
	// query := `
	// 	SELECT id, title, content, tags, updated_at
	// 	FROM posts
	// 	WHERE user_id = $1
	// `
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
		WHERE u.id = $1
	`
	var posts []Post
	rows, err := s.db.QueryContext(ctx, query, id, loggedInUserID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, errors.New("ErrNotFound")
		default:
			return nil, err
		}
	}
	for rows.Next() {
		var post Post
		err = rows.Scan(
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
			return nil, err
		}
		posts = append(posts, post)
	}

	return posts, nil
}

func (s *UserStore) GetByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, username, email, password, created_at FROM users
		WHERE email = $1 AND is_active = true
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	user := &User{}
	err := s.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password.hash,
		&user.CreatedAt,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return user, nil
}

func (s *UserStore) CreateAndInvite(ctx context.Context, user *User, token string, invitationExp time.Duration) error {
	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		if err := s.Create(ctx, tx, user); err != nil {
			return err
		}

		if err := s.createUserInvitation(ctx, tx, token, invitationExp, user.ID); err != nil {
			return err
		}

		return nil
	})
}

func (s *UserStore) createUserInvitation(ctx context.Context, tx *sql.Tx, token string, exp time.Duration, userID int64) error {
	query := `INSERT INTO user_invitations (token, user_id, expiry) VALUES ($1, $2, $3)`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, token, userID, time.Now().Add(exp))
	if err != nil {
		return err
	}

	return nil
}

func (s *UserStore) Activate(ctx context.Context, token string) error {
	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		// 1. find the user that this token belongs to
		user, err := s.getUserFromInvitation(ctx, tx, token)
		if err != nil {
			return err
		}

		// 2. update the user
		user.IsActive = true
		if err := s.update(ctx, tx, user); err != nil {
			return err
		}

		// 3. clean the invitations
		if err := s.deleteUserInvitations(ctx, tx, user.ID); err != nil {
			return err
		}

		return nil
	})
}

func (s *UserStore) getUserFromInvitation(ctx context.Context, tx *sql.Tx, token string) (*User, error) {
	query := `
		SELECT u.id, u.username, u.email, u.created_at, u.is_active
		FROM users u
		JOIN user_invitations ui ON u.id = ui.user_id
		WHERE ui.token = $1 AND ui.expiry > $2
	`

	hash := sha256.Sum256([]byte(token))
	hashToken := hex.EncodeToString(hash[:])

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	user := &User{}
	err := tx.QueryRowContext(ctx, query, hashToken, time.Now()).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.CreatedAt,
		&user.IsActive,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return user, nil
}

func (s *UserStore) update(ctx context.Context, tx *sql.Tx, user *User) error {
	query := `UPDATE users SET username = $1, email = $2, is_active = $3 WHERE id = $4`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, user.Username, user.Email, user.IsActive, user.ID)
	if err != nil {
		return err
	}

	return nil
}

func (s *UserStore) deleteUserInvitations(ctx context.Context, tx *sql.Tx, userID int64) error {
	query := `DELETE FROM user_invitations WHERE user_id = $1`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, userID)
	if err != nil {
		return err
	}

	return nil
}

func (s *UserStore) Delete(ctx context.Context, userID int64) error {
	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		if err := s.delete(ctx, tx, userID); err != nil {
			return err
		}

		if err := s.deleteUserInvitations(ctx, tx, userID); err != nil {
			return err
		}

		return nil
	})
}

func (s *UserStore) delete(ctx context.Context, tx *sql.Tx, id int64) error {
	query := `DELETE FROM users WHERE id = $1`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	return nil
}

func (s *UserStore) SaveUnsavePost(ctx context.Context, userID int64, postID int64) (bool, error) {
	var (
		exists bool
		query  string
	)
	err := s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM savedpost WHERE user_id = $1 AND savedpost_id = $2)", userID, postID).Scan(&exists)
	if err != nil {
		return exists, err
	}

	if !exists { // save the post
		query = `
			INSERT INTO savedpost (user_id, savedpost_id)
			VALUES ($1, $2) 
		`
	} else { // unsave the post
		query = `
			DELETE FROM savedpost WHERE user_id = $1 AND savedpost_id = $2 
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

func (s *UserStore) GetSavedPostsByUser(ctx context.Context, userID int64) ([]int64, error) {
	query := `SELECT savedpost_id FROM savedpost WHERE user_id = $1`

	var postIDSlice []int64
	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, errors.New("ErrNotFound")
		default:
			return nil, err
		}
	}
	for rows.Next() {
		var postID int64
		err = rows.Scan(
			&postID,
		)
		if err != nil {
			return nil, err
		}
		postIDSlice = append(postIDSlice, postID)
	}

	return postIDSlice, nil
}

func (s *UserStore) GetFollowers(ctx context.Context, userID int64) ([]int64, error) {
	query := `
		SELECT user_id FROM followers WHERE following_id = $1;
	`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var followers []int64
	for rows.Next() {
		var followerID int64
		if err := rows.Scan(&followerID); err != nil {
			return nil, err
		}
		followers = append(followers, followerID)
	}
	return followers, nil
}

func (s *UserStore) GetFollowings(ctx context.Context, userID int64) ([]int64, error) {
	query := `
		SELECT following_id FROM followers WHERE user_id = $1;
	`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var followings []int64
	for rows.Next() {
		var followerID int64
		if err := rows.Scan(&followerID); err != nil {
			return nil, err
		}
		followings = append(followings, followerID)
	}
	return followings, nil
}

func (s *UserStore) GetByWildCard(ctx context.Context, name string) ([]compactUserPayload, error) {
	query := `
		SELECT id, username FROM users WHERE username ILIKE $1;
	`

	var (
		user  compactUserPayload
		users []compactUserPayload
	)
	rows, err := s.db.QueryContext(ctx, query, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&user.Id, &user.Name); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}
