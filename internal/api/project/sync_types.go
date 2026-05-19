package projectapi

type ProjectPushRequest struct {
	EncryptedEnvironment map[string]any `json:"encryptedEnvironment"`
	GithubUserID         string         `json:"githubUserId"`
	ParentVersionID      int64          `json:"parentVersionId"`
}

type ProjectPushResponse struct {
	Message         string `json:"message"`
	EnvName         string `json:"envName"`
	HistoryID       int64  `json:"historyId"`
	ProjectID       int64  `json:"projectId"`
	VersionID       int64  `json:"versionId"`
	ParentVersionID int64  `json:"parentVersionId"`
}

type ProjectPullResponse struct {
	EncryptedEnvironment map[string]any `json:"encryptedEnvironment"`
	Message              string         `json:"message"`
	EnvName              string         `json:"envName"`
	WrappedMasterKey     string         `json:"wrappedMasterKey"`
	CreatedAt            string         `json:"createdAt,omitempty"`
	UpdatedAt            string         `json:"updatedAt,omitempty"`
	HistoryID            int64          `json:"historyId"`
	ProjectID            int64          `json:"projectId"`
	VersionID            int64          `json:"versionId"`
}
