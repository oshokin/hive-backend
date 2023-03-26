package config

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/oshokin/hive-backend/internal/db"
	"github.com/spf13/viper"
)

type Configuration struct {
	AppName      string
	LogLevel     string
	ServerPort   uint16
	JWTSecretKey []byte
	DBConfig     *db.DatabaseConfiguration
}

const (
	defaultAppName              = "hive-backend"
	defaultEnvPrefix            = "HIVE_BACKEND"
	defaultServerPort           = uint16(8080)
	defaultDBMaxConnections     = 10
	defaultDBConnectionLifetime = 1 * time.Minute
)

var (
	errConfigIsEmpty                              = errors.New("configuration is empty")
	errJWTKeyIsEmpty                              = errors.New("jwt secret key is empty")
	errDatabaseConnectionConfigurationIsIncorrect = errors.New("database connection configuration is incorrect")
)

func GetDefaults(ctx context.Context) (*Configuration, error) {
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
		AppName:      defaultAppName,
		LogLevel:     viper.GetString("LOG_LEVEL"),
		ServerPort:   viper.GetUint16("SERVER_PORT"),
		JWTSecretKey: []byte(viper.GetString("JWT_SECRET_KEY")),
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

func (c *Configuration) Validate() error {
	if c == nil {
		return errConfigIsEmpty
	}

	if len(c.JWTSecretKey) == 0 {
		return errJWTKeyIsEmpty
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
