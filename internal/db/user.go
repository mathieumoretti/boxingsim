package db

import (
	"database/sql"
	"errors"
	"time"

	"github.com/mormm/boxing/internal/model"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrDuplicateUser = errors.New("user already exists")
)

// GetUserByID retrieves a user by ID
func GetUserByID(db *sql.DB, id int) (*model.User, error) {
	query := `
		SELECT id, username, email, hashed_password, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	user := &model.User{}
	err := db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.HashedPassword,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

// GetUserByUsername retrieves a user by username
func GetUserByUsername(db *sql.DB, username string) (*model.User, error) {
	query := `
		SELECT id, username, email, hashed_password, created_at, updated_at
		FROM users
		WHERE username = $1
	`
	user := &model.User{}
	err := db.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.HashedPassword,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

// UpdateUser updates a user's information
func UpdateUser(db *sql.DB, user *model.User) error {
	query := `
		UPDATE users
		SET username = $1, email = $2, hashed_password = $3, updated_at = $4
		WHERE id = $5
	`
	_, err := db.Exec(query, user.Username, user.Email, user.HashedPassword, time.Now(), user.ID)
	return err
}

// ListUsers retrieves all users
func ListUsers(db *sql.DB) ([]*model.User, error) {
	query := `
		SELECT id, username, email, hashed_password, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
	`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []*model.User{}
	for rows.Next() {
		user := &model.User{}
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.HashedPassword,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, rows.Err()
}