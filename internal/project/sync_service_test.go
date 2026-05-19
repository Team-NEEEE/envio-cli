package project

import (
	"context"
	"encoding/base64"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/Team-NEEEE/envio-cli/internal/api"
	projectapi "github.com/Team-NEEEE/envio-cli/internal/api/project"
	"github.com/Team-NEEEE/envio-cli/internal/command"
	envcrypto "github.com/Team-NEEEE/envio-cli/internal/crypto"
	"github.com/Team-NEEEE/envio-cli/internal/workspace"
)

type fakeSyncAPI struct {
	pullErr      error
	pushErr      error
	historyErr   error
	pullResp     *projectapi.ProjectPullResponse
	pushResp     *projectapi.ProjectPushResponse
	historyResp  *projectapi.ProjectHistoryResponse
	auth         string
	pushReq      projectapi.ProjectPushRequest
	projectID    int64
	pullCalls    int
	pushCalls    int
	historyCalls int
}

func (f *fakeSyncAPI) PullLatest(
	_ context.Context,
	projectID int64,
	githubUserID string,
	deviceID int64,
	authorization string,
) (*projectapi.ProjectPullResponse, error) {
	f.pullCalls++
	f.projectID = projectID
	f.auth = authorization
	if githubUserID != "octocat" {
		return nil, errors.New("unexpected github user")
	}
	if deviceID != 20 {
		return nil, errors.New("unexpected device")
	}
	return f.pullResp, f.pullErr
}

func (f *fakeSyncAPI) Push(
	_ context.Context,
	projectID int64,
	req projectapi.ProjectPushRequest,
	authorization string,
) (*projectapi.ProjectPushResponse, error) {
	f.pushCalls++
	f.projectID = projectID
	f.pushReq = req
	f.auth = authorization
	return f.pushResp, f.pushErr
}

func (f *fakeSyncAPI) History(
	_ context.Context,
	projectID int64,
	authorization string,
) (*projectapi.ProjectHistoryResponse, error) {
	f.historyCalls++
	f.projectID = projectID
	f.auth = authorization
	return f.historyResp, f.historyErr
}

func TestSyncPushEncryptsEnvironmentAndSavesVersion(t *testing.T) {
	t.Parallel()

	client := &fakeSyncAPI{pushResp: &projectapi.ProjectPushResponse{
		ProjectID:       1,
		HistoryID:       11,
		VersionID:       3,
		ParentVersionID: 2,
	}}
	service, saved := newTestSyncService(client)
	service.readFile = func(string) ([]byte, error) {
		return []byte("API_KEY=secret\nDATABASE_URL=postgres://localhost/db\n"), nil
	}

	got, appErr := service.Push(context.Background(), "C:/repo", "", command.NoopReporter{})
	if appErr != nil {
		t.Fatalf("Push() appErr = %v", appErr)
	}
	if got.VersionID != 3 || got.ParentVersionID != 2 || got.VariableCount != 2 {
		t.Fatalf("Push() = %#v", got)
	}
	if client.pushCalls != 1 || client.projectID != 1 || client.auth != "" {
		t.Fatalf("client state = %#v", client)
	}
	if client.pushReq.GithubUserID != "octocat" || client.pushReq.ParentVersionID != 2 {
		t.Fatalf("push request = %#v", client.pushReq)
	}
	decrypted, err := envcrypto.DecryptEnvironment(client.pushReq.EncryptedEnvironment, testSyncMasterKey())
	if err != nil {
		t.Fatalf("DecryptEnvironment() error = %v", err)
	}
	if string(decrypted) != "API_KEY=secret\nDATABASE_URL=postgres://localhost/db\n" {
		t.Fatalf("decrypted = %q", string(decrypted))
	}
	if saved.versionID != 3 {
		t.Fatalf("saved version = %d, want 3", saved.versionID)
	}
}

