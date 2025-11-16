package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/AtoyanMikhail/PRAssignmentService/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTeamRepository - мок репозитория для тестирования
type MockTeamRepository struct {
	mock.Mock
}

func (m *MockTeamRepository) Create(ctx context.Context, teamName string) (models.Team, error) {
	args := m.Called(ctx, teamName)
	return args.Get(0).(models.Team), args.Error(1)
}

func (m *MockTeamRepository) GetByName(ctx context.Context, teamName string) (models.Team, error) {
	args := m.Called(ctx, teamName)
	return args.Get(0).(models.Team), args.Error(1)
}

func (m *MockTeamRepository) GetTeamByID(ctx context.Context, id int64) (models.Team, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(models.Team), args.Error(1)
}

func (m *MockTeamRepository) List(ctx context.Context) ([]models.Team, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.Team), args.Error(1)
}

func (m *MockTeamRepository) Exists(ctx context.Context, teamName string) (bool, error) {
	args := m.Called(ctx, teamName)
	return args.Bool(0), args.Error(1)
}

func TestTeamService_CreateTeam(t *testing.T) {
	ctx := context.Background()

	t.Run("успешное создание команды", func(t *testing.T) {
		mockRepo := new(MockTeamRepository)
		service := NewTeamService(mockRepo)

		expectedTeam := models.Team{
			ID:        1,
			TeamName:  "backend",
			CreatedAt: time.Now(),
		}

		mockRepo.On("Exists", ctx, "backend").Return(false, nil)
		mockRepo.On("Create", ctx, "backend").Return(expectedTeam, nil)

		team, err := service.CreateTeam(ctx, "backend")

		assert.NoError(t, err)
		assert.Equal(t, expectedTeam, team)
		mockRepo.AssertExpectations(t)
	})

	t.Run("ошибка - команда уже существует", func(t *testing.T) {
		mockRepo := new(MockTeamRepository)
		service := NewTeamService(mockRepo)

		mockRepo.On("Exists", ctx, "backend").Return(true, nil)

		team, err := service.CreateTeam(ctx, "backend")

		assert.Error(t, err)
		assert.Equal(t, ErrTeamAlreadyExists, err)
		assert.Equal(t, models.Team{}, team)
		mockRepo.AssertExpectations(t)
	})

	t.Run("ошибка при проверке существования", func(t *testing.T) {
		mockRepo := new(MockTeamRepository)
		service := NewTeamService(mockRepo)

		dbError := errors.New("database error")
		mockRepo.On("Exists", ctx, "backend").Return(false, dbError)

		team, err := service.CreateTeam(ctx, "backend")

		assert.Error(t, err)
		assert.Equal(t, dbError, err)
		assert.Equal(t, models.Team{}, team)
		mockRepo.AssertExpectations(t)
	})

	t.Run("ошибка при создании команды", func(t *testing.T) {
		mockRepo := new(MockTeamRepository)
		service := NewTeamService(mockRepo)

		dbError := errors.New("insert error")
		mockRepo.On("Exists", ctx, "backend").Return(false, nil)
		mockRepo.On("Create", ctx, "backend").Return(models.Team{}, dbError)

		team, err := service.CreateTeam(ctx, "backend")

		assert.Error(t, err)
		assert.Equal(t, dbError, err)
		assert.Equal(t, models.Team{}, team)
		mockRepo.AssertExpectations(t)
	})
}

