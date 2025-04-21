package rest

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"

	"github.com/vilasle/metrics/internal/logger"
	"github.com/vilasle/metrics/internal/metric"
	"github.com/vilasle/metrics/internal/service"
	middleware "github.com/vilasle/metrics/internal/transport/rest/middleware"
)

type rawData struct {
	Name  string
	Type  string
	Value string
}

/*
auto-tests use filled Content-Type header only for iter1
that's why handle any Content-Type as text/plain with exception of application/json
*/
func updateMetric(svc service.MetricService, r *http.Request) Response {
	switch r.Header.Get("Content-Type") {
	case "application/json":
		return handleUpdateAsTextJSON(svc, r)
	default:
		return handleUpdateAsTextPlain(svc, r)
	}
}

func handleUpdateAsTextPlain(svc service.MetricService, r *http.Request) Response {
	raw := getRawDataFromContext(r.Context())
	logger.Debugw("raw data from url", "raw", raw, "url", r.URL.String())

	if m, err := metric.ParseMetric(raw.Name, raw.Value, raw.Type); err == nil {
		err := svc.Save(r.Context(), m)
		return newTextResponse(emptyBody(), err)
	} else {
		return newTextResponse(emptyBody(), err)
	}
}

func handleUpdateAsTextJSON(svc service.MetricService, r *http.Request) Response {
	defer r.Body.Close()
	if r.Body == http.NoBody {
		return newTextResponse(emptyBody(), ErrEmptyRequestBody)
	}
	content, err := io.ReadAll(r.Body)
	if err != nil {
		return newTextResponse(emptyBody(), ErrReadingRequestBody)
	}

	ok, err := checkHashSum(&content, r)
	if err != nil {
		return newTextResponse(emptyBody(), err)
	}

	if !ok {
		return newTextResponse(emptyBody(), ErrInvalidHashSum)
	}

	decompressedContent, err := unpackContent(content, r.Header.Get("Content-Encoding") == "gzip")
	if err != nil {
		return newTextResponse(emptyBody(), ErrReadingRequestBody)
	}

	logger.Debugw("request body", "url", r.URL.String(), "body", string(decompressedContent))

	m, err := metric.FromJSON(decompressedContent)
	if err != nil {
		return newTextResponse(emptyBody(), err)
	}

	if err = svc.Save(r.Context(), m); err != nil {
		return newTextResponse(emptyBody(), err)
	}

	logger.Debugw("updated metric", "metric", m)

	updContent, err := json.Marshal(m)
	return newJSONResponse(updContent, err)
}

func unpackContent(content []byte, isCompressed bool) ([]byte, error) {
	if !isCompressed {
		return content, nil
	}

	rd := bytes.NewReader(content)
	grd, err := gzip.NewReader(rd)
	if err != nil {
		return nil, err
	}

	defer grd.Close()
	if c, err := io.ReadAll(grd); err == io.EOF {
		return c, nil
	} else {
		return c, err
	}
}

func updateMetrics(svc service.MetricService, r *http.Request) Response {
	switch r.Header.Get("Content-Type") {
	case "application/json":
		return handleUpdateMetricsAsBatch(svc, r)
	default:
		return newTextResponse(emptyBody(), ErrUnknownContentType)
	}
}

func handleUpdateMetricsAsBatch(svc service.MetricService, r *http.Request) Response {
	defer r.Body.Close()
	if r.Body == http.NoBody {
		return newTextResponse(emptyBody(), ErrEmptyRequestBody)
	}
	content, err := io.ReadAll(r.Body)
	if err != nil {
		return newTextResponse(emptyBody(), ErrReadingRequestBody)
	}

	ok, err := checkHashSum(&content, r)
	if err != nil {
		return newTextResponse(emptyBody(), err)
	}

	if !ok {
		return newTextResponse(emptyBody(), ErrInvalidHashSum)
	}

	decompressedContent, err := unpackContent(content, r.Header.Get("Content-Encoding") == "gzip")
	if err != nil {
		return newTextResponse(emptyBody(), ErrReadingRequestBody)
	}

	logger.Debugw("request body", "url", r.URL.String(), "body", string(decompressedContent))

	ms, err := metric.FromJSONArray(decompressedContent)
	if err != nil {
		return newTextResponse(emptyBody(), err)
	}

	if err = svc.Save(r.Context(), ms...); err != nil {
		return newTextResponse(emptyBody(), err)
	}

	logger.Debugw("updated metrics", "metric", ms)

	return newTextResponse(emptyBody(), err)
}

func checkHashSum(pC *[]byte, req *http.Request) (bool, error) {
	key := req.Context().Value(middleware.HashContextKey)
	hashSum := req.Header.Get("HashSHA256")

	//nothing check
	if hashSum == "" {
		return true, nil
	}

	sign, ok := key.(string)
	if !ok {
		return false, ErrInvalidKeyType
	}

	//nothing key for getting hash sum
	if sign == "" {
		return true, nil
	}

	logger.Debug("check key", "key", sign)

	reqHash, err := base64.URLEncoding.DecodeString(hashSum)
	if err != nil {
		return false, err
	}

	logger.Debug("source hash", "hash", reqHash)

	hashSumFromContext, err := getHashSumWithKey(pC, sign)
	if err != nil {
		return false, err
	}

	logger.Debug("generated hash", "hash", hashSumFromContext)

	return hmac.Equal(reqHash, hashSumFromContext), nil
}

func getHashSumWithKey(pC *[]byte, key string) ([]byte, error) {
	h := hmac.New(sha256.New, []byte(key))

	if _, err := h.Write(*pC); err != nil {
		return []byte{}, err
	}

	return h.Sum(nil), nil
}
