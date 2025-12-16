package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
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

type createTeamRequest struct {
	Name         string `json:"name"`
	SupervisorID *int64 `json:"supervisor_id"`
}

type updateTeamSupervisorRequest struct {
	TeamID       int64  `json:"team_id"`
	SupervisorID *int64 `json:"supervisor_id"`
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

// CreateTeamHandler handles the creation of a new team by an admin.
func (h *Handler) CreateTeamHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("→ %s %s", r.Method, r.URL.Path)

	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	if err := h.Authorize(r); err != nil {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	currentUser, err := h.GetCurrentUser(r)
	if err != nil || currentUser.Role != db.RoleEnumAdmin {
		http.Error(w, `{"error":"Forbidden: Admin role required"}`, http.StatusForbidden)
		return
	}

	var req createTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, `{"error":"Team name is required"}`, http.StatusBadRequest)
		return
	}

	var supervisorID pgtype.Int8
	if req.SupervisorID != nil {
		supervisorID = pgtype.Int8{Int64: *req.SupervisorID, Valid: true}
	}

	team, err := h.Q.CreateTeam(r.Context(), db.CreateTeamParams{
		Name:         req.Name,
		SupervisorID: supervisorID,
	})
	if err != nil {
		log.Printf("CreateTeamHandler: Failed to create team: %v", err)
		http.Error(w, `{"error":"Failed to create team. Name might already exist."}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(team)
}

// ListTeamsHandler returns all teams with supervisor details
func (h *Handler) ListTeamsHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("→ %s %s", r.Method, r.URL.Path)

	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	if err := h.Authorize(r); err != nil {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	currentUser, err := h.GetCurrentUser(r)
	if err != nil || currentUser.Role != db.RoleEnumAdmin {
		http.Error(w, `{"error":"Forbidden: Admin role required"}`, http.StatusForbidden)
		return
	}

	teams, err := h.Q.ListTeamsWithDetails(r.Context())
	if err != nil {
		log.Printf("ListTeamsHandler: Failed to list teams: %v", err)
		http.Error(w, `{"error":"Failed to list teams"}`, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(teams)
}

// GetTeamMembersHandler returns all members of a specific team
func (h *Handler) GetTeamMembersHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("→ %s %s", r.Method, r.URL.Path)

	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	if err := h.Authorize(r); err != nil {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	currentUser, err := h.GetCurrentUser(r)
	if err != nil || currentUser.Role != db.RoleEnumAdmin {
		http.Error(w, `{"error":"Forbidden: Admin role required"}`, http.StatusForbidden)
		return
	}

	teamIDStr := r.URL.Query().Get("team_id")
	if teamIDStr == "" {
		http.Error(w, `{"error":"team_id is required"}`, http.StatusBadRequest)
		return
	}

	teamID, err := strconv.ParseInt(teamIDStr, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"Invalid team_id"}`, http.StatusBadRequest)
		return
	}

	members, err := h.Q.GetTeamMembers(r.Context(), pgtype.Int8{Int64: teamID, Valid: true})
	if err != nil {
		log.Printf("GetTeamMembersHandler: Failed to get team members: %v", err)
		http.Error(w, `{"error":"Failed to get team members"}`, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(members)
}

// UpdateTeamSupervisorHandler updates the supervisor of a team
func (h *Handler) UpdateTeamSupervisorHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("→ %s %s", r.Method, r.URL.Path)

	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	if err := h.Authorize(r); err != nil {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	currentUser, err := h.GetCurrentUser(r)
	if err != nil || currentUser.Role != db.RoleEnumAdmin {
		http.Error(w, `{"error":"Forbidden: Admin role required"}`, http.StatusForbidden)
		return
	}

	var req updateTeamSupervisorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	var supervisorID pgtype.Int8
	if req.SupervisorID != nil {
		supervisorID = pgtype.Int8{Int64: *req.SupervisorID, Valid: true}
	}

	team, err := h.Q.UpdateTeamSupervisor(r.Context(), db.UpdateTeamSupervisorParams{
		ID:           req.TeamID,
		SupervisorID: supervisorID,
	})
	if err != nil {
		log.Printf("UpdateTeamSupervisorHandler: Failed to update team: %v", err)
		http.Error(w, `{"error":"Failed to update team supervisor"}`, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(team)
}

// GetAllSupervisorsHandler returns all users with supervisor role
func (h *Handler) GetAllSupervisorsHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("→ %s %s", r.Method, r.URL.Path)

	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	if err := h.Authorize(r); err != nil {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	currentUser, err := h.GetCurrentUser(r)
	if err != nil || currentUser.Role != db.RoleEnumAdmin {
		http.Error(w, `{"error":"Forbidden: Admin role required"}`, http.StatusForbidden)
		return
	}

	supervisors, err := h.Q.GetAllSupervisors(r.Context())
	if err != nil {
		log.Printf("GetAllSupervisorsHandler: Failed to get supervisors: %v", err)
		http.Error(w, `{"error":"Failed to get supervisors"}`, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(supervisors)
}

// AdminTeamHinzufuegenHandler serves the Team Hinzufügen page
func (h *Handler) AdminTeamHinzufuegenHandler(w http.ResponseWriter, r *http.Request) {
	if err := h.Authorize(r); err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	user, err := h.GetCurrentUser(r)
	if err != nil || user.Role != db.RoleEnumAdmin {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	http.ServeFile(w, r, "./public/public-static/admin_teamHinzufuegen.html")
}

// AdminTeamVerwaltenHandler serves the Team Verwalten page
func (h *Handler) AdminTeamVerwaltenHandler(w http.ResponseWriter, r *http.Request) {
	if err := h.Authorize(r); err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	user, err := h.GetCurrentUser(r)
	if err != nil || user.Role != db.RoleEnumAdmin {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	http.ServeFile(w, r, "./public/public-static/admin_teamVerwalten.html")
}

// AdminBenutzerAnlegenHandler serves the Benutzer Anlegen page
func (h *Handler) AdminBenutzerAnlegenHandler(w http.ResponseWriter, r *http.Request) {
	if err := h.Authorize(r); err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	user, err := h.GetCurrentUser(r)
	if err != nil || user.Role != db.RoleEnumAdmin {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	http.ServeFile(w, r, "./public/public-static/admin_benutzerAnlegen.html")
}

// AdminBenutzerVerwaltenHandler serves the Benutzer Verwalten page
func (h *Handler) AdminBenutzerVerwaltenHandler(w http.ResponseWriter, r *http.Request) {
	if err := h.Authorize(r); err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	user, err := h.GetCurrentUser(r)
	if err != nil || user.Role != db.RoleEnumAdmin {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	http.ServeFile(w, r, "./public/public-static/admin_benutzerVerwalten.html")
}

type updateUserRequest struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	TeamID    *int64 `json:"team_id"`
	Password  string `json:"password,omitempty"`
}

// UpdateUserHandler handles updating user information by an admin
func (h *Handler) UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("→ %s %s", r.Method, r.URL.Path)

	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if r.Method != http.MethodPut && r.Method != http.MethodPost {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	if err := h.Authorize(r); err != nil {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	currentUser, err := h.GetCurrentUser(r)
	if err != nil || currentUser.Role != db.RoleEnumAdmin {
		http.Error(w, `{"error":"Forbidden: Admin role required"}`, http.StatusForbidden)
		return
	}

	var req updateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("UpdateUserHandler: Failed to decode request: %v", err)
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Update user details
	arg := db.UpdateUserParams{
		ID: req.ID,
		FirstName: pgtype.Text{
			String: req.FirstName,
			Valid:  req.FirstName != "",
		},
		LastName: pgtype.Text{
			String: req.LastName,
			Valid:  req.LastName != "",
		},
		Email: req.Email,
		Role:  db.RoleEnum(req.Role),
	}

	if req.TeamID != nil {
		arg.TeamID = pgtype.Int8{
			Int64: *req.TeamID,
			Valid: true,
		}
	}

	_, err = h.Q.UpdateUser(r.Context(), arg)
	if err != nil {
		log.Printf("UpdateUserHandler: Failed to update user: %v", err)
		http.Error(w, `{"error":"Failed to update user"}`, http.StatusInternalServerError)
		return
	}

	// If password is provided, update it separately
	if req.Password != "" {
		hashedPassword, err := util.HashPassword(req.Password)
		if err != nil {
			log.Printf("UpdateUserHandler: Failed to hash password: %v", err)
			http.Error(w, `{"error":"Failed to hash password"}`, http.StatusInternalServerError)
			return
		}

		err = h.Q.UpdateUserPassword(r.Context(), db.UpdateUserPasswordParams{
			ID:           req.ID,
			PasswordHash: hashedPassword,
		})
		if err != nil {
			log.Printf("UpdateUserHandler: Failed to update password: %v", err)
			http.Error(w, `{"error":"Failed to update password"}`, http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "User updated successfully"})
}

type deactivateUserRequest struct {
	ID int64 `json:"id"`
}

// DeactivateUserHandler handles deactivating a user by an admin
func (h *Handler) DeactivateUserHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("→ %s %s", r.Method, r.URL.Path)

	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if r.Method != http.MethodPut && r.Method != http.MethodPost {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	if err := h.Authorize(r); err != nil {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	currentUser, err := h.GetCurrentUser(r)
	if err != nil || currentUser.Role != db.RoleEnumAdmin {
		http.Error(w, `{"error":"Forbidden: Admin role required"}`, http.StatusForbidden)
		return
	}

	var req deactivateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("DeactivateUserHandler: Failed to decode request: %v", err)
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Delete the user (or you could add an "active" field and set it to false)
	err = h.Q.DeleteUser(r.Context(), req.ID)
	if err != nil {
		log.Printf("DeactivateUserHandler: Failed to deactivate user: %v", err)
		http.Error(w, `{"error":"Failed to deactivate user"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "User deactivated successfully"})
}

type deleteTeamRequest struct {
	ID int64 `json:"id"`
}

// DeleteTeamHandler handles deleting a team by an admin
func (h *Handler) DeleteTeamHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("→ %s %s", r.Method, r.URL.Path)

	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if r.Method != http.MethodDelete && r.Method != http.MethodPost {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	if err := h.Authorize(r); err != nil {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	currentUser, err := h.GetCurrentUser(r)
	if err != nil || currentUser.Role != db.RoleEnumAdmin {
		http.Error(w, `{"error":"Forbidden: Admin role required"}`, http.StatusForbidden)
		return
	}

	var req deleteTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("DeleteTeamHandler: Failed to decode request: %v", err)
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Delete the team
	err = h.Q.DeleteTeam(r.Context(), req.ID)
	if err != nil {
		log.Printf("DeleteTeamHandler: Failed to delete team: %v", err)
		http.Error(w, `{"error":"Failed to delete team. Make sure no users are assigned to this team."}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Team deleted successfully"})
}
