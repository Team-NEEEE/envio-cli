package project

import (
	"context"
	"strconv"
	"strings"

	projectapi "github.com/Team-NEEEE/envio-cli/internal/api/project"
	"github.com/Team-NEEEE/envio-cli/internal/command"
)

func (s *SyncService) ListHistory(
	ctx context.Context,
	cwd string,
	reporter command.Reporter,
) (*HistoryListResult, *command.AppError) {
	if reporter == nil {
		reporter = command.NoopReporter{}
	}
	if s == nil {
		return nil, newSyncAppError(ErrorHistoryFailed, "history service is not configured", "", 1)
	}
	s.ensureDefaults()

	local, histories, appErr := s.loadHistory(ctx, cwd, reporter)
	if appErr != nil {
		return nil, appErr
	}

	return &HistoryListResult{
		ProjectID:       local.projectID,
		LocalRepository: local.repositoryRoot,
		Histories:       histories,
	}, nil
}

func (s *SyncService) DecryptHistoryVersion(
	ctx context.Context,
	cwd string,
	versionSelector string,
	reporter command.Reporter,
) (*HistoryVersionResult, *command.AppError) {
	if reporter == nil {
		reporter = command.NoopReporter{}
	}
	if s == nil {
		return nil, newSyncAppError(ErrorHistoryFailed, "history service is not configured", "", 1)
	}
	s.ensureDefaults()

	local, histories, appErr := s.loadHistory(ctx, cwd, reporter)
	if appErr != nil {
		return nil, appErr
	}

	entry, ok := SelectHistoryVersion(histories, versionSelector)
	if !ok {
		return nil, newSyncAppError(
			ErrorHistoryVersionNotFound,
			"history version was not found",
			"Choose a version from `envio history`.",
			1,
		)
	}
	if len(entry.EncryptedEnvironment) == 0 {
		return nil, newSyncAppError(
			ErrorHistoryResponseInvalid,
			"history response is missing encryptedEnvironment",
			"Server response did not include encrypted environment data for the selected version.",
			1,
		)
	}

	reporter.UpdateStep(command.StepUpdate{ID: StepDecryptEnvironment, Status: command.StatusRunning})
	raw, err := s.decryptEnvironment(entry.EncryptedEnvironment, local.masterKey)
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

	return &HistoryVersionResult{
		ProjectID:       entry.ProjectID,
		VersionID:       entry.VersionID,
		BaseVersionID:   entry.BaseVersionID,
		HistoryID:       entry.HistoryID,
		GithubID:        entry.GithubID,
		CreatedAt:       entry.CreatedAt,
		Environment:     string(raw),
		VariableCount:   len(values),
		LocalRepository: local.repositoryRoot,
	}, nil
}

func (s *SyncService) loadHistory(
	ctx context.Context,
	cwd string,
	reporter command.Reporter,
) (localProjectContext, []HistoryEntry, *command.AppError) {
	reporter.UpdateStep(command.StepUpdate{ID: StepLoadProjectContext, Status: command.StatusRunning})
	local, appErr := s.loadLocalContext(ctx, cwd)
	if appErr != nil {
		reporter.UpdateStep(command.StepUpdate{ID: StepLoadProjectContext, Status: command.StatusError})
		return localProjectContext{}, nil, appErr
	}
	reporter.UpdateStep(command.StepUpdate{ID: StepLoadProjectContext, Status: command.StatusSuccess})

	reporter.UpdateStep(command.StepUpdate{ID: StepListHistory, Status: command.StatusRunning})
	if appErr := s.ensureClientReady(ErrorHistoryFailed); appErr != nil {
		reporter.UpdateStep(command.StepUpdate{ID: StepListHistory, Status: command.StatusError})
		return localProjectContext{}, nil, appErr
	}
	response, err := s.client.History(ctx, local.projectID, "")
	if err != nil {
		reporter.UpdateStep(command.StepUpdate{ID: StepListHistory, Status: command.StatusError})
		return localProjectContext{}, nil, syncAppErrorFromAPI(ErrorHistoryFailed, "history request failed", err)
	}
	if appErr := validateHistoryResponse(response, local.projectID); appErr != nil {
		reporter.UpdateStep(command.StepUpdate{ID: StepListHistory, Status: command.StatusError})
		return localProjectContext{}, nil, appErr
	}
	reporter.UpdateStep(command.StepUpdate{ID: StepListHistory, Status: command.StatusSuccess})

	return local, normalizeHistoryEntries(response.Histories, local.projectID), nil
}

func SelectHistoryVersion(histories []HistoryEntry, selector string) (HistoryEntry, bool) {
	selector = normalizeHistoryVersionSelector(selector)
	if selector == "" {
		return HistoryEntry{}, false
	}
	for _, history := range histories {
		if strconv.FormatInt(history.VersionID, 10) == selector {
			return history, true
		}
	}
	return HistoryEntry{}, false
}

func validateHistoryResponse(response *projectapi.ProjectHistoryResponse, projectID int64) *command.AppError {
	if response == nil {
		return newSyncAppError(ErrorHistoryResponseInvalid, "history response is empty", "Server response did not include history data.", 1)
	}
	if len(response.Histories) == 0 {
		return newSyncAppError(ErrorEnvironmentVersionNotReady, "history response has no versions", "Run `envio push` to create the first environment version.", 1)
	}
	for _, history := range response.Histories {
		if history.ProjectID != 0 && history.ProjectID != projectID {
			return newSyncAppError(ErrorHistoryResponseInvalid, "history response projectId does not match local project", "", 1)
		}
		if history.HistoryID <= 0 {
			return newSyncAppError(ErrorHistoryResponseInvalid, "history response is missing historyId", "Server response did not include a valid historyId.", 1)
		}
		if history.VersionID <= 0 {
			return newSyncAppError(ErrorHistoryResponseInvalid, "history response is missing versionId", "Server response did not include a valid versionId.", 1)
		}
	}
	return nil
}

func normalizeHistoryEntries(entries []projectapi.ProjectHistoryEntry, projectID int64) []HistoryEntry {
	histories := make([]HistoryEntry, 0, len(entries))
	hasLatest := false
	var maxVersionID int64
	for _, entry := range entries {
		if entry.ProjectID == 0 {
			entry.ProjectID = projectID
		}
		if entry.Latest {
			hasLatest = true
		}
		if entry.VersionID > maxVersionID {
			maxVersionID = entry.VersionID
		}
		histories = append(histories, HistoryEntry{
			EncryptedEnvironment: entry.EncryptedEnvironment,
			GithubID:             entry.GithubID,
			CreatedAt:            entry.CreatedAt,
			HistoryID:            entry.HistoryID,
			ProjectID:            entry.ProjectID,
			VersionID:            entry.VersionID,
			BaseVersionID:        entry.BaseVersionID,
			Latest:               entry.Latest,
		})
	}
	if hasLatest {
		return histories
	}
	for i := range histories {
		if histories[i].VersionID == maxVersionID {
			histories[i].Latest = true
			break
		}
	}
	return histories
}

func normalizeHistoryVersionSelector(selector string) string {
	selector = strings.ToLower(strings.TrimSpace(selector))
	selector = strings.TrimPrefix(selector, "version:")
	selector = strings.TrimPrefix(selector, "version-")
	selector = strings.TrimPrefix(selector, "v")
	selector = strings.TrimSpace(selector)
	if selector == "" {
		return ""
	}
	if _, err := strconv.ParseInt(selector, 10, 64); err != nil {
		return "invalid:" + selector
	}
	return selector
}
