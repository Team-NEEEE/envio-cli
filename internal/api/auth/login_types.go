package authapi

type LoginStartResponse struct {
	Message        string `json:"message"`
	LoginSessionID string `json:"loginSessionId"`
	AuthURL        string `json:"authUrl"`
	ExpiresIn      int    `json:"expiresIn"`
}

type LoginStatusResponse struct {
	Message  string `json:"message"`
	Status   string `json:"status"`
	GithubID string `json:"githubId"`
	Email    string `json:"email"`
}

type LoginStatusRequest struct {
	LoginSessionID string `json:"loginSessionId"`
}

type RegisterKeyRequest struct {
	LoginSessionID string `json:"loginSessionId"`
	PublicKey      string `json:"publicKey"`
	DeviceName     string `json:"deviceName"`
}

type RegisterKeyResponse struct {
	Message  string `json:"message"`
	UserID   int64  `json:"userId"`
	GithubID string `json:"githubId"`
	Email    string `json:"email"`
	DeviceID int64  `json:"deviceId"`
}
