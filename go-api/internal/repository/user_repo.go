package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// UserRepo handles user-related database operations.
type UserRepo struct {
	DB *pgxpool.Pool
}

// ExistsByEmail checks whether a user with the given email exists.
func (r *UserRepo) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := r.DB.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", email).Scan(&exists)
	return exists, err
}

// Create inserts a new user and returns its ID.
func (r *UserRepo) Create(ctx context.Context, email, hashedPwd, firstName, lastName string) (string, error) {
	var userID string
	err := r.DB.QueryRow(ctx,
		`INSERT INTO users (email, password_hash, first_name, last_name)
		 VALUES ($1, $2, $3, $4) RETURNING id`,
		email, hashedPwd, firstName, lastName).Scan(&userID)
	return userID, err
}

// FindByEmail returns (userID, passwordHash) for the given email.
func (r *UserRepo) FindByEmail(ctx context.Context, email string) (userID, hashedPwd string, err error) {
	err = r.DB.QueryRow(ctx,
		"SELECT id, password_hash FROM users WHERE email = $1", email).Scan(&userID, &hashedPwd)
	return
}

// FindEmailByID returns the email for a given user ID.
func (r *UserRepo) FindEmailByID(ctx context.Context, userID string) (string, error) {
	var email string
	err := r.DB.QueryRow(ctx,
		"SELECT email FROM users WHERE id = $1", userID).Scan(&email)
	return email, err
}

// StoreRefreshToken persists a refresh token with its expiry.
func (r *UserRepo) StoreRefreshToken(ctx context.Context, userID, token string, expiresAt time.Time) error {
	_, err := r.DB.Exec(ctx,
		`INSERT INTO refresh_tokens (user_id, token, expires_at)
		 VALUES ($1, $2, $3)`,
		userID, token, expiresAt)
	return err
}

// RefreshTokenExists checks whether a non-expired refresh token is in the DB.
func (r *UserRepo) RefreshTokenExists(ctx context.Context, token string) (bool, error) {
	var exists bool
	err := r.DB.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM refresh_tokens WHERE token = $1 AND expires_at > NOW())",
		token).Scan(&exists)
	return exists, err
}
