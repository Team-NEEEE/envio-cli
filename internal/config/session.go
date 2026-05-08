package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	appName           = "envio"
	globalSessionFile = "session.json"
)

// GlobalSession json으로 직렬화 해야하기 때문에 string 타입으로 설정
type GlobalSession struct {
	UserID     int64  `json:"userId"`
	GithubID   string `json:"githubId"`
	DeviceID   int64  `json:"deviceId"`
	DeviceName string `json:"deviceName"`
	PublicKey  string `json:"publicKey"`
}

type GlobalConfig struct {
	GlobalSession GlobalSession `json:"globalSession"`
}

func SaveGlobalSession(session GlobalSession) error {
	path, err := GlobalSessionPath()
	if err != nil {
		return err
	}

	cfg := GlobalConfig{GlobalSession: session}

	return writeJSONFile(path, cfg, 0600)
}

// GlobalSessionPath envio 저장 위치 설정
//
//	Windows: C:\Users\<사용자명>\AppData\Roaming\envio\session.json
//	macOS:   ~/Library/Application Support/envio/session.json
//	Linux:   ~/.config/envio/session.json
func GlobalSessionPath() (string, error) {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("사용자 config 디렉토리 확인 실패: %w", err)
	}

	return filepath.Join(userConfigDir, appName, globalSessionFile), nil
}

func writeJSONFile(path string, obj interface{}, perm os.FileMode) error {
	dir := filepath.Dir(path)

	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("디렉토리 생성 실패: %w", err)
	}

	raw, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return fmt.Errorf("객체 직렬화 실패: %w", err)
	}

	tempPath := path + ".tmp"

	if err := os.WriteFile(tempPath, raw, perm); err != nil {
		return fmt.Errorf("임시 config 파일 저장 실패: %w", err)
	}

	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("config 파일 교체 실패: %w", err)
	}

	return nil
}
