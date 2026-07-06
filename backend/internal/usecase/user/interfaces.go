package user

import (
	"context"
	domainuser "server_nesting_optimizer/internal/domain/user"
)

type UserRepository interface {
	Create(ctx context.Context, user domainuser.User) (domainuser.User, error)
	ExistsByLogin(ctx context.Context, login string) (bool, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}

type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(passwordHash string, password string) error
}
