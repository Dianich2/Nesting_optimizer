package container

import (
	"server_nesting_optimizer/internal/config"
	"server_nesting_optimizer/internal/repository/postgres"
	userusecase "server_nesting_optimizer/internal/usecase/user"
	"server_nesting_optimizer/pkg/password"
	"time"

	domainsession "server_nesting_optimizer/internal/domain/session"
	jwtpkg "server_nesting_optimizer/pkg/jwt"

	"github.com/jmoiron/sqlx"
)

type Container struct {
	CreateUserUseCase     *userusecase.CreateUserUseCase
	LoginUseCase          *userusecase.LoginUseCase
	RefreshUseCase        *userusecase.RefreshUseCase
	LogoutUseCase         *userusecase.LogoutUseCase
	GetCurrentUserUseCase *userusecase.GetCurrentUserUseCase
	JWTManager            *jwtpkg.Manager
	SessionRepository     *postgres.SessionRepository
}

func New(
	db *sqlx.DB,
	cfg config.Config,
) *Container {
	userRepo := postgres.NewUserRepository(db)
	sessionRepo := postgres.NewSessionRepository(db)

	passwordHasher := password.NewBcryptHasher(cfg.Security.BcryptCost)

	jwtManager := jwtpkg.NewManager(
		cfg.JWT.AccessSecret,
		cfg.JWT.RefreshSecret,
		time.Duration(cfg.JWT.AccessTTLMin)*time.Minute,
		time.Duration(cfg.JWT.RefreshTTLMin)*time.Minute,
		cfg.JWT.Issuer,
	)

	createUserUseCase := userusecase.NewCreateUserUseCase(
		userRepo,
		passwordHasher,
	)

	sessionFactory := domainsession.NewFactory(
		time.Duration(cfg.JWT.RefreshTTLMin) * time.Minute,
	)

	loginUseCase := userusecase.NewLoginUseCase(
		userRepo,
		sessionRepo,
		passwordHasher,
		jwtManager,
		sessionFactory,
	)

	refreshUseCase := userusecase.NewRefreshUseCase(
		sessionRepo,
		sessionFactory,
		jwtManager,
	)

	logoutUseCase := userusecase.NewLogoutUseCase(
		sessionRepo,
	)

	getCurrentUserUseCase := userusecase.NewGetCurrentUserUseCase(
		userRepo,
	)

	return &Container{
		CreateUserUseCase:     createUserUseCase,
		LoginUseCase:          loginUseCase,
		RefreshUseCase:        refreshUseCase,
		LogoutUseCase:         logoutUseCase,
		GetCurrentUserUseCase: getCurrentUserUseCase,
		JWTManager:            jwtManager,
		SessionRepository:     sessionRepo,
	}
}
