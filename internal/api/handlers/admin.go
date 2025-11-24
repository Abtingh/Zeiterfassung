package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	db "github.com/Abtingh/Zeiterfassung/internal/db/sqlc"
	"github.com/Abtingh/Zeiterfassung/internal/util"
)

type createUserRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	Role      string `json:"role"`
	TeamID    *int64 `json:"team_id"`
	StartDate string `json:"start_date"` // Format: YYYY-MM-DD
}

type userResponse struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	TeamID    *int64 `json:"team_id,omitempty"`
}

func newUserResponse(user db.CreateUserRow) userResponse {
	var firstName, lastName string
	if user.FirstName.Valid {
		firstName = user.FirstName.String
	}
	if user.LastName.Valid {
		lastName = user.LastName.String
	}

	return userResponse{
		ID:        user.ID,
		FirstName: firstName,
		LastName:  lastName,
		Email:     user.Email,
		Role:      string(user.Role),
		TeamID:    nil, // CreateUserRow doesn't return TeamID
	}
}

func newListUserResponse(user db.ListUsersRow) userResponse {
	var firstName, lastName string
	var teamID *int64

	if user.FirstName.Valid {
		firstName = user.FirstName.String
	}
	if user.LastName.Valid {
		lastName = user.LastName.String
	}
	if user.TeamID.Valid {
		teamID = &user.TeamID.Int64
	}

	return userResponse{
		ID:        user.ID,
		FirstName: firstName,
		LastName:  lastName,
		Email:     user.Email,
		Role:      string(user.Role),
		TeamID:    teamID,
	}
}

// CreateUserHandler handles the creation of a new user by an admin.
//
// Security: This endpoint requires admin role authorization.
//
// How it works:
// 1. Validates the request method (POST only) and handles CORS
// 2. Verifies the user is authenticated via session token
// 3. Checks if the authenticated user has the 'admin' role
// 4. Parses and validates the request body (first name, last name, email, password, role)
// 5. Validates password strength (minimum 8 characters)
// 6. Validates the role is one of: admin, student, supervisor, accounting
// 7. Hashes the password using bcrypt before storing
// 8. Converts string values to appropriate PostgreSQL types (pgtype.Text, pgtype.Int8, etc.)
// 9. Inserts the new user into the database
// 10. Returns the created user (without sensitive data like password hash)
//
// Request body:
//   - first_name: User's first name (required)
//   - last_name: User's last name (required)
//   - email: User's email address (required, must be unique)
//   - password: User's password (required, min 8 characters)
//   - role: User's role (required, one of: admin, student, supervisor, accounting)
//   - team_id: Optional team assignment (for students/supervisors)
//   - start_date: Optional start date in YYYY-MM-DD format
//
// Returns:
//   - 201 Created: User successfully created with user details
//   - 400 Bad Request: Invalid input data
//   - 401 Unauthorized: Not authenticated
//   - 403 Forbidden: Not an admin user
//   - 409 Conflict: Email already exists
//   - 500 Internal Server Error: Database or server error
func (h *Handler) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	// Step 1: Validate user is authenticated
	if err := h.Authorize(r); err != nil {
		log.Printf("CreateUserHandler: Authorization failed: %v", err)
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// Step 2: Get current user and verify they have admin role
	currentUser, err := h.GetCurrentUser(r)
	if err != nil {
		log.Printf("CreateUserHandler: Failed to get current user: %v", err)
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// Step 3: Check if user has admin role
	if currentUser.Role != db.RoleEnumAdmin {
		log.Printf("CreateUserHandler: Access denied for user %s with role %s", currentUser.Email, currentUser.Role)
		http.Error(w, `{"error":"Forbidden: Admin role required"}`, http.StatusForbidden)
		return
	}

	log.Printf("CreateUserHandler: Admin user %s authorized to create users", currentUser.Email)

	// Parse request body
	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("CreateUserHandler: Failed to decode request: %v", err)
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.FirstName == "" || req.LastName == "" || req.Email == "" || req.Password == "" || req.Role == "" {
		http.Error(w, `{"error":"Missing required fields"}`, http.StatusBadRequest)
		return
	}

	// Validate password length
	if len(req.Password) < 8 {
		http.Error(w, `{"error":"Password must be at least 8 characters"}`, http.StatusBadRequest)
		return
	}

	// Validate role
	if req.Role != "admin" && req.Role != "student" && req.Role != "supervisor" && req.Role != "accounting" {
		http.Error(w, `{"error":"Invalid role. Must be: admin, student, supervisor, or accounting"}`, http.StatusBadRequest)
		return
	}

	// Hash password
	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		log.Printf("CreateUserHandler: Failed to hash password: %v", err)
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}

	// Convert role string to RoleEnum
	var roleEnum db.RoleEnum
	switch req.Role {
	case "admin":
		roleEnum = db.RoleEnumAdmin
	case "student":
		roleEnum = db.RoleEnumStudent
	case "supervisor":
		roleEnum = db.RoleEnumSupervisor
	case "accounting":
		roleEnum = db.RoleEnumAccounting
	}

	// Parse start date
	var startDate pgtype.Date
	if req.StartDate != "" {
		t, err := time.Parse("2006-01-02", req.StartDate)
		if err != nil {
			http.Error(w, `{"error":"Invalid start_date format. Use YYYY-MM-DD"}`, http.StatusBadRequest)
			return
		}
		startDate = pgtype.Date{
			Time:  t,
			Valid: true,
		}
	}

	// Create user in database
	arg := db.CreateUserParams{
		FirstName: pgtype.Text{
			String: req.FirstName,
			Valid:  true,
		},
		LastName: pgtype.Text{
			String: req.LastName,
			Valid:  true,
		},
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Role:         roleEnum,
		TeamID: pgtype.Int8{
			Int64: 0,
			Valid: false,
		},
		StartDate: startDate,
	}

	// Set team ID if provided
	if req.TeamID != nil {
		arg.TeamID = pgtype.Int8{
			Int64: *req.TeamID,
			Valid: true,
		}
	}

	user, err := h.Q.CreateUser(r.Context(), arg)
	if err != nil {
		log.Printf("CreateUserHandler: Failed to create user: %v", err)
		// Check if email already exists
		if err.Error() == "duplicate key value violates unique constraint" {
			http.Error(w, `{"error":"Email already exists"}`, http.StatusConflict)
			return
		}
		http.Error(w, `{"error":"Failed to create user"}`, http.StatusInternalServerError)
		return
	}

	// Return user response
	rsp := newUserResponse(user)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(rsp)
}

