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

	"github.com/vilasle/metrics/internal/logger"
)

type UnpackFunc func([]byte, *http.Request) ([]byte, error)

type UnpackerChain []UnpackFunc

func NewUnpackerChain(unpackers ...UnpackFunc) UnpackerChain {
	return unpackers
}

func (c UnpackerChain) Unpack(b []byte, req *http.Request) (result []byte, err error) {
	result = b
	for _, unpack := range c {
		result, err = unpack(result, req)
		if err != nil {
			return nil, err
		}
	}
	return result, err
}

func CheckHashSum(hashKey []byte) UnpackFunc {
	return func(b []byte, req *http.Request) ([]byte, error) {
		if len(hashKey) == 0 {
			return b, nil
		}

		key := req.Context().Value(HashContextKey)
		hashSum := req.Header.Get("HashSHA256")

		sign, ok := key.(string)
		if !ok {
			return b, ErrInvalidKeyType
		}

		// nothing key for getting hash sum
		if sign == "" {
			return b, nil
		}

		logger.Debug("check key", "key", sign)

		reqHash, err := base64.URLEncoding.DecodeString(hashSum)
		if err != nil {
			return b, err
		}

		logger.Debug("source hash", "hash", reqHash)

		hashSumFromContext, err := getHashSumWithKey(b, sign)
		if err != nil {
			return b, err
		}

		logger.Debug("generated hash", "hash", hashSumFromContext)

		match := hmac.Equal(reqHash, hashSumFromContext)

		if match {
			return b, nil
		}

		return b, ErrInvalidHashSum
	}
}

func getHashSumWithKey(b []byte, key string) ([]byte, error) {
	h := hmac.New(sha256.New, []byte(key))

	if _, err := h.Write(b); err != nil {
		return []byte{}, err
	}

	return h.Sum(nil), nil
}

func DecryptContent(key *rsa.PrivateKey) UnpackFunc {
	return func(b []byte, _ *http.Request) ([]byte, error) {
		if key == nil {
			return b, nil
		}
		return rsa.DecryptOAEP(sha256.New(), rand.Reader, key, b, []byte{})
	}
}

func DecompressContent() UnpackFunc {
	return func(b []byte, req *http.Request) ([]byte, error) {
		rd := bytes.NewReader(b)
		grd, err := gzip.NewReader(rd)
		if err != nil {
			return nil, err
		}

		defer grd.Close()
		c, err := io.ReadAll(grd)
		if err == io.EOF {
			err = nil
		}
		return c, err
	}
}
