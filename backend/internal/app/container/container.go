package container

import (
	"fmt"
	"server_nesting_optimizer/internal/config"
	"server_nesting_optimizer/internal/geometry/nfp"
	"server_nesting_optimizer/internal/geometry/simplefeatures"
	"server_nesting_optimizer/internal/nesting"
	"server_nesting_optimizer/internal/repository/postgres"
	nestingusecase "server_nesting_optimizer/internal/usecase/nesting"
	patternusecase "server_nesting_optimizer/internal/usecase/pattern"
	placementusecase "server_nesting_optimizer/internal/usecase/placement"
	projectusecase "server_nesting_optimizer/internal/usecase/project"
	projectpatternusecase "server_nesting_optimizer/internal/usecase/project_pattern"
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
	ListProjectSurfacesUseCase   *projectsurfaceusecase.ListProjectSurfacesUseCase
	UpdateProjectSurfaceUseCase  *projectsurfaceusecase.UpdateProjectSurfaceUseCase
	DeleteProjectSurfaceUseCase  *projectsurfaceusecase.DeleteProjectSurfaceUseCase
	CreatePatternUseCase         *patternusecase.CreatePatternUseCase
	GetPatternByIDUseCase        *patternusecase.GetPatternByIDUseCase
	ListPatternsUseCase          *patternusecase.ListPatternsUseCase
	UpdatePatternUseCase         *patternusecase.UpdatePatternUseCase
	DeletePatternUseCase         *patternusecase.DeletePatternUseCase
	CreateProjectPatternUseCase  *projectpatternusecase.CreateProjectPatternUseCase
	GetProjectPatternByIDUseCase *projectpatternusecase.GetProjectPatternByIDUseCase
	ListProjectPatternsUseCase   *projectpatternusecase.ListProjectPatternsUseCase
	UpdateProjectPatternUseCase  *projectpatternusecase.UpdateProjectPatternUseCase
	DeleteProjectPatternUseCase  *projectpatternusecase.DeleteProjectPatternUseCase
	CreatePlacementUseCase       *placementusecase.CreatePlacementUseCase
	GetPlacementByIDUseCase      *placementusecase.GetPlacementByIDUseCase
	ListPlacementsUseCase        *placementusecase.ListPlacementsUseCase
	UpdatePlacementUseCase       *placementusecase.UpdatePlacementUseCase
	DeletePlacementUseCase       *placementusecase.DeletePlacementUseCase
	RunNestingUseCase            *nestingusecase.RunNestingUseCase
	JWTManager                   *jwtpkg.Manager
	SessionRepository            *postgres.SessionRepository
}

