package project

import (
	"encoding/json"
	"os"
	"path/filepath"
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
