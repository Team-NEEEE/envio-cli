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
	// StepCheckRepository StepCheckRepository는 입력된 저장소 URL과 현재 로컬 Git origin이 같은 저장소인지 확인하는 단계다.
	StepCheckRepository = "check-repository"
	// StepCreateProject StepCreateProject는 백엔드에 프로젝트를 생성하고 멤버 공개키 목록을 받는 단계다.
	StepCreateProject = "create-project"
	// StepWrapProjectKey StepWrapProjectKey는 새 프로젝트 마스터 키를 멤버 디바이스 공개키별로 암호화하는 단계다.
	StepWrapProjectKey = "wrap-project-key"
	// StepSaveWrappedKeys StepSaveWrappedKeys는 암호화된 프로젝트 마스터 키 목록을 백엔드에 등록하는 단계다.
	StepSaveWrappedKeys = "save-wrapped-keys"
	// StepSaveProjectConfig stores the project config under .envio/config.
	StepSaveProjectConfig = "save-project-config"

	ErrorGitRepositoryRequired     = "GIT_REPOSITORY_REQUIRED"
	ErrorRepositoryURLInvalid      = "REPOSITORY_URL_INVALID"
	ErrorRepositoryContextMismatch = "REPOSITORY_CONTEXT_MISMATCH"
	ErrorCreateProjectFailed       = "CREATE_PROJECT_FAILED"
	ErrorWrapProjectKeyFailed      = "WRAP_PROJECT_KEY_FAILED"
	ErrorSaveWrappedKeysFailed     = "SAVE_WRAPPED_KEYS_FAILED"
	ErrorSaveProjectConfigFailed   = "SAVE_PROJECT_CONFIG_FAILED"
	ErrorCreateResponseInvalid     = "CREATE_RESPONSE_INVALID"
	ErrorLoginRequired             = "LOGIN_REQUIRED"
	ErrorGitHubAppNotInstalled     = "GITHUB_APP_NOT_INSTALLED"
	ErrorRepositoryAccessDenied    = "REPOSITORY_ACCESS_DENIED"
	ErrorProjectAlreadyExists      = "PROJECT_ALREADY_EXISTS"
	ErrorNoAvailableMemberDevice   = "NO_AVAILABLE_MEMBER_DEVICE"
)

const (
	projectMasterKeyAlgorithm = "project-master-key-v1"
	projectMasterKeyEncoding  = "base64"
)

// CreateResult CreateResult는 프로젝트 생성 성공 후 렌더러로 넘기는 요약 정보다.
// 민감한 masterKey 값은 결과에 넣지 않고 저장소 로컬 .envio 파일에만 저장한다.
type CreateResult struct {
	ProjectName           string
	GithubRepoName        string
	LocalRepositoryRoot   string
	LocalConfigPath       string
	ProjectID             int64
	WrappedKeyTargetCount int
	UpdatedCount          int
}

// CreateService CreateService는 저장소 검증, 백엔드 프로젝트 생성, 키 래핑, 로컬 세션 저장을 조율한다.
// 테스트에서는 각 의존성을 바꿔 끼워 네트워크, Git, 파일 시스템을 직접 건드리지 않고 흐름을 검증한다.
type CreateService struct {
	client                   createAPI
	clientErr                error
	git                      workspace.GitInspector
	generateProjectMasterKey func() ([]byte, error)
	wrapProjectMasterKey     func(string, []byte) (string, error)
	saveLocalLinkConfig      func(string, LocalLinkConfig) error
	loadGlobalSession        func() (config.GlobalSession, error)
}

// createAPI create 흐름에서 필요한 백엔드 호출만 담은 좁은 인터페이스다.
type createAPI interface {
	CreateProject(context.Context, projectapi.CreateProjectRequest) (*projectapi.CreateProjectResponse, error)
	SaveWrappedKeys(context.Context, int64, projectapi.SaveWrappedKeysRequest, string) (*projectapi.SaveWrappedKeysResponse, error)
}

// NewCreateService 운영 환경에서 사용할 프로젝트 생성 서비스를 만든다.
// HTTP 클라이언트 생성 실패는 보관해 두었다가 실제 create 단계에서 command.AppError로 변환한다.
func NewCreateService(apiURL string) *CreateService {
	client, err := projectapi.NewHTTPClient(config.APIURLOrDefault(apiURL), nil)
	return &CreateService{
		client:                   client,
		clientErr:                err,
		git:                      workspace.NewGitInspector(nil),
		generateProjectMasterKey: envcrypto.GenerateProjectMasterKey,
		wrapProjectMasterKey:     envcrypto.WrapProjectMasterKey,
		saveLocalLinkConfig:      saveLocalLinkConfig,
		loadGlobalSession:        config.LoadGlobalSession,
	}
}

