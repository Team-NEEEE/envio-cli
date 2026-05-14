package workspace

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

var ErrGitRepositoryRequired = errors.New("git repository with origin remote is required")

type GitRepository struct {
	Root      string
	OriginURL string
}

type GitInspector interface {
	Inspect(ctx context.Context, cwd string) (GitRepository, error)
}

type GitRunner interface {
	Run(ctx context.Context, cwd string, args ...string) (string, error)
}

type Inspector struct {
	runner GitRunner
}

func NewGitInspector(runner GitRunner) *Inspector {
	if runner == nil {
		runner = commandGitRunner{}
	}
	return &Inspector{runner: runner}
}

func (i *Inspector) Inspect(ctx context.Context, cwd string) (GitRepository, error) {
	if i == nil {
		return GitRepository{}, errors.New("git inspector is nil")
	}
	if strings.TrimSpace(cwd) == "" {
		cwd = "."
	}

	root, err := i.runner.Run(ctx, cwd, "rev-parse", "--show-toplevel")
	if err != nil {
		return GitRepository{}, fmt.Errorf("%w: %w", ErrGitRepositoryRequired, err)
	}
	root = strings.TrimSpace(root)
	if root == "" {
		return GitRepository{}, fmt.Errorf("%w: git root is empty", ErrGitRepositoryRequired)
	}

	originURL, err := i.runner.Run(ctx, root, "remote", "get-url", "origin")
	if err != nil {
		return GitRepository{}, fmt.Errorf("%w: %w", ErrGitRepositoryRequired, err)
	}
	originURL = strings.TrimSpace(originURL)
	if originURL == "" {
		return GitRepository{}, fmt.Errorf("%w: origin URL is empty", ErrGitRepositoryRequired)
	}

	return GitRepository{Root: root, OriginURL: originURL}, nil
}

type commandGitRunner struct{}

func (commandGitRunner) Run(ctx context.Context, cwd string, args ...string) (string, error) {
	allArgs := append([]string{"-C", cwd}, args...)
	cmd := exec.CommandContext(ctx, "git", allArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		detail := strings.TrimSpace(string(output))
		if detail == "" {
			return "", fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
		}
		return "", fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, detail)
	}
	return string(output), nil
}
