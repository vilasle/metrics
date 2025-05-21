package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_hashSumWriter(t *testing.T) {
	key := []byte("KeyForHashSum")
	j := NewJSONWriter(WithCalculateHashSum(key))

	ob := struct {
		name  string
		value int
	}{
		name:  "some name",
		value: 54321,
	}

	expected := "6xjcKlVAfsIm0CmaPG54lCWamTjggHBgy1pnYwADH44="
	err := j.Write(ob)
	require.NoError(t, err)

	
	hash, ok := j.headers["HashSHA256"]
	require.Equal(t, true, ok)

	assert.Equal(t, expected, hash)
}
