package apikeys

import (
	"testing"

	"github.com/jackc/pgx/v5"
)

func TestTokenPrefix(t *testing.T) {
	if got := TokenPrefix("short"); got != "short" {
		t.Fatalf("TokenPrefix short = %q", got)
	}
	if got := TokenPrefix("12345678"); got != "12345678" {
		t.Fatalf("TokenPrefix exact = %q", got)
	}
	if got := TokenPrefix("123456789"); got != "12345678" {
		t.Fatalf("TokenPrefix long = %q", got)
	}
}

func TestHashToken(t *testing.T) {
	if got := HashToken(" token "); got != HashToken("token") {
		t.Fatalf("HashToken should trim input")
	}
}

func TestScopeAllowsMethod(t *testing.T) {
	if !ScopeRead.AllowsMethod("GET") {
		t.Fatal("read should allow GET")
	}
	if ScopeRead.AllowsMethod("POST") {
		t.Fatal("read should not allow POST")
	}
	if ScopeWrite.AllowsMethod("GET") {
		t.Fatal("write should not allow GET")
	}
	if !ScopeWrite.AllowsMethod("POST") {
		t.Fatal("write should allow POST")
	}
	if !ScopeReadWrite.AllowsMethod("GET") || !ScopeReadWrite.AllowsMethod("POST") {
		t.Fatal("read_write should allow GET and POST")
	}
}

func TestScopeValid(t *testing.T) {
	if !ScopeRead.Valid() || !ScopeWrite.Valid() || !ScopeReadWrite.Valid() {
		t.Fatal("expected valid scopes to be valid")
	}
	if Scope("bad").Valid() {
		t.Fatal("expected invalid scope to be invalid")
	}
}

func TestIsNotFound(t *testing.T) {
	if !IsNotFound(ErrNotFound) {
		t.Fatal("expected ErrNotFound to match")
	}
	if !IsNotFound(pgx.ErrNoRows) {
		t.Fatal("expected pgx.ErrNoRows to match")
	}
}
