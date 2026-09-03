package models

import (
	"time"

	"github.com/google/uuid"
)

type Project struct {
	ID            uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name          string    `json:"name" gorm:"not null;unique"`
	UserID        uuid.UUID
	RepositoryURL string    `json:"repository_url" gorm:"not null"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
