package version

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_ShowVersion(t *testing.T) {

	testCases := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "ShowVersion, args are filled",
			args: []string{"1.0.0", "2019-01-01", "abcdefg"},
			want: "Build version: 1.0.0\nBuild date: 2019-01-01\nBuild commit: abcdefg\n\n",
		},
		{
			name: "ShowVersion, args are not filled",
			args: []string{"", "", ""},
			want: "Build version: N/A\nBuild date: N/A\nBuild commit: N/A\n\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			old := os.Stdout

			r, w, _ := os.Pipe()

			os.Stdout = w
			ShowVersion(tc.args[0], tc.args[1], tc.args[2])
			os.Stdout = old

			w.Close()

			content := make([]byte, 1024)

			n, err := r.Read(content)
			require.NoError(t, err)

			content = content[:n]

			assert.Equal(t, tc.want, string(content))
		})
	}
}
