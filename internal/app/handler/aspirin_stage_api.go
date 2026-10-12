package handler

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"aspirin_back/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type AspirinSynthesisStageResponse struct {
	AspirinSynthesisStageID          int      `json:"aspirin_stage_id"`
	AspirinSynthesisStageName        string   `json:"aspirin_stage_name"`
	AspirinSynthesisStageDescription string   `json:"aspirin_stage_description"`
	AspirinSynthesisStageImage       *string  `json:"aspirin_stage_image"`
	AspirinSynthesisStageVideo       *string  `json:"aspirin_stage_video"`
	SynthesisDurationMinutes         *int     `json:"synthesis_duration_minutes"`
	PureAspirinYieldPercent          *float64 `json:"pure_aspirin_yield_percent"`
	AspirinSynthesisLikeCount        int      `json:"aspirin_synthesis_like_count"`
	IsCreator                        int      `json:"is_creator"`
	IsLiked                          int      `json:"is_liked"`
}

type PublishAspirinSynthesisStageRequest struct {
	AspirinSynthesisStageDescription string   `json:"aspirin_stage_description" binding:"required"`
	SynthesisDurationMinutes         int      `json:"synthesis_duration_minutes" binding:"required"`
	PureAspirinYieldPercent          *float64 `json:"pure_aspirin_yield_percent" binding:"required"`
}

func (h *Handler) serializeAspirinSynthesisStage(
	stage repository.AspirinSynthesisStage,
	currentUserID int,
) AspirinSynthesisStageResponse {

	isCreator := 0

	if currentUserID > 0 && stage.AspirinProductionUserID == currentUserID {
		isCreator = 1
	}

	isLiked := 0
	for _, like := range stage.AspirinSynthesisStageLikes {
		if currentUserID > 0 && like.AspirinProductionUserID == currentUserID {
			isLiked = 1
			break
		}
	}
	return AspirinSynthesisStageResponse{
		IsLiked:                 isLiked,
		AspirinSynthesisStageID: stage.AspirinSynthesisStageID,

		AspirinSynthesisStageName: stage.AspirinSynthesisStageName,

		AspirinSynthesisStageDescription: stage.AspirinSynthesisStageDescription,

		AspirinSynthesisStageImage: h.Repository.AspirinStageMediaURL(
			stage.AspirinSynthesisStageImageURL,
		),

		AspirinSynthesisStageVideo: h.Repository.AspirinStageMediaURL(
			stage.AspirinSynthesisStageVideoURL,
		),

		SynthesisDurationMinutes: stage.SynthesisDurationMinutes,

		PureAspirinYieldPercent: stage.PureAspirinYieldPercent,

		AspirinSynthesisLikeCount: stage.AspirinSynthesisLikeCount,

		IsCreator: isCreator,
	}
}

// GET /api/aspirin-stages
// @Summary Опубликованные этапы синтеза аспирина
// @Tags Этапы синтеза аспирина
// @Produce json
// @Param Authorization header string false "Необязательный Bearer JWT для is_creator и is_liked"
// @Param max_synthesis_duration_minutes query int false "Максимальная длительность в минутах" minimum(0)
// @Success 200 {object} AspirinSynthesisStagesDataResponse
// @Failure 400,500 {object} AspirinProductionMessageResponse
// @Router /aspirin-stages [get]
func (h *Handler) GetAspirinSynthesisStagesAPI(
	ctx *gin.Context,
) {

	currentUserID :=
		GetCurrentAspirinProductionUserID(ctx)

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
			strconv.Atoi(durationStr)

		if convErr != nil ||
			maxDuration < 0 {

			ctx.JSON(
				http.StatusBadRequest,
				gin.H{
					"message": "Неверное значение длительности этапа синтеза",
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
			"Ошибка получения этапов синтеза через API: ",
			err,
		)

		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{
				"message": "Ошибка получения этапов синтеза",
			},
		)

		return
	}

	response :=
		make(
			[]AspirinSynthesisStageResponse,
			0,
			len(stages),
		)

	for _, stage := range stages {

		response =
			append(
				response,
				h.serializeAspirinSynthesisStage(
					stage,
					currentUserID,
				),
			)
	}

	ctx.JSON(
		http.StatusOK,
		gin.H{
			"data": response,
		},
	)
}