func TestSyncPullDecryptsEnvironmentWritesFileAndSavesVersion(t *testing.T) {
	t.Parallel()

	encrypted, err := envcrypto.EncryptEnvironment([]byte("API_KEY=secret\n"), testSyncMasterKey())
	if err != nil {
		t.Fatalf("EncryptEnvironment() error = %v", err)
	}
	client := &fakeSyncAPI{pullResp: &projectapi.ProjectPullResponse{
		EncryptedEnvironment: encrypted,
		ProjectID:            1,
		HistoryID:            11,
		VersionID:            3,
	}}
	service, saved := newTestSyncService(client)

	got, appErr := service.Pull(context.Background(), "C:/repo", ".env.local", command.NoopReporter{})
	if appErr != nil {
		t.Fatalf("Pull() appErr = %v", appErr)
	}
	if got.VersionID != 3 || got.VariableCount != 1 {
		t.Fatalf("Pull() = %#v", got)
	}
	if saved.path != filepath.Join("C:/repo", ".env.local") ||
		string(saved.raw) != "API_KEY=secret\n" ||
		saved.perm != 0600 ||
		saved.versionID != 3 {
		t.Fatalf("saved = %#v", saved)
	}
}

func TestSyncPullUsesWrappedMasterKeyWhenReturned(t *testing.T) {
	t.Parallel()

	localMasterKey := []byte("12345678901234567890123456789012")
	remoteMasterKey := []byte("abcdefghijklmnopqrstuvwxzy123456")
	encrypted, err := envcrypto.EncryptEnvironment([]byte("API_KEY=remote\n"), remoteMasterKey)
	if err != nil {
		t.Fatalf("EncryptEnvironment() error = %v", err)
	}
	client := &fakeSyncAPI{pullResp: &projectapi.ProjectPullResponse{
		EncryptedEnvironment: encrypted,
		WrappedMasterKey:     "wrapped-key",
		ProjectID:            1,
		HistoryID:            11,
		VersionID:            3,
	}}
	service, saved := newTestSyncService(client)
	service.loadLocalContext = func(context.Context, string) (localProjectContext, *command.AppError) {
		return localProjectContext{
			repositoryRoot: "C:/repo",
			githubUserID:   "octocat",
			masterKey:      localMasterKey,
			source:         localProjectSourceSession,
			projectID:      1,
			deviceID:       20,
			versionID:      2,
		}, nil
	}
	service.loadPrivateKey = func(deviceID int64) (string, error) {
		if deviceID != 20 {
			t.Fatalf("deviceID = %d, want 20", deviceID)
		}
		return "private-key", nil
	}
	service.unwrapMasterKey = func(privateKey string, wrappedMasterKey string) ([]byte, error) {
		if privateKey != "private-key" || wrappedMasterKey != "wrapped-key" {
			t.Fatalf("unwrap args = %q, %q", privateKey, wrappedMasterKey)
		}
		return remoteMasterKey, nil
	}

	got, appErr := service.Pull(context.Background(), "C:/repo", "", command.NoopReporter{})
	if appErr != nil {
		t.Fatalf("Pull() appErr = %v", appErr)
	}
	if got.VariableCount != 1 || string(saved.raw) != "API_KEY=remote\n" {
		t.Fatalf("Pull() = %#v, saved = %#v", got, saved)
	}
}

func TestSyncHistoryListsMetadataWithoutDecrypting(t *testing.T) {
	t.Parallel()

	client := &fakeSyncAPI{historyResp: &projectapi.ProjectHistoryResponse{
		Histories: []projectapi.ProjectHistoryEntry{
			{
				HistoryID:            11,
				ProjectID:            1,
				VersionID:            8,
				BaseVersionID:        7,
				GithubID:             "octocat",
				CreatedAt:            "2026-04-20 11:00:00",
				EncryptedEnvironment: map[string]any{"ciphertext": "ciphertext"},
			},
			{
				HistoryID: 10,
				ProjectID: 1,
				VersionID: 7,
				GithubID:  "dev1",
				CreatedAt: "2026-04-19 10:00:00",
			},
		},
	}}
	service, _ := newTestSyncService(client)
	service.decryptEnvironment = func(map[string]any, []byte) ([]byte, error) {
		t.Fatal("ListHistory must not decrypt environments")
		return nil, nil
	}

	got, appErr := service.ListHistory(context.Background(), "C:/repo", command.NoopReporter{})
	if appErr != nil {
		t.Fatalf("ListHistory() appErr = %v", appErr)
	}
	if client.historyCalls != 1 || got.ProjectID != 1 || len(got.Histories) != 2 {
		t.Fatalf("history result = %#v, client = %#v", got, client)
	}
	if !got.Histories[0].Latest || got.Histories[0].VersionID != 8 || got.Histories[0].BaseVersionID != 7 {
		t.Fatalf("history latest normalization = %#v", got.Histories[0])
	}
}

