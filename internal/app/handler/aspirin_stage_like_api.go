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
// @Summary Поставить или снять лайк текущего пользователя
// @Tags Этапы синтеза аспирина
// @Accept json
// @Produce json
// @Security AspirinProductionBearer
// @Param id path int true "ID опубликованного этапа" minimum(1)
// @Param aspirin_stage_like body AspirinSynthesisStageLikeRequest true "1 — поставить, 0 — снять"
// @Success 200 {object} AspirinSynthesisStageLikeDataResponse
// @Failure 400,401,404,500,503 {object} AspirinProductionMessageResponse
// @Router /aspirin-stages/{id}/like [post]
func (h *Handler) SetAspirinSynthesisStageLikeAPI(
	ctx *gin.Context,
) {

	currentUserID :=
		GetCurrentAspirinProductionUserID(ctx)

	stageID, err :=
		strconv.Atoi(
			ctx.Param("id"),
		)

	if err != nil ||
		stageID < 1 {

		ctx.JSON(
			http.StatusBadRequest,
			gin.H{
				"message": "Некорректный ID этапа синтеза",
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
				"message": "Некорректное значение лайка",
			},
		)

		return
	}

	if request.Like != 0 &&
		request.Like != 1 {

		ctx.JSON(
			http.StatusBadRequest,
			gin.H{
				"message": "Поле like должно иметь значение 0 или 1",
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
					"message": "Опубликованный этап не найден",
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
				"message": "Ошибка изменения лайка",
			},
		)

		return
	}

	ctx.JSON(
		http.StatusOK,
		gin.H{
			"data": gin.H{
				"like": request.Like,

				"aspirin_synthesis_like_count": likeCount,
			},
		},
	)
}

type AspirinSynthesisStageLikeResponse struct {
	Like                      int `json:"like"`
	AspirinSynthesisLikeCount int `json:"aspirin_synthesis_like_count"`
}
type AspirinSynthesisStageLikeDataResponse struct {
	Data AspirinSynthesisStageLikeResponse `json:"data"`
}
