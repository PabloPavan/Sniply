package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/PabloPavan/Sniply/internal"
	"github.com/PabloPavan/Sniply/internal/users"
)

type UsersRepo interface {
	Create(ctx context.Context, u *users.User) error
	GetByID(ctx context.Context, id string) (*users.User, error)
	List(ctx context.Context, f users.UserFilter) ([]*users.User, error)
	Update(ctx context.Context, u *users.User) error
	Delete(ctx context.Context, id string) error
}

type UsersHandler struct {
	Repo           UsersRepo
	PasswordHasher func(plain string) (string, error)
}

func (h *UsersHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req users.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Password = strings.TrimSpace(req.Password)

	if req.Email == "" || req.Password == "" {
		http.Error(w, "email and password are required", http.StatusBadRequest)
		return
	}
	// Validação mínima. Se quiser, faça validação mais forte (regex).
	if !strings.Contains(req.Email, "@") {
		http.Error(w, "invalid email", http.StatusBadRequest)
		return
	}

	hasher := h.PasswordHasher
	if hasher == nil {
		hasher = defaultPasswordHasher
	}

	hash, err := hasher(req.Password)
	if err != nil {
		http.Error(w, "failed to process password", http.StatusInternalServerError)
		return
	}

	u := &users.User{
		ID:           "usr_" + internal.RandomHex(12),
		Email:        req.Email,
		PasswordHash: hash,
	}

	if err := h.Repo.Create(r.Context(), u); err != nil {
		if users.IsUniqueViolationEmail(err) {
			http.Error(w, "email already exists", http.StatusConflict)
			return
		}

		http.Error(w, "failed to create user", http.StatusInternalServerError)
		return
	}

	resp := users.UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *UsersHandler) List(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))

	limit := 100
	offset := 0

	if l := strings.TrimSpace(r.URL.Query().Get("limit")); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}
	if o := strings.TrimSpace(r.URL.Query().Get("offset")); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}

	f := users.UserFilter{
		Query:  q,
		Limit:  limit,
		Offset: offset,
	}

	list, err := h.Repo.List(r.Context(), f)
	if err != nil {
		http.Error(w, "failed to list users", http.StatusInternalServerError)
		return
	}

	resp := make([]users.UserResponse, 0, len(list))
	for _, u := range list {
		resp = append(resp, users.UserResponse{
			ID:        u.ID,
			Email:     u.Email,
			CreatedAt: u.CreatedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *UsersHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(chi.URLParam(r, "id"))
	if id == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	var req users.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Password = strings.TrimSpace(req.Password)

	if req.Email == "" && req.Password == "" {
		http.Error(w, "email or password must be provided", http.StatusBadRequest)
		return
	}
	if req.Email != "" && !strings.Contains(req.Email, "@") {
		http.Error(w, "invalid email", http.StatusBadRequest)
		return
	}

	cur, err := h.Repo.GetByID(r.Context(), id)
	if err != nil {
		if users.IsNotFound(err) {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to load user", http.StatusInternalServerError)
		return
	}

	if req.Email != "" {
		cur.Email = req.Email
	}
	if req.Password != "" {
		hasher := h.PasswordHasher
		if hasher == nil {
			hasher = defaultPasswordHasher
		}
		hash, err := hasher(req.Password)
		if err != nil {
			http.Error(w, "failed to process password", http.StatusInternalServerError)
			return
		}
		cur.PasswordHash = hash
	}

	if err := h.Repo.Update(r.Context(), cur); err != nil {
		if users.IsUniqueViolationEmail(err) {
			http.Error(w, "email already exists", http.StatusConflict)
			return
		}
		if users.IsNotFound(err) {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to update user", http.StatusInternalServerError)
		return
	}

	resp := users.UserResponse{
		ID:        cur.ID,
		Email:     cur.Email,
		CreatedAt: cur.CreatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *UsersHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(chi.URLParam(r, "id"))
	if id == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	if err := h.Repo.Delete(r.Context(), id); err != nil {
		if users.IsNotFound(err) {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to update user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
