package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Deployment struct {
	ID        uuid.UUID        `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ProjectID uuid.UUID        `json:"project_id"`
	Status    DeploymentStatus `json:"status"`
	CreatedAt time.Time        `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time        `json:"updated_at" gorm:"autoUpdateTime"`
}

func (d *Deployment) TransitionTo(
	status DeploymentStatus,
) error {

	switch d.Status {
	case QUEUED:
		if status != BUILDING {
			return fmt.Errorf("invalid transition from %s to %s", d.Status, status)
		}
	case BUILDING:
		if status != RUNNING && status != FAILED {
			return fmt.Errorf("invalid transition from %s to %s", d.Status, status)
		}
	case RUNNING:
		if status != STOPPED && status != FAILED {
			return fmt.Errorf("invalid transition from %s to %s", d.Status, status)
		}
	case STOPPED:
		if status != RUNNING && status != FAILED {
			return fmt.Errorf("invalid transition from %s to %s", d.Status, status)
		}
	case FAILED:
		return fmt.Errorf("invalid transition from %s to %s", d.Status, status)
	}

	d.Status = status
	d.UpdatedAt = time.Now()
	return nil
}
