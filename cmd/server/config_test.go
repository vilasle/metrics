package main

import (
	"os"
	"testing"

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
				"-a", "localhost:17000",
				"-c", "server_custom.json",
				"-k", "hashSum.key",
				"-crypto-key", "private.key",
				"-r", "true",
			},
			jsonConfigContent: []byte(
				`{
					"address": "localhost:70",
					"restore": "false"
					"store_interval": 130,
					"store_file": "storage.file",
					"database_dsn": "postgres://user:password@localhost:5432/database"
					"crypto_key": "crypto.pem"
				}`,
			),
			envvars: map[string]string{
				"ADDRESS":           "localhost:10000",
				"DATABASE_DSN":      "postgres://user:password@localhost:5432/db",
				"STORAGE_INTERNAL":  "5000",
				"CONFIG":            "server.json",
				"KEY":               "hash.key",
				"FILE_STORAGE_PATH": "dump_file",
			},
			want: runConfig{
				address:        "localhost:10000",
				dumpFilePath:   "dump_file",
				dumpInterval:   5000,
				restore:        true,
				databaseDSN:    "postgres://user:password@localhost:5432/db",
				hashSumKey:     "hash.key",
				privateKeyPath: "private.key",
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if tt.jsonConfigContent != nil {
				fd, err := os.OpenFile("server.json", os.O_RDWR|os.O_CREATE, 0666)
				require.NoError(t, err)
				_, err = fd.Write(tt.jsonConfigContent)
				require.NoError(t, err)
				fd.Close()
				defer os.Remove("server.json")

			}

			for k, v := range tt.envvars {
				err := os.Setenv(k, v)
				require.NoError(t, err)
			}

			os.Args = append([]string{os.Args[0]}, tt.cliArgs...)

			got := getConfig()
			assert.Equal(t, tt.want, got)
		})
	}

}
