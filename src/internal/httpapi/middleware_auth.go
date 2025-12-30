package httpapi

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/PabloPavan/sniply_api/internal/apikeys"
	"github.com/PabloPavan/sniply_api/internal/identity"
	"github.com/PabloPavan/sniply_api/internal/session"
)

type APIKeyStore interface {
	GetByTokenHash(ctx context.Context, hash string) (*apikeys.Key, error)
}

func AuthMiddleware(mgr *session.Manager, cookieCfg session.CookieConfig, apiKeyStore APIKeyStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if token := apiKeyFromRequest(r); token != "" {
				if apiKeyStore == nil {
					http.Error(w, "unauthorized", http.StatusUnauthorized)
					return
				}

				key, err := apiKeyStore.GetByTokenHash(r.Context(), apikeys.HashToken(token))
				if err != nil {
					if apikeys.IsNotFound(err) {
						http.Error(w, "unauthorized", http.StatusUnauthorized)
						return
					}
					http.Error(w, "failed to authenticate", http.StatusInternalServerError)
					return
				}

				if !key.Scope.AllowsMethod(r.Method) {
					http.Error(w, "forbidden", http.StatusForbidden)
					return
				}

				ctx := identity.WithUser(r.Context(), key.UserID, key.UserRole)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			if mgr == nil {
				http.Error(w, "auth not configured", http.StatusInternalServerError)
				return
			}

			name := cookieCfg.Name
			if name == "" {
				name = "sniply_session"
			}

			reqCookie, err := r.Cookie(name)
			if err != nil || reqCookie.Value == "" {
				http.Error(w, "missing session", http.StatusUnauthorized)
				return
			}

			sess, err := mgr.Get(r.Context(), reqCookie.Value)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			if requiresCSRFToken(r.Method) {
				token := r.Header.Get("X-CSRF-Token")
				if token == "" || token != sess.CSRFToken {
					http.Error(w, "forbidden", http.StatusForbidden)
					return
				}
			}

			var refreshed bool
			sess, refreshed, err = mgr.Refresh(r.Context(), sess)
			if err != nil {
				if errors.Is(err, session.ErrNotFound) {
					http.Error(w, "unauthorized", http.StatusUnauthorized)
					return
				}
				http.Error(w, "failed to refresh session", http.StatusInternalServerError)
				return
			}

			if refreshed && sess != nil {
				cookieCfg.Write(w, sess.ID, sess.ExpiresAt)
			}

			ctx := identity.WithUser(r.Context(), sess.UserID, sess.Role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func apiKeyFromRequest(r *http.Request) string {
	if v := strings.TrimSpace(r.Header.Get("X-API-Key")); v != "" {
		return v
	}

	auth := strings.TrimSpace(r.Header.Get("Authorization"))
	if auth == "" {
		return ""
	}
	parts := strings.Fields(auth)
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return strings.TrimSpace(parts[1])
	}
	return ""
}

func requiresCSRFToken(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return false
	default:
		return true
	}
}
