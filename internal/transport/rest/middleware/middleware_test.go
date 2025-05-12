package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
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
