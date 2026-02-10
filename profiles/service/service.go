package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"matchmaking-service/profiles/model"
	"matchmaking-service/profiles/repository"
)

var ErrForbidden = errors.New("forbidden")

type Service struct{ store repository.Store }

func New(store repository.Store) *Service { return &Service{store: store} }

func (s *Service) CreateProfile(ctx context.Context, actorUserID string, p model.Profile) (model.Profile, error) {
	if p.UserID != actorUserID {
		return model.Profile{}, ErrForbidden
	}
	if err := validate(p); err != nil {
		return model.Profile{}, err
	}
	return s.store.Create(ctx, p)
}

func (s *Service) UpdateProfile(ctx context.Context, actorUserID string, p model.Profile) (model.Profile, error) {
	if p.UserID != actorUserID {
		return model.Profile{}, ErrForbidden
	}
	if err := validate(p); err != nil {
		return model.Profile{}, err
	}
	return s.store.Update(ctx, p)
}

func (s *Service) GetProfile(ctx context.Context, userID string) (model.Profile, error) {
	return s.store.GetByID(ctx, userID)
}

func (s *Service) ListProfiles(ctx context.Context, page, size int) (model.ListResult, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	if size > 100 {
		size = 100
	}
	offset := (page - 1) * size
	items, err := s.store.List(ctx, offset, size)
	if err != nil {
		return model.ListResult{}, err
	}
	return model.ListResult{Items: items, Page: page, Size: size}, nil
}

func validate(p model.Profile) error {
	if p.UserID == "" {
		return fmt.Errorf("user_id is required")
	}
	if p.Age < 18 || p.Age > 100 {
		return fmt.Errorf("age must be between 18 and 100")
	}
	if strings.TrimSpace(p.Profession) == "" {
		return fmt.Errorf("profession is required")
	}
	if strings.TrimSpace(p.Education) == "" {
		return fmt.Errorf("education is required")
	}
	if p.Income < 0 {
		return fmt.Errorf("income must be >= 0")
	}
	if strings.TrimSpace(p.ResidencyStatus) == "" {
		return fmt.Errorf("residency_status is required")
	}
	if strings.TrimSpace(p.Location) == "" {
		return fmt.Errorf("location is required")
	}
	if strings.TrimSpace(p.MaritalStatus) == "" {
		return fmt.Errorf("marital_status is required")
	}
	if len(p.Description) > 1000 {
		return fmt.Errorf("description too long")
	}
	return nil
}
