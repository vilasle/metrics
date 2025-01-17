package postgresql

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vilasle/metrics/internal/metric"
)

type getter interface {
	get(...string) ([]metric.Metric, error)
}

type unknownGetter struct{}

func (g *unknownGetter) get(...string) ([]metric.Metric, error) {
	return []metric.Metric{}, metric.ErrUnknownMetricType
}

type counterGetter struct {
	db *pgxpool.Pool
}

func (g *counterGetter) get(filterName ...string) ([]metric.Metric, error) {
	if len(filterName) == 0 {
		return g.getAll()
	}
	return g.getByFilter(filterName...)
}

func (g *counterGetter) getByFilter(name ...string) ([]metric.Metric, error) {
	txt := `
		SELECT id, SUM(value) 
		FROM counters 
		WHERE "id" = any($1)
		GROUP BY id
		`

	if r, err := g.db.Query(context.TODO(), txt, name); err == nil {
		return g.parseResult(r)
	} else {
		return []metric.Metric{}, err
	}
}

func (g *counterGetter) getAll() ([]metric.Metric, error) {
	txt := `SELECT id, value FROM counters`
	if r, err := g.db.Query(context.TODO(), txt); err == nil {
		return g.parseResult(r)
	} else {
		return []metric.Metric{}, err
	}
}

func (g *counterGetter) parseResult(rows pgx.Rows) ([]metric.Metric, error) {
	rs := make([]metric.Metric, 0)
	for rows.Next() {
		var name string
		var value int64
		if err := rows.Scan(&name, &value); err != nil {
			return nil, err
		}
		rs = append(rs, metric.NewCounterMetric(name, value))
	}
	return rs, nil

}

type gaugeGetter struct {
	db *pgxpool.Pool
}

func (g *gaugeGetter) get(filterName ...string) ([]metric.Metric, error) {
	if len(filterName) == 0 {
		return g.getAll()
	}
	return g.getByFilter(filterName...)
}

func (g *gaugeGetter) getByFilter(name ...string) ([]metric.Metric, error) {
	txt := `SELECT id, value FROM gauges WHERE "id" = any($1)`
	if r, err := g.db.Query(context.TODO(), txt, name); err == nil {
		return g.parseResult(r)
	} else {
		return []metric.Metric{}, err
	}
}

func (g *gaugeGetter) getAll() ([]metric.Metric, error) {
	txt := `SELECT id, value FROM gauges`
	if r, err := g.db.Query(context.TODO(), txt); err == nil {
		return g.parseResult(r)
	} else {
		return []metric.Metric{}, err
	}
}

func (g *gaugeGetter) parseResult(rows pgx.Rows) ([]metric.Metric, error) {
	rs := make([]metric.Metric, 0)
	for rows.Next() {
		var name string
		var value float64
		if err := rows.Scan(&name, &value); err != nil {
			return nil, err
		}
		rs = append(rs, metric.NewGaugeMetric(name, value))
	}
	return rs, nil

}
