-- name: GetAssignmentStats :many
-- Статистика назначений по пользователям
SELECT 
    u.user_id,
    u.username,
    t.team_name,
    COUNT(pr.id) as total_assignments,
    COUNT(CASE WHEN p.status = 'OPEN' THEN 1 END) as open_prs,
    COUNT(CASE WHEN p.status = 'MERGED' THEN 1 END) as merged_prs
FROM users u
LEFT JOIN teams t ON u.team_id = t.id
LEFT JOIN pr_reviewers pr ON u.user_id = pr.user_id
LEFT JOIN pull_requests p ON pr.pull_request_id = p.pull_request_id
GROUP BY u.user_id, u.username, t.team_name
ORDER BY total_assignments DESC;

-- name: GetPRStats :many
-- Статистика по Pull Request'ам
SELECT 
    pr.pull_request_id,
    pr.pull_request_name,
    pr.author_id,
    pr.status,
    COUNT(r.id) as reviewers_count,
    pr.created_at,
    pr.merged_at
FROM pull_requests pr
LEFT JOIN pr_reviewers r ON pr.pull_request_id = r.pull_request_id
GROUP BY pr.pull_request_id, pr.pull_request_name, pr.author_id, pr.status, pr.created_at, pr.merged_at
ORDER BY pr.created_at DESC;

-- name: GetTeamStats :many
-- Статистика по командам
SELECT 
    t.team_name,
    COUNT(DISTINCT u.user_id) as total_members,
    COUNT(DISTINCT CASE WHEN u.is_active THEN u.user_id END) as active_members,
    COUNT(DISTINCT pr.pull_request_id) as total_prs_authored,
    COUNT(DISTINCT r.pull_request_id) as total_prs_reviewed
FROM teams t
LEFT JOIN users u ON t.id = u.team_id
LEFT JOIN pull_requests pr ON u.user_id = pr.author_id
LEFT JOIN pr_reviewers r ON u.user_id = r.user_id
GROUP BY t.team_name
ORDER BY t.team_name;

-- name: GetUserWorkload :many
-- Рабочая нагрузка пользователей (только открытые PR)
SELECT 
    u.user_id,
    u.username,
    t.team_name,
    u.is_active,
    COUNT(r.id) as open_reviews_count
FROM users u
LEFT JOIN teams t ON u.team_id = t.id
LEFT JOIN pr_reviewers r ON u.user_id = r.user_id
LEFT JOIN pull_requests pr ON r.pull_request_id = pr.pull_request_id AND pr.status = 'OPEN'
WHERE u.is_active = true
GROUP BY u.user_id, u.username, t.team_name, u.is_active
ORDER BY open_reviews_count DESC, u.username;