// Create Create는 현재 작업 디렉터리가 속한 Git 저장소를 기준으로 프로젝트를 생성한다.
// 저장소 검증, 서버 프로젝트 생성, 멤버별 키 래핑, 서버 저장, 로컬 .envio 저장 순서로 진행한다.
func (s *CreateService) Create(
	ctx context.Context,
	cwd string,
	repositoryURL string,
	reporter command.Reporter,
) (*CreateResult, *command.AppError) {
	if reporter == nil {
		reporter = command.NoopReporter{}
	}
	if s == nil {
		return nil, newCreateAppError(ErrorCreateProjectFailed, "create service is not configured", "", 1)
	}
	s.ensureDefaults()
	repositoryURL = strings.TrimSpace(repositoryURL)

	// git rev-parse로 찾은 저장소 루트와 origin URL을 먼저 검증한다.
	// 이 단계가 통과해야 .envio를 어느 저장소 루트에 써야 하는지도 확정된다.
	reporter.UpdateStep(command.StepUpdate{ID: StepCheckRepository, Status: command.StatusRunning})
	repository, inputRef, appErr := s.validateRepository(ctx, cwd, repositoryURL)
	if appErr != nil {
		reporter.UpdateStep(command.StepUpdate{ID: StepCheckRepository, Status: command.StatusError})
		return nil, appErr
	}
	reporter.UpdateStep(command.StepUpdate{ID: StepCheckRepository, Status: command.StatusSuccess})

	reporter.UpdateStep(command.StepUpdate{ID: StepCreateProject, Status: command.StatusRunning})
	globalSession, appErr := s.requireGlobalSession()
	if appErr != nil {
		reporter.UpdateStep(command.StepUpdate{ID: StepCreateProject, Status: command.StatusError})
		return nil, appErr
	}
	// 클라이언트 생성 오류를 여기서 보고하면 사용자는 어느 단계에서 실패했는지 같은 UI 흐름으로 볼 수 있다.
	if s.clientErr != nil {
		reporter.UpdateStep(command.StepUpdate{ID: StepCreateProject, Status: command.StatusError})
		return nil, newCreateAppError(
			ErrorCreateProjectFailed,
			"create project API client is not available",
			s.clientErr.Error(),
			1,
		)
	}
	// 테스트나 수동 구성에서 nil 클라이언트가 들어와도 panic 대신 명령 오류로 돌려준다.
	if s.client == nil {
		reporter.UpdateStep(command.StepUpdate{ID: StepCreateProject, Status: command.StatusError})
		return nil, newCreateAppError(
			ErrorCreateProjectFailed,
			"create project API client is not configured",
			"",
			1,
		)
	}

	createResponse, err := s.client.CreateProject(ctx, projectapi.CreateProjectRequest{
		RepositoryURL: repositoryURL,
		DeviceID:      globalSession.DeviceID,
		PublicKey:     globalSession.PublicKey,
	})
	if err != nil {
		reporter.UpdateStep(command.StepUpdate{ID: StepCreateProject, Status: command.StatusError})
		return nil, createAppErrorFromAPI(err)
	}
	// 서버 응답에 projectId와 멤버 공개키가 없으면 이후 키 래핑이 성립하지 않으므로 즉시 중단한다.
	if appErr := validateCreateResponse(createResponse); appErr != nil {
		reporter.UpdateStep(command.StepUpdate{ID: StepCreateProject, Status: command.StatusError})
		return nil, appErr
	}
	reporter.UpdateStep(command.StepUpdate{ID: StepCreateProject, Status: command.StatusSuccess})

	reporter.UpdateStep(command.StepUpdate{ID: StepWrapProjectKey, Status: command.StatusRunning})
	// 프로젝트 마스터 키는 한 번만 생성하고, 같은 키를 멤버 디바이스 공개키별로 암호화한다.
	wrappedKeys, projectMasterKey, appErr := s.wrapMemberKeys(createResponse.Members)
	if appErr != nil {
		reporter.UpdateStep(command.StepUpdate{ID: StepWrapProjectKey, Status: command.StatusError})
		return nil, appErr
	}
	reporter.UpdateStep(command.StepUpdate{ID: StepWrapProjectKey, Status: command.StatusSuccess})

	reporter.UpdateStep(command.StepUpdate{ID: StepSaveWrappedKeys, Status: command.StatusRunning})
	saveResponse, err := s.client.SaveWrappedKeys(ctx, createResponse.ProjectID, projectapi.SaveWrappedKeysRequest{
		WrappedKeys: wrappedKeys,
		DeviceID:    globalSession.DeviceID,
		PublicKey:   globalSession.PublicKey,
	}, "")
	if err != nil {
		reporter.UpdateStep(command.StepUpdate{ID: StepSaveWrappedKeys, Status: command.StatusError})
		return nil, newCreateAppError(
			ErrorSaveWrappedKeysFailed,
			"save wrapped keys request failed",
			err.Error(),
			1,
		)
	}
	reporter.UpdateStep(command.StepUpdate{ID: StepSaveWrappedKeys, Status: command.StatusSuccess})

	// 출력용 결과에는 masterKey를 넣지 않는다. 민감 값은 아래 .envio 저장 단계에서만 다룬다.
	result := &CreateResult{
		ProjectID:             createResponse.ProjectID,
		ProjectName:           createResponse.ProjectName,
		GithubRepoName:        createResponse.GithubRepoName,
		LocalRepositoryRoot:   repository.Root,
		LocalConfigPath:       localLinkConfigPath(repository.Root),
		WrappedKeyTargetCount: len(wrappedKeys),
	}
	if result.ProjectName == "" {
		result.ProjectName = inputRef.Repo
	}
	if result.GithubRepoName == "" {
		result.GithubRepoName = inputRef.Repo
	}
	if saveResponse != nil {
		// TODO: Validate projectId and updatedCount after the backend defines partial-update semantics.
		result.UpdatedCount = saveResponse.UpdatedCount
	}

	reporter.UpdateStep(command.StepUpdate{ID: StepSaveProjectConfig, Status: command.StatusRunning})
	// .envio는 Git 저장소 루트 바로 아래에 저장된다. repository.Root는 .git이 있는 작업 트리 루트다.
	if err := s.saveLocalLinkConfig(repository.Root, LocalLinkConfig{
		LinkedProjectID: result.ProjectID,
		ProjectName:     result.ProjectName,
		GithubRepoName:  result.GithubRepoName,
		RepositoryURL:   repositoryURL,
		UserGithubID:    globalSession.GithubID,
		DeviceID:        globalSession.DeviceID,
		MasterKey: MasterKeySession{
			Algorithm: projectMasterKeyAlgorithm,
			Encoding:  projectMasterKeyEncoding,
			Value:     base64.StdEncoding.EncodeToString(projectMasterKey),
		},
	}); err != nil {
		reporter.UpdateStep(command.StepUpdate{ID: StepSaveProjectConfig, Status: command.StatusError})
		return nil, newCreateAppError(
			ErrorSaveProjectConfigFailed,
			"save project config failed",
			err.Error(),
			1,
		)
	}
	reporter.UpdateStep(command.StepUpdate{ID: StepSaveProjectConfig, Status: command.StatusSuccess})

	return result, nil
}

