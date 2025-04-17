package dumper

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vilasle/metrics/internal/metric"
)

func Test_FileStream_NewFileStream(t *testing.T) {
	testCases := []struct {
		name      string
		file      string
		wantError bool
	}{
		{
			name:      "not existed file",
			file:      "test.txt",
			wantError: false,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			fs, err := NewFileStream(tt.file)
			if tt.wantError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, fs)

			require.NoError(t, fs.Close())

			_, err = os.Stat(tt.file)
			require.NoError(t, err)

			require.Equal(t, true, !os.IsNotExist(err))

			os.RemoveAll(tt.file)
		})
	}

}

func Test_FileStream_Write(t *testing.T) {
	file := "test.txt"
	fs, err := NewFileStream(file)
	//check create
	require.NoError(t, err)
	require.NotNil(t, fs)

	msg := []byte("test message\n and another message\n\t")

	n, err := fs.Write(msg)

	require.NoError(t, err)
	assert.Equal(t, len(msg), n)

	require.NoError(t, fs.Close())

	content, err := os.ReadFile(file)
	require.NoError(t, err)
	assert.Equal(t, msg, content)

	os.RemoveAll(file)
}

func Test_FileStream_Rewrite(t *testing.T) {
	file := "test.txt"
	fs, err := NewFileStream(file)
	//check create
	require.NoError(t, err)
	require.NotNil(t, fs)

	msg := []byte("test message\n and another message\n\t")

	for i := 0; i < 10; i++ {
		n, err := fs.Write(msg)
		require.NoError(t, err)
		assert.Equal(t, len(msg), n)
	}

	content, err := os.ReadFile(file)
	require.NoError(t, err)
	assert.NotEqual(t, msg, content)

	n, err := fs.Rewrite(msg)
	require.NoError(t, err)
	assert.Equal(t, len(msg), n)

	require.NoError(t, fs.Close())

	content, err = os.ReadFile(file)
	require.NoError(t, err)
	assert.Equal(t, msg, content)

	os.RemoveAll(file)
}

func Test_FileStream_ScanAll(t *testing.T) {
	filename := "test.txt"
	content := []string{"0;gauge1;123.123", "1;counter1;123"}
	lines := make([]string, 0, 20)
	for i := 0; i < 10; i++ {
		lines = append(lines, content[0], content[1])
	}

	fd, err := os.Create(filename)
	require.NoError(t, err)

	for _, line := range lines {
		wrLn := []byte(line)
		wrLn = append(wrLn, byte('\n'))
		_, err = fd.Write(wrLn)
		require.NoError(t, err)
	}

	require.NoError(t, fd.Close())

	fs, err := NewFileStream(filename)
	require.NoError(t, err)
	require.NotNil(t, fs)

	result, err := fs.ScanAll()
	require.NoError(t, err)

	for i := range result {
		assert.Equal(t, lines[i], result[i])
	}

	require.NoError(t, fs.Close())

	require.NoError(t, os.RemoveAll(filename))

}
func Test_FileStream_Clear(t *testing.T) {
	testCases := []struct {
		name      string
		filename  string
		getFile   func(filename string) (*os.File, error)
		wantError bool
	}{
		{
			name:     "normal situation",
			filename: "test.txt",
			getFile: func(filename string) (*os.File, error) {
				fd, err := os.Create(filename)
				if err != nil {
					return nil, err
				}
				//file will not be empty
				_, err = fd.Write([]byte("test"))
				if err == nil {
					err = fd.Sync()
				}
				return fd, err
			},
			wantError: false,
		},
		{
			name:     "nil pointer between file pointer",
			filename: "test.txt",
			getFile: func(filename string) (*os.File, error) {
				return nil, nil
			},
			wantError: true,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			file, err := tt.getFile(tt.filename)
			require.NoError(t, err)

			fs := File{
				fd: file,
				mx: &sync.Mutex{},
			}
			err = fs.Clear()
			if tt.wantError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NoError(t, file.Close())

			stat, err := os.Stat(tt.filename)
			require.NoError(t, err)
			assert.Equal(t, int64(0), stat.Size())

			require.NoError(t, os.Remove(tt.filename))
		})
	}

}

