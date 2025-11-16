package service

import (
	"context"
	"database/sql"
"errors"
"github.com/jackc/pgx/v5"
	"fmt"

	"github.com/AtoyanMikhail/PRAssignmentService/internal/models"
	"github.com/AtoyanMikhail/PRAssignmentService/internal/repository"
)

// ReviewerServiceImpl реализует ReviewerService
type ReviewerServiceImpl struct {
	reviewerRepo repository.PRReviewerRepository
	userRepo     repository.UserRepository
	prRepo       repository.PullRequestRepository
	statsRepo    repository.StatisticsRepository
	store        *repository.Store
}

// NewReviewerService создает новый ReviewerService
func NewReviewerService(
	reviewerRepo repository.PRReviewerRepository,
	userRepo repository.UserRepository,
	prRepo repository.PullRequestRepository,
	statsRepo repository.StatisticsRepository,
	store *repository.Store,
) ReviewerService {
	return &ReviewerServiceImpl{
		reviewerRepo: reviewerRepo,
		userRepo:     userRepo,
		prRepo:       prRepo,
		statsRepo:    statsRepo,
		store:        store,
	}
}

// AssignReviewer назначает ревьюера на Pull Request
func (s *ReviewerServiceImpl) AssignReviewer(ctx context.Context, pullRequestID, userID string) (models.PRReviewer, error) {
	// Проверка существования PR
	pr, err := s.prRepo.GetPullRequestByPRID(ctx, pullRequestID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return models.PRReviewer{}, ErrPullRequestNotFound
		}
		return models.PRReviewer{}, err
	}

	// Проверка, что PR открыт
	if pr.Status != models.PullRequestStatusOpen {
		return models.PRReviewer{}, fmt.Errorf("cannot assign reviewer to non-open PR")
	}

	// Проверка существования пользователя
	user, err := s.userRepo.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return models.PRReviewer{}, ErrUserNotFound
		}
		return models.PRReviewer{}, err
	}

	// Проверка, что пользователь активен
	if !user.IsActive {
		return models.PRReviewer{}, ErrUserInactive
	}

	// Проверка, что пользователь не является автором PR
	if pr.AuthorID == userID {
		return models.PRReviewer{}, ErrCannotAssignAuthor
	}

	// Проверка, что ревьюер еще не назначен
	isAssigned, err := s.reviewerRepo.IsUserAssigned(ctx, pullRequestID, userID)
	if err != nil {
		return models.PRReviewer{}, err
	}
	if isAssigned {
		return models.PRReviewer{}, ErrReviewerAlreadyAssigned
	}

	// Назначение ревьюера
	reviewer, err := s.reviewerRepo.Add(ctx, pullRequestID, userID)
	if err != nil {
		return models.PRReviewer{}, err
	}

	return reviewer, nil
}

