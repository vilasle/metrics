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

func (r repeater) exec(ctx context.Context, sql string, args ...interface{}) (err error) {
	return r.repeat(func() error {
		_, err := r.db.ExecContext(ctx, sql, args...)
		return err
	})
}

func (r repeater) query(ctx context.Context, sql string, args ...interface{}) (rows *sql.Rows, err error) {
	r.repeat(func() error {
		rows, err = r.db.QueryContext(ctx, sql, args...)
		if err == nil && rows.Err() != nil {
			err = rows.Err()
		}
		return err
	})

	return rows, err
}

func (r repeater) ping(ctx context.Context) (err error) {
	return r.repeat(func() error {
		return r.db.PingContext(ctx)
	})
}

func (r repeater) close() {
	r.db.Close()
}

func (r repeater) begin(ctx context.Context) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, nil)
}
