package middleware

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http"

	"github.com/vilasle/metrics/internal/logger"
)

type key string

const HashContextKey key = "hashKey"

func HashKey(hashKey string) func(h http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			if len(hashKey) > 0 {
				ctx = context.WithValue(ctx, HashContextKey, hashKey)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)

	}

}

type hashCalculatedResponse struct {
	headerIsWrote bool
	http.ResponseWriter
	code int
	sign []byte
	key  []byte
}

func (hw *hashCalculatedResponse) WriteHeader(code int) {
	hw.headerIsWrote = true
	hw.code = code
}

func (hw *hashCalculatedResponse) Write(p []byte) (int, error) {
	if !hw.headerIsWrote {
		hw.WriteHeader(http.StatusOK)
	}

	hash, err := calculateHash(&p, hw.key)
	if err != nil {
		logger.Error("can not calculate a hash sum of body", "error", err)
	} else {
		hw.sign = hash
		hash := base64.URLEncoding.EncodeToString(hw.sign)
		hw.Header().Set("HashSHA256", hash)
	}
	hw.ResponseWriter.WriteHeader(hw.code)

	return hw.ResponseWriter.Write(p)
}

func calculateHash(pC *[]byte, key []byte) ([]byte, error) {
	h := hmac.New(sha256.New, key)
	_, err := h.Write(*pC)
	if err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}

func CalculateHashSum(hashKey string) func(h http.Handler) http.Handler {
	needCalculateHashSum := len(hashKey) > 0
	key := []byte(hashKey)
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			if needCalculateHashSum {
				next.ServeHTTP(&hashCalculatedResponse{ResponseWriter: w, key: key}, r)
				return
			}
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}
