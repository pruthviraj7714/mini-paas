package events

import "github.com/google/uuid"

const DeploymentCreatedRoutingKey = "deployment.created"

type DeploymentJobPayload struct {
	DeploymentID uuid.UUID `json:"deployment_id"`
	ProjectID    uuid.UUID `json:"project_id"`
	RepoURL      string    `json:"repo_url"`
	CommitSHA    string    `json:"commit_sha"`
}
