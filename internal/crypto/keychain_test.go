package crypto

import (
	"errors"
	"testing"

	"github.com/zalando/go-keyring"
)

// TestSaveAndLoadPrivateKey 개인키 저장 후 로드 동작을 검증한다.
func TestSaveAndLoadPrivateKey(t *testing.T) {
	// 실제 OS keychain을 건드리지 않도록 go-keyring mock provider를 사용한다.
	keyring.MockInit()

	const privateKeyPEM = "private-key-pem"

	if err := SavePrivateKey(privateKeyPEM); err != nil {
		t.Fatalf("SavePrivateKey() error = %v", err)
	}

	got, err := LoadPrivateKey()
	if err != nil {
		t.Fatalf("LoadPrivateKey() error = %v", err)
	}
	if got != privateKeyPEM {
		t.Fatalf("LoadPrivateKey() = %q", got)
	}
}

func TestSaveAndLoadDevicePrivateKey(t *testing.T) {
	keyring.MockInit()

	const privateKeyPEM = "private-key-pem"
	if err := SaveDevicePrivateKey(20, privateKeyPEM); err != nil {
		t.Fatalf("SaveDevicePrivateKey() error = %v", err)
	}

	got, err := LoadDevicePrivateKey(20)
	if err != nil {
		t.Fatalf("LoadDevicePrivateKey() error = %v", err)
	}
	if got != privateKeyPEM {
		t.Fatalf("LoadDevicePrivateKey() = %q", got)
	}
}

// TestSavePublicKey 공개키 저장 시 keychain service와 key 이름을 검증한다.
func TestSavePublicKey(t *testing.T) {
	// 실제 OS keychain을 건드리지 않도록 go-keyring mock provider를 사용한다.
	keyring.MockInit()

	const publicKeyPEM = "public-key-pem"

	if err := SavePublicKey(publicKeyPEM); err != nil {
		t.Fatalf("SavePublicKey() error = %v", err)
	}

	got, err := keyring.Get(serviceName, "device-public-key")
	if err != nil {
		t.Fatalf("keyring.Get() error = %v", err)
	}
	if got != publicKeyPEM {
		t.Fatalf("saved public key = %q", got)
	}
}

func TestDeleteProjectMasterKey(t *testing.T) {
	keyring.MockInit()

	if err := keyring.Set(serviceName, projectMasterKeyAccount(1, 20), "legacy-master-key"); err != nil {
		t.Fatalf("keyring.Set() error = %v", err)
	}
	if err := DeleteProjectMasterKey(1, 20); err != nil {
		t.Fatalf("DeleteProjectMasterKey() error = %v", err)
	}
	if _, err := keyring.Get(serviceName, projectMasterKeyAccount(1, 20)); !errors.Is(err, keyring.ErrNotFound) {
		t.Fatalf("keyring.Get() error = %v, want ErrNotFound", err)
	}
	if err := DeleteProjectMasterKey(1, 20); err != nil {
		t.Fatalf("DeleteProjectMasterKey() missing key error = %v", err)
	}
}

// TestKeychainErrors keychain 오류가 호출자에게 그대로 전달되는지 검증한다.
func TestKeychainErrors(t *testing.T) {
	wantErr := errors.New("keyring failed")
	// mock provider가 항상 오류를 반환하도록 설정한다.
	keyring.MockInitWithError(wantErr)

	if err := SavePrivateKey("private-key-pem"); !errors.Is(err, wantErr) {
		t.Fatalf("SavePrivateKey() error = %v, want %v", err, wantErr)
	}
	if err := SaveDevicePrivateKey(20, "private-key-pem"); !errors.Is(err, wantErr) {
		t.Fatalf("SaveDevicePrivateKey() error = %v, want %v", err, wantErr)
	}
	if err := SavePublicKey("public-key-pem"); !errors.Is(err, wantErr) {
		t.Fatalf("SavePublicKey() error = %v, want %v", err, wantErr)
	}
	if _, err := LoadPrivateKey(); !errors.Is(err, wantErr) {
		t.Fatalf("LoadPrivateKey() error = %v, want %v", err, wantErr)
	}
	if _, err := LoadDevicePrivateKey(20); !errors.Is(err, wantErr) {
		t.Fatalf("LoadDevicePrivateKey() error = %v, want %v", err, wantErr)
	}
	if err := DeleteProjectMasterKey(1, 20); !errors.Is(err, wantErr) {
		t.Fatalf("DeleteProjectMasterKey() error = %v, want %v", err, wantErr)
	}
}
