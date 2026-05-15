package projectapi

type LinkProjectRequest struct {
	PublicKey     string `json:"publicKey"`
	UserGithubID  string `json:"userGithubId"`
	RepositoryURL string `json:"repositoryUrl"`
	Owner         string `json:"owner,omitempty"`
	RepoName      string `json:"repoName,omitempty"`
	DeviceID      int64  `json:"deviceId"`
}

type LinkProjectResponse struct {
	Message          string        `json:"message"`
	WrappedMasterKey string        `json:"wrappedMasterKey"`
	JoinStatus       string        `json:"joinStatus"`
	Project          LinkedProject `json:"project"`
}

type LinkedProject struct {
	ProjectName    string `json:"projectName"`
	GithubRepoName string `json:"githubRepoName"`
	Owner          string `json:"owner"`
	ProjectID      int64  `json:"projectId"`
}
