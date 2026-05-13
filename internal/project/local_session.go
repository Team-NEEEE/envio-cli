package project

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const localSessionFile = ".envio"

type SessionFile struct {
	Session Session `json:"session"`
}

type Session struct {
	MasterKey      MasterKeySession `json:"masterKey"`
	ProjectName    string           `json:"projectName"`
	GithubRepoName string           `json:"githubRepoName"`
	RepositoryURL  string           `json:"repositoryUrl"`
	ProjectID      int64            `json:"projectId"`
}

type MasterKeySession struct {
	Algorithm string `json:"algorithm"`
	Encoding  string `json:"encoding"`
	Value     string `json:"value"`
}

func saveProjectSession(repositoryRoot string, session Session) error {
	if strings.TrimSpace(repositoryRoot) == "" {
		return errors.New("repository root is required")
	}
	if err := validateProjectSession(session); err != nil {
		return err
	}

	path := localSessionPath(repositoryRoot)
	raw, err := json.MarshalIndent(SessionFile{Session: session}, "", "  ")
	if err != nil {
		return fmt.Errorf("encode project session: %w", err)
	}

	tempPath := path + ".tmp"
	if err := os.WriteFile(tempPath, raw, 0600); err != nil {
		return fmt.Errorf("write temporary project session: %w", err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("replace project session: %w", err)
	}
	return nil
}

func localSessionPath(repositoryRoot string) string {
	return filepath.Join(repositoryRoot, localSessionFile)
}

func validateProjectSession(session Session) error {
	if session.ProjectID <= 0 {
		return errors.New("project session projectId must be positive")
	}
	if strings.TrimSpace(session.RepositoryURL) == "" {
		return errors.New("project session repositoryUrl is required")
	}
	if strings.TrimSpace(session.MasterKey.Algorithm) == "" {
		return errors.New("project session masterKey.algorithm is required")
	}
	if strings.TrimSpace(session.MasterKey.Encoding) == "" {
		return errors.New("project session masterKey.encoding is required")
	}
	if strings.TrimSpace(session.MasterKey.Value) == "" {
		return errors.New("project session masterKey.value is required")
	}
	return nil
}
