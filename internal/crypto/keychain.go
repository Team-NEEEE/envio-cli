package crypto

import "github.com/zalando/go-keyring"

const serviceName = "envio-cli"

// SavePrivateKey 개인키 저장
func SavePrivateKey(privateKeyPEM string) error {
	return keyring.Set(serviceName, "device-private-key", privateKeyPEM)
}

// SavePublicKey 공개키 저장
func SavePublicKey(publicKeyPEM string) error {
	return keyring.Set(serviceName, "device-public-key", publicKeyPEM)
}

// LoadPrivateKey 개인키 로드
func LoadPrivateKey() (string, error) {
	return keyring.Get(serviceName, "device-private-key")
}
