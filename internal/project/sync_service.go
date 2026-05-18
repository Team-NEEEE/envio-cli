package project

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
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
	StepLoadProjectContext  = "load-project-context"
	StepReadEnvironmentFile = "read-environment-file"
	StepEncryptEnvironment  = "encrypt-environment"
	StepPushEnvironment     = "push-environment"
	StepPullEnvironment     = "pull-environment"
	StepDecryptEnvironment  = "decrypt-environment"
	StepWriteEnvironment    = "write-environment-file" // #nosec G101 -- step identifier, not a credential.
	StepSaveSyncState       = "save-sync-state"

	ErrorProjectLinkRequired         = "PROJECT_LINK_REQUIRED"
	ErrorLocalSessionMismatch        = "LOCAL_SESSION_MISMATCH"
	ErrorProjectMasterKeyRequired    = "PROJECT_MASTER_KEY_REQUIRED"
	ErrorEnvironmentFileNotFound     = "ENVIRONMENT_FILE_NOT_FOUND"
	ErrorEnvironmentFileInvalid      = "ENVIRONMENT_FILE_INVALID"
	ErrorEncryptEnvironmentFailed    = "ENCRYPT_ENVIRONMENT_FAILED"
	ErrorDecryptEnvironmentFailed    = "DECRYPT_ENVIRONMENT_FAILED"
	ErrorPushEnvironmentFailed       = "PUSH_ENVIRONMENT_FAILED"
	ErrorPullEnvironmentFailed       = "PULL_ENVIRONMENT_FAILED"
	ErrorPullResponseInvalid         = "PULL_RESPONSE_INVALID"
	ErrorPushResponseInvalid         = "PUSH_RESPONSE_INVALID"
	ErrorWriteEnvironmentFileFailed  = "WRITE_ENVIRONMENT_FILE_FAILED"
	ErrorSaveSyncStateFailed         = "SAVE_SYNC_STATE_FAILED"
	ErrorEnvironmentVersionNotReady  = "ENVIRONMENT_VERSION_NOT_INITIALIZED"
	ErrorVersionConflict             = "VERSION_CONFLICT"
	defaultEnvironmentFileName       = ".env"
	backendCodeProjectNotFound       = "PR-001"
	backendCodeVersionNotInitialized = "PR-002"
	backendCodeVersionConflict       = "PR-003"
	backendCodeAccessDenied          = "C-005"
)

type PushResult struct {
	EnvironmentFile string
	LocalRepository string
	VariableCount   int
	ProjectID       int64
	VersionID       int64
	ParentVersionID int64
	HistoryID       int64
}

type PullResult struct {
	EnvironmentFile string
	LocalRepository string
	VariableCount   int
	ProjectID       int64
	VersionID       int64
	HistoryID       int64
}

type SyncService struct {
	client             syncAPI
	clientErr          error
	git                workspace.GitInspector
	loadGlobalSession  func() (config.GlobalSession, error)
	readFile           func(string) ([]byte, error)
	writeFile          func(string, []byte, os.FileMode) error
	parseEnvironment   func([]byte) (map[string]string, error)
	encryptEnvironment func([]byte, []byte) (map[string]any, error)
	decryptEnvironment func(map[string]any, []byte) ([]byte, error)
	loadLocalContext   func(context.Context, string) (localProjectContext, *command.AppError)
	saveLocalVersion   func(localProjectContext, int64) error
}

type syncAPI interface {
	PullLatest(context.Context, int64, string, string) (*projectapi.ProjectPullResponse, error)
	Push(context.Context, int64, projectapi.ProjectPushRequest, string) (*projectapi.ProjectPushResponse, error)
}

type localProjectSource string

const (
	localProjectSourceLinkConfig localProjectSource = "link-config"
	localProjectSourceSession    localProjectSource = "session"
)

type localProjectContext struct {
	repositoryRoot string
	repositoryURL  string
	githubUserID   string
	source         localProjectSource
	masterKey      []byte
	session        Session
	linkConfig     LocalLinkConfig
	projectID      int64
	deviceID       int64
	versionID      int64
}

func NewSyncService(apiURL string) *SyncService {
	client, err := projectapi.NewHTTPClient(config.APIURLOrDefault(apiURL), nil)
	return &SyncService{
		client:    client,
		clientErr: err,
	}
}

