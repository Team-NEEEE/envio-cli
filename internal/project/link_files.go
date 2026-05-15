package project

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	projectMetadataFile = "envio.json"
	localLinkConfigDir  = ".envio"
	localLinkConfigName = "config"
)

type ProjectMetadata struct {
	ProjectName    string `json:"projectName"`
	Owner          string `json:"owner"`
	RepoName       string `json:"repoName"`
	GithubRepoName string `json:"githubRepoName"`
	SchemaVersion  int    `json:"schemaVersion"`
	ProjectID      int64  `json:"projectId"`
}

type LocalLinkConfig struct {
	RepositoryURL   string `json:"repositoryUrl"`
	UserGithubID    string `json:"userGithubId"`
	LinkedProjectID int64  `json:"linkedProjectId"`
	DeviceID        int64  `json:"deviceId"`
}

func saveProjectMetadata(repositoryRoot string, metadata ProjectMetadata) error {
	if strings.TrimSpace(repositoryRoot) == "" {
		return errors.New("repository root is required")
	}
	if err := validateProjectMetadata(metadata); err != nil {
		return err
	}
	return writeJSONAtomic(projectMetadataPath(repositoryRoot), metadata, 0644)
}

func saveLocalLinkConfig(repositoryRoot string, cfg LocalLinkConfig) error {
	if strings.TrimSpace(repositoryRoot) == "" {
		return errors.New("repository root is required")
	}
	if err := validateLocalLinkConfig(cfg); err != nil {
		return err
	}
	if err := ensureLocalLinkConfigPathAvailable(repositoryRoot); err != nil {
		return err
	}
	if err := ensureProjectSessionIgnored(repositoryRoot); err != nil {
		return err
	}
	return writeJSONAtomic(localLinkConfigPath(repositoryRoot), cfg, 0600)
}

func projectMetadataPath(repositoryRoot string) string {
	return filepath.Join(repositoryRoot, projectMetadataFile)
}

func localLinkConfigPath(repositoryRoot string) string {
	return filepath.Join(repositoryRoot, localLinkConfigDir, localLinkConfigName)
}

func ensureLocalLinkConfigPathAvailable(repositoryRoot string) error {
	path := filepath.Join(repositoryRoot, localLinkConfigDir)
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("inspect local link config path: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s exists as a file; .envio/config cannot be created without migrating the existing file", path)
	}
	return nil
}

func validateProjectMetadata(metadata ProjectMetadata) error {
	if metadata.SchemaVersion <= 0 {
		return errors.New("project metadata schemaVersion must be positive")
	}
	if metadata.ProjectID <= 0 {
		return errors.New("project metadata projectId must be positive")
	}
	if strings.TrimSpace(metadata.ProjectName) == "" {
		return errors.New("project metadata projectName is required")
	}
	if strings.TrimSpace(metadata.Owner) == "" {
		return errors.New("project metadata owner is required")
	}
	if strings.TrimSpace(metadata.RepoName) == "" {
		return errors.New("project metadata repoName is required")
	}
	if strings.TrimSpace(metadata.GithubRepoName) == "" {
		return errors.New("project metadata githubRepoName is required")
	}
	return nil
}

func validateLocalLinkConfig(cfg LocalLinkConfig) error {
	if cfg.LinkedProjectID <= 0 {
		return errors.New("local link config linkedProjectId must be positive")
	}
	if strings.TrimSpace(cfg.RepositoryURL) == "" {
		return errors.New("local link config repositoryUrl is required")
	}
	if strings.TrimSpace(cfg.UserGithubID) == "" {
		return errors.New("local link config userGithubId is required")
	}
	if cfg.DeviceID <= 0 {
		return errors.New("local link config deviceId must be positive")
	}
	return nil
}

func writeJSONAtomic(path string, obj any, perm os.FileMode) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	raw, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return fmt.Errorf("encode JSON: %w", err)
	}
	raw = append(raw, '\n')

	tempPath := path + ".tmp"
	if err := os.WriteFile(tempPath, raw, perm); err != nil {
		return fmt.Errorf("write temporary JSON file: %w", err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("replace JSON file: %w", err)
	}
	return nil
}
