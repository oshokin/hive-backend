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

	// connString := "postgres://user:password@localhost/dbname"
	// poolConfig, err := pgxpool.ParseConfig(connString)
	// if err != nil {
	// 	return err
	// }
	// poolConfig.MaxConns = 10
	// poolConfig.MinConns = 2
	// db, err = pgxpool.ConnectConfig(context.Background(), poolConfig)
	// if err != nil {
	// 	return err
	// }
	// return nil
}
