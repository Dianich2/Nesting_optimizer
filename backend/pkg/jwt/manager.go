package jwt

import (
	"errors"
	"fmt"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Manager struct {
	accessSecret  []byte
	refreshSecret []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration
	issuer        string
}

func NewManager(
	accessSecret string,
	refreshSecret string,
	accessTTL time.Duration,
	refreshTTL time.Duration,
	issuer string,
) *Manager {
	return &Manager{
		accessSecret:  []byte(accessSecret),
		refreshSecret: []byte(refreshSecret),
		accessTTL:     accessTTL,
		refreshTTL:    refreshTTL,
		issuer:        issuer,
	}
}

func (m *Manager) GenerateAccessToken(
	userID int64,
	sessionID uuid.UUID,
) (string, error) {
	return m.generateToken(
		userID,
		sessionID,
		TokenTypeAccess,
		m.accessSecret,
		m.accessTTL,
	)
}

func (m *Manager) GenerateRefreshToken(
	userID int64,
	sessionID uuid.UUID,
) (string, error) {
	return m.generateToken(
		userID,
		sessionID,
		TokenTypeRefresh,
		m.refreshSecret,
		m.refreshTTL,
	)
}

func (m *Manager) ParseAccessToken(
	tokenString string,
) (*Claims, error) {
	return m.parseToken(
		tokenString,
		m.accessSecret,
		TokenTypeAccess,
	)
}

func (m *Manager) ParseRefreshToken(
	tokenString string,
) (*Claims, error) {
	return m.parseToken(
		tokenString,
		m.refreshSecret,
		TokenTypeRefresh,
	)
}

func (m *Manager) generateToken(
	userID int64,
	sessionID uuid.UUID,
	tokenType string,
	secret []byte,
	ttl time.Duration,
) (string, error) {
	if userID <= 0 {
		return "", fmt.Errorf(
			"generate %s token: user id must be positive",
			tokenType,
		)
	}

	if sessionID == uuid.Nil {
		return "", fmt.Errorf(
			"generate %s token: session id must not be empty",
			tokenType,
		)
	}

	now := time.Now()

	claims := Claims{
		UserID:    userID,
		SessionID: sessionID,
		TokenType: tokenType,
		RegisteredClaims: jwtlib.RegisteredClaims{
			ExpiresAt: jwtlib.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwtlib.NewNumericDate(now),
			Issuer:    m.issuer,
		},
	}

	token := jwtlib.NewWithClaims(
		jwtlib.SigningMethodHS256,
		claims,
	)

	signedToken, err := token.SignedString(
		secret,
	)
	if err != nil {
		return "", fmt.Errorf(
			"sign %s token: %w",
			tokenType,
			err,
		)
	}

	return signedToken, nil
}

func (m *Manager) parseToken(
	tokenString string,
	secret []byte,
	expectedTokenType string,
) (*Claims, error) {
	var claims Claims

	token, err := jwtlib.ParseWithClaims(
		tokenString,
		&claims,
		func(token *jwtlib.Token) (interface{}, error) {
			return secret, nil
		},
		jwtlib.WithValidMethods(
			[]string{jwtlib.SigningMethodHS256.Alg()},
		),
		jwtlib.WithExpirationRequired(),
		jwtlib.WithIssuer(m.issuer),
		jwtlib.WithIssuedAt(),
	)
	if err != nil {
		if errors.Is(err, jwtlib.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}

		return nil, ErrInvalidToken
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	if claims.TokenType != expectedTokenType {
		return nil, ErrUnexpectedTokenType
	}

	if claims.UserID <= 0 {
		return nil, ErrInvalidToken
	}

	if claims.SessionID == uuid.Nil {
		return nil, ErrInvalidToken
	}

	return &claims, nil
}
