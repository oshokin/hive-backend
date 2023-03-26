package db

import (
	"context"

	"github.com/IBM/pgxpoolprometheus"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
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

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, err
	}

	collector := pgxpoolprometheus.NewCollector(pool,
		map[string]string{"db_name": v.Name})
	prometheus.MustRegister(collector)

	return pool, nil
}
