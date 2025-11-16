package models

import (
	"time"

	"github.com/AtoyanMikhail/PRAssignmentService/internal/db"
	"github.com/jackc/pgx/v5/pgtype"
)

// User представляет пользователя в доменной модели
type User struct {
	ID        int64
	UserID    string
	Username  string
	TeamID    int64
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ToDBUser преобразует доменную модель в модель базы данных
func (u *User) ToDBUser() db.User {
	return db.User{
		ID:       u.ID,
		UserID:   u.UserID,
		Username: u.Username,
		TeamID:   u.TeamID,
		IsActive: u.IsActive,
		CreatedAt: pgtype.Timestamp{
			Time:  u.CreatedAt,
			Valid: true,
		},
		UpdatedAt: pgtype.Timestamp{
			Time:  u.UpdatedAt,
			Valid: true,
		},
	}
}

// UserFromDB преобразует модель базы данных в доменную модель
func UserFromDB(dbUser db.User) User {
	return User{
		ID:        dbUser.ID,
		UserID:    dbUser.UserID,
		Username:  dbUser.Username,
		TeamID:    dbUser.TeamID,
		IsActive:  dbUser.IsActive,
		CreatedAt: dbUser.CreatedAt.Time,
		UpdatedAt: dbUser.UpdatedAt.Time,
	}
}

// UsersFromDB преобразует список моделей базы данных в доменные модели
func UsersFromDB(dbUsers []db.User) []User {
	users := make([]User, len(dbUsers))
	for i, dbUser := range dbUsers {
		users[i] = UserFromDB(dbUser)
	}
	return users
}

// UserWithTeam представляет пользователя с информацией о команде
type UserWithTeam struct {
	ID        int64
	UserID    string
	Username  string
	TeamID    int64
	TeamName  string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// UserWithTeamFromDB преобразует результат запроса GetUserWithTeam в доменную модель
func UserWithTeamFromDB(dbRow db.GetUserWithTeamRow) UserWithTeam {
	return UserWithTeam{
		ID:        dbRow.ID,
		UserID:    dbRow.UserID,
		Username:  dbRow.Username,
		TeamID:    dbRow.TeamID,
		TeamName:  dbRow.TeamName,
		IsActive:  dbRow.IsActive,
		CreatedAt: dbRow.CreatedAt.Time,
		UpdatedAt: dbRow.UpdatedAt.Time,
	}
}
