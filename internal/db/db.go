package db

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/oshokin/hive-backend/internal/common"
	pgx_pool_collector "github.com/oshokin/hive-backend/internal/util/pgx-pool-prometheus"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	dbCount         = 3
	readOnlyDBCount = 2
)

var roundRobinIndex uint32

// NewCluster takes a context and a ClusterConfiguration and returns a Cluster,
// which consists of three connection pools:
// a master pool, a synchronous replica pool, and an asynchronous replica pool.
// It creates the three pools concurrently and returns an error if any of the pools failed to be created.
func NewCluster(ctx context.Context, v *ClusterConfiguration) (*Cluster, error) {
	var (
		master *pgxpool.Pool
		sync   *pgxpool.Pool
		async  *pgxpool.Pool
		wg     = common.GetDefaultWG(dbCount)
	)

	wg.Add(
		func() (localErr error) {
			master, localErr = newPool(ctx, v.Master)
			if localErr != nil {
				return fmt.Errorf("failed to create master database pool: %w", localErr)
			}

			return nil
		},
		func() (localErr error) {
			sync, localErr = newPool(ctx, v.Sync)
			if localErr != nil {
				return fmt.Errorf("failed to create sync database pool: %w", localErr)
			}

			return nil
		},
		func() (localErr error) {
			async, localErr = newPool(ctx, v.Async)
			if localErr != nil {
				return fmt.Errorf("failed to create async database pool: %w", localErr)
			}

			return nil
		})

	if err := wg.Start().GetLastError(); err != nil {
		return nil, err
	}

	staters := map[string]pgx_pool_collector.Stater{
		"master": master,
		"sync":   sync,
		"async":  async,
	}

	collector := pgx_pool_collector.NewCollector(staters, nil)
	prometheus.MustRegister(collector)

	return &Cluster{
		Master: master,
		Sync:   sync,
		Async:  async,
	}, nil
}

func newPool(ctx context.Context, v *DatabaseConfiguration) (*pgxpool.Pool, error) {
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

	return pool, nil
}

// Write returns the connection pool for the master database.
func (c *Cluster) Write() *pgxpool.Pool {
	return c.Master
}

// ReadSync returns the connection pool for the synchronous replica database.
func (c *Cluster) ReadSync() *pgxpool.Pool {
	return c.Sync
}

// ReadAsync returns the connection pool for the asynchronous replica database.
func (c *Cluster) ReadAsync() *pgxpool.Pool {
	return c.Async
}

// RR returns a connection pool based on a round robin algorithm.
func (c *Cluster) RR() *pgxpool.Pool {
	idx := atomic.AddUint32(&roundRobinIndex, 1) % dbCount
	switch idx {
	case 1:
		return c.Master
	case 2:
		return c.Sync
	default:
		return c.Async
	}
}

// ReadRR returns a read only connection pool based on a round robin algorithm.
func (c *Cluster) ReadRR() *pgxpool.Pool {
	idx := atomic.AddUint32(&roundRobinIndex, 1) % readOnlyDBCount
	if idx == 0 {
		return c.Sync
	}

	return c.Async
}

// Close closes the connections to all databases in the cluster.
// If any of the connections are nil, they will not be closed.
// This method should always be called when the cluster is no longer needed.
func (c *Cluster) Close() {
	if c == nil {
		return
	}

	if c.Master != nil {
		c.Master.Close()
	}

	if c.Sync != nil {
		c.Sync.Close()
	}

	if c.Async != nil {
		c.Async.Close()
	}
}
