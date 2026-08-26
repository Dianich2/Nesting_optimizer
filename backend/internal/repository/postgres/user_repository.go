package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	domainuser "server_nesting_optimizer/internal/domain/user"
)

type UserRepository struct {
	db DBTX
}

func NewUserRepository(db DBTX) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func mapCreateUserError(err error) error {
	constraint, ok := uniqueViolationConstraint(err)
	if !ok {
		return err
	}

	switch constraint {
	case "users_login_unique":
		return domainuser.ErrLoginAlreadyExists

	case "users_email_unique":
		return domainuser.ErrEmailAlreadyExists

	default:
		return err
	}
}

func (r *UserRepository) Create(
	ctx context.Context,
	user domainuser.User,
) (domainuser.User, error) {
	var createdUser domainuser.User
	if err := r.db.GetContext(
		ctx,
		&createdUser,
		createUserQuery,
		user.Login,
		user.Email,
		user.PasswordHash,
		user.FirstName,
		user.LastName,
	); err != nil {
		mappedErr := mapCreateUserError(err)
		return domainuser.User{}, fmt.Errorf(
			"create user: %w",
			mappedErr,
		)
	}
	return createdUser, nil
}

func (r *UserRepository) ExistsByLogin(
	ctx context.Context,
	login string,
) (bool, error) {
	var exists bool
	if err := r.db.GetContext(ctx, &exists, existsUserByLoginQuery, login); err != nil {
		return false, fmt.Errorf(
			"exists by login: %w",
			err,
		)
	}
	return exists, nil
}

func (r *UserRepository) ExistsByEmail(
	ctx context.Context,
	email string,
) (bool, error) {
	var exists bool
	if err := r.db.GetContext(ctx, &exists, existsUserByEmailQuery, email); err != nil {
		return false, fmt.Errorf(
			"exists by email: %w",
			err,
		)
	}
	return exists, nil
}

func (r *UserRepository) GetByIdentifier(
	ctx context.Context,
	identifier string,
) (domainuser.User, error) {
	var user domainuser.User
	if err := r.db.GetContext(ctx, &user, getByIdentifierQuery, identifier); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domainuser.User{}, fmt.Errorf(
				"get user by identifier: %w",
				domainuser.ErrNotFound,
			)
		}

		return domainuser.User{}, fmt.Errorf(
			"get user by identifier: %w",
			err,
		)
	}
	return user, nil
}

func (r *UserRepository) GetByID(
	ctx context.Context,
	id int64,
) (domainuser.User, error) {
	var user domainuser.User
	if err := r.db.GetContext(
		ctx,
		&user,
		getByIDQuery,
		id,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domainuser.User{}, fmt.Errorf(
				"get user by id: %w",
				domainuser.ErrNotFound,
			)
		}

		return domainuser.User{}, fmt.Errorf(
			"get user by id: %w",
			err,
		)
	}

	return user, nil
}

func (r *UserRepository) UpdateProfile(
	ctx context.Context,
	firstName *string,
	lastName *string,
	id int64,
) (domainuser.User, error) {
	var updatedUser domainuser.User
	if err := r.db.GetContext(
		ctx,
		&updatedUser,
		updateProfileQuery,
		firstName,
		lastName,
		id,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domainuser.User{}, fmt.Errorf(
				"update user profile: %w",
				domainuser.ErrNotFound,
			)
		}

		return domainuser.User{}, fmt.Errorf(
			"update user profile: %w",
			err,
		)
	}

	return updatedUser, nil
}

func (r *UserRepository) ChangePassword(
	ctx context.Context,
	userID int64,
	oldPasswordHash string,
	newPasswordHash string,
) error {
	affected, err := execAffected(
		ctx,
		r.db,
		updatePasswordQuery,
		newPasswordHash,
		oldPasswordHash,
		userID,
	)
	if err != nil {
		return fmt.Errorf(
			"change password: %w",
			err,
		)
	}

	if affected == 0 {
		return fmt.Errorf(
			"change user password: %w",
			domainuser.ErrPasswordChanged,
		)
	}

	return nil
}

func (r *UserRepository) SoftDelete(
	ctx context.Context,
	userID int64,
	expectedPasswordHash string,
) error {
	affected, err := execAffected(
		ctx,
		r.db,
		softDeleteUserQuery,
		userID,
		expectedPasswordHash,
	)
	if err != nil {
		return fmt.Errorf("soft delete user: %w", err)
	}

	if affected == 0 {
		return fmt.Errorf(
			"soft delete user: %w",
			domainuser.ErrUserChanged,
		)
	}

	return nil
}
