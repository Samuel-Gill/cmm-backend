package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"matchmaking-service/auth/model"
	"matchmaking-service/auth/password"
	"matchmaking-service/auth/repository"
	"matchmaking-service/auth/token"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type Config struct {
	AccessTokenTTL        time.Duration
	RefreshTokenTTL       time.Duration
	ResetPasswordTokenTTL time.Duration
}

type Service struct {
	store  repository.Store
	hasher password.Hasher
	tokens token.Manager
	cfg    Config
}

type AuthTokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func New(store repository.Store, hasher password.Hasher, tokens token.Manager, cfg Config) *Service {
	return &Service{store: store, hasher: hasher, tokens: tokens, cfg: cfg}
}

func (s *Service) Signup(ctx context.Context, email, pass string) (model.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" || len(pass) < 8 {
		return model.User{}, fmt.Errorf("invalid payload")
	}
	hash, err := s.hasher.Hash(pass)
	if err != nil {
		return model.User{}, err
	}
	return s.store.CreateUser(ctx, email, hash)
}

func (s *Service) Login(ctx context.Context, email, pass string) (AuthTokens, error) {
	u, err := s.store.GetUserByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
	if err != nil {
		return AuthTokens{}, ErrInvalidCredentials
	}
	if err := s.hasher.Compare(u.PasswordHash, pass); err != nil {
		return AuthTokens{}, ErrInvalidCredentials
	}
	access, err := s.tokens.CreateAccessToken(u.ID, s.cfg.AccessTokenTTL)
	if err != nil {
		return AuthTokens{}, err
	}
	refreshPlain, refreshHash, err := token.NewOpaqueToken()
	if err != nil {
		return AuthTokens{}, err
	}
	err = s.store.SaveRefreshToken(ctx, model.RefreshToken{UserID: u.ID, TokenHash: refreshHash, ExpiresAt: time.Now().Add(s.cfg.RefreshTokenTTL)})
	if err != nil {
		return AuthTokens{}, err
	}
	return AuthTokens{AccessToken: access, RefreshToken: refreshPlain}, nil
}

func (s *Service) Logout(ctx context.Context, refreshPlain string) error {
	_, refreshHash, err := token.NewOpaqueTokenFromPlain(refreshPlain)
	if err != nil {
		return err
	}
	return s.store.RevokeRefreshToken(ctx, refreshHash)
}

func (s *Service) RequestPasswordReset(ctx context.Context, email string) (string, error) {
	u, err := s.store.GetUserByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
	if err != nil {
		return "", nil
	}
	plain, hash, err := token.NewOpaqueToken()
	if err != nil {
		return "", err
	}
	err = s.store.SavePasswordReset(ctx, model.PasswordReset{UserID: u.ID, TokenHash: hash, ExpiresAt: time.Now().Add(s.cfg.ResetPasswordTokenTTL)})
	if err != nil {
		return "", err
	}
	return plain, nil
}

func (s *Service) ResetPassword(ctx context.Context, resetPlain, newPassword string) error {
	if len(newPassword) < 8 {
		return fmt.Errorf("password too short")
	}
	_, hash, err := token.NewOpaqueTokenFromPlain(resetPlain)
	if err != nil {
		return err
	}
	reset, err := s.store.ConsumePasswordReset(ctx, hash)
	if err != nil {
		return ErrInvalidCredentials
	}
	pwdHash, err := s.hasher.Hash(newPassword)
	if err != nil {
		return err
	}
	return s.store.UpdatePassword(ctx, reset.UserID, pwdHash)
}

func (s *Service) ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	u, err := s.store.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	if err := s.hasher.Compare(u.PasswordHash, oldPassword); err != nil {
		return ErrInvalidCredentials
	}
	if len(newPassword) < 8 {
		return fmt.Errorf("password too short")
	}
	h, err := s.hasher.Hash(newPassword)
	if err != nil {
		return err
	}
	return s.store.UpdatePassword(ctx, userID, h)
}

func (s *Service) UserFromAccessToken(token string) (string, error) {
	return s.tokens.ParseAccessToken(token)
}
