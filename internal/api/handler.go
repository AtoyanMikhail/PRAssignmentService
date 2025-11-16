package api

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/AtoyanMikhail/PRAssignmentService/internal/models"
	"github.com/AtoyanMikhail/PRAssignmentService/internal/service"
	"github.com/gin-gonic/gin"
)

// Handler реализует ServerInterface для обработки HTTP запросов
type Handler struct {
	services *service.Services
}

// NewHandler создает новый HTTP handler
func NewHandler(services *service.Services) *Handler {
	return &Handler{
		services: services,
	}
}

// GetHealth возвращает статус работоспособности сервиса
func (h *Handler) GetHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// PostPullRequestCreate создает PR и автоматически назначает ревьюеров
func (h *Handler) PostPullRequestCreate(c *gin.Context) {
	var req PostPullRequestCreateJSONRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: struct {
				Code    ErrorResponseErrorCode `json:"code"`
				Message string                 `json:"message"`
			}{
				Code:    NOTFOUND,
				Message: "Invalid request body: " + err.Error(),
			},
		})
		return
	}

	// Создаем PR
	pr, err := h.services.PullRequest.CreatePR(c.Request.Context(), req.PullRequestId, req.PullRequestName, req.AuthorId)
	if err != nil {
		switch err {
		case service.ErrPullRequestAlreadyExists:
			c.JSON(http.StatusConflict, ErrorResponse{
				Error: struct {
					Code    ErrorResponseErrorCode `json:"code"`
					Message string                 `json:"message"`
				}{
					Code:    PREXISTS,
					Message: "Pull request already exists",
				},
			})
		case service.ErrUserNotFound:
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error: struct {
					Code    ErrorResponseErrorCode `json:"code"`
					Message string                 `json:"message"`
				}{
					Code:    NOTFOUND,
					Message: "Author not found",
				},
			})
		default:
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error: struct {
					Code    ErrorResponseErrorCode `json:"code"`
					Message string                 `json:"message"`
				}{
					Code:    NOTFOUND,
					Message: err.Error(),
				},
			})
		}
		return
	}

	// Автоматически назначаем ревьюеров (до 2)
	reviewers, err := h.services.Reviewer.AutoAssignReviewers(c.Request.Context(), pr.PullRequestID, 2)
	if err != nil && err != service.ErrNoActiveReviewers {
		// Логируем ошибку, но не прерываем выполнение
		c.Error(err)
	}

	// Формируем ответ
	assignedReviewerIDs := make([]string, 0, len(reviewers))
	for _, r := range reviewers {
		assignedReviewerIDs = append(assignedReviewerIDs, r.UserID)
	}

	response := PullRequest{
		PullRequestId:     pr.PullRequestID,
		PullRequestName:   pr.PullRequestName,
		AuthorId:          pr.AuthorID,
		Status:            PullRequestStatus(pr.Status),
		AssignedReviewers: assignedReviewerIDs,
		CreatedAt:         &pr.CreatedAt,
	}

	c.JSON(http.StatusCreated, response)
}

// PostPullRequestMerge помечает PR как MERGED
func (h *Handler) PostPullRequestMerge(c *gin.Context) {
	var req PostPullRequestMergeJSONRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: struct {
				Code    ErrorResponseErrorCode `json:"code"`
				Message string                 `json:"message"`
			}{
				Code:    NOTFOUND,
				Message: "Invalid request body: " + err.Error(),
			},
		})
		return
	}

	pr, err := h.services.PullRequest.MergePR(c.Request.Context(), req.PullRequestId)
	if err != nil {
		switch err {
		case service.ErrPullRequestNotFound:
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error: struct {
					Code    ErrorResponseErrorCode `json:"code"`
					Message string                 `json:"message"`
				}{
					Code:    NOTFOUND,
					Message: "Pull request not found",
				},
			})
		default:
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error: struct {
					Code    ErrorResponseErrorCode `json:"code"`
					Message string                 `json:"message"`
				}{
					Code:    NOTFOUND,
					Message: err.Error(),
				},
			})
		}
		return
	}

	// Получаем ревьюеров
	reviewers, _ := h.services.Reviewer.GetPRReviewers(c.Request.Context(), pr.PullRequestID)
	assignedReviewerIDs := make([]string, 0, len(reviewers))
	for _, r := range reviewers {
		assignedReviewerIDs = append(assignedReviewerIDs, r.UserID)
	}

	response := PullRequest{
		PullRequestId:     pr.PullRequestID,
		PullRequestName:   pr.PullRequestName,
		AuthorId:          pr.AuthorID,
		Status:            PullRequestStatus(pr.Status),
		AssignedReviewers: assignedReviewerIDs,
		CreatedAt:         &pr.CreatedAt,
		MergedAt:          pr.MergedAt,
	}

	c.JSON(http.StatusOK, response)
}

