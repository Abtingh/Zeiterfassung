package handlers

import (
	"encoding/json"
	"fmt"
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

	log.Printf("GetPendingTimeEntriesHandler: User ID=%d, Email=%s, Role=%s", user.ID, user.Email, user.Role)

	// Check if the user is supervisor
	if user.Role != db.RoleEnumSupervisor {
		http.Error(w, "Forbidden: Only supervisors can access this", http.StatusForbidden)
		return
	}

	// Get all pending submissions for this supervisor
	log.Printf("GetPendingTimeEntriesHandler: Fetching pending submissions for supervisor ID=%d", user.ID)
	submissions, err := h.Q.GetPendingSubmissionsForSupervisor(r.Context(), pgtype.Int8{Int64: user.ID, Valid: true})
	if err != nil {
		log.Printf("Error fetching pending submissions: %v", err)
		http.Error(w, "Failed to fetch pending submissions", http.StatusInternalServerError)
		return
	}

	log.Printf("GetPendingTimeEntriesHandler: Found %d pending submissions", len(submissions))

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

// GetTimeEntriesForSubmissionHandler returns time entries for a specific submission
func (h *Handler) GetTimeEntriesForSubmissionHandler(w http.ResponseWriter, r *http.Request) {
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

	// Get submission_id from query params
	submissionIDStr := r.URL.Query().Get("submission_id")
	if submissionIDStr == "" {
		http.Error(w, "Bad Request: submission_id is required", http.StatusBadRequest)
		return
	}

	submissionID, err := uuid.Parse(submissionIDStr)
	if err != nil {
		http.Error(w, "Bad Request: Invalid submission_id", http.StatusBadRequest)
		return
	}

	// Verify this submission belongs to a student under this supervisor
	authorized, err := h.Q.VerifySubmissionBelongsToSupervisor(r.Context(), db.VerifySubmissionBelongsToSupervisorParams{
		ID:           submissionID,
		SupervisorID: pgtype.Int8{Int64: user.ID, Valid: true},
	})

	if err != nil || !authorized {
		log.Printf("Submission %s not authorized for supervisor %d: %v", submissionID, user.ID, err)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Get time entries for this submission
	entries, err := h.Q.GetTimeEntriesBySubmissionID(r.Context(), submissionID)
	if err != nil {
		log.Printf("Error fetching time entries for submission %s: %v", submissionID, err)
		http.Error(w, "Failed to fetch time entries", http.StatusInternalServerError)
		return
	}

	log.Printf("GetTimeEntriesForSubmissionHandler: Found %d entries for submission %s", len(entries), submissionID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}

// GetStudentWeekStatusHandler returns the submission status for a specific student's week
func (h *Handler) GetStudentWeekStatusHandler(w http.ResponseWriter, r *http.Request) {
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

	// Parse query parameters
	weekStr := r.URL.Query().Get("week")
	yearStr := r.URL.Query().Get("year")
	studentIDStr := r.URL.Query().Get("student_id")

	if weekStr == "" || yearStr == "" || studentIDStr == "" {
		http.Error(w, "Bad Request: week, year, and student_id are required", http.StatusBadRequest)
		return
	}

	var week, year int32
	var studentID int64
	fmt.Sscanf(weekStr, "%d", &week)
	fmt.Sscanf(yearStr, "%d", &year)
	fmt.Sscanf(studentIDStr, "%d", &studentID)

	// Verify the student belongs to this supervisor's team
	students, err := h.Q.GetStudentsForSupervisor(r.Context(), pgtype.Int8{Int64: user.ID, Valid: true})
	if err != nil {
		http.Error(w, "Failed to verify student", http.StatusInternalServerError)
		return
	}

	isAuthorized := false
	for _, s := range students {
		if s.ID == studentID {
			isAuthorized = true
			break
		}
	}

	if !isAuthorized {
		http.Error(w, "Forbidden: Student not in your team", http.StatusForbidden)
		return
	}

	// Get submission for this student and week
	submission, err := h.Q.GetWeeklySubmission(r.Context(), db.GetWeeklySubmissionParams{
		UserID:     studentID,
		WeekNumber: week,
		Year:       year,
	})

	if err != nil {
		// No submission found = open status
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "offen"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": string(submission.Status)})
}
