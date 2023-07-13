package domain

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"time"
)

var (
	// ErrInvalidToken returns when requested token is not found in store.
	ErrInvalidToken = errors.New("invalid user token")
	ErrNoTokenFound = errors.New("token not found")
)

// WrapToken wraps token for user representation.
type WrapToken struct {
	Token Token `json:"token"`
}

// Token represents token model.
type Token struct {
	Hashed []byte    `json:"-"`
	Plain  string    `json:"plain_token" db:"-"`
	UserID int       `json:"-" db:"user_id"`
	Expiry time.Time `json:"expiry"`
}

// TokenService represents a service for managing tokens.
type TokenService interface {
	Create(ctx context.Context, token Token) error
	GetUserID(ctx context.Context, hashedToken string) (int, error)
}

// GenerateToken returns generated token.
func GenerateToken(id int, length int, expiry time.Duration) (Token, error) {
	randBytes := make([]byte, length)
	_, err := rand.Read(randBytes)
	if err != nil {
		return Token{}, err
	}

	plainToken := base64.RawURLEncoding.EncodeToString(randBytes)
	hashedToken := HashToken(plainToken)

	expiryTime := time.Now().Add(expiry)

	token := Token{
		Hashed: hashedToken,
		Plain:  plainToken,
		UserID: id,
		Expiry: expiryTime,
	}

	return token, nil
}

// HashToken hashes token.
func HashToken(s string) []byte {
	b := sha256.Sum256([]byte(s))
	return b[:]
}
