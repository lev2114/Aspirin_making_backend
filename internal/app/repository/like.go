package repository

import (
	"errors"

	"gorm.io/gorm"
)

func (r *Repository) SetAspirinSynthesisStageLike(
	stageID int,
	aspirinProductionUserID int,
	likeValue int,
) (int, error) {

	if likeValue != 0 && likeValue != 1 {
		return 0, errors.New("значение лайка должно быть 0 или 1")
	}

	// Лайкать можно только опубликованную услугу.
	var stage AspirinSynthesisStage

	err := r.db.
		Where(
			"aspirin_stage_id = ? AND aspirin_stage_status = ?",
			stageID,
			AspirinSynthesisStageStatusPublished,
		).
		First(&stage).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, ErrAspirinSynthesisStageNotFound
	}

	if err != nil {
		return 0, err
	}

	if likeValue == 1 {

		stageLike := AspirinSynthesisStageLike{
			AspirinProductionUserID:
				aspirinProductionUserID,

			AspirinSynthesisStageID:
				stageID,
		}

		// Повторный like=1 не создаст второй лайк.
		if err :=
			r.db.
				Where(
					"production_user_id = ? AND aspirin_stage_id = ?",
					aspirinProductionUserID,
					stageID,
				).
				FirstOrCreate(&stageLike).Error; err != nil {

			return 0, err
		}

	} else {

		// like=0 — отмена лайка.
		if err :=
			r.db.
				Where(
					"production_user_id = ? AND aspirin_stage_id = ?",
					aspirinProductionUserID,
					stageID,
				).
				Delete(
					&AspirinSynthesisStageLike{},
				).Error; err != nil {

			return 0, err
		}
	}

	var likeCount int64

	if err :=
		r.db.
			Model(
				&AspirinSynthesisStageLike{},
			).
			Where(
				"aspirin_stage_id = ?",
				stageID,
			).
			Count(
				&likeCount,
			).Error; err != nil {

		return 0, err
	}

	return int(likeCount), nil
}