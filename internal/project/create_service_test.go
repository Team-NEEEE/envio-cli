package project

import (
	"context"
	"encoding/base64"
	"errors"
	"strings"
	"testing"

	projectapi "github.com/Team-NEEEE/envio-cli/internal/api/project"
	"github.com/Team-NEEEE/envio-cli/internal/command"
	"github.com/Team-NEEEE/envio-cli/internal/workspace"
)

type fakeCreateAPI struct {
	createReq   projectapi.CreateProjectRequest
	saveReq     projectapi.SaveWrappedKeysRequest
	createResp  *projectapi.CreateProjectResponse
	saveResp    *projectapi.SaveWrappedKeysResponse
	createErr   error
	saveErr     error
	saveAuth    string
	saveProject int64
	createCalls int
	saveCalls   int
}

func (f *fakeCreateAPI) CreateProject(
	_ context.Context,
	req projectapi.CreateProjectRequest,
) (*projectapi.CreateProjectResponse, error) {
	f.createCalls++
	f.createReq = req
	return f.createResp, f.createErr
}

func (f *fakeCreateAPI) SaveWrappedKeys(
	_ context.Context,
	projectID int64,
	req projectapi.SaveWrappedKeysRequest,
	authorization string,
) (*projectapi.SaveWrappedKeysResponse, error) {
	f.saveCalls++
	f.saveProject = projectID
	f.saveReq = req
	f.saveAuth = authorization
	return f.saveResp, f.saveErr
}

type fakeProjectGit struct {
	repository workspace.GitRepository
	err        error
	calls      int
}

func (f *fakeProjectGit) Inspect(context.Context, string) (workspace.GitRepository, error) {
	f.calls++
	return f.repository, f.err
}

func TestCreateDoesNotRequireGlobalSessionBeforeGitOrAPI(t *testing.T) {
	t.Parallel()

	client := &fakeCreateAPI{createResp: validCreateResponse(), saveResp: validSaveResponse()}
	service, saved := newTestCreateService(client)

	_, appErr := service.Create(context.Background(), "C:/repo", "https://github.com/Team-NEEEE/envio-cli", command.NoopReporter{})
	if appErr != nil {
		t.Fatalf("Create() appErr = %v", appErr)
	}
	if client.createCalls != 1 || client.saveCalls != 1 {
		t.Fatalf("api calls = create %d save %d, want both called", client.createCalls, client.saveCalls)
	}
	if saved.session.MasterKey.Value == "" {
		t.Fatalf("saved session should include masterKey: %#v", saved.session)
	}
}

func TestCreateStopsBeforeAPIWhenRepositoryContextMismatches(t *testing.T) {
	t.Parallel()

	client := &fakeCreateAPI{}
	service, _ := newTestCreateService(client)

	_, appErr := service.Create(
		context.Background(),
		"C:/repo",
		"https://github.com/Team-NEEEE/envio-server",
		command.NoopReporter{},
	)
	if appErr == nil || appErr.Code != ErrorRepositoryContextMismatch {
		t.Fatalf("Create() appErr = %#v, want %s", appErr, ErrorRepositoryContextMismatch)
	}
	if client.createCalls != 0 || client.saveCalls != 0 {
		t.Fatalf("api calls = create %d save %d, want none", client.createCalls, client.saveCalls)
	}
	if appErr.Hint == "" ||
		!containsAll(appErr.Hint, "Team-NEEEE/envio-cli", "Team-NEEEE/envio-server") {
		t.Fatalf("mismatch hint should include normalized repos, got %q", appErr.Hint)
	}
}

func TestCreateCreatesProjectWrapsMembersSavesWrappedKeysAndLocalSession(t *testing.T) {
	t.Parallel()

	client := &fakeCreateAPI{
		createResp: validCreateResponse(),
		saveResp:   validSaveResponse(),
	}
	service, saved := newTestCreateService(client)

	got, appErr := service.Create(
		context.Background(),
		"C:/repo",
		"https://github.com/Team-NEEEE/envio-cli.git",
		command.NoopReporter{},
	)
	if appErr != nil {
		t.Fatalf("Create() appErr = %v", appErr)
	}
	if got.ProjectID != 1 || got.WrappedKeyTargetCount != 2 || got.UpdatedCount != 2 {
		t.Fatalf("Create() = %#v", got)
	}
	if client.createCalls != 1 || client.saveCalls != 1 {
		t.Fatalf("api calls = create %d save %d", client.createCalls, client.saveCalls)
	}
	if client.createReq.RepositoryURL != "https://github.com/Team-NEEEE/envio-cli.git" {
		t.Fatalf("create request = %#v", client.createReq)
	}
	if client.saveProject != 1 || len(client.saveReq.WrappedKeys) != 2 {
		t.Fatalf("save request project=%d body=%#v", client.saveProject, client.saveReq)
	}
	if client.saveAuth != "" {
		t.Fatalf("save authorization = %q, want empty because create does not use global session", client.saveAuth)
	}
	if client.saveReq.WrappedKeys[0].EncryptedKey != "wrapped-public-key-1" ||
		client.saveReq.WrappedKeys[1].EncryptedKey != "wrapped-public-key-2" {
		t.Fatalf("save request = %#v", client.saveReq)
	}
	if saved.root != "C:/repo" {
		t.Fatalf("saved root = %q", saved.root)
	}
	if saved.session.ProjectID != 1 ||
		saved.session.MasterKey.Algorithm != projectMasterKeyAlgorithm ||
		saved.session.MasterKey.Encoding != projectMasterKeyEncoding {
		t.Fatalf("saved session = %#v", saved.session)
	}
	wantMasterKey := base64.StdEncoding.EncodeToString([]byte("12345678901234567890123456789012"))
	if saved.session.MasterKey.Value != wantMasterKey {
		t.Fatalf("saved master key = %q, want %q", saved.session.MasterKey.Value, wantMasterKey)
	}
}

