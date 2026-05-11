package crypto

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"
)

// TestGenerateRSAKeyPairPEM RSA 키 쌍 생성 결과가 유효한 PEM인지 검증한다.
func TestGenerateRSAKeyPairPEM(t *testing.T) {
	privatePEM, publicPEM, err := GenerateRSAKeyPairPEM()
	if err != nil {
		t.Fatalf("GenerateRSAKeyPairPEM() error = %v", err)
	}

	if privatePEM == "" {
		t.Fatal("privatePEM is empty")
	}
	if publicPEM == "" {
		t.Fatal("publicPEM is empty")
	}

	// private PEM은 PKCS#1 RSA PRIVATE KEY 형식이어야 한다.
	privateBlock, rest := pem.Decode([]byte(privatePEM))
	if privateBlock == nil {
		t.Fatal("privatePEM is not valid PEM")
	}
	if len(bytes.TrimSpace(rest)) != 0 {
		t.Fatalf("privatePEM has trailing data: %q", string(rest))
	}
	if privateBlock.Type != "RSA PRIVATE KEY" {
		t.Fatalf("private block type = %q", privateBlock.Type)
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(privateBlock.Bytes)
	if err != nil {
		t.Fatalf("parse private key: %v", err)
	}
	if err := privateKey.Validate(); err != nil {
		t.Fatalf("validate private key: %v", err)
	}
	if got := privateKey.N.BitLen(); got != 2048 {
		t.Fatalf("private key bit length = %d", got)
	}

	// public PEM은 PKIX PUBLIC KEY 형식이어야 한다.
	publicBlock, rest := pem.Decode([]byte(publicPEM))
	if publicBlock == nil {
		t.Fatal("publicPEM is not valid PEM")
	}
	if len(bytes.TrimSpace(rest)) != 0 {
		t.Fatalf("publicPEM has trailing data: %q", string(rest))
	}
	if publicBlock.Type != "PUBLIC KEY" {
		t.Fatalf("public block type = %q", publicBlock.Type)
	}

	parsedPublicKey, err := x509.ParsePKIXPublicKey(publicBlock.Bytes)
	if err != nil {
		t.Fatalf("parse public key: %v", err)
	}

	publicKey, ok := parsedPublicKey.(*rsa.PublicKey)
	if !ok {
		t.Fatalf("public key type = %T", parsedPublicKey)
	}
	// 생성된 public key가 private key에서 파생된 값인지 확인한다.
	if publicKey.N.Cmp(privateKey.N) != 0 || publicKey.E != privateKey.E {
		t.Fatal("public key does not match private key")
	}
}
