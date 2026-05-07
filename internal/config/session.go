package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Session TODO key type 고려
type Session struct {
	UserID     int64  `json:"userId"`
	GithubID   string `json:"githubId"`
	DeviceID   int64  `json:"deviceId"`
	DeviceName string `json:"deviceName"`
	PublicKey  string `json:"publicKey"`
	PrivateKey string `json:"privateKey"`
}

type LocalConfig struct {
	Session Session `json:"session"`
}

func SaveLocalSession(session Session) error {
	dir := ".envio"

	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf(".envio 디렉토리 생성 실패: %w", err)
	}

	cfg := LocalConfig{
		Session: session,
	}

	raw, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("config 직렬화 실패: %w", err)
	}

	path := filepath.Join(dir, "config")

	if err := os.WriteFile(path, raw, 0600); err != nil {
		return fmt.Errorf(".envio/config 저장 실패: %w", err)
	}

	return nil
}
