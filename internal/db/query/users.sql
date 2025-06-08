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
  supervisor_id,
  start_date
FROM users
WHERE email = $1;


-- name: UpdateUserTokens :exec
UPDATE users
SET    session_token = $1,
       csrf_token    = $2
WHERE  id            = $3;

-- name: GetUserBySessionToken :one
SELECT * FROM users
WHERE session_token = $1 AND session_token IS NOT NULL
LIMIT 1;