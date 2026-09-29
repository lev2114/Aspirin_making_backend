package repository

import (
	"context"
	"errors"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	AspirinSynthesisStageStatusDraft     = "draft"
	AspirinSynthesisStageStatusPublished = "published"
	AspirinSynthesisStageStatusDeleted   = "deleted"
)

var (
	ErrAspirinSynthesisStageNotFound = errors.New("этап синтеза аспирина не найден")
	ErrAspirinSynthesisDraftExists   = errors.New("у пользователя уже есть черновик этапа синтеза аспирина")
)

type AspirinProductionUser struct {
	ProductionUserID       int    `gorm:"column:production_user_id;primaryKey"`
	ProductionUsername     string `gorm:"column:production_username;size:100;not null;uniqueIndex"`
	ProductionUserPassword string `gorm:"column:production_user_password;size:255;not null"`
}

type AspirinSynthesisStageLike struct {
	AspirinSynthesisStageLikeID int `gorm:"column:aspirin_stage_like_id;primaryKey"`
	AspirinProductionUserID     int `gorm:"column:production_user_id;not null;uniqueIndex:idx_aspirin_stage_user_like"`
	AspirinSynthesisStageID     int `gorm:"column:aspirin_stage_id;not null;uniqueIndex:idx_aspirin_stage_user_like"`
}

func (AspirinProductionUser) TableName() string {
	return "aspirin_production_users"
}

type AspirinSynthesisStage struct {
	AspirinSynthesisStageID          int        `gorm:"column:aspirin_stage_id;primaryKey"`
	AspirinSynthesisStageName        string     `gorm:"column:aspirin_stage_name"`
	AspirinSynthesisStageDescription string     `gorm:"column:aspirin_stage_description"`
	AspirinSynthesisStageStatus      string     `gorm:"column:aspirin_stage_status"`
	AspirinSynthesisStageImageURL    *string    `gorm:"column:aspirin_stage_image_url"`
	AspirinSynthesisStageVideoURL    *string    `gorm:"column:aspirin_stage_video_url"`
	SynthesisDurationMinutes         *int       `gorm:"column:synthesis_duration_minutes"`
	PureAspirinYieldPercent          *float64   `gorm:"column:pure_aspirin_yield_percent"`
	AspirinSynthesisStageCreatedAt   time.Time  `gorm:"column:aspirin_stage_created_at"`
	AspirinSynthesisStageFormedAt    *time.Time `gorm:"column:aspirin_stage_formed_at"`
	AspirinProductionUserID          int        `gorm:"column:production_user_id"`

	AspirinSynthesisStageLikes []AspirinSynthesisStageLike `gorm:"foreignKey:AspirinSynthesisStageID;references:AspirinSynthesisStageID"`
	AspirinSynthesisLikeCount  int                         `gorm:"-"`
}

func (AspirinSynthesisStage) TableName() string {
	return "aspirin_stages"
}

func (AspirinSynthesisStageLike) TableName() string {
	return "aspirin_stage_likes"
}

type RepositorySettings struct {
	PostgresDSN     string
	MinioEndpoint   string
	MinioAccessKey  string
	MinioSecretKey  string
	MinioBucketName string
}

type Repository struct {
	db              *gorm.DB
	minioClient     *minio.Client
	minioBucketName string
	minioEndpoint   string
}

