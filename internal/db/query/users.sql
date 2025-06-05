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
