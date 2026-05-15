package project

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSaveProjectMetadataWritesEnvioJSONWithoutSecrets(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := saveProjectMetadata(root, projectMetadata{
		SchemaVersion:  1,
		ProjectID:      1,
		ProjectName:    "envio-cli",
		Owner:          "Team-NEEEE",
		RepoName:       "envio-cli",
		GithubRepoName: "Team-NEEEE/envio-cli",
	}); err != nil {
		t.Fatalf("saveProjectMetadata() error = %v", err)
	}

	raw, err := os.ReadFile(filepath.Join(root, projectMetadataFile))
	if err != nil {
		t.Fatalf("ReadFile(envio.json) error = %v", err)
	}
	content := string(raw)
	for _, forbidden := range []string{"masterKey", "wrappedMasterKey", "accessToken", "privateKey"} {
		if strings.Contains(content, forbidden) {
			t.Fatalf("envio.json leaked %s: %s", forbidden, content)
		}
	}
	var got projectMetadata
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if got.ProjectID != 1 || got.Owner != "Team-NEEEE" || got.GithubRepoName != "Team-NEEEE/envio-cli" {
		t.Fatalf("metadata = %#v", got)
	}
}

func TestSaveLocalLinkConfigWritesIgnoredEnvioConfig(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := saveLocalLinkConfig(root, LocalLinkConfig{
		LinkedProjectID: 1,
		RepositoryURL:   "https://github.com/Team-NEEEE/envio-cli.git",
		UserGithubID:    "octocat",
		DeviceID:        20,
	}); err != nil {
		t.Fatalf("saveLocalLinkConfig() error = %v", err)
	}

	raw, err := os.ReadFile(filepath.Join(root, localLinkConfigDir, localLinkConfigName))
	if err != nil {
		t.Fatalf("ReadFile(.envio/config) error = %v", err)
	}
	var got LocalLinkConfig
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if got.LinkedProjectID != 1 || got.UserGithubID != "octocat" || got.DeviceID != 20 {
		t.Fatalf("local config = %#v", got)
	}

	gitignore, err := os.ReadFile(filepath.Join(root, gitignoreFile))
	if err != nil {
		t.Fatalf("ReadFile(.gitignore) error = %v", err)
	}
	if !strings.Contains(string(gitignore), ".envio/") {
		t.Fatalf(".gitignore = %q, want .envio/", string(gitignore))
	}
}

func TestSaveLocalLinkConfigRejectsLegacyEnvioFileConflict(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, localLinkConfigDir), []byte("legacy"), 0600); err != nil {
		t.Fatalf("WriteFile(.envio) error = %v", err)
	}

	err := saveLocalLinkConfig(root, LocalLinkConfig{
		LinkedProjectID: 1,
		RepositoryURL:   "https://github.com/Team-NEEEE/envio-cli.git",
		UserGithubID:    "octocat",
		DeviceID:        20,
	})
	if err == nil {
		t.Fatal("saveLocalLinkConfig() error = nil, want conflict")
	}
	if !strings.Contains(err.Error(), ".envio/config") {
		t.Fatalf("error = %v, want mentioning .envio/config", err)
	}
}