func New(settings RepositorySettings) (*Repository, error) {
	db, err := gorm.Open(postgres.Open(settings.PostgresDSN), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	minioClient, err := minio.New(settings.MinioEndpoint, &minio.Options{
		Creds: credentials.NewStaticV4(
			settings.MinioAccessKey,
			settings.MinioSecretKey,
			"",
		),
		Secure: false,
	})
	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	exists, err := minioClient.BucketExists(ctx, settings.MinioBucketName)
	if err != nil {
		return nil, err
	}

	if !exists {
		if err := minioClient.MakeBucket(
			ctx,
			settings.MinioBucketName,
			minio.MakeBucketOptions{},
		); err != nil {
			return nil, err
		}
	}

	return &Repository{
		db:              db,
		minioClient:     minioClient,
		minioBucketName: settings.MinioBucketName,
		minioEndpoint:   settings.MinioEndpoint,
	}, nil
}

func calculateAspirinSynthesisLikeCount(stages []AspirinSynthesisStage) {
	for i := range stages {
		stages[i].AspirinSynthesisLikeCount =
			len(stages[i].AspirinSynthesisStageLikes)
	}
}

func (r *Repository) GetPublishedAspirinSynthesisStages() ([]AspirinSynthesisStage, error) {
	var stages []AspirinSynthesisStage

	err := r.db.
		Preload("AspirinSynthesisStageLikes").
		Where(
			"aspirin_stage_status = ?",
			AspirinSynthesisStageStatusPublished,
		).
		Order("aspirin_stage_id ASC").
		Find(&stages).Error

	if err != nil {
		return nil, err
	}

	calculateAspirinSynthesisLikeCount(stages)

	return stages, nil
}

func (r *Repository) FilterPublishedAspirinSynthesisStagesByDuration(
	maxDuration int,
) ([]AspirinSynthesisStage, error) {

	var stages []AspirinSynthesisStage

	err := r.db.
		Preload("AspirinSynthesisStageLikes").
		Where(
			"aspirin_stage_status = ? AND synthesis_duration_minutes <= ?",
			AspirinSynthesisStageStatusPublished,
			maxDuration,
		).
		Order("aspirin_stage_id ASC").
		Find(&stages).Error

	if err != nil {
		return nil, err
	}

	calculateAspirinSynthesisLikeCount(stages)

	return stages, nil
}

func (r *Repository) GetAspirinSynthesisStageForFeed(
	stageID *int,
	next bool,
) (*AspirinSynthesisStage, error) {

	stages, err := r.GetPublishedAspirinSynthesisStages()
	if err != nil {
		return nil, err
	}

	if len(stages) == 0 {
		return nil, ErrAspirinSynthesisStageNotFound
	}

	if stageID == nil {
		return &stages[0], nil
	}

	currentIndex := -1

	for i := range stages {
		if stages[i].AspirinSynthesisStageID == *stageID {
			currentIndex = i
			break
		}
	}

	if currentIndex == -1 {
		return nil, ErrAspirinSynthesisStageNotFound
	}

	if !next {
		return &stages[currentIndex], nil
	}

	nextIndex := currentIndex + 1

	if nextIndex >= len(stages) {
		nextIndex = 0
	}

	return &stages[nextIndex], nil
}

func (r *Repository) GetAspirinSynthesisStageDraft(
	aspirinProductionUserID int,
) (*AspirinSynthesisStage, error) {

	var draft AspirinSynthesisStage

	err := r.db.
		Where(
			"production_user_id = ? AND aspirin_stage_status = ?",
			aspirinProductionUserID,
			AspirinSynthesisStageStatusDraft,
		).
		First(&draft).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &draft, nil
}

func (r *Repository) CreateAspirinSynthesisStageDraft(
	stageName string,
	aspirinProductionUserID int,
) error {

	existingDraft, err :=
		r.GetAspirinSynthesisStageDraft(
			aspirinProductionUserID,
		)

	if err != nil {
		return err
	}

	if existingDraft != nil {
		return ErrAspirinSynthesisDraftExists
	}

	stage := AspirinSynthesisStage{
		AspirinSynthesisStageName:      stageName,
		AspirinSynthesisStageStatus:    AspirinSynthesisStageStatusDraft,
		AspirinProductionUserID:        aspirinProductionUserID,
		AspirinSynthesisStageCreatedAt: time.Now(),
	}

	return r.db.Create(&stage).Error
}

func (r *Repository) CreateAspirinSynthesisStageDraftWithMedia(
	stageName string,
	aspirinProductionUserID int,
	imageFilename *string,
	videoFilename *string,
) (*AspirinSynthesisStage, error) {

	existingDraft, err :=
		r.GetAspirinSynthesisStageDraft(
			aspirinProductionUserID,
		)

	if err != nil {
		return nil, err
	}

	if existingDraft != nil {
		return nil, ErrAspirinSynthesisDraftExists
	}

	stage := AspirinSynthesisStage{
		AspirinSynthesisStageName:      stageName,
		AspirinSynthesisStageStatus:    AspirinSynthesisStageStatusDraft,
		AspirinProductionUserID:        aspirinProductionUserID,
		AspirinSynthesisStageCreatedAt: time.Now(),

		// Здесь в БД сохраняются только имена файлов.
		AspirinSynthesisStageImageURL: imageFilename,
		AspirinSynthesisStageVideoURL: videoFilename,
	}

	if err := r.db.Create(&stage).Error; err != nil {
		return nil, err
	}

	return &stage, nil
}

func (r *Repository) PublishAspirinSynthesisStageDraft(
	aspirinProductionUserID int,
	description string,
	durationMinutes int,
	pureAspirinYieldPercent float64,
) (*AspirinSynthesisStage, error) {

	var draft AspirinSynthesisStage

	err := r.db.
		Where(
			"production_user_id = ? AND aspirin_stage_status = ?",
			aspirinProductionUserID,
			AspirinSynthesisStageStatusDraft,
		).
		First(&draft).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrAspirinSynthesisStageNotFound
	}

	if err != nil {
		return nil, err
	}

	now := time.Now()

	err = r.db.
		Model(&draft).
		Updates(map[string]any{
			"aspirin_stage_description":
				description,

			"synthesis_duration_minutes":
				durationMinutes,

			"pure_aspirin_yield_percent":
				pureAspirinYieldPercent,

			"aspirin_stage_status":
				AspirinSynthesisStageStatusPublished,

			"aspirin_stage_formed_at":
				now,
		}).Error

	if err != nil {
		return nil, err
	}

	// Обновляем поля в объекте, который вернём handler'у.
	draft.AspirinSynthesisStageDescription =
		description

	draft.SynthesisDurationMinutes =
		&durationMinutes

	draft.PureAspirinYieldPercent =
		&pureAspirinYieldPercent

	draft.AspirinSynthesisStageStatus =
		AspirinSynthesisStageStatusPublished

	draft.AspirinSynthesisStageFormedAt =
		&now

	return &draft, nil
}

