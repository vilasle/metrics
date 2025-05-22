package http

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_hashSumWriter(t *testing.T) {
	testCases := []struct {
		name     string
		key      []byte
		data     []byte
		expected string
		wantErr  bool
	}{
		{
			name:     "valid key",
			key:      []byte("KeyForHashSum"),
			data:     []byte("some data"),
			expected: "6xjcKlVAfsIm0CmaPG54lCWamTjggHBgy1pnYwADH44=",
			wantErr:  false,
		},
		{
			name:     "empty key",
			key:      []byte{},
			data:     []byte("some data"),
			expected: "",
			wantErr:  false,
		},
	}

	for _, tt := range testCases {

		j := NewJSONWriter(WithCalculateHashSum(tt.key))

		ob := struct {
			name  string
			value int
		}{
			name:  "some name",
			value: 54321,
		}

		err := j.Write(ob)

		if tt.wantErr {
			require.Error(t, err)
			continue
		} else {
			require.NoError(t, err)
		}

		hash := j.headers["HashSHA256"]
		assert.Equal(t, tt.expected, hash)
	}
}

func Test_gzipWriter(t *testing.T) {
	testCases := []struct {
		name     string
		data     []byte
		expected []byte
		wantErr  bool
	}{
		{
			name: "compressing data",
			data: []byte("some data"),
			expected: []byte{
				31, 139, 8, 0, 0, 0, 0, 0, 2, 255, 170, 174, 5,
				4, 0, 0, 255, 255, 67, 191, 166, 163, 2, 0, 0, 0,
			},
			wantErr: false,
		},
	}

	for _, tt := range testCases {

		j := NewJSONWriter(WithCompressing())

		ob := struct {
			name  string
			value int
		}{
			name:  "some name",
			value: 54321,
		}

		err := j.Write(ob)

		if tt.wantErr {
			require.Error(t, err)
			continue
		} else {
			require.NoError(t, err)
		}

		actual := j.Bytes()
		assert.Equal(t, tt.expected, actual)
	}
}

func Test_encryptWriter(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	require.NoError(t, err)

	type object struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	testCases := []struct {
		name string
		object
		wantErr  bool
		expected []byte
	}{
		{
			name: "encryption data",
			object: object{
				Name:  "some name",
				Value: 54321,
			},
			wantErr:  false,
			expected: []byte(`{"name":"some name","value":54321}`),
		},
	}

	for _, tt := range testCases {
		j := NewJSONWriter(WithEncryption(&privateKey.PublicKey))

		err := j.Write(tt.object)

		if tt.wantErr {
			require.Error(t, err)
			continue
		} else {
			require.NoError(t, err)
		}

		data := j.Bytes()

		actual, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, data, []byte{})
		require.NoError(t, err)

		assert.Equal(t, tt.expected, actual)
	}
}