func (s *SyncService) Push(
	ctx context.Context,
	cwd string,
	environmentFile string,
	reporter command.Reporter,
) (*PushResult, *command.AppError) {
	if reporter == nil {
		reporter = command.NoopReporter{}
	}
	if s == nil {
		return nil, newSyncAppError(ErrorPushEnvironmentFailed, "sync service is not configured", "", 1)
	}
	s.ensureDefaults()

	reporter.UpdateStep(command.StepUpdate{ID: StepLoadProjectContext, Status: command.StatusRunning})
	local, appErr := s.loadLocalContext(ctx, cwd)
	if appErr != nil {
		reporter.UpdateStep(command.StepUpdate{ID: StepLoadProjectContext, Status: command.StatusError})
		return nil, appErr
	}
	reporter.UpdateStep(command.StepUpdate{ID: StepLoadProjectContext, Status: command.StatusSuccess})

	envPath := resolveEnvironmentPath(cwd, local.repositoryRoot, environmentFile)
	reporter.UpdateStep(command.StepUpdate{ID: StepReadEnvironmentFile, Status: command.StatusRunning})
	raw, err := s.readFile(envPath)
	if err != nil {
		reporter.UpdateStep(command.StepUpdate{ID: StepReadEnvironmentFile, Status: command.StatusError})
		if os.IsNotExist(err) {
			return nil, newSyncAppError(
				ErrorEnvironmentFileNotFound,
				"environment file does not exist",
				fmt.Sprintf("Create %s before running `envio push`.", envPath),
				1,
			)
		}
		return nil, newSyncAppError(ErrorEnvironmentFileInvalid, "environment file could not be read", err.Error(), 1)
	}
	values, err := s.parseEnvironment(raw)
	if err != nil {
		reporter.UpdateStep(command.StepUpdate{ID: StepReadEnvironmentFile, Status: command.StatusError})
		return nil, newSyncAppError(ErrorEnvironmentFileInvalid, "environment file is invalid", err.Error(), 2)
	}
	reporter.UpdateStep(command.StepUpdate{ID: StepReadEnvironmentFile, Status: command.StatusSuccess})

	reporter.UpdateStep(command.StepUpdate{ID: StepEncryptEnvironment, Status: command.StatusRunning})
	encrypted, err := s.encryptEnvironment(raw, local.masterKey)
	if err != nil {
		reporter.UpdateStep(command.StepUpdate{ID: StepEncryptEnvironment, Status: command.StatusError})
		return nil, newSyncAppError(ErrorEncryptEnvironmentFailed, "environment encryption failed", err.Error(), 1)
	}
	reporter.UpdateStep(command.StepUpdate{ID: StepEncryptEnvironment, Status: command.StatusSuccess})

	reporter.UpdateStep(command.StepUpdate{ID: StepPushEnvironment, Status: command.StatusRunning})
	if appErr := s.ensureClientReady(ErrorPushEnvironmentFailed); appErr != nil {
		reporter.UpdateStep(command.StepUpdate{ID: StepPushEnvironment, Status: command.StatusError})
		return nil, appErr
	}
	response, err := s.client.Push(ctx, local.projectID, projectapi.ProjectPushRequest{
		GithubUserID:         local.githubUserID,
		EncryptedEnvironment: encrypted,
		ParentVersionID:      local.versionID,
	}, "")
	if err != nil {
		reporter.UpdateStep(command.StepUpdate{ID: StepPushEnvironment, Status: command.StatusError})
		return nil, syncAppErrorFromAPI(ErrorPushEnvironmentFailed, "push environment request failed", err)
	}
	if appErr := validatePushResponse(response, local.projectID); appErr != nil {
		reporter.UpdateStep(command.StepUpdate{ID: StepPushEnvironment, Status: command.StatusError})
		return nil, appErr
	}
	reporter.UpdateStep(command.StepUpdate{ID: StepPushEnvironment, Status: command.StatusSuccess})

	reporter.UpdateStep(command.StepUpdate{ID: StepSaveSyncState, Status: command.StatusRunning})
	if err := s.saveLocalVersion(local, response.VersionID); err != nil {
		reporter.UpdateStep(command.StepUpdate{ID: StepSaveSyncState, Status: command.StatusError})
		return nil, newSyncAppError(ErrorSaveSyncStateFailed, "save local sync state failed", err.Error(), 1)
	}
	reporter.UpdateStep(command.StepUpdate{ID: StepSaveSyncState, Status: command.StatusSuccess})

	return &PushResult{
		ProjectID:       response.ProjectID,
		VersionID:       response.VersionID,
		ParentVersionID: response.ParentVersionID,
		HistoryID:       response.HistoryID,
		VariableCount:   len(values),
		EnvironmentFile: envPath,
		LocalRepository: local.repositoryRoot,
	}, nil
}

