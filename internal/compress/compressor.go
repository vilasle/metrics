package compress

import (
	"bytes"
	"compress/gzip"
	"io"
)

type (
	CompressorWriter interface {
		io.Writer
		Bytes() []byte
		Reset()
	}

	compressor struct {
		*gzip.Writer
		*bytes.Buffer
	}
)

func NewCompressor(level int) CompressorWriter {
	buf := &bytes.Buffer{}
	gw, err := gzip.NewWriterLevel(buf, level)
	if err != nil {
		return nil
	}

	return &compressor{
		Writer: gw,
		Buffer: buf,
	}
}

func (c *compressor) Write(content []byte) (int, error) {
	n, err := c.Writer.Write(content)
	c.Writer.Close()
	return n, err
}

func (c *compressor) Bytes() []byte {
	return c.Buffer.Bytes()
}

func (c *compressor) Reset() {
	c.Writer.Reset(c.Buffer)
	c.Buffer.Reset()
}