func Test_dumpedMetric_dumpedContent(t *testing.T) {
	tests := []struct {
		name   string
		metric dumpedMetric
		want   []byte
	}{
		{
			name: "dump counter",
			metric: dumpedMetric{
				metric.NewCounterMetric("counter1", 123),
			},
			want: []byte("1;counter1;123\n"),
		},
		{
			name: "dump gauge",
			metric: dumpedMetric{
				metric.NewGaugeMetric("gauge1", 123.123),
			},
			want: []byte("0;gauge1;123.123\n"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.metric.dumpedContent())
		})
	}
}

func Test_FileDumper_NewFileDumper(t *testing.T) {
}

func Test_FileDumper_Save(t *testing.T) {
	storageErr := errors.New("storage error")
	fsErr := errors.New("file stream error")
	setupStorage := func(mock *MockMetricRepository, ctx context.Context, metrics []metric.Metric, err error) {
		mock.EXPECT().Save(ctx, metrics).Return(err)
	}

	setupWriter := func(mock *MockSerialWriter, input []byte, n int, err error) {
		mock.EXPECT().Write(input).Return(n, err)
	}

	type saveArgs struct {
		metrics []metric.Metric

		err error
	}

	type writeArgs struct {
		content []byte
		n       int
		err     error
	}

	tmpMetrics := []metric.Metric{
		metric.NewCounterMetric("counter1", 123),
	}

	testCases := []struct {
		name     string
		syncSave bool
		*saveArgs
		*writeArgs
		ctx     context.Context
		metrics []metric.Metric
		err     error
	}{
		{
			name:     "success, no sync mode",
			syncSave: false,
			saveArgs: &saveArgs{
				metrics: tmpMetrics,
				err:     storageErr,
			},
			metrics: tmpMetrics,
			err:     storageErr,
		},
		{
			name:     "storage failed, no sync mode",
			syncSave: false,
			saveArgs: &saveArgs{
				metrics: tmpMetrics,
				err:     nil,
			},
			metrics: tmpMetrics,
			err:     nil,
		},
		{
			name:     "success, sync mode",
			syncSave: true,
			saveArgs: &saveArgs{
				metrics: tmpMetrics,
				err:     nil,
			},
			writeArgs: &writeArgs{
				content: []byte("1;counter1;123\n"),
				n:       1,
				err:     nil,
			},
			metrics: tmpMetrics,
			err:     nil,
		},
		{
			name:     "serial writer failed, sync mode",
			syncSave: true,
			saveArgs: &saveArgs{
				metrics: tmpMetrics,
				err:     nil,
			},
			writeArgs: &writeArgs{
				content: []byte("1;counter1;123\n"),
				n:       1,
				err:     fsErr,
			},
			metrics: tmpMetrics,
			err:     fsErr,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := NewMockMetricRepository(ctrl)
			fs := NewMockSerialWriter(ctrl)

			if tt.saveArgs != nil {
				setupStorage(repo, tt.ctx, tt.saveArgs.metrics, tt.saveArgs.err)
			}

			if tt.writeArgs != nil {
				setupWriter(fs, tt.writeArgs.content, tt.writeArgs.n, tt.writeArgs.err)
			}

			fd := FileDumper{
				fs:       fs,
				storage:  repo,
				syncSave: tt.syncSave,
				srvMx:    &sync.Mutex{},
			}

			err := fd.Save(tt.ctx, tt.metrics...)

			if tt.err != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}

}

func Test_FileDumper_DumpAll(t *testing.T) {
	setupWriter := func(mock *MockSerialWriter, input []byte, n int, err error) {
		mock.EXPECT().Rewrite(input).Return(n, err)
	}

	setupStorage := func(mock *MockMetricRepository, ctx context.Context, mtype string, metrics []metric.Metric, err error) {
		mock.EXPECT().Get(ctx, mtype).Return(metrics, err)
	}

	storageErr := errors.New("storage error")
	fsErr := errors.New("file stream error")

	type storageArgs struct {
		mtype   string
		metrics []metric.Metric
		err     error
	}

	type writerArgs struct {
		content []byte
		n       int
		err     error
	}

	testsCases := []struct {
		name        string
		ctx         context.Context
		storageArgs []storageArgs
		writerArgs  []writerArgs
		want        error
	}{
		{
			name: "success dumping",
			want: nil,
			ctx:  context.Background(),
			storageArgs: []storageArgs{
				{
					mtype: metric.TypeGauge,
					metrics: []metric.Metric{
						metric.NewGaugeMetric("gauge1", 123.123),
					},
					err: nil,
				},
				{
					mtype: metric.TypeCounter,
					metrics: []metric.Metric{
						metric.NewCounterMetric("counter1", 123),
					},
					err: nil,
				},
			},
			writerArgs: []writerArgs{
				{
					content: []byte("0;gauge1;123.123\n1;counter1;123\n"),
					n:       1,
					err:     nil,
				},
			},
		},
		{
			name: "failed getting metrics",
			want: storageErr,
			ctx:  context.Background(),
			storageArgs: []storageArgs{
				{
					mtype: metric.TypeGauge,
					metrics: []metric.Metric{
						metric.NewGaugeMetric("gauge1", 123.123),
					},
					err: nil,
				},
				{
					mtype:   metric.TypeCounter,
					metrics: []metric.Metric{},
					err:     storageErr,
				},
			},
			writerArgs: []writerArgs{},
		},
		{
			name: "failed write metrics",
			want: fsErr,
			ctx:  context.Background(),
			storageArgs: []storageArgs{
				{
					mtype: metric.TypeGauge,
					metrics: []metric.Metric{
						metric.NewGaugeMetric("gauge1", 123.123),
					},
					err: nil,
				},
				{
					mtype: metric.TypeCounter,
					metrics: []metric.Metric{
						metric.NewCounterMetric("counter1", 123),
					},
					err: nil,
				},
			},
			writerArgs: []writerArgs{
				{
					content: []byte("0;gauge1;123.123\n1;counter1;123\n"),
					n:       1,
					err:     fsErr,
				},
			},
		},
	}

	for _, tt := range testsCases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := NewMockMetricRepository(ctrl)

			for _, args := range tt.storageArgs {
				setupStorage(repo, tt.ctx, args.mtype, args.metrics, args.err)
			}

			fs := NewMockSerialWriter(ctrl)

			for _, args := range tt.writerArgs {
				setupWriter(fs, args.content, args.n, args.err)
			}

			fd := FileDumper{
				storage: repo,
				fs:      fs,
				srvMx:   &sync.Mutex{},
			}

			assert.Equal(t, tt.want, fd.DumpAll(tt.ctx))
		})
	}

}

