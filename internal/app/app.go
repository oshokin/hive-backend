package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/oshokin/hive-backend/internal/api"
	"github.com/oshokin/hive-backend/internal/config"
	"github.com/oshokin/hive-backend/internal/db"
	"github.com/oshokin/hive-backend/internal/logger"
	city_repo "github.com/oshokin/hive-backend/internal/repository/city"
	randomizing_job_repo "github.com/oshokin/hive-backend/internal/repository/randomizing_job"
	user_repo "github.com/oshokin/hive-backend/internal/repository/user"
	city_service "github.com/oshokin/hive-backend/internal/service/city"
	randomizing_job_service "github.com/oshokin/hive-backend/internal/service/randomizing_job"
	user_service "github.com/oshokin/hive-backend/internal/service/user"
)

// Application represents the main application struct.
type Application struct {
	config                *config.Configuration           // Application configuration
	dbCluster             *db.Cluster                     // Database cluster
	cityRepo              city_repo.Repository            // Repository for managing city data
	cityService           city_service.Service            // Service for managing city data
	userRepo              user_repo.Repository            // Repository for managing user data
	userService           user_service.Service            // Service for managing user data
	randomizingJobRepo    randomizing_job_repo.Repository // Repository for managing user randomizing job data
	randomizingJobService randomizing_job_service.Service // Service for managing user randomizing job data
	server                api.Server                      // HTTP server for handling API requests
}

// NewApplication creates a new Application instance with the given context.
func NewApplication(ctx context.Context) (*Application, error) {
	var err error

	config, err := config.GetDefaults()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	dbCluster, err := db.NewCluster(ctx, config.DBClusterConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	logger.SetLevel(config.LogLevel)

	cityRepo := city_repo.NewRepository(dbCluster)
	cityService := city_service.NewService(cityRepo)
	userRepo := user_repo.NewRepository(dbCluster)
	userService := user_service.NewService(userRepo, cityService, config.FakeUserPassword)
	randomizingJobRepo := randomizing_job_repo.NewRepository(dbCluster)
	randomizingJobService := randomizing_job_service.NewService(randomizingJobRepo, userService)
	server := api.NewServer(userService,
		cityService,
		randomizingJobService,
		config)

	return &Application{
		config:                config,
		dbCluster:             dbCluster,
		cityRepo:              cityRepo,
		cityService:           cityService,
		userRepo:              userRepo,
		userService:           userService,
		randomizingJobRepo:    randomizingJobRepo,
		randomizingJobService: randomizingJobService,
		server:                server,
	}, nil
}

// Run starts the application and blocks until the context is done.
func (app *Application) Run(ctx context.Context) {
	ctx, stopReceivingSignals := signal.NotifyContext(ctx,
		os.Interrupt,
		os.Kill,
		syscall.SIGTERM,
		syscall.SIGINT)

	defer stopReceivingSignals()
	defer app.dbCluster.Close()

	app.server.Start(ctx, app.config.ServerPort)
	app.randomizingJobService.Start(ctx)

	<-ctx.Done()
	stopReceivingSignals()

	app.randomizingJobService.Stop(ctx)
	app.server.Stop(ctx)
}
