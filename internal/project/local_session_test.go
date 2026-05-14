package project

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSaveProjectSessionWritesEnvioFileWithMasterKey(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	session := Session{
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

	if err := saveProjectSession(root, session); err != nil {
		t.Fatalf("saveProjectSession() error = %v", err)
	}

	raw, err := os.ReadFile(filepath.Join(root, localSessionFile))
	if err != nil {
		t.Fatalf("ReadFile(.envio) error = %v", err)
	}
	var got SessionFile
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if got.Session.MasterKey.Value != "base64-master-key" ||
		got.Session.MasterKey.Encoding != projectMasterKeyEncoding ||
		got.Session.ProjectID != 1 {
		t.Fatalf("saved session = %#v", got.Session)
	}

	gitignore, err := os.ReadFile(filepath.Join(root, gitignoreFile))
	if err != nil {
		t.Fatalf("ReadFile(.gitignore) error = %v", err)
	}
	for _, pattern := range localSessionIgnorePatterns {
		if !strings.Contains(string(gitignore), pattern) {
			t.Fatalf(".gitignore = %q, want containing %q", string(gitignore), pattern)
		}
	}
}

func TestSaveProjectSessionAppendsMissingGitignorePatterns(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, gitignoreFile), []byte("dist\n.envio\n"), 0600); err != nil {
		t.Fatalf("WriteFile(.gitignore) error = %v", err)
	}

	if err := saveProjectSession(root, validProjectSession()); err != nil {
		t.Fatalf("saveProjectSession() error = %v", err)
	}

	gitignore, err := os.ReadFile(filepath.Join(root, gitignoreFile))
	if err != nil {
		t.Fatalf("ReadFile(.gitignore) error = %v", err)
	}
	content := string(gitignore)
	if strings.Count(content, ".envio\n") != 1 {
		t.Fatalf(".gitignore should not duplicate .envio entry: %q", content)
	}
	for _, pattern := range []string{".envio.tmp", ".envio/"} {
		if !strings.Contains(content, pattern) {
			t.Fatalf(".gitignore = %q, want containing %q", content, pattern)
		}
	}
}

func TestSaveProjectSessionRequiresMasterKeyFields(t *testing.T) {
	t.Parallel()

	err := saveProjectSession(t.TempDir(), Session{
		ProjectID:     1,
		RepositoryURL: "https://github.com/Team-NEEEE/envio-cli",
	})
	if err == nil {
		t.Fatal("saveProjectSession() error = nil, want masterKey validation error")
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
