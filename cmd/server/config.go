package main

import (
	"cmp"
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strconv"

	"github.com/vilasle/metrics/internal/logger"
)

type jsonConfig struct {
	Address         string `json:"address"`
	Restore         bool   `json:"restore"`
	StorageInternal int    `json:"store_interval"`
	StorageFile     string `json:"store_file"`
	DatabaseDSN     string `json:"database_dsn"`
	CryptoKeyPath   string `json:"crypto_key"`
}

type runConfig struct {
	address        string
	dumpFilePath   string
	dumpInterval   int64
	restore        bool
	databaseDSN    string
	hashSumKey     string
	privateKeyPath string
}

func (c runConfig) String() string {
	return fmt.Sprintf("address: %s; dumpFilePath: %s; dumpInterval: %d; restore: %t; databaseDSN: %s; key for hash sum: %s",
		c.address, c.dumpFilePath, c.dumpInterval, c.restore, c.databaseDSN, c.hashSumKey)
}

func (c runConfig) DNS() (string, error) {
	link, err := url.Parse(c.databaseDSN)
	if err != nil {
		return "", err
	}
	_ = link
	return c.databaseDSN, nil
}

func getConfig() runConfig {
	config := runConfig{
		dumpInterval: 300,
		restore:      true,
		address:      "localhost:8080",
		dumpFilePath: "a.metrics",
	}

	address := flag.String("a", "", "address for server")
	storageInternal := flag.Int64("i", 0, "dumping timeout")
	dumpFile := flag.String("f", "", "dump file")
	restore := flag.Bool("r", false, "need to restore metrics from dump")
	dbDSN := flag.String("d", "", "database dns e.g. 'postgres://user:password@host:port/database?option=value'")
	hashSumKey := flag.String("k", "", "key for hash sum")
	cryptoKey := flag.String("crypto-key", "", "path to private key")

	var configPath string
	flag.StringVar(&configPath, "config", "", "path to json config file")
	flag.StringVar(&configPath, "c", "", "path to json config file")

	flag.Parse()

	externalConfig, err := readConfig(cmp.Or(os.Getenv("CONFIG"), configPath))
	if err != nil {
		logger.Warn("can not read config file, will use default values",
			"path", configPath,
			"error", err)
	}
	config.address = cmp.Or(
		os.Getenv("ADDRESS"),
		*address,
		externalConfig.Address,
		config.address)

	config.dumpFilePath = cmp.Or(
		os.Getenv("FILE_STORAGE_PATH"),
		*dumpFile, config.dumpFilePath,
		config.dumpFilePath)

	config.databaseDSN = cmp.Or(
		os.Getenv("DATABASE_DSN"),
		*dbDSN,
		externalConfig.DatabaseDSN)

	config.dumpInterval = cmp.Or(
		int64(parseInt(os.Getenv("STORAGE_INTERNAL"), 0)),
		*storageInternal,
		int64(externalConfig.StorageInternal),
		config.dumpInterval,
	)

	config.restore = cmp.Or(
		parseBool(os.Getenv("RESTORE"), false),
		*restore,
		externalConfig.Restore,
		config.restore)

	config.hashSumKey = cmp.Or(
		os.Getenv("KEY"),
		*hashSumKey)

	config.privateKeyPath = cmp.Or(
		os.Getenv("CRYPTO_KEY"),
		*cryptoKey,
		externalConfig.CryptoKeyPath)

	return config
}

func readConfig(path string) (jsonConfig, error) {
	var config jsonConfig
	if path == "" {
		return config, nil
	}

	file, err := os.Open(path)
	if err != nil {
		return config, err
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(&config)
	return config, err
}

func parseInt(s string, defVal int) int {
	v, err := strconv.Atoi(s)
	if err != nil {
		return defVal
	}
	return v
}

func parseBool(b string, defVal bool) bool {
	v, err := strconv.ParseBool(b)
	if err != nil {
		return defVal
	}
	return v
}