func (s *SyncService) Pull(
	ctx context.Context,
	cwd string,
	environmentFile string,
	reporter command.Reporter,
) (*PullResult, *command.AppError) {
	if reporter == nil {
		reporter = command.NoopReporter{}
	}
	if s == nil {
		return nil, newSyncAppError(ErrorPullEnvironmentFailed, "sync service is not configured", "", 1)
	}
	s.ensureDefaults()

	reporter.UpdateStep(command.StepUpdate{ID: StepLoadProjectContext, Status: command.StatusRunning})
	local, appErr := s.loadLocalContext(ctx, cwd)
	if appErr != nil {
		reporter.UpdateStep(command.StepUpdate{ID: StepLoadProjectContext, Status: command.StatusError})
		return nil, appErr
	}
	reporter.UpdateStep(command.StepUpdate{ID: StepLoadProjectContext, Status: command.StatusSuccess})

	reporter.UpdateStep(command.StepUpdate{ID: StepPullEnvironment, Status: command.StatusRunning})
	if appErr := s.ensureClientReady(ErrorPullEnvironmentFailed); appErr != nil {
		reporter.UpdateStep(command.StepUpdate{ID: StepPullEnvironment, Status: command.StatusError})
		return nil, appErr
	}
	response, err := s.client.PullLatest(ctx, local.projectID, local.githubUserID, "")
	if err != nil {
		reporter.UpdateStep(command.StepUpdate{ID: StepPullEnvironment, Status: command.StatusError})
		return nil, syncAppErrorFromAPI(ErrorPullEnvironmentFailed, "pull environment request failed", err)
	}
	if appErr := validatePullResponse(response, local.projectID); appErr != nil {
		reporter.UpdateStep(command.StepUpdate{ID: StepPullEnvironment, Status: command.StatusError})
		return nil, appErr
	}
	reporter.UpdateStep(command.StepUpdate{ID: StepPullEnvironment, Status: command.StatusSuccess})

	reporter.UpdateStep(command.StepUpdate{ID: StepDecryptEnvironment, Status: command.StatusRunning})
	raw, err := s.decryptEnvironment(response.EncryptedEnvironment, local.masterKey)
	if err != nil {
		reporter.UpdateStep(command.StepUpdate{ID: StepDecryptEnvironment, Status: command.StatusError})
		return nil, newSyncAppError(ErrorDecryptEnvironmentFailed, "environment decryption failed", err.Error(), 1)
	}
	values, err := s.parseEnvironment(raw)
	if err != nil {
		reporter.UpdateStep(command.StepUpdate{ID: StepDecryptEnvironment, Status: command.StatusError})
		return nil, newSyncAppError(ErrorDecryptEnvironmentFailed, "decrypted environment file is invalid", err.Error(), 1)
	}
	reporter.UpdateStep(command.StepUpdate{ID: StepDecryptEnvironment, Status: command.StatusSuccess})

	envPath := resolveEnvironmentPath(cwd, local.repositoryRoot, environmentFile)
	reporter.UpdateStep(command.StepUpdate{ID: StepWriteEnvironment, Status: command.StatusRunning})
	if err := s.writeFile(envPath, raw, 0600); err != nil {
		reporter.UpdateStep(command.StepUpdate{ID: StepWriteEnvironment, Status: command.StatusError})
		return nil, newSyncAppError(ErrorWriteEnvironmentFileFailed, "write environment file failed", err.Error(), 1)
	}
	reporter.UpdateStep(command.StepUpdate{ID: StepWriteEnvironment, Status: command.StatusSuccess})

	reporter.UpdateStep(command.StepUpdate{ID: StepSaveSyncState, Status: command.StatusRunning})
	if err := s.saveLocalVersion(local, response.VersionID); err != nil {
		reporter.UpdateStep(command.StepUpdate{ID: StepSaveSyncState, Status: command.StatusError})
		return nil, newSyncAppError(ErrorSaveSyncStateFailed, "save local sync state failed", err.Error(), 1)
	}
	reporter.UpdateStep(command.StepUpdate{ID: StepSaveSyncState, Status: command.StatusSuccess})

	return &PullResult{
		ProjectID:       response.ProjectID,
		VersionID:       response.VersionID,
		HistoryID:       response.HistoryID,
		VariableCount:   len(values),
		EnvironmentFile: envPath,
		LocalRepository: local.repositoryRoot,
	}, nil
}

