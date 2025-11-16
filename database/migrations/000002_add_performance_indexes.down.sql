-- Remove performance indexes
DROP INDEX IF EXISTS idx_pull_requests_status_created_at;
DROP INDEX IF EXISTS idx_pr_reviewers_pr_id_user_id;
DROP INDEX IF EXISTS idx_users_is_active_team_id;