func TestSyncHistoryDecryptsSelectedVersion(t *testing.T) {
	t.Parallel()

	encrypted, err := envcrypto.EncryptEnvironment([]byte("API_KEY=secret\n"), testSyncMasterKey())
	if err != nil {
		t.Fatalf("EncryptEnvironment() error = %v", err)
	}
	client := &fakeSyncAPI{historyResp: &projectapi.ProjectHistoryResponse{
		Histories: []projectapi.ProjectHistoryEntry{
			{
				HistoryID:            11,
				ProjectID:            1,
				VersionID:            8,
				BaseVersionID:        7,
				GithubID:             "octocat",
				CreatedAt:            "2026-04-20 11:00:00",
				EncryptedEnvironment: encrypted,
			},
		},
	}}
	service, _ := newTestSyncService(client)

	got, appErr := service.DecryptHistoryVersion(context.Background(), "C:/repo", "v8", command.NoopReporter{})
	if appErr != nil {
		t.Fatalf("DecryptHistoryVersion() appErr = %v", appErr)
	}
	if got.VersionID != 8 || got.BaseVersionID != 7 || got.VariableCount != 1 {
		t.Fatalf("DecryptHistoryVersion() = %#v", got)
	}
	if got.Environment != "API_KEY=secret\n" {
		t.Fatalf("environment = %q", got.Environment)
	}
}

func TestSyncPushMapsVersionConflict(t *testing.T) {
	t.Parallel()

	client := &fakeSyncAPI{pushErr: &api.ErrorResponse{
		Status:  api.ErrorStatus("409 CONFLICT"),
		Code:    backendCodeVersionConflict,
		Message: "conflict",
	}}
	service, _ := newTestSyncService(client)
	service.readFile = func(string) ([]byte, error) {
		return []byte("API_KEY=secret\n"), nil
	}

	_, appErr := service.Push(context.Background(), "C:/repo", "", command.NoopReporter{})
	if appErr == nil || appErr.Code != ErrorVersionConflict {
		t.Fatalf("Push() appErr = %#v, want %s", appErr, ErrorVersionConflict)
	}
}

func TestSyncLoadContextSupportsLegacyProjectSession(t *testing.T) {
	t.Parallel()

	local, appErr := localContextFromSession(
		workspace.GitRepository{
			Root:      "C:/repo",
			OriginURL: "git@github.com:Team-NEEEE/envio-cli.git",
		},
		validLinkSession(),
		Session{
			ProjectID:      1,
			ProjectName:    "envio-cli",
			GithubRepoName: "Team-NEEEE/envio-cli",
			RepositoryURL:  "https://github.com/Team-NEEEE/envio-cli",
			VersionID:      4,
			MasterKey: MasterKeySession{
				Algorithm: projectMasterKeyAlgorithm,
				Encoding:  projectMasterKeyEncoding,
				Value:     base64ForTest(testSyncMasterKey()),
			},
		},
	)
	if appErr != nil {
		t.Fatalf("localContextFromSession() appErr = %v", appErr)
	}
	if local.projectID != 1 || local.versionID != 4 || local.githubUserID != "octocat" {
		t.Fatalf("local context = %#v", local)
	}
}

type savedSyncState struct {
	path      string
	raw       []byte
	versionID int64
	perm      os.FileMode
}

func newTestSyncService(client *fakeSyncAPI) (*SyncService, *savedSyncState) {
	saved := &savedSyncState{}
	service := &SyncService{
		client: client,
		loadLocalContext: func(context.Context, string) (localProjectContext, *command.AppError) {
			return localProjectContext{
				repositoryRoot: "C:/repo",
				githubUserID:   "octocat",
				masterKey:      testSyncMasterKey(),
				source:         localProjectSourceSession,
				projectID:      1,
				deviceID:       20,
				versionID:      2,
			}, nil
		},
		saveLocalVersion: func(_ localProjectContext, versionID int64) error {
			saved.versionID = versionID
			return nil
		},
		writeFile: func(path string, raw []byte, perm os.FileMode) error {
			saved.path = path
			saved.raw = append([]byte(nil), raw...)
			saved.perm = perm
			return nil
		},
	}
	return service, saved
}

func testSyncMasterKey() []byte {
	return []byte("12345678901234567890123456789012")
}

func base64ForTest(raw []byte) string {
	return base64.StdEncoding.EncodeToString(raw)
}
