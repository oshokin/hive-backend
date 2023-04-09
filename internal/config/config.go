package config

import (
	"errors"
	"fmt"
	"time"

	"github.com/oshokin/hive-backend/internal/db"
	"github.com/spf13/viper"
)

// Configuration represents the application configuration.
type Configuration struct {
	AppName          string                    // Name of the application.
	LogLevel         string                    // Logging level of the application.
	ServerPort       uint16                    // Port on which the application listens for requests.
	JWTSecretKey     []byte                    // Secret key used to sign and verify JSON Web Tokens.
	FakeUserPassword string                    // Password string used for generating random users.
	DBConfig         *db.DatabaseConfiguration // Database configuration.
}

// Constants with default values used for initialization.
const (
	defaultAppName              = "hive-backend"
	defaultEnvPrefix            = "HIVE_BACKEND"
	defaultServerPort           = uint16(8080)
	defaultDBMaxConnections     = 10
	defaultDBConnectionLifetime = 1 * time.Minute
)

// Errors that can occur during configuration validation.
var (
	errConfigIsEmpty                              = errors.New("configuration is empty")
	errJWTKeyIsEmpty                              = errors.New("jwt secret key is empty")
	errFakeUserPasswordIsEmpty                    = errors.New("fake user password is empty")
	errDatabaseConnectionConfigurationIsIncorrect = errors.New("database connection configuration is incorrect")
)

// GetDefaults loads configuration from environment variables and returns a pointer to Configuration.
func GetDefaults() (*Configuration, error) {
	viper.AutomaticEnv()
	viper.SetEnvPrefix(defaultEnvPrefix)

	config := getConfigFromEnvVars()
	config.enrichEmptyFieldsWithDefaults()

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}

	return config, nil
}

func getConfigFromEnvVars() *Configuration {
	return &Configuration{
		AppName:          defaultAppName,
		LogLevel:         viper.GetString("LOG_LEVEL"),
		ServerPort:       viper.GetUint16("SERVER_PORT"),
		JWTSecretKey:     []byte(viper.GetString("JWT_SECRET_KEY")),
		FakeUserPassword: viper.GetString("FAKE_USER_PASSWORD"),
		DBConfig: &db.DatabaseConfiguration{
			Host:               viper.GetString("DB_HOST"),
			Port:               viper.GetUint16("DB_PORT"),
			Name:               viper.GetString("DB_NAME"),
			User:               viper.GetString("DB_USER"),
			Password:           viper.GetString("DB_PASSWORD"),
			MaxConnections:     viper.GetUint16("DB_MAX_CONNECTIONS"),
			ConnectionLifetime: viper.GetDuration("DB_CONNECTION_LIFETIME"),
		},
	}
}

// Validate checks if Configuration is valid.
func (c *Configuration) Validate() error {
	if c == nil {
		return errConfigIsEmpty
	}

	if len(c.JWTSecretKey) == 0 {
		return errJWTKeyIsEmpty
	}

	if c.FakeUserPassword == "" {
		return errFakeUserPasswordIsEmpty
	}

	dbConfig := c.DBConfig
	if dbConfig == nil {
		return errDatabaseConnectionConfigurationIsIncorrect
	}

	if dbConfig.Host == "" ||
		dbConfig.Port == 0 ||
		dbConfig.Name == "" ||
		dbConfig.User == "" ||
		dbConfig.Password == "" {
		return errDatabaseConnectionConfigurationIsIncorrect
	}

	return nil
}

func (c *Configuration) enrichEmptyFieldsWithDefaults() {
	if c == nil {
		return
	}

	if c.ServerPort == 0 {
		c.ServerPort = defaultServerPort
	}

	dbConfig := c.DBConfig
	if dbConfig == nil {
		return
	}

	if dbConfig.MaxConnections == 0 {
		dbConfig.MaxConnections = defaultDBMaxConnections
	}

	if dbConfig.ConnectionLifetime == 0 {
		dbConfig.ConnectionLifetime = defaultDBConnectionLifetime
	}
}
