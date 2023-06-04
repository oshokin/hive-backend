package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/oshokin/hive-backend/internal/common"
	"github.com/oshokin/hive-backend/internal/config"
	"github.com/oshokin/hive-backend/internal/logger"
	city_service "github.com/oshokin/hive-backend/internal/service/city"
	randomizing_job_service "github.com/oshokin/hive-backend/internal/service/randomizing_job"
	user_service "github.com/oshokin/hive-backend/internal/service/user"
	chi_prometheus "github.com/oshokin/hive-backend/internal/util/chi-prometheus"
	go_cache "github.com/patrickmn/go-cache"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Server is the interface for starting and stopping the HTTP server.
type Server interface {
	Start(ctx context.Context, port uint16)
	Stop(ctx context.Context)
}

type server struct {
	server                *http.Server
	router                chi.Router
	userService           user_service.Service
	cityService           city_service.Service
	randomizingJobService randomizing_job_service.Service
	cache                 *go_cache.Cache
	jwtSecretKey          []byte
}

const (
	readHeaderTimeout     = 5 * time.Second
	serverShutdownTimeout = 10 * time.Second
)

// NewServer creates and returns a new Server instance.
func NewServer(userService user_service.Service,
	cityService city_service.Service,
	randomizingJobService randomizing_job_service.Service,
	config *config.Configuration) Server {
	r := chi.NewRouter()
	s := &server{
		router:                r,
		userService:           userService,
		cityService:           cityService,
		randomizingJobService: randomizingJobService,
		cache:                 go_cache.New(cacheExpirationTime, cacheCleanupInterval),
		jwtSecretKey:          config.JWTSecretKey,
	}

	r.Use(
		chi_prometheus.NewMiddleware(config.AppName),
		middleware.RequestID,
		middleware.Recoverer,
		middleware.Heartbeat("/ping"),
		middleware.Timeout(config.RequestTimeout))

	r.Handle("/metrics", promhttp.Handler())
	r.Get("/v1/city/list", s.getCitiesHandler)
	r.Get("/v1/randomizing-job/list", s.getRandomizingJobsHandler)
	r.Post("/v1/randomizing-job/create", s.createRandomizingJobHandler)
	r.Post("/v1/randomizing-job/cancel", s.cancelRandomizingJobHandler)
	r.Post("/v1/user/create", s.createUserHandler)
	r.Post("/v1/user/login", s.loginUserHandler)
	r.With(s.authMiddleware).Post("/v1/user/logout", s.logoutUserHandler)
	r.Get("/v1/user/{id}", s.getUserHandler)
	r.Get("/v1/user/search", s.searchUsersHandler)

	return s
}

// Start starts the HTTP server on the specified port.
func (s *server) Start(ctx context.Context, port uint16) {
	s.server = &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           s.router,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	logger.Infof(ctx, "starting server on port %d", port)

	go func(server *http.Server) {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.FatalKV(ctx, "failed to start server", common.ErrorTag, err)
		}
	}(s.server)

	logger.Infof(ctx, "server is up and listening on port %d", port)
}

// Stop stops the HTTP server.
func (s *server) Stop(ctx context.Context) {
	logger.Info(ctx, "shutting down server")

	ctx, cancel := context.WithTimeout(ctx, serverShutdownTimeout)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		logger.ErrorKV(ctx, "failed to stop server", common.ErrorTag, err)
	}

	logger.Info(ctx, "server stopped")
}
