package service

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/AtoyanMikhail/PRAssignmentService/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockPullRequestRepository - мок репозитория для тестирования
type MockPullRequestRepository struct {
	mock.Mock
}

func (m *MockPullRequestRepository) CreatePR(ctx context.Context, pullRequestID, pullRequestName, authorID string, status models.PullRequestStatus) (models.PullRequest, error) {
	args := m.Called(ctx, pullRequestID, pullRequestName, authorID, status)
	return args.Get(0).(models.PullRequest), args.Error(1)
}

func (m *MockPullRequestRepository) GetPullRequestByPRID(ctx context.Context, pullRequestID string) (models.PullRequest, error) {
	args := m.Called(ctx, pullRequestID)
	return args.Get(0).(models.PullRequest), args.Error(1)
}

func (m *MockPullRequestRepository) GetPullRequestByID(ctx context.Context, id int64) (models.PullRequest, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(models.PullRequest), args.Error(1)
}

func (m *MockPullRequestRepository) UpdateStatus(ctx context.Context, pullRequestID string, status models.PullRequestStatus) (models.PullRequest, error) {
	args := m.Called(ctx, pullRequestID, status)
	return args.Get(0).(models.PullRequest), args.Error(1)
}

func (m *MockPullRequestRepository) Merge(ctx context.Context, pullRequestID string) (models.PullRequest, error) {
	args := m.Called(ctx, pullRequestID)
	return args.Get(0).(models.PullRequest), args.Error(1)
}

func (m *MockPullRequestRepository) ListByStatus(ctx context.Context, status models.PullRequestStatus) ([]models.PullRequest, error) {
	args := m.Called(ctx, status)
	return args.Get(0).([]models.PullRequest), args.Error(1)
}

func (m *MockPullRequestRepository) PRExists(ctx context.Context, pullRequestID string) (bool, error) {
	args := m.Called(ctx, pullRequestID)
	return args.Bool(0), args.Error(1)
}

func TestPullRequestService_CreatePR(t *testing.T) {
	ctx := context.Background()

	t.Run("успешное создание PR", func(t *testing.T) {
		mockRepo := new(MockPullRequestRepository)
		service := NewPullRequestService(mockRepo)

		expectedPR := models.PullRequest{
			ID:              1,
			PullRequestID:   "PR-123",
			PullRequestName: "Feature: Add tests",
			AuthorID:        "user1",
			Status:          models.PullRequestStatusOpen,
			CreatedAt:       time.Now(),
		}

		mockRepo.On("PRExists", ctx, "PR-123").Return(false, nil)
		mockRepo.On("CreatePR", ctx, "PR-123", "Feature: Add tests", "user1", models.PullRequestStatusOpen).
			Return(expectedPR, nil)

		pr, err := service.CreatePR(ctx, "PR-123", "Feature: Add tests", "user1")

		assert.NoError(t, err)
		assert.Equal(t, expectedPR, pr)
		mockRepo.AssertExpectations(t)
	})

	t.Run("ошибка - PR уже существует", func(t *testing.T) {
		mockRepo := new(MockPullRequestRepository)
		service := NewPullRequestService(mockRepo)

		mockRepo.On("PRExists", ctx, "PR-123").Return(true, nil)

		pr, err := service.CreatePR(ctx, "PR-123", "Feature: Add tests", "user1")

		assert.Error(t, err)
		assert.Equal(t, ErrPullRequestAlreadyExists, err)
		assert.Equal(t, models.PullRequest{}, pr)
		mockRepo.AssertExpectations(t)
	})
}

func TestPullRequestService_GetPR(t *testing.T) {
	ctx := context.Background()

	t.Run("успешное получение PR", func(t *testing.T) {
		mockRepo := new(MockPullRequestRepository)
		service := NewPullRequestService(mockRepo)

		expectedPR := models.PullRequest{
			ID:              1,
			PullRequestID:   "PR-123",
			PullRequestName: "Feature: Add tests",
			AuthorID:        "user1",
			Status:          models.PullRequestStatusOpen,
			CreatedAt:       time.Now(),
		}

		mockRepo.On("GetPullRequestByPRID", ctx, "PR-123").Return(expectedPR, nil)

		pr, err := service.GetPR(ctx, "PR-123")

		assert.NoError(t, err)
		assert.Equal(t, expectedPR, pr)
		mockRepo.AssertExpectations(t)
	})

	t.Run("ошибка - PR не найден", func(t *testing.T) {
		mockRepo := new(MockPullRequestRepository)
		service := NewPullRequestService(mockRepo)

		mockRepo.On("GetPullRequestByPRID", ctx, "PR-999").Return(models.PullRequest{}, sql.ErrNoRows)

		pr, err := service.GetPR(ctx, "PR-999")

		assert.Error(t, err)
		assert.Equal(t, ErrPullRequestNotFound, err)
		assert.Equal(t, models.PullRequest{}, pr)
		mockRepo.AssertExpectations(t)
	})
}

