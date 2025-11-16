package service

import (
	"context"

	"github.com/AtoyanMikhail/PRAssignmentService/internal/models"
	"github.com/AtoyanMikhail/PRAssignmentService/internal/repository"
)

// StatisticsServiceImpl реализует StatisticsService
type StatisticsServiceImpl struct {
	repo repository.StatisticsRepository
}

// NewStatisticsService создает новый StatisticsService
func NewStatisticsService(repo repository.StatisticsRepository) StatisticsService {
	return &StatisticsServiceImpl{
		repo: repo,
	}
}

// GetAssignmentStats возвращает статистику по назначениям ревьюеров
func (s *StatisticsServiceImpl) GetAssignmentStats(ctx context.Context) ([]models.AssignmentStats, error) {
	return s.repo.GetAssignmentStats(ctx)
}

// GetPRStats возвращает статистику по Pull Request'ам
func (s *StatisticsServiceImpl) GetPRStats(ctx context.Context) ([]models.PRStats, error) {
	return s.repo.GetPRStats(ctx)
}

// GetTeamStats возвращает статистику по командам
func (s *StatisticsServiceImpl) GetTeamStats(ctx context.Context) ([]models.TeamStats, error) {
	return s.repo.GetTeamStats(ctx)
}

// GetUserWorkload возвращает нагрузку пользователей
func (s *StatisticsServiceImpl) GetUserWorkload(ctx context.Context) ([]models.UserWorkload, error) {
	return s.repo.GetUserWorkload(ctx)
}
