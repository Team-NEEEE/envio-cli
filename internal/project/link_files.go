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
	localLinkConfigDir  = ".envio"
	localLinkConfigName = "config"
)

type LocalLinkConfig struct {
	RepositoryURL   string           `json:"repositoryUrl"`
	UserGithubID    string           `json:"userGithubId"`
	ProjectName     string           `json:"projectName,omitempty"`
	GithubRepoName  string           `json:"githubRepoName,omitempty"`
	MasterKey       MasterKeySession `json:"masterKey"`
	LinkedProjectID int64            `json:"linkedProjectId"`
	DeviceID        int64            `json:"deviceId"`
	VersionID       int64            `json:"versionId,omitempty"`
}

func saveLocalLinkConfig(repositoryRoot string, cfg LocalLinkConfig) error {
	if strings.TrimSpace(repositoryRoot) == "" {
		return errors.New("repository root is required")
	}
	if err := validateLocalLinkConfig(cfg); err != nil {
		return err
	}
	if err := ensureLocalEnvioDir(repositoryRoot); err != nil {
		return err
	}
	if err := ensureProjectSessionIgnored(repositoryRoot); err != nil {
		return err
	}
	return writeJSONAtomic(localLinkConfigPath(repositoryRoot), cfg, 0600)
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
		if _, err := readProjectSessionFile(path); err == nil {
			return nil
		}
		return fmt.Errorf("%s exists as a file; .envio/config cannot be created without migrating the existing file", path)
	}
	return nil
}

func ensureLocalEnvioDir(repositoryRoot string) error {
	if err := ensureLocalLinkConfigPathAvailable(repositoryRoot); err != nil {
		return err
	}
	path := filepath.Join(repositoryRoot, localLinkConfigDir)
	if info, err := os.Stat(path); err == nil && !info.IsDir() {
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("migrate legacy .envio file: %w", err)
		}
	}
	if err := os.MkdirAll(path, 0700); err != nil {
		return fmt.Errorf("create .envio directory: %w", err)
	}
	if err := hideLocalEnvioDir(path); err != nil {
		return fmt.Errorf("hide .envio directory: %w", err)
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
	if strings.TrimSpace(cfg.MasterKey.Algorithm) == "" {
		return errors.New("local link config masterKey.algorithm is required")
	}
	if strings.TrimSpace(cfg.MasterKey.Encoding) == "" {
		return errors.New("local link config masterKey.encoding is required")
	}
	if strings.TrimSpace(cfg.MasterKey.Value) == "" {
		return errors.New("local link config masterKey.value is required")
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