func Test_FileDumper_Get(t *testing.T) {
	ctx := context.Background()
	metricType := metric.TypeGauge
	filter := "gauge1"

	result := []metric.Metric{
		metric.NewGaugeMetric("gauge1", 123.123),
	}
	var resultErr error = nil

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := NewMockMetricRepository(ctrl)
	repo.EXPECT().Get(ctx, metricType, filter).Return(result, resultErr)

	fd := FileDumper{
		storage: repo,
	}

	r, err := fd.Get(ctx, metricType, filter)

	assert.Equal(t, resultErr, err)
	assert.Equal(t, result, r)
}

func Test_FileDumper_Ping(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := NewMockMetricRepository(ctrl)
	repo.EXPECT().Ping(ctx).Return(nil)

	fd := FileDumper{
		storage: repo,
	}
	assert.NoError(t, fd.Ping(ctx))
}

func Test_FileDumper_Close(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := NewMockMetricRepository(ctrl)
	repo.EXPECT().Close()

	fd := FileDumper{
		storage: repo,
	}

	fd.Close()
}

func Test_FileDumper_all(t *testing.T) {
	gaugeErr := errors.New("gauge error")
	counterErr := errors.New("counter error")

	setup := func(r *MockMetricRepository, ctx context.Context, mtype string, res []metric.Metric, err error) {
		r.EXPECT().Get(ctx, mtype).Return(res, err)
	}

	testCases := []struct {
		name       string
		gauges     []metric.Metric
		counters   []metric.Metric
		result     []metric.Metric
		ctx        context.Context
		err        error
		gaugeErr   error
		counterErr error
	}{
		{
			name: "success",
			gauges: []metric.Metric{
				metric.NewGaugeMetric("gauge1", 123.123), metric.NewGaugeMetric("gauge2", 321.321),
			},
			counters: []metric.Metric{
				metric.NewCounterMetric("counter1", 123), metric.NewCounterMetric("counter2", 321),
			},
			result: []metric.Metric{
				metric.NewGaugeMetric("gauge1", 123.123), metric.NewGaugeMetric("gauge2", 321.321),
				metric.NewCounterMetric("counter1", 123), metric.NewCounterMetric("counter2", 321),
			},
			ctx:        context.Background(),
			err:        nil,
			gaugeErr:   nil,
			counterErr: nil,
		},
		{
			name:       "getting gauges failed",
			gauges:     nil,
			counters:   nil,
			result:     nil,
			ctx:        context.Background(),
			err:        gaugeErr,
			gaugeErr:   gaugeErr,
			counterErr: nil,
		},
		{
			name: "getting counters failed",
			gauges: []metric.Metric{
				metric.NewGaugeMetric("gauge1", 123.123), metric.NewGaugeMetric("gauge2", 321.321),
			},
			counters:   nil,
			result:     nil,
			ctx:        context.Background(),
			err:        counterErr,
			gaugeErr:   nil,
			counterErr: counterErr,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := NewMockMetricRepository(ctrl)

			setup(repo, tt.ctx, metric.TypeGauge, tt.gauges, tt.gaugeErr)

			if tt.gaugeErr == nil {
				setup(repo, tt.ctx, metric.TypeCounter, tt.counters, tt.counterErr)
			}

			fd := &FileDumper{storage: repo}

			r, err := fd.all(tt.ctx)

			assert.Equal(t, tt.err, err)
			assert.Equal(t, tt.result, r)
		})
	}

}

