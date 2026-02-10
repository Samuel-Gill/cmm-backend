package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	DatabaseURL           string
	JWTSecret             string
	AccessTokenTTL        time.Duration
	RefreshTokenTTL       time.Duration
	ResetPasswordTokenTTL time.Duration
	PasswordAlgorithm     string
	ServerAddr            string
}

func Load() Config {
	return Config{
		DatabaseURL:           getenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/matchmaking?sslmode=disable"),
		JWTSecret:             getenv("JWT_SECRET", "dev-secret"),
		AccessTokenTTL:        durationEnv("ACCESS_TOKEN_TTL_MINUTES", 15, time.Minute),
		RefreshTokenTTL:       durationEnv("REFRESH_TOKEN_TTL_HOURS", 24*7, time.Hour),
		ResetPasswordTokenTTL: durationEnv("RESET_TOKEN_TTL_MINUTES", 30, time.Minute),
		PasswordAlgorithm:     getenv("PASSWORD_ALGO", "bcrypt"),
		ServerAddr:            getenv("SERVER_ADDR", ":8080"),
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func durationEnv(key string, fallback int, unit time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return time.Duration(fallback) * unit
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return time.Duration(fallback) * unit
	}
	return time.Duration(i) * unit
}
