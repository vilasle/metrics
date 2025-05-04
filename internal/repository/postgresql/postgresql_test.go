package postgresql

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vilasle/metrics/internal/metric"
	"github.com/vilasle/metrics/internal/repository"
)

func Test_unknownGetter_get(t *testing.T) {
	getter := unknownGetter{}

	r, err := getter.get(context.Background())

	assert.Equal(t, metric.ErrUnknownMetricType, err)
	assert.Len(t, r, 0)
}

func Test_counterGetter_get(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err, "can not create sqlmock")

	r := repeater{db: db, repeatSteps: []time.Duration{time.Second}}
	getter := counterGetter{r}

	mock.
		ExpectQuery(`SELECT id, value FROM counters`).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "value"}).
				AddRow("counter1", 1))

	result, err := getter.get(context.Background())

	expected := []metric.Metric{
		metric.NewCounterMetric("counter1", 1),
	}

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func Test_gaugeGetter_get(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err, "can not create sqlmock")

	r := repeater{db: db, repeatSteps: []time.Duration{time.Second}}
	getter := gaugeGetter{r}

	mock.
		ExpectQuery(`SELECT id, value FROM gauges`).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "value"}).
				AddRow("gauge1", 1.123))

	result, err := getter.get(context.Background())

	expected := []metric.Metric{
		metric.NewGaugeMetric("gauge1", 1.123),
	}

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func Test_unknownSaver_save(t *testing.T) {
	saver := unknownSaver{}

	err := saver.save(context.Background(), metric.NewCounterMetric("counter1", 1))

	assert.Equal(t, metric.ErrUnknownMetricType, err)
}

func Test_counterSaver_save(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err, "can not create sqlmock")

	r := repeater{db: db, repeatSteps: []time.Duration{time.Second}}
	s := counterSaver{r}

	mock.ExpectExec(`INSERT INTO counters`).WithArgs("counter1", int64(1)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = s.save(context.Background(), metric.NewCounterMetric("counter1", 1))
	assert.NoError(t, err)
}

func Test_gaugeSaver_save(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err, "can not create sqlmock")

	r := repeater{db: db, repeatSteps: []time.Duration{time.Second}}
	s := gaugeSaver{r}

	mock.ExpectExec(`INSERT INTO gauges`).WithArgs("gauge1", 1.123).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = s.save(context.Background(), metric.NewGaugeMetric("gauge1", 1.123))
	assert.NoError(t, err)
}

type mockMetric struct{}

func (mockMetric) Name() string {
	return "mock"
}

func (mockMetric) Value() string {
	return "1"
}
func (mockMetric) Type() string {
	return "some"
}
func (mockMetric) SetValue(any) error {
	return nil
}
func (mockMetric) AddValue(any) error {
	return nil
}
func (mockMetric) Float64() float64 {
	return 1
}
func (mockMetric) Int64() int64 {
	return 1
}
func (mockMetric) String() string {
	return "mock"
}
func (mockMetric) MarshalJSON() ([]byte, error) {
	return []byte("{}"), nil
}

func TestPostgresqlMetricRepository_Save(t *testing.T) {
	setup := func(mock sqlmock.Sqlmock, metrics []metric.Metric) {
		if len(metrics) > 1 {
			mock.ExpectBegin()
		}
		for _, m := range metrics {
			if m.Type() == metric.TypeCounter {
				mock.ExpectExec("INSERT INTO counters").
					WithArgs(m.Name(), m.Int64()).
					WillReturnResult(sqlmock.NewResult(1, 1))
			} else if m.Type() == metric.TypeGauge {
				mock.ExpectExec("INSERT INTO gauges").
					WithArgs(m.Name(), m.Float64()).
					WillReturnResult(sqlmock.NewResult(1, 1))
			}
		}
		if len(metrics) > 1 {
			mock.ExpectCommit()
			mock.ExpectRollback()

		}
	}
	testCases := []struct {
		name    string
		ctx     context.Context
		metrics []metric.Metric
		want    error
	}{
		{
			name: "success saving counter",
			ctx:  context.Background(),
			metrics: []metric.Metric{
				metric.NewCounterMetric("counter1", 1),
			},
			want: nil,
		},
		{
			name: "success saving gauge",
			ctx:  context.Background(),
			metrics: []metric.Metric{
				metric.NewGaugeMetric("gauge1", 1.123),
			},
			want: nil,
		},
		{
			name: "success batch saving",
			ctx:  context.Background(),
			metrics: []metric.Metric{
				metric.NewGaugeMetric("gauge1", 1.123),
				metric.NewCounterMetric("counter1", 1),
			},
			want: nil,
		},
		{
			name: "failed: unknown metric type",
			ctx:  context.Background(),
			metrics: []metric.Metric{
				mockMetric{},
			},
			want: metric.ErrUnknownMetricType,
		},
		{
			name: "failed batch: one of metrics has unknown metric type",
			ctx:  context.Background(),
			metrics: []metric.Metric{
				metric.NewGaugeMetric("gauge1", 1.123),
				mockMetric{},
			},
			want: metric.ErrUnknownMetricType,
		},
		{
			name:    "failed: empty set",
			ctx:     context.Background(),
			metrics: []metric.Metric{},
			want:    repository.ErrEmptySetOfMetric,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err, "can not create sqlmock")

			setup(mock, tt.metrics)

			r := repeater{db: db, repeatSteps: []time.Duration{time.Second}}

			repo := PostgresqlMetricRepository{r}

			err = repo.Save(tt.ctx, tt.metrics...)
			if tt.want != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.want)
			}
		})

	}

}
