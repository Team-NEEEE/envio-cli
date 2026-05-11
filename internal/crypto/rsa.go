package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

// GenerateRSAKeyPairPEM generates a 2048-bit RSA key pair and returns both keys as PEM strings.
func GenerateRSAKeyPairPEM() (privatePEM string, publicPEM string, err error) {
	// RSA 방식으로 private key를 생성한다.
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", fmt.Errorf("RSA private key 생성 실패: %w", err)
	}

	// Private key를 PKCS#1 DER로 직렬화한 뒤 PEM으로 감싼다.
	privateDER := x509.MarshalPKCS1PrivateKey(privateKey)
	privateBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateDER,
	}

	// Private key에서 public key를 추출해 PKIX DER로 직렬화한다.
	publicDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return "", "", fmt.Errorf("RSA public key 직렬화 실패: %w", err)
	}

	publicBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicDER,
	}

	return string(pem.EncodeToMemory(privateBlock)),
		string(pem.EncodeToMemory(publicBlock)),
		nil
}
