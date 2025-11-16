package models

import (
	"time"

	"github.com/AtoyanMikhail/PRAssignmentService/internal/db"
	"github.com/jackc/pgx/v5/pgtype"
)

// PRReviewer представляет назначение ревьювера на PR
type PRReviewer struct {
	ID            int64
	PullRequestID string
	UserID        string
	AssignedAt    time.Time
}

// ToDBPRReviewer преобразует доменную модель в модель базы данных
func (r *PRReviewer) ToDBPRReviewer() db.PrReviewer {
	return db.PrReviewer{
		ID:            r.ID,
		PullRequestID: r.PullRequestID,
		UserID:        r.UserID,
		AssignedAt: pgtype.Timestamp{
			Time:  r.AssignedAt,
			Valid: true,
		},
	}
}

// PRReviewerFromDB преобразует модель базы данных в доменную модель
func PRReviewerFromDB(dbReviewer db.PrReviewer) PRReviewer {
	return PRReviewer{
		ID:            dbReviewer.ID,
		PullRequestID: dbReviewer.PullRequestID,
		UserID:        dbReviewer.UserID,
		AssignedAt:    dbReviewer.AssignedAt.Time,
	}
}

// PRReviewersFromDB преобразует список моделей базы данных в доменные модели
func PRReviewersFromDB(dbReviewers []db.PrReviewer) []PRReviewer {
	reviewers := make([]PRReviewer, len(dbReviewers))
	for i, dbReviewer := range dbReviewers {
		reviewers[i] = PRReviewerFromDB(dbReviewer)
	}
	return reviewers
}

// ReviewerInfo представляет информацию о ревьювере с деталями
type ReviewerInfo struct {
	ID            int64
	PullRequestID string
	UserID        string
	Username      string
	TeamID        int64
	IsActive      bool
	AssignedAt    time.Time
}

// ReviewerInfoFromDBRow преобразует результат запроса GetReviewersByPRID
func ReviewerInfoFromDBRow(dbRow db.GetReviewersByPRIDRow) ReviewerInfo {
	return ReviewerInfo{
		ID:            dbRow.ID,
		PullRequestID: dbRow.PullRequestID,
		UserID:        dbRow.UserID,
		Username:      dbRow.Username,
		TeamID:        dbRow.TeamID,
		IsActive:      dbRow.IsActive,
		AssignedAt:    dbRow.AssignedAt.Time,
	}
}

// ReviewerInfosFromDBRows преобразует список результатов запроса
func ReviewerInfosFromDBRows(dbRows []db.GetReviewersByPRIDRow) []ReviewerInfo {
	reviewers := make([]ReviewerInfo, len(dbRows))
	for i, dbRow := range dbRows {
		reviewers[i] = ReviewerInfoFromDBRow(dbRow)
	}
	return reviewers
}

// InactiveReviewerInfo представляет информацию о неактивном ревьювере на PR
type InactiveReviewerInfo struct {
	PullRequestID      string
	AuthorID           string
	InactiveReviewerID string
	TeamID             int64
}

// InactiveReviewerInfoFromDBRow преобразует результат запроса GetOpenPRsWithInactiveReviewers
func InactiveReviewerInfoFromDBRow(dbRow db.GetOpenPRsWithInactiveReviewersRow) InactiveReviewerInfo {
	return InactiveReviewerInfo{
		PullRequestID:      dbRow.PullRequestID,
		AuthorID:           dbRow.AuthorID,
		InactiveReviewerID: dbRow.InactiveReviewerID,
		TeamID:             dbRow.TeamID,
	}
}

// InactiveReviewerInfosFromDBRows преобразует список результатов запроса
func InactiveReviewerInfosFromDBRows(dbRows []db.GetOpenPRsWithInactiveReviewersRow) []InactiveReviewerInfo {
	infos := make([]InactiveReviewerInfo, len(dbRows))
	for i, dbRow := range dbRows {
		infos[i] = InactiveReviewerInfoFromDBRow(dbRow)
	}
	return infos
}
