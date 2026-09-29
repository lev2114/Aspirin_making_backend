package handler

import (
	"aspirin_back/internal/app/repository"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const (
	defaultAspirinStageImageURL = "/static/img/default.png"
	defaultAspirinStageVideoURL = "/static/img/default.mp4"
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

func parseOptionalAspirinSynthesisStageID(
	ctx *gin.Context,
) (*int, error) {

	stageIDText :=
		strings.TrimSpace(
			ctx.Query(
				"aspirin_synthesis_stage_id",
			),
		)

	if stageIDText == "" {
		return nil, nil
	}

	stageID, err :=
		strconv.Atoi(
			stageIDText,
		)

	if err != nil ||
		stageID < 1 {

		return nil,
			errors.New(
				"некорректный ID этапа синтеза аспирина",
			)
	}

	return &stageID, nil
}

func (h *Handler) prepareAspirinStageMediaForView(
	stage *repository.AspirinSynthesisStage,
) {

	if stage.AspirinSynthesisStageImageURL == nil ||
		strings.TrimSpace(
			*stage.AspirinSynthesisStageImageURL,
		) == "" {

		imageURL :=
			defaultAspirinStageImageURL

		stage.AspirinSynthesisStageImageURL =
			&imageURL

	} else {

		stage.AspirinSynthesisStageImageURL =
			h.Repository.
				AspirinStageMediaURL(
					stage.AspirinSynthesisStageImageURL,
				)
	}

	if stage.AspirinSynthesisStageVideoURL == nil ||
		strings.TrimSpace(
			*stage.AspirinSynthesisStageVideoURL,
		) == "" {

		videoURL :=
			defaultAspirinStageVideoURL

		stage.AspirinSynthesisStageVideoURL =
			&videoURL

	} else {

		stage.AspirinSynthesisStageVideoURL =
			h.Repository.
				AspirinStageMediaURL(
					stage.AspirinSynthesisStageVideoURL,
				)
	}
}

func (h *Handler) prepareAspirinStagesMediaForView(
	stages []repository.AspirinSynthesisStage,
) {

	for i := range stages {

		h.prepareAspirinStageMediaForView(
			&stages[i],
		)
	}
}

func (h *Handler) ShowAspirinSynthesisStage(
	ctx *gin.Context,
) {

	stageID, err :=
		parseOptionalAspirinSynthesisStageID(
			ctx,
		)

	if err != nil {

		ctx.JSON(
			http.StatusBadRequest,
			gin.H{
				"error": err.Error(),
			},
		)

		return
	}

	nextAspirinSynthesisStage :=
		ctx.Query("next") == "true"

	stage, err :=
		h.Repository.
			GetAspirinSynthesisStageForFeed(
				stageID,
				nextAspirinSynthesisStage,
			)

	if err != nil {

		if errors.Is(
			err,
			repository.ErrAspirinSynthesisStageNotFound,
		) {

			ctx.JSON(
				http.StatusNotFound,
				gin.H{
					"error":
						"Опубликованный этап синтеза не найден или удалён",
				},
			)

			return
		}

		logrus.Error(
			"Ошибка получения этапа синтеза аспирина: ",
			err,
		)

		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{
				"error":
					"Ошибка получения этапа синтеза аспирина",
			},
		)

		return
	}

	h.prepareAspirinStageMediaForView(
		stage,
	)

	ctx.HTML(
		http.StatusOK,
		"aspirin_synthesis_stage_feed.html",
		gin.H{
			"aspirinSynthesisStage":
				stage,

			"aspirinSynthesisStageDescriptionExpanded":
				ctx.Query(
					"aspirin_synthesis_stage_description",
				) == "full",

			"defaultAspirinStageVideoURL":
				defaultAspirinStageVideoURL,
		},
	)
}

func (h *Handler) ShowAspirinSynthesisStageDraft(
	ctx *gin.Context,
) {

	draft, err :=
		h.Repository.
			GetAspirinSynthesisStageDraft(
				GetCurrentAspirinProductionUserID(),
			)

	if err != nil {

		logrus.Error(
			"Ошибка получения черновика этапа синтеза аспирина: ",
			err,
		)

		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{
				"error":
					"Ошибка получения черновика этапа синтеза аспирина",
			},
		)

		return
	}

	if draft != nil {

		h.prepareAspirinStageMediaForView(
			draft,
		)
	}

	ctx.HTML(
		http.StatusOK,
		"aspirin_synthesis_stage_draft.html",
		gin.H{
			"aspirinSynthesisStageDraft":
				draft,

			"defaultAspirinStageImageURL":
				defaultAspirinStageImageURL,

			"defaultAspirinStageVideoURL":
				defaultAspirinStageVideoURL,
		},
	)
}

func (h *Handler) ShowAspirinSynthesisStages(
	ctx *gin.Context,
) {

	durationStr :=
		ctx.Query(
			"max_synthesis_duration_minutes",
		)

	var stages []repository.AspirinSynthesisStage
	var err error

	if durationStr == "" {

		stages, err =
			h.Repository.
				GetPublishedAspirinSynthesisStages()

	} else {

		maxDuration, convErr :=
			strconv.Atoi(
				durationStr,
			)

		if convErr != nil {

			ctx.JSON(
				http.StatusBadRequest,
				gin.H{
					"error":
						"Неверное значение длительности этапа синтеза",
				},
			)

			return
		}

		stages, err =
			h.Repository.
				FilterPublishedAspirinSynthesisStagesByDuration(
					maxDuration,
				)
	}

	if err != nil {

		logrus.Error(
			"Ошибка получения этапов синтеза: ",
			err,
		)

		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{
				"error":
					"Ошибка получения этапов синтеза",
			},
		)

		return
	}

	h.prepareAspirinStagesMediaForView(
		stages,
	)

	ctx.HTML(
		http.StatusOK,
		"aspirin_synthesis_stages.html",
		gin.H{
			"aspirinSynthesisStages":
				stages,

			"maxSynthesisDurationMinutes":
				durationStr,
		},
	)
}