func (s *SyncService) ensureDefaults() {
	if s.git == nil {
		s.git = workspace.NewGitInspector(nil)
	}
	if s.loadGlobalSession == nil {
		s.loadGlobalSession = config.LoadGlobalSession
	}
	if s.readFile == nil {
		s.readFile = os.ReadFile
	}
	if s.writeFile == nil {
		s.writeFile = writeEnvironmentFileAtomic
	}
	if s.parseEnvironment == nil {
		s.parseEnvironment = config.ParseDotenv
	}
	if s.encryptEnvironment == nil {
		s.encryptEnvironment = envcrypto.EncryptEnvironment
	}
	if s.decryptEnvironment == nil {
		s.decryptEnvironment = envcrypto.DecryptEnvironment
	}
	if s.loadLocalContext == nil {
		s.loadLocalContext = s.loadDefaultLocalContext
	}
	if s.saveLocalVersion == nil {
		s.saveLocalVersion = saveLocalVersion
	}
}

func (s *SyncService) ensureClientReady(code string) *command.AppError {
	if s.clientErr != nil {
		return newSyncAppError(code, "project sync API client is not available", s.clientErr.Error(), 1)
	}
	if s.client == nil {
		return newSyncAppError(code, "project sync API client is not configured", "", 1)
	}
	return nil
}

func (s *SyncService) loadDefaultLocalContext(ctx context.Context, cwd string) (localProjectContext, *command.AppError) {
	repository, err := s.git.Inspect(ctx, cwd)
	if err != nil {
		return localProjectContext{}, newSyncAppError(
			ErrorGitRepositoryRequired,
			"git repository with origin remote is required",
			"Run this command inside a Git repository with an origin remote.",
			1,
		)
	}

	globalSession, appErr := requireValidGlobalSession(s.loadGlobalSession)
	if appErr != nil {
		return localProjectContext{}, appErr
	}

	linkConfig, linkErr := loadLocalLinkConfig(repository.Root)
	session, sessionErr := loadProjectSession(repository.Root)
	switch {
	case linkErr == nil:
		return s.localContextFromLinkConfig(repository, globalSession, linkConfig)
	case sessionErr == nil:
		return localContextFromSession(repository, globalSession, session)
	case os.IsNotExist(linkErr) && os.IsNotExist(sessionErr):
		return localProjectContext{}, newSyncAppError(
			ErrorProjectLinkRequired,
			"local project link config is missing",
			"Run `envio link` or `envio create` before syncing environment variables.",
			1,
		)
	case !os.IsNotExist(linkErr):
		return localProjectContext{}, newSyncAppError(ErrorProjectLinkRequired, "local project link config could not be loaded", linkErr.Error(), 1)
	default:
		return localProjectContext{}, newSyncAppError(ErrorProjectLinkRequired, "local project session could not be loaded", sessionErr.Error(), 1)
	}
}

