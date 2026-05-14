package authapi

import "encoding/json"

type LoginStartResponse struct {
	Message        string `json:"message"`
	LoginSessionID string `json:"loginSessionId"`
	AuthURL        string `json:"loginUrl"`
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
	GithubID       string `json:"githubId"`
	PublicKey      string `json:"publicKey"`
	DeviceName     string `json:"deviceName"`
}

type RegisterKeyResponse struct {
	Message  string `json:"message"`
	GithubID string `json:"githubId"`
	Email    string `json:"email"`
	UserID   int64  `json:"userId"`
	DeviceID int64  `json:"deviceId"`
}

func (r *RegisterKeyResponse) UnmarshalJSON(data []byte) error {
	type registerKeyResponse RegisterKeyResponse
	var response registerKeyResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return err
	}

	var aliases struct {
		UserIDSnake   int64 `json:"user_id"`
		UserIDUpper   int64 `json:"userID"`
		DeviceIDSnake int64 `json:"device_id"`
		DeviceIDUpper int64 `json:"deviceID"`
	}
	if err := json.Unmarshal(data, &aliases); err != nil {
		return err
	}

	*r = RegisterKeyResponse(response)
	if r.UserID == 0 {
		switch {
		case aliases.UserIDSnake != 0:
			r.UserID = aliases.UserIDSnake
		case aliases.UserIDUpper != 0:
			r.UserID = aliases.UserIDUpper
		}
	}
	if r.DeviceID == 0 {
		switch {
		case aliases.DeviceIDSnake != 0:
			r.DeviceID = aliases.DeviceIDSnake
		case aliases.DeviceIDUpper != 0:
			r.DeviceID = aliases.DeviceIDUpper
		}
	}

	return nil
}
