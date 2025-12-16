package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	db "github.com/Abtingh/Zeiterfassung/internal/db/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// GetApprovedSubmissionsHandler fetches all approved time entries for accounting to process
func (h *Handler) GetApprovedSubmissionsHandler(w http.ResponseWriter, r *http.Request) {
	user, err := h.GetCurrentUser(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if the user is accounting
	if user.Role != db.RoleEnumAccounting {
		http.Error(w, "Forbidden: Only accounting can access this", http.StatusForbidden)
		return
	}

	// Get all approved submissions
	submissions, err := h.Q.GetApprovedSubmissions(r.Context())
	if err != nil {
		log.Printf("Error fetching approved submissions: %v", err)
		http.Error(w, "Failed to fetch approved submissions", http.StatusInternalServerError)
		return
	}

	// Return the submissions as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(submissions)
}

// GetAllStudentsHandler returns all students in the system for accounting dropdown
func (h *Handler) GetAllStudentsHandler(w http.ResponseWriter, r *http.Request) {
	user, err := h.GetCurrentUser(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if the user is accounting
	if user.Role != db.RoleEnumAccounting {
		http.Error(w, "Forbidden: Only accounting can access this", http.StatusForbidden)
		return
	}

	// Get all students
	students, err := h.Q.GetAllStudents(r.Context())
	if err != nil {
		log.Printf("Error fetching students: %v", err)
		http.Error(w, "Failed to fetch students", http.StatusInternalServerError)
		return
	}

	// Return the students as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(students)
}

// GetApprovedSubmissionsForStudentHandler returns approved submissions for a specific student
func (h *Handler) GetApprovedSubmissionsForStudentHandler(w http.ResponseWriter, r *http.Request) {
	user, err := h.GetCurrentUser(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if the user is accounting
	if user.Role != db.RoleEnumAccounting {
		http.Error(w, "Forbidden: Only accounting can access this", http.StatusForbidden)
		return
	}

	// Get student ID from query params
	studentIDStr := r.URL.Query().Get("student_id")
	if studentIDStr == "" {
		http.Error(w, "Missing student_id parameter", http.StatusBadRequest)
		return
	}

	var studentID int64
	if _, err := json.Number(studentIDStr).Int64(); err != nil {
		http.Error(w, "Invalid student_id", http.StatusBadRequest)
		return
	}
	studentID, _ = json.Number(studentIDStr).Int64()

	// Get approved submissions for the student
	submissions, err := h.Q.GetApprovedSubmissionsForStudent(r.Context(), studentID)
	if err != nil {
		log.Printf("Error fetching submissions for student %d: %v", studentID, err)
		http.Error(w, "Failed to fetch submissions", http.StatusInternalServerError)
		return
	}

	// Return the submissions as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(submissions)
}

// GetAllSubmissionsForStudentHandler returns ALL submissions for a specific student (any status)
func (h *Handler) GetAllSubmissionsForStudentHandler(w http.ResponseWriter, r *http.Request) {
	user, err := h.GetCurrentUser(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if the user is accounting
	if user.Role != db.RoleEnumAccounting {
		http.Error(w, "Forbidden: Only accounting can access this", http.StatusForbidden)
		return
	}

	// Get student ID from query params
	studentIDStr := r.URL.Query().Get("student_id")
	if studentIDStr == "" {
		http.Error(w, "Missing student_id parameter", http.StatusBadRequest)
		return
	}

	var studentID int64
	if _, err := json.Number(studentIDStr).Int64(); err != nil {
		http.Error(w, "Invalid student_id", http.StatusBadRequest)
		return
	}
	studentID, _ = json.Number(studentIDStr).Int64()

	// Get ALL submissions for the student (any status)
	submissions, err := h.Q.GetSubmissionsForStudent(r.Context(), studentID)
	if err != nil {
		log.Printf("Error fetching all submissions for student %d: %v", studentID, err)
		http.Error(w, "Failed to fetch submissions", http.StatusInternalServerError)
		return
	}

	// Return the submissions as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(submissions)
}

// MarkSubmissionAsProcessedHandler marks a submission as 'erledigt' (completed) by accounting
func (h *Handler) MarkSubmissionAsProcessedHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := h.GetCurrentUser(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if the user is accounting
	if user.Role != db.RoleEnumAccounting {
		http.Error(w, "Forbidden: Only accounting can access this", http.StatusForbidden)
		return
	}

	var req struct {
		SubmissionID uuid.UUID `json:"submission_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad Request: Invalid JSON", http.StatusBadRequest)
		return
	}

	// Mark submission as processed
	result, err := h.Q.MarkSubmissionAsProcessed(r.Context(), db.MarkSubmissionAsProcessedParams{
		ID:          req.SubmissionID,
		ProcessedBy: pgtype.Int8{Int64: user.ID, Valid: true},
	})
	if err != nil {
		log.Printf("Error marking submission as processed: %v", err)
		http.Error(w, "Failed to mark submission as processed", http.StatusInternalServerError)
		return
	}

	// Return success JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "erledigt",
		"result": result,
	})
}

// GetProcessedSubmissionsHandler fetches all processed (erledigt) submissions
func (h *Handler) GetProcessedSubmissionsHandler(w http.ResponseWriter, r *http.Request) {
	user, err := h.GetCurrentUser(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if the user is accounting
	if user.Role != db.RoleEnumAccounting {
		http.Error(w, "Forbidden: Only accounting can access this", http.StatusForbidden)
		return
	}

	// Get all processed submissions
	submissions, err := h.Q.GetProcessedSubmissions(r.Context())
	if err != nil {
		log.Printf("Error fetching processed submissions: %v", err)
		http.Error(w, "Failed to fetch processed submissions", http.StatusInternalServerError)
		return
	}

	// Return the submissions as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(submissions)
}

// AccountingZeitenUebersichtHandler serves the Zeiten Übersicht page for accounting
func (h *Handler) AccountingZeitenUebersichtHandler(w http.ResponseWriter, r *http.Request) {
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

	// Only accounting can access this page
	if user.Role != db.RoleEnumAccounting {
		http.Error(w, "Forbidden: Only accounting can access this", http.StatusForbidden)
		return
	}

	http.ServeFile(w, r, "./public/public-static/buch_ZeitenUebersicht.html")
}