// PostPullRequestReassign переназначает ревьюера
func (h *Handler) PostPullRequestReassign(c *gin.Context) {
	var req PostPullRequestReassignJSONRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: struct {
				Code    ErrorResponseErrorCode `json:"code"`
				Message string                 `json:"message"`
			}{
				Code:    NOTFOUND,
				Message: "Invalid request body: " + err.Error(),
			},
		})
		return
	}

	err := h.services.Reviewer.ReplaceReviewer(c.Request.Context(), req.PullRequestId, req.OldUserId, "")
	if err != nil {
		switch err {
		case service.ErrPullRequestNotFound:
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error: struct {
					Code    ErrorResponseErrorCode `json:"code"`
					Message string                 `json:"message"`
				}{
					Code:    NOTFOUND,
					Message: "Pull request not found",
				},
			})
		case service.ErrReviewerNotAssigned:
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error: struct {
					Code    ErrorResponseErrorCode `json:"code"`
					Message string                 `json:"message"`
				}{
					Code:    NOTASSIGNED,
					Message: "Reviewer not assigned to this PR",
				},
			})
		case service.ErrNoActiveReviewers:
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error: struct {
					Code    ErrorResponseErrorCode `json:"code"`
					Message string                 `json:"message"`
				}{
					Code:    NOCANDIDATE,
					Message: "No active reviewers available for reassignment",
				},
			})
		default:
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error: struct {
					Code    ErrorResponseErrorCode `json:"code"`
					Message string                 `json:"message"`
				}{
					Code:    NOTFOUND,
					Message: err.Error(),
				},
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "reassigned"})
}

// PostTeamAdd создает команду с участниками
func (h *Handler) PostTeamAdd(c *gin.Context) {
	var req PostTeamAddJSONRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: struct {
				Code    ErrorResponseErrorCode `json:"code"`
				Message string                 `json:"message"`
			}{
				Code:    NOTFOUND,
				Message: "Invalid request body: " + err.Error(),
			},
		})
		return
	}

	// Создаем/получаем команду
	team, err := h.services.Team.GetTeam(c.Request.Context(), req.TeamName)
	if err != nil {
		if err == service.ErrTeamNotFound || err == sql.ErrNoRows {
			team, err = h.services.Team.CreateTeam(c.Request.Context(), req.TeamName)
			if err != nil {
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Error: struct {
						Code    ErrorResponseErrorCode `json:"code"`
						Message string                 `json:"message"`
					}{
						Code:    NOTFOUND,
						Message: err.Error(),
					},
				})
				return
			}
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error: struct {
					Code    ErrorResponseErrorCode `json:"code"`
					Message string                 `json:"message"`
				}{
					Code:    NOTFOUND,
					Message: err.Error(),
				},
			})
			return
		}
	}

	// Создаем/обновляем пользователей
	for _, member := range req.Members {
		_, err := h.services.User.GetUser(c.Request.Context(), member.UserId)
		if err != nil {
			if err == service.ErrUserNotFound || err == sql.ErrNoRows {
				_, err = h.services.User.CreateUser(c.Request.Context(), member.UserId, member.Username, team.ID)
				if err != nil {
					c.JSON(http.StatusInternalServerError, ErrorResponse{
						Error: struct {
							Code    ErrorResponseErrorCode `json:"code"`
							Message string                 `json:"message"`
						}{
							Code:    NOTFOUND,
							Message: err.Error(),
						},
					})
					return
				}
				// После создания обновляем is_active если нужно
				if !member.IsActive {
					_, _ = h.services.User.DeactivateUser(c.Request.Context(), member.UserId)
				}
			} else {
				// Другая ошибка
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Error: struct {
						Code    ErrorResponseErrorCode `json:"code"`
						Message string                 `json:"message"`
					}{
						Code:    NOTFOUND,
						Message: err.Error(),
					},
				})
				return
			}
		} else {
			// Обновляем существующего пользователя
			_, err = h.services.User.UpdateUser(c.Request.Context(), member.UserId, member.Username, team.ID, member.IsActive)
			if err != nil {
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Error: struct {
						Code    ErrorResponseErrorCode `json:"code"`
						Message string                 `json:"message"`
					}{
						Code:    NOTFOUND,
						Message: err.Error(),
					},
				})
				return
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "team_name": req.TeamName})
}