func TestTeamService_GetTeam(t *testing.T) {
	ctx := context.Background()

	t.Run("успешное получение команды", func(t *testing.T) {
		mockRepo := new(MockTeamRepository)
		service := NewTeamService(mockRepo)

		expectedTeam := models.Team{
			ID:        1,
			TeamName:  "backend",
			CreatedAt: time.Now(),
		}

		mockRepo.On("GetByName", ctx, "backend").Return(expectedTeam, nil)

		team, err := service.GetTeam(ctx, "backend")

		assert.NoError(t, err)
		assert.Equal(t, expectedTeam, team)
		mockRepo.AssertExpectations(t)
	})

	t.Run("ошибка - команда не найдена", func(t *testing.T) {
		mockRepo := new(MockTeamRepository)
		service := NewTeamService(mockRepo)

		mockRepo.On("GetByName", ctx, "nonexistent").Return(models.Team{}, sql.ErrNoRows)

		team, err := service.GetTeam(ctx, "nonexistent")

		assert.Error(t, err)
		assert.Equal(t, ErrTeamNotFound, err)
		assert.Equal(t, models.Team{}, team)
		mockRepo.AssertExpectations(t)
	})

	t.Run("ошибка БД", func(t *testing.T) {
		mockRepo := new(MockTeamRepository)
		service := NewTeamService(mockRepo)

		dbError := errors.New("database error")
		mockRepo.On("GetByName", ctx, "backend").Return(models.Team{}, dbError)

		team, err := service.GetTeam(ctx, "backend")

		assert.Error(t, err)
		assert.Equal(t, dbError, err)
		assert.Equal(t, models.Team{}, team)
		mockRepo.AssertExpectations(t)
	})
}

func TestTeamService_GetTeamByID(t *testing.T) {
	ctx := context.Background()

	t.Run("успешное получение команды по ID", func(t *testing.T) {
		mockRepo := new(MockTeamRepository)
		service := NewTeamService(mockRepo)

		expectedTeam := models.Team{
			ID:        1,
			TeamName:  "backend",
			CreatedAt: time.Now(),
		}

		mockRepo.On("GetTeamByID", ctx, int64(1)).Return(expectedTeam, nil)

		team, err := service.GetTeamByID(ctx, 1)

		assert.NoError(t, err)
		assert.Equal(t, expectedTeam, team)
		mockRepo.AssertExpectations(t)
	})

	t.Run("ошибка - команда не найдена", func(t *testing.T) {
		mockRepo := new(MockTeamRepository)
		service := NewTeamService(mockRepo)

		mockRepo.On("GetTeamByID", ctx, int64(999)).Return(models.Team{}, sql.ErrNoRows)

		team, err := service.GetTeamByID(ctx, 999)

		assert.Error(t, err)
		assert.Equal(t, ErrTeamNotFound, err)
		assert.Equal(t, models.Team{}, team)
		mockRepo.AssertExpectations(t)
	})
}

func TestTeamService_ListTeams(t *testing.T) {
	ctx := context.Background()

	t.Run("успешное получение списка команд", func(t *testing.T) {
		mockRepo := new(MockTeamRepository)
		service := NewTeamService(mockRepo)

		expectedTeams := []models.Team{
			{ID: 1, TeamName: "backend", CreatedAt: time.Now()},
			{ID: 2, TeamName: "frontend", CreatedAt: time.Now()},
		}

		mockRepo.On("List", ctx).Return(expectedTeams, nil)

		teams, err := service.ListTeams(ctx)

		assert.NoError(t, err)
		assert.Equal(t, expectedTeams, teams)
		assert.Len(t, teams, 2)
		mockRepo.AssertExpectations(t)
	})

	t.Run("пустой список команд", func(t *testing.T) {
		mockRepo := new(MockTeamRepository)
		service := NewTeamService(mockRepo)

		mockRepo.On("List", ctx).Return([]models.Team{}, nil)

		teams, err := service.ListTeams(ctx)

		assert.NoError(t, err)
		assert.Empty(t, teams)
		mockRepo.AssertExpectations(t)
	})

	t.Run("ошибка БД", func(t *testing.T) {
		mockRepo := new(MockTeamRepository)
		service := NewTeamService(mockRepo)

		dbError := errors.New("database error")
		mockRepo.On("List", ctx).Return([]models.Team(nil), dbError)

		teams, err := service.ListTeams(ctx)

		assert.Error(t, err)
		assert.Equal(t, dbError, err)
		assert.Nil(t, teams)
		mockRepo.AssertExpectations(t)
	})
}
