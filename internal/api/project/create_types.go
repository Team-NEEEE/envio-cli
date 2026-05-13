package projectapi

type CreateProjectRequest struct {
	RepositoryURL string `json:"repositoryUrl"`
}

type CreateProjectResponse struct {
	Message        string          `json:"message"`
	ProjectName    string          `json:"projectName"`
	GithubRepoName string          `json:"githubRepoName"`
	Members        []ProjectMember `json:"members"`
	ProjectID      int64           `json:"projectId"`
	InstallationID int64           `json:"installationId"`
}

type ProjectMember struct {
	GithubID     string `json:"githubId"`
	PublicKey    string `json:"publicKey"`
	ProjectRole  string `json:"projectRole"`
	UserID       int64  `json:"userId"`
	UserDeviceID int64  `json:"userDeviceId"`
}

type SaveWrappedKeysRequest struct {
	WrappedKeys []WrappedKey `json:"wrappedKeys"`
}

type WrappedKey struct {
	EncryptedKey string `json:"encryptedKey"`
	UserID       int64  `json:"userId"`
	UserDeviceID int64  `json:"userDeviceId"`
}

type SaveWrappedKeysResponse struct {
	Message      string `json:"message"`
	ProjectID    int64  `json:"projectId"`
	UpdatedCount int    `json:"updatedCount"`
}
