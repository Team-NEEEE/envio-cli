package project

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSaveLocalLinkConfigWritesIgnoredEnvioConfig(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := saveLocalLinkConfig(root, validLocalLinkConfigForTest()); err != nil {
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
	if got.LinkedProjectID != 1 ||
		got.UserGithubID != "octocat" ||
		got.DeviceID != 20 ||
		got.MasterKey.Value == "" {
		t.Fatalf("local config = %#v", got)
	}

	gitignore, err := os.ReadFile(filepath.Join(root, gitignoreFile))
	if err != nil {
		t.Fatalf("ReadFile(.gitignore) error = %v", err)
	}
	if !strings.Contains(string(gitignore), ".envio/") {
		t.Fatalf(".gitignore = %q, want .envio/", string(gitignore))
	}
	if _, err := os.Stat(filepath.Join(root, "envio.json")); !os.IsNotExist(err) {
		t.Fatalf("envio.json should not be created, err = %v", err)
	}
}

func TestSaveLocalLinkConfigRejectsLegacyEnvioFileConflict(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, localLinkConfigDir), []byte("legacy"), 0600); err != nil {
		t.Fatalf("WriteFile(.envio) error = %v", err)
	}

	err := saveLocalLinkConfig(root, validLocalLinkConfigForTest())
	if err == nil {
		t.Fatal("saveLocalLinkConfig() error = nil, want conflict")
	}
	if !strings.Contains(err.Error(), ".envio/config") {
		t.Fatalf("error = %v, want mentioning .envio/config", err)
	}
}

func validLocalLinkConfigForTest() LocalLinkConfig {
	return LocalLinkConfig{
		LinkedProjectID: 1,
		RepositoryURL:   "https://github.com/Team-NEEEE/envio-cli.git",
		UserGithubID:    "octocat",
		DeviceID:        20,
		ProjectName:     "envio-cli",
		GithubRepoName:  "Team-NEEEE/envio-cli",
		MasterKey: MasterKeySession{
			Algorithm: projectMasterKeyAlgorithm,
			Encoding:  projectMasterKeyEncoding,
			Value:     base64.StdEncoding.EncodeToString(testProjectMasterKey()),
		},
	}
}
