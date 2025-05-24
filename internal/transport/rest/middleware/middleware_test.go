package middleware

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vilasle/metrics/internal/compress"
	"github.com/vilasle/metrics/internal/logger"
)

func Test_WithCompress(t *testing.T) {
	handler := WithCompress("application/json", "text/html")
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message": "hello world"}`))
		w.WriteHeader(http.StatusOK)
	})

	resp := httptest.NewRecorder()

	want := []byte{31, 139, 8, 0, 0, 0, 0, 0, 2, 255, 170, 86, 202, 77, 45, 46, 78, 76, 79, 85, 178, 82, 80, 202, 72, 205, 201, 201, 87, 40, 207, 47, 202, 73, 81, 170, 5, 4, 0, 0, 255, 255, 48, 91, 218, 238, 26, 0, 0, 0}

	req, _ := http.NewRequest("POST", "/", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept-Encoding", "gzip")

	handler(next).ServeHTTP(resp, req)

	got := resp.Body.Bytes()

	assert.Equal(t, want, got)
}

func Test_Compress(t *testing.T) {
	handler := Compress("application/json", "text/html")
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message": "hello world"}`))
		w.WriteHeader(http.StatusOK)
	})

	resp := httptest.NewRecorder()

	want := []byte{31, 139, 8, 0, 0, 0, 0, 0, 2, 255, 170, 86, 202, 77, 45, 46, 78, 76, 79, 85, 178, 82, 80, 202, 72, 205, 201, 201, 87, 40, 207, 47, 202, 73, 81, 170, 5, 4, 0, 0, 255, 255, 48, 91, 218, 238, 26, 0, 0, 0}

	req, _ := http.NewRequest("POST", "/", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept-Encoding", "gzip")

	handler(next).ServeHTTP(resp, req)

	got := resp.Body.Bytes()

	assert.Equal(t, want, got)
}

func Test_WithUnpackBody(t *testing.T) {
	//source content
	srcContent := []byte(`{"message": "hello world"}`)

	//create hash key and generate private key
	hashKey := []byte("hashKey")
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	require.NoError(t, err)

	//create unpacker chain
	chain := NewUnpackerChain(
		CheckHashSum(hashKey),
		DecryptContent(privateKey, "update", "updates"),
		DecompressContent("gzip"),
	)
	handler := WithUnpackBody(chain)

	//compress content
	wc := compress.NewCompressor(gzip.BestCompression)
	_, err = wc.Write(srcContent)
	require.NoError(t, err)

	compressedContent := wc.Bytes()

	//encrypt content
	content, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, &privateKey.PublicKey, compressedContent, []byte{})
	require.NoError(t, err)

	//calculate hash sum
	hasher := hmac.New(sha256.New, []byte(hashKey))
	_, err = hasher.Write(content)
	require.NoError(t, err)

	srcHash := hasher.Sum(nil)
	hash := base64.URLEncoding.EncodeToString(srcHash)

	//create request with prepared content
	req, err := http.NewRequest("POST", "/updates", bytes.NewReader(content))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("HashSHA256", hash)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//check what unpacked content is same as source content
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		assert.Equal(t, srcContent, body)
		w.WriteHeader(http.StatusOK)
	})

	resp := httptest.NewRecorder()

	handler(next).ServeHTTP(resp, req)
}

type syncer struct {
	*bytes.Buffer
}

func (s *syncer) Sync() error {
	return nil
}

func Test_WithLogger(t *testing.T) {

	handler := WithLogger()

	h := testHandler()

	buffer := &bytes.Buffer{}
	out := &syncer{buffer}

	logger.Init(out, false)

	r := regexp.MustCompile(`{"level":"info","ts":"[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}.[0-9]{3}.*","msg":"handle request","uuid":"[0-9a-z]{8}-[0-9a-z]{4}-[0-9a-z]{4}-[0-9a-z]{4}-[0-9a-z]{12}","uri":"\/update","method":"POST","code":200,"delay":[0-9]{1,10}.[0-9]{1,10},"size":7}`)

	resp := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodPost, "/update", nil)
	require.NoError(t, err)

	handler(h).ServeHTTP(resp, req)

	content := buffer.String()

	assert.Regexp(t, r, content)
}
