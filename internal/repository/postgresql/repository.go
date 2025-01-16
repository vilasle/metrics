package postgresql

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vilasle/metrics/internal/metric"
)

type PostgresqlMetricRepository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *PostgresqlMetricRepository {
	return &PostgresqlMetricRepository{db: db}
}

func (r *PostgresqlMetricRepository) Save(entity metric.Metric) error {
	//TODO implement it
	return nil 

}

func (r *PostgresqlMetricRepository) Get(metricType string, filterName ...string) ([]metric.Metric, error) {
	//TODO implement it
	return []metric.Metric{}, nil
}

func (r *PostgresqlMetricRepository) Ping(ctx context.Context) error {
	return r.db.Ping(ctx)
}

func (r *PostgresqlMetricRepository) Close() {
	r.db.Close()
}