// ensureDefaults는 테스트에서 일부 의존성만 채운 CreateService가 들어와도 운영 기본값으로 보강한다.
func (s *CreateService) ensureDefaults() {
	if s.git == nil {
		s.git = workspace.NewGitInspector(nil)
	}
	if s.generateProjectMasterKey == nil {
		s.generateProjectMasterKey = envcrypto.GenerateProjectMasterKey
	}
	if s.wrapProjectMasterKey == nil {
		s.wrapProjectMasterKey = envcrypto.WrapProjectMasterKey
	}
	if s.saveLocalLinkConfig == nil {
		s.saveLocalLinkConfig = saveLocalLinkConfig
	}
	if s.loadGlobalSession == nil {
		s.loadGlobalSession = config.LoadGlobalSession
	}
}

// validateRepository는 현재 작업 디렉터리가 속한 Git 저장소와 입력 repositoryURL이 같은 GitHub 저장소인지 확인한다.
// 비교는 raw URL 문자열이 아니라 정규화된 owner/repo 기준으로 수행해 SSH/HTTPS 표기 차이를 허용한다.
func (s *CreateService) validateRepository(
	ctx context.Context,
	cwd string,
	repositoryURL string,
) (workspace.GitRepository, github.RepositoryRef, *command.AppError) {
	repository, err := s.git.Inspect(ctx, cwd)
	if err != nil {
		return workspace.GitRepository{}, github.RepositoryRef{}, newCreateAppError(
			ErrorGitRepositoryRequired,
			"git repository with origin remote is required",
			"Run this command inside a Git repository with an origin remote.",
			1,
		)
	}

	inputRef, err := github.ParseRepositoryURL(repositoryURL)
	if err != nil {
		return workspace.GitRepository{}, github.RepositoryRef{}, newCreateAppError(
			ErrorRepositoryURLInvalid,
			"repository URL is invalid",
			"Use a GitHub repository URL such as https://github.com/owner/repo.",
			2,
		)
	}
	originRef, err := github.ParseRepositoryURL(repository.OriginURL)
	if err != nil {
		return workspace.GitRepository{}, github.RepositoryRef{}, newCreateAppError(
			ErrorRepositoryURLInvalid,
			"origin URL is invalid",
			"Set origin to a GitHub repository URL before running create.",
			1,
		)
	}
	if !inputRef.Equal(originRef) {
		return workspace.GitRepository{}, github.RepositoryRef{}, newCreateAppError(
			ErrorRepositoryContextMismatch,
			"repository URL does not match current Git repository",
			fmt.Sprintf("Current repository: %s\nInput repository: %s", originRef.String(), inputRef.String()),
			1,
		)
	}

	return repository, inputRef, nil
}

