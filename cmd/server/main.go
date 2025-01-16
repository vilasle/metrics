		
package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vilasle/metrics/internal/logger"
	"github.com/vilasle/metrics/internal/repository"
	"github.com/vilasle/metrics/internal/service"
	srvSvc "github.com/vilasle/metrics/internal/service/server"
	"github.com/vilasle/metrics/internal/service/server/dumper"

	"github.com/vilasle/metrics/internal/repository/memory"
	"github.com/vilasle/metrics/internal/repository/postgresql"

	mdw "github.com/vilasle/metrics/internal/transport/rest/middlieware"
	rest "github.com/vilasle/metrics/internal/transport/rest/server"
)

type runConfig struct {
	address      string
	dumpFilePath string
	dumpInterval int64
	restore      bool
	databaseDns  string
}

func (c runConfig) String() string {
	return fmt.Sprintf("address: %s; dumpFilePath: %s; dumpInterval: %d; restore: %t; databaseDns: %s",
		c.address, c.dumpFilePath, c.dumpInterval, c.restore, c.databaseDns)
}

func (c runConfig) DNS() (string, error) {
	link, err := url.Parse(c.databaseDns)
	if err != nil {
		return "", err
	}
	_ = link
	return c.databaseDns, nil
}

func getConfig() runConfig {
	address := flag.String("a", "localhost:8080", "address for server")
	storageInternal := flag.Int64("i", 300, "dumping timeout")
	dumpFile := flag.String("f", "a.metrics", "dump file")
	restore := flag.Bool("r", true, "need to restore metrics from dump")
	dbDSN := flag.String("d", "'postgres://user:password@host:port/database?option=value'", "database dns")

	flag.Parse()

	envAddress := os.Getenv("ADDRESS")
	if envAddress != "" {
		*address = envAddress
	}

	envInternal := os.Getenv("STORAGE_INTERNAL")
	if envInternal != "" {
		if v, err := strconv.ParseInt(envInternal, 10, 64); err != nil {
			*storageInternal = v
		}
	}

	envDumpFile := os.Getenv("FILE_STORAGE_PATH")
	if envDumpFile != "" {
		*dumpFile = envDumpFile
	}

	envRestore := os.Getenv("RESTORE")
	if envRestore != "" {
		if v, err := strconv.ParseBool(envRestore); err != nil {
			*restore = v
		}
	}

	envDSN := os.Getenv("DATABASE_DSN")
	if envDSN != "" {
		*dbDSN = envDSN
	}

	return runConfig{
		address:      *address,
		restore:      *restore,
		dumpFilePath: *dumpFile,
		dumpInterval: *storageInternal,
		databaseDns:  *dbDSN,
	}
}

func main() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("application is in a panic", err)
		}
	}()

	logger.Init(os.Stdout, true)

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

	cancelDumper()
	time.Sleep(time.Second * 3)

	if !server.IsRunning() {
		logger.Error("server stopped unexpected")
		os.Exit(1)
	}

	shutdown(server)
}

func shutdown(srv *rest.HTTPServer) {
	tickForce := time.NewTicker(time.Second * 5)
	tickKill := time.NewTicker(time.Second * 10)

	stopErr := make(chan error)
	defer close(stopErr)

	go func() { stopErr <- srv.Stop() }()

	for {
		select {
		case err := <-stopErr:
			if err != nil {
				fmt.Println("server stopped with error", err)
				srv.ForceStop()
			} else {
				os.Exit(0)
			}
		case <-tickForce.C:
			go srv.ForceStop()
		case <-tickKill.C:
			fmt.Println("server did not stop during expected time")
			os.Exit(1)
		}
	}
}

func subscribeToStopSignals() chan os.Signal {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	return stop
}

func createRepositoryService(config runConfig) (service.MetricService, context.CancelFunc) {
	storage, err := getStorage(config)
	if err != nil {
		logger.Fatalf("can not create storage", "error", err)
	}

	fs, err := dumper.NewFileStream(config.dumpFilePath)
	if err != nil {
		logger.Fatalf("can not create file dumper", "error", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	svc := srvSvc.NewMetricService(storage)
	//FIXME dumper must wrap storage not service
	if svcDumper, err := dumper.NewFileDumper(ctx, dumper.Config{
		Timeout: (time.Second * time.Duration(config.dumpInterval)),
		Restore: config.restore,
		Service: svc,
		Stream:  fs,
	}); err == nil {
		return svcDumper, cancel
	} else {
		panic(err)
	}
}

func getStorage(config runConfig) (repository.MetricRepository, error) {
	if config.databaseDns == "" {
		return memory.NewMetricRepository(), nil
	}
	pool, err := pgxpool.New(context.TODO(), config.databaseDns)
	if err != nil {
		return nil, err
	}
	return postgresql.NewRepository(pool), nil
}

func createAndPreparingServer(config runConfig) (*rest.HTTPServer, context.CancelFunc) {
	server := rest.NewHTTPServer(config.address,
		mdw.WithLogger(),
		mdw.Compress("application/json", "text/html"))

	svc, cancel := createRepositoryService(config)

	registerHandlers(server, svc)
	return server, cancel
}

func registerHandlers(srv *rest.HTTPServer, svc service.MetricService) {
	srv.Register("/", nil, nil, rest.DisplayAllMetrics(svc))
	srv.Register("/update/", toSlice(http.MethodPost), nil, rest.UpdateMetric(svc))
	srv.Register("/value/", toSlice(http.MethodPost), nil, rest.DisplayMetric(svc))
	srv.Register("/value/{type}/{name}", toSlice(http.MethodGet), nil, rest.DisplayMetric(svc))
	srv.Register("/update/{type}/{name}/{value}", toSlice(http.MethodPost), nil, rest.UpdateMetric(svc))
	srv.Register("/ping", toSlice(http.MethodGet), nil, rest.Ping(svc))
}

func toSlice(it ...string) []string {
	return it
}
