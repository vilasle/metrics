package main

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"database/sql"
	"encoding/pem"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vilasle/metrics/internal/logger"
	"github.com/vilasle/metrics/internal/repository"
	"github.com/vilasle/metrics/internal/service"
	srvSvc "github.com/vilasle/metrics/internal/service/server"
	"github.com/vilasle/metrics/internal/version"

	"github.com/vilasle/metrics/internal/repository/memory"
	"github.com/vilasle/metrics/internal/repository/memory/dumper"
	"github.com/vilasle/metrics/internal/repository/postgresql"

	_ "github.com/jackc/pgx/v5/stdlib"
	mdw "github.com/vilasle/metrics/internal/transport/rest/middleware"
	rest "github.com/vilasle/metrics/internal/transport/rest/server"
)

var buildVersion, buildDate, buildCommit string

func main() {
	version.ShowVersion(buildVersion, buildDate, buildCommit)

	logger.Init(os.Stdout, false)

	defer func() {
		if err := recover(); err != nil {
			logger.Error("application is in a panic", "err", err)
		}
	}()

	defer logger.Close()

	conf := getConfig()

	server, cancelDumper := createAndPreparingServer(conf)

	stop := subscribeToStopSignals()
	defer close(stop)

	logger.Infow("run server", "config", conf)

	go func() {
		if err := server.Start(); err != nil {
			logger.Errorw("server starting got error", "error", err)
		}
		stop <- os.Interrupt
	}()

	<-stop

	logger.Debug("got signal")

	if !server.IsRunning() {
		logger.Fatal("server stopped unexpected")
	}

	connectionClosed := make(chan struct{})

	shutdown(server, connectionClosed)

	<-connectionClosed

	cancelDumper()

	time.Sleep(time.Second * 2)
}

func shutdown(srv *rest.HTTPServer, closed chan struct{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)

	defer close(closed)
	defer cancel()

	err := srv.Stop(ctx)

	if err == nil {
		return nil
	}

	logger.Error("server stopping gracefully failed", "error", err)
	return srv.ForceStop()
}

func subscribeToStopSignals() chan os.Signal {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGQUIT, syscall.SIGTERM)
	return stop
}

func createRepositoryService(config runConfig) (service.MetricService, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())

	storage, err := getStorage(ctx, config)
	if err != nil {
		logger.Errorw("can not create storage", "error", err)
		os.Exit(1)
	}

	return srvSvc.NewMetricService(storage), cancel
}

func getStorage(ctx context.Context, config runConfig) (repository.MetricRepository, error) {
	if config.databaseDSN == "" {
		return memoryStorage(ctx, config)
	}
	return postgresStorage(config)
}

func memoryStorage(ctx context.Context, config runConfig) (repository.MetricRepository, error) {
	if fs, err := dumper.NewFileStream(config.dumpFilePath); err == nil {
		return dumper.NewFileDumper(ctx, dumper.Config{
			Timeout:      (time.Second * time.Duration(config.dumpInterval)),
			Restore:      config.restore,
			Storage:      memory.NewMetricRepository(),
			SerialWriter: fs,
		})
	} else {
		return nil, err
	}
}

func postgresStorage(config runConfig) (repository.MetricRepository, error) {
	db, err := sql.Open("pgx/v5", config.databaseDSN)
	if err != nil {
		return nil, err
	}
	return postgresql.NewRepository(db)
}

func createAndPreparingServer(config runConfig) (*rest.HTTPServer, context.CancelFunc) {
	hashKey, err := getHashKeyFromFile(config.hashSumKey)
	if err != nil {
		logger.Error("can not get hash key from file", "error", err)
	}

	key, err := getPrivateKeyFromFile(config.privateKeyPath)
	if err != nil {
		logger.Error("can not get private key from file", "error", err)
	}

	contentUnpackers := mdw.NewUnpackerChain(
		mdw.CheckHashSum(hashKey),
		mdw.DecryptContent(key, "update", "updates"),
		mdw.DecompressContent("gzip"),
	)

	middlewares := make([]func(http.Handler) http.Handler, 0, 4)
	middlewares = append(middlewares,
		mdw.WithLogger(),
		mdw.Compress("application/json", "text/html"),
		mdw.WithUnpackBody(contentUnpackers),
	)

	server := rest.NewHTTPServer(config.address, middlewares...)

	svc, cancel := createRepositoryService(config)

	registerHandlers(server, svc)
	return server, cancel
}

func registerHandlers(srv *rest.HTTPServer, svc service.MetricService) {
	srv.Register("/", rest.DisplayAllMetrics(svc), http.MethodGet)
	srv.Register("/ping", rest.Ping(svc), http.MethodGet)
	srv.Register("/value/", rest.DisplayMetric(svc), http.MethodPost)
	srv.Register("/update/", rest.UpdateMetric(svc), http.MethodPost)
	srv.Register("/updates/", rest.BatchUpdate(svc), http.MethodPost)
	srv.Register("/value/{type}/{name}", rest.DisplayMetric(svc), http.MethodGet)
	srv.Register("/update/{type}/{name}/{value}", rest.UpdateMetric(svc), http.MethodPost)
}

func getHashKeyFromFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func getPrivateKeyFromFile(path string) (*rsa.PrivateKey, error) {
	if path == "" {
		return nil, nil
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	privateBlock, _ := pem.Decode(content)
	return x509.ParsePKCS1PrivateKey(privateBlock.Bytes)
}