func (s *SyncService) localContextFromLinkConfig(
	repository workspace.GitRepository,
	globalSession config.GlobalSession,
	linkConfig LocalLinkConfig,
) (localProjectContext, *command.AppError) {
	if err := validateLocalLinkConfig(linkConfig); err != nil {
		return localProjectContext{}, newSyncAppError(ErrorProjectLinkRequired, "local project link config is invalid", err.Error(), 1)
	}
	if appErr := validateRepositoryMatches(repository.OriginURL, linkConfig.RepositoryURL); appErr != nil {
		return localProjectContext{}, appErr
	}
	if !strings.EqualFold(linkConfig.UserGithubID, globalSession.GithubID) || linkConfig.DeviceID != globalSession.DeviceID {
		return localProjectContext{}, newSyncAppError(
			ErrorLocalSessionMismatch,
			"local project session does not match current login",
			"Run `envio link` again with the current login session.",
			1,
		)
	}

	masterKey, err := projectMasterKeyFromConfig(linkConfig)
	if err != nil {
		return localProjectContext{}, newSyncAppError(
			ErrorProjectMasterKeyRequired,
			"project master key could not be loaded from local config",
			"Run `envio link` again to restore this project's key on the current device.",
			1,
		)
	}
	if err := envcrypto.ValidateProjectMasterKey(masterKey); err != nil {
		return localProjectContext{}, newSyncAppError(ErrorProjectMasterKeyRequired, "project master key is invalid", err.Error(), 1)
	}

	return localProjectContext{
		source:         localProjectSourceLinkConfig,
		linkConfig:     linkConfig,
		repositoryRoot: repository.Root,
		repositoryURL:  linkConfig.RepositoryURL,
		githubUserID:   linkConfig.UserGithubID,
		projectID:      linkConfig.LinkedProjectID,
		deviceID:       linkConfig.DeviceID,
		versionID:      linkConfig.VersionID,
		masterKey:      masterKey,
	}, nil
}

func localContextFromSession(
	repository workspace.GitRepository,
	globalSession config.GlobalSession,
	session Session,
) (localProjectContext, *command.AppError) {
	if err := validateProjectSession(session); err != nil {
		return localProjectContext{}, newSyncAppError(ErrorProjectLinkRequired, "local project session is invalid", err.Error(), 1)
	}
	if appErr := validateRepositoryMatches(repository.OriginURL, session.RepositoryURL); appErr != nil {
		return localProjectContext{}, appErr
	}

	masterKey, err := projectMasterKeyFromSession(session)
	if err != nil {
		return localProjectContext{}, newSyncAppError(ErrorProjectMasterKeyRequired, "project master key is invalid", err.Error(), 1)
	}

	return localProjectContext{
		source:         localProjectSourceSession,
		session:        session,
		repositoryRoot: repository.Root,
		repositoryURL:  session.RepositoryURL,
		githubUserID:   globalSession.GithubID,
		projectID:      session.ProjectID,
		deviceID:       globalSession.DeviceID,
		versionID:      session.VersionID,
		masterKey:      masterKey,
	}, nil
}

func requireValidGlobalSession(load func() (config.GlobalSession, error)) (config.GlobalSession, *command.AppError) {
	session, err := load()
	if err != nil {
		if os.IsNotExist(err) {
			return config.GlobalSession{}, newSyncAppError(
				ErrorLoginRequired,
				"login is required",
				"Run `envio login` before syncing environment variables.",
				1,
			)
		}
		return config.GlobalSession{}, newSyncAppError(ErrorLoginRequired, "login session could not be loaded", err.Error(), 1)
	}
	if !session.Valid() {
		return config.GlobalSession{}, newSyncAppError(
			ErrorLoginRequired,
			"login session is invalid",
			"Run `envio login` before syncing environment variables.",
			1,
		)
	}
	return session, nil
}

func validateRepositoryMatches(originURL, configuredURL string) *command.AppError {
	originRef, err := github.ParseRepositoryURL(originURL)
	if err != nil {
		return newSyncAppError(
			ErrorRepositoryURLInvalid,
			"origin URL is invalid",
			"Set origin to a GitHub repository URL before syncing.",
			1,
		)
	}
	configuredRef, err := github.ParseRepositoryURL(configuredURL)
	if err != nil {
		return newSyncAppError(
			ErrorRepositoryURLInvalid,
			"linked repository URL is invalid",
			"Run `envio link` again with a valid GitHub repository URL.",
			1,
		)
	}
	if !originRef.Equal(configuredRef) {
		return newSyncAppError(
			ErrorRepositoryContextMismatch,
			"linked repository URL does not match current Git repository",
			fmt.Sprintf("Current repository: %s\nLinked repository: %s", originRef.String(), configuredRef.String()),
			1,
		)
	}
	return nil
}

