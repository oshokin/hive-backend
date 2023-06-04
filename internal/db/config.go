package db

import (
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type (
	// DatabaseConfiguration represents the configuration needed to establish a connection with a database.
	DatabaseConfiguration struct {
		// Hostname or IP address of the database server.
		Host string `mapstructure:"host"`
		// Port number used to connect to the database server.
		Port uint16 `mapstructure:"port"`
		// Name of the database to connect to.
		Name string `mapstructure:"name"`
		// Username to authenticate with the database server.
		User string `mapstructure:"user"`
		// Password to authenticate with the database server.
		Password string `mapstructure:"password"`
		// Maximum number of simultaneous database connections.
		MaxConnections uint16 `mapstructure:"maxConnections"`
		// Maximum amount of time that a connection can live before being closed and reestablished.
		ConnectionLifetime time.Duration `mapstructure:"connectionLifetime"`
	}

	// ClusterConfiguration represents the configuration needed to set up a PostgreSQL cluster
	// with synchronous and asynchronous replicas.
	ClusterConfiguration struct {
		// Configuration for the master PostgreSQL server, which all other servers will replicate from.
		Master *DatabaseConfiguration
		// Configuration for the synchronous replica PostgreSQL server.
		Sync *DatabaseConfiguration
		// Configuration for the asynchronous replica PostgreSQL server.
		Async *DatabaseConfiguration
	}

	// Cluster represents a PostgreSQL cluster consisting of a master database and synchronous and asynchronous replicas.
	Cluster struct {
		// Connection pool for the master database.
		Master *pgxpool.Pool
		// Connection pool for the synchronous replica database.
		Sync *pgxpool.Pool
		// Connection pool for the asynchronous replica database.
		Async *pgxpool.Pool
	}
)

// Validate checks database configuration for errors.
func (v *DatabaseConfiguration) Validate(name string) error {
	if v == nil {
		return fmt.Errorf("%s database connection configuration is incorrect", name)
	}

	if v.Host == "" ||
		v.Port == 0 ||
		v.Name == "" ||
		v.User == "" ||
		v.Password == "" {
		return fmt.Errorf("%s database connection configuration is incorrect", name)
	}

	return nil
}
