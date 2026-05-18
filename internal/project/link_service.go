package project

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/Team-NEEEE/envio-cli/internal/api"
	projectapi "github.com/Team-NEEEE/envio-cli/internal/api/project"
	"github.com/Team-NEEEE/envio-cli/internal/command"
	"github.com/Team-NEEEE/envio-cli/internal/config"
	envcrypto "github.com/Team-NEEEE/envio-cli/internal/crypto"
	"github.com/Team-NEEEE/envio-cli/internal/github"
	"github.com/Team-NEEEE/envio-cli/internal/workspace"
)

const (
	StepLinkProject      = "link-project"
	StepUnwrapProjectKey = "unwrap-project-key"
	StepSaveLinkConfig   = "save-link-config"

	ErrorLinkProjectFailed       = "LINK_PROJECT_FAILED"
	ErrorLinkResponseInvalid     = "LINK_RESPONSE_INVALID"
	ErrorUnwrapProjectKeyFailed  = "UNWRAP_PROJECT_KEY_FAILED"
	ErrorSaveLinkConfigFailed    = "SAVE_LINK_CONFIG_FAILED"
	ErrorLocalLinkConfigConflict = "LOCAL_LINK_CONFIG_CONFLICT"

	JoinStatusApproved              = "APPROVED"
	ErrorRepositoryParseFailed      = "REPOSITORY_PARSE_FAILED"
	ErrorInvalidPublicKeyFormat     = "INVALID_PUBLIC_KEY_FORMAT"
	ErrorInvalidDeviceID            = "INVALID_DEVICE_ID"
	ErrorProjectAccessDenied        = "PROJECT_ACCESS_DENIED"
	ErrorProjectNotFound            = "PROJECT_NOT_FOUND"
	ErrorJoinStatusPending          = "JOIN_STATUS_PENDING"
	ErrorWrappedMasterKeyNotReady   = "WRAPPED_MASTER_KEY_NOT_READY"
	ErrorUserDeviceKeyNotRegistered = "USER_DEVICE_KEY_NOT_REGISTERED"
	ErrorPublicKeyMismatch          = "PUBLIC_KEY_MISMATCH"
	ErrorUnauthorized               = "UNAUTHORIZED"
	ErrorInternalServer             = "INTERNAL_SERVER_ERROR"
	ErrorInvalidRepositoryURL       = "INVALID_REPOSITORY_URL"
)

type LinkResult struct {
	ProjectName         string
	GithubRepoName      string
	Owner               string
	RepoName            string
	JoinStatus          string
	LocalRepositoryRoot string
	LocalConfigPath     string
	ProjectID           int64
}

type LinkService struct {
	client                  linkAPI
	clientErr               error
	git                     workspace.GitInspector
	loadGlobalSession       func() (config.GlobalSession, error)
	loadPrivateKey          func(int64) (string, error)
	unwrapProjectMasterKey  func(string, string) ([]byte, error)
	deleteProjectMasterKey  func(int64, int64) error
	saveLocalLinkConfig     func(string, LocalLinkConfig) error
	ensureLocalConfigTarget func(string) error
}

type linkAPI interface {
	LinkProject(context.Context, projectapi.LinkProjectRequest, string) (*projectapi.LinkProjectResponse, error)
}

func NewLinkService(apiURL string) *LinkService {
	baseURL := config.APIURLOrDefault(apiURL)
	client, err := projectapi.NewHTTPClient(baseURL, nil)
	return &LinkService{
		client:                  client,
		clientErr:               err,
		git:                     workspace.NewGitInspector(nil),
		loadGlobalSession:       config.LoadGlobalSession,
		loadPrivateKey:          envcrypto.LoadDevicePrivateKey,
		unwrapProjectMasterKey:  envcrypto.UnwrapProjectMasterKey,
		deleteProjectMasterKey:  envcrypto.DeleteProjectMasterKey,
		saveLocalLinkConfig:     saveLocalLinkConfig,
		ensureLocalConfigTarget: ensureLocalLinkConfigPathAvailable,
	}
}