// PostTeamDeactivate массово деактивирует всех пользователей команды и переназначает их PR
func (h *Handler) PostTeamDeactivate(c *gin.Context) {
	var req PostTeamDeactivateJSONRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: struct {
				Code    ErrorResponseErrorCode `json:"code"`
				Message string                 `json:"message"`
			}{
				Code:    NOTFOUND,
				Message: "Invalid request body: " + err.Error(),
			},
		})
		return
	}

	// Засекаем время начала операции
	start := time.Now()

	// Получаем команду для проверки существования
	team, err := h.services.Team.GetTeam(c.Request.Context(), req.TeamName)
	if err != nil {
		if err == service.ErrTeamNotFound || err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error: struct {
					Code    ErrorResponseErrorCode `json:"code"`
					Message string                 `json:"message"`
				}{
					Code:    NOTFOUND,
					Message: "Team not found",
				},
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error: struct {
					Code    ErrorResponseErrorCode `json:"code"`
					Message string                 `json:"message"`
				}{
					Code:    NOTFOUND,
					Message: err.Error(),
				},
			})
		}
		return
	}

	// Выполняем массовую деактивацию с переназначением PR
	deactivatedUsers, reassignedPRs, err := h.services.User.DeactivateTeamUsers(c.Request.Context(), team.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: struct {
				Code    ErrorResponseErrorCode `json:"code"`
				Message string                 `json:"message"`
			}{
				Code:    NOTFOUND,
				Message: "Failed to deactivate team users: " + err.Error(),
			},
		})
		return
	}

	// Вычисляем время выполнения
	duration := time.Since(start).Milliseconds()

	c.JSON(http.StatusOK, gin.H{
		"deactivated_users": deactivatedUsers,
		"reassigned_prs":    reassignedPRs,
		"duration_ms":       duration,
	})
}

// GetTeamGet возвращает команду с участниками
func (h *Handler) GetTeamGet(c *gin.Context, params GetTeamGetParams) {
	team, err := h.services.Team.GetTeam(c.Request.Context(), params.TeamName)
	if err != nil {
		if err == service.ErrTeamNotFound || err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error: struct {
					Code    ErrorResponseErrorCode `json:"code"`
					Message string                 `json:"message"`
				}{
					Code:    NOTFOUND,
					Message: "Team not found",
				},
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error: struct {
					Code    ErrorResponseErrorCode `json:"code"`
					Message string                 `json:"message"`
				}{
					Code:    NOTFOUND,
					Message: err.Error(),
				},
			})
		}
		return
	}

	// Получаем пользователей команды
	users, err := h.services.User.ListTeamUsers(c.Request.Context(), team.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: struct {
				Code    ErrorResponseErrorCode `json:"code"`
				Message string                 `json:"message"`
			}{
				Code:    NOTFOUND,
				Message: err.Error(),
			},
		})
		return
	}

	members := make([]TeamMember, 0, len(users))
	for _, user := range users {
		members = append(members, TeamMember{
			UserId:   user.UserID,
			Username: user.Username,
			IsActive: user.IsActive,
		})
	}

	response := Team{
		TeamName: team.TeamName,
		Members:  members,
	}

	c.JSON(http.StatusOK, response)
}

