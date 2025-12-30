package apikeys

import "time"

type Scope string

const (
	ScopeRead      Scope = "read"
	ScopeWrite     Scope = "write"
	ScopeReadWrite Scope = "read_write"
)

func (s Scope) Valid() bool {
	switch s {
	case ScopeRead, ScopeWrite, ScopeReadWrite:
		return true
	default:
		return false
	}
}

func (s Scope) AllowsMethod(method string) bool {
	switch method {
	case "GET", "HEAD", "OPTIONS":
		return s == ScopeRead || s == ScopeReadWrite
	default:
		return s == ScopeWrite || s == ScopeReadWrite
	}
}

type Key struct {
	ID          string     `json:"id"`
	UserID      string     `json:"user_id"`
	UserRole    string     `json:"user_role,omitempty"`
	Name        string     `json:"name"`
	Scope       Scope      `json:"scope"`
	TokenHash   string     `json:"-"`
	TokenPrefix string     `json:"token_prefix"`
	CreatedAt   time.Time  `json:"created_at"`
	RevokedAt   *time.Time `json:"revoked_at,omitempty"`
}
