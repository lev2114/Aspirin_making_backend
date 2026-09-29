package handler

import (
	"errors"
	"net/http"
	"strings"

	"aspirin_back/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type RegisterAspirinProductionUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AspirinProductionUserResponse struct {
	ProductionUserID int `json:"production_user_id"`

	ProductionUsername string `json:"production_username"`
}

// POST /api/users
func (h *Handler) RegisterAspirinProductionUserAPI(
	ctx *gin.Context,
) {

	var request RegisterAspirinProductionUserRequest

	if err :=
		ctx.ShouldBindJSON(
			&request,
		); err != nil {

		ctx.JSON(
			http.StatusBadRequest,
			gin.H{
				"status":  "fail",
				"message": "Некорректные данные пользователя",
			},
		)

		return
	}

	request.Username =
		strings.TrimSpace(
			request.Username,
		)

	request.Password =
		strings.TrimSpace(
			request.Password,
		)

	if request.Username == "" {

		ctx.JSON(
			http.StatusBadRequest,
			gin.H{
				"status":  "fail",
				"message": "Имя пользователя не может быть пустым",
			},
		)

		return
	}

	if request.Password == "" {

		ctx.JSON(
			http.StatusBadRequest,
			gin.H{
				"status":  "fail",
				"message": "Пароль не может быть пустым",
			},
		)

		return
	}

	user, err :=
		h.Repository.
			CreateAspirinProductionUser(
				request.Username,
				request.Password,
			)

	if err != nil {

		if errors.Is(
			err,
			repository.ErrAspirinProductionUserExists,
		) {

			ctx.JSON(
				http.StatusConflict,
				gin.H{
					"status":  "fail",
					"message": "Пользователь с таким именем уже существует",
				},
			)

			return
		}

		logrus.Error(
			"Ошибка регистрации пользователя: ",
			err,
		)

		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{
				"status":  "fail",
				"message": "Ошибка регистрации пользователя",
			},
		)

		return
	}

	ctx.JSON(
		http.StatusCreated,
		gin.H{
			"status": "success",
			"data":
				AspirinProductionUserResponse{
					ProductionUserID:
						user.ProductionUserID,

					ProductionUsername:
						user.ProductionUsername,
				},
		},
	)
}

// POST /api/authentication
// Заглушка для ЛР-4.
func (h *Handler) AuthenticateAspirinProductionUserAPI(
	ctx *gin.Context,
) {

	ctx.JSON(
		http.StatusOK,
		gin.H{
			"status": "success",
			"message":
				"Аутентификация будет реализована в лабораторной работе №4",
			"stub": true,
		},
	)
}

// POST /api/deauthentication
// Заглушка для ЛР-4.
func (h *Handler) DeauthenticateAspirinProductionUserAPI(
	ctx *gin.Context,
) {

	ctx.JSON(
		http.StatusOK,
		gin.H{
			"status": "success",
			"message":
				"Деавторизация будет реализована в лабораторной работе №4",
			"stub": true,
		},
	)
}