package workspace

import (
	"context"
	"mini-paas/internal/git"
	"os"
	"path/filepath"
)

type WorkspaceManager struct {
	GitRunner *git.Runner
	baseDir   string
}

func NewWorkspaceManager(runner *git.Runner) *WorkspaceManager {
	return &WorkspaceManager{
		GitRunner: runner,
		baseDir:   "workspaces",
	}
}

func (wm *WorkspaceManager) GetWorkspaceDir(id string) (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, wm.baseDir, id), nil
}

func (wm *WorkspaceManager) PrepareWorkspace(ctx context.Context, repoURL string, deploymentID string) (string, error) {
	id := deploymentID
	workDir, err := wm.GetWorkspaceDir(id)
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(workDir, 0755); err != nil {
		return "", err
	}

	_, err = wm.GitRunner.Clone(ctx, repoURL, workDir)
	if err != nil {
		return "", err
	}

	return workDir, nil
}
