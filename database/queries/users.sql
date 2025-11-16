-- name: CreateUser :one
INSERT INTO users (user_id, username, team_id, is_active)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetUserByUserID :one
SELECT * FROM users
WHERE user_id = $1 LIMIT 1;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;

-- name: UpdateUser :one
UPDATE users
SET username = $2, team_id = $3, is_active = $4, updated_at = NOW()
WHERE user_id = $1
RETURNING *;

-- name: UpdateUserIsActive :one
UPDATE users
SET is_active = $2, updated_at = NOW()
WHERE user_id = $1
RETURNING *;

-- name: ListUsersByTeamID :many
SELECT * FROM users
WHERE team_id = $1
ORDER BY username;

-- name: ListActiveUsersByTeamID :many
SELECT * FROM users
WHERE team_id = $1 AND is_active = true
ORDER BY username;

-- name: ListActiveUsersByTeamIDExcludingUser :many
SELECT * FROM users
WHERE team_id = $1 AND is_active = true AND user_id != $2
ORDER BY username;

-- name: GetUserWithTeam :one
SELECT u.*, t.team_name
FROM users u
JOIN teams t ON u.team_id = t.id
WHERE u.user_id = $1 LIMIT 1;

-- name: UserExists :one
SELECT EXISTS(SELECT 1 FROM users WHERE user_id = $1);

-- name: DeactivateTeamUsers :many
UPDATE users
SET is_active = false, updated_at = NOW()
WHERE team_id = $1 AND is_active = true
RETURNING *;
