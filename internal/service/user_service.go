package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/AtoyanMikhail/PRAssignmentService/internal/models"
	"github.com/AtoyanMikhail/PRAssignmentService/internal/repository"
	"github.com/jackc/pgx/v5"
)

// UserServiceImpl реализует UserService
type UserServiceImpl struct {
	userRepo     repository.UserRepository
	teamRepo     repository.TeamRepository
	reviewerRepo repository.PRReviewerRepository
	store        *repository.Store
}

// NewUserService создает новый UserService
func NewUserService(
	userRepo repository.UserRepository,
	teamRepo repository.TeamRepository,
	reviewerRepo repository.PRReviewerRepository,
	store *repository.Store,
) UserService {
	return &UserServiceImpl{
		userRepo:     userRepo,
		teamRepo:     teamRepo,
		reviewerRepo: reviewerRepo,
		store:        store,
	}
}

// CreateUser создает нового пользователя
func (s *UserServiceImpl) CreateUser(ctx context.Context, userID, username string, teamID int64) (models.User, error) {
	exists, err := s.userRepo.UserExists(ctx, userID)
	if err != nil {
		return models.User{}, err
	}
	if exists {
		return models.User{}, ErrUserAlreadyExists
	}

	_, err = s.teamRepo.GetTeamByID(ctx, teamID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, ErrTeamNotFound
		}
		return models.User{}, err
	}

	user, err := s.userRepo.CreateUser(ctx, userID, username, teamID, true)
	if err != nil {
		return models.User{}, err
	}

	return user, nil
}

// GetUser возвращает пользователя по userID
func (s *UserServiceImpl) GetUser(ctx context.Context, userID string) (models.User, error) {
	user, err := s.userRepo.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, ErrUserNotFound
		}
		return models.User{}, err
	}
	return user, nil
}

// GetUserWithTeam возвращает пользователя с информацией о команде
func (s *UserServiceImpl) GetUserWithTeam(ctx context.Context, userID string) (models.UserWithTeam, error) {
	user, err := s.userRepo.GetWithTeam(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return models.UserWithTeam{}, ErrUserNotFound
		}
		return models.UserWithTeam{}, err
	}
	return user, nil
}

// UpdateUser обновляет данные пользователя
func (s *UserServiceImpl) UpdateUser(ctx context.Context, userID, username string, teamID int64, isActive bool) (models.User, error) {
	_, err := s.GetUser(ctx, userID)
	if err != nil {
		return models.User{}, err
	}

	_, err = s.teamRepo.GetTeamByID(ctx, teamID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, ErrTeamNotFound
		}
		return models.User{}, err
	}

	user, err := s.userRepo.Update(ctx, userID, username, teamID, isActive)
	if err != nil {
		return models.User{}, err
	}

	return user, nil
}

// DeactivateUser деактивирует пользователя
func (s *UserServiceImpl) DeactivateUser(ctx context.Context, userID string) (models.User, error) {
	_, err := s.GetUser(ctx, userID)
	if err != nil {
		return models.User{}, err
	}

	// Деактивация пользователя
	user, err := s.userRepo.UpdateIsActive(ctx, userID, false)
	if err != nil {
		return models.User{}, err
	}

	return user, nil
}

// ActivateUser активирует пользователя
func (s *UserServiceImpl) ActivateUser(ctx context.Context, userID string) (models.User, error) {
	_, err := s.GetUser(ctx, userID)
	if err != nil {
		return models.User{}, err
	}

	// Активация пользователя
	user, err := s.userRepo.UpdateIsActive(ctx, userID, true)
	if err != nil {
		return models.User{}, err
	}

	return user, nil
}

// ListTeamUsers возвращает всех пользователей команды
func (s *UserServiceImpl) ListTeamUsers(ctx context.Context, teamID int64) ([]models.User, error) {
	_, err := s.teamRepo.GetTeamByID(ctx, teamID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTeamNotFound
		}
		return nil, err
	}

	users, err := s.userRepo.ListByTeamID(ctx, teamID)
	if err != nil {
		return nil, err
	}

	return users, nil
}

// ListActiveTeamUsers возвращает активных пользователей команды
func (s *UserServiceImpl) ListActiveTeamUsers(ctx context.Context, teamID int64) ([]models.User, error) {
	_, err := s.teamRepo.GetTeamByID(ctx, teamID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTeamNotFound
		}
		return nil, err
	}

	users, err := s.userRepo.ListActiveByTeamID(ctx, teamID)
	if err != nil {
		return nil, err
	}

	return users, nil
}

