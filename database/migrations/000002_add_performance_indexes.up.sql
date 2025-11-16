-- Add additional indexes for better performance on statistics queries

-- Composite index for faster filtering by status and created_at
CREATE INDEX IF NOT EXISTS idx_pull_requests_status_created_at ON pull_requests(status, created_at DESC);

-- Index for counting reviewers per PR
CREATE INDEX IF NOT EXISTS idx_pr_reviewers_pr_id_user_id ON pr_reviewers(pull_request_id, user_id);

-- Index for user activity queries
CREATE INDEX IF NOT EXISTS idx_users_is_active_team_id ON users(is_active, team_id) WHERE is_active = true;
