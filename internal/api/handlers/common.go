package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgtype"
)

func (h *Handler) HomeHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("→ %s %s", r.Method, r.URL.Path)
	log.Printf("HomeHandler: Cookies received: %v", r.Header.Get("Cookie"))

	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.WriteHeader(http.StatusNoContent)
		return
	} // Ensure user is authenticated
	if err := h.Authorize(r); err != nil {
		log.Printf("Authorization failed: %v", err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get current user from session
	user, err := h.GetCurrentUser(r)
	if err != nil {
		log.Printf("GetCurrentUser failed: %v", err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	log.Printf("Serving home page for user %s with role: '%s' (length: %d)", user.Email, user.Role, len(user.Role))

	// Serve the correct home page based on user's role
	switch user.Role {
	case "supervisor":
		log.Printf("Serving teamleiter_home.html for supervisor")
		http.ServeFile(w, r, "./public/public-static/teamleiter_home.html")
	case "admin":
		log.Printf("Serving admin_home.html for admin")
		http.ServeFile(w, r, "./public/public-static/admin_home.html")
	case "accounting":
		log.Printf("Serving buch_home.html for accounting")
		http.ServeFile(w, r, "./public/public-static/buch_home.html")
	default: // student or any other role
		log.Printf("Serving student_home.html for role: '%s'", user.Role)
		http.ServeFile(w, r, "./public/public-static/student_home.html")
	}
}

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
		ID        int64  `json:"id"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
		Role      string `json:"role"`
		TeamID    *int64 `json:"team_id,omitempty"`
		StartDate string `json:"start_date,omitempty"`
	}

	response := UserResponse{
		ID:        user.ID,
		FirstName: user.FirstName.String,
		LastName:  user.LastName.String,
		Email:     user.Email,
		Role:      string(user.Role),
		TeamID:    nil,
		StartDate: "",
	}

	if user.TeamID.Valid {
		teamID := user.TeamID.Int64
		response.TeamID = &teamID
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
