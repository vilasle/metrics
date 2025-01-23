package postgresql

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/vilasle/metrics/internal/metric"
)

type getter interface {
	get(context.Context, ...string) ([]metric.Metric, error)
}

type unknownGetter struct{}

func (g *unknownGetter) get(context.Context, ...string) ([]metric.Metric, error) {
	return []metric.Metric{}, metric.ErrUnknownMetricType
}

type counterGetter struct {
	db repeater
}

func (g *counterGetter) get(ctx context.Context, filterName ...string) ([]metric.Metric, error) {
	if len(filterName) == 0 {
		return g.getAll(ctx)
	}
	return g.getByFilter(ctx, filterName...)
}

func (g *counterGetter) getByFilter(ctx context.Context, name ...string) ([]metric.Metric, error) {
	txt := `
		SELECT id, SUM(value) 
		FROM counters 
		WHERE "id" = any($1)
		GROUP BY id
		`

	if r, err := g.db.Query(ctx, txt, name); err == nil {
		return g.parseResult(r)
	} else {
		return []metric.Metric{}, err
	}
}

func (g *counterGetter) getAll(ctx context.Context) ([]metric.Metric, error) {
	txt := `SELECT id, value FROM counters`
	if r, err := g.db.Query(ctx, txt); err == nil {
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
	db repeater
}

func (g *gaugeGetter) get(ctx context.Context, filterName ...string) ([]metric.Metric, error) {
	if len(filterName) == 0 {
		return g.getAll(ctx)
	}
	return g.getByFilter(ctx, filterName...)
}

func (g *gaugeGetter) getByFilter(ctx context.Context, name ...string) ([]metric.Metric, error) {
	txt := `SELECT id, value FROM gauges WHERE "id" = any($1)`
	if r, err := g.db.Query(ctx, txt, name); err == nil {
		return g.parseResult(r)
	} else {
		return []metric.Metric{}, err
	}
}

func (g *gaugeGetter) getAll(ctx context.Context) ([]metric.Metric, error) {
	txt := `SELECT id, value FROM gauges`
	if r, err := g.db.Query(ctx, txt); err == nil {
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