// GetUsersGetReview возвращает PR'ы где пользователь назначен ревьюером
func (h *Handler) GetUsersGetReview(c *gin.Context, params GetUsersGetReviewParams) {
	prs, err := h.services.Reviewer.GetUserPRs(c.Request.Context(), params.UserId)
	if err != nil {
		if err == service.ErrUserNotFound || err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error: struct {
					Code    ErrorResponseErrorCode `json:"code"`
					Message string                 `json:"message"`
				}{
					Code:    NOTFOUND,
					Message: "User not found",
				},
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error: struct {
					Code    ErrorResponseErrorCode `json:"code"`
					Message string                 `json:"message"`
				}{
					Code:    NOTFOUND,
					Message: err.Error(),
				},
			})
		}
		return
	}

	response := make([]PullRequestShort, 0, len(prs))
	for _, pr := range prs {
		response = append(response, PullRequestShort{
			PullRequestId:   pr.PullRequestID,
			PullRequestName: pr.PullRequestName,
			AuthorId:        pr.AuthorID,
			Status:          PullRequestShortStatus(pr.Status),
		})
	}

	c.JSON(http.StatusOK, response)
}

// PostUsersSetIsActive устанавливает флаг активности пользователя
func (h *Handler) PostUsersSetIsActive(c *gin.Context) {
	var req PostUsersSetIsActiveJSONRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: struct {
				Code    ErrorResponseErrorCode `json:"code"`
				Message string                 `json:"message"`
			}{
				Code:    NOTFOUND,
				Message: "Invalid request body: " + err.Error(),
			},
		})
		return
	}

	// Получаем пользователя
	user, err := h.services.User.GetUser(c.Request.Context(), req.UserId)
	if err != nil {
		if err == service.ErrUserNotFound || err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error: struct {
					Code    ErrorResponseErrorCode `json:"code"`
					Message string                 `json:"message"`
				}{
					Code:    NOTFOUND,
					Message: "User not found",
				},
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error: struct {
					Code    ErrorResponseErrorCode `json:"code"`
					Message string                 `json:"message"`
				}{
					Code:    NOTFOUND,
					Message: err.Error(),
				},
			})
		}
		return
	}

	// Обновляем пользователя
	var updatedUser models.User
	if req.IsActive {
		updatedUser, err = h.services.User.ActivateUser(c.Request.Context(), user.UserID)
	} else {
		updatedUser, err = h.services.User.DeactivateUser(c.Request.Context(), user.UserID)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: struct {
				Code    ErrorResponseErrorCode `json:"code"`
				Message string                 `json:"message"`
			}{
				Code:    NOTFOUND,
				Message: err.Error(),
			},
		})
		return
	}

	// Если деактивируем пользователя, удаляем его из PR
	if !req.IsActive {
		// Переназначаем PR от неактивного ревьюера
		err = h.services.Reviewer.ReassignFromInactiveReviewers(c.Request.Context())
		if err != nil {
			c.Error(err)
		}
	}

	response := User{
		UserId:   updatedUser.UserID,
		Username: updatedUser.Username,
		TeamName: "", // Можно получить из team если нужно
		IsActive: updatedUser.IsActive,
	}

	c.JSON(http.StatusOK, response)
}

// GetStatisticsAssignments возвращает статистику назначений по пользователям
func (h *Handler) GetStatisticsAssignments(c *gin.Context) {
	stats, err := h.services.Statistics.GetAssignmentStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: struct {
				Code    ErrorResponseErrorCode `json:"code"`
				Message string                 `json:"message"`
			}{
				Code:    NOTFOUND,
				Message: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"statistics": stats,
	})
}

// GetStatisticsWorkload возвращает рабочую нагрузку активных пользователей
func (h *Handler) GetStatisticsWorkload(c *gin.Context) {
	workload, err := h.services.Statistics.GetUserWorkload(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: struct {
				Code    ErrorResponseErrorCode `json:"code"`
				Message string                 `json:"message"`
			}{
				Code:    NOTFOUND,
				Message: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"workload": workload,
	})
}