func loadLocalLinkConfig(repositoryRoot string) (LocalLinkConfig, error) {
	raw, err := os.ReadFile(localLinkConfigPath(repositoryRoot))
	if err != nil {
		return LocalLinkConfig{}, err
	}
	var cfg LocalLinkConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return LocalLinkConfig{}, fmt.Errorf("decode local link config: %w", err)
	}
	return cfg, nil
}

func loadProjectSession(repositoryRoot string) (Session, error) {
	legacyPath := legacyLocalSessionPath(repositoryRoot)
	if info, err := os.Stat(legacyPath); err == nil && !info.IsDir() {
		return readProjectSessionFile(legacyPath)
	} else if err != nil && !os.IsNotExist(err) {
		return Session{}, err
	}

	session, err := readProjectSessionFile(localSessionPath(repositoryRoot))
	if err == nil {
		return session, nil
	}
	if !os.IsNotExist(err) {
		return Session{}, err
	}
	return Session{}, err
}

func readProjectSessionFile(path string) (Session, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Session{}, err
	}
	var sessionFile SessionFile
	if err := json.Unmarshal(raw, &sessionFile); err != nil {
		return Session{}, fmt.Errorf("decode project session: %w", err)
	}
	return sessionFile.Session, nil
}

func projectMasterKeyFromSession(session Session) ([]byte, error) {
	return projectMasterKeyFromMasterKeySession(session.MasterKey)
}

func projectMasterKeyFromConfig(cfg LocalLinkConfig) ([]byte, error) {
	return projectMasterKeyFromMasterKeySession(cfg.MasterKey)
}

func projectMasterKeyFromMasterKeySession(masterKeySession MasterKeySession) ([]byte, error) {
	if !strings.EqualFold(strings.TrimSpace(masterKeySession.Encoding), projectMasterKeyEncoding) {
		return nil, fmt.Errorf("unsupported project master key encoding: %s", masterKeySession.Encoding)
	}
	masterKey, err := base64.StdEncoding.DecodeString(strings.TrimSpace(masterKeySession.Value))
	if err != nil {
		return nil, fmt.Errorf("decode project master key: %w", err)
	}
	if err := envcrypto.ValidateProjectMasterKey(masterKey); err != nil {
		return nil, err
	}
	return masterKey, nil
}

func saveLocalVersion(local localProjectContext, versionID int64) error {
	if versionID <= 0 {
		return fmt.Errorf("versionId must be positive")
	}
	switch local.source {
	case localProjectSourceLinkConfig:
		cfg := local.linkConfig
		cfg.VersionID = versionID
		return writeJSONAtomic(localLinkConfigPath(local.repositoryRoot), cfg, 0600)
	case localProjectSourceSession:
		session := local.session
		session.VersionID = versionID
		return saveLocalLinkConfig(local.repositoryRoot, LocalLinkConfig{
			LinkedProjectID: session.ProjectID,
			ProjectName:     session.ProjectName,
			GithubRepoName:  session.GithubRepoName,
			RepositoryURL:   session.RepositoryURL,
			UserGithubID:    local.githubUserID,
			DeviceID:        local.deviceID,
			VersionID:       session.VersionID,
			MasterKey:       session.MasterKey,
		})
	default:
		return fmt.Errorf("unknown local project source: %s", local.source)
	}
}

func resolveEnvironmentPath(cwd, repositoryRoot, environmentFile string) string {
	environmentFile = strings.TrimSpace(environmentFile)
	if environmentFile == "" {
		return filepath.Join(repositoryRoot, defaultEnvironmentFileName)
	}
	if filepath.IsAbs(environmentFile) {
		return filepath.Clean(environmentFile)
	}
	if strings.TrimSpace(cwd) == "" {
		cwd = repositoryRoot
	}
	return filepath.Join(cwd, environmentFile)
}

