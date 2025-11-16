package models

import (
	"time"

	"github.com/AtoyanMikhail/PRAssignmentService/internal/db"
	"github.com/jackc/pgx/v5/pgtype"
)

// Team представляет команду в доменной модели
type Team struct {
	ID        int64
	TeamName  string
	CreatedAt time.Time
}

// ToDBTeam преобразует доменную модель в модель базы данных
func (t *Team) ToDBTeam() db.Team {
	return db.Team{
		ID:       t.ID,
		TeamName: t.TeamName,
		CreatedAt: pgtype.Timestamp{
			Time:  t.CreatedAt,
			Valid: true,
		},
	}
}

// TeamFromDB преобразует модель базы данных в доменную модель
func TeamFromDB(dbTeam db.Team) Team {
	return Team{
		ID:        dbTeam.ID,
		TeamName:  dbTeam.TeamName,
		CreatedAt: dbTeam.CreatedAt.Time,
	}
}

// TeamsFromDB преобразует список моделей базы данных в доменные модели
func TeamsFromDB(dbTeams []db.Team) []Team {
	teams := make([]Team, len(dbTeams))
	for i, dbTeam := range dbTeams {
		teams[i] = TeamFromDB(dbTeam)
	}
	return teams
}
