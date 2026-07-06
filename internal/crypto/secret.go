package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
)

// TODO: M2 管理端接入密钥保存时，统一通过本文件加密 api_key 和数据库密码，未配置 CONFIG_ENCRYPTION_KEY 时拒绝保存 secret。
func EncryptString(key string, plaintext string) (string, error) {
	if key == "" {
		return "", fmt.Errorf("CONFIG_ENCRYPTION_KEY is required")
	}
	block, err := aes.NewCipher(normalizeKey(key))
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func DecryptString(key string, encoded string) (string, error) {
	if key == "" {
		return "", fmt.Errorf("CONFIG_ENCRYPTION_KEY is required")
	}
	ciphertext, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(normalizeKey(key))
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(ciphertext) < gcm.NonceSize() {
		return "", fmt.Errorf("ciphertext is too short")
	}
	nonce, body := ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, body, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

func normalizeKey(key string) []byte {
	sum := sha256.Sum256([]byte(key))
	return sum[:]
}
