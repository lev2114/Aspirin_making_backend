package repository

import (
	"errors"
)

var ErrAspirinProductionUserExists =
	errors.New("пользователь с таким именем уже существует")

func (r *Repository) CreateAspirinProductionUser(
	username string,
	password string,
) (*AspirinProductionUser, error) {

	var count int64

	if err :=
		r.db.
			Model(
				&AspirinProductionUser{},
			).
			Where(
				"production_username = ?",
				username,
			).
			Count(
				&count,
			).Error; err != nil {

		return nil, err
	}

	if count > 0 {
		return nil, ErrAspirinProductionUserExists
	}

	user := AspirinProductionUser{
		ProductionUsername:
			username,

		ProductionUserPassword:
			password,
	}

	if err :=
		r.db.
			Create(
				&user,
			).Error; err != nil {

		return nil, err
	}

	return &user, nil
}