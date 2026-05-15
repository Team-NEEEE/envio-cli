package project

import (
	"context"
	"errors"
	"testing"

	projectapi "github.com/Team-NEEEE/envio-cli/internal/api/project"
	"github.com/Team-NEEEE/envio-cli/internal/command"
	"github.com/Team-NEEEE/envio-cli/internal/config"
	"github.com/Team-NEEEE/envio-cli/internal/workspace"
)

type fakeLinkAPI struct {
	auth  string
	err   error
	resp  *projectapi.LinkProjectResponse
	req   projectapi.LinkProjectRequest
	calls int
}

func (f *fakeLinkAPI) LinkProject(
	_ context.Context,
	req projectapi.LinkProjectRequest,
	authorization string,
) (*projectapi.LinkProjectResponse, error) {
	f.calls++
	f.req = req
	f.auth = authorization
	return f.resp, f.err
}

func TestLinkUsesOriginWhenURLIsOmittedAndSavesNonSensitiveState(t *testing.T) {
	t.Parallel()

	client := &fakeLinkAPI{resp: validLinkResponse()}
	service, saved := newTestLinkService(client)

	got, appErr := service.Link(context.Background(), "C:/repo", "", command.NoopReporter{})
	if appErr != nil {
		t.Fatalf("Link() appErr = %v", appErr)
	}
	if got.ProjectID != 1 || got.ProjectName != "envio-cli" || got.JoinStatus != JoinStatusApproved {
		t.Fatalf("Link() = %#v", got)
	}
	if client.calls != 1 {
		t.Fatalf("api calls = %d, want 1", client.calls)
	}
	if client.auth != "" {
		t.Fatalf("authorization = %q", client.auth)
	}
	if client.req.UserGithubID != "octocat" ||
		client.req.RepositoryURL != "git@github.com:Team-NEEEE/envio-cli.git" ||
		client.req.PublicKey != "public-key" ||
		client.req.Owner != "Team-NEEEE" ||
		client.req.RepoName != "envio-cli" ||
		client.req.DeviceID != 20 {
		t.Fatalf("link request = %#v", client.req)
	}
	if saved.masterProject != 1 || saved.masterDevice != 20 || string(saved.masterKey) != string(testProjectMasterKey()) {
		t.Fatalf("saved master key = %#v", saved)
	}
	if !saved.metadataCalled || !saved.configCalled {
		t.Fatalf("metadata/config should be saved: %#v", saved)
	}
	if saved.metadata.ProjectID != 1 ||
		saved.metadata.Owner != "Team-NEEEE" ||
		saved.metadata.RepoName != "envio-cli" ||
		saved.metadata.GithubRepoName != "Team-NEEEE/envio-cli" {
		t.Fatalf("metadata = %#v", saved.metadata)
	}
	if saved.config.LinkedProjectID != 1 ||
		saved.config.RepositoryURL != "git@github.com:Team-NEEEE/envio-cli.git" ||
		saved.config.UserGithubID != "octocat" ||
		saved.config.DeviceID != 20 {
		t.Fatalf("local config = %#v", saved.config)
	}
}

func TestLinkStopsBeforeAPIWhenRepositoryContextMismatches(t *testing.T) {
	t.Parallel()

	client := &fakeLinkAPI{resp: validLinkResponse()}
	service, saved := newTestLinkService(client)

	_, appErr := service.Link(
		context.Background(),
		"C:/repo",
		"https://github.com/Team-NEEEE/envio-server",
		command.NoopReporter{},
	)
	if appErr == nil || appErr.Code != ErrorRepositoryContextMismatch {
		t.Fatalf("Link() appErr = %#v, want %s", appErr, ErrorRepositoryContextMismatch)
	}
	if client.calls != 0 {
		t.Fatalf("api calls = %d, want 0", client.calls)
	}
	if saved.metadataCalled || saved.configCalled || saved.masterCalled {
		t.Fatalf("nothing should be saved on mismatch: %#v", saved)
	}
}

func TestLinkReturnsJoinStatusPendingForUnapprovedSuccessResponse(t *testing.T) {
	t.Parallel()

	client := &fakeLinkAPI{resp: &projectapi.LinkProjectResponse{
		Project:          projectapi.LinkedProject{ProjectID: 1},
		WrappedMasterKey: "wrapped-key",
		JoinStatus:       "PENDING",
	}}
	service, saved := newTestLinkService(client)

	_, appErr := service.Link(context.Background(), "C:/repo", "", command.NoopReporter{})
	if appErr == nil || appErr.Code != ErrorJoinStatusPending {
		t.Fatalf("Link() appErr = %#v, want %s", appErr, ErrorJoinStatusPending)
	}
	if saved.masterCalled || saved.metadataCalled || saved.configCalled {
		t.Fatalf("nothing should be saved before approval: %#v", saved)
	}
}

