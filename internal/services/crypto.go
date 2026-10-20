package services

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	mathrand "math/rand"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func HashPasswordMD5(password string) string {
	hash := md5.Sum([]byte(password))
	return hex.EncodeToString(hash[:])
}

func HashPasswordSHA1(password string) string {
	hash := sha1.Sum([]byte(password))
	return hex.EncodeToString(hash[:])
}

func VerifyMD5Password(password, hash string) bool {
	return HashPasswordMD5(password) == hash
}

func GenerateRandomToken() string {
	mathrand.Seed(time.Now().UnixNano())
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	token := make([]byte, 32)
	for i := range token {
		token[i] = chars[mathrand.Intn(len(chars))]
	}
	return string(token)
}

func GenerateSessionID() string {
	mathrand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%d-%d", time.Now().Unix(), mathrand.Int63())
}

type CryptoService struct {
	key []byte
}

func NewCryptoService(key string) *CryptoService {
	return &CryptoService{key: []byte(key)}
}

func (c *CryptoService) GetKey() []byte {
	return c.key
}

func (c *CryptoService) Encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

func (c *CryptoService) Decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	return gcm.Open(nil, nonce, ciphertext, nil)
}

type KeyChain struct {
	keys map[string][]byte
}

func NewKeyChain() *KeyChain {
	return &KeyChain{keys: make(map[string][]byte)}
}

func (kc *KeyChain) AddKey(name string, key []byte) {
	kc.keys[name] = key
}

func (kc *KeyChain) GetKey(name string) []byte {
	return kc.keys[name]
}

func (kc *KeyChain) ExportKeys() map[string][]byte {
	return kc.keys
}

func HashPasswordBcrypt(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func VerifyPasswordBcrypt(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func GenerateRandomBytes(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func NewSessionID() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func DeriveKey(password, salt []byte) []byte {
	combined := append(password, salt...)
	hash := sha1.Sum(combined)
	return hash[:]
}