func Test_withClear(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	fs := NewMockSerialWriter(ctrl)

	fd := &FileDumper{
		fs: fs,
	}

	ctx := context.Background()

	fs.EXPECT().Clear().Return(nil)

	assert.NoError(t, withClear(ctx, fd))
}

func Test_FileDumper_restore(t *testing.T) {
	//fs scan all setup
	setupFs := func(mock *MockSerialWriter, result []string, err error) {
		mock.EXPECT().ScanAll().Return(result, err)
	}
	//storage save setup
	setupStorage := func(mock *MockMetricRepository, ctx context.Context, metrics []metric.Metric, err error) {
		mock.EXPECT().Save(ctx, metrics).Return(err)
	}

	ctx := context.Background()

	type writerArg struct {
		result []string
		err    error
	}

	type storageArg struct {
		metrics []metric.Metric
		err     error
	}

	testCase := []struct {
		name string
		writerArg
		storageArgs []storageArg
		want        error
	}{
		{
			name: "success",
			writerArg: writerArg{
				result: []string{"0;gauge1;123.123", "1;counter1;321"},
				err:    nil,
			},
			storageArgs: []storageArg{
				{
					metrics: []metric.Metric{
						metric.NewGaugeMetric("gauge1", 123.123),
					},
					err: nil,
				},
				{
					metrics: []metric.Metric{
						metric.NewCounterMetric("counter1", 321),
					},
					err: nil,
				},
			},
		},
	}

	for _, tt := range testCase {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			fs := NewMockSerialWriter(ctrl)
			storage := NewMockMetricRepository(ctrl)

			setupFs(fs, tt.result, tt.err)

			for _, arg := range tt.storageArgs {
				setupStorage(storage, ctx, arg.metrics, arg.err)
			}

			fd := &FileDumper{
				fs:      fs,
				storage: storage,
			}

			err := fd.restore(ctx)

			assert.Equal(t, tt.want, err)
		})
	}
}