func (s *LinkService) Link(
	ctx context.Context,
	cwd string,
	repositoryURL string,
	reporter command.Reporter,
) (*LinkResult, *command.AppError) {
	if reporter == nil {
		reporter = command.NoopReporter{}
	}
	if s == nil {
		return nil, newLinkAppError(ErrorLinkProjectFailed, "link service is not configured", "", 1)
	}
	s.ensureDefaults()

	reporter.UpdateStep(command.StepUpdate{ID: StepCheckRepository, Status: command.StatusRunning})
	repository, inputRef, resolvedURL, appErr := s.validateRepository(ctx, cwd, repositoryURL)
	if appErr != nil {
		reporter.UpdateStep(command.StepUpdate{ID: StepCheckRepository, Status: command.StatusError})
		return nil, appErr
	}
	reporter.UpdateStep(command.StepUpdate{ID: StepCheckRepository, Status: command.StatusSuccess})

	reporter.UpdateStep(command.StepUpdate{ID: StepLinkProject, Status: command.StatusRunning})
	session, appErr := s.requireGlobalSession()
	if appErr != nil {
		reporter.UpdateStep(command.StepUpdate{ID: StepLinkProject, Status: command.StatusError})
		return nil, appErr
	}
	if s.clientErr != nil {
		reporter.UpdateStep(command.StepUpdate{ID: StepLinkProject, Status: command.StatusError})
		return nil, newLinkAppError(ErrorLinkProjectFailed, "link project API client is not available", s.clientErr.Error(), 1)
	}
	if s.client == nil {
		reporter.UpdateStep(command.StepUpdate{ID: StepLinkProject, Status: command.StatusError})
		return nil, newLinkAppError(ErrorLinkProjectFailed, "link project API client is not configured", "", 1)
	}

	linkResponse, err := s.client.LinkProject(ctx, projectapi.LinkProjectRequest{
		PublicKey:     session.PublicKey,
		DeviceID:      session.DeviceID,
		UserGithubID:  session.GithubID,
		RepositoryURL: resolvedURL,
		Owner:         inputRef.Owner,
		RepoName:      inputRef.Repo,
	}, "")
	if err != nil {
		reporter.UpdateStep(command.StepUpdate{ID: StepLinkProject, Status: command.StatusError})
		return nil, linkAppErrorFromAPI(err)
	}
	if appErr := validateLinkResponse(linkResponse); appErr != nil {
		reporter.UpdateStep(command.StepUpdate{ID: StepLinkProject, Status: command.StatusError})
		return nil, appErr
	}
	reporter.UpdateStep(command.StepUpdate{ID: StepLinkProject, Status: command.StatusSuccess})

	reporter.UpdateStep(command.StepUpdate{ID: StepUnwrapProjectKey, Status: command.StatusRunning})
	privateKey, err := s.loadPrivateKey(session.DeviceID)
	if err != nil {
		reporter.UpdateStep(command.StepUpdate{ID: StepUnwrapProjectKey, Status: command.StatusError})
		return nil, newLinkAppError(
			ErrorUserDeviceKeyNotRegistered,
			"current device private key could not be loaded",
			"Run `envio login` again to register this device key.",
			1,
		)
	}
	projectMasterKey, err := s.unwrapProjectMasterKey(privateKey, linkResponse.WrappedMasterKey)
	if err != nil {
		reporter.UpdateStep(command.StepUpdate{ID: StepUnwrapProjectKey, Status: command.StatusError})
		return nil, newLinkAppError(ErrorUnwrapProjectKeyFailed, "unwrap project master key failed", err.Error(), 1)
	}
	reporter.UpdateStep(command.StepUpdate{ID: StepUnwrapProjectKey, Status: command.StatusSuccess})

	projectID := linkResponse.Project.ProjectID
	localConfig := LocalLinkConfig{
		LinkedProjectID: projectID,
		RepositoryURL:   resolvedURL,
		UserGithubID:    session.GithubID,
		DeviceID:        session.DeviceID,
		ProjectName:     projectNameFromLinkResponse(linkResponse, inputRef),
		GithubRepoName:  githubRepoNameFromLinkResponse(linkResponse, inputRef),
		MasterKey: MasterKeySession{
			Algorithm: projectMasterKeyAlgorithm,
			Encoding:  projectMasterKeyEncoding,
			Value:     base64.StdEncoding.EncodeToString(projectMasterKey),
		},
	}

	reporter.UpdateStep(command.StepUpdate{ID: StepSaveLinkConfig, Status: command.StatusRunning})
	if err := s.saveLocalLinkConfig(repository.Root, localConfig); err != nil {
		reporter.UpdateStep(command.StepUpdate{ID: StepSaveLinkConfig, Status: command.StatusError})
		return nil, newLinkAppError(ErrorSaveLinkConfigFailed, "save local link config failed", err.Error(), 1)
	}
	if err := s.deleteProjectMasterKey(projectID, session.DeviceID); err != nil {
		reporter.UpdateStep(command.StepUpdate{ID: StepSaveLinkConfig, Status: command.StatusError})
		return nil, newLinkAppError(ErrorSaveLinkConfigFailed, "delete legacy project master key failed", err.Error(), 1)
	}
	reporter.UpdateStep(command.StepUpdate{ID: StepSaveLinkConfig, Status: command.StatusSuccess})

	return &LinkResult{
		ProjectID:           projectID,
		ProjectName:         localConfig.ProjectName,
		GithubRepoName:      localConfig.GithubRepoName,
		Owner:               inputRef.Owner,
		RepoName:            inputRef.Repo,
		JoinStatus:          strings.TrimSpace(linkResponse.JoinStatus),
		LocalRepositoryRoot: repository.Root,
		LocalConfigPath:     localLinkConfigPath(repository.Root),
	}, nil
}

