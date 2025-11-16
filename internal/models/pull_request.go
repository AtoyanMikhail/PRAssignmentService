package models

import (
	"time"

	"github.com/AtoyanMikhail/PRAssignmentService/internal/db"
	"github.com/jackc/pgx/v5/pgtype"
)

// PullRequestStatus представляет статус Pull Request
type PullRequestStatus string

const (
	PullRequestStatusOpen   PullRequestStatus = "OPEN"
	PullRequestStatusMerged PullRequestStatus = "MERGED"
)

// IsValid проверяет, является ли статус валидным
func (s PullRequestStatus) IsValid() bool {
	return s == PullRequestStatusOpen || s == PullRequestStatusMerged
}

// PullRequest представляет Pull Request в доменной модели
type PullRequest struct {
	ID              int64
	PullRequestID   string
	PullRequestName string
	AuthorID        string
	Status          PullRequestStatus
	CreatedAt       time.Time
	MergedAt        *time.Time
}

// ToDBPullRequest преобразует доменную модель в модель базы данных
func (pr *PullRequest) ToDBPullRequest() db.PullRequest {
	dbPR := db.PullRequest{
		ID:              pr.ID,
		PullRequestID:   pr.PullRequestID,
		PullRequestName: pr.PullRequestName,
		AuthorID:        pr.AuthorID,
		Status:          string(pr.Status),
		CreatedAt: pgtype.Timestamp{
			Time:  pr.CreatedAt,
			Valid: true,
		},
	}

	if pr.MergedAt != nil {
		dbPR.MergedAt = pgtype.Timestamp{
			Time:  *pr.MergedAt,
			Valid: true,
		}
	}

	return dbPR
}

// PullRequestFromDB преобразует модель базы данных в доменную модель
func PullRequestFromDB(dbPR db.PullRequest) PullRequest {
	pr := PullRequest{
		ID:              dbPR.ID,
		PullRequestID:   dbPR.PullRequestID,
		PullRequestName: dbPR.PullRequestName,
		AuthorID:        dbPR.AuthorID,
		Status:          PullRequestStatus(dbPR.Status),
		CreatedAt:       dbPR.CreatedAt.Time,
	}

	if dbPR.MergedAt.Valid {
		pr.MergedAt = &dbPR.MergedAt.Time
	}

	return pr
}

// PullRequestsFromDB преобразует список моделей базы данных в доменные модели
func PullRequestsFromDB(dbPRs []db.PullRequest) []PullRequest {
	prs := make([]PullRequest, len(dbPRs))
	for i, dbPR := range dbPRs {
		prs[i] = PullRequestFromDB(dbPR)
	}
	return prs
}

// PullRequestShort представляет краткую информацию о PR
type PullRequestShort struct {
	PullRequestID   string
	PullRequestName string
	AuthorID        string
	Status          PullRequestStatus
}

// PullRequestShortFromDBRow преобразует результат запроса GetPullRequestsByReviewerUserID
func PullRequestShortFromDBRow(dbRow db.GetPullRequestsByReviewerUserIDRow) PullRequestShort {
	return PullRequestShort{
		PullRequestID:   dbRow.PullRequestID,
		PullRequestName: dbRow.PullRequestName,
		AuthorID:        dbRow.AuthorID,
		Status:          PullRequestStatus(dbRow.Status),
	}
}

// PullRequestShortsFromDBRows преобразует список результатов запроса
func PullRequestShortsFromDBRows(dbRows []db.GetPullRequestsByReviewerUserIDRow) []PullRequestShort {
	prs := make([]PullRequestShort, len(dbRows))
	for i, dbRow := range dbRows {
		prs[i] = PullRequestShortFromDBRow(dbRow)
	}
	return prs
}
