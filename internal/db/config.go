package db

import "time"

// DatabaseConfiguration represents the configuration needed to establish a connection with a database.
type DatabaseConfiguration struct {
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
