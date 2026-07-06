package container

import (
	"server_nesting_optimizer/internal/config"
	"server_nesting_optimizer/internal/repository/postgres"
	userusecase "server_nesting_optimizer/internal/usecase/user"
	"server_nesting_optimizer/pkg/password"

	"github.com/jmoiron/sqlx"
)

type Container struct {
	CreateUserUseCase *userusecase.CreateUserUseCase
}

func New(
	db *sqlx.DB,
	cfg config.Config,
) *Container {
	userRepo := postgres.NewUserRepository(db)
	passwordHasher := password.NewBcryptHasher(cfg.Security.BcryptCost)

	createUserUseCase := userusecase.NewCreateUserUseCase(
		userRepo, passwordHasher,
	)

	return &Container{
		CreateUserUseCase: createUserUseCase,
	}
}
