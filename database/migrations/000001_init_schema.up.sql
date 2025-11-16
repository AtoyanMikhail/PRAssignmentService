-- Create teams table
CREATE TABLE IF NOT EXISTS teams (
    id BIGSERIAL PRIMARY KEY,
    team_name VARCHAR(255) NOT NULL UNIQUE CHECK (LENGTH(TRIM(team_name)) > 0),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_teams_team_name ON teams(team_name);

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL UNIQUE CHECK (LENGTH(TRIM(user_id)) > 0),
    username VARCHAR(255) NOT NULL CHECK (LENGTH(TRIM(username)) > 0),
    team_id BIGINT NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_user_id ON users(user_id);
CREATE INDEX idx_users_team_id ON users(team_id);
CREATE INDEX idx_users_team_id_is_active ON users(team_id, is_active);

-- Create pull_requests table
CREATE TABLE IF NOT EXISTS pull_requests (
    id BIGSERIAL PRIMARY KEY,
    pull_request_id VARCHAR(255) NOT NULL UNIQUE CHECK (LENGTH(TRIM(pull_request_id)) > 0),
    pull_request_name VARCHAR(255) NOT NULL CHECK (LENGTH(TRIM(pull_request_name)) > 0),
    author_id VARCHAR(255) NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'OPEN' CHECK (status IN ('OPEN', 'MERGED')),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    merged_at TIMESTAMP NULL
);

CREATE INDEX idx_pull_requests_pr_id ON pull_requests(pull_request_id);
CREATE INDEX idx_pull_requests_author_id ON pull_requests(author_id);
CREATE INDEX idx_pull_requests_status ON pull_requests(status);

-- Create pr_reviewers junction table
CREATE TABLE IF NOT EXISTS pr_reviewers (
    id BIGSERIAL PRIMARY KEY,
    pull_request_id VARCHAR(255) NOT NULL REFERENCES pull_requests(pull_request_id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    assigned_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(pull_request_id, user_id)
);

CREATE INDEX idx_pr_reviewers_pr_id ON pr_reviewers(pull_request_id);
CREATE INDEX idx_pr_reviewers_user_id ON pr_reviewers(user_id);
