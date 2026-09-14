package git

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
)

type Result struct {
	Stdout  string
	Stderr  string
	Success bool
	Message string
}

type GitRunner interface {
	Clone(ctx context.Context, repoURL string, destination string) (*Result, error)
	Cleanup(deploymentID string) error
}

type Runner struct {
}

func (r *Runner) Clone(ctx context.Context, repoURL string, destination string) (*Result, error) {

	cmd := exec.CommandContext(ctx, "git", "clone", repoURL, destination)
	// 2. Create buffers to capture stdout and stderr
	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err := cmd.Run()

	stdoutStr := stdoutBuf.String()
	stderrStr := stderrBuf.String()

	fmt.Printf("--- STDOUT ---\n%s\n", stdoutStr)
	fmt.Printf("--- STDERR ---\n%s\n", stderrStr)

	exitCode := 0
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			// The program ran but exited with a non-zero status
			exitCode = exitErr.ExitCode()
		} else {
			// The command failed to start entirely (e.g., executable not found)
			return &Result{
				Stdout:  stdoutBuf.String(),
				Stderr:  stderrBuf.String(),
				Success: false,
				Message: err.Error(),
			}, err
		}
	}

	fmt.Printf("--- EXIT CODE ---\n%d\n", exitCode)

	if exitCode != 0 {
		return &Result{
			Stdout:  stdoutBuf.String(),
			Stderr:  stderrBuf.String(),
			Success: false,
			Message: stderrStr,
		}, fmt.Errorf("git clone failed with exit code %d", exitCode)
	}

	return &Result{
		Stdout:  stdoutBuf.String(),
		Stderr:  stderrBuf.String(),
		Success: true,
		Message: "Clone successful",
	}, nil
}

func (r *Runner) Cleanup(deploymentID string) error {
	return nil
}
