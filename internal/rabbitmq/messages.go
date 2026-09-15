package rabbitmq

type DeploymentJob struct {
	DeploymentID int `json:"deployment_id"`
	ProjectID    int `json:"project_id"`
}
