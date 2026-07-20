package user

import (
	"context"
	domainsession "server_nesting_optimizer/internal/domain/session"
	domainuser "server_nesting_optimizer/internal/domain/user"
	"server_nesting_optimizer/pkg/jwt"

	"github.com/google/uuid"
)

type UserRepository interface {
	Create(
		ctx context.Context,
		user domainuser.User,
	) (domainuser.User, error)

	ExistsByLogin(
		ctx context.Context,
		login string,
	) (bool, error)

	ExistsByEmail(
		ctx context.Context,
		email string,
	) (bool, error)

	GetByIdentifier(
		ctx context.Context,
		identifier string,
	) (domainuser.User, error)

	GetByID(
		ctx context.Context,
		id int64,
	) (domainuser.User, error)
}

type PasswordHasher interface {
	Hash(
		password string,
	) (string, error)

	Compare(
		passwordHash string,
		password string,
	) error
}

type TokenManager interface {
	GenerateAccessToken(
		userID int64,
		sessionID uuid.UUID,
	) (string, error)

	GenerateRefreshToken(
		userID int64,
		sessionID uuid.UUID,
	) (string, error)

	ParseRefreshToken(
		tokenString string,
	) (*jwt.Claims, error)
}

type SessionRepository interface {
	Upsert(
		ctx context.Context,
		session domainsession.Session,
	) (domainsession.Session, error)

	Rotate(
		ctx context.Context,
		oldSessionID uuid.UUID,
		oldRefreshTokenHash string,
		newSession domainsession.Session,
	) (domainsession.Session, error)

	GetBySessionID(
		ctx context.Context,
		sessionID uuid.UUID,
	) (domainsession.Session, error)

	GetByUserID(
		ctx context.Context,
		userID int64,
	) (domainsession.Session, error)

	DeleteBySessionID(
		ctx context.Context,
		sessionID uuid.UUID,
	) error

	DeleteExpired(
		ctx context.Context,
	) error
}

type SessionFactory interface {
	New(
		userID int64,
	) domainsession.Session
}
