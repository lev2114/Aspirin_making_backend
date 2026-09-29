package api

import (
	"aspirin_back/internal/app/dsn"
	"aspirin_back/internal/app/handler"
	"aspirin_back/internal/app/repository"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func StartServer() {
	_ = godotenv.Load()

	repo, err :=
		repository.New(
			repository.RepositorySettings{
				PostgresDSN: dsn.FromEnv(),

				MinioEndpoint: os.Getenv(
					"MINIO_ENDPOINT",
				),

				MinioAccessKey: os.Getenv(
					"MINIO_ACCESS_KEY",
				),

				MinioSecretKey: os.Getenv(
					"MINIO_SECRET_KEY",
				),

				MinioBucketName: os.Getenv(
					"MINIO_BUCKET_NAME",
				),
			},
		)

	if err != nil {
		logrus.Fatalf(
			"Ошибка инициализации Repository: %v",
			err,
		)
	}

	h := handler.NewHandler(repo)

	r := gin.Default()

	r.LoadHTMLGlob("templates/*")

	r.Static(
		"/static",
		"./resources",
	)

	// ЛР-2.
	r.GET(
		"/aspirin-synthesis-stage",
		h.ShowAspirinSynthesisStage,
	)

	r.GET(
		"/aspirin-synthesis-stage-draft",
		h.ShowAspirinSynthesisStageDraft,
	)

	r.GET(
		"/aspirin-synthesis-stages",
		h.ShowAspirinSynthesisStages,
	)

	r.POST(
		"/aspirin-synthesis-stage-draft",
		h.CreateAspirinSynthesisStageDraft,
	)

	r.POST(
		"/aspirin-synthesis-stage-publication",
		h.PublishAspirinSynthesisStageDraft,
	)

	r.POST(
		"/aspirin-synthesis-stage-deletion",
		h.DeleteAspirinSynthesisStage,
	)

	// API лабораторной №3.
    apiGroup := r.Group("/api")
    
    {
    	// -------------------------
    	// Домен услуг
    	// -------------------------
    
    	// Список опубликованных услуг + фильтрация.
    	apiGroup.GET(
    		"/aspirin-stages",
    		h.GetAspirinSynthesisStagesAPI,
    	)
    
    	// Получение единственного черновика текущего пользователя.
    	apiGroup.GET(
    		"/aspirin-stages/draft",
    		h.GetAspirinSynthesisStageDraftAPI,
    	)
    
    	// Создание черновика с изображением и видео.
    	apiGroup.POST(
    		"/aspirin-stages",
    		h.CreateAspirinSynthesisStageAPI,
    	)
    
    	// Публикация единственного черновика.
    	apiGroup.PUT(
    		"/aspirin-stages/draft/publication",
    		h.PublishAspirinSynthesisStageAPI,
    	)
    
    	// Лента без ID.
    	apiGroup.GET(
    		"/aspirin-stages/feed",
    		h.GetAspirinSynthesisStageFeedAPI,
    	)
    
    	// Лента по ID, в том числе ?next=true.
    	apiGroup.GET(
    		"/aspirin-stages/feed/:id",
    		h.GetAspirinSynthesisStageFeedByIDAPI,
    	)
    
    	// Поставить/снять лайк.
    	apiGroup.POST(
    		"/aspirin-stages/:id/like",
    		h.SetAspirinSynthesisStageLikeAPI,
    	)
    
    	// Soft delete только собственной услуги.
    	apiGroup.DELETE(
    		"/aspirin-stages/:id",
    		h.DeleteAspirinSynthesisStageAPI,
    	)
    
    	// -------------------------
    	// Домен пользователей
    	// -------------------------
    
    	// Регистрация.
    	apiGroup.POST(
    		"/users",
    		h.RegisterAspirinProductionUserAPI,
    	)
    
    	// Заглушка ЛР-4.
    	apiGroup.POST(
    		"/authentication",
    		h.AuthenticateAspirinProductionUserAPI,
    	)
    
    	// Заглушка ЛР-4.
    	apiGroup.POST(
    		"/deauthentication",
    		h.DeauthenticateAspirinProductionUserAPI,
    	)
    }

	logrus.Info(
		"Сервер запущен на http://localhost:8080/aspirin-synthesis-stage",
	)

	if err :=
		r.Run(
			"0.0.0.0:8080",
		); err != nil {

		logrus.Fatalf(
			"Ошибка запуска HTTP-сервера: %v",
			err,
		)
	}
}
