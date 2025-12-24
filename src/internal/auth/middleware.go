package auth

import (
	"net/http"
	"strings"
)

func Middleware(svc *Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := r.Header.Get("Authorization")
			if h == "" {
				http.Error(w, "missing authorization", http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(h, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				http.Error(w, "invalid authorization", http.StatusUnauthorized)
				return
			}

			claims, err := svc.ValidateAccessToken(strings.TrimSpace(parts[1]))
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			ctx := WithUser(r.Context(), claims.Subject, claims.Role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