// GET /api/aspirin-stages/draft
// @Summary Черновик текущего пользователя производства аспирина
// @Tags Этапы синтеза аспирина
// @Produce json
// @Security AspirinProductionBearer
// @Success 200 {object} AspirinSynthesisStageDataResponse
// @Failure 401,503,500 {object} AspirinProductionMessageResponse
// @Router /aspirin-stages/draft [get]
func (h *Handler) GetAspirinSynthesisStageDraftAPI(
	ctx *gin.Context,
) {

	currentUserID :=
		GetCurrentAspirinProductionUserID(ctx)

	draft, err :=
		h.Repository.
			GetAspirinSynthesisStageDraft(
				currentUserID,
			)

	if err != nil {

		logrus.Error(
			"Ошибка получения черновика через API: ",
			err,
		)

		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{
				"message": "Ошибка получения черновика этапа синтеза",
			},
		)

		return
	}

	if draft == nil {

		ctx.JSON(
			http.StatusOK,
			gin.H{
				"data": nil,
			},
		)

		return
	}

	ctx.JSON(
		http.StatusOK,
		gin.H{
			"data": h.serializeAspirinSynthesisStage(
				*draft,
				currentUserID,
			),
		},
	)
}

const (
	maxAspirinStageImageSize = 10 << 20
	maxAspirinStageVideoSize = 50 << 20
)

func detectUploadedFileContentType(
	header *multipart.FileHeader,
) (string, error) {

	file, err := header.Open()
	if err != nil {
		return "", err
	}

	defer file.Close()

	buffer := make([]byte, 512)

	n, err := file.Read(buffer)

	if err != nil &&
		!errors.Is(err, io.EOF) {

		return "", err
	}

	if n == 0 {
		return "", errors.New("пустой файл")
	}

	return http.DetectContentType(
		buffer[:n],
	), nil
}

func validateAspirinStageImage(
	header *multipart.FileHeader,
) (string, error) {

	if header.Size >
		maxAspirinStageImageSize {

		return "",
			fmt.Errorf(
				"размер изображения не должен превышать 10 МБ",
			)
	}

	contentType, err :=
		detectUploadedFileContentType(
			header,
		)

	if err != nil {
		return "", err
	}

	switch contentType {

	case "image/jpeg",
		"image/png",
		"image/gif",
		"image/webp":

		return contentType, nil

	default:

		return "",
			fmt.Errorf(
				"неподдерживаемый формат изображения",
			)
	}
}

func validateAspirinStageVideo(
	header *multipart.FileHeader,
) (string, error) {

	if header.Size >
		maxAspirinStageVideoSize {

		return "",
			fmt.Errorf(
				"размер видео не должен превышать 50 МБ",
			)
	}

	contentType, err :=
		detectUploadedFileContentType(
			header,
		)

	if err != nil {
		return "", err
	}

	switch contentType {

	case "video/mp4",
		"video/webm",
		"video/quicktime":

		return contentType, nil

	default:

		return "",
			fmt.Errorf(
				"неподдерживаемый формат видео",
			)
	}
}

