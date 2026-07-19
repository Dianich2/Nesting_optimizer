package session

import (
	"server_nesting_optimizer/pkg/tokenhash"
	"time"

	"github.com/google/uuid"
)

type Session struct {
	SessionID        uuid.UUID `db:"session_id"`
	UserID           int64     `db:"user_id"`
	RefreshTokenHash string    `db:"refresh_token_hash"`
	CreatedAt        time.Time `db:"created_at"`
	ExpiresAt        time.Time `db:"expires_at"`
}

func (s Session) IsExpired(now time.Time) bool {
	return now.Compare(s.ExpiresAt) >= 0
}

func (s Session) IsActive(now time.Time) bool {
	return !s.IsExpired(now)
}

func (s *Session) SetRefreshToken(refreshToken string) {
	s.RefreshTokenHash = tokenhash.SHA256(refreshToken)
}

func (s *Session) MatchesRefreshToken(refreshToken string) bool {
	return s.RefreshTokenHash == tokenhash.SHA256(refreshToken)
}
