package middleware

import (
	"context"
	"net/http"
	"strings"

	"matchmaking-service/auth/service"
)

type ctxKey string

const UserIDKey ctxKey = "user_id"

func Protected(svc *service.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			parts := strings.SplitN(auth, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				http.Error(w, "missing token", http.StatusUnauthorized)
				return
			}
			userID, err := svc.UserFromAccessToken(parts[1])
			if err != nil {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			w.Header().Set("X-User-ID", userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
