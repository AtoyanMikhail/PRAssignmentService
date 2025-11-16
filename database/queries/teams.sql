-- name: CreateTeam :one
INSERT INTO teams (team_name)
VALUES ($1)
RETURNING *;

-- name: GetTeamByName :one
SELECT * FROM teams
WHERE team_name = $1 LIMIT 1;

-- name: GetTeamByID :one
SELECT * FROM teams
WHERE id = $1 LIMIT 1;

-- name: ListTeams :many
SELECT * FROM teams
ORDER BY team_name;

-- name: TeamExists :one
SELECT EXISTS(SELECT 1 FROM teams WHERE team_name = $1);
