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

//TODO add godoc
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

//TODO add godoc
func (c *compressor) Write(content []byte) (int, error) {
	n, err := c.Writer.Write(content)
	c.Writer.Close()
	return n, err
}

//TODO add godoc
func (c *compressor) Bytes() []byte {
	return c.Buffer.Bytes()
}

//TODO add godoc
func (c *compressor) Reset() {
	c.Writer.Reset(c.Buffer)
	c.Buffer.Reset()
}