func TestLinkAllowsMissingOptionalJoinStatus(t *testing.T) {
	t.Parallel()

	response := validLinkResponse()
	response.JoinStatus = ""
	client := &fakeLinkAPI{resp: response}
	service, saved := newTestLinkService(client)

	got, appErr := service.Link(context.Background(), "C:/repo", "", command.NoopReporter{})
	if appErr != nil {
		t.Fatalf("Link() appErr = %v", appErr)
	}
	if got.JoinStatus != "" {
		t.Fatalf("JoinStatus = %q, want empty optional value", got.JoinStatus)
	}
	if !saved.masterCalled || !saved.metadataCalled || !saved.configCalled {
		t.Fatalf("link should save state without optional joinStatus: %#v", saved)
	}
}

func TestLinkReturnsUnwrapFailureBeforeSaving(t *testing.T) {
	t.Parallel()

	client := &fakeLinkAPI{resp: validLinkResponse()}
	service, saved := newTestLinkService(client)
	service.unwrapProjectMasterKey = func(string, string) ([]byte, error) {
		return nil, errors.New("bad wrapped key")
	}

	_, appErr := service.Link(context.Background(), "C:/repo", "", command.NoopReporter{})
	if appErr == nil || appErr.Code != ErrorUnwrapProjectKeyFailed {
		t.Fatalf("Link() appErr = %#v, want %s", appErr, ErrorUnwrapProjectKeyFailed)
	}
	if saved.masterCalled || saved.metadataCalled || saved.configCalled {
		t.Fatalf("nothing should be saved when unwrap fails: %#v", saved)
	}
}

type savedLinkState struct {
	masterKey      []byte
	metadataRoot   string
	configRoot     string
	metadata       projectMetadata
	config         LocalLinkConfig
	metadataCalled bool
	configCalled   bool
	masterCalled   bool
	masterProject  int64
	masterDevice   int64
}

func newTestLinkService(client *fakeLinkAPI) (*LinkService, *savedLinkState) {
	saved := &savedLinkState{}
	service := &LinkService{
		client: client,
		git: &fakeProjectGit{
			repository: workspace.GitRepository{
				Root:      "C:/repo",
				OriginURL: "git@github.com:Team-NEEEE/envio-cli.git",
			},
		},
		loadGlobalSession: func() (config.GlobalSession, error) {
			return validLinkSession(), nil
		},
		loadPrivateKey: func(deviceID int64) (string, error) {
			if deviceID != 20 {
				return "", errors.New("unexpected device id")
			}
			return "private-key", nil
		},
		unwrapProjectMasterKey: func(privateKey string, wrapped string) ([]byte, error) {
			if privateKey != "private-key" || wrapped != "wrapped-key" {
				return nil, errors.New("unexpected key input")
			}
			return testProjectMasterKey(), nil
		},
		saveProjectMasterKey: func(projectID int64, deviceID int64, projectMasterKey []byte) error {
			saved.masterCalled = true
			saved.masterProject = projectID
			saved.masterDevice = deviceID
			saved.masterKey = append([]byte(nil), projectMasterKey...)
			return nil
		},
		saveProjectMetadata: func(root string, metadata projectMetadata) error {
			saved.metadataCalled = true
			saved.metadataRoot = root
			saved.metadata = metadata
			return nil
		},
		saveLocalLinkConfig: func(root string, cfg LocalLinkConfig) error {
			saved.configCalled = true
			saved.configRoot = root
			saved.config = cfg
			return nil
		},
		ensureLocalConfigTarget: func(string) error {
			return nil
		},
	}
	return service, saved
}

func validLinkSession() config.GlobalSession {
	return config.GlobalSession{
		UserID:     10,
		GithubID:   "octocat",
		DeviceID:   20,
		DeviceName: "desktop",
		PublicKey:  "public-key",
	}
}

func validLinkResponse() *projectapi.LinkProjectResponse {
	return &projectapi.LinkProjectResponse{
		Message: "linked",
		Project: projectapi.LinkedProject{
			ProjectID:      1,
			ProjectName:    "envio-cli",
			GithubRepoName: "Team-NEEEE/envio-cli",
			Owner:          "Team-NEEEE",
		},
		WrappedMasterKey: "wrapped-key",
		JoinStatus:       JoinStatusApproved,
	}
}

func testProjectMasterKey() []byte {
	return []byte("12345678901234567890123456789012")
}
