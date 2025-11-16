package repository

import (
	"context"

	"github.com/AtoyanMikhail/PRAssignmentService/internal/db"
	"github.com/AtoyanMikhail/PRAssignmentService/internal/models"
)

// --- PullRequestRepository implementation ---

func (r *PostgresRepository) CreatePR(ctx context.Context, pullRequestID, pullRequestName, authorID string, status models.PullRequestStatus) (models.PullRequest, error) {
	dbPR, err := r.queries.CreatePullRequest(ctx, db.CreatePullRequestParams{
		PullRequestID:   pullRequestID,
		PullRequestName: pullRequestName,
		AuthorID:        authorID,
		Status:          string(status),
	})
	if err != nil {
		return models.PullRequest{}, err
	}
	return models.PullRequestFromDB(dbPR), nil
}

func (r *PostgresRepository) GetPullRequestByPRID(ctx context.Context, pullRequestID string) (models.PullRequest, error) {
	dbPR, err := r.queries.GetPullRequestByPRID(ctx, pullRequestID)
	if err != nil {
		return models.PullRequest{}, err
	}
	return models.PullRequestFromDB(dbPR), nil
}

func (r *PostgresRepository) GetPullRequestByID(ctx context.Context, id int64) (models.PullRequest, error) {
	dbPR, err := r.queries.GetPullRequestByID(ctx, id)
	if err != nil {
		return models.PullRequest{}, err
	}
	return models.PullRequestFromDB(dbPR), nil
}

func (r *PostgresRepository) UpdateStatus(ctx context.Context, pullRequestID string, status models.PullRequestStatus) (models.PullRequest, error) {
	dbPR, err := r.queries.UpdatePullRequestStatus(ctx, db.UpdatePullRequestStatusParams{
		PullRequestID: pullRequestID,
		Status:        string(status),
	})
	if err != nil {
		return models.PullRequest{}, err
	}
	return models.PullRequestFromDB(dbPR), nil
}

func (r *PostgresRepository) Merge(ctx context.Context, pullRequestID string) (models.PullRequest, error) {
	dbPR, err := r.queries.MergePullRequest(ctx, pullRequestID)
	if err != nil {
		return models.PullRequest{}, err
	}
	return models.PullRequestFromDB(dbPR), nil
}

func (r *PostgresRepository) ListByStatus(ctx context.Context, status models.PullRequestStatus) ([]models.PullRequest, error) {
	dbPRs, err := r.queries.ListPullRequestsByStatus(ctx, string(status))
	if err != nil {
		return nil, err
	}
	return models.PullRequestsFromDB(dbPRs), nil
}

func (r *PostgresRepository) PRExists(ctx context.Context, pullRequestID string) (bool, error) {
	return r.queries.PullRequestExists(ctx, pullRequestID)
}

// --- PRReviewerRepository implementation ---

func (r *PostgresRepository) Add(ctx context.Context, pullRequestID, userID string) (models.PRReviewer, error) {
	dbReviewer, err := r.queries.AddReviewer(ctx, db.AddReviewerParams{
		PullRequestID: pullRequestID,
		UserID:        userID,
	})
	if err != nil {
		return models.PRReviewer{}, err
	}
	return models.PRReviewerFromDB(dbReviewer), nil
}

func (r *PostgresRepository) Remove(ctx context.Context, pullRequestID, userID string) error {
	return r.queries.RemoveReviewer(ctx, db.RemoveReviewerParams{
		PullRequestID: pullRequestID,
		UserID:        userID,
	})
}

func (r *PostgresRepository) GetReviewersByPRID(ctx context.Context, pullRequestID string) ([]models.ReviewerInfo, error) {
	dbRows, err := r.queries.GetReviewersByPRID(ctx, pullRequestID)
	if err != nil {
		return nil, err
	}
	return models.ReviewerInfosFromDBRows(dbRows), nil
}

func (r *PostgresRepository) GetPRsByReviewerUserID(ctx context.Context, userID string) ([]models.PullRequestShort, error) {
	dbRows, err := r.queries.GetPullRequestsByReviewerUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return models.PullRequestShortsFromDBRows(dbRows), nil
}

func (r *PostgresRepository) IsUserAssigned(ctx context.Context, pullRequestID, userID string) (bool, error) {
	return r.queries.IsUserAssignedToPR(ctx, db.IsUserAssignedToPRParams{
		PullRequestID: pullRequestID,
		UserID:        userID,
	})
}

func (r *PostgresRepository) Count(ctx context.Context, pullRequestID string) (int64, error) {
	return r.queries.CountReviewersByPRID(ctx, pullRequestID)
}

func (r *PostgresRepository) Replace(ctx context.Context, pullRequestID, oldUserID, newUserID string) error {
	return r.queries.ReplaceReviewer(ctx, db.ReplaceReviewerParams{
		PullRequestID: pullRequestID,
		UserID:        oldUserID,
		UserID_2:      newUserID,
	})
}

func (r *PostgresRepository) GetOpenPRsWithInactiveReviewers(ctx context.Context) ([]models.InactiveReviewerInfo, error) {
	dbRows, err := r.queries.GetOpenPRsWithInactiveReviewers(ctx)
	if err != nil {
		return nil, err
	}
	return models.InactiveReviewerInfosFromDBRows(dbRows), nil
}

func (r *PostgresRepository) RemoveInactiveReviewers(ctx context.Context, pullRequestID string, userIDs []string) error {
	return r.queries.RemoveInactiveReviewers(ctx, db.RemoveInactiveReviewersParams{
		PullRequestID: pullRequestID,
		Column2:       userIDs,
	})
}

// --- StatisticsRepository implementation ---

func (r *PostgresRepository) GetAssignmentStats(ctx context.Context) ([]models.AssignmentStats, error) {
	dbRows, err := r.queries.GetAssignmentStats(ctx)
	if err != nil {
		return nil, err
	}
	return models.AssignmentStatsListFromDBRows(dbRows), nil
}

func (r *PostgresRepository) GetPRStats(ctx context.Context) ([]models.PRStats, error) {
	dbRows, err := r.queries.GetPRStats(ctx)
	if err != nil {
		return nil, err
	}
	return models.PRStatsListFromDBRows(dbRows), nil
}

func (r *PostgresRepository) GetTeamStats(ctx context.Context) ([]models.TeamStats, error) {
	dbRows, err := r.queries.GetTeamStats(ctx)
	if err != nil {
		return nil, err
	}
	return models.TeamStatsListFromDBRows(dbRows), nil
}

func (r *PostgresRepository) GetUserWorkload(ctx context.Context) ([]models.UserWorkload, error) {
	dbRows, err := r.queries.GetUserWorkload(ctx)
	if err != nil {
		return nil, err
	}
	return models.UserWorkloadListFromDBRows(dbRows), nil
}
