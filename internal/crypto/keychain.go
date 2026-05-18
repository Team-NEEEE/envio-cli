package crypto

import (
	"errors"
	"fmt"

	"github.com/zalando/go-keyring"
)

const (
	serviceName       = "envio"
	legacyServiceName = "envio-cli"
)

// SavePrivateKey 개인키 저장
func SavePrivateKey(privateKeyPEM string) error {
	return keyring.Set(serviceName, "device-private-key", privateKeyPEM)
}

func SaveDevicePrivateKey(deviceID int64, privateKeyPEM string) error {
	if deviceID <= 0 {
		return fmt.Errorf("deviceId must be positive")
	}
	return keyring.Set(serviceName, devicePrivateKeyAccount(deviceID), privateKeyPEM)
}

// SavePublicKey 공개키 저장
func SavePublicKey(publicKeyPEM string) error {
	return keyring.Set(serviceName, "device-public-key", publicKeyPEM)
}

// LoadPrivateKey 개인키 로드
func LoadPrivateKey() (string, error) {
	return keyring.Get(serviceName, "device-private-key")
}

func LoadDevicePrivateKey(deviceID int64) (string, error) {
	if deviceID <= 0 {
		return "", fmt.Errorf("deviceId must be positive")
	}
	privateKeyPEM, err := keyring.Get(serviceName, devicePrivateKeyAccount(deviceID))
	if err == nil {
		return privateKeyPEM, nil
	}
	if legacyPrivateKeyPEM, legacyErr := LoadPrivateKey(); legacyErr == nil {
		return legacyPrivateKeyPEM, nil
	}
	if legacyPrivateKeyPEM, legacyErr := keyring.Get(legacyServiceName, "device-private-key"); legacyErr == nil {
		return legacyPrivateKeyPEM, nil
	}
	return "", err
}

func DeleteProjectMasterKey(projectID int64, deviceID int64) error {
	if projectID <= 0 {
		return fmt.Errorf("projectId must be positive")
	}
	if deviceID <= 0 {
		return fmt.Errorf("deviceId must be positive")
	}
	err := keyring.Delete(serviceName, projectMasterKeyAccount(projectID, deviceID))
	if errors.Is(err, keyring.ErrNotFound) {
		return nil
	}
	return err
}

func devicePrivateKeyAccount(deviceID int64) string {
	return fmt.Sprintf("private-key:%d", deviceID)
}

func projectMasterKeyAccount(projectID int64, deviceID int64) string {
	return fmt.Sprintf("project-master-key:%d:%d", projectID, deviceID)
}