// POST /api/aspirin-stages
// @Summary Создание черновика этапа синтеза аспирина
// @Tags Этапы синтеза аспирина
// @Accept multipart/form-data
// @Produce json
// @Security AspirinProductionBearer
// @Param aspirin_synthesis_stage_name formData string true "Название этапа"
// @Param aspirin_synthesis_stage_image formData file false "JPEG, PNG, GIF или WebP, до 10 МБ"
// @Param aspirin_synthesis_stage_video formData file false "MP4, WebM или QuickTime, до 50 МБ"
// @Success 201 {object} AspirinSynthesisStageDataResponse
// @Failure 400,401,409,500,503 {object} AspirinProductionMessageResponse
// @Router /aspirin-stages [post]
func (h *Handler) CreateAspirinSynthesisStageAPI(
	ctx *gin.Context,
) {

	currentUserID :=
		GetCurrentAspirinProductionUserID(ctx)

	existingDraft, err :=
		h.Repository.
			GetAspirinSynthesisStageDraft(
				currentUserID,
			)

	if err != nil {

		logrus.Error(
			"Ошибка проверки черновика: ",
			err,
		)

		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{
				"message": "Ошибка проверки существующего черновика",
			},
		)

		return
	}

	if existingDraft != nil {

		ctx.JSON(
			http.StatusConflict,
			gin.H{
				"message": "У пользователя уже есть черновик",
			},
		)

		return
	}

	if err :=
		ctx.Request.ParseMultipartForm(
			64 << 20,
		); err != nil {

		ctx.JSON(
			http.StatusBadRequest,
			gin.H{
				"message": "Некорректная multipart-форма",
			},
		)

		return
	}

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
				"message": "Не указано название этапа синтеза",
			},
		)

		return
	}

	var imageFilename *string
	var videoFilename *string

	imageHeader, imageErr :=
		ctx.FormFile(
			"aspirin_synthesis_stage_image",
		)

	if imageErr != nil &&
		imageErr != http.ErrMissingFile {

		ctx.JSON(
			http.StatusBadRequest,
			gin.H{
				"message": "Ошибка получения изображения",
			},
		)

		return
	}

	if imageErr == nil {

		contentType, err :=
			validateAspirinStageImage(
				imageHeader,
			)

		if err != nil {

			ctx.JSON(
				http.StatusBadRequest,
				gin.H{
					"message": err.Error(),
				},
			)

			return
		}

		filename, err :=
			h.Repository.
				UploadAspirinStageFile(
					imageHeader,
					"image",
					contentType,
				)

		if err != nil {

			logrus.Error(
				"Ошибка загрузки изображения в MinIO: ",
				err,
			)

			ctx.JSON(
				http.StatusInternalServerError,
				gin.H{
					"message": "Ошибка сохранения изображения",
				},
			)

			return
		}

		imageFilename = &filename
	}

	videoHeader, videoErr :=
		ctx.FormFile(
			"aspirin_synthesis_stage_video",
		)

	if videoErr != nil &&
		videoErr != http.ErrMissingFile {

		if imageFilename != nil {

			_ = h.Repository.
				DeleteAspirinStageFile(
					*imageFilename,
				)
		}

		ctx.JSON(
			http.StatusBadRequest,
			gin.H{
				"message": "Ошибка получения видео",
			},
		)

		return
	}

	if videoErr == nil {

		contentType, err :=
			validateAspirinStageVideo(
				videoHeader,
			)

		if err != nil {

			if imageFilename != nil {

				_ = h.Repository.
					DeleteAspirinStageFile(
						*imageFilename,
					)
			}

			ctx.JSON(
				http.StatusBadRequest,
				gin.H{
					"message": err.Error(),
				},
			)

			return
		}

		filename, err :=
			h.Repository.
				UploadAspirinStageFile(
					videoHeader,
					"video",
					contentType,
				)

		if err != nil {

			if imageFilename != nil {

				_ = h.Repository.
					DeleteAspirinStageFile(
						*imageFilename,
					)
			}

			logrus.Error(
				"Ошибка загрузки видео в MinIO: ",
				err,
			)

			ctx.JSON(
				http.StatusInternalServerError,
				gin.H{
					"message": "Ошибка сохранения видео",
				},
			)

			return
		}

		videoFilename = &filename
	}

	stage, err :=
		h.Repository.
			CreateAspirinSynthesisStageDraftWithMedia(
				stageName,
				currentUserID,
				imageFilename,
				videoFilename,
			)

	if err != nil {

		if imageFilename != nil {

			_ = h.Repository.
				DeleteAspirinStageFile(
					*imageFilename,
				)
		}

		if videoFilename != nil {

			_ = h.Repository.
				DeleteAspirinStageFile(
					*videoFilename,
				)
		}

		if errors.Is(
			err,
			repository.ErrAspirinSynthesisDraftExists,
		) {

			ctx.JSON(
				http.StatusConflict,
				gin.H{
					"message": "У пользователя уже есть черновик",
				},
			)

			return
		}

		logrus.Error(
			"Ошибка создания этапа синтеза: ",
			err,
		)

		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{
				"message": "Ошибка создания этапа синтеза",
			},
		)

		return
	}

	ctx.JSON(
		http.StatusCreated,
		gin.H{
			"data": h.serializeAspirinSynthesisStage(
				*stage,
				currentUserID,
			),
		},
	)
}

