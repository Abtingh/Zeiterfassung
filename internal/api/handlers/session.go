package handlers

import (
	"database/sql"
	"errors"
	"net/http"
)

var AuthError = errors.New("unauthorized")

// Authorize validates the requester using data stored in the database.
//
// It replaces the previous in‑memory `users[email]` lookup with a DB query.
// The function is now a *method* on Handler so it can access h.Q.
//
// Workflow:
//  1. Extract email from form data.
//  2. Fetch user row with `GetUserByEmail`.
//  3. Compare session_token cookie and csrf header with DB values.
//
// Any mismatch → AuthError.
func (h *Handler) Authorize(r *http.Request) error {
	email := r.FormValue("email")
	if email == "" {
		return AuthError
	}

	// ── Get user from DB ─────────────────────────────────────────────
	ctx := r.Context()
	user, err := h.Q.GetUserByEmail(ctx, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return AuthError
		}
		return err // unexpected DB error
	}

	// ── Validate session token ───────────────────────────────────────
	stCookie, err := r.Cookie("session_token")
	if err != nil || stCookie.Value == "" {
		return AuthError
	}
	if !user.SessionToken.Valid || stCookie.Value != user.SessionToken.String {
		return AuthError
	}

	// ── Validate CSRF token ──────────────────────────────────────────
	csrfHeader := r.Header.Get("X-CSRF-Token")
	if csrfHeader == "" {
		return AuthError
	}
	if !user.CsrfToken.Valid || csrfHeader != user.CsrfToken.String {
		return AuthError
	}

	return nil // authorized
}
