package compress

import (
	"bytes"
	"compress/gzip"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCompressor(t *testing.T) {
	type args struct {
		level int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "NewCompressor",
			args: args{
				level: gzip.BestCompression,
			},
			wantErr: false,
		},
		{
			name: "NewCompressor: Wrong level",
			args: args{
				level: -3,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewCompressor(tt.args.level)
			if tt.wantErr {
				assert.Equal(t, got, nil)
			} else {
				assert.NotEqual(t, got, nil)
			}
		})
	}
}

func Test_compressor_Write(t *testing.T) {
	type fields struct {
		Writer *gzip.Writer
		Buffer *bytes.Buffer
	}
	type args struct {
		content []byte
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wroteLen   int
		wroteBytes []byte
	}{
		{
			name: "Compress data",
			fields: fields{
				Writer: gzip.NewWriter(nil),
				Buffer: bytes.NewBuffer(nil),
			},
			args: args{
				content: []byte("test"),
			},
			wroteLen: 4,
			wroteBytes: []byte{
				31, 139, 8, 0, 0, 0, 0, 0,
				2, 255, 42, 73, 45, 46,
				1, 4, 0, 0, 255, 255, 12,
				126, 127, 216, 4, 0, 0, 0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			buf := &bytes.Buffer{}
			gw, err := gzip.NewWriterLevel(buf, gzip.BestCompression)
			require.NoError(t, err)

			c := &compressor{
				Writer: gw,
				Buffer: buf,
			}

			got, err := c.Write(tt.args.content)
			require.NoError(t, err)

			assert.Equal(t, got, tt.wroteLen)
			assert.Equal(t, c.Bytes(), tt.wroteBytes)
			c.Reset()

			assert.Equal(t, buf.Len(), 0)
			

		})
	}
}
