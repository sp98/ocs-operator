package server

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"time"

	"github.com/golang-jwt/jwt"
)

type AuthManager struct {
	secretKey     string
	tokenDuration time.Duration
}

type Claims struct {
	jwt.StandardClaims
}

func NewAuthManager(secretKey string, duration time.Duration) *AuthManager {
	return &AuthManager{secretKey, duration}
}

// GetAccessToken generates a new JWT token
func (a *AuthManager) GetAccessToken() (string, error) {
	claims := Claims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(a.tokenDuration).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(a.secretKey))
}

// VerifyAccessToken verifies the jwt token and return a claim if the token is valid
func (a *AuthManager) VerifyAccessToken(accessToken string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		accessToken,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			_, ok := token.Method.(*jwt.SigningMethodHMAC)
			if !ok {
				return nil, fmt.Errorf("unexpected token signing method")
			}

			return []byte(a.secretKey), nil
		},
	)

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

// GetEncryptedID encryptes the StorageConsumer UID
func (a *AuthManager) GetEncryptedID(id string) ([]byte, error) {
	hashedID := sha256.Sum256([]byte(id))
	return encrypt(hashedID, a.secretKey)
}

// VerifyEncryptedID verifies the encrypted StorageConsumer UID
func (a *AuthManager) VerifyEncryptedID(id string) ([]byte, error) {
	return decrypt([]byte(id), a.secretKey)
}

func encrypt(s [32]byte, key string) ([]byte, error) {
	c, err := aes.NewCipher([]byte(key))
	if err != nil {
		return []byte{}, fmt.Errorf("failed to create new cipher. %v", err)
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to generate gcm. %v", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return []byte{}, fmt.Errorf("failed generate nonce with secure random sequence. %v", err)
	}

	return gcm.Seal(nonce, nonce, s[:], nil), nil
}

func decrypt(s []byte, key string) ([]byte, error) {
	c, err := aes.NewCipher([]byte(key))
	if err != nil {
		return []byte{}, fmt.Errorf("failed to create new cipher. %v", err)
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to generate gcm. %v", err)
	}

	nonceSize := gcm.NonceSize()
	if len(s) < nonceSize {
		// TODO: return new error here
		return []byte{}, fmt.Errorf("invalid nonce size. %v", err)
	}

	nonce, ciphertext := s[:nonceSize], s[nonceSize:]
	res, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to decrypt data. %v", err)
	}

	return res, err
}