// DeactivateTeamUsers деактивирует всех пользователей команды и перераспределяет их PR
// Возвращает количество деактивированных пользователей и количество переназначенных PR
func (s *UserServiceImpl) DeactivateTeamUsers(ctx context.Context, teamID int64) (int, int, error) {
	_, err := s.teamRepo.GetTeamByID(ctx, teamID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return 0, 0, ErrTeamNotFound
		}
		return 0, 0, err
	}

	var deactivatedCount int
	var reassignedCount int

	// Выполнение в транзакции для атомарности
	err = s.store.ExecTx(ctx, func(txRepo *repository.PostgresRepository) error {
		// 1. Деактивировать всех пользователей команды
		deactivatedUsers, err := txRepo.DeactivateTeamUsers(ctx, teamID)
		if err != nil {
			return fmt.Errorf("failed to deactivate team users: %w", err)
		}

		deactivatedCount = len(deactivatedUsers)

		if len(deactivatedUsers) == 0 {
			return nil // Нет пользователей для деактивации
		}

		// 2. Получить все открытые PR с неактивными ревьюерами
		inactiveReviewerInfos, err := txRepo.GetOpenPRsWithInactiveReviewers(ctx)
		if err != nil {
			return fmt.Errorf("failed to get PRs with inactive reviewers: %w", err)
		}

		// Группировка по PR для batch операций
		prToInactiveUsers := make(map[string][]string)
		for _, info := range inactiveReviewerInfos {
			prToInactiveUsers[info.PullRequestID] = append(prToInactiveUsers[info.PullRequestID], info.InactiveReviewerID)
		}

		// 3. Для каждого PR удалить неактивных ревьюеров
		for prID, inactiveUserIDs := range prToInactiveUsers {
			err := txRepo.RemoveInactiveReviewers(ctx, prID, inactiveUserIDs)
			if err != nil {
				return fmt.Errorf("failed to remove inactive reviewers from PR %s: %w", prID, err)
			}
		}

		// 4. Получить нагрузку пользователей для балансировки
		workloads, err := txRepo.GetUserWorkload(ctx)
		if err != nil {
			return fmt.Errorf("failed to get user workload: %w", err)
		}

		workloadMap := make(map[string]int64)
		for _, w := range workloads {
			workloadMap[w.UserID] = w.OpenReviewsCount
		}

		// 5. Для PR с недостаточным количеством ревьюеров назначить новых
		for prID := range prToInactiveUsers {
			// Проверить текущее количество ревьюеров
			currentCount, err := txRepo.Count(ctx, prID)
			if err != nil {
				return fmt.Errorf("failed to count reviewers for PR %s: %w", prID, err)
			}

			// Если ревьюеров меньше 2, назначить недостающих
			if currentCount < 2 {
				needed := 2 - int(currentCount)

				// Получить информацию о PR для определения команды автора
				pr, err := txRepo.GetPullRequestByPRID(ctx, prID)
				if err != nil {
					return fmt.Errorf("failed to get PR %s: %w", prID, err)
				}

				// Получить автора для определения его команды
				author, err := txRepo.GetByUserID(ctx, pr.AuthorID)
				if err != nil {
					return fmt.Errorf("failed to get author %s: %w", pr.AuthorID, err)
				}

				// Получить активных пользователей команды (исключая автора)
				activeUsers, err := txRepo.ListActiveByTeamIDExcludingUser(ctx, author.TeamID, pr.AuthorID)
				if err != nil {
					return fmt.Errorf("failed to get active users for team %d: %w", author.TeamID, err)
				}

				if len(activeUsers) == 0 {
					continue // Нет доступных ревьюеров для этого PR
				}

				// Исключить уже назначенных ревьюеров
				assignedReviewers, err := txRepo.GetReviewersByPRID(ctx, prID)
				if err != nil {
					return fmt.Errorf("failed to get assigned reviewers for PR %s: %w", prID, err)
				}

				assignedMap := make(map[string]bool)
				for _, r := range assignedReviewers {
					assignedMap[r.UserID] = true
				}

				// Фильтрация доступных пользователей
				var availableUsers []models.User
				for _, user := range activeUsers {
					if !assignedMap[user.UserID] {
						availableUsers = append(availableUsers, user)
					}
				}

				// Выбор пользователей с минимальной нагрузкой
				selectedUsers := selectUsersWithMinWorkload(availableUsers, workloadMap, needed)

				// Назначение выбранных ревьюеров
				for _, user := range selectedUsers {
					_, err := txRepo.Add(ctx, prID, user.UserID)
					if err != nil {
						return fmt.Errorf("failed to assign reviewer %s to PR %s: %w", user.UserID, prID, err)
					}
					reassignedCount++ // Увеличиваем счетчик переназначенных PR
				}
			}
		}

		return nil
	})

	if err != nil {
		return 0, 0, err
	}

	return deactivatedCount, reassignedCount, nil
}
