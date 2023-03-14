package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v4/pgxpool"
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
	userService := user_service.NewService(userRepo)
	server := api.NewServer(ctx, userService, cityService, config.JWTSecretKey)

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

// User registration - POST /v1/users
// User login - POST /v1/users/login
// Get user info by ID - GET /v1/users/{id}
// Get cities list - GET /v1/cities
// Note that the version number is typically included at the beginning of the URI path, and it's common to start with "v1" for the initial version of an API. Also, note that the endpoint names are in the plural form to follow RESTful conventions.

// Олег Шокин
// my project is a social network,
// so i have users, cities, also I want to have a table that shows who is friend of who. what's the common and best name for such table that holds connections between friends?
// A common and standard name for such a table is "friends" or "user_friends".

// Олег Шокин
// okay so what's common name for the table with dialogues?
// The most common name for a table that holds dialogues is simply "dialogs" or "messages". However, some people may prefer more descriptive names like "conversations", "chat_history", or "message_logs". Ultimately, the best name for a table will depend on the specific requirements and context of your project.

// Олег Шокин
// and what about table with posts or comments?
// if some friend want to discuss something in some topic or friends' wall.
// what would be tables structure?
// The table for posts could be named something like posts and the table for comments could be named something like comments.

// For the discussion feature, you could create a new table that represents a discussion thread, and another table for the messages exchanged between users in that thread. For example:

// sql
// Copy code
// CREATE TABLE discussion_threads (
//     id bigserial PRIMARY KEY,
//     title varchar(255) NOT NULL,
//     created_at timestamp NOT NULL DEFAULT now()
// );

// CREATE TABLE discussion_messages (
//     id bigserial PRIMARY KEY,
//     thread_id bigint NOT NULL REFERENCES discussion_threads(id) ON DELETE CASCADE,
//     author_id bigint NOT NULL REFERENCES users(id) ON DELETE CASCADE,
//     message_text text NOT NULL,
//     created_at timestamp NOT NULL DEFAULT now()
// );
// In this structure, the discussion_threads table represents a discussion thread and the discussion_messages table holds the messages exchanged in that thread. The author_id field in the discussion_messages table would reference the id field in the users table, indicating the user who wrote the message.

// Олег Шокин
// what is the common structure for table "user_friends"?
// A common structure for a table that holds connections between friends is to have two columns that store the user IDs of the two friends. This table could be named "user_friends" or "friends" and might have the following structure:

// sql
// Copy code
// CREATE TABLE user_friends (
//     user_id_1 BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
//     user_id_2 BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
//     PRIMARY KEY (user_id_1, user_id_2)
// );