func (s *LinkService) ensureDefaults() {
	if s.git == nil {
		s.git = workspace.NewGitInspector(nil)
	}
	if s.loadGlobalSession == nil {
		s.loadGlobalSession = config.LoadGlobalSession
	}
	if s.loadPrivateKey == nil {
		s.loadPrivateKey = envcrypto.LoadDevicePrivateKey
	}
	if s.unwrapProjectMasterKey == nil {
		s.unwrapProjectMasterKey = envcrypto.UnwrapProjectMasterKey
	}
	if s.deleteProjectMasterKey == nil {
		s.deleteProjectMasterKey = envcrypto.DeleteProjectMasterKey
	}
	if s.saveLocalLinkConfig == nil {
		s.saveLocalLinkConfig = saveLocalLinkConfig
	}
	if s.ensureLocalConfigTarget == nil {
		s.ensureLocalConfigTarget = ensureLocalLinkConfigPathAvailable
	}
}

func (s *LinkService) validateRepository(
	ctx context.Context,
	cwd string,
	repositoryURL string,
) (workspace.GitRepository, github.RepositoryRef, string, *command.AppError) {
	repository, err := s.git.Inspect(ctx, cwd)
	if err != nil {
		return workspace.GitRepository{}, github.RepositoryRef{}, "", newLinkAppError(
			ErrorGitRepositoryRequired,
			"git repository with origin remote is required",
			"Run this command inside a Git repository with an origin remote.",
			1,
		)
	}
	if err := s.ensureLocalConfigTarget(repository.Root); err != nil {
		return workspace.GitRepository{}, github.RepositoryRef{}, "", newLinkAppError(
			ErrorLocalLinkConfigConflict,
			"local link config path conflicts with an existing file",
			err.Error(),
			1,
		)
	}

	resolvedURL := strings.TrimSpace(repositoryURL)
	if resolvedURL == "" {
		resolvedURL = strings.TrimSpace(repository.OriginURL)
	}

	inputRef, err := github.ParseRepositoryURL(resolvedURL)
	if err != nil {
		return workspace.GitRepository{}, github.RepositoryRef{}, "", newLinkAppError(
			ErrorRepositoryURLInvalid,
			"repository URL is invalid",
			"Use a GitHub repository URL such as https://github.com/owner/repo.",
			2,
		)
	}
	originRef, err := github.ParseRepositoryURL(repository.OriginURL)
	if err != nil {
		return workspace.GitRepository{}, github.RepositoryRef{}, "", newLinkAppError(
			ErrorRepositoryURLInvalid,
			"origin URL is invalid",
			"Set origin to a GitHub repository URL before running link.",
			1,
		)
	}
	if !inputRef.Equal(originRef) {
		return workspace.GitRepository{}, github.RepositoryRef{}, "", newLinkAppError(
			ErrorRepositoryContextMismatch,
			"repository URL does not match current Git repository",
			fmt.Sprintf("Current repository: %s\nInput repository: %s", originRef.String(), inputRef.String()),
			1,
		)
	}
	return repository, inputRef, resolvedURL, nil
}

func (s *LinkService) requireGlobalSession() (config.GlobalSession, *command.AppError) {
	session, err := s.loadGlobalSession()
	if err != nil {
		if os.IsNotExist(err) {
			return config.GlobalSession{}, newLinkAppError(
				ErrorLoginRequired,
				"login is required",
				"Run `envio login` before linking a project.",
				1,
			)
		}
		return config.GlobalSession{}, newLinkAppError(
			ErrorLoginRequired,
			"login session could not be loaded",
			err.Error(),
			1,
		)
	}
	if !session.Valid() {
		return config.GlobalSession{}, newLinkAppError(
			ErrorLoginRequired,
			"login session is invalid",
			"Run `envio login` before linking a project.",
			1,
		)
	}
	return session, nil
}

