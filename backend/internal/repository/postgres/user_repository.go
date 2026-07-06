package postgres

import (
	"context"

	domainuser "server_nesting_optimizer/internal/domain/user"

	"github.com/jmoiron/sqlx"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (ur *UserRepository) Create(
	ctx context.Context,
	user domainuser.User,
) (domainuser.User, error) {
	var createdUser domainuser.User
	if err := ur.db.GetContext(
		ctx,
		&createdUser,
		createUserQuery,
		user.Login,
		user.Email,
		user.PasswordHash,
		user.FirstName,
		user.LastName,
	); err != nil {
		return domainuser.User{}, err
	}
	return createdUser, nil
}

func (ur *UserRepository) ExistsByLogin(
	ctx context.Context,
	login string,
) (bool, error) {
	var exists bool
	if err := ur.db.GetContext(ctx, &exists, existUserByLoginQuery, login); err != nil {
		return false, err
	}
	return exists, nil
}

func (ur *UserRepository) ExistsByEmail(
	ctx context.Context,
	email string,
) (bool, error) {
	var exists bool
	if err := ur.db.GetContext(ctx, &exists, existUserByEmailQuery, email); err != nil {
		return false, err
	}
	return exists, nil
}