// AssignReviewers назначает несколько ревьюеров на Pull Request
func (s *ReviewerServiceImpl) AssignReviewers(ctx context.Context, pullRequestID string, userIDs []string) ([]models.PRReviewer, error) {
	var reviewers []models.PRReviewer

	// Выполнение в транзакции для атомарности
	err := s.store.ExecTx(ctx, func(txRepo *repository.PostgresRepository) error {
		for _, userID := range userIDs {
			// Используем репозиторий из транзакции
			reviewerService := &ReviewerServiceImpl{
				reviewerRepo: txRepo,
				userRepo:     txRepo,
				prRepo:       txRepo,
				statsRepo:    txRepo,
				store:        s.store,
			}

			reviewer, err := reviewerService.AssignReviewer(ctx, pullRequestID, userID)
			if err != nil {
				return err
			}
			reviewers = append(reviewers, reviewer)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return reviewers, nil
}

// RemoveReviewer удаляет ревьюера с Pull Request
func (s *ReviewerServiceImpl) RemoveReviewer(ctx context.Context, pullRequestID, userID string) error {
	// Проверка существования PR
	_, err := s.prRepo.GetPullRequestByPRID(ctx, pullRequestID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return ErrPullRequestNotFound
		}
		return err
	}

	// Проверка, что ревьюер назначен
	isAssigned, err := s.reviewerRepo.IsUserAssigned(ctx, pullRequestID, userID)
	if err != nil {
		return err
	}
	if !isAssigned {
		return ErrReviewerNotAssigned
	}

	// Удаление ревьюера
	return s.reviewerRepo.Remove(ctx, pullRequestID, userID)
}

// ReplaceReviewer заменяет одного ревьюера другим
func (s *ReviewerServiceImpl) ReplaceReviewer(ctx context.Context, pullRequestID, oldUserID, newUserID string) error {
	// Выполнение в транзакции для атомарности
	return s.store.ExecTx(ctx, func(txRepo *repository.PostgresRepository) error {
		// Проверка существования PR
		pr, err := txRepo.GetPullRequestByPRID(ctx, pullRequestID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
				return ErrPullRequestNotFound
			}
			return err
		}

		// Проверка, что PR открыт
		if pr.Status != models.PullRequestStatusOpen {
			return fmt.Errorf("cannot replace reviewer in non-open PR")
		}

		// Проверка, что старый ревьюер назначен
		isAssigned, err := txRepo.IsUserAssigned(ctx, pullRequestID, oldUserID)
		if err != nil {
			return err
		}
		if !isAssigned {
			return ErrReviewerNotAssigned
		}

		// Проверка существования нового пользователя
		newUser, err := txRepo.GetByUserID(ctx, newUserID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
				return ErrUserNotFound
			}
			return err
		}

		// Проверка, что новый пользователь активен
		if !newUser.IsActive {
			return ErrUserInactive
		}

		// Проверка, что новый пользователь не является автором PR
		if pr.AuthorID == newUserID {
			return ErrCannotAssignAuthor
		}

		// Проверка, что новый ревьюер еще не назначен
		isNewAssigned, err := txRepo.IsUserAssigned(ctx, pullRequestID, newUserID)
		if err != nil {
			return err
		}
		if isNewAssigned {
			return ErrReviewerAlreadyAssigned
		}

		// Замена ревьюера
		return txRepo.Replace(ctx, pullRequestID, oldUserID, newUserID)
	})
}

// GetPRReviewers возвращает список ревьюеров для Pull Request
func (s *ReviewerServiceImpl) GetPRReviewers(ctx context.Context, pullRequestID string) ([]models.ReviewerInfo, error) {
	// Проверка существования PR
	_, err := s.prRepo.GetPullRequestByPRID(ctx, pullRequestID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPullRequestNotFound
		}
		return nil, err
	}

	reviewers, err := s.reviewerRepo.GetReviewersByPRID(ctx, pullRequestID)
	if err != nil {
		return nil, err
	}

	return reviewers, nil
}

// GetUserPRs возвращает Pull Request'ы, где пользователь является ревьюером
func (s *ReviewerServiceImpl) GetUserPRs(ctx context.Context, userID string) ([]models.PullRequestShort, error) {
	// Проверка существования пользователя
	_, err := s.userRepo.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	prs, err := s.reviewerRepo.GetPRsByReviewerUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return prs, nil
}