func (r *Repository) DeleteAspirinSynthesisStage(
	stageID int,
) error {

	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}

	result, err := sqlDB.ExecContext(
		context.Background(),
		`UPDATE aspirin_stages
         SET aspirin_stage_status = $1
         WHERE aspirin_stage_id = $2
           AND aspirin_stage_status <> $1`,
		AspirinSynthesisStageStatusDeleted,
		stageID,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrAspirinSynthesisStageNotFound
	}

	return nil
}

func (r *Repository) GetAspirinSynthesisStageByID(
	stageID int,
) (*AspirinSynthesisStage, error) {

	var stage AspirinSynthesisStage

	err := r.db.
		Preload("AspirinSynthesisStageLikes").
		Where(
			"aspirin_stage_id = ? AND aspirin_stage_status <> ?",
			stageID,
			AspirinSynthesisStageStatusDeleted,
		).
		First(&stage).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrAspirinSynthesisStageNotFound
	}

	if err != nil {
		return nil, err
	}

	stage.AspirinSynthesisLikeCount =
		len(stage.AspirinSynthesisStageLikes)

	return &stage, nil
}

func (r *Repository) DeleteAspirinSynthesisStageForUser(
	stageID int,
	aspirinProductionUserID int,
) error {

	result := r.db.
		Model(&AspirinSynthesisStage{}).
		Where(
			"aspirin_stage_id = ? AND production_user_id = ? AND aspirin_stage_status <> ?",
			stageID,
			aspirinProductionUserID,
			AspirinSynthesisStageStatusDeleted,
		).
		Update(
			"aspirin_stage_status",
			AspirinSynthesisStageStatusDeleted,
		)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrAspirinSynthesisStageNotFound
	}

	return nil
}