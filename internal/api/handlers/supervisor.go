package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	db "github.com/Abtingh/Zeiterfassung/internal/db/sqlc"
	"github.com/google/uuid"
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

// GetStudentsForSupervisorHandler returns all students under this supervisor
func (h *Handler) GetStudentsForSupervisorHandler(w http.ResponseWriter, r *http.Request) {
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

	// Get all students for this supervisor
	students, err := h.Q.GetStudentsForSupervisor(r.Context(), pgtype.Int8{Int64: user.ID, Valid: true})
	if err != nil {
		log.Printf("Error fetching students: %v", err)
		http.Error(w, "Failed to fetch students", http.StatusInternalServerError)
		return
	}

	// Return the students as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(students)
}

func (h *Handler) ApproveSubmissionHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

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

	//?
	var req struct {
		SubmissionID uuid.UUID `json:"submission_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad Request: Invalid JSON", http.StatusBadRequest)
		return
	}

	authorized, err := h.Q.VerifySubmissionBelongsToSupervisor(r.Context(), db.VerifySubmissionBelongsToSupervisorParams{
		ID:           req.SubmissionID,
		SupervisorID: pgtype.Int8{Int64: user.ID, Valid: true},
	})

	if err != nil || !authorized {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	_, err = h.Q.ApproveWeeklySubmission(r.Context(), db.ApproveWeeklySubmissionParams{
		ID:          req.SubmissionID,
		ProcessedBy: pgtype.Int8{Int64: user.ID, Valid: true},
	})
	if err != nil {
		http.Error(w, "Failed to approve", http.StatusInternalServerError)
		return
	}

	// 7. Return success JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "approved"})
}

func (h *Handler) RejectSubmissionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

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

	var req struct {
		SubmissionID uuid.UUID `json:"submission_id"`
		Reason       string    `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad Request: Invalid JSON", http.StatusBadRequest)
		return
	}

	authorized, err := h.Q.VerifySubmissionBelongsToSupervisor(r.Context(), db.VerifySubmissionBelongsToSupervisorParams{
		ID:           req.SubmissionID,
		SupervisorID: pgtype.Int8{Int64: user.ID, Valid: true},
	})

	if err != nil || !authorized {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	_, err = h.Q.RejectWeeklySubmission(r.Context(), db.RejectWeeklySubmissionParams{
		ID:          req.SubmissionID,
		ProcessedBy: pgtype.Int8{Int64: user.ID, Valid: true},
	})

	if err != nil {
		http.Error(w, "Failed to reject", http.StatusInternalServerError)
		return
	}

	// Return success JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "rejected"})
}

// SupervisorZeitGenehmigenHandler serves the Zeit genehmigen page for supervisors
func (h *Handler) SupervisorZeitGenehmigenHandler(w http.ResponseWriter, r *http.Request) {
	// Ensure user is authenticated
	if err := h.Authorize(r); err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get current user and check role
	user, err := h.GetCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Only supervisors can access this page
	if user.Role != db.RoleEnumSupervisor {
		http.Error(w, "Forbidden: Only supervisors can access this", http.StatusForbidden)
		return
	}

	http.ServeFile(w, r, "./public/public-static/teamLeiter_ZeitGenehmigen.html")
}
