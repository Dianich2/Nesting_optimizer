package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	domainsession "server_nesting_optimizer/internal/domain/session"

	"github.com/google/uuid"
)

type SessionRepository struct {
	db DBTX
}

func NewSessionRepository(db DBTX) *SessionRepository {
	return &SessionRepository{
		db: db,
	}
}

func (r *SessionRepository) Upsert(
	ctx context.Context,
	session domainsession.Session,
) (domainsession.Session, error) {
	var upsertedSession domainsession.Session
	if err := r.db.GetContext(
		ctx,
		&upsertedSession,
		upsertSessionQuery,
		session.SessionID,
		session.UserID,
		session.RefreshTokenHash,
		session.ExpiresAt,
	); err != nil {
		return domainsession.Session{}, fmt.Errorf(
			"upsert session: %w",
			err,
		)
	}
	return upsertedSession, nil
}

func (r *SessionRepository) Rotate(
	ctx context.Context,
	oldSessionID uuid.UUID,
	oldRefreshTokenHash string,
	newSession domainsession.Session,
) (domainsession.Session, error) {
	var rotatedSession domainsession.Session
	if err := r.db.GetContext(
		ctx,
		&rotatedSession,
		rotateSessionQuery,
		newSession.SessionID,
		newSession.RefreshTokenHash,
		newSession.ExpiresAt,
		oldSessionID,
		oldRefreshTokenHash,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domainsession.Session{}, fmt.Errorf(
				"rotate session: %w",
				domainsession.ErrSessionChanged,
			)
		}

		return domainsession.Session{}, fmt.Errorf(
			"rotate session: %w",
			err,
		)
	}

	return rotatedSession, nil
}

func (r *SessionRepository) GetBySessionID(
	ctx context.Context,
	sessionID uuid.UUID,
) (domainsession.Session, error) {
	var session domainsession.Session
	if err := r.db.GetContext(
		ctx,
		&session,
		getSessionBySessionIDQuery,
		sessionID,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domainsession.Session{}, fmt.Errorf(
				"get session by session id: %w",
				domainsession.ErrNotFound,
			)
		}

		return domainsession.Session{}, fmt.Errorf(
			"get session by session id: %w",
			err,
		)
	}

	return session, nil
}

func (r *SessionRepository) GetByUserID(
	ctx context.Context,
	userID int64,
) (domainsession.Session, error) {
	var session domainsession.Session
	if err := r.db.GetContext(
		ctx,
		&session,
		getSessionByUserIDQuery,
		userID,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domainsession.Session{}, fmt.Errorf(
				"get session by user id: %w",
				domainsession.ErrNotFound,
			)
		}

		return domainsession.Session{}, fmt.Errorf(
			"get session by user id: %w",
			err,
		)
	}

	return session, nil
}

func (r *SessionRepository) DeleteBySessionID(
	ctx context.Context,
	sessionID uuid.UUID,
) error {
	if _, err := r.db.ExecContext(
		ctx,
		deleteSessionBySessionIDQuery,
		sessionID,
	); err != nil {
		return fmt.Errorf(
			"delete session by session id: %w",
			err,
		)
	}

	return nil
}

func (r *SessionRepository) DeleteExpired(
	ctx context.Context,
) error {
	if _, err := r.db.ExecContext(
		ctx,
		deleteExpiredSessionsQuery,
	); err != nil {
		return fmt.Errorf(
			"delete expired sessions: %w",
			err,
		)
	}

	return nil
}

func (r *SessionRepository) DeleteByUserID(
	ctx context.Context,
	userID int64,
) error {
	if _, err := r.db.ExecContext(
		ctx,
		deleteSessionByUserIDQuery,
		userID,
	); err != nil {
		return fmt.Errorf(
			"delete session by user id: %w",
			err,
		)
	}

	return nil
}
