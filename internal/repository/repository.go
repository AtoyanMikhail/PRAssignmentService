package repository

import (
	"context"

	"github.com/AtoyanMikhail/PRAssignmentService/internal/models"
)

// TeamRepository описывает операции с командами
type TeamRepository interface {
	Create(ctx context.Context, teamName string) (models.Team, error)
	GetByName(ctx context.Context, teamName string) (models.Team, error)
	GetTeamByID(ctx context.Context, id int64) (models.Team, error)
	List(ctx context.Context) ([]models.Team, error)
	Exists(ctx context.Context, teamName string) (bool, error)
}

// UserRepository описывает операции с пользователями
type UserRepository interface {
	CreateUser(ctx context.Context, userID, username string, teamID int64, isActive bool) (models.User, error)
	GetByUserID(ctx context.Context, userID string) (models.User, error)
	GetUserByID(ctx context.Context, id int64) (models.User, error)
	Update(ctx context.Context, userID, username string, teamID int64, isActive bool) (models.User, error)
	UpdateIsActive(ctx context.Context, userID string, isActive bool) (models.User, error)
	ListByTeamID(ctx context.Context, teamID int64) ([]models.User, error)
	ListActiveByTeamID(ctx context.Context, teamID int64) ([]models.User, error)
	ListActiveByTeamIDExcludingUser(ctx context.Context, teamID int64, excludeUserID string) ([]models.User, error)
	GetWithTeam(ctx context.Context, userID string) (models.UserWithTeam, error)
	UserExists(ctx context.Context, userID string) (bool, error)
	DeactivateTeamUsers(ctx context.Context, teamID int64) ([]models.User, error)
}

// PullRequestRepository описывает операции с Pull Request'ами
type PullRequestRepository interface {
	CreatePR(ctx context.Context, pullRequestID, pullRequestName, authorID string, status models.PullRequestStatus) (models.PullRequest, error)
	GetPullRequestByPRID(ctx context.Context, pullRequestID string) (models.PullRequest, error)
	GetPullRequestByID(ctx context.Context, id int64) (models.PullRequest, error)
	UpdateStatus(ctx context.Context, pullRequestID string, status models.PullRequestStatus) (models.PullRequest, error)
	Merge(ctx context.Context, pullRequestID string) (models.PullRequest, error)
	ListByStatus(ctx context.Context, status models.PullRequestStatus) ([]models.PullRequest, error)
	PRExists(ctx context.Context, pullRequestID string) (bool, error)
}

// PRReviewerRepository описывает операции с ревьюерами
type PRReviewerRepository interface {
	Add(ctx context.Context, pullRequestID, userID string) (models.PRReviewer, error)
	Remove(ctx context.Context, pullRequestID, userID string) error
	GetReviewersByPRID(ctx context.Context, pullRequestID string) ([]models.ReviewerInfo, error)
	GetPRsByReviewerUserID(ctx context.Context, userID string) ([]models.PullRequestShort, error)
	IsUserAssigned(ctx context.Context, pullRequestID, userID string) (bool, error)
	Count(ctx context.Context, pullRequestID string) (int64, error)
	Replace(ctx context.Context, pullRequestID, oldUserID, newUserID string) error
	GetOpenPRsWithInactiveReviewers(ctx context.Context) ([]models.InactiveReviewerInfo, error)
	RemoveInactiveReviewers(ctx context.Context, pullRequestID string, userIDs []string) error
}

// StatisticsRepository описывает операции для получения статистики
type StatisticsRepository interface {
	GetAssignmentStats(ctx context.Context) ([]models.AssignmentStats, error)
	GetPRStats(ctx context.Context) ([]models.PRStats, error)
	GetTeamStats(ctx context.Context) ([]models.TeamStats, error)
	GetUserWorkload(ctx context.Context) ([]models.UserWorkload, error)
}
