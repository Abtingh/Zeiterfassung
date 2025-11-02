package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	db "github.com/Abtingh/Zeiterfassung/internal/db/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TimeEntryInput struct {
	Date      string  `json:"date" binding:"required"`
	StartTime string  `json:"startTime"`
	EndTime   string  `json:"endTime"`
	BreakMin  float64 `json:"breakMin"`
	DurationH float64 `json:"durationH"`
	Note      string  `json:"note"`
}

type SubmitWeekRequest struct {
	WeekNumber int              `json:"weekNumber" binding:"required"`
	Year       int              `json:"year" binding:"required"`
	Entries    []TimeEntryInput `json:"entries" binding:"required"`
}

type WeekResponse struct {
	Entries []db.TimeEntry `json:"entries"`
	Status  string         `json:"status"`
}

// SubmitWeeklyTimeEntriesHandler handles submission of a full week's time entries
func (h *Handler) SubmitWeeklyTimeEntriesHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("→ %s %s", r.Method, r.URL.Path)

	// Handle CORS preflight
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Ensure user is authenticated
	if err := h.Authorize(r); err != nil {
		log.Printf("Authorization failed: %v", err)
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// Get current user
	user, err := h.GetCurrentUser(r)
	if err != nil {
		log.Printf("GetCurrentUser failed: %v", err)
		http.Error(w, `{"error":"Failed to get user"}`, http.StatusInternalServerError)
		return
	}

	// Parse JSON request body
	var req SubmitWeekRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("JSON decode error: %v", err)
		http.Error(w, fmt.Sprintf(`{"error":"Invalid request body: %s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	log.Printf("Submitting week %d/%d for user %s with %d entries", req.WeekNumber, req.Year, user.Email, len(req.Entries))

	ctx := r.Context()

	// Get the database pool and start transaction
	pool, ok := h.DB.(*pgxpool.Pool)
	if !ok {
		log.Printf("Failed to cast DB to *pgxpool.Pool")
		http.Error(w, `{"error":"Database connection error"}`, http.StatusInternalServerError)
		return
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		log.Printf("Failed to start transaction: %v", err)
		http.Error(w, `{"error":"Failed to start transaction"}`, http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(ctx)

	qtx := h.Q.WithTx(tx)

	// Check if submission already exists
	existingSubmission, err := qtx.GetWeeklySubmission(ctx, db.GetWeeklySubmissionParams{
		UserID:     user.ID,
		WeekNumber: int32(req.WeekNumber),
		Year:       int32(req.Year),
	})

	var submissionID uuid.UUID
	if err != nil {
		// If no rows found, create new submission
		if err.Error() == "no rows in result set" || err == sql.ErrNoRows {
			// Create new submission
			log.Printf("Creating new weekly submission")
			submission, err := qtx.CreateWeeklySubmission(ctx, db.CreateWeeklySubmissionParams{
				UserID:     user.ID,
				WeekNumber: int32(req.WeekNumber),
				Year:       int32(req.Year),
			})
			if err != nil {
				log.Printf("Failed to create weekly submission: %v", err)
				http.Error(w, `{"error":"Failed to create weekly submission"}`, http.StatusInternalServerError)
				return
			}
			submissionID = submission.ID
		} else {
			// Other database error
			log.Printf("Database error checking submission: %v", err)
			http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
			return
		}
	} else {
		// Use existing submission
		submissionID = existingSubmission.ID
		log.Printf("Found existing submission ID: %d with status: %s", submissionID, existingSubmission.Status)

		// Check if status allows editing
		if existingSubmission.Status == "bestaetigt" || existingSubmission.Status == "erledigt" {
			log.Printf("Cannot edit submission with status: %s", existingSubmission.Status)
			http.Error(w, `{"error":"Cannot edit confirmed submissions"}`, http.StatusForbidden)
			return
		}

		// Delete existing entries
		if err := qtx.DeleteTimeEntriesBySubmission(ctx, submissionID); err != nil {
			log.Printf("Failed to delete old entries: %v", err)
			http.Error(w, `{"error":"Failed to delete old entries"}`, http.StatusInternalServerError)
			return
		}
	}

	// Create time entries
	for i, entry := range req.Entries {
		// Skip entries with no time data (empty days or 00:00-00:00)
		if (entry.StartTime == "" && entry.EndTime == "") ||
			(entry.StartTime == "00:00" && entry.EndTime == "00:00") {
			log.Printf("Skipping empty entry %d", i)
			continue
		}

		entryDate, err := time.Parse("2006-01-02", entry.Date)
		if err != nil {
			log.Printf("Invalid date format for entry %d: %s", i, entry.Date)
			http.Error(w, fmt.Sprintf(`{"error":"Invalid date format: %s"}`, entry.Date), http.StatusBadRequest)
			return
		}

		var startTime, endTime pgtype.Time
		if entry.StartTime != "" {
			st, err := time.Parse("15:04", entry.StartTime)
			if err == nil {
				microseconds := int64(st.Hour())*3600000000 + int64(st.Minute())*60000000
				startTime = pgtype.Time{Microseconds: microseconds, Valid: true}
			}
		}
		if entry.EndTime != "" {
			et, err := time.Parse("15:04", entry.EndTime)
			if err == nil {
				microseconds := int64(et.Hour())*3600000000 + int64(et.Minute())*60000000
				endTime = pgtype.Time{Microseconds: microseconds, Valid: true}
			}
		}

		// Convert break minutes to integer
		breakMinutes := int32(entry.BreakMin)

		// Convert duration to pgtype.Numeric
		var durationH pgtype.Numeric
		durationH.Valid = true // Always mark as valid
		if err := durationH.Scan(fmt.Sprintf("%.2f", entry.DurationH)); err != nil {
			log.Printf("Failed to convert duration: %v, using 0", err)
			durationH.Scan("0.00")
		}

		_, err = qtx.CreateTimeEntry(ctx, db.CreateTimeEntryParams{
			SubmissionID: submissionID,
			EntryDate:    pgtype.Date{Time: entryDate, Valid: true},
			StartTime:    startTime,
			EndTime:      endTime,
			BreakMin:     pgtype.Int4{Int32: breakMinutes, Valid: true},
			DurationH:    durationH,
			Note:         pgtype.Text{String: entry.Note, Valid: entry.Note != ""},
		})
		if err != nil {
			log.Printf("Failed to create time entry %d: %v", i, err)
			http.Error(w, `{"error":"Failed to create time entry"}`, http.StatusInternalServerError)
			return
		}
	}

	// Update status to "gesendet" (sent)
	_, err = qtx.UpdateWeeklySubmissionStatus(ctx, db.UpdateWeeklySubmissionStatusParams{
		ID:     submissionID,
		Status: db.SubmissionStatusEnumGesendet,
	})
	if err != nil {
		log.Printf("Failed to update status: %v", err)
		http.Error(w, `{"error":"Failed to update status"}`, http.StatusInternalServerError)
		return
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		http.Error(w, `{"error":"Failed to commit transaction"}`, http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully submitted week %d/%d for user %s", req.WeekNumber, req.Year, user.Email)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":      "Time entries submitted successfully",
		"submissionId": submissionID,
	})
}

// GetWeeklyTimeEntriesHandler retrieves time entries for a specific week
func (h *Handler) GetWeeklyTimeEntriesHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("→ %s %s", r.Method, r.URL.Path)

	// Handle CORS preflight
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Only allow GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Ensure user is authenticated
	if err := h.Authorize(r); err != nil {
		log.Printf("Authorization failed: %v", err)
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// Get current user
	user, err := h.GetCurrentUser(r)
	if err != nil {
		log.Printf("GetCurrentUser failed: %v", err)
		http.Error(w, `{"error":"Failed to get user"}`, http.StatusInternalServerError)
		return
	}

	// Get query parameters
	weekStr := r.URL.Query().Get("week")
	yearStr := r.URL.Query().Get("year")

	if weekStr == "" || yearStr == "" {
		http.Error(w, `{"error":"Week number and year are required"}`, http.StatusBadRequest)
		return
	}

	weekNum, err := strconv.Atoi(weekStr)
	if err != nil {
		http.Error(w, `{"error":"Invalid week number"}`, http.StatusBadRequest)
		return
	}

	yearNum, err := strconv.Atoi(yearStr)
	if err != nil {
		http.Error(w, `{"error":"Invalid year"}`, http.StatusBadRequest)
		return
	}

	log.Printf("Fetching week %d/%d for user %s", weekNum, yearNum, user.Email)

	ctx := r.Context()

	// Get weekly submission
	submission, err := h.Q.GetWeeklySubmission(ctx, db.GetWeeklySubmissionParams{
		UserID:     user.ID,
		WeekNumber: int32(weekNum),
		Year:       int32(yearNum),
	})

	if err != nil && (err.Error() == "no rows in result set" || err == sql.ErrNoRows) {
		// No submission exists yet, return empty data
		log.Printf("No submission found for week %d/%d", weekNum, yearNum)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(WeekResponse{
			Entries: []db.TimeEntry{},
			Status:  "offen",
		})
		return
	} else if err != nil {
		log.Printf("Database error: %v", err)
		http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
		return
	}

	// Get time entries for this submission
	entries, err := h.Q.GetTimeEntriesBySubmission(ctx, submission.ID)
	if err != nil {
		log.Printf("Failed to fetch entries: %v", err)
		http.Error(w, `{"error":"Failed to fetch entries"}`, http.StatusInternalServerError)
		return
	}

	log.Printf("Found %d entries with status: %s", len(entries), submission.Status)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(WeekResponse{
		Entries: entries,
		Status:  string(submission.Status),
	})
}
