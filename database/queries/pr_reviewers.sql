-- name: AddReviewer :one
INSERT INTO pr_reviewers (pull_request_id, user_id)
VALUES ($1, $2)
RETURNING *;

-- name: RemoveReviewer :exec
DELETE FROM pr_reviewers
WHERE pull_request_id = $1 AND user_id = $2;

-- name: GetReviewersByPRID :many
SELECT pr.*, u.username, u.team_id, u.is_active
FROM pr_reviewers pr
JOIN users u ON pr.user_id = u.user_id
WHERE pr.pull_request_id = $1
ORDER BY pr.assigned_at;

-- name: GetPullRequestsByReviewerUserID :many
SELECT pr.*, p.pull_request_name, p.author_id, p.status, p.created_at, p.merged_at
FROM pr_reviewers pr
JOIN pull_requests p ON pr.pull_request_id = p.pull_request_id
WHERE pr.user_id = $1
ORDER BY p.created_at DESC;

-- name: IsUserAssignedToPR :one
SELECT EXISTS(
    SELECT 1 FROM pr_reviewers
    WHERE pull_request_id = $1 AND user_id = $2
);

-- name: CountReviewersByPRID :one
SELECT COUNT(*) FROM pr_reviewers
WHERE pull_request_id = $1;

-- name: ReplaceReviewer :exec
UPDATE pr_reviewers
SET user_id = $3, assigned_at = NOW()
WHERE pull_request_id = $1 AND user_id = $2;

-- name: GetOpenPRsWithInactiveReviewers :many
SELECT DISTINCT p.pull_request_id, p.author_id, pr.user_id as inactive_reviewer_id, u.team_id
FROM pull_requests p
JOIN pr_reviewers pr ON p.pull_request_id = pr.pull_request_id
JOIN users u ON pr.user_id = u.user_id
WHERE p.status = 'OPEN' AND u.is_active = false
ORDER BY p.pull_request_id;

-- name: RemoveInactiveReviewers :exec
DELETE FROM pr_reviewers
WHERE pull_request_id = $1 AND user_id = ANY($2::text[]);
