package config

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/oshokin/hive-backend/internal/db"
	"github.com/spf13/viper"
)

// Configuration represents the application configuration.
type Configuration struct {
	AppName          string                   // Name of the application.
	LogLevel         string                   // Logging level of the application.
	ServerPort       uint16                   // Port on which the application listens for requests.
	RequestTimeout   time.Duration            // Maximum duration for a request to complete before timing out.
	JWTSecretKey     []byte                   // Secret key used to sign and verify JSON Web Tokens.
	FakeUserPassword string                   // Password string used for generating random users.
	DBClusterConfig  *db.ClusterConfiguration // Database cluster configuration.
}

// Constants with default values used for initialization.
const (
	defaultAppName              = "hive-backend"
	defaultEnvPrefix            = "HIVE_BACKEND"
	defaultServerPort           = uint16(8080)
	defaultRequestTimeout       = 5 * time.Second
	defaultDBMaxConnections     = 100
	defaultDBConnectionLifetime = 1 * time.Minute
)

// Errors that can occur during configuration validation.
var (
	errConfigIsEmpty           = errors.New("configuration is empty")
	errJWTKeyIsEmpty           = errors.New("jwt secret key is empty")
	errFakeUserPasswordIsEmpty = errors.New("fake user password is empty")
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
		DBClusterConfig: &db.ClusterConfiguration{
			Master: getDatabaseConfiguration("MASTER"),
			Sync:   getDatabaseConfiguration("SYNC"),
			Async:  getDatabaseConfiguration("ASYNC"),
		},
	}
}

func getDatabaseConfiguration(prefix string) *db.DatabaseConfiguration {
	addPrefix := func(key string) string {
		return strings.Join([]string{"DB", prefix, key}, "_")
	}

	return &db.DatabaseConfiguration{
		Host:               viper.GetString(addPrefix("HOST")),
		Port:               viper.GetUint16(addPrefix("PORT")),
		Name:               viper.GetString(addPrefix("NAME")),
		User:               viper.GetString(addPrefix("USER")),
		Password:           viper.GetString(addPrefix("PASSWORD")),
		MaxConnections:     viper.GetUint16(addPrefix("MAX_CONNECTIONS")),
		ConnectionLifetime: viper.GetDuration(addPrefix("CONNECTION_LIFETIME")),
	}
}

// Validate checks if configuration is valid.
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

	dbc := c.DBClusterConfig
	if err := dbc.Master.Validate("master"); err != nil {
		return err
	}

	if err := dbc.Sync.Validate("sync"); err != nil {
		return err
	}

	return dbc.Async.Validate("async")
}

func (c *Configuration) enrichEmptyFieldsWithDefaults() {
	if c.AppName == "" {
		c.AppName = defaultAppName
	}

	if c.ServerPort == 0 {
		c.ServerPort = defaultServerPort
	}

	if c.RequestTimeout == 0 {
		c.RequestTimeout = defaultRequestTimeout
	}

	if dbc := c.DBClusterConfig; dbc != nil {
		c.enrichEmptyDBConfig(dbc.Master)
		c.enrichEmptyDBConfig(dbc.Sync)
		c.enrichEmptyDBConfig(dbc.Async)
	}
}

func (c *Configuration) enrichEmptyDBConfig(v *db.DatabaseConfiguration) {
	if v == nil {
		v = &db.DatabaseConfiguration{}
	}

	if v.MaxConnections == 0 {
		v.MaxConnections = defaultDBMaxConnections
	}

	if v.ConnectionLifetime == 0 {
		v.ConnectionLifetime = defaultDBConnectionLifetime
	}
}
