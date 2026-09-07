package database

import (
	"mini-paas/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConnectPostgres(connStr string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{})

	db.AutoMigrate(&models.User{}, &models.Project{}, &models.Deployment{})

	return db, err
}
