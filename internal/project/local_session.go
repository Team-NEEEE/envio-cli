package project

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const localSessionFile = "session"
const legacyLocalSessionFile = ".envio"
const gitignoreFile = ".gitignore"

var localSessionIgnorePatterns = []string{
	".envio",
	".envio/",
}

type SessionFile struct {
	Session Session `json:"session"`
}

type Session struct {
	MasterKey      MasterKeySession `json:"masterKey"`
	ProjectName    string           `json:"projectName"`
	GithubRepoName string           `json:"githubRepoName"`
	RepositoryURL  string           `json:"repositoryUrl"`
	ProjectID      int64            `json:"projectId"`
	VersionID      int64            `json:"versionId,omitempty"`
}

type MasterKeySession struct {
	Algorithm string `json:"algorithm"`
	Encoding  string `json:"encoding"`
	Value     string `json:"value"`
}

func localSessionPath(repositoryRoot string) string {
	return filepath.Join(repositoryRoot, localLinkConfigDir, localSessionFile)
}

func legacyLocalSessionPath(repositoryRoot string) string {
	return filepath.Join(repositoryRoot, legacyLocalSessionFile)
}

func ensureProjectSessionIgnored(repositoryRoot string) error {
	path := filepath.Join(repositoryRoot, gitignoreFile)
	raw, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read .gitignore: %w", err)
	}

	existing := map[string]bool{}
	if err == nil {
		for _, line := range strings.Split(string(raw), "\n") {
			existing[strings.TrimSpace(strings.TrimSuffix(line, "\r"))] = true
		}
	}

	missing := make([]string, 0, len(localSessionIgnorePatterns))
	for _, pattern := range localSessionIgnorePatterns {
		if !existing[pattern] {
			missing = append(missing, pattern)
		}
	}
	if len(missing) == 0 {
		return nil
	}

	var builder strings.Builder
	builder.Write(raw)
	if len(raw) > 0 && !strings.HasSuffix(string(raw), "\n") {
		builder.WriteString("\n")
	}
	for _, pattern := range missing {
		builder.WriteString(pattern)
		builder.WriteString("\n")
	}

	if err := os.WriteFile(path, []byte(builder.String()), 0600); err != nil {
		return fmt.Errorf("write .gitignore: %w", err)
	}
	return nil
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
