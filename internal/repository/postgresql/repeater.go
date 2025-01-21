package postgresql

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type repeater struct {
	db     *pgxpool.Pool
	repeat []time.Duration
}

func (r repeater) Exec(ctx context.Context, sql string, args ...interface{}) (err error) {
	for _, d := range r.repeat {
		if _, err = r.db.Exec(ctx, sql, args...); err == nil {
			return nil
		}
		time.Sleep(d)
	}
	return err
}

func (r repeater) Query(ctx context.Context, sql string, args ...interface{}) (rows pgx.Rows, err error) {
	for _, d := range r.repeat {
		if rows, err = r.db.Query(ctx, sql, args...); err == nil {
			return rows, nil
		}
		time.Sleep(d)
	}
	return nil, err
}

func (r repeater) Ping(ctx context.Context) (err error) {
	for _, d := range r.repeat {
		if err = r.db.Ping(ctx); err == nil {
			return nil
		}
		time.Sleep(d)
	}
	return err
}

func (r repeater) Close() {
	r.db.Close()
}

func (r repeater) Begin(ctx context.Context) (pgx.Tx, error) {
	return r.db.Begin(ctx)
}
