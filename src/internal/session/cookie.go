package session

import (
	"net/http"
	"strings"
	"time"
)

const (
	DefaultSessionCookieName = "sniply_session"
	DefaultCSRFCookieName    = "sniply_csrf"
)

type CookieConfig struct {
	Name     string
	Path     string
	Domain   string
	Secure   bool
	SameSite http.SameSite
}

func (c CookieConfig) Write(w http.ResponseWriter, value string, expiresAt time.Time) {
	path := c.Path
	if path == "" {
		path = "/"
	}
	name := SessionCookieName(c.Name)

	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     path,
		Domain:   c.Domain,
		Expires:  expiresAt,
		MaxAge:   int(time.Until(expiresAt).Seconds()),
		Secure:   c.Secure,
		HttpOnly: true,
		SameSite: c.SameSite,
	})
}

func (c CookieConfig) Clear(w http.ResponseWriter) {
	path := c.Path
	if path == "" {
		path = "/"
	}
	name := SessionCookieName(c.Name)

	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     path,
		Domain:   c.Domain,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		Secure:   c.Secure,
		HttpOnly: true,
		SameSite: c.SameSite,
	})
}

type CSRFCookieConfig struct {
	Name     string
	Path     string
	Domain   string
	Secure   bool
	SameSite http.SameSite
}

func (c CSRFCookieConfig) Write(w http.ResponseWriter, value string, expiresAt time.Time) {
	path := c.Path
	if path == "" {
		path = "/"
	}
	name := CSRFCookieName(c.Name)

	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     path,
		Domain:   c.Domain,
		Expires:  expiresAt,
		MaxAge:   int(time.Until(expiresAt).Seconds()),
		Secure:   c.Secure,
		HttpOnly: false,
		SameSite: c.SameSite,
	})
}

func (c CSRFCookieConfig) Clear(w http.ResponseWriter) {
	path := c.Path
	if path == "" {
		path = "/"
	}
	name := CSRFCookieName(c.Name)

	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     path,
		Domain:   c.Domain,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		Secure:   c.Secure,
		HttpOnly: false,
		SameSite: c.SameSite,
	})
}

func SessionCookieName(name string) string {
	if strings.TrimSpace(name) == "" {
		return DefaultSessionCookieName
	}
	return name
}

func CSRFCookieName(name string) string {
	if strings.TrimSpace(name) == "" {
		return DefaultCSRFCookieName
	}
	return name
}
