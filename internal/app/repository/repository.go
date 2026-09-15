package repository

import (
	"fmt"
)

type Repository struct{}

func NewRepository() (*Repository, error) {
	return &Repository{}, nil
}

type AspirinStage struct {
	ID          int
	Title       string
	Description string
	DurationMin int // Индивидуальное поле 1 (числовое)
	YieldPct    int // Индивидуальное поле 2 (числовое)
	Likes       int
	ImageURL    string
	VideoURL    string
}

func (r *Repository) GetAspirinStages() ([]AspirinStage, error) {
	stages := []AspirinStage{
		{
			ID:          1,
			Title:       "Подготовка материалов",
			Description: "Сбор реагентов и посуды.",
			DurationMin: 10,
			YieldPct:    100,
			Likes:       21,
			ImageURL:    "http://localhost:9000/aspirin-synthesis-media/materials_gathering.png",
			VideoURL:    "http://localhost:9000/aspirin-synthesis-media/materials_gathering.mp4",
		},
		{
			ID:          2,
			Title:       "Ацетилирование",
			Description: "Взаимодействие исходного вещества с ацетилирующим реагентом. После завершения реакции продукт направляется на охлаждение и кристаллизацию.",
			DurationMin: 20,
			YieldPct:    85,
			Likes:       12,
			ImageURL:    "http://localhost:9000/aspirin-synthesis-media/acetylation.png",
			VideoURL:    "http://localhost:9000/aspirin-synthesis-media/acetylation.mp4",
		},
		{
			ID:          3,
			Title:       "Фильтрация",
			Description: "Очистка от примесей.",
			DurationMin: 10,
			YieldPct:    80,
			Likes:       8,
			ImageURL:    "http://localhost:9000/aspirin-synthesis-media/filtering.png",
			VideoURL:    "http://localhost:9000/aspirin-synthesis-media/filtering.mp4",
		},
		{
			ID:          4,
			Title:       "Перекристаллизация",
			Description: "Получение чистого вещества.",
			DurationMin: 25,
			YieldPct:    75,
			Likes:       25,
			ImageURL:    "http://localhost:9000/aspirin-synthesis-media/recrystallisation.png",
			VideoURL:    "http://localhost:9000/aspirin-synthesis-media/recrystallisation.mp4",
		},
	}
	return stages, nil
}

func (r *Repository) GetStageByID(id int) (AspirinStage, error) {
	stages, _ := r.GetAspirinStages()
	for _, stage := range stages {
		if stage.ID == id {
			return stage, nil
		}
	}
	return AspirinStage{}, fmt.Errorf("этап синтеза не найден")
}
