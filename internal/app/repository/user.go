package repository

import (
	"errors"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var ErrAspirinProductionUserExists = errors.New("пользователь с таким именем уже существует")
var ErrInvalidAspirinProductionCredentials = errors.New("неверное имя пользователя или пароль")

func (r *Repository) CreateAspirinProductionUser(username, password string) (*AspirinProductionUser, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := AspirinProductionUser{ProductionUsername: username, ProductionUserPassword: string(hash)}
	if err := r.db.Create(&user).Error; err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrAspirinProductionUserExists
		}
		return nil, err
	}
	return &user, nil
}

func (r *Repository) AuthenticateAspirinProductionUser(username, password string) (*AspirinProductionUser, error) {
	var user AspirinProductionUser
	err := r.db.Where("production_username = ?", username).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrInvalidAspirinProductionCredentials
	}
	if err != nil {
		return nil, err
	}
	if bcrypt.CompareHashAndPassword([]byte(user.ProductionUserPassword), []byte(password)) != nil {
		return nil, ErrInvalidAspirinProductionCredentials
	}
	return &user, nil
}
