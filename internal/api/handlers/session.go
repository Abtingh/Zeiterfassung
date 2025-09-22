package handlers

import (
	"errors"
	"log"
	"net/http"

	db "github.com/Abtingh/Zeiterfassung/internal/db/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

var ErrUnauthorized = errors.New("unauthorized")

// Authorize checks if the incoming HTTP request is authorized by validating the session token
// and, for non-GET requests, the CSRF token. It retrieves the session token from the request's
// cookies and looks up the associated user. For non-GET requests, it also validates the CSRF
// token from the cookies against the user's stored CSRF token. Returns ErrUnauthorized if any check
// fails, otherwise returns nil to indicate successful authorization.
func (h *Handler) Authorize(r *http.Request) error {
	// Get session token from cookie
	sessionCookie, err := r.Cookie("session_token")
	if err != nil || sessionCookie.Value == "" {
		log.Printf("Authorize failed: no session token cookie. Error: %v", err)
		return ErrUnauthorized
	}

	log.Printf("Authorize: Found session token: %s", sessionCookie.Value[:10]+"...")

	// Look up user by session token
	ctx := r.Context()
	user, err := h.Q.GetUserBySessionToken(ctx, pgtype.Text{String: sessionCookie.Value, Valid: true})
	if err != nil {
		log.Printf("Authorize failed: GetUserBySessionToken error: %v", err)
		return ErrUnauthorized
	}

	log.Printf("Authorize: Found user %s for session token", user.Email)

	// For non-GET requests, validate CSRF token
	if r.Method != http.MethodGet {
		csrfCookie, err := r.Cookie("csrf_token")
		if err != nil || csrfCookie.Value == "" {
			return ErrUnauthorized
		}

		if !user.CsrfToken.Valid || csrfCookie.Value != user.CsrfToken.String {
			return ErrUnauthorized
		}
	}

	return nil // authorized
}

// GetCurrentUser retrieves the current user from the session token in the request
func (h *Handler) GetCurrentUser(r *http.Request) (*db.User, error) {
	// Get session token from cookie
	sessionCookie, err := r.Cookie("session_token")
	if err != nil || sessionCookie.Value == "" {
		log.Printf("GetCurrentUser failed: no session token cookie. Error: %v", err)
		return nil, ErrUnauthorized
	}

	// Look up user by session token
	ctx := r.Context()
	user, err := h.Q.GetUserBySessionToken(ctx, pgtype.Text{String: sessionCookie.Value, Valid: true})
	if err != nil {
		log.Printf("GetCurrentUser failed: GetUserBySessionToken error: %v", err)
		return nil, ErrUnauthorized
	}

	log.Printf("GetCurrentUser: Found user %s", user.Email)
	return &user, nil
}