// PUT /api/aspirin-stages/draft/publication
// @Summary Публикация черновика текущего пользователя
// @Tags Этапы синтеза аспирина
// @Accept json
// @Produce json
// @Security AspirinProductionBearer
// @Param aspirin_stage_publication body PublishAspirinSynthesisStageRequest true "Описание, длительность и выход аспирина"
// @Success 200 {object} AspirinSynthesisStagePublicationResponse
// @Failure 400,401,404,500,503 {object} AspirinProductionMessageResponse
// @Router /aspirin-stages/draft/publication [put]
func (h *Handler) PublishAspirinSynthesisStageAPI(
	ctx *gin.Context,
) {

	currentUserID :=
		GetCurrentAspirinProductionUserID(ctx)

	var request PublishAspirinSynthesisStageRequest

	if err :=
		ctx.ShouldBindJSON(
			&request,
		); err != nil {

		ctx.JSON(
			http.StatusBadRequest,
			gin.H{
				"message": "Некорректные данные публикации",
			},
		)

		return
	}

	request.AspirinSynthesisStageDescription =
		strings.TrimSpace(
			request.AspirinSynthesisStageDescription,
		)

	if request.AspirinSynthesisStageDescription == "" {

		ctx.JSON(
			http.StatusBadRequest,
			gin.H{
				"message": "Описание этапа не может быть пустым",
			},
		)

		return
	}

	if request.SynthesisDurationMinutes < 1 {

		ctx.JSON(
			http.StatusBadRequest,
			gin.H{
				"message": "Длительность синтеза должна быть положительным числом",
			},
		)

		return
	}

	if *request.PureAspirinYieldPercent < 0 ||
		*request.PureAspirinYieldPercent > 100 {

		ctx.JSON(
			http.StatusBadRequest,
			gin.H{
				"message": "Выход чистого аспирина должен быть от 0 до 100 процентов",
			},
		)

		return
	}

	stage, err :=
		h.Repository.
			PublishAspirinSynthesisStageDraft(
				currentUserID,
				request.AspirinSynthesisStageDescription,
				request.SynthesisDurationMinutes,
				*request.PureAspirinYieldPercent,
			)

	if err != nil {

		if errors.Is(
			err,
			repository.ErrAspirinSynthesisStageNotFound,
		) {

			ctx.JSON(
				http.StatusNotFound,
				gin.H{
					"message": "Черновик текущего пользователя не найден",
				},
			)

			return
		}

		logrus.Error(
			"Ошибка публикации этапа через API: ",
			err,
		)

		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{
				"message": "Ошибка публикации этапа синтеза",
			},
		)

		return
	}

	ctx.JSON(
		http.StatusOK,
		gin.H{
			"message": "Этап синтеза опубликован",

			"data": h.serializeAspirinSynthesisStage(
				*stage,
				currentUserID,
			),
		},
	)
}

