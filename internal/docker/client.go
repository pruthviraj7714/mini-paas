package docker

import (
	"fmt"
	"log"

	"github.com/moby/moby/client"
)

func NewDockerClient() (*client.Client, error) {
	apiClient, err := client.New(client.FromEnv)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Docker client created successfully")

	return apiClient, nil
}
