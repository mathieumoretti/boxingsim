package db

import (
	"database/sql"
	"errors"
	"time"

	"boxing/internal/model"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrDuplicateUser = errors.New("user already exists")
)

// GetUserByID retrieves a user by ID
func GetUserByID(db *sql.DB, id int) (*model.User, error) {
	query := `
		SELECT id, username, email, password_hash, role, created_at, updated_at
		FROM users
		WHERE id = ?
	`
	user := &model.User{}
	err := db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
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
		SELECT id, username, email, password_hash, role, created_at, updated_at
		FROM users
		WHERE username = ?
	`
	user := &model.User{}
	err := db.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
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
		SET username = ?, email = ?, password_hash = ?, role = ?, updated_at = ?
		WHERE id = ?
	`
	_, err := db.Exec(query, user.Username, user.Email, user.PasswordHash, user.Role, time.Now(), user.ID)
	return err
}

// ListUsers retrieves all users
func ListUsers(db *sql.DB) ([]*model.User, error) {
	query := `
		SELECT id, username, email, password_hash, role, created_at, updated_at
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
			&user.PasswordHash,
			&user.Role,
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