func (s *CreateService) requireGlobalSession() (config.GlobalSession, *command.AppError) {
	session, err := s.loadGlobalSession()
	if err != nil {
		if os.IsNotExist(err) {
			return config.GlobalSession{}, newCreateAppError(
				ErrorLoginRequired,
				"login is required",
				"Run `envio login` before creating a project.",
				1,
			)
		}
		return config.GlobalSession{}, newCreateAppError(
			ErrorLoginRequired,
			"login session could not be loaded",
			err.Error(),
			1,
		)
	}
	if !session.Valid() {
		return config.GlobalSession{}, newCreateAppError(
			ErrorLoginRequired,
			"login session is invalid",
			"Run `envio login` before creating a project.",
			1,
		)
	}
	return session, nil
}

// validateCreateResponse는 백엔드 create 응답이 이후 암호화 작업에 필요한 최소 계약을 만족하는지 확인한다.
func validateCreateResponse(response *projectapi.CreateProjectResponse) *command.AppError {
	if response == nil {
		return newCreateAppError(
			ErrorCreateResponseInvalid,
			"create project response is empty",
			"Server response did not include project metadata.",
			1,
		)
	}
	if response.ProjectID <= 0 {
		return newCreateAppError(
			ErrorCreateResponseInvalid,
			"create project response is missing projectId",
			"Server response did not include a valid projectId.",
			1,
		)
	}
	if len(response.Members) == 0 {
		return newCreateAppError(
			ErrorCreateResponseInvalid,
			"create project response is missing members",
			"Server response did not include any member public keys.",
			1,
		)
	}
	for _, member := range response.Members {
		// TODO: Re-enable userId validation if the create response contract makes it mandatory.
		if member.UserDeviceID <= 0 || strings.TrimSpace(member.PublicKey) == "" {
			return newCreateAppError(
				ErrorCreateResponseInvalid,
				"create project response contains invalid member data",
				"Server response did not include valid userDeviceId and publicKey values.",
				1,
			)
		}
	}
	return nil
}

func createAppErrorFromAPI(err error) *command.AppError {
	var apiErr *api.ErrorResponse
	if errors.As(err, &apiErr) {
		code := normalizeCreateAPIErrorCode(apiErr)
		if code == "" {
			code = ErrorCreateProjectFailed
		}
		message := strings.TrimSpace(apiErr.Message)
		if message == "" {
			message = "create project request failed"
		}
		return newCreateAppError(code, message, createHintForCode(code), exitCodeForCreateError(code))
	}

	var httpErr *api.HTTPResponseError
	if errors.As(err, &httpErr) {
		switch httpErr.StatusCode {
		case http.StatusForbidden, http.StatusUnprocessableEntity:
			return newCreateAppError(
				ErrorGitHubAppNotInstalled,
				"GitHub App is not installed for this repository",
				githubAppInstallHint(),
				1,
			)
		case http.StatusUnauthorized:
			return newCreateAppError(
				ErrorUnauthorized,
				"create project request is unauthorized",
				createHintForCode(ErrorUnauthorized),
				1,
			)
		}
	}

	return newCreateAppError(ErrorCreateProjectFailed, "create project request failed", err.Error(), 1)
}

