package service

import (
	"context"
	"fmt"

	"matchmaking-service/matches/model"
	"matchmaking-service/matches/repository"
)

type Service struct{ store repository.Store }

func New(store repository.Store) *Service { return &Service{store: store} }

func (s *Service) Browse(ctx context.Context, actorUserID string, f model.BrowseFilter) (model.BrowseResult, error) {
	if f.Page <= 0 {
		f.Page = 1
	}
	if f.Size <= 0 {
		f.Size = 20
	}
	if f.Size > 100 {
		f.Size = 100
	}
	if f.MinAge > 0 && f.MaxAge > 0 && f.MinAge > f.MaxAge {
		return model.BrowseResult{}, fmt.Errorf("min_age must be <= max_age")
	}
	if f.MinIncome > 0 && f.MaxIncome > 0 && f.MinIncome > f.MaxIncome {
		return model.BrowseResult{}, fmt.Errorf("min_income must be <= max_income")
	}

	items, err := s.store.Browse(ctx, actorUserID, f)
	if err != nil {
		return model.BrowseResult{}, err
	}
	return model.BrowseResult{Items: items, Page: f.Page, Size: f.Size}, nil
}

func (s *Service) Like(ctx context.Context, actorUserID, likedUserID string) (map[string]any, error) {
	if actorUserID == "" || likedUserID == "" {
		return nil, fmt.Errorf("actor and liked user IDs are required")
	}
	if actorUserID == likedUserID {
		return nil, fmt.Errorf("cannot like your own profile")
	}
	matched, err := s.store.LikeProfile(ctx, actorUserID, likedUserID)
	if err != nil {
		return nil, err
	}
	return map[string]any{"liked_user_id": likedUserID, "matched": matched}, nil
}

func (s *Service) Relationships(ctx context.Context, actorUserID string) (model.RelationshipSummary, error) {
	likedBy, err := s.store.LikedBy(ctx, actorUserID)
	if err != nil {
		return model.RelationshipSummary{}, err
	}
	liked, err := s.store.Liked(ctx, actorUserID)
	if err != nil {
		return model.RelationshipSummary{}, err
	}
	matches, err := s.store.MutualMatches(ctx, actorUserID)
	if err != nil {
		return model.RelationshipSummary{}, err
	}
	return model.RelationshipSummary{LikedBy: likedBy, Liked: liked, MutualMatches: matches}, nil
}
