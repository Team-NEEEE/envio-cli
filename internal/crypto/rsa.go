package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"encoding/pem"
	"fmt"
	"math/big"
)

// GenerateRSAKeyPairPEM generates a 2048-bit RSA key pair and returns both keys as PEM strings.
func GenerateRSAKeyPairPEM() (privatePEM string, publicPEM string, err error) {
	privateKey, err := generateRSAKey()
	if err != nil {
		return "", "", err
	}

	publicPEM, err = encodeRSAPublicKeyPEM(&privateKey.PublicKey)
	if err != nil {
		return "", "", err
	}

	return encodeRSAPrivateKeyPEM(privateKey), publicPEM, nil
}

// GenerateRSAKeyPairForLogin returns a private PEM and an SSH authorized_keys public key.
func GenerateRSAKeyPairForLogin() (privatePEM string, publicKey string, err error) {
	privateKey, err := generateRSAKey()
	if err != nil {
		return "", "", err
	}

	publicKey, err = encodeRSAPublicKeySSH(&privateKey.PublicKey)
	if err != nil {
		return "", "", err
	}

	return encodeRSAPrivateKeyPEM(privateKey), publicKey, nil
}

func generateRSAKey() (*rsa.PrivateKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("RSA private key 생성 실패: %w", err)
	}
	return privateKey, nil
}

func encodeRSAPrivateKeyPEM(privateKey *rsa.PrivateKey) string {
	privateDER := x509.MarshalPKCS1PrivateKey(privateKey)
	privateBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateDER,
	}
	return string(pem.EncodeToMemory(privateBlock))
}

func encodeRSAPublicKeyPEM(publicKey *rsa.PublicKey) (string, error) {
	publicDER, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return "", fmt.Errorf("RSA public key 직렬화 실패: %w", err)
	}

	publicBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicDER,
	}

	return string(pem.EncodeToMemory(publicBlock)), nil
}

func encodeRSAPublicKeySSH(publicKey *rsa.PublicKey) (string, error) {
	if publicKey == nil {
		return "", fmt.Errorf("RSA public key is nil")
	}

	blob := make([]byte, 0, 3+4+3+4+256)
	var err error
	if blob, err = appendSSHString(blob, []byte("ssh-rsa")); err != nil {
		return "", err
	}
	if blob, err = appendSSHString(blob, marshalSSHMPInt(big.NewInt(int64(publicKey.E)))); err != nil {
		return "", err
	}
	if blob, err = appendSSHString(blob, marshalSSHMPInt(publicKey.N)); err != nil {
		return "", err
	}

	return "ssh-rsa " + base64.StdEncoding.EncodeToString(blob), nil
}

func appendSSHString(dst []byte, value []byte) ([]byte, error) {
	lengthValue, err := uint32Length(value)
	if err != nil {
		return nil, err
	}

	var length [4]byte
	binary.BigEndian.PutUint32(length[:], lengthValue)
	dst = append(dst, length[:]...)
	return append(dst, value...), nil
}

func uint32Length(value []byte) (uint32, error) {
	valueLength := len(value)
	if uint64(valueLength) > uint64(^uint32(0)) {
		return 0, fmt.Errorf("SSH string length exceeds uint32: %d", valueLength)
	}
	return uint32(valueLength), nil
}

func marshalSSHMPInt(value *big.Int) []byte {
	if value == nil || value.Sign() == 0 {
		return nil
	}

	encoded := value.Bytes()
	if encoded[0]&0x80 != 0 {
		encoded = append([]byte{0}, encoded...)
	}
	return encoded
}
