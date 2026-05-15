package projectapi

type LinkProjectRequest struct {
	PublicKey     string `json:"publicKey"`
	DeviceID      int64  `json:"deviceId"`
	UserGithubID  string `json:"userGithubId"`
	RepositoryURL string `json:"repositoryUrl"`
	Owner         string `json:"owner,omitempty"`
	RepoName      string `json:"repoName,omitempty"`
}

type LinkProjectResponse struct {
	Message          string        `json:"message"`
	Project          LinkedProject `json:"project"`
	WrappedMasterKey string        `json:"wrappedMasterKey"`
	JoinStatus       string        `json:"joinStatus"`
}

type LinkedProject struct {
	ProjectName    string `json:"projectName"`
	GithubRepoName string `json:"githubRepoName"`
	Owner          string `json:"owner"`
	ProjectID      int64  `json:"projectId"`
}
