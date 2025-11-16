package service

import (
	"context"
	"database/sql"
"errors"
"github.com/jackc/pgx/v5"

	"github.com/AtoyanMikhail/PRAssignmentService/internal/models"
	"github.com/AtoyanMikhail/PRAssignmentService/internal/repository"
)

// PullRequestServiceImpl реализует PullRequestService
type PullRequestServiceImpl struct {
	repo repository.PullRequestRepository
}

// NewPullRequestService создает новый PullRequestService
func NewPullRequestService(repo repository.PullRequestRepository) PullRequestService {
	return &PullRequestServiceImpl{
		repo: repo,
	}
}

// CreatePR создает новый Pull Request
func (s *PullRequestServiceImpl) CreatePR(ctx context.Context, pullRequestID, pullRequestName, authorID string) (models.PullRequest, error) {
	// Проверка существования PR
	exists, err := s.repo.PRExists(ctx, pullRequestID)
	if err != nil {
		return models.PullRequest{}, err
	}
	if exists {
		return models.PullRequest{}, ErrPullRequestAlreadyExists
	}

	// Создание PR со статусом OPEN
	pr, err := s.repo.CreatePR(ctx, pullRequestID, pullRequestName, authorID, models.PullRequestStatusOpen)
	if err != nil {
		return models.PullRequest{}, err
	}

	return pr, nil
}

// GetPR возвращает Pull Request по ID
func (s *PullRequestServiceImpl) GetPR(ctx context.Context, pullRequestID string) (models.PullRequest, error) {
	pr, err := s.repo.GetPullRequestByPRID(ctx, pullRequestID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return models.PullRequest{}, ErrPullRequestNotFound
		}
		return models.PullRequest{}, err
	}
	return pr, nil
}

// UpdatePRStatus обновляет статус Pull Request
func (s *PullRequestServiceImpl) UpdatePRStatus(ctx context.Context, pullRequestID string, status models.PullRequestStatus) (models.PullRequest, error) {
	// Проверка существования PR
	_, err := s.GetPR(ctx, pullRequestID)
	if err != nil {
		return models.PullRequest{}, err
	}

	// Валидация статуса
	if !status.IsValid() {
		return models.PullRequest{}, ErrInvalidStatus
	}

	// Обновление статуса
	pr, err := s.repo.UpdateStatus(ctx, pullRequestID, status)
	if err != nil {
		return models.PullRequest{}, err
	}

	return pr, nil
}

// MergePR помечает Pull Request как merged
func (s *PullRequestServiceImpl) MergePR(ctx context.Context, pullRequestID string) (models.PullRequest, error) {
	// Проверка существования PR
	_, err := s.GetPR(ctx, pullRequestID)
	if err != nil {
		return models.PullRequest{}, err
	}

	// Merge PR (устанавливает статус MERGED и merged_at)
	pr, err := s.repo.Merge(ctx, pullRequestID)
	if err != nil {
		return models.PullRequest{}, err
	}

	return pr, nil
}

// ListOpenPRs возвращает список открытых Pull Request'ов
func (s *PullRequestServiceImpl) ListOpenPRs(ctx context.Context) ([]models.PullRequest, error) {
	return s.repo.ListByStatus(ctx, models.PullRequestStatusOpen)
}

// ListPRsByStatus возвращает Pull Request'ы по статусу
func (s *PullRequestServiceImpl) ListPRsByStatus(ctx context.Context, status models.PullRequestStatus) ([]models.PullRequest, error) {
	if !status.IsValid() {
		return nil, ErrInvalidStatus
	}
	return s.repo.ListByStatus(ctx, status)
}
