package handler

import (
	"aspirin_back/internal/app/repository"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{Repository: r}
}

func (h *Handler) GetFeed(ctx *gin.Context) {
	stages, _ := h.Repository.GetAspirinStages()

	// Ищем текущий этап по ID из URL, по умолчанию показываем первый
	idQuery := ctx.Query("id")
	currentIndex := 0
	if idQuery != "" {
		id, _ := strconv.Atoi(idQuery)
		for i, s := range stages {
			if s.ID == id {
				currentIndex = i
				break
			}
		}
	}

	currentStage := stages[currentIndex]

	// Определяем ID следующего этапа (зацикливаем на начало, если это последний)
	nextID := stages[0].ID
	if currentIndex+1 < len(stages) {
		nextID = stages[currentIndex+1].ID
	}

	ctx.HTML(http.StatusOK, "feed.html", gin.H{
		"stage":  currentStage,
		"nextID": nextID,
	})
}

func (h *Handler) GetGrid(ctx *gin.Context) {
	stages, _ := h.Repository.GetAspirinStages()

	// Логика фильтрации по ползунку
	durationQuery := ctx.Query("duration")
	currentDuration := 30 // Максимальное значение по умолчанию

	if durationQuery != "" {
		val, err := strconv.Atoi(durationQuery)
		if err == nil {
			currentDuration = val
		}
	}

	var filtered []repository.AspirinStage
	for _, s := range stages {
		if s.DurationMin <= currentDuration {
			filtered = append(filtered, s)
		}
	}

	ctx.HTML(http.StatusOK, "grid.html", gin.H{
		"stages":          filtered,
		"currentDuration": currentDuration,
	})
}

func (h *Handler) GetStageDetail(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error("Ошибка преобразования ID: ", err)
	}

	stages, _ := h.Repository.GetAspirinStages()

	var currentStage repository.AspirinStage
	currentIndex := 0

	for i, stage := range stages {
		if stage.ID == id {
			currentStage = stage
			currentIndex = i
			break
		}
	}

	nextID := stages[0].ID
	if currentIndex+1 < len(stages) {
		nextID = stages[currentIndex+1].ID
	}

	ctx.HTML(http.StatusOK, "detail.html", gin.H{
		"stage":  currentStage,
		"nextID": nextID,
	})
}

func (h *Handler) GetAddPage(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "add.html", gin.H{})
}
