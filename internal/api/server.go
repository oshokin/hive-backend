package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/oshokin/hive-backend/internal/common"
	"github.com/oshokin/hive-backend/internal/logger"
	city_service "github.com/oshokin/hive-backend/internal/service/city"
	user_service "github.com/oshokin/hive-backend/internal/service/user"
	go_cache "github.com/patrickmn/go-cache"
)

type Server interface {
	Start(ctx context.Context, port uint16)
	Stop(ctx context.Context)
}

type server struct {
	server       *http.Server
	router       chi.Router
	userService  user_service.Service
	cityService  city_service.Service
	cache        *go_cache.Cache
	jwtSecretKey []byte
}

const serverShutdownTimeout = 10 * time.Second

func NewServer(ctx context.Context,
	userService user_service.Service,
	cityService city_service.Service,
	jwtSecretKey []byte) Server {
	r := chi.NewRouter()
	s := &server{
		router:       r,
		userService:  userService,
		cityService:  cityService,
		cache:        go_cache.New(cacheExpirationTime, cacheCleanupInterval),
		jwtSecretKey: jwtSecretKey,
	}

	r.Use(middleware.RequestID, middleware.Recoverer)
	r.Use(middleware.Heartbeat("/ping"))

	r.Get("/v1/city/list", s.getCitiesHandler)
	r.Post("/v1/user/create", s.createUserHandler)
	r.Post("/v1/user/login", s.loginUserHandler)
	r.With(s.authMiddleware).Post("/v1/user/logout", s.logoutUserHandler)
	r.Get("/v1/user/{id}", s.getUserHandler)

	return s
}

func (s *server) Start(ctx context.Context, port uint16) {
	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: s.router,
	}

	logger.Infof(ctx, "starting server on port %d", port)
	go func(server *http.Server) {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.FatalKV(ctx, "failed to start server", common.ErrorTag, err)
		}
	}(s.server)

	logger.Infof(ctx, "server is up and listening on port %d", port)
}

func (s *server) Stop(ctx context.Context) {
	logger.Info(ctx, "shutting down server")

	ctx, cancel := context.WithTimeout(ctx, serverShutdownTimeout)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		logger.ErrorKV(ctx, "failed to stop server", common.ErrorTag, err)
	}

	logger.Info(ctx, "server stopped")
}
