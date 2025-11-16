package service

import (
	"context"
	"errors"

	"github.com/AtoyanMikhail/PRAssignmentService/internal/models"
)

// Ошибки сервисного слоя
var (
	ErrTeamNotFound             = errors.New("team not found")
	ErrTeamAlreadyExists        = errors.New("team already exists")
	ErrUserNotFound             = errors.New("user not found")
	ErrUserAlreadyExists        = errors.New("user already exists")
	ErrUserInactive             = errors.New("user is inactive")
	ErrPullRequestNotFound      = errors.New("pull request not found")
	ErrPullRequestAlreadyExists = errors.New("pull request already exists")
	ErrReviewerAlreadyAssigned  = errors.New("reviewer already assigned")
	ErrReviewerNotAssigned      = errors.New("reviewer not assigned")
	ErrCannotAssignAuthor       = errors.New("cannot assign PR author as reviewer")
	ErrNoActiveReviewers        = errors.New("no active reviewers available")
	ErrInvalidStatus            = errors.New("invalid pull request status")
)

// TeamService управляет операциями с командами
type TeamService interface {
	// CreateTeam создает новую команду
	CreateTeam(ctx context.Context, teamName string) (models.Team, error)

	// GetTeam возвращает команду по имени
	GetTeam(ctx context.Context, teamName string) (models.Team, error)

	// GetTeamByID возвращает команду по ID
	GetTeamByID(ctx context.Context, id int64) (models.Team, error)

	// ListTeams возвращает список всех команд
	ListTeams(ctx context.Context) ([]models.Team, error)
}

// UserService управляет операциями с пользователями
type UserService interface {
	// CreateUser создает нового пользователя
	CreateUser(ctx context.Context, userID, username string, teamID int64) (models.User, error)

	// GetUser возвращает пользователя по userID
	GetUser(ctx context.Context, userID string) (models.User, error)

	// GetUserWithTeam возвращает пользователя с информацией о команде
	GetUserWithTeam(ctx context.Context, userID string) (models.UserWithTeam, error)

	// UpdateUser обновляет данные пользователя
	UpdateUser(ctx context.Context, userID, username string, teamID int64, isActive bool) (models.User, error)

	// DeactivateUser деактивирует пользователя
	DeactivateUser(ctx context.Context, userID string) (models.User, error)

	// ActivateUser активирует пользователя
	ActivateUser(ctx context.Context, userID string) (models.User, error)

	// ListTeamUsers возвращает всех пользователей команды
	ListTeamUsers(ctx context.Context, teamID int64) ([]models.User, error)

	// ListActiveTeamUsers возвращает активных пользователей команды
	ListActiveTeamUsers(ctx context.Context, teamID int64) ([]models.User, error)

	// DeactivateTeamUsers деактивирует всех пользователей команды и перераспределяет их PR
	// Возвращает количество деактивированных пользователей и количество переназначенных PR
	DeactivateTeamUsers(ctx context.Context, teamID int64) (deactivatedUsers int, reassignedPRs int, err error)
}

// PullRequestService управляет операциями с Pull Request'ами
type PullRequestService interface {
	// CreatePR создает новый Pull Request
	CreatePR(ctx context.Context, pullRequestID, pullRequestName, authorID string) (models.PullRequest, error)

	// GetPR возвращает Pull Request по ID
	GetPR(ctx context.Context, pullRequestID string) (models.PullRequest, error)

	// UpdatePRStatus обновляет статус Pull Request
	UpdatePRStatus(ctx context.Context, pullRequestID string, status models.PullRequestStatus) (models.PullRequest, error)

	// MergePR помечает Pull Request как merged
	MergePR(ctx context.Context, pullRequestID string) (models.PullRequest, error)

	// ListOpenPRs возвращает список открытых Pull Request'ов
	ListOpenPRs(ctx context.Context) ([]models.PullRequest, error)

	// ListPRsByStatus возвращает Pull Request'ы по статусу
	ListPRsByStatus(ctx context.Context, status models.PullRequestStatus) ([]models.PullRequest, error)
}

// ReviewerService управляет назначением ревьюеров
type ReviewerService interface {
	// AssignReviewer назначает ревьюера на Pull Request
	AssignReviewer(ctx context.Context, pullRequestID, userID string) (models.PRReviewer, error)

	// AssignReviewers назначает несколько ревьюеров на Pull Request
	AssignReviewers(ctx context.Context, pullRequestID string, userIDs []string) ([]models.PRReviewer, error)

	// RemoveReviewer удаляет ревьюера с Pull Request
	RemoveReviewer(ctx context.Context, pullRequestID, userID string) error

	// ReplaceReviewer заменяет одного ревьюера другим
	ReplaceReviewer(ctx context.Context, pullRequestID, oldUserID, newUserID string) error

	// GetPRReviewers возвращает список ревьюеров для Pull Request
	GetPRReviewers(ctx context.Context, pullRequestID string) ([]models.ReviewerInfo, error)

	// GetUserPRs возвращает Pull Request'ы, где пользователь является ревьюером
	GetUserPRs(ctx context.Context, userID string) ([]models.PullRequestShort, error)

	// AutoAssignReviewers автоматически назначает ревьюеров на Pull Request
	// на основе алгоритма балансировки нагрузки
	AutoAssignReviewers(ctx context.Context, pullRequestID string, count int) ([]models.PRReviewer, error)

	// ReassignFromInactiveReviewers находит все PR с неактивными ревьюерами
	// и переназначает их на активных пользователей
	ReassignFromInactiveReviewers(ctx context.Context) error
}

// StatisticsService предоставляет статистику
type StatisticsService interface {
	// GetAssignmentStats возвращает статистику по назначениям ревьюеров
	GetAssignmentStats(ctx context.Context) ([]models.AssignmentStats, error)

	// GetPRStats возвращает статистику по Pull Request'ам
	GetPRStats(ctx context.Context) ([]models.PRStats, error)

	// GetTeamStats возвращает статистику по командам
	GetTeamStats(ctx context.Context) ([]models.TeamStats, error)

	// GetUserWorkload возвращает нагрузку пользователей
	GetUserWorkload(ctx context.Context) ([]models.UserWorkload, error)
}
