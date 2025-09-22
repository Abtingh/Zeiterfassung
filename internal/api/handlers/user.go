package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgtype"
)

func (h *Handler) MeHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("→ %s %s", r.Method, r.URL.Path)
	log.Printf("MeHandler: Cookies received: %v", r.Header.Get("Cookie"))

	// Set CORS headers for all requests
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	// Handle CORS preflight
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Only allow GET requests
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	// Validate session authorization
	if err := h.Authorize(r); err != nil {
		log.Printf("MeHandler: Authorization failed: %v", err)
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	log.Printf("MeHandler: Authorization successful")

	// Get session token from cookie
	sessionCookie, err := r.Cookie("session_token")
	if err != nil || sessionCookie.Value == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return
	}

	// Get user data from database using session token
	ctx := r.Context()
	user, err := h.Q.GetUserBySessionToken(ctx, pgtype.Text{String: sessionCookie.Value, Valid: true})
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error":"User not found"}`, http.StatusNotFound)
		return
	}

	// Prepare response data (exclude sensitive fields)
	type UserResponse struct {
		ID           int64  `json:"id"`
		FirstName    string `json:"first_name"`
		LastName     string `json:"last_name"`
		Email        string `json:"email"`
		Role         string `json:"role"`
		SupervisorID *int64 `json:"supervisor_id,omitempty"`
		StartDate    string `json:"start_date,omitempty"`
	}

	response := UserResponse{
		ID:           user.ID,
		FirstName:    user.FirstName.String,
		LastName:     user.LastName.String,
		Email:        user.Email,
		Role:         string(user.Role),
		SupervisorID: nil,
		StartDate:    "",
	}

	if user.SupervisorID.Valid {
		supervisorID := user.SupervisorID.Int64
		response.SupervisorID = &supervisorID
	}
	if user.StartDate.Valid {
		response.StartDate = user.StartDate.Time.Format("2006-01-02")
	}

	// Set response headers and send JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, `{"error":"failed to encode response"}`, http.StatusInternalServerError)
		return
	}

}
