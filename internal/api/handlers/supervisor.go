package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	db "github.com/Abtingh/Zeiterfassung/internal/db/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

// GetPendingTimeEntriesHandler fetches all submitted time entries for students under this supervisor
func (h *Handler) GetPendingTimeEntriesHandler(w http.ResponseWriter, r *http.Request) {
	user, err := h.GetCurrentUser(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if the user is supervisor
	if user.Role != db.RoleEnumSupervisor {
		http.Error(w, "Forbidden: Only supervisors can access this", http.StatusForbidden)
		return
	}

	// Get all pending submissions for this supervisor
	submissions, err := h.Q.GetPendingSubmissionsForSupervisor(r.Context(), pgtype.Int8{Int64: user.ID, Valid: true})
	if err != nil {
		log.Printf("Error fetching pending submissions: %v", err)
		http.Error(w, "Failed to fetch pending submissions", http.StatusInternalServerError)
		return
	}

	// Return the submissions as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(submissions)
}