// ListUsersHandler handles listing all users in the system.
//
// Security: This endpoint requires admin role authorization.
//
// How it works:
// 1. Validates the request method (GET only) and handles CORS
// 2. Verifies the user is authenticated via session token
// 3. Checks if the authenticated user has the 'admin' role
// 4. Queries the database for all users (ordered by ID)
// 5. Converts database types (pgtype.Text, pgtype.Int8) to JSON-friendly types
// 6. Returns the list of users with their details (excluding sensitive data)
//
// Note: Currently returns all users without pagination. For large user bases,
// consider adding LIMIT and OFFSET parameters to the SQL query.
//
// Returns:
//   - 200 OK: Array of user objects
//   - 401 Unauthorized: Not authenticated
//   - 403 Forbidden: Not an admin user
//   - 405 Method Not Allowed: Wrong HTTP method
//   - 500 Internal Server Error: Database error
func (h *Handler) ListUsersHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("→ %s %s", r.Method, r.URL.Path)

	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

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

	// Step 1: Validate user is authenticated
	if err := h.Authorize(r); err != nil {
		log.Printf("ListUsersHandler: Authorization failed: %v", err)
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// Step 2: Get current user and verify they have admin role
	currentUser, err := h.GetCurrentUser(r)
	if err != nil {
		log.Printf("ListUsersHandler: Failed to get current user: %v", err)
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// Step 3: Check if user has admin role
	if currentUser.Role != db.RoleEnumAdmin {
		log.Printf("ListUsersHandler: Access denied for user %s with role %s", currentUser.Email, currentUser.Role)
		http.Error(w, `{"error":"Forbidden: Admin role required"}`, http.StatusForbidden)
		return
	}

	log.Printf("ListUsersHandler: Admin user %s authorized to list users", currentUser.Email)

	// Get users from database (no pagination parameters needed)
	users, err := h.Q.ListUsers(r.Context())
	if err != nil {
		log.Printf("ListUsersHandler: Failed to list users: %v", err)
		http.Error(w, `{"error":"Failed to list users"}`, http.StatusInternalServerError)
		return
	}

	// Convert to response format
	var response []userResponse
	for _, user := range users {
		response = append(response, newListUserResponse(user))
	}

	// Return users list
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
