package docker

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/moby/go-archive"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

type Manager interface {
	Build(ctx context.Context, contextDir string, imageName string) error
	Run(ctx context.Context, imageName string, containerName string, command string) error
	Stop(ctx context.Context, containerID string) error
	Remove(ctx context.Context, containerID string) error
	Logs(ctx context.Context, containerID string) (string, error)
	Inspect(ctx context.Context, containerID string) (string, error)
}

type DockerManager struct {
	client *client.Client
}

func NewManager(cli *client.Client) *DockerManager {
	return &DockerManager{
		client: cli,
	}
}

func (m *DockerManager) Build(
	ctx context.Context,
	contextDir string,
	imageName string,
) error {
	buildContext, err := archive.TarWithOptions(
		contextDir,
		&archive.TarOptions{},
	)
	if err != nil {
		return fmt.Errorf("create build context: %w", err)
	}

	res, err := m.client.ImageBuild(
		ctx,
		buildContext,
		client.ImageBuildOptions{
			Dockerfile: "Dockerfile",
			Tags:       []string{imageName},
			Remove:     true,
		},
	)
	if err != nil {
		return fmt.Errorf("docker image build: %w", err)
	}
	defer res.Body.Close()

	_, err = io.Copy(os.Stdout, res.Body)
	if err != nil {
		return fmt.Errorf("read docker build output: %w", err)
	}

	return nil
}

func (m *DockerManager) Run(ctx context.Context, imageName string, containerName string, command string) error {

	ops := client.ContainerCreateOptions{
		Name: containerName,
		Config: &container.Config{
			Image: imageName,
		},
	}

	res, err := m.client.ContainerCreate(ctx, ops)
	if err != nil {
		return fmt.Errorf("create container: %w", err)
	}

	fmt.Println("Container created:", res.ID)

	return nil
}

func (m *DockerManager) Stop(ctx context.Context, containerID string) error {
	return nil
}

func (m *DockerManager) Remove(ctx context.Context, containerID string) error {
	return nil
}

func (m *DockerManager) Logs(ctx context.Context, containerID string) (string, error) {
	return "", nil
}

func (m *DockerManager) Inspect(ctx context.Context, containerID string) (string, error) {
	return "", nil
}
