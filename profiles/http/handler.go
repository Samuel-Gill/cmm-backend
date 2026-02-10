package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"matchmaking-service/auth/middleware"
	"matchmaking-service/profiles/model"
	"matchmaking-service/profiles/repository"
	"matchmaking-service/profiles/service"
)

type Handler struct{ svc *service.Service }

func NewHandler(svc *service.Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	actorID, _ := r.Context().Value(middleware.UserIDKey).(string)
	var p model.Profile
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	profile, err := h.svc.CreateProfile(r.Context(), actorID, p)
	if err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(profile)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	actorID, _ := r.Context().Value(middleware.UserIDKey).(string)
	userID := chi.URLParam(r, "userID")
	var p model.Profile
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	p.UserID = userID
	profile, err := h.svc.UpdateProfile(r.Context(), actorID, p)
	if err != nil {
		writeError(w, err)
		return
	}
	json.NewEncoder(w).Encode(profile)
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	profile, err := h.svc.GetProfile(r.Context(), userID)
	if err != nil {
		writeError(w, err)
		return
	}
	json.NewEncoder(w).Encode(profile)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	size, _ := strconv.Atoi(r.URL.Query().Get("size"))
	result, err := h.svc.ListProfiles(r.Context(), page, size)
	if err != nil {
		writeError(w, err)
		return
	}
	json.NewEncoder(w).Encode(result)
}

func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrForbidden):
		http.Error(w, "forbidden", http.StatusForbidden)
	case errors.Is(err, repository.ErrNotFound):
		http.Error(w, "not found", http.StatusNotFound)
	default:
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}
