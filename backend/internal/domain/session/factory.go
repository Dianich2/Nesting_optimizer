package session

import (
	"time"

	"github.com/google/uuid"
)

type Factory struct {
	refreshTTL time.Duration
}

func NewFactory(
	refreshTTL time.Duration,
) *Factory {
	return &Factory{
		refreshTTL: refreshTTL,
	}
}

func (f *Factory) New(
	userID int64,
) Session {
	now := time.Now()
	return Session{
		SessionID: uuid.New(),
		UserID:    userID,
		ExpiresAt: now.Add(f.refreshTTL),
	}
}
