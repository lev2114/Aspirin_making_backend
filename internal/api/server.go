package api

import (
	"aspirin_back/internal/app/dsn"
	"aspirin_back/internal/app/handler"
	"aspirin_back/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func StartServer() {
	_ = godotenv.Load()

	repo, err := repository.New(dsn.FromEnv())
	if err != nil {
		logrus.Fatalf("Ошибка подключения к PostgreSQL: %v", err)
	}

	h := handler.NewHandler(repo)
	r := gin.Default()

	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./resources")

	// Ровно 3 GET метода лабораторной №2.
	r.GET("/aspirin-synthesis-stage", h.ShowAspirinSynthesisStage)
	r.GET("/aspirin-synthesis-stage-draft", h.ShowAspirinSynthesisStageDraft)
	r.GET("/aspirin-synthesis-stages", h.ShowAspirinSynthesisStages)

	// Ровно 3 POST метода лабораторной №2.
	r.POST("/aspirin-synthesis-stage-draft", h.CreateAspirinSynthesisStageDraft)
	r.POST("/aspirin-synthesis-stage-publication", h.PublishAspirinSynthesisStageDraft)
	r.POST("/aspirin-synthesis-stage-deletion", h.DeleteAspirinSynthesisStage)

	logrus.Info("Сервер запущен на http://localhost:8080/aspirin-synthesis-stage")
	if err := r.Run("0.0.0.0:8080"); err != nil {
		logrus.Fatalf("Ошибка запуска HTTP-сервера: %v", err)
	}
}
