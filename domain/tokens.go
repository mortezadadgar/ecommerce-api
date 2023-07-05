package domain

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"time"
)

// WrapToken wraps token for user representation.
type WrapToken struct {
	Token Tokens `json:"token"`
}

// Tokens represents tokens model.
type Tokens struct {
	Hashed []byte    `json:"-"`
	Plain  string    `json:"plain_token" db:"-"`
	UserID int       `json:"-" db:"user_id"`
	Expiry time.Time `json:"expiry"`
}

// TokensService represents a service for managing tokens.
type TokensService interface {
	Create(ctx context.Context, token Tokens) error
	GetUser(ctx context.Context, hashedToken string) (Users, error)
}

// GenerateToken returns generated token.
func GenerateToken(id int, length int, expiry time.Duration) (Tokens, error) {
	randBytes := make([]byte, length)
	_, err := rand.Read(randBytes)
	if err != nil {
		return Tokens{}, err
	}

	plainToken := base64.RawURLEncoding.EncodeToString(randBytes)
	hashedToken := HashToken(plainToken)

	expiryTime := time.Now().Add(expiry)

	token := Tokens{
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
