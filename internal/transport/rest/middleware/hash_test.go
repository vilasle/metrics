package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_HashKey(t *testing.T) {
	key := []byte("test")

	handler := CalculateHashSum(string(key))

	resp := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodPost, "/", nil)
	require.NoError(t, err)

	h := testHandler()
	handler(h).ServeHTTP(resp, req)

	hashSum := resp.Header().Get("Hashsha256")
	require.NotEmpty(t, hashSum)
	assert.Equal(t, "PIEiCxCDirly_U9nljBNqxuzzN9loO3Ixc5-7zAZG2w=", hashSum)

}

func testHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("content"))
		w.WriteHeader(http.StatusOK)
	})
}