func TestCreateRejectsInvalidCreateResponseBeforeWrapping(t *testing.T) {
	t.Parallel()

	client := &fakeCreateAPI{
		createResp: &projectapi.CreateProjectResponse{ProjectID: 1},
	}
	service, saved := newTestCreateService(client)

	_, appErr := service.Create(
		context.Background(),
		"C:/repo",
		"https://github.com/Team-NEEEE/envio-cli",
		command.NoopReporter{},
	)
	if appErr == nil || appErr.Code != ErrorCreateResponseInvalid {
		t.Fatalf("Create() appErr = %#v, want %s", appErr, ErrorCreateResponseInvalid)
	}
	if client.saveCalls != 0 {
		t.Fatalf("save calls = %d, want 0", client.saveCalls)
	}
	if saved.called {
		t.Fatal("project session should not be saved for invalid create response")
	}
}

func TestCreateStopsBeforeSaveWhenWrappingFails(t *testing.T) {
	t.Parallel()

	client := &fakeCreateAPI{
		createResp: &projectapi.CreateProjectResponse{
			ProjectID: 1,
			Members: []projectapi.ProjectMember{
				{UserID: 10, UserDeviceID: 20, GithubID: "octocat", PublicKey: "bad-key"},
			},
		},
	}
	service, saved := newTestCreateService(client)
	service.wrapProjectMasterKey = func(string, []byte) (string, error) {
		return "", errors.New("invalid public key")
	}

	_, appErr := service.Create(
		context.Background(),
		"C:/repo",
		"https://github.com/Team-NEEEE/envio-cli",
		command.NoopReporter{},
	)
	if appErr == nil || appErr.Code != ErrorWrapProjectKeyFailed {
		t.Fatalf("Create() appErr = %#v, want %s", appErr, ErrorWrapProjectKeyFailed)
	}
	if client.saveCalls != 0 {
		t.Fatalf("save calls = %d, want 0", client.saveCalls)
	}
	if saved.called {
		t.Fatal("project session should not be saved when wrapping fails")
	}
}

func TestCreateReturnsSaveWrappedKeysFailure(t *testing.T) {
	t.Parallel()

	client := &fakeCreateAPI{
		createResp: validCreateResponse(),
		saveErr:    errors.New("save failed"),
	}
	service, saved := newTestCreateService(client)

	_, appErr := service.Create(
		context.Background(),
		"C:/repo",
		"https://github.com/Team-NEEEE/envio-cli",
		command.NoopReporter{},
	)
	if appErr == nil || appErr.Code != ErrorSaveWrappedKeysFailed {
		t.Fatalf("Create() appErr = %#v, want %s", appErr, ErrorSaveWrappedKeysFailed)
	}
	if saved.called {
		t.Fatal("project session should not be saved when wrapped key save fails")
	}
}

func TestCreateReturnsLocalSessionSaveFailure(t *testing.T) {
	t.Parallel()

	client := &fakeCreateAPI{
		createResp: validCreateResponse(),
		saveResp:   validSaveResponse(),
	}
	service, _ := newTestCreateService(client)
	service.saveProjectSession = func(string, Session) error {
		return errors.New("disk full")
	}

	_, appErr := service.Create(
		context.Background(),
		"C:/repo",
		"https://github.com/Team-NEEEE/envio-cli",
		command.NoopReporter{},
	)
	if appErr == nil || appErr.Code != ErrorSaveProjectSessionFailed {
		t.Fatalf("Create() appErr = %#v, want %s", appErr, ErrorSaveProjectSessionFailed)
	}
}

type savedProjectSession struct {
	root    string
	session Session
	called  bool
}

func newTestCreateService(client *fakeCreateAPI) (*CreateService, *savedProjectSession) {
	saved := &savedProjectSession{}
	service := &CreateService{
		client: client,
		git: &fakeProjectGit{
			repository: workspace.GitRepository{
				Root:      "C:/repo",
				OriginURL: "git@github.com:Team-NEEEE/envio-cli.git",
			},
		},
		generateProjectMasterKey: func() ([]byte, error) {
			return []byte("12345678901234567890123456789012"), nil
		},
		wrapProjectMasterKey: func(publicKey string, _ []byte) (string, error) {
			return "wrapped-" + publicKey, nil
		},
		saveProjectSession: func(root string, session Session) error {
			saved.called = true
			saved.root = root
			saved.session = session
			return nil
		},
	}
	return service, saved
}

func validCreateResponse() *projectapi.CreateProjectResponse {
	return &projectapi.CreateProjectResponse{
		ProjectID:      1,
		ProjectName:    "envio-cli",
		GithubRepoName: "envio-cli",
		Members: []projectapi.ProjectMember{
			{UserID: 10, UserDeviceID: 20, GithubID: "octocat", PublicKey: "public-key-1"},
			{UserID: 11, UserDeviceID: 21, GithubID: "mona", PublicKey: "public-key-2"},
		},
	}
}

func validSaveResponse() *projectapi.SaveWrappedKeysResponse {
	return &projectapi.SaveWrappedKeysResponse{
		ProjectID:    1,
		UpdatedCount: 2,
	}
}

func containsAll(value string, parts ...string) bool {
	for _, part := range parts {
		if !strings.Contains(value, part) {
			return false
		}
	}
	return true
}
