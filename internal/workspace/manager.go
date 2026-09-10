package workspace

import "mini-paas/internal/git"

type WorkspaceManager struct {
	GitRunner *git.Runner
}

func NewWorkspaceManager(runner *git.Runner) *WorkspaceManager {
	return &WorkspaceManager{
		GitRunner: runner,
	}
}
