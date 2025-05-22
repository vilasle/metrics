package http

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"sync"

	"github.com/vilasle/metrics/internal/compress"
)

type WriterOption func(*JSONWriter)

func WithCompressing() WriterOption {
	return func(e *JSONWriter) {
		e.wrt = newGzipWriter(e.wrt)
		e.headers["Content-Encoding"] = "gzip"
	}
}

func WithEncryption(key *rsa.PublicKey) WriterOption {
	return func(e *JSONWriter) {
		if key == nil {
			return
		}
		e.wrt = newEncryptWriter(e.wrt, key)
	}
}

func WithCalculateHashSum(hashSumKey []byte) WriterOption {
	return func(e *JSONWriter) {
		if len(hashSumKey) == 0 {
			return
		}
		e.wrt = newHashSumWriter(e, hashSumKey)
	}
}

type JSONWriter struct {
	headers map[string]string
	buf     *bytes.Buffer
	wrt     io.Writer
}

func NewJSONWriter(opts ...WriterOption) *JSONWriter {
	buf := &bytes.Buffer{}
	e := &JSONWriter{
		buf:     buf,
		wrt:     buf,
		headers: make(map[string]string),
	}

	e.headers["Content-Type"] = "application/json"

	for _, opt := range opts {
		opt(e)
	}

	return e
}

func (e JSONWriter) Write(object any) error {
	content, err := json.Marshal(&object)
	if err != nil {
		return err
	}

	e.buf.Reset()

	_, err = e.wrt.Write(content)
	return err
}

func (e JSONWriter) Bytes() []byte {
	return e.buf.Bytes()
}

type gzipWriter struct {
	wrt          io.Writer
	encodersPool *sync.Pool
}

func newGzipWriter(wrt io.Writer) *gzipWriter {
	return &gzipWriter{
		wrt: wrt,
		encodersPool: &sync.Pool{
			New: func() interface{} {
				return compress.NewCompressor(gzip.BestCompression)
			},
		},
	}
}

func (e gzipWriter) Write(d []byte) (int, error) {
	w := e.encodersPool.Get().(compress.CompressorWriter)
	n, err := w.Write(d)
	if err != nil {
		return n, err
	}
	content := w.Bytes()

	w.Reset()
	e.encodersPool.Put(w)

	return e.wrt.Write(content)
}

type encryptWriter struct {
	key *rsa.PublicKey
	wrt io.Writer
}

func newEncryptWriter(wrt io.Writer, key *rsa.PublicKey) *encryptWriter {
	return &encryptWriter{
		key: key,
		wrt: wrt,
	}
}

func (e encryptWriter) Write(d []byte) (int, error) {
	content, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, e.key, d, []byte{})
	if err != nil {
		return 0, err
	}
	return e.wrt.Write(content)
}

type hashSumWriter struct {
	key []byte
	wrt io.Writer
	jw  *JSONWriter
}

func newHashSumWriter(jw *JSONWriter, key []byte) *hashSumWriter {
	return &hashSumWriter{
		key: key,
		wrt: jw.wrt,
		jw:  jw,
	}
}

func (e hashSumWriter) Write(d []byte) (int, error) {
	w := hmac.New(sha256.New, []byte(e.key))
	if _, err := w.Write(d); err != nil {
		return 0, err
	}

	srcHash := w.Sum(nil)
	hash := base64.URLEncoding.EncodeToString(srcHash)

	e.jw.headers["HashSHA256"] = hash

	return e.wrt.Write(d)
}