func writeEnvironmentFileAtomic(path string, raw []byte, perm os.FileMode) error {
	if strings.TrimSpace(path) == "" {
		return errors.New("environment file path is required")
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create environment file directory: %w", err)
	}
	tempPath := path + ".tmp"
	if err := os.WriteFile(tempPath, raw, perm); err != nil {
		return fmt.Errorf("write temporary environment file: %w", err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("replace environment file: %w", err)
	}
	return nil
}

func validatePushResponse(response *projectapi.ProjectPushResponse, projectID int64) *command.AppError {
	if response == nil {
		return newSyncAppError(ErrorPushResponseInvalid, "push response is empty", "Server response did not include version metadata.", 1)
	}
	if response.ProjectID != 0 && response.ProjectID != projectID {
		return newSyncAppError(ErrorPushResponseInvalid, "push response projectId does not match local project", "", 1)
	}
	if response.VersionID <= 0 {
		return newSyncAppError(ErrorPushResponseInvalid, "push response is missing versionId", "Server response did not include a valid versionId.", 1)
	}
	if response.ProjectID == 0 {
		response.ProjectID = projectID
	}
	return nil
}

func validatePullResponse(response *projectapi.ProjectPullResponse, projectID int64) *command.AppError {
	if response == nil {
		return newSyncAppError(ErrorPullResponseInvalid, "pull response is empty", "Server response did not include encrypted environment data.", 1)
	}
	if response.ProjectID != 0 && response.ProjectID != projectID {
		return newSyncAppError(ErrorPullResponseInvalid, "pull response projectId does not match local project", "", 1)
	}
	if response.VersionID <= 0 {
		return newSyncAppError(ErrorPullResponseInvalid, "pull response is missing versionId", "Server response did not include a valid versionId.", 1)
	}
	if len(response.EncryptedEnvironment) == 0 {
		return newSyncAppError(ErrorPullResponseInvalid, "pull response is missing encryptedEnvironment", "Server response did not include encrypted environment data.", 1)
	}
	if response.ProjectID == 0 {
		response.ProjectID = projectID
	}
	return nil
}

func syncAppErrorFromAPI(defaultCode, defaultMessage string, err error) *command.AppError {
	var apiErr *api.ErrorResponse
	if !errors.As(err, &apiErr) {
		return newSyncAppError(defaultCode, defaultMessage, err.Error(), 1)
	}

	code := normalizeSyncAPIErrorCode(apiErr)
	if code == "" {
		code = defaultCode
	}
	message := strings.TrimSpace(apiErr.Message)
	if message == "" {
		message = defaultMessage
	}
	return newSyncAppError(code, message, syncHintForCode(code), exitCodeForSyncError(code))
}

func normalizeSyncAPIErrorCode(apiErr *api.ErrorResponse) string {
	if apiErr == nil {
		return ErrorPullEnvironmentFailed
	}
	switch strings.TrimSpace(apiErr.Code) {
	case backendCodeProjectNotFound:
		return ErrorProjectNotFound
	case backendCodeVersionNotInitialized:
		return ErrorEnvironmentVersionNotReady
	case backendCodeVersionConflict:
		return ErrorVersionConflict
	case backendCodeAccessDenied:
		return ErrorProjectAccessDenied
	}

	statusText := strings.TrimSpace(string(apiErr.Status))
	if fields := strings.Fields(statusText); len(fields) > 0 {
		statusText = fields[0]
	}
	status, err := strconv.Atoi(statusText)
	if err != nil {
		return strings.TrimSpace(apiErr.Code)
	}
	switch status {
	case http.StatusForbidden:
		return ErrorProjectAccessDenied
	case http.StatusNotFound:
		return ErrorProjectNotFound
	case http.StatusConflict:
		return ErrorVersionConflict
	case http.StatusUnprocessableEntity:
		return ErrorEnvironmentVersionNotReady
	default:
		return strings.TrimSpace(apiErr.Code)
	}
}

func syncHintForCode(code string) string {
	switch code {
	case ErrorProjectNotFound:
		return "Run `envio create` or `envio link` before syncing this repository."
	case ErrorProjectAccessDenied:
		return "Check that the current GitHub user is a member of this Envio project."
	case ErrorEnvironmentVersionNotReady:
		return "Run `envio push` to create the first environment version."
	case ErrorVersionConflict:
		return "Run `envio pull` before pushing again."
	default:
		return ""
	}
}

func exitCodeForSyncError(code string) int {
	if code == ErrorEnvironmentFileInvalid {
		return 2
	}
	return 1
}

func newSyncAppError(code, message, hint string, exitCode int) *command.AppError {
	return command.NewAppError(code, message, hint, exitCode, command.SeverityError)
}
