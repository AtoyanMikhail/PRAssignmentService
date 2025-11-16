package service

import "github.com/AtoyanMikhail/PRAssignmentService/internal/repository"

// Services содержит все сервисы приложения
type Services struct {
	Team        TeamService
	User        UserService
	PullRequest PullRequestService
	Reviewer    ReviewerService
	Statistics  StatisticsService
}

// NewServices создает новый экземпляр Services
func NewServices(store *repository.Store) *Services {
	return &Services{
		Team:        NewTeamService(store),
		User:        NewUserService(store, store, store, store),
		PullRequest: NewPullRequestService(store),
		Reviewer:    NewReviewerService(store, store, store, store, store),
		Statistics:  NewStatisticsService(store),
	}
}
