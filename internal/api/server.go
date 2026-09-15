package api

import (
	"aspirin_back/internal/app/handler"
	"aspirin_back/internal/app/repository"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func StartServer() {
	repo, err := repository.NewRepository()
	if err != nil {
		logrus.Error("ошибка инициализации репозитория")
	}

	h := handler.NewHandler(repo)
	r := gin.Default()

	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./resources")

	r.GET("/", h.GetFeed)
	r.GET("/grid", h.GetGrid)
	r.GET("/add", h.GetAddPage)
	r.GET("/stage/:id", h.GetStageDetail)

	log.Println("Сервер запущен на :8080")
	r.Run("0.0.0.0:8080")
}
