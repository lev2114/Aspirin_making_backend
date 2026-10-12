// Command migrate prepares the schema and upgrades Lab 3 plaintext passwords.
package main

import (
	"log"

	"aspirin_back/internal/app/dsn"
	"aspirin_back/internal/app/repository"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	_ = godotenv.Load()
	db, err := gorm.Open(postgres.Open(dsn.FromEnv()), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&repository.AspirinProductionUser{}, &repository.AspirinSynthesisStage{}, &repository.AspirinSynthesisStageLike{}); err != nil {
			return err
		}
		if err := tx.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_aspirin_stage_user_draft ON aspirin_stages (production_user_id) WHERE aspirin_stage_status = 'draft'`).Error; err != nil {
			return err
		}
		var users []repository.AspirinProductionUser
		if err := tx.Find(&users).Error; err != nil {
			return err
		}
		for _, user := range users {
			if _, err := bcrypt.Cost([]byte(user.ProductionUserPassword)); err == nil {
				continue
			}
			hash, err := bcrypt.GenerateFromPassword([]byte(user.ProductionUserPassword), bcrypt.DefaultCost)
			if err != nil {
				return err
			}
			if err := tx.Model(&user).Update("production_user_password", string(hash)).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		log.Fatal(err)
	}
	log.Print("Таблицы производства аспирина и хеши паролей готовы")
}