// GET /api/aspirin-stages/feed
// @Summary Первый опубликованный этап в ленте
// @Tags Лента синтеза аспирина
// @Produce json
// @Param Authorization header string false "Необязательный Bearer JWT для is_creator и is_liked"
// @Success 200 {object} AspirinSynthesisStageDataResponse
// @Failure 404,500 {object} AspirinProductionMessageResponse
// @Router /aspirin-stages/feed [get]
func (h *Handler) GetAspirinSynthesisStageFeedAPI(
	ctx *gin.Context,
) {

	currentUserID :=
		GetCurrentAspirinProductionUserID(ctx)

	stage, err :=
		h.Repository.
			GetAspirinSynthesisStageForFeed(
				nil,
				false,
			)

	if err != nil {

		if errors.Is(
			err,
			repository.ErrAspirinSynthesisStageNotFound,
		) {

			ctx.JSON(
				http.StatusNotFound,
				gin.H{
					"message": "В ленте нет опубликованных этапов",
				},
			)

			return
		}

		logrus.Error(
			"Ошибка получения ленты: ",
			err,
		)

		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{
				"message": "Ошибка получения ленты",
			},
		)

		return
	}

	ctx.JSON(
		http.StatusOK,
		gin.H{
			"data": h.serializeAspirinSynthesisStage(
				*stage,
				currentUserID,
			),
		},
	)
}

// GET /api/aspirin-stages/feed/:id
// GET /api/aspirin-stages/feed/:id?next=true
// @Summary Этап ленты по ID или следующий этап
// @Tags Лента синтеза аспирина
// @Produce json
// @Param Authorization header string false "Необязательный Bearer JWT для is_creator и is_liked"
// @Param id path int true "ID этапа синтеза" minimum(1)
// @Param next query bool false "Следующий этап; после последнего возвращается первый"
// @Success 200 {object} AspirinSynthesisStageDataResponse
// @Failure 400,404,500 {object} AspirinProductionMessageResponse
// @Router /aspirin-stages/feed/{id} [get]
func (h *Handler) GetAspirinSynthesisStageFeedByIDAPI(
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

	next := false

	nextText :=
		strings.TrimSpace(
			ctx.Query("next"),
		)

	if nextText != "" {

		nextValue, err :=
			strconv.ParseBool(
				nextText,
			)

		if err != nil {

			ctx.JSON(
				http.StatusBadRequest,
				gin.H{
					"message": "Параметр next должен быть true или false",
				},
			)

			return
		}

		next = nextValue
	}

	stage, err :=
		h.Repository.
			GetAspirinSynthesisStageForFeed(
				&stageID,
				next,
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
			"Ошибка получения элемента ленты: ",
			err,
		)

		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{
				"message": "Ошибка получения элемента ленты",
			},
		)

		return
	}

	ctx.JSON(
		http.StatusOK,
		gin.H{
			"data": h.serializeAspirinSynthesisStage(
				*stage,
				currentUserID,
			),
		},
	)
}

// DELETE /api/aspirin-stages/:id
// @Summary Логическое удаление своего этапа синтеза
// @Tags Этапы синтеза аспирина
// @Produce json
// @Security AspirinProductionBearer
// @Param id path int true "ID этапа синтеза" minimum(1)
// @Success 200 {object} AspirinProductionMessageResponse
// @Failure 400,401,404,500,503 {object} AspirinProductionMessageResponse
// @Router /aspirin-stages/{id} [delete]
func (h *Handler) DeleteAspirinSynthesisStageAPI(
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

	err =
		h.Repository.
			DeleteAspirinSynthesisStageForUser(
				stageID,
				currentUserID,
			)

	if err != nil {

		if errors.Is(
			err,
			repository.ErrAspirinSynthesisStageNotFound,
		) {

			ctx.JSON(
				http.StatusNotFound,
				gin.H{
					"message": "Этап не найден или не принадлежит текущему пользователю",
				},
			)

			return
		}

		logrus.Error(
			"Ошибка удаления этапа через API: ",
			err,
		)

		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{
				"message": "Ошибка удаления этапа",
			},
		)

		return
	}

	ctx.JSON(
		http.StatusOK,
		gin.H{
			"message": "Этап синтеза логически удалён",
		},
	)
}

// Response envelopes shared by the API and its Swagger specification.
type AspirinSynthesisStagesDataResponse struct {
	Data []AspirinSynthesisStageResponse `json:"data"`
}
type AspirinSynthesisStageDataResponse struct {
	Data *AspirinSynthesisStageResponse `json:"data"`
}
type AspirinSynthesisStagePublicationResponse struct {
	Message string                        `json:"message"`
	Data    AspirinSynthesisStageResponse `json:"data"`
}
