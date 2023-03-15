package db

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
)

func GetDBPool(ctx context.Context, v *DatabaseConfiguration) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig("")
	if err != nil {
		return nil, err
	}

	connConfig := poolConfig.ConnConfig
	connConfig.Host = v.Host
	connConfig.Port = v.Port
	connConfig.Database = v.Name
	connConfig.User = v.User
	connConfig.Password = v.Password

	poolConfig.MaxConnLifetime = v.ConnectionLifetime
	poolConfig.MaxConns = int32(v.MaxConnections)

	return pgxpool.ConnectConfig(ctx, poolConfig)
}
