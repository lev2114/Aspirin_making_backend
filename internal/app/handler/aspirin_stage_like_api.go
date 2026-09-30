package handler

import (
	"errors"
	"net/http"
	"strconv"

	"aspirin_back/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type AspirinSynthesisStageLikeRequest struct {
	Like int `json:"like"`
}

// POST /api/aspirin-stages/:id/like
func (h *Handler) SetAspirinSynthesisStageLikeAPI(
	ctx *gin.Context,
) {

	currentUserID :=
		GetCurrentAspirinProductionUserID()

	stageID, err :=
		strconv.Atoi(
			ctx.Param("id"),
		)

	if err != nil ||
		stageID < 1 {

		ctx.JSON(
			http.StatusBadRequest,
			gin.H{
				"message":
					"Некорректный ID этапа синтеза",
			},
		)

		return
	}

	var request AspirinSynthesisStageLikeRequest

	if err :=
		ctx.ShouldBindJSON(
			&request,
		); err != nil {

		ctx.JSON(
			http.StatusBadRequest,
			gin.H{
				"message":
					"Некорректное значение лайка",
			},
		)

		return
	}

	if request.Like != 0 &&
		request.Like != 1 {

		ctx.JSON(
			http.StatusBadRequest,
			gin.H{
				"message":
					"Поле like должно иметь значение 0 или 1",
			},
		)

		return
	}

	likeCount, err :=
		h.Repository.
			SetAspirinSynthesisStageLike(
				stageID,
				currentUserID,
				request.Like,
			)

	if err != nil {

		if errors.Is(
			err,
			repository.ErrAspirinSynthesisStageNotFound,
		) {

			ctx.JSON(
				http.StatusNotFound,
				gin.H{
					"message":
						"Опубликованный этап не найден",
				},
			)

			return
		}

		logrus.Error(
			"Ошибка изменения лайка: ",
			err,
		)

		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{
				"message":
					"Ошибка изменения лайка",
			},
		)

		return
	}

	ctx.JSON(
		http.StatusOK,
		gin.H{
			"data": gin.H{
				"like":
					request.Like,

				"aspirin_synthesis_like_count":
					likeCount,
			},
		},
	)
}