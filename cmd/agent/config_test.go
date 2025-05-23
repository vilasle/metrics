package main

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_getConfig(t *testing.T) {
	testCases := []struct {
		name              string
		cliArgs           []string
		envvars           map[string]string
		jsonConfigContent []byte
		want              runConfig
	}{
		{
			name: "there cli args and config file, and there conflicts",
			cliArgs: []string{
				"-r", "10",
				"-k", "key",
				"-l", "1",
				"-crypto-key", "crypto-key",
				"-config", "agent1234.json",
			},
			jsonConfigContent: []byte(
				`{
				   "address": "localhost:8090",
				   "report_interval": "1s",
				   "poll_interval": "1s",
				   "crypto_key": "/path/to/key.pem"
				}`,
			),
			envvars: map[string]string{
				"ADDRESS":         "localhost:9000",
				"REPORT_INTERVAL": "100",
				"CONFIG":          "agent.json",
				"CRYPTO_KEY":      "key.pem",
			},
			want: runConfig{
				endpoint:   "localhost:9000",
				report:     time.Second * 100,
				poll:       time.Second,
				rateLimit:  1,
				hashSumKey: "key",
				cryptoKey:  "key.pem",
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if tt.jsonConfigContent != nil {
				fd, err := os.OpenFile("agent.json", os.O_RDWR|os.O_CREATE, 0666)
				require.NoError(t, err)
				_, err = fd.Write(tt.jsonConfigContent)
				require.NoError(t, err)
				fd.Close()

			}

			for k, v := range tt.envvars {
				err := os.Setenv(k, v)
				require.NoError(t, err)
			}

			os.Args = append([]string{os.Args[0]}, tt.cliArgs...)

			got := getConfig()
			assert.Equal(t, tt.want, got)
			os.Remove("agent.json")
		})
	}

}
