package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/oshokin/hive-backend/internal/api"
	"github.com/oshokin/hive-backend/internal/config"
	"github.com/oshokin/hive-backend/internal/db"
	"github.com/oshokin/hive-backend/internal/logger"
	city_repo "github.com/oshokin/hive-backend/internal/repository/city"
	user_repo "github.com/oshokin/hive-backend/internal/repository/user"
	city_service "github.com/oshokin/hive-backend/internal/service/city"
	user_service "github.com/oshokin/hive-backend/internal/service/user"
)

type Application struct {
	config      *config.Configuration
	dbPool      *pgxpool.Pool
	cityRepo    city_repo.Repository
	cityService city_service.Service
	userRepo    user_repo.Repository
	userService user_service.Service
	server      api.Server
}

func NewApplication(ctx context.Context) (*Application, error) {
	var err error
	config, err := config.GetDefaults(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	dbPool, err := db.GetDBPool(ctx, config.DBConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	logger.SetLevel(config.LogLevel)

	cityRepo := city_repo.NewRepository(dbPool)
	cityService := city_service.NewService(cityRepo)
	userRepo := user_repo.NewRepository(dbPool)
	userService := user_service.NewService(userRepo, cityService)
	server := api.NewServer(ctx, userService, cityService, config)

	return &Application{
		config:      config,
		dbPool:      dbPool,
		cityRepo:    cityRepo,
		cityService: cityService,
		userRepo:    userRepo,
		userService: userService,
		server:      server,
	}, nil
}

func (app *Application) Run(ctx context.Context) {
	ctx, stopReceivingSignals := signal.NotifyContext(ctx,
		os.Interrupt,
		os.Kill,
		syscall.SIGTERM,
		syscall.SIGINT)

	defer stopReceivingSignals()
	defer app.dbPool.Close()

	app.server.Start(ctx, app.config.ServerPort)

	<-ctx.Done()
	stopReceivingSignals()

	app.server.Stop(ctx)
}