func normalizeCreateAPIErrorCode(apiErr *api.ErrorResponse) string {
	if apiErr == nil {
		return ErrorCreateProjectFailed
	}
	code := strings.TrimSpace(apiErr.Code)
	if code != "" {
		return code
	}

	status, ok := createHTTPStatusCode(apiErr.Status)
	if !ok {
		return ""
	}
	switch status {
	case http.StatusBadRequest:
		return ErrorInvalidRepositoryURL
	case http.StatusUnauthorized:
		return ErrorUnauthorized
	case http.StatusForbidden, http.StatusUnprocessableEntity:
		return ErrorGitHubAppNotInstalled
	case http.StatusConflict:
		return ErrorProjectAlreadyExists
	case http.StatusInternalServerError:
		return ErrorInternalServer
	default:
		return ""
	}
}

func createHTTPStatusCode(status api.ErrorStatus) (int, bool) {
	statusText := strings.TrimSpace(string(status))
	if fields := strings.Fields(statusText); len(fields) > 0 {
		statusText = fields[0]
	}
	code, err := strconv.Atoi(statusText)
	if err != nil {
		return 0, false
	}
	return code, true
}

func createHintForCode(code string) string {
	switch code {
	case ErrorGitHubAppNotInstalled, ErrorRepositoryAccessDenied:
		return githubAppInstallHint()
	case ErrorInvalidRepositoryURL, ErrorRepositoryURLInvalid:
		return "Use a valid GitHub repository URL."
	case ErrorRepositoryParseFailed:
		return "Use a GitHub repository URL that can be parsed as owner/repo."
	case ErrorUnauthorized:
		return "Run `envio login` before creating a project."
	case ErrorProjectAlreadyExists:
		return "Run `envio link` to connect this repository to the existing Envio project."
	case ErrorNoAvailableMemberDevice:
		return "Ask repository collaborators to run `envio login`, then run `envio create` again."
	case ErrorInternalServer:
		return "Try again after the server recovers."
	default:
		return ""
	}
}

func githubAppInstallHint() string {
	return fmt.Sprintf(
		"Install the Envio GitHub App for this repository: %s. After installation, run `envio create` again.",
		config.GitHubAppInstallURL,
	)
}

func exitCodeForCreateError(code string) int {
	if code == ErrorInvalidRepositoryURL ||
		code == ErrorRepositoryParseFailed ||
		code == ErrorRepositoryURLInvalid {
		return 2
	}
	return 1
}

// wrapMemberKeys는 하나의 프로젝트 마스터 키를 만들고 각 멤버 디바이스 공개키로 암호화한다.
// 평문 masterKey는 서버로 보내지 않고, 호출자가 로컬 .envio 저장에만 사용하도록 함께 반환한다.
func (s *CreateService) wrapMemberKeys(members []projectapi.ProjectMember) ([]projectapi.WrappedKey, []byte, *command.AppError) {
	projectMasterKey, err := s.generateProjectMasterKey()
	if err != nil {
		return nil, nil, newCreateAppError(
			ErrorWrapProjectKeyFailed,
			"project master key generation failed",
			err.Error(),
			1,
		)
	}

	wrappedKeys := make([]projectapi.WrappedKey, 0, len(members))
	for _, member := range members {
		encryptedKey, err := s.wrapProjectMasterKey(member.PublicKey, projectMasterKey)
		if err != nil {
			return nil, nil, newCreateAppError(
				ErrorWrapProjectKeyFailed,
				"project master key wrapping failed",
				fmt.Sprintf("Failed to wrap key for githubId=%s userId=%d userDeviceId=%d: %v", member.GithubID, member.UserID, member.UserDeviceID, err),
				1,
			)
		}
		wrappedKeys = append(wrappedKeys, projectapi.WrappedKey{
			UserID:       member.UserID,
			UserDeviceID: member.UserDeviceID,
			EncryptedKey: encryptedKey,
		})
	}
	return wrappedKeys, projectMasterKey, nil
}

// newCreateAppError는 create 서비스에서 사용하는 오류 형식을 command 패키지 계약에 맞춘다.
func newCreateAppError(code, message, hint string, exitCode int) *command.AppError {
	return command.NewAppError(code, message, hint, exitCode, command.SeverityError)
}
