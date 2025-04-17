package postgresql

import (
	"context"
	"database/sql"
	"time"
)

type repeater struct {
	db          *sql.DB
	repeatSteps []time.Duration
}

func (r repeater) repeat(fn func() error) (err error) {
	for _, d := range r.repeatSteps {
		if err = fn(); err == nil {
			return nil
		}
		time.Sleep(d)
	}
	return err
}

func (r repeater) Exec(ctx context.Context, sql string, args ...interface{}) (err error) {
	return r.repeat(func() error {
		_, err := r.db.ExecContext(ctx, sql, args...)
		return err
	})
}

func (r repeater) Query(ctx context.Context, sql string, args ...interface{}) (rows *sql.Rows, err error) {
	r.repeat(func() error {
		rows, err = r.db.QueryContext(ctx, sql, args...)
		
		return err
	})

	return rows, err
}

func (r repeater) Ping(ctx context.Context) (err error) {
	return r.repeat(func() error {
		return r.db.PingContext(ctx)
	})
}

func (r repeater) Close() {
	r.db.Close()
}

func (r repeater) Begin(ctx context.Context) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, nil)
}
