package models

type DeploymentStatus string

const (
	QUEUED   DeploymentStatus = "queued"
	BUILDING DeploymentStatus = "building"
	STARTING DeploymentStatus = "starting"
	RUNNING  DeploymentStatus = "running"
	FAILED   DeploymentStatus = "failed"
	STOPPED  DeploymentStatus = "stopped"
)
