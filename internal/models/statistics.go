package models

import (
	"github.com/AtoyanMikhail/PRAssignmentService/internal/db"
)

// AssignmentStats представляет статистику назначений по пользователю
type AssignmentStats struct {
	UserID           string
	Username         string
	TeamName         *string
	TotalAssignments int64
	OpenPRs          int64
	MergedPRs        int64
}

// AssignmentStatsFromDBRow преобразует результат запроса GetAssignmentStats
func AssignmentStatsFromDBRow(dbRow db.GetAssignmentStatsRow) AssignmentStats {
	return AssignmentStats{
		UserID:           dbRow.UserID,
		Username:         dbRow.Username,
		TeamName:         dbRow.TeamName,
		TotalAssignments: dbRow.TotalAssignments,
		OpenPRs:          dbRow.OpenPrs,
		MergedPRs:        dbRow.MergedPrs,
	}
}

// AssignmentStatsListFromDBRows преобразует список результатов запроса
func AssignmentStatsListFromDBRows(dbRows []db.GetAssignmentStatsRow) []AssignmentStats {
	stats := make([]AssignmentStats, len(dbRows))
	for i, dbRow := range dbRows {
		stats[i] = AssignmentStatsFromDBRow(dbRow)
	}
	return stats
}

// PRStats представляет статистику по Pull Request
type PRStats struct {
	PullRequestID   string
	PullRequestName string
	AuthorID        string
	Status          PullRequestStatus
	ReviewersCount  int64
}

// PRStatsFromDBRow преобразует результат запроса GetPRStats
func PRStatsFromDBRow(dbRow db.GetPRStatsRow) PRStats {
	return PRStats{
		PullRequestID:   dbRow.PullRequestID,
		PullRequestName: dbRow.PullRequestName,
		AuthorID:        dbRow.AuthorID,
		Status:          PullRequestStatus(dbRow.Status),
		ReviewersCount:  dbRow.ReviewersCount,
	}
}

// PRStatsListFromDBRows преобразует список результатов запроса
func PRStatsListFromDBRows(dbRows []db.GetPRStatsRow) []PRStats {
	stats := make([]PRStats, len(dbRows))
	for i, dbRow := range dbRows {
		stats[i] = PRStatsFromDBRow(dbRow)
	}
	return stats
}

// TeamStats представляет статистику по команде
type TeamStats struct {
	TeamName         string
	TotalMembers     int64
	ActiveMembers    int64
	TotalPRsAuthored int64
	TotalPRsReviewed int64
}

// TeamStatsFromDBRow преобразует результат запроса GetTeamStats
func TeamStatsFromDBRow(dbRow db.GetTeamStatsRow) TeamStats {
	return TeamStats{
		TeamName:         dbRow.TeamName,
		TotalMembers:     dbRow.TotalMembers,
		ActiveMembers:    dbRow.ActiveMembers,
		TotalPRsAuthored: dbRow.TotalPrsAuthored,
		TotalPRsReviewed: dbRow.TotalPrsReviewed,
	}
}

// TeamStatsListFromDBRows преобразует список результатов запроса
func TeamStatsListFromDBRows(dbRows []db.GetTeamStatsRow) []TeamStats {
	stats := make([]TeamStats, len(dbRows))
	for i, dbRow := range dbRows {
		stats[i] = TeamStatsFromDBRow(dbRow)
	}
	return stats
}

// UserWorkload представляет рабочую нагрузку пользователя
type UserWorkload struct {
	UserID           string
	Username         string
	TeamName         *string
	IsActive         bool
	OpenReviewsCount int64
}

// UserWorkloadFromDBRow преобразует результат запроса GetUserWorkload
func UserWorkloadFromDBRow(dbRow db.GetUserWorkloadRow) UserWorkload {
	return UserWorkload{
		UserID:           dbRow.UserID,
		Username:         dbRow.Username,
		TeamName:         dbRow.TeamName,
		IsActive:         dbRow.IsActive,
		OpenReviewsCount: dbRow.OpenReviewsCount,
	}
}

// UserWorkloadListFromDBRows преобразует список результатов запроса
func UserWorkloadListFromDBRows(dbRows []db.GetUserWorkloadRow) []UserWorkload {
	workloads := make([]UserWorkload, len(dbRows))
	for i, dbRow := range dbRows {
		workloads[i] = UserWorkloadFromDBRow(dbRow)
	}
	return workloads
}