func validateLinkResponse(response *projectapi.LinkProjectResponse) *command.AppError {
	if response == nil {
		return newLinkAppError(ErrorLinkResponseInvalid, "link project response is empty", "Server response did not include project metadata.", 1)
	}
	if response.Project.ProjectID <= 0 {
		return newLinkAppError(ErrorLinkResponseInvalid, "link project response is missing projectId", "Server response did not include a valid projectId.", 1)
	}
	if strings.TrimSpace(response.WrappedMasterKey) == "" {
		return newLinkAppError(ErrorWrappedMasterKeyNotReady, "wrapped master key is not ready", "Project key distribution is required before linking.", 1)
	}
	if strings.TrimSpace(response.JoinStatus) == "" {
		return nil
	}
	if !strings.EqualFold(strings.TrimSpace(response.JoinStatus), JoinStatusApproved) {
		return newLinkAppError(ErrorJoinStatusPending, "project join approval is not completed", "Project join approval is not completed yet.", 1)
	}
	return nil
}

func projectNameFromLinkResponse(response *projectapi.LinkProjectResponse, ref github.RepositoryRef) string {
	name := strings.TrimSpace(response.Project.ProjectName)
	if name == "" {
		return ref.Repo
	}
	return name
}

func githubRepoNameFromLinkResponse(response *projectapi.LinkProjectResponse, ref github.RepositoryRef) string {
	name := strings.TrimSpace(response.Project.GithubRepoName)
	if name == "" {
		return ref.String()
	}
	return name
}

func linkAppErrorFromAPI(err error) *command.AppError {
	var apiErr *api.ErrorResponse
	if errors.As(err, &apiErr) {
		code := strings.TrimSpace(apiErr.Code)
		if code == "" {
			code = codeFromHTTPStatus(apiErr.Status)
		}
		if code == "" {
			code = ErrorLinkProjectFailed
		}
		message := strings.TrimSpace(apiErr.Message)
		if message == "" {
			message = "link project request failed"
		}
		return newLinkAppError(code, message, linkHintForCode(code), exitCodeForLinkError(code))
	}
	return newLinkAppError(ErrorLinkProjectFailed, "link project request failed", err.Error(), 1)
}

func codeFromHTTPStatus(status api.ErrorStatus) string {
	code, err := strconv.Atoi(string(status))
	if err != nil {
		return ""
	}
	switch code {
	case http.StatusUnauthorized:
		return ErrorUnauthorized
	case http.StatusForbidden:
		return ErrorProjectAccessDenied
	case http.StatusNotFound:
		return ErrorProjectNotFound
	case http.StatusConflict:
		return ErrorJoinStatusPending
	case http.StatusUnprocessableEntity:
		return ErrorWrappedMasterKeyNotReady
	case http.StatusInternalServerError:
		return ErrorInternalServer
	default:
		return ""
	}
}

func linkHintForCode(code string) string {
	switch code {
	case ErrorInvalidRepositoryURL:
		return "Use a valid GitHub repository URL."
	case ErrorRepositoryParseFailed:
		return "Use a GitHub repository URL that can be parsed as owner/repo."
	case ErrorInvalidPublicKeyFormat:
		return "Run `envio login` again to register a valid device public key."
	case ErrorInvalidDeviceID:
		return "Run `envio login` again to register this device."
	case ErrorUnauthorized:
		return "Run `envio login` before linking a project."
	case ErrorProjectAccessDenied:
		return "You do not have Envio access to this GitHub repository."
	case ErrorProjectNotFound:
		return "Run `envio create` before linking this repository."
	case ErrorJoinStatusPending:
		return "Project join approval is not completed yet."
	case ErrorWrappedMasterKeyNotReady:
		return "Project key distribution is required before linking."
	case ErrorUserDeviceKeyNotRegistered:
		return "Run `envio login` again to register this device key."
	case ErrorPublicKeyMismatch:
		return "Run `envio login` again because the local public key does not match the registered device key."
	case ErrorInternalServer:
		return "Try again after the server recovers."
	default:
		return ""
	}
}

func exitCodeForLinkError(code string) int {
	if code == ErrorInvalidRepositoryURL || code == ErrorRepositoryURLInvalid {
		return 2
	}
	return 1
}

func newLinkAppError(code, message, hint string, exitCode int) *command.AppError {
	return command.NewAppError(code, message, hint, exitCode, command.SeverityError)
}
