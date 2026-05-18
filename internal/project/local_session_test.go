package project

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnsureProjectSessionIgnoredAppendsMissingGitignorePatterns(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, gitignoreFile), []byte("dist\n.envio\n"), 0600); err != nil {
		t.Fatalf("WriteFile(.gitignore) error = %v", err)
	}

	if err := ensureProjectSessionIgnored(root); err != nil {
		t.Fatalf("ensureProjectSessionIgnored() error = %v", err)
	}

	gitignore, err := os.ReadFile(filepath.Join(root, gitignoreFile))
	if err != nil {
		t.Fatalf("ReadFile(.gitignore) error = %v", err)
	}
	content := string(gitignore)
	if strings.Count(content, ".envio\n") != 1 {
		t.Fatalf(".gitignore should not duplicate .envio entry: %q", content)
	}
	if !strings.Contains(content, ".envio/") {
		t.Fatalf(".gitignore = %q, want containing .envio/", content)
	}
}

func TestValidateProjectSessionRequiresMasterKeyFields(t *testing.T) {
	t.Parallel()

	err := validateProjectSession(Session{
		ProjectID:     1,
		RepositoryURL: "https://github.com/Team-NEEEE/envio-cli",
	})
	if err == nil {
		t.Fatal("validateProjectSession() error = nil, want masterKey validation error")
	}
}

func validProjectSession() Session {
	return Session{
		ProjectID:      1,
		ProjectName:    "envio-cli",
		GithubRepoName: "Team-NEEEE/envio-cli",
		RepositoryURL:  "https://github.com/Team-NEEEE/envio-cli",
		MasterKey: MasterKeySession{
			Algorithm: projectMasterKeyAlgorithm,
			Encoding:  projectMasterKeyEncoding,
			Value:     "base64-master-key",
		},
	}
}
