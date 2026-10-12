package api

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	_ "aspirin_back/docs"
	"aspirin_back/internal/app/auth"
	"aspirin_back/internal/app/dsn"
	"aspirin_back/internal/app/handler"
	"aspirin_back/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func StartServer() {
	_ = godotenv.Load()
	authentication, redisClient, err := newAspirinProductionAuthentication()
	if err != nil {
		logrus.Fatalf("Ошибка настройки аутентификации: %v", err)
	}
	defer redisClient.Close()
	repo, err := repository.New(repository.RepositorySettings{
		PostgresDSN: dsn.FromEnv(), MinioEndpoint: os.Getenv("MINIO_ENDPOINT"),
		MinioAccessKey: os.Getenv("MINIO_ACCESS_KEY"), MinioSecretKey: os.Getenv("MINIO_SECRET_KEY"),
		MinioBucketName: os.Getenv("MINIO_BUCKET_NAME"),
	})
	if err != nil {
		logrus.Fatalf("Ошибка инициализации Repository: %v", err)
	}
	h := handler.NewHandler(repo, authentication)
	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./resources")
	RegisterAspirinProductionRoutes(r, h)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	logrus.Info("API: http://localhost:8080/api; Swagger: http://localhost:8080/swagger/index.html")
	if err := r.Run("0.0.0.0:8080"); err != nil {
		logrus.Fatalf("Ошибка запуска HTTP-сервера: %v", err)
	}
}

// The same access rules apply to the legacy SSR routes and the REST API.
func RegisterAspirinProductionRoutes(r *gin.Engine, h *handler.Handler) {
	optional := h.OptionalAspirinProductionAuthentication
	required := h.RequireAspirinProductionAuthentication
	r.GET("/aspirin-synthesis-stage", optional, h.ShowAspirinSynthesisStage)
	r.GET("/aspirin-synthesis-stages", optional, h.ShowAspirinSynthesisStages)
	r.GET("/aspirin-synthesis-stage-draft", required, h.ShowAspirinSynthesisStageDraft)
	r.POST("/aspirin-synthesis-stage-draft", required, h.CreateAspirinSynthesisStageDraft)
	r.POST("/aspirin-synthesis-stage-publication", required, h.PublishAspirinSynthesisStageDraft)
	r.POST("/aspirin-synthesis-stage-deletion", required, h.DeleteAspirinSynthesisStage)
	apiGroup := r.Group("/api")
	apiGroup.GET("/aspirin-stages", optional, h.GetAspirinSynthesisStagesAPI)
	apiGroup.GET("/aspirin-stages/feed", optional, h.GetAspirinSynthesisStageFeedAPI)
	apiGroup.GET("/aspirin-stages/feed/:id", optional, h.GetAspirinSynthesisStageFeedByIDAPI)
	apiGroup.POST("/users", h.RegisterAspirinProductionUserAPI)
	apiGroup.POST("/authentication", h.AuthenticateAspirinProductionUserAPI)
	creator := apiGroup.Group("", required)
	creator.GET("/aspirin-stages/draft", h.GetAspirinSynthesisStageDraftAPI)
	creator.POST("/aspirin-stages", h.CreateAspirinSynthesisStageAPI)
	creator.PUT("/aspirin-stages/draft/publication", h.PublishAspirinSynthesisStageAPI)
	creator.POST("/aspirin-stages/:id/like", h.SetAspirinSynthesisStageLikeAPI)
	creator.DELETE("/aspirin-stages/:id", h.DeleteAspirinSynthesisStageAPI)
	creator.POST("/deauthentication", h.DeauthenticateAspirinProductionUserAPI)
}

func newAspirinProductionAuthentication() (*auth.AspirinProductionAuthentication, *redis.Client, error) {
	lifetime := time.Hour
	if value := os.Getenv("JWT_TTL"); value != "" {
		parsed, err := time.ParseDuration(value)
		if err != nil {
			return nil, nil, fmt.Errorf("JWT_TTL: %w", err)
		}
		lifetime = parsed
	}
	redisDB := 0
	if value := os.Getenv("REDIS_DB"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 0 {
			return nil, nil, fmt.Errorf("REDIS_DB должен быть неотрицательным целым числом")
		}
		redisDB = parsed
	}
	address := os.Getenv("REDIS_ADDR")
	if address == "" {
		address = "localhost:6379"
	}
	client := redis.NewClient(&redis.Options{Addr: address, Password: os.Getenv("REDIS_PASSWORD"), DB: redisDB,
		DialTimeout: 3 * time.Second, ReadTimeout: 3 * time.Second, WriteTimeout: 3 * time.Second})
	authentication, err := auth.New(os.Getenv("JWT_SECRET"), lifetime, auth.RedisAspirinProductionTokenBlacklist{Client: client})
	if err != nil {
		client.Close()
		return nil, nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, nil, fmt.Errorf("Redis: %w", err)
	}
	return authentication, client, nil
}
