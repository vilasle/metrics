package http

import (
	"crypto/x509"
	"encoding/pem"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_WrappingBodyWriter(t *testing.T) {
	object := []struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}{
		{
			Name:  "SomeName",
			Value: 123456,
		},
		{
			Name:  "SomeName",
			Value: 123456,
		},
		{
			Name:  "SomeName",
			Value: 123456,
		},
		{
			Name:  "SomeName",
			Value: 123456,
		},
		{
			Name:  "SomeName",
			Value: 123456,
		},
		{
			Name:  "SomeName",
			Value: 123456,
		},
		{
			Name:  "SomeName",
			Value: 123456,
		},
	}

	content, err := os.ReadFile("metric.pub")
	require.NoError(t, err)

	publicBlock, _ := pem.Decode(content)
	key, err := x509.ParsePKCS1PublicKey(publicBlock.Bytes)
	require.NoError(t, err)

	hashKey := []byte("some hash key")
	json := NewJSONWriter(
		WithCalculateHashSum(hashKey),
		WithEncryption(key),
		WithCompressing())

	err = json.Write(object)
	require.NoError(t, err)

	actual := json.Bytes()

	_ = actual

	err = json.Write(object)
	_ = err
}
