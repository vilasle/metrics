package compress

import (
	"bytes"
	"compress/gzip"
	"io"
)

type (
	//CompressorWriter implements compression algorithm Flate, compress data and writes to Buffer
	CompressorWriter interface {
		io.Writer
		Bytes() []byte
		Reset()
	}

	compressor struct {
		wrt *gzip.Writer
		buf *bytes.Buffer
	}
)

// NewCompressor returns new instance of CompressorWriter
// allowed levels see gzip.NewWriterLevel
func NewCompressor(level int) CompressorWriter {
	buf := &bytes.Buffer{}
	gw, err := gzip.NewWriterLevel(buf, level)
	if err != nil {
		return nil
	}

	return &compressor{
		wrt: gw,
		buf: buf,
	}
}

// Write compress data and writes to Buffer
func (c *compressor) Write(content []byte) (int, error) {
	n, err := c.wrt.Write(content)
	c.wrt.Close()
	return n, err
}

// Bytes returns compressed data from buffer
func (c *compressor) Bytes() []byte {
	return c.buf.Bytes()
}

// Reset clears buffer from data
func (c *compressor) Reset() {
	c.wrt.Reset(c.buf)
	c.buf.Reset()
}
