package models

import (
	"time"

	"github.com/google/uuid"
)

type Project struct {
	ID            uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name          string    `json:"name" gorm:"not null"`
	UserID        uuid.UUID `json:"user_id" gorm:"not null"`
	RepositoryURL string    `json:"repository_url" gorm:"not null;uniqueIndex"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type ProjectAddRequest struct {
	Name          string `json:"name" gorm:"not null;unique"`
	RepositoryURL string `json:"repository_url" gorm:"not null;unique"`
}
