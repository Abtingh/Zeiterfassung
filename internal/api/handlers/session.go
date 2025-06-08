package handlers

import (
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5/pgtype"
)

var AuthError = errors.New("unauthorized")

func (h *Handler) Authorize(r *http.Request) error {
	// Get session token from cookie
	sessionCookie, err := r.Cookie("session_token")
	if err != nil || sessionCookie.Value == "" {
		return AuthError
	}

	// Look up user by session token
	ctx := r.Context()
	user, err := h.Q.GetUserBySessionToken(ctx, pgtype.Text{String: sessionCookie.Value, Valid: true})
	if err != nil {
		return AuthError
	}

	// For non-GET requests, validate CSRF token
	if r.Method != http.MethodGet {
		csrfCookie, err := r.Cookie("csrf_token")
		if err != nil || csrfCookie.Value == "" {
			return AuthError
		}

		if !user.CsrfToken.Valid || csrfCookie.Value != user.CsrfToken.String {
			return AuthError
		}
	}

	return nil // authorized
}