func TestPullRequestService_UpdatePRStatus(t *testing.T) {
	ctx := context.Background()

	t.Run("успешное обновление статуса", func(t *testing.T) {
		mockRepo := new(MockPullRequestRepository)
		service := NewPullRequestService(mockRepo)

		existingPR := models.PullRequest{
			ID:              1,
			PullRequestID:   "PR-123",
			PullRequestName: "Feature: Add tests",
			AuthorID:        "user1",
			Status:          models.PullRequestStatusOpen,
			CreatedAt:       time.Now(),
		}

		updatedPR := existingPR
		updatedPR.Status = models.PullRequestStatusMerged

		mockRepo.On("GetPullRequestByPRID", ctx, "PR-123").Return(existingPR, nil)
		mockRepo.On("UpdateStatus", ctx, "PR-123", models.PullRequestStatusMerged).Return(updatedPR, nil)

		pr, err := service.UpdatePRStatus(ctx, "PR-123", models.PullRequestStatusMerged)

		assert.NoError(t, err)
		assert.Equal(t, models.PullRequestStatusMerged, pr.Status)
		mockRepo.AssertExpectations(t)
	})

	t.Run("ошибка - невалидный статус", func(t *testing.T) {
		mockRepo := new(MockPullRequestRepository)
		service := NewPullRequestService(mockRepo)

		existingPR := models.PullRequest{
			ID:            1,
			PullRequestID: "PR-123",
			Status:        models.PullRequestStatusOpen,
		}

		mockRepo.On("GetPullRequestByPRID", ctx, "PR-123").Return(existingPR, nil)

		pr, err := service.UpdatePRStatus(ctx, "PR-123", models.PullRequestStatus("INVALID"))

		assert.Error(t, err)
		assert.Equal(t, ErrInvalidStatus, err)
		assert.Equal(t, models.PullRequest{}, pr)
		mockRepo.AssertExpectations(t)
	})

	t.Run("ошибка - PR не найден", func(t *testing.T) {
		mockRepo := new(MockPullRequestRepository)
		service := NewPullRequestService(mockRepo)

		mockRepo.On("GetPullRequestByPRID", ctx, "PR-999").Return(models.PullRequest{}, sql.ErrNoRows)

		pr, err := service.UpdatePRStatus(ctx, "PR-999", models.PullRequestStatusMerged)

		assert.Error(t, err)
		assert.Equal(t, ErrPullRequestNotFound, err)
		assert.Equal(t, models.PullRequest{}, pr)
		mockRepo.AssertExpectations(t)
	})
}

func TestPullRequestService_MergePR(t *testing.T) {
	ctx := context.Background()

	t.Run("успешный merge PR", func(t *testing.T) {
		mockRepo := new(MockPullRequestRepository)
		service := NewPullRequestService(mockRepo)

		existingPR := models.PullRequest{
			ID:            1,
			PullRequestID: "PR-123",
			Status:        models.PullRequestStatusOpen,
			CreatedAt:     time.Now(),
		}

		mergedTime := time.Now()
		mergedPR := existingPR
		mergedPR.Status = models.PullRequestStatusMerged
		mergedPR.MergedAt = &mergedTime

		mockRepo.On("GetPullRequestByPRID", ctx, "PR-123").Return(existingPR, nil)
		mockRepo.On("Merge", ctx, "PR-123").Return(mergedPR, nil)

		pr, err := service.MergePR(ctx, "PR-123")

		assert.NoError(t, err)
		assert.Equal(t, models.PullRequestStatusMerged, pr.Status)
		assert.NotNil(t, pr.MergedAt)
		mockRepo.AssertExpectations(t)
	})

	t.Run("ошибка - PR не найден", func(t *testing.T) {
		mockRepo := new(MockPullRequestRepository)
		service := NewPullRequestService(mockRepo)

		mockRepo.On("GetPullRequestByPRID", ctx, "PR-999").Return(models.PullRequest{}, sql.ErrNoRows)

		pr, err := service.MergePR(ctx, "PR-999")

		assert.Error(t, err)
		assert.Equal(t, ErrPullRequestNotFound, err)
		assert.Equal(t, models.PullRequest{}, pr)
		mockRepo.AssertExpectations(t)
	})
}

func TestPullRequestService_ListOpenPRs(t *testing.T) {
	ctx := context.Background()

	t.Run("успешное получение открытых PR", func(t *testing.T) {
		mockRepo := new(MockPullRequestRepository)
		service := NewPullRequestService(mockRepo)

		expectedPRs := []models.PullRequest{
			{ID: 1, PullRequestID: "PR-1", Status: models.PullRequestStatusOpen},
			{ID: 2, PullRequestID: "PR-2", Status: models.PullRequestStatusOpen},
		}

		mockRepo.On("ListByStatus", ctx, models.PullRequestStatusOpen).Return(expectedPRs, nil)

		prs, err := service.ListOpenPRs(ctx)

		assert.NoError(t, err)
		assert.Len(t, prs, 2)
		mockRepo.AssertExpectations(t)
	})
}

func TestPullRequestService_ListPRsByStatus(t *testing.T) {
	ctx := context.Background()

	t.Run("успешное получение PR по статусу", func(t *testing.T) {
		mockRepo := new(MockPullRequestRepository)
		service := NewPullRequestService(mockRepo)

		expectedPRs := []models.PullRequest{
			{ID: 1, PullRequestID: "PR-1", Status: models.PullRequestStatusMerged},
			{ID: 2, PullRequestID: "PR-2", Status: models.PullRequestStatusMerged},
		}

		mockRepo.On("ListByStatus", ctx, models.PullRequestStatusMerged).Return(expectedPRs, nil)

		prs, err := service.ListPRsByStatus(ctx, models.PullRequestStatusMerged)

		assert.NoError(t, err)
		assert.Len(t, prs, 2)
		mockRepo.AssertExpectations(t)
	})

	t.Run("ошибка - невалидный статус", func(t *testing.T) {
		mockRepo := new(MockPullRequestRepository)
		service := NewPullRequestService(mockRepo)

		prs, err := service.ListPRsByStatus(ctx, models.PullRequestStatus("INVALID"))

		assert.Error(t, err)
		assert.Equal(t, ErrInvalidStatus, err)
		assert.Nil(t, prs)
		mockRepo.AssertNotCalled(t, "ListByStatus")
	})
}
