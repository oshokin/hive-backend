package pgx_pool_prometheus

import (
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
)

type (
	// Stater is a provider of the Stat() function. Implemented by pgxpool.Pool.
	Stater interface {
		Stat() *pgxpool.Stat
	}

	// Collector is a prometheus.Collector that will collect the nine statistics produced by pgxpool.Stat.
	Collector struct {
		staters map[string]Stater

		acquireCountDesc         *prometheus.Desc
		acquireDurationDesc      *prometheus.Desc
		acquiredConnsDesc        *prometheus.Desc
		canceledAcquireCountDesc *prometheus.Desc
		constructingConnsDesc    *prometheus.Desc
		emptyAcquireCountDesc    *prometheus.Desc
		idleConnsDesc            *prometheus.Desc
		maxConnsDesc             *prometheus.Desc
		totalConnsDesc           *prometheus.Desc
		newConnsCount            *prometheus.Desc
		maxLifetimeDestroyCount  *prometheus.Desc
		maxIdleDestroyCount      *prometheus.Desc
	}
)

const poolNameLabelTag = "db"

// NewCollector creates a new Collector to collect stats from PGX pools.
func NewCollector(staters map[string]Stater, labels prometheus.Labels) *Collector {
	poolNameLabel := prometheus.UnconstrainedLabels{poolNameLabelTag}

	return &Collector{
		staters: staters,
		acquireCountDesc: prometheus.NewDesc(
			"pgxpool_acquire_count",
			"Cumulative count of successful acquires from the pool.",
			poolNameLabel,
			labels),
		acquireDurationDesc: prometheus.NewDesc(
			"pgxpool_acquire_duration_ns",
			"Total duration of all successful acquires from the pool in nanoseconds.",
			poolNameLabel,
			labels),
		acquiredConnsDesc: prometheus.NewDesc(
			"pgxpool_acquired_conns",
			"Number of currently acquired connections in the pool.",
			poolNameLabel,
			labels),
		canceledAcquireCountDesc: prometheus.NewDesc(
			"pgxpool_canceled_acquire_count",
			"Cumulative count of acquires from the pool that were canceled by a context.",
			poolNameLabel,
			labels),
		constructingConnsDesc: prometheus.NewDesc(
			"pgxpool_constructing_conns",
			"Number of conns with construction in progress in the pool.",
			poolNameLabel,
			labels),
		emptyAcquireCountDesc: prometheus.NewDesc(
			"pgxpool_empty_acquire",
			"Cumulative count of successful acquires from the pool that waited for "+
				"a resource to be released or constructed because the pool was empty.",
			poolNameLabel,
			labels),
		idleConnsDesc: prometheus.NewDesc(
			"pgxpool_idle_conns",
			"Number of currently idle conns in the pool.",
			poolNameLabel,
			labels),
		maxConnsDesc: prometheus.NewDesc(
			"pgxpool_max_conns",
			"Maximum size of the pool.",
			poolNameLabel,
			labels),
		totalConnsDesc: prometheus.NewDesc(
			"pgxpool_total_conns",
			"Total number of resources currently in the pool. "+
				"The value is the sum of ConstructingConns, AcquiredConns, and IdleConns.",
			poolNameLabel,
			labels),
		newConnsCount: prometheus.NewDesc(
			"pgxpool_new_conns_count",
			"Cumulative count of new connections opened.",
			poolNameLabel,
			labels),
		maxLifetimeDestroyCount: prometheus.NewDesc(
			"pgxpool_max_lifetime_destroy_count",
			"Cumulative count of connections destroyed because they exceeded MaxConnLifetime. ",
			poolNameLabel,
			labels),
		maxIdleDestroyCount: prometheus.NewDesc(
			"pgxpool_max_idle_destroy_count",
			"Cumulative count of connections destroyed because they exceeded MaxConnIdleTime.",
			poolNameLabel,
			labels),
	}
}

// Describe implements the prometheus.Collector interface.
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(c, ch)
}

// Collect implements the prometheus.Collector interface.
func (c *Collector) Collect(metrics chan<- prometheus.Metric) {
	var wg sync.WaitGroup

	wg.Add(len(c.staters))

	for name, stater := range c.staters {
		go func(name string, stater Stater) {
			defer wg.Done()
			c.collectFromPool(name, stater, metrics)
		}(name, stater)
	}

	wg.Wait()
}

func (c *Collector) collectFromPool(poolName string, p Stater, metrics chan<- prometheus.Metric) {
	stats := p.Stat()
	metrics <- prometheus.MustNewConstMetric(
		c.acquireCountDesc,
		prometheus.CounterValue,
		float64(stats.AcquireCount()),
		poolName,
	)
	metrics <- prometheus.MustNewConstMetric(
		c.acquireDurationDesc,
		prometheus.CounterValue,
		float64(stats.AcquireDuration()),
		poolName,
	)
	metrics <- prometheus.MustNewConstMetric(
		c.acquiredConnsDesc,
		prometheus.GaugeValue,
		float64(stats.AcquiredConns()),
		poolName,
	)
	metrics <- prometheus.MustNewConstMetric(
		c.canceledAcquireCountDesc,
		prometheus.CounterValue,
		float64(stats.CanceledAcquireCount()),
		poolName,
	)
	metrics <- prometheus.MustNewConstMetric(
		c.constructingConnsDesc,
		prometheus.GaugeValue,
		float64(stats.ConstructingConns()),
		poolName,
	)
	metrics <- prometheus.MustNewConstMetric(
		c.emptyAcquireCountDesc,
		prometheus.CounterValue,
		float64(stats.EmptyAcquireCount()),
		poolName,
	)
	metrics <- prometheus.MustNewConstMetric(
		c.idleConnsDesc,
		prometheus.GaugeValue,
		float64(stats.IdleConns()),
		poolName,
	)
	metrics <- prometheus.MustNewConstMetric(
		c.maxConnsDesc,
		prometheus.GaugeValue,
		float64(stats.MaxConns()),
		poolName,
	)
	metrics <- prometheus.MustNewConstMetric(
		c.totalConnsDesc,
		prometheus.GaugeValue,
		float64(stats.TotalConns()),
		poolName,
	)
	metrics <- prometheus.MustNewConstMetric(
		c.newConnsCount,
		prometheus.CounterValue,
		float64(stats.NewConnsCount()),
		poolName,
	)
	metrics <- prometheus.MustNewConstMetric(
		c.maxLifetimeDestroyCount,
		prometheus.CounterValue,
		float64(stats.MaxLifetimeDestroyCount()),
		poolName,
	)
	metrics <- prometheus.MustNewConstMetric(
		c.maxIdleDestroyCount,
		prometheus.CounterValue,
		float64(stats.MaxIdleDestroyCount()),
		poolName,
	)
}
