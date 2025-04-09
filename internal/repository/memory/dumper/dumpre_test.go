package dumper

import (
	"context"
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
		{
			name:      "access denied",
			file:      "/root/test.txt",
			wantError: true,
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
