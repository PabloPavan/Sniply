package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/PabloPavan/Sniply/internal"
	"github.com/PabloPavan/Sniply/internal/snippets"
)

type SnippetsRepo interface {
	Create(ctx context.Context, s *snippets.Snippet) error
	GetByIDPublicOnly(ctx context.Context, id string) (*snippets.Snippet, error)
	List(ctx context.Context, f snippets.SnippetFilter) ([]*snippets.Snippet, error)
}

type SnippetsHandler struct {
	Repo SnippetsRepo
}

func (h *SnippetsHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req snippets.CreateSnippetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Content = strings.TrimSpace(req.Content)
	req.Language = strings.TrimSpace(req.Language)

	if req.Name == "" || req.Content == "" {
		http.Error(w, "name and content are required", http.StatusBadRequest)
		return
	}
	if req.Language == "" {
		req.Language = "txt"
	}
	if req.Visibility == "" {
		req.Visibility = snippets.VisibilityPublic // para testar no Insomnia
	}

	s := &snippets.Snippet{
		ID:         "snp_" + internal.RandomHex(12),
		Name:       req.Name,
		Content:    req.Content,
		Language:   req.Language,
		Tags:       req.Tags,
		Visibility: req.Visibility,

		// Quando entrar auth, isso vem do token.
		CreatorID: "usr_demo",
	}

	if err := h.Repo.Create(r.Context(), s); err != nil {
		http.Error(w, "failed to create snippet", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(s)
}

func (h *SnippetsHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	id = strings.TrimSpace(id)
	if id == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	s, err := h.Repo.GetByIDPublicOnly(r.Context(), id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(s)
}

func (h *SnippetsHandler) List(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	creator := strings.TrimSpace(r.URL.Query().Get("creator"))
	language := strings.TrimSpace(r.URL.Query().Get("language"))

	limit := 100
	offset := 0
	if l := strings.TrimSpace(r.URL.Query().Get("limit")); l != "" {
		// ignore parse error and keep default
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}
	if o := strings.TrimSpace(r.URL.Query().Get("offset")); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}

	f := snippets.SnippetFilter{
		Query:    q,
		Creator:  creator,
		Language: language,
		Limit:    limit,
		Offset:   offset,
	}

	s, err := h.Repo.List(r.Context(), f)
	if err != nil {
		http.Error(w, "failed to list snippets", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(s)
}
