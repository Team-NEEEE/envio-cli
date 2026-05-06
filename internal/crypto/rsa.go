package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

func GenerateRSAKeyPairPEM() (privatePEM string, publicPEM string, err error) {
	//private key 생성 - rsa 방식
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", fmt.Errorf("RSA private key 생성 실패: %w", err)
	}

	//private key 직렬화
	privateDER := x509.MarshalPKCS1PrivateKey(privateKey)
	privateBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateDER,
	}

	//public key 직렬화 - privateKey 기준
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
