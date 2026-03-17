-- name: InsertUserSession :exec
INSERT INTO user_sessions (id, user_id, cognito_username, refresh_token, expires_at)
VALUES ($1, $2, $3, $4, $5);

-- name: GetSessionByRefreshToken :one
SELECT id, user_id, cognito_username, refresh_token, expires_at
FROM user_sessions
WHERE refresh_token = $1
  AND expires_at > NOW();

-- name: DeleteSessionByRefreshToken :exec
DELETE FROM user_sessions
WHERE refresh_token = $1;

-- name: DeleteSessionsByUserID :exec
DELETE FROM user_sessions
WHERE user_id = $1;
