-- name: GetUserByEmail :one
SELECT
    id,
    first_name,
    last_name,
    email,
    password_hash,
    session_token,
    csrf_token,
    role,
    team_id,
    start_date
FROM users
WHERE
    email = $1;

-- name: UpdateUserTokens :exec
UPDATE users
SET
    session_token = $1,
    csrf_token = $2
WHERE
    id = $3;

-- name: GetUserBySessionToken :one
SELECT *
FROM users
WHERE
    session_token = $1
    AND session_token IS NOT NULL
LIMIT 1;

-- name: GetPublicUserBySessionToken :one
SELECT
    id,
    first_name,
    last_name,
    email,
    role,
    team_id,
    start_date
FROM users
WHERE
    session_token = $1
LIMIT 1;

-- name: CreateUser :one
INSERT INTO
    users (
        first_name,
        last_name,
        email,
        password_hash,
        role,
        team_id,
        start_date
    )
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING
    id,
    first_name,
    last_name,
    email,
    role,
    created_at;

-- name: CreateTeam :one
INSERT INTO
    teams (name, supervisor_id)
VALUES ($1, $2)
RETURNING
    id,
    name,
    supervisor_id;

-- name: GetUser :one
SELECT
    id,
    first_name,
    last_name,
    email,
    role,
    team_id,
    start_date,
    created_at
FROM users
WHERE
    id = $1
LIMIT 1;

-- name: ListUsers :many
SELECT
    id,
    first_name,
    last_name,
    email,
    role,
    team_id,
    start_date
FROM users
ORDER BY id;

-- name: ListTeams :many
SELECT id, name, supervisor_id FROM teams ORDER BY id;

-- name: UpdateUser :one
UPDATE users
SET
    first_name = COALESCE($2, first_name),
    last_name = COALESCE($3, last_name),
    email = COALESCE($4, email),
    role = COALESCE($5, role),
    team_id = COALESCE($6, team_id),
    updated_at = NOW()
WHERE
    id = $1
RETURNING
    id,
    first_name,
    last_name,
    email,
    role,
    updated_at;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;

-- name: DeleteTeam :exec
DELETE FROM teams WHERE id = $1;