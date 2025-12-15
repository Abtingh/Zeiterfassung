-- name: CreateWeeklySubmission :one
INSERT INTO
    weekly_submissions (
        user_id,
        week_number,
        year,
        status
    )
VALUES ($1, $2, $3, 'offen')
RETURNING
    *;

-- name: GetWeeklySubmission :one
SELECT *
FROM weekly_submissions
WHERE
    user_id = $1
    AND week_number = $2
    AND year = $3
LIMIT 1;

-- name: GetWeeklySubmissionsByID :one
SELECT * FROM weekly_submissions WHERE id = $1;

-- name: CreateTimeEntry :one
INSERT INTO
    time_entries (
        submission_id,
        entry_date,
        start_time,
        end_time,
        break_min,
        duration_h,
        note
    )
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING
    *;

-- name: GetTimeEntriesBySubmission :many
SELECT *
FROM time_entries
WHERE
    submission_id = $1
ORDER BY entry_date;

-- name: UpdateWeeklySubmissionStatus :one
UPDATE weekly_submissions
SET
    status = $2,
    updated_at = NOW()
WHERE
    id = $1
RETURNING
    *;

-- name: DeleteTimeEntriesBySubmission :exec
DELETE FROM time_entries WHERE submission_id = $1;

-- name: UpdateTimeEntry :one
UPDATE time_entries
SET
    start_time = $2,
    end_time = $3,
    break_min = $4,
    duration_h = $5,
    note = $6
WHERE
    id = $1
RETURNING
    *;

-- name: GetSupervisorTeams :many
-- Get all teams where the user is a supervisor
SELECT id, name, supervisor_id FROM teams WHERE supervisor_id = $1;

-- name: GetStudentsForSupervisor :many
-- Get all students under this supervisor (from all their teams)
SELECT u.id, u.first_name, u.last_name, u.email, u.team_id
FROM users u
    JOIN teams t ON u.team_id = t.id
WHERE
    t.supervisor_id = $1
    AND u.role = 'student'
ORDER BY u.last_name, u.first_name;

-- name: GetStudentByTeamID :many
-- Get all students in a specific team
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
    team_id = $1
    AND role = 'student'
ORDER BY last_name, first_name;

-- name: GetPendingSubmissionsForSupervisor :many
-- Get all weekly submissions with status 'gesendet' for students under this supervisor
SELECT ws.id, ws.user_id, ws.week_number, ws.year, ws.status, ws.submitted_at, u.first_name, u.last_name, u.email
FROM
    weekly_submissions ws
    JOIN users u ON ws.user_id = u.id
    JOIN teams t ON u.team_id = t.id
WHERE
    t.supervisor_id = $1
    AND ws.status = 'gesendet'
ORDER BY ws.year DESC, ws.week_number DESC, u.last_name;

-- name: GetTimeEntriesBySubmissionID :many
-- Get all time entries for a specific weekly submission
SELECT
    id,
    submission_id,
    entry_date,
    start_time,
    end_time,
    break_min,
    duration_h,
    note
FROM time_entries
WHERE
    submission_id = $1
ORDER BY entry_date;

-- name: ApproveWeeklySubmission :one
-- Approve a weekly submission (set status to 'bestaetigt')
UPDATE weekly_submissions
SET
    status = 'bestaetigt',
    approved_at = NOW(),
    processed_by = $2,
    updated_at = NOW()
WHERE
    id = $1
RETURNING
    id,
    user_id,
    week_number,
    year,
    status,
    approved_at,
    processed_by;

-- name: RejectWeeklySubmission :one
-- Reject a weekly submission (set status to 'korrektur')
UPDATE weekly_submissions
SET
    status = 'korrektur',
    approved_at = NULL,
    processed_by = $2,
    updated_at = NOW()
WHERE
    id = $1
RETURNING
    id,
    user_id,
    week_number,
    year,
    status,
    approved_at,
    processed_by;

-- name: GetWeeklySubmissionByID :one
-- Get a specific weekly submission with user details
SELECT ws.id, ws.user_id, ws.week_number, ws.year, ws.status, ws.submitted_at, ws.approved_at, ws.processed_by, u.first_name, u.last_name, u.email, u.team_id
FROM
    weekly_submissions ws
    JOIN users u ON ws.user_id = u.id
WHERE
    ws.id = $1;

-- name: VerifySubmissionBelongsToSupervisor :one
-- Verify that a submission belongs to a student under this supervisor
SELECT EXISTS (
        SELECT 1
        FROM
            weekly_submissions ws
            JOIN users u ON ws.user_id = u.id
            JOIN teams t ON u.team_id = t.id
        WHERE
            ws.id = $1
            AND t.supervisor_id = $2
    ) AS is_authorized;