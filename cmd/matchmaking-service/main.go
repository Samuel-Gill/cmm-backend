package main

import (
	"database/sql"
	"log"
	"net/http"
	"strings"

	_ "github.com/lib/pq"
	"matchmaking-service/auth/password"
	authrepository "matchmaking-service/auth/repository"
	authservice "matchmaking-service/auth/service"
	"matchmaking-service/auth/token"
	"matchmaking-service/config"
	matchesrepository "matchmaking-service/matches/repository"
	matchesservice "matchmaking-service/matches/service"
	notificationsrepository "matchmaking-service/notifications/repository"
	notificationsservice "matchmaking-service/notifications/service"
	profilesrepository "matchmaking-service/profiles/repository"
	profilesservice "matchmaking-service/profiles/service"
	"matchmaking-service/routing"
)

func main() {
	cfg := config.Load()
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var hasher password.Hasher = password.BCryptHasher{}
	if strings.EqualFold(cfg.PasswordAlgorithm, "argon2") {
		hasher = password.Argon2Hasher{}
	}

	authSvc := authservice.New(
		authrepository.NewPostgresStore(db),
		hasher,
		token.NewManager(cfg.JWTSecret),
		authservice.Config{
			AccessTokenTTL:        cfg.AccessTokenTTL,
			RefreshTokenTTL:       cfg.RefreshTokenTTL,
			ResetPasswordTokenTTL: cfg.ResetPasswordTokenTTL,
		},
	)
	profilesSvc := profilesservice.New(profilesrepository.NewPostgresStore(db))
	matchesSvc := matchesservice.New(matchesrepository.NewPostgresStore(db))
	notificationsSvc := notificationsservice.New(notificationsrepository.NewPostgresStore(db))

	log.Printf("starting matchmaking-service on %s", cfg.ServerAddr)
	if err := http.ListenAndServe(cfg.ServerAddr, routing.NewRouter(authSvc, profilesSvc, matchesSvc, notificationsSvc)); err != nil {
		log.Fatal(err)
	}
}
