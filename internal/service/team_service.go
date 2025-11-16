package service

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jackc/pgx/v5"

	"github.com/AtoyanMikhail/PRAssignmentService/internal/models"
	"github.com/AtoyanMikhail/PRAssignmentService/internal/repository"
)

// TeamServiceImpl реализует TeamService
type TeamServiceImpl struct {
	repo repository.TeamRepository
}

// NewTeamService создает новый TeamService
func NewTeamService(repo repository.TeamRepository) TeamService {
	return &TeamServiceImpl{
		repo: repo,
	}
}

// CreateTeam создает новую команду
func (s *TeamServiceImpl) CreateTeam(ctx context.Context, teamName string) (models.Team, error) {
	exists, err := s.repo.Exists(ctx, teamName)
	if err != nil {
		return models.Team{}, err
	}
	if exists {
		return models.Team{}, ErrTeamAlreadyExists
	}

	team, err := s.repo.Create(ctx, teamName)
	if err != nil {
		return models.Team{}, err
	}

	return team, nil
}

// GetTeam возвращает команду по имени
func (s *TeamServiceImpl) GetTeam(ctx context.Context, teamName string) (models.Team, error) {
	team, err := s.repo.GetByName(ctx, teamName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return models.Team{}, ErrTeamNotFound
		}
		return models.Team{}, err
	}
	return team, nil
}

// GetTeamByID возвращает команду по ID
func (s *TeamServiceImpl) GetTeamByID(ctx context.Context, id int64) (models.Team, error) {
	team, err := s.repo.GetTeamByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return models.Team{}, ErrTeamNotFound
		}
		return models.Team{}, err
	}
	return team, nil
}

// ListTeams возвращает список всех команд
func (s *TeamServiceImpl) ListTeams(ctx context.Context) ([]models.Team, error) {
	teams, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	return teams, nil
}