// POST №1 через ORM.
func (h *Handler) CreateAspirinSynthesisStageDraft(
	ctx *gin.Context,
) {

	stageName :=
		strings.TrimSpace(
			ctx.PostForm(
				"aspirin_synthesis_stage_name",
			),
		)

	if stageName == "" {

		ctx.JSON(
			http.StatusBadRequest,
			gin.H{
				"error":
					"Укажите название этапа синтеза аспирина",
			},
		)

		return
	}

	err :=
		h.Repository.
			CreateAspirinSynthesisStageDraft(
				stageName,
				GetCurrentAspirinProductionUserID(),
			)

	if err != nil {

		if errors.Is(
			err,
			repository.ErrAspirinSynthesisDraftExists,
		) {

			ctx.Redirect(
				http.StatusFound,
				"/aspirin-synthesis-stage-draft",
			)

			return
		}

		logrus.Error(
			"Ошибка создания черновика этапа синтеза аспирина: ",
			err,
		)

		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{
				"error":
					"Ошибка создания черновика этапа синтеза аспирина",
			},
		)

		return
	}

	ctx.Redirect(
		http.StatusFound,
		"/aspirin-synthesis-stage-draft",
	)
}

// POST №2 через ORM.
func (h *Handler) PublishAspirinSynthesisStageDraft(
	ctx *gin.Context,
) {

	stageID, err :=
		strconv.Atoi(
			ctx.PostForm(
				"aspirin_synthesis_stage_id",
			),
		)

	if err != nil ||
		stageID < 1 {

		ctx.JSON(
			http.StatusBadRequest,
			gin.H{
				"error":
					"Некорректный ID этапа синтеза аспирина",
			},
		)

		return
	}

	description :=
		strings.TrimSpace(
			ctx.PostForm(
				"aspirin_synthesis_stage_description",
			),
		)

	if description == "" {

		ctx.JSON(
			http.StatusBadRequest,
			gin.H{
				"error":
					"Заполните краткое описание этапа синтеза аспирина",
			},
		)

		return
	}

	durationMinutes, err :=
		strconv.Atoi(
			ctx.PostForm(
				"synthesis_duration_minutes",
			),
		)

	if err != nil ||
		durationMinutes < 1 {

		ctx.JSON(
			http.StatusBadRequest,
			gin.H{
				"error":
					"Длительность синтеза должна быть положительным числом",
			},
		)

		return
	}

	pureAspirinYieldPercent, err :=
		strconv.ParseFloat(
			ctx.PostForm(
				"pure_aspirin_yield_percent",
			),
			64,
		)

	if err != nil ||
		pureAspirinYieldPercent < 0 ||
		pureAspirinYieldPercent > 100 {

		ctx.JSON(
			http.StatusBadRequest,
			gin.H{
				"error":
					"Выход чистого аспирина должен быть от 0 до 100%",
			},
		)

		return
	}

	_, err =
	    h.Repository.
	    	PublishAspirinSynthesisStageDraft(
	    		GetCurrentAspirinProductionUserID(),
	    		description,
	    		durationMinutes,
	    		pureAspirinYieldPercent,
	    	)

	if err != nil {

		if errors.Is(
			err,
			repository.ErrAspirinSynthesisStageNotFound,
		) {

			ctx.JSON(
				http.StatusNotFound,
				gin.H{
					"error":
						"Черновик этапа синтеза аспирина не найден",
				},
			)

			return
		}

		logrus.Error(
			"Ошибка публикации этапа синтеза аспирина: ",
			err,
		)

		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{
				"error":
					"Ошибка публикации этапа синтеза аспирина",
			},
		)

		return
	}

	ctx.Redirect(
		http.StatusFound,
		"/aspirin-synthesis-stages",
	)
}

// POST №3 ЛР-2: логическое удаление чистым SQL.
func (h *Handler) DeleteAspirinSynthesisStage(
	ctx *gin.Context,
) {

	stageID, err :=
		strconv.Atoi(
			ctx.PostForm(
				"aspirin_synthesis_stage_id",
			),
		)

	if err != nil ||
		stageID < 1 {

		ctx.JSON(
			http.StatusBadRequest,
			gin.H{
				"error":
					"Некорректный ID этапа синтеза аспирина",
			},
		)

		return
	}

	err =
		h.Repository.
			DeleteAspirinSynthesisStage(
				stageID,
			)

	if err != nil {

		if errors.Is(
			err,
			repository.ErrAspirinSynthesisStageNotFound,
		) {

			ctx.JSON(
				http.StatusNotFound,
				gin.H{
					"error":
						"Этап синтеза аспирина не найден",
				},
			)

			return
		}

		logrus.Error(
			"Ошибка логического удаления этапа синтеза аспирина: ",
			err,
		)

		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{
				"error":
					"Ошибка удаления этапа синтеза аспирина",
			},
		)

		return
	}

	ctx.Redirect(
		http.StatusFound,
		"/aspirin-synthesis-stages",
	)
}