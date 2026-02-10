package repository

import (
	"context"
	"database/sql"
	"errors"

	"matchmaking-service/profiles/model"
)

var ErrNotFound = errors.New("not found")

type Store interface {
	Create(ctx context.Context, p model.Profile) (model.Profile, error)
	Update(ctx context.Context, p model.Profile) (model.Profile, error)
	GetByID(ctx context.Context, userID string) (model.Profile, error)
	List(ctx context.Context, offset, limit int) ([]model.Profile, error)
}

type PostgresStore struct{ DB *sql.DB }

func NewPostgresStore(db *sql.DB) *PostgresStore { return &PostgresStore{DB: db} }

func (s *PostgresStore) Create(ctx context.Context, p model.Profile) (model.Profile, error) {
	q := `INSERT INTO matchmaking_profiles
	(user_id, age, profession, education, income, residency_status, location, marital_status, description,
	 hide_contact_info, hide_address, hide_income, hide_visa_status)
	VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
	RETURNING user_id, age, profession, education, income, residency_status, location, marital_status, description,
	 hide_contact_info, hide_address, hide_income, hide_visa_status, created_at, updated_at`
	err := s.DB.QueryRowContext(ctx, q,
		p.UserID, p.Age, p.Profession, p.Education, p.Income, p.ResidencyStatus, p.Location, p.MaritalStatus, p.Description,
		p.HideContactInfo, p.HideAddress, p.HideIncome, p.HideVisaStatus,
	).Scan(
		&p.UserID, &p.Age, &p.Profession, &p.Education, &p.Income, &p.ResidencyStatus, &p.Location, &p.MaritalStatus, &p.Description,
		&p.HideContactInfo, &p.HideAddress, &p.HideIncome, &p.HideVisaStatus, &p.CreatedAt, &p.UpdatedAt,
	)
	return p, err
}

func (s *PostgresStore) Update(ctx context.Context, p model.Profile) (model.Profile, error) {
	q := `UPDATE matchmaking_profiles SET
	age=$2, profession=$3, education=$4, income=$5, residency_status=$6, location=$7, marital_status=$8, description=$9,
	hide_contact_info=$10, hide_address=$11, hide_income=$12, hide_visa_status=$13, updated_at=NOW()
	WHERE user_id=$1
	RETURNING user_id, age, profession, education, income, residency_status, location, marital_status, description,
	 hide_contact_info, hide_address, hide_income, hide_visa_status, created_at, updated_at`
	err := s.DB.QueryRowContext(ctx, q,
		p.UserID, p.Age, p.Profession, p.Education, p.Income, p.ResidencyStatus, p.Location, p.MaritalStatus, p.Description,
		p.HideContactInfo, p.HideAddress, p.HideIncome, p.HideVisaStatus,
	).Scan(
		&p.UserID, &p.Age, &p.Profession, &p.Education, &p.Income, &p.ResidencyStatus, &p.Location, &p.MaritalStatus, &p.Description,
		&p.HideContactInfo, &p.HideAddress, &p.HideIncome, &p.HideVisaStatus, &p.CreatedAt, &p.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return p, ErrNotFound
	}
	return p, err
}

func (s *PostgresStore) GetByID(ctx context.Context, userID string) (model.Profile, error) {
	q := `SELECT user_id, age, profession, education, income, residency_status, location, marital_status, description,
	hide_contact_info, hide_address, hide_income, hide_visa_status, created_at, updated_at
	FROM matchmaking_profiles WHERE user_id=$1`
	var p model.Profile
	err := s.DB.QueryRowContext(ctx, q, userID).Scan(
		&p.UserID, &p.Age, &p.Profession, &p.Education, &p.Income, &p.ResidencyStatus, &p.Location, &p.MaritalStatus, &p.Description,
		&p.HideContactInfo, &p.HideAddress, &p.HideIncome, &p.HideVisaStatus, &p.CreatedAt, &p.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return p, ErrNotFound
	}
	return p, err
}

func (s *PostgresStore) List(ctx context.Context, offset, limit int) ([]model.Profile, error) {
	q := `SELECT user_id, age, profession, education, income, residency_status, location, marital_status, description,
	hide_contact_info, hide_address, hide_income, hide_visa_status, created_at, updated_at
	FROM matchmaking_profiles ORDER BY created_at DESC OFFSET $1 LIMIT $2`
	rows, err := s.DB.QueryContext(ctx, q, offset, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]model.Profile, 0)
	for rows.Next() {
		var p model.Profile
		if err := rows.Scan(
			&p.UserID, &p.Age, &p.Profession, &p.Education, &p.Income, &p.ResidencyStatus, &p.Location, &p.MaritalStatus, &p.Description,
			&p.HideContactInfo, &p.HideAddress, &p.HideIncome, &p.HideVisaStatus, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}
