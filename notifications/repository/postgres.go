package repository

import (
	"context"
	"database/sql"
	"errors"

	"matchmaking-service/notifications/model"
)

var ErrNotFound = errors.New("not found")

type Store interface {
	ListUnread(ctx context.Context, userID string, limit int) ([]model.Notification, error)
	MarkRead(ctx context.Context, userID, notificationID string) error
}

type PostgresStore struct{ DB *sql.DB }

func NewPostgresStore(db *sql.DB) *PostgresStore { return &PostgresStore{DB: db} }

func (s *PostgresStore) ListUnread(ctx context.Context, userID string, limit int) ([]model.Notification, error) {
	q := `SELECT id, user_id, type, title, body, read_at, created_at
	FROM notifications
	WHERE user_id=$1 AND read_at IS NULL
	ORDER BY created_at DESC
	LIMIT $2`
	rows, err := s.DB.QueryContext(ctx, q, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]model.Notification, 0)
	for rows.Next() {
		var n model.Notification
		if err := rows.Scan(&n.ID, &n.UserID, &n.Type, &n.Title, &n.Body, &n.ReadAt, &n.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

func (s *PostgresStore) MarkRead(ctx context.Context, userID, notificationID string) error {
	res, err := s.DB.ExecContext(ctx, `UPDATE notifications SET read_at=NOW() WHERE id=$1 AND user_id=$2 AND read_at IS NULL`, notificationID, userID)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}
