package service

import (
	"context"

	"matchmaking-service/notifications/model"
	"matchmaking-service/notifications/repository"
)

type Service struct{ store repository.Store }

func New(store repository.Store) *Service { return &Service{store: store} }

func (s *Service) Unread(ctx context.Context, userID string, limit int) ([]model.Notification, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	return s.store.ListUnread(ctx, userID, limit)
}

func (s *Service) MarkRead(ctx context.Context, userID, notificationID string) error {
	return s.store.MarkRead(ctx, userID, notificationID)
}
