package repository

import (
	"context"
	"database/sql"
	"errors"

	"matchmaking-service/auth/model"
)

var ErrNotFound = errors.New("not found")

type Store interface {
	CreateUser(ctx context.Context, email, passwordHash string) (model.User, error)
	GetUserByEmail(ctx context.Context, email string) (model.User, error)
	GetUserByID(ctx context.Context, id string) (model.User, error)
	UpdatePassword(ctx context.Context, userID, passwordHash string) error
	SaveRefreshToken(ctx context.Context, token model.RefreshToken) error
	RevokeRefreshToken(ctx context.Context, tokenHash string) error
	GetActiveRefreshToken(ctx context.Context, tokenHash string) (model.RefreshToken, error)
	SavePasswordReset(ctx context.Context, reset model.PasswordReset) error
	ConsumePasswordReset(ctx context.Context, tokenHash string) (model.PasswordReset, error)
}

type PostgresStore struct{ DB *sql.DB }

func NewPostgresStore(db *sql.DB) *PostgresStore { return &PostgresStore{DB: db} }

func (s *PostgresStore) CreateUser(ctx context.Context, email, passwordHash string) (model.User, error) {
	q := `INSERT INTO auth_users (email, password_hash) VALUES ($1,$2)
	RETURNING id,email,password_hash,account_state,created_at`
	var u model.User
	err := s.DB.QueryRowContext(ctx, q, email, passwordHash).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.AccountState, &u.CreatedAt)
	return u, err
}

func (s *PostgresStore) GetUserByEmail(ctx context.Context, email string) (model.User, error) {
	q := `SELECT id,email,password_hash,account_state,created_at FROM auth_users WHERE email=$1`
	var u model.User
	err := s.DB.QueryRowContext(ctx, q, email).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.AccountState, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return u, ErrNotFound
	}
	return u, err
}

func (s *PostgresStore) GetUserByID(ctx context.Context, id string) (model.User, error) {
	q := `SELECT id,email,password_hash,account_state,created_at FROM auth_users WHERE id=$1`
	var u model.User
	err := s.DB.QueryRowContext(ctx, q, id).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.AccountState, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return u, ErrNotFound
	}
	return u, err
}

func (s *PostgresStore) UpdatePassword(ctx context.Context, userID, passwordHash string) error {
	_, err := s.DB.ExecContext(ctx, `UPDATE auth_users SET password_hash=$1,updated_at=NOW() WHERE id=$2`, passwordHash, userID)
	return err
}

func (s *PostgresStore) SaveRefreshToken(ctx context.Context, token model.RefreshToken) error {
	_, err := s.DB.ExecContext(ctx, `INSERT INTO auth_refresh_tokens (user_id,token_hash,expires_at) VALUES ($1,$2,$3)`, token.UserID, token.TokenHash, token.ExpiresAt)
	return err
}

func (s *PostgresStore) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	_, err := s.DB.ExecContext(ctx, `UPDATE auth_refresh_tokens SET revoked_at=NOW() WHERE token_hash=$1`, tokenHash)
	return err
}

func (s *PostgresStore) GetActiveRefreshToken(ctx context.Context, tokenHash string) (model.RefreshToken, error) {
	q := `SELECT user_id,token_hash,expires_at,revoked_at FROM auth_refresh_tokens WHERE token_hash=$1 AND revoked_at IS NULL AND expires_at > NOW()`
	var t model.RefreshToken
	err := s.DB.QueryRowContext(ctx, q, tokenHash).Scan(&t.UserID, &t.TokenHash, &t.ExpiresAt, &t.RevokedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return t, ErrNotFound
	}
	return t, err
}

func (s *PostgresStore) SavePasswordReset(ctx context.Context, reset model.PasswordReset) error {
	_, err := s.DB.ExecContext(ctx, `INSERT INTO auth_password_resets (user_id,token_hash,expires_at) VALUES ($1,$2,$3)`, reset.UserID, reset.TokenHash, reset.ExpiresAt)
	return err
}

func (s *PostgresStore) ConsumePasswordReset(ctx context.Context, tokenHash string) (model.PasswordReset, error) {
	q := `UPDATE auth_password_resets SET used_at=NOW() WHERE token_hash=$1 AND used_at IS NULL AND expires_at > NOW()
	RETURNING user_id,token_hash,expires_at,used_at`
	var p model.PasswordReset
	err := s.DB.QueryRowContext(ctx, q, tokenHash).Scan(&p.UserID, &p.TokenHash, &p.ExpiresAt, &p.UsedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return p, ErrNotFound
	}
	return p, err
}
