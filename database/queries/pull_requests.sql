-- name: CreatePullRequest :one
INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetPullRequestByPRID :one
SELECT * FROM pull_requests
WHERE pull_request_id = $1 LIMIT 1;

-- name: GetPullRequestByID :one
SELECT * FROM pull_requests
WHERE id = $1 LIMIT 1;

-- name: UpdatePullRequestStatus :one
UPDATE pull_requests
SET status = $2, merged_at = CASE WHEN $2 = 'MERGED' THEN NOW() ELSE merged_at END
WHERE pull_request_id = $1
RETURNING *;

-- name: MergePullRequest :one
UPDATE pull_requests
SET status = 'MERGED', merged_at = COALESCE(merged_at, NOW())
WHERE pull_request_id = $1
RETURNING *;

-- name: ListPullRequests :many
SELECT * FROM pull_requests
ORDER BY created_at DESC;

-- name: ListPullRequestsByStatus :many
SELECT * FROM pull_requests
WHERE status = $1
ORDER BY created_at DESC;

-- name: PullRequestExists :one
SELECT EXISTS(SELECT 1 FROM pull_requests WHERE pull_request_id = $1);
