package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"matchmaking-service/matches/model"
)

type Store interface {
	Browse(ctx context.Context, actorUserID string, f model.BrowseFilter) ([]model.BrowseProfile, error)
	LikeProfile(ctx context.Context, actorUserID, likedUserID string) (bool, error)
	LikedBy(ctx context.Context, userID string) ([]string, error)
	Liked(ctx context.Context, userID string) ([]string, error)
	MutualMatches(ctx context.Context, userID string) ([]string, error)
}

type PostgresStore struct{ DB *sql.DB }

func NewPostgresStore(db *sql.DB) *PostgresStore { return &PostgresStore{DB: db} }

func (s *PostgresStore) Browse(ctx context.Context, actorUserID string, f model.BrowseFilter) ([]model.BrowseProfile, error) {
	args := []any{actorUserID}
	conditions := []string{"p.user_id <> $1", "p.is_discoverable = TRUE"}

	if f.MinAge > 0 {
		args = append(args, f.MinAge)
		conditions = append(conditions, fmt.Sprintf("p.age >= $%d", len(args)))
	}
	if f.MaxAge > 0 {
		args = append(args, f.MaxAge)
		conditions = append(conditions, fmt.Sprintf("p.age <= $%d", len(args)))
	}
	if strings.TrimSpace(f.Gender) != "" {
		args = append(args, strings.TrimSpace(f.Gender))
		conditions = append(conditions, fmt.Sprintf("p.gender = $%d", len(args)))
	}
	if f.MinIncome > 0 {
		args = append(args, f.MinIncome)
		conditions = append(conditions, fmt.Sprintf("p.income >= $%d", len(args)))
	}
	if f.MaxIncome > 0 {
		args = append(args, f.MaxIncome)
		conditions = append(conditions, fmt.Sprintf("p.income <= $%d", len(args)))
	}
	if strings.TrimSpace(f.Location) != "" {
		args = append(args, strings.TrimSpace(f.Location))
		conditions = append(conditions, fmt.Sprintf("p.location = $%d", len(args)))
	}
	if strings.TrimSpace(f.ResidencyStatus) != "" {
		args = append(args, strings.TrimSpace(f.ResidencyStatus))
		conditions = append(conditions, fmt.Sprintf("p.residency_status = $%d", len(args)))
	}

	offset := (f.Page - 1) * f.Size
	args = append(args, offset, f.Size)
	offsetIdx := len(args) - 1
	limitIdx := len(args)

	q := fmt.Sprintf(`
		SELECT p.user_id, p.age, p.gender, p.profession, p.education, p.income, p.location, p.residency_status, p.marital_status, p.description, p.created_at
		FROM matchmaking_profiles p
		WHERE %s
		ORDER BY p.created_at DESC
		OFFSET $%d LIMIT $%d`, strings.Join(conditions, " AND "), offsetIdx, limitIdx)

	rows, err := s.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]model.BrowseProfile, 0)
	for rows.Next() {
		var p model.BrowseProfile
		if err := rows.Scan(&p.UserID, &p.Age, &p.Gender, &p.Profession, &p.Education, &p.Income, &p.Location, &p.ResidencyStatus, &p.MaritalStatus, &p.Description, &p.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *PostgresStore) LikeProfile(ctx context.Context, actorUserID, likedUserID string) (bool, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO profile_likes (liker_user_id, liked_user_id, status, created_at, updated_at)
		VALUES ($1,$2,'liked',NOW(),NOW())
		ON CONFLICT (liker_user_id, liked_user_id)
		DO UPDATE SET status='liked', updated_at=NOW()`, actorUserID, likedUserID)
	if err != nil {
		return false, err
	}
	if err := notifyLike(ctx, tx, likedUserID, actorUserID); err != nil {
		return false, err
	}

	var reciprocal bool
	err = tx.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM profile_likes
			WHERE liker_user_id=$1 AND liked_user_id=$2 AND status='liked'
		)`, likedUserID, actorUserID).Scan(&reciprocal)
	if err != nil {
		return false, err
	}

	if reciprocal {
		low, high := actorUserID, likedUserID
		if high < low {
			low, high = high, low
		}
		_, err = tx.ExecContext(ctx, `
			INSERT INTO matches (user_low_id, user_high_id, status, matched_at, updated_at)
			VALUES ($1,$2,'active',NOW(),NOW())
			ON CONFLICT (user_low_id, user_high_id)
			DO UPDATE SET status='active', updated_at=NOW()`, low, high)
		if err != nil {
			return false, err
		}
		if err := notifyMutual(ctx, tx, actorUserID, likedUserID); err != nil {
			return false, err
		}
		if err := notifyMutual(ctx, tx, likedUserID, actorUserID); err != nil {
			return false, err
		}
	}

	if err := tx.Commit(); err != nil {
		return false, err
	}
	return reciprocal, nil
}

func (s *PostgresStore) LikedBy(ctx context.Context, userID string) ([]string, error) {
	q := `SELECT liker_user_id FROM profile_likes WHERE liked_user_id=$1 AND status='liked' ORDER BY created_at DESC`
	return s.queryIDs(ctx, q, userID)
}

func (s *PostgresStore) Liked(ctx context.Context, userID string) ([]string, error) {
	q := `SELECT liked_user_id FROM profile_likes WHERE liker_user_id=$1 AND status='liked' ORDER BY created_at DESC`
	return s.queryIDs(ctx, q, userID)
}

func (s *PostgresStore) MutualMatches(ctx context.Context, userID string) ([]string, error) {
	q := `
		SELECT CASE WHEN user_low_id=$1 THEN user_high_id ELSE user_low_id END AS matched_user
		FROM matches
		WHERE (user_low_id=$1 OR user_high_id=$1) AND status='active'
		ORDER BY matched_at DESC`
	return s.queryIDs(ctx, q, userID)
}

func (s *PostgresStore) queryIDs(ctx context.Context, q, userID string) ([]string, error) {
	rows, err := s.DB.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]string, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

func notifyLike(ctx context.Context, tx *sql.Tx, toUserID, fromUserID string) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO notifications (user_id, type, title, body)
	VALUES ($1, 'liked', 'You got a new like', $2)`, toUserID, "User "+fromUserID+" liked your profile")
	return err
}

func notifyMutual(ctx context.Context, tx *sql.Tx, userID, matchedUserID string) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO notifications (user_id, type, title, body)
	VALUES ($1, 'match', 'It's a match!', $2)`, userID, "You matched with user "+matchedUserID)
	return err
}
