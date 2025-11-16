package repository

import (
	"context"

	"github.com/AtoyanMikhail/PRAssignmentService/internal/db"
	"github.com/AtoyanMikhail/PRAssignmentService/internal/models"
)

// --- TeamRepository implementation ---

func (r *PostgresRepository) Create(ctx context.Context, teamName string) (models.Team, error) {
	dbTeam, err := r.queries.CreateTeam(ctx, teamName)
	if err != nil {
		return models.Team{}, err
	}
	return models.TeamFromDB(dbTeam), nil
}

func (r *PostgresRepository) GetByName(ctx context.Context, teamName string) (models.Team, error) {
	dbTeam, err := r.queries.GetTeamByName(ctx, teamName)
	if err != nil {
		return models.Team{}, err
	}
	return models.TeamFromDB(dbTeam), nil
}

func (r *PostgresRepository) GetTeamByID(ctx context.Context, id int64) (models.Team, error) {
	dbTeam, err := r.queries.GetTeamByID(ctx, id)
	if err != nil {
		return models.Team{}, err
	}
	return models.TeamFromDB(dbTeam), nil
}

func (r *PostgresRepository) List(ctx context.Context) ([]models.Team, error) {
	dbTeams, err := r.queries.ListTeams(ctx)
	if err != nil {
		return nil, err
	}
	return models.TeamsFromDB(dbTeams), nil
}

func (r *PostgresRepository) Exists(ctx context.Context, teamName string) (bool, error) {
	return r.queries.TeamExists(ctx, teamName)
}

// --- UserRepository implementation ---

func (r *PostgresRepository) CreateUser(ctx context.Context, userID, username string, teamID int64, isActive bool) (models.User, error) {
	dbUser, err := r.queries.CreateUser(ctx, db.CreateUserParams{
		UserID:   userID,
		Username: username,
		TeamID:   teamID,
		IsActive: isActive,
	})
	if err != nil {
		return models.User{}, err
	}
	return models.UserFromDB(dbUser), nil
}

func (r *PostgresRepository) GetByUserID(ctx context.Context, userID string) (models.User, error) {
	dbUser, err := r.queries.GetUserByUserID(ctx, userID)
	if err != nil {
		return models.User{}, err
	}
	return models.UserFromDB(dbUser), nil
}

func (r *PostgresRepository) GetUserByID(ctx context.Context, id int64) (models.User, error) {
	dbUser, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		return models.User{}, err
	}
	return models.UserFromDB(dbUser), nil
}

func (r *PostgresRepository) Update(ctx context.Context, userID, username string, teamID int64, isActive bool) (models.User, error) {
	dbUser, err := r.queries.UpdateUser(ctx, db.UpdateUserParams{
		UserID:   userID,
		Username: username,
		TeamID:   teamID,
		IsActive: isActive,
	})
	if err != nil {
		return models.User{}, err
	}
	return models.UserFromDB(dbUser), nil
}

func (r *PostgresRepository) UpdateIsActive(ctx context.Context, userID string, isActive bool) (models.User, error) {
	dbUser, err := r.queries.UpdateUserIsActive(ctx, db.UpdateUserIsActiveParams{
		UserID:   userID,
		IsActive: isActive,
	})
	if err != nil {
		return models.User{}, err
	}
	return models.UserFromDB(dbUser), nil
}

func (r *PostgresRepository) ListByTeamID(ctx context.Context, teamID int64) ([]models.User, error) {
	dbUsers, err := r.queries.ListUsersByTeamID(ctx, teamID)
	if err != nil {
		return nil, err
	}
	return models.UsersFromDB(dbUsers), nil
}

func (r *PostgresRepository) ListActiveByTeamID(ctx context.Context, teamID int64) ([]models.User, error) {
	dbUsers, err := r.queries.ListActiveUsersByTeamID(ctx, teamID)
	if err != nil {
		return nil, err
	}
	return models.UsersFromDB(dbUsers), nil
}

func (r *PostgresRepository) ListActiveByTeamIDExcludingUser(ctx context.Context, teamID int64, excludeUserID string) ([]models.User, error) {
	dbUsers, err := r.queries.ListActiveUsersByTeamIDExcludingUser(ctx, db.ListActiveUsersByTeamIDExcludingUserParams{
		TeamID: teamID,
		UserID: excludeUserID,
	})
	if err != nil {
		return nil, err
	}
	return models.UsersFromDB(dbUsers), nil
}

func (r *PostgresRepository) GetWithTeam(ctx context.Context, userID string) (models.UserWithTeam, error) {
	dbRow, err := r.queries.GetUserWithTeam(ctx, userID)
	if err != nil {
		return models.UserWithTeam{}, err
	}
	return models.UserWithTeamFromDB(dbRow), nil
}

func (r *PostgresRepository) UserExists(ctx context.Context, userID string) (bool, error) {
	return r.queries.UserExists(ctx, userID)
}

func (r *PostgresRepository) DeactivateTeamUsers(ctx context.Context, teamID int64) ([]models.User, error) {
	dbUsers, err := r.queries.DeactivateTeamUsers(ctx, teamID)
	if err != nil {
		return nil, err
	}
	return models.UsersFromDB(dbUsers), nil
}
