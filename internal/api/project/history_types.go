package projectapi

import "encoding/json"

type ProjectHistoryResponse struct {
	Message   string                `json:"message"`
	Histories []ProjectHistoryEntry `json:"histories"`
}

func (r *ProjectHistoryResponse) UnmarshalJSON(data []byte) error {
	var histories []ProjectHistoryEntry
	if err := json.Unmarshal(data, &histories); err == nil {
		r.Message = ""
		r.Histories = histories
		return nil
	}

	type projectHistoryResponse ProjectHistoryResponse
	var response projectHistoryResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return err
	}
	*r = ProjectHistoryResponse(response)
	return nil
}

type ProjectHistoryEntry struct {
	EncryptedEnvironment map[string]any `json:"encryptedEnvironment,omitempty"`
	Message              string         `json:"message,omitempty"`
	GithubID             string         `json:"githubId"`
	CreatedAt            string         `json:"createdAt"`
	UpdatedAt            string         `json:"updatedAt,omitempty"`
	HistoryID            int64          `json:"historyId"`
	ProjectID            int64          `json:"projectId"`
	VersionID            int64          `json:"versionId"`
	BaseVersionID        int64          `json:"baseVersionId,omitempty"`
	Latest               bool           `json:"latest,omitempty"`
}

func (h *ProjectHistoryEntry) UnmarshalJSON(data []byte) error {
	type projectHistoryEntry ProjectHistoryEntry
	var entry projectHistoryEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return err
	}

	var aliases struct {
		EncryptedEnvironment map[string]any `json:"encrypted_environment"`
		GithubID             string         `json:"github_id"`
		CreatedAt            string         `json:"created_at"`
		UpdatedAt            string         `json:"updated_at"`
		HistoryID            int64          `json:"history_id"`
		HistoriesID          int64          `json:"histories_id"`
		ProjectID            int64          `json:"project_id"`
		VersionID            int64          `json:"version_id"`
		BaseVersionID        int64          `json:"base_version_id"`
		IsLatest             bool           `json:"is_latest"`
	}
	if err := json.Unmarshal(data, &aliases); err != nil {
		return err
	}

	*h = ProjectHistoryEntry(entry)
	if len(h.EncryptedEnvironment) == 0 && len(aliases.EncryptedEnvironment) > 0 {
		h.EncryptedEnvironment = aliases.EncryptedEnvironment
	}
	if h.GithubID == "" {
		h.GithubID = aliases.GithubID
	}
	if h.CreatedAt == "" {
		h.CreatedAt = aliases.CreatedAt
	}
	if h.UpdatedAt == "" {
		h.UpdatedAt = aliases.UpdatedAt
	}
	if h.HistoryID == 0 {
		h.HistoryID = aliases.HistoryID
	}
	if h.HistoryID == 0 {
		h.HistoryID = aliases.HistoriesID
	}
	if h.ProjectID == 0 {
		h.ProjectID = aliases.ProjectID
	}
	if h.VersionID == 0 {
		h.VersionID = aliases.VersionID
	}
	if h.BaseVersionID == 0 {
		h.BaseVersionID = aliases.BaseVersionID
	}
	if !h.Latest {
		h.Latest = aliases.IsLatest
	}

	return nil
}
