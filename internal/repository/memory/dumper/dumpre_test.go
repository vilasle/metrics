package dumper

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vilasle/metrics/internal/metric"
)

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

func Test_FileStream_NewFileStream(t *testing.T) {
	file := "test.txt"
	fs, err := NewFileStream(file)

	require.NoError(t, err)
	require.NotNil(t, fs)
	require.NoError(t, fs.Close())

	_, err = os.Stat(file)
	require.NoError(t, err)

	exists := !os.IsNotExist(err)

	require.Equal(t, true, exists)

	os.RemoveAll(file)
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

func Test_FileStream_ScanAll(t *testing.T) {}

func Test_FileStream_Clear(t *testing.T) {}