func New(
	db *sqlx.DB,
	cfg config.Config,
) (*Container, error) {
	unitOfWork := postgres.NewUnitOfWork(db)
	sfCodec := simplefeatures.NewCodec()
	sfEngine := simplefeatures.NewEngine()
	nestingUnitOfWork := postgres.NewNestingUnitOfWork(db, sfCodec)

	scopedDB := &postgres.ScopedDB{DB: db}

	userRepo := postgres.NewUserRepository(scopedDB)
	sessionRepo := postgres.NewSessionRepository(scopedDB)
	projectRepo := postgres.NewProjectRepository(scopedDB)
	surfaceRepo := postgres.NewSurfaceRepository(
		scopedDB,
		sfCodec,
	)
	projectSurfaceRepo := postgres.NewProjectSurfaceRepository(
		scopedDB,
		sfCodec,
	)
	patternRepo := postgres.NewPatternRepository(
		scopedDB,
		sfCodec,
	)
	projectPatternRepo := postgres.NewProjectPatternRepository(
		scopedDB,
		sfCodec,
	)
	placementRepo := postgres.NewPlacementRepository(
		scopedDB,
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

	listProjectSurfacesUseCase := projectsurfaceusecase.NewListProjectSurfacesUseCase(
		projectSurfaceRepo,
	)

	updateProjectSurfaceUseCase := projectsurfaceusecase.NewUpdateProjectSurfaceUseCase(
		projectSurfaceRepo,
		sfEngine,
	)

	deleteProjectSurfaceUseCase := projectsurfaceusecase.NewDeleteProjectSurfaceUseCase(
		projectSurfaceRepo,
	)

	createPatternUseCase := patternusecase.NewCreatePatternUseCase(
		patternRepo,
		sfEngine,
	)

	getPatternByIDUseCase := patternusecase.NewGetPatternByIDUseCase(
		patternRepo,
	)

	listPatternsUseCase := patternusecase.NewListPatternsUseCase(
		patternRepo,
	)

	updatePatternUseCase := patternusecase.NewUpdatePatternUseCase(
		patternRepo,
		sfEngine,
	)

	deletePatternUseCase := patternusecase.NewDeletePatternUseCase(
		patternRepo,
	)

	createProjectPatternUseCase := projectpatternusecase.NewCreateProjectPatternUseCase(
		projectPatternRepo,
		projectRepo,
		patternRepo,
		sfEngine,
	)

	getProjectPatternByIDUseCase := projectpatternusecase.NewGetProjectPatternByIDUseCase(
		projectPatternRepo,
	)

	listProjectPatternsUseCase := projectpatternusecase.NewListProjectPatternsUseCase(
		projectPatternRepo,
	)

	updateProjectPatternUseCase := projectpatternusecase.NewUpdateProjectPatternUseCase(
		projectPatternRepo,
		sfEngine,
	)

	deleteProjectPatternUseCase := projectpatternusecase.NewDeleteProjectPatternUseCase(
		projectPatternRepo,
	)

	createPlacementUseCase := placementusecase.NewCreatePlacementUseCase(
		placementRepo,
		projectSurfaceRepo,
		projectPatternRepo,
		sfEngine,
	)

	getPlacementByIDUseCase := placementusecase.NewGetPlacementByIDUseCase(
		placementRepo,
		sfEngine,
	)

	listPlacementsUseCase := placementusecase.NewListPlacementsUseCase(
		placementRepo,
		projectSurfaceRepo,
		sfEngine,
	)

	updatePlacementUseCase := placementusecase.NewUpdatePlacementUseCase(
		placementRepo,
		projectSurfaceRepo,
		sfEngine,
	)

	deletePlacementUseCase := placementusecase.NewDeletePlacementUseCase(
		placementRepo,
	)

	optimizerFactories := map[nesting.Algorithm]nesting.OptimizerFactory{
		nesting.BaselineAlgorithm: func() (nesting.Optimizer, error) {
			engine := simplefeatures.NewEngine()

			return nesting.NewBaselineOptimizer(engine), nil
		},

		nesting.NFPGreedyAlgorithm: func() (nesting.Optimizer, error) {
			engine := simplefeatures.NewEngine()
			builder := nfp.NewCachedBuilder(
				nfp.NewBuilder(engine),
			)

			return nesting.NewNFPGreedyOptimizer(
				engine,
				builder,
			), nil
		},

		nesting.GeneticAlgorithm: func() (nesting.Optimizer, error) {
			engine := simplefeatures.NewEngine()
			builder := nfp.NewCachedBuilder(
				nfp.NewBuilder(engine),
			)

			return nesting.NewGeneticOptimizer(
				engine,
				builder,
				nesting.DefaultGeneticConfig(),
			)
		},

		nesting.SimulatedAnnealingAlgorithm: func() (nesting.Optimizer, error) {
			engine := simplefeatures.NewEngine()
			builder := nfp.NewCachedBuilder(
				nfp.NewBuilder(engine),
			)

			return nesting.NewSimulatedAnnealingOptimizer(
				engine,
				builder,
				nesting.DefaultAnnealingConfig(),
			)
		},

		nesting.HybridAlgorithm: func() (nesting.Optimizer, error) {
			engine := simplefeatures.NewEngine()
			builder := nfp.NewCachedBuilder(
				nfp.NewBuilder(engine),
			)

			geneticOptimizer, err := nesting.NewGeneticOptimizer(
				engine,
				builder,
				nesting.DefaultGeneticConfig(),
			)
			if err != nil {
				return nil, fmt.Errorf(
					"create hybrid genetic optimizer: %w",
					err,
				)
			}

			annealingOptimizer, err := nesting.NewSimulatedAnnealingOptimizer(
				engine,
				builder,
				nesting.DefaultAnnealingConfig(),
			)
			if err != nil {
				return nil, fmt.Errorf(
					"create hybrid annealing optimizer: %w",
					err,
				)
			}

			return nesting.NewHybridOptimizer(
				engine,
				geneticOptimizer,
				annealingOptimizer,
				nesting.DefaultCompactionConfig(),
			)
		},
	}

	optimizerRegistry, err := nesting.NewOptimizerFactoryRegistry(
		optimizerFactories,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"create optimizer registry: %w",
			err,
		)
	}

	runNestingUseCase := nestingusecase.NewRunNestingUseCase(
		projectSurfaceRepo,
		projectPatternRepo,
		placementRepo,
		sfEngine,
		optimizerRegistry,
		nestingUnitOfWork,
	)

	createProjectSurfaceUseCase.Transactions = scopedDB
	updateProjectSurfaceUseCase.Transactions = scopedDB
	deleteProjectSurfaceUseCase.Transactions = scopedDB
	createProjectPatternUseCase.Transactions = scopedDB
	updateProjectPatternUseCase.Transactions = scopedDB
	deleteProjectPatternUseCase.Transactions = scopedDB
	createPlacementUseCase.Transactions = scopedDB
	updatePlacementUseCase.Transactions = scopedDB
	deletePlacementUseCase.Transactions = scopedDB
	updateProjectUseCase.Transactions = scopedDB
	deleteProjectUseCase.Transactions = scopedDB
	runNestingUseCase.Transactions = scopedDB
	runNestingUseCase.Budgets = scopedDB
	runNestingUseCase.Limiter = nestingusecase.NewComputeLimiter(cfg.Nesting.Workers, cfg.Nesting.Queue)
	runNestingUseCase.Previews = postgres.NewNestingPreviewRepository(scopedDB)
	runNestingUseCase.History = postgres.NewNestingRunRepository(scopedDB)
	runNestingUseCase.Operations = postgres.NewNestingOperationRepository(scopedDB)

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
		ListProjectSurfacesUseCase:   listProjectSurfacesUseCase,
		UpdateProjectSurfaceUseCase:  updateProjectSurfaceUseCase,
		DeleteProjectSurfaceUseCase:  deleteProjectSurfaceUseCase,
		CreatePatternUseCase:         createPatternUseCase,
		GetPatternByIDUseCase:        getPatternByIDUseCase,
		ListPatternsUseCase:          listPatternsUseCase,
		UpdatePatternUseCase:         updatePatternUseCase,
		DeletePatternUseCase:         deletePatternUseCase,
		CreateProjectPatternUseCase:  createProjectPatternUseCase,
		GetProjectPatternByIDUseCase: getProjectPatternByIDUseCase,
		ListProjectPatternsUseCase:   listProjectPatternsUseCase,
		UpdateProjectPatternUseCase:  updateProjectPatternUseCase,
		DeleteProjectPatternUseCase:  deleteProjectPatternUseCase,
		CreatePlacementUseCase:       createPlacementUseCase,
		GetPlacementByIDUseCase:      getPlacementByIDUseCase,
		ListPlacementsUseCase:        listPlacementsUseCase,
		UpdatePlacementUseCase:       updatePlacementUseCase,
		DeletePlacementUseCase:       deletePlacementUseCase,
		RunNestingUseCase:            runNestingUseCase,
		JWTManager:                   jwtManager,
		SessionRepository:            sessionRepo,
	}, nil
}
