package db

import "time"

type DatabaseConfiguration struct {
	Host               string        `mapstructure:"host"`
	Port               uint16        `mapstructure:"port"`
	Name               string        `mapstructure:"name"`
	User               string        `mapstructure:"user"`
	Password           string        `mapstructure:"password"`
	MaxConnections     uint16        `mapstructure:"maxConnections"`
	ConnectionLifetime time.Duration `mapstructure:"connectionLifetime"`
}
