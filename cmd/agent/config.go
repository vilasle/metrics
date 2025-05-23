package main

import (
	"cmp"
	"encoding/json"
	"errors"
	"flag"
	"os"
	"strconv"
	"time"

	"github.com/vilasle/metrics/internal/logger"
)

type runConfig struct {
	report     time.Duration
	poll       time.Duration
	rateLimit  int
	endpoint   string
	hashSumKey string
	cryptoKey  string
}

type jsonConfig struct {
	Address        string   `json:"address"`
	ReportInterval Duration `json:"report_interval"`
	PollInterval   Duration `json:"poll_interval"`
	CryptoKey      string   `json:"crypto_key"`
}

// there are three sources of config:
// 1. command line arguments
// 2. environment variables
// 3. json config file
// the priority is environment variables arguments > json config file
func getConfig() runConfig {
	//if there are not this data in environment variables, cli args or json config file
	//then use default values
	config := runConfig{
		report:    time.Second * 10,
		poll:      time.Second * 2,
		rateLimit: 1,
		endpoint:  "localhost:8080",
	}

	endpoint := flag.String("a", "", "endpoint to send metrics")
	reportSec := flag.Int("r", 0, "timeout(sec) for sending report to server")
	pollSec := flag.Int("p", 0, "timeout(sec) for polling metrics")
	hashSumKey := flag.String("k", "", "path to key for hash sum")
	rateLimit := flag.Int("l", 0, "rate limit for sending metrics")
	cryptoKey := flag.String("crypto-key", "", "path to public key")
	
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

	config.endpoint = cmp.Or(
		os.Getenv("ADDRESS"),
		*endpoint,
		externalConfig.Address,
		config.endpoint,
	)

	config.report = cmp.Or(
		parseDuration(os.Getenv("REPORT_INTERVAL"), 0),
		time.Duration(*reportSec)*time.Second,
		externalConfig.ReportInterval.Duration,
		config.report,
	)

	config.poll = cmp.Or(
		parseDuration(os.Getenv("POLL_INTERVAL"), 0),
		time.Duration(*pollSec)*time.Second,
		externalConfig.PollInterval.Duration,
		config.poll,
	)
	config.hashSumKey = cmp.Or(
		os.Getenv("KEY"),
		*hashSumKey,
	)

	config.rateLimit = cmp.Or(
		parseInt(os.Getenv("RATE_LIMIT"), 0),
		*rateLimit,
		config.rateLimit,
	)

	config.cryptoKey = cmp.Or(
		os.Getenv("CRYPTO_KEY"),
		*cryptoKey,
		externalConfig.CryptoKey,
	)

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

func parseDuration(s string, defVal time.Duration) time.Duration {
	v, err := strconv.Atoi(s)
	if err != nil {
		return defVal
	}
	return time.Second * time.Duration(v)
}

type Duration struct {
	time.Duration
}

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case string:
		var err error
		d.Duration, err = time.ParseDuration(value)
		return err
	default:
		return errors.New("invalid duration")
	}
}
