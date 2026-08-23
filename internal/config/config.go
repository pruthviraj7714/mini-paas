package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	HTTPPort            string `mapstructure:"HTTP_PORT"`
	DatabaseURL         string `mapstructure:"DATABASE_URL"`
	RabbitMQURL         string `mapstructure:"RABBITMQ_URL"`
	DockerHost          string `mapstructure:"DOCKER_HOST"`
	DeploymentWorkspace string `mapstructure:"DEPLOYMENT_WORKSPACE"`
}

func getEnv(key, fallback string) string {
	val := os.Getenv(key)

	if val == "" {
		return fallback
	}

	return val
}

func LoadConfig() *Config {
	err := godotenv.Load()

	if err != nil {
		panic("error while loading .env")
	}

	return &Config{
		HTTPPort:            getEnv("PORT", "8080"),
		DatabaseURL:         getEnv("DATABASE_URL", "postgresql://postgres:password@localhost:5432"),
		RabbitMQURL:         getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672"),
		DockerHost:          getEnv("DOCKER_HOST", "unix:///var/run/docker.sock"),
		DeploymentWorkspace: getEnv("DEPLOYMENT_WORKSPACE", "/app"),
	}
}