// AutoAssignReviewers автоматически назначает ревьюеров на Pull Request
func (s *ReviewerServiceImpl) AutoAssignReviewers(ctx context.Context, pullRequestID string, count int) ([]models.PRReviewer, error) {
	// Проверка существования PR
	pr, err := s.prRepo.GetPullRequestByPRID(ctx, pullRequestID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPullRequestNotFound
		}
		return nil, err
	}

	// Получение автора PR для определения его команды
	author, err := s.userRepo.GetByUserID(ctx, pr.AuthorID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	// Получение нагрузки всех пользователей для балансировки
	workloads, err := s.statsRepo.GetUserWorkload(ctx)
	if err != nil {
		return nil, err
	}

	// Создание мапы нагрузки для быстрого доступа
	workloadMap := make(map[string]int64)
	for _, w := range workloads {
		workloadMap[w.UserID] = w.OpenReviewsCount
	}

	// Получение активных пользователей команды (исключая автора)
	activeUsers, err := s.userRepo.ListActiveByTeamIDExcludingUser(ctx, author.TeamID, pr.AuthorID)
	if err != nil {
		return nil, err
	}

	if len(activeUsers) == 0 {
		return nil, ErrNoActiveReviewers
	}

	// Выбор пользователей с минимальной нагрузкой
	selectedUsers := selectUsersWithMinWorkload(activeUsers, workloadMap, count)

	// Назначение выбранных ревьюеров
	var reviewers []models.PRReviewer
	err = s.store.ExecTx(ctx, func(txRepo *repository.PostgresRepository) error {
		for _, user := range selectedUsers {
			reviewer, err := txRepo.Add(ctx, pullRequestID, user.UserID)
			if err != nil {
				return err
			}
			reviewers = append(reviewers, reviewer)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return reviewers, nil
}

// ReassignFromInactiveReviewers переназначает ревьюеров с неактивных на активных
func (s *ReviewerServiceImpl) ReassignFromInactiveReviewers(ctx context.Context) error {
	return s.store.ExecTx(ctx, func(txRepo *repository.PostgresRepository) error {
		// Получение PR с неактивными ревьюерами
		inactiveInfos, err := txRepo.GetOpenPRsWithInactiveReviewers(ctx)
		if err != nil {
			return fmt.Errorf("failed to get PRs with inactive reviewers: %w", err)
		}

		if len(inactiveInfos) == 0 {
			return nil // Нет PR для обработки
		}

		// Группировка по PR
		prToInactive := make(map[string][]models.InactiveReviewerInfo)
		for _, info := range inactiveInfos {
			prToInactive[info.PullRequestID] = append(prToInactive[info.PullRequestID], info)
		}

		// Получение нагрузки пользователей
		workloads, err := txRepo.GetUserWorkload(ctx)
		if err != nil {
			return fmt.Errorf("failed to get user workload: %w", err)
		}

		workloadMap := make(map[string]int64)
		for _, w := range workloads {
			workloadMap[w.UserID] = w.OpenReviewsCount
		}

		// Обработка каждого PR
		for prID, inactives := range prToInactive {
			// Получение команды по первому неактивному пользователю
			teamID := inactives[0].TeamID

			// Получение активных пользователей команды
			activeUsers, err := txRepo.ListActiveByTeamID(ctx, teamID)
			if err != nil {
				return fmt.Errorf("failed to get active users for team %d: %w", teamID, err)
			}

			if len(activeUsers) == 0 {
				// Просто удаляем неактивных ревьюеров
				inactiveUserIDs := make([]string, 0, len(inactives))
				for _, info := range inactives {
					inactiveUserIDs = append(inactiveUserIDs, info.InactiveReviewerID)
				}
				if err := txRepo.RemoveInactiveReviewers(ctx, prID, inactiveUserIDs); err != nil {
					return fmt.Errorf("failed to remove inactive reviewers: %w", err)
				}
				continue
			}

			// Замена каждого неактивного ревьюера на активного
			for _, inactive := range inactives {
				// Выбор пользователя с минимальной нагрузкой
				newReviewer := selectUserWithMinWorkload(activeUsers, workloadMap, prID)
				if newReviewer != nil {
					// Замена ревьюера
					if err := txRepo.Replace(ctx, prID, inactive.InactiveReviewerID, newReviewer.UserID); err != nil {
						return fmt.Errorf("failed to replace reviewer: %w", err)
					}
					// Обновление нагрузки
					workloadMap[newReviewer.UserID]++
				} else {
					// Если нет подходящих ревьюеров, просто удаляем неактивного
					if err := txRepo.Remove(ctx, prID, inactive.InactiveReviewerID); err != nil {
						return fmt.Errorf("failed to remove inactive reviewer: %w", err)
					}
				}
			}
		}

		return nil
	})
}

// selectUsersWithMinWorkload выбирает N пользователей с минимальной нагрузкой
func selectUsersWithMinWorkload(users []models.User, workloadMap map[string]int64, count int) []models.User {
	if count > len(users) {
		count = len(users)
	}

	// Копируем пользователей для сортировки
	sortedUsers := make([]models.User, len(users))
	copy(sortedUsers, users)

	// Простая сортировка по нагрузке (bubble sort для простоты)
	for i := 0; i < len(sortedUsers)-1; i++ {
		for j := 0; j < len(sortedUsers)-i-1; j++ {
			workloadJ := workloadMap[sortedUsers[j].UserID]
			workloadJNext := workloadMap[sortedUsers[j+1].UserID]
			if workloadJ > workloadJNext {
				sortedUsers[j], sortedUsers[j+1] = sortedUsers[j+1], sortedUsers[j]
			}
		}
	}

	return sortedUsers[:count]
}

// selectUserWithMinWorkload выбирает одного пользователя с минимальной нагрузкой
// который еще не назначен на данный PR
func selectUserWithMinWorkload(users []models.User, workloadMap map[string]int64, excludePRID string) *models.User {
	var minUser *models.User
	minWorkload := int64(-1)

	for i := range users {
		workload := workloadMap[users[i].UserID]
		if minWorkload == -1 || workload < minWorkload {
			minWorkload = workload
			minUser = &users[i]
		}
	}

	return minUser
}
