package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"matchmaking-service/auth/middleware"
	"matchmaking-service/matches/model"
	"matchmaking-service/matches/service"
)

type Handler struct{ svc *service.Service }

func NewHandler(svc *service.Service) *Handler { return &Handler{svc: svc} }

type likeRequest struct {
	LikedUserID string `json:"liked_user_id"`
}

func (h *Handler) Browse(w http.ResponseWriter, r *http.Request) {
	actorID, _ := r.Context().Value(middleware.UserIDKey).(string)
	filter := model.BrowseFilter{
		MinAge:          atoi(r.URL.Query().Get("min_age")),
		MaxAge:          atoi(r.URL.Query().Get("max_age")),
		Gender:          r.URL.Query().Get("gender"),
		MinIncome:       atoi(r.URL.Query().Get("min_income")),
		MaxIncome:       atoi(r.URL.Query().Get("max_income")),
		Location:        r.URL.Query().Get("location"),
		ResidencyStatus: r.URL.Query().Get("residency_status"),
		Page:            atoi(r.URL.Query().Get("page")),
		Size:            atoi(r.URL.Query().Get("size")),
	}
	result, err := h.svc.Browse(r.Context(), actorID, filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(result)
}

func (h *Handler) LikeProfile(w http.ResponseWriter, r *http.Request) {
	actorID, _ := r.Context().Value(middleware.UserIDKey).(string)
	var req likeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	resp, err := h.svc.Like(r.Context(), actorID, req.LikedUserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) Relationships(w http.ResponseWriter, r *http.Request) {
	actorID, _ := r.Context().Value(middleware.UserIDKey).(string)
	resp, err := h.svc.Relationships(r.Context(), actorID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(resp)
}

func atoi(v string) int {
	i, _ := strconv.Atoi(v)
	return i
}
