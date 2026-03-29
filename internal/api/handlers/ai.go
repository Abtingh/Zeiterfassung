package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// ParsedTimeEntry represents the AI-parsed time entry
type ParsedTimeEntry struct {
	Project         *string `json:"project"`
	Activity        *string `json:"activity"`
	DurationMinutes int     `json:"durationMinutes"`
	Date            string  `json:"date"`
	Description     string  `json:"description"`
	StartTime       *string `json:"startTime"`
	EndTime         *string `json:"endTime"`
	Confidence      int     `json:"confidence"` // AI confidence score 0-100
}

// AIParseRequest is the request body for the AI parse endpoint
type AIParseRequest struct {
	Text string `json:"text"`
}

// AIParseResponse is the response from the AI parse endpoint
type AIParseResponse struct {
	Success bool             `json:"success"`
	Data    *ParsedTimeEntry `json:"data,omitempty"`
	Error   string           `json:"error,omitempty"`
}

// AIParseTimeEntryHandler handles natural language time entry parsing via n8n
func (h *Handler) AIParseTimeEntryHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("→ %s %s", r.Method, r.URL.Path)

	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	// Handle CORS preflight
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Only allow POST requests
	if r.Method != http.MethodPost {
		json.NewEncoder(w).Encode(AIParseResponse{
			Success: false,
			Error:   "Method not allowed",
		})
		return
	}

	// Ensure user is authenticated
	if err := h.Authorize(r); err != nil {
		log.Printf("Authorization failed: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(AIParseResponse{
			Success: false,
			Error:   "Unauthorized",
		})
		return
	}

	// Parse request body
	var req AIParseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("JSON decode error: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AIParseResponse{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	if req.Text == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AIParseResponse{
			Success: false,
			Error:   "Text field is required",
		})
		return
	}

	log.Printf("AI Parse request: %s", req.Text)

	// Get n8n webhook URL from environment variable
	n8nWebhookURL := os.Getenv("N8N_WEBHOOK_URL")
	if n8nWebhookURL == "" {
		// Default to local n8n instance
		n8nWebhookURL = "http://n8n:5678/webhook/parse-time-entry"
	}

	// Call n8n webhook
	parsedEntry, err := callN8NWebhook(n8nWebhookURL, req.Text)
	if err != nil {
		log.Printf("n8n webhook error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(AIParseResponse{
			Success: false,
			Error:   fmt.Sprintf("AI parsing failed: %v", err),
		})
		return
	}

	log.Printf("AI Parse result: project=%v, duration=%d, date=%s",
		parsedEntry.Project, parsedEntry.DurationMinutes, parsedEntry.Date)

	// Return successful response
	json.NewEncoder(w).Encode(AIParseResponse{
		Success: true,
		Data:    parsedEntry,
	})
}

// callN8NWebhook calls the n8n webhook and returns the parsed time entry
func callN8NWebhook(webhookURL string, text string) (*ParsedTimeEntry, error) {
	// Prepare request body with current date for context
	// This helps the AI understand relative dates like "heute", "gestern", etc.
	currentDate := time.Now().Format("2006-01-02")
	reqBody, err := json.Marshal(map[string]string{
		"text":        text,
		"currentDate": currentDate,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Make request to n8n
	resp, err := client.Post(webhookURL, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to call n8n webhook: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	log.Printf("n8n response status: %d, body: %s", resp.StatusCode, string(body))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("n8n returned status %d: %s", resp.StatusCode, string(body))
	}

	// Check for empty response (usually means n8n workflow error)
	if len(body) == 0 {
		return nil, fmt.Errorf("AI service returned empty response - please try again in a moment")
	}

	// Parse response - n8n returns an array, so we need to handle both cases
	var parsedEntry ParsedTimeEntry

	// First try to parse as array
	var parsedEntries []ParsedTimeEntry
	if err := json.Unmarshal(body, &parsedEntries); err == nil && len(parsedEntries) > 0 {
		parsedEntry = parsedEntries[0]
	} else {
		// Try to parse as single object
		if err := json.Unmarshal(body, &parsedEntry); err != nil {
			return nil, fmt.Errorf("failed to parse n8n response: %w", err)
		}
	}

	return &parsedEntry, nil
}
