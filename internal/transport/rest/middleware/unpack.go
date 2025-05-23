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
	"strings"

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
			return result, err
		}
	}
	return result, err
}

func CheckHashSum(key []byte) UnpackFunc {
	return func(b []byte, req *http.Request) ([]byte, error) {
		if len(key) == 0 {
			return b, nil
		}

		sum := req.Header.Get("HashSHA256")
		//nothing key for getting hash sum
		if sum == "" {
			return b, nil
		}

		reqHash, err := base64.URLEncoding.DecodeString(sum)
		if err != nil {
			return b, err
		}

		logger.Debug("request hash sum", "sum", reqHash)

		hashSum, err := getHashSumWithKey(b, key)
		if err != nil {
			return b, err
		}

		logger.Debug("generated hash", "hash", hashSum)

		match := hmac.Equal(reqHash, hashSum)

		if match {
			return b, nil
		}

		return b, ErrInvalidHashSum
	}
}

func getHashSumWithKey(b []byte, key []byte) ([]byte, error) {
	h := hmac.New(sha256.New, key)

	if _, err := h.Write(b); err != nil {
		return []byte{}, err
	}

	return h.Sum(nil), nil
}

func DecryptContent(key *rsa.PrivateKey, path ...string) UnpackFunc {
	encryptedPath := make(map[string]struct{}, len(path))
	for i := range path {
		rs := path[i]

		if !strings.HasPrefix(rs, "/") {
			rs = "/" + rs
		}

		if !strings.HasSuffix(rs, "/") {
			rs = rs + "/"
		}

		encryptedPath[rs] = struct{}{}
	}

	return func(b []byte, req *http.Request) ([]byte, error) {
		if key == nil {
			return b, nil
		}

		path := req.RequestURI
		
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
		if !strings.HasSuffix(path, "/") {
			path = path + "/"
		}

		if _, ok := encryptedPath[path]; !ok {
			return b, nil
		}

		return rsa.DecryptOAEP(sha256.New(), rand.Reader, key, b, []byte{})
	}
}

func DecompressContent(types ...string) UnpackFunc {
	supportedEncodings := make(map[string]struct{}, len(types))

	for i := range types {
		supportedEncodings[types[i]] = struct{}{}
	}

	return func(b []byte, req *http.Request) ([]byte, error) {
		encoding := req.Header.Get("Content-Encoding")

		if _, ok := supportedEncodings[encoding]; !ok {
			return b, nil
		}

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
