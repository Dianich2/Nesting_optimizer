package jwt

import (
	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

type Claims struct {
	UserID    int64     `json:"user_id"`
	SessionID uuid.UUID `json:"session_id"`
	TokenType string    `json:"token_type"`
	jwtlib.RegisteredClaims
}
