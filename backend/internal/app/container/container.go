package container

import (
	"server_nesting_optimizer/internal/config"
	"server_nesting_optimizer/internal/geometry/simplefeatures"
	"server_nesting_optimizer/internal/repository/postgres"
	projectusecase "server_nesting_optimizer/internal/usecase/project"
	projectsurfaceusecase "server_nesting_optimizer/internal/usecase/project_surface"
	surfaceusecase "server_nesting_optimizer/internal/usecase/surface"
	userusecase "server_nesting_optimizer/internal/usecase/user"
	"server_nesting_optimizer/pkg/password"
	"time"

	domainsession "server_nesting_optimizer/internal/domain/session"
	jwtpkg "server_nesting_optimizer/pkg/jwt"

	"github.com/jmoiron/sqlx"
)

type Container struct {
	CreateUserUseCase            *userusecase.CreateUserUseCase
	LoginUseCase                 *userusecase.LoginUseCase
	RefreshUseCase               *userusecase.RefreshUseCase
	LogoutUseCase                *userusecase.LogoutUseCase
	GetCurrentUserUseCase        *userusecase.GetCurrentUserUseCase
	UpdateProfileUseCase         *userusecase.UpdateProfileUseCase
	ChangePasswordUseCase        *userusecase.ChangePasswordUseCase
	DeleteCurrentUserUseCase     *userusecase.DeleteCurrentUserUseCase
	CreateProjectUseCase         *projectusecase.CreateProjectUseCase
	GetProjectByIDUseCase        *projectusecase.GetProjectByIDUseCase
	ListProjectsUseCase          *projectusecase.ListProjectsUseCase
	UpdateProjectUseCase         *projectusecase.UpdateProjectUseCase
	DeleteProjectUseCase         *projectusecase.DeleteProjectUseCase
	CreateSurfaceUseCase         *surfaceusecase.CreateSurfaceUseCase
	GetSurfaceByIDUseCase        *surfaceusecase.GetSurfaceByIDUseCase
	ListSurfacesUseCase          *surfaceusecase.ListSurfacesUseCase
	UpdateSurfaceUseCase         *surfaceusecase.UpdateSurfaceUseCase
	DeleteSurfaceUseCase         *surfaceusecase.DeleteSurfaceUseCase
	CreateProjectSurfaceUseCase  *projectsurfaceusecase.CreateProjectSurfaceUseCase
	GetProjectSurfaceByIDUseCase *projectsurfaceusecase.GetProjectSurfaceByIDUseCase
	JWTManager                   *jwtpkg.Manager
	SessionRepository            *postgres.SessionRepository
}

func New(
	db *sqlx.DB,
	cfg config.Config,
) *Container {
	unitOfWork := postgres.NewUnitOfWork(db)
	sfCodec := simplefeatures.NewCodec()
	sfEngine := simplefeatures.NewEngine()

	userRepo := postgres.NewUserRepository(db)
	sessionRepo := postgres.NewSessionRepository(db)
	projectRepo := postgres.NewProjectRepository(db)
	surfaceRepo := postgres.NewSurfaceRepository(
		db,
		sfCodec,
	)
	projectSurfaceRepo := postgres.NewProjectSurfaceRepository(
		db,
		sfCodec,
	)

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

	updateProfileUseCase := userusecase.NewUpdateProfileUseCase(
		userRepo,
	)

	changePasswordUseCase := userusecase.NewChangePasswordUseCase(
		userRepo,
		passwordHasher,
		unitOfWork,
	)

	deleteCurrentUserUseCase := userusecase.NewDeleteCurrentUserUseCase(
		userRepo,
		passwordHasher,
		unitOfWork,
	)

	createProjectUseCase := projectusecase.NewCreateProjectUseCase(
		projectRepo,
	)

	getProjectByIDUseCase := projectusecase.NewGetProjectByIDUseCase(
		projectRepo,
	)

	listProjectsUseCase := projectusecase.NewListProjectsUseCase(
		projectRepo,
	)

	updateProjectUseCase := projectusecase.NewUpdateProjectUseCase(
		projectRepo,
	)

	deleteProjectUseCase := projectusecase.NewDeleteProjectUseCase(
		projectRepo,
	)

	createSurfaceUseCase := surfaceusecase.NewCreateSurfaceUseCase(
		surfaceRepo,
		sfEngine,
	)

	getSurfaceByIDUseCase := surfaceusecase.NewGetSurfaceByIDUseCase(
		surfaceRepo,
	)

	listSurfacesUseCase := surfaceusecase.NewListSurfacesUseCase(
		surfaceRepo,
	)

	updateSurfaceUseCase := surfaceusecase.NewUpdateSurfaceUseCase(
		surfaceRepo,
		sfEngine,
	)

	deleteSurfaceUseCase := surfaceusecase.NewDeleteSurfaceUseCase(
		surfaceRepo,
	)

	createProjectSurfaceUseCase := projectsurfaceusecase.NewCreateProjectSurfaceUseCase(
		projectSurfaceRepo,
		projectRepo,
		surfaceRepo,
		sfEngine,
	)

	getProjectSurfaceByIDUseCase := projectsurfaceusecase.NewGetProjectSurfaceByIDUseCase(
		projectSurfaceRepo,
	)

	return &Container{
		CreateUserUseCase:            createUserUseCase,
		LoginUseCase:                 loginUseCase,
		RefreshUseCase:               refreshUseCase,
		LogoutUseCase:                logoutUseCase,
		UpdateProfileUseCase:         updateProfileUseCase,
		GetCurrentUserUseCase:        getCurrentUserUseCase,
		ChangePasswordUseCase:        changePasswordUseCase,
		DeleteCurrentUserUseCase:     deleteCurrentUserUseCase,
		CreateProjectUseCase:         createProjectUseCase,
		GetProjectByIDUseCase:        getProjectByIDUseCase,
		ListProjectsUseCase:          listProjectsUseCase,
		UpdateProjectUseCase:         updateProjectUseCase,
		DeleteProjectUseCase:         deleteProjectUseCase,
		CreateSurfaceUseCase:         createSurfaceUseCase,
		GetSurfaceByIDUseCase:        getSurfaceByIDUseCase,
		ListSurfacesUseCase:          listSurfacesUseCase,
		UpdateSurfaceUseCase:         updateSurfaceUseCase,
		DeleteSurfaceUseCase:         deleteSurfaceUseCase,
		CreateProjectSurfaceUseCase:  createProjectSurfaceUseCase,
		GetProjectSurfaceByIDUseCase: getProjectSurfaceByIDUseCase,
		JWTManager:                   jwtManager,
		SessionRepository:            sessionRepo,
	}
}
