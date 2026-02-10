package routing

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	authhttp "matchmaking-service/auth/http"
	authmw "matchmaking-service/auth/middleware"
	authservice "matchmaking-service/auth/service"
	applog "matchmaking-service/logging"
	matcheshttp "matchmaking-service/matches/http"
	matchesservice "matchmaking-service/matches/service"
	notificationshttp "matchmaking-service/notifications/http"
	notificationsservice "matchmaking-service/notifications/service"
	profileshttp "matchmaking-service/profiles/http"
	profilesservice "matchmaking-service/profiles/service"
)

func NewRouter(authSvc *authservice.Service, profilesSvc *profilesservice.Service, matchesSvc *matchesservice.Service, notificationsSvc *notificationsservice.Service) http.Handler {
	r := chi.NewRouter()
	r.Use(applog.RequestID())
	r.Use(applog.Recoverer())
	r.Use(applog.RequestLogger())
	authHandler := authhttp.NewHandler(authSvc)
	profilesHandler := profileshttp.NewHandler(profilesSvc)
	matchesHandler := matcheshttp.NewHandler(matchesSvc)
	notificationsHandler := notificationshttp.NewHandler(notificationsSvc)

	r.Route("/auth", func(r chi.Router) {
		r.Post("/signup", authHandler.Signup)
		r.Post("/login", authHandler.Login)
		r.Post("/logout", authHandler.Logout)
		r.Post("/password/reset/request", authHandler.RequestResetPassword)
		r.Post("/password/reset/confirm", authHandler.ResetPassword)

		r.Group(func(r chi.Router) {
			r.Use(authmw.Protected(authSvc))
			r.Post("/password/change", authHandler.ChangePassword)
			r.Get("/me", authHandler.Me)
		})
	})

	r.Route("/profiles", func(r chi.Router) {
		r.Get("/", profilesHandler.List)
		r.Get("/{userID}", profilesHandler.GetByID)

		r.Group(func(r chi.Router) {
			r.Use(authmw.Protected(authSvc))
			r.Post("/", profilesHandler.Create)
			r.Put("/{userID}", profilesHandler.Update)
		})
	})

	r.Route("/matches", func(r chi.Router) {
		r.Use(authmw.Protected(authSvc))
		r.Get("/browse", matchesHandler.Browse)
		r.Post("/like", matchesHandler.LikeProfile)
		r.Get("/relationships", matchesHandler.Relationships)
	})

	r.Route("/notifications", func(r chi.Router) {
		r.Use(authmw.Protected(authSvc))
		r.Get("/unread", notificationsHandler.Unread)
		r.Post("/{notificationID}/read", notificationsHandler.MarkRead)
	})

	return r
}
