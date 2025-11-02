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