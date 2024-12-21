package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	service "github.com/vilasle/metrics/internal/service/server"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/vilasle/metrics/internal/repository/memory"
	rest "github.com/vilasle/metrics/internal/transport/rest/server"
)

type runConfig struct {
	address string
}

func (c runConfig) String() string {
	return fmt.Sprintf("address: %s", c.address)
}

func getConfig() runConfig {
	address := flag.String("a", "localhost:8080", "address for server")

	flag.Parse()

	envAddress := os.Getenv("ADDRESS")
	if envAddress != "" {
		*address = envAddress
	}

	return runConfig{
		address: *address,
	}
}

func main() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("application is in a panic", err)
		}
	}()

	conf, logger := getConfig(), createLogger()
	defer logger.Sync()

	sugar := logger.Sugar()

	server := createAndPreparingServer(conf.address, sugar)

	stop := subscribeToStopSignals()
	defer close(stop)

	sugar.Infow("run server", "config", conf)

	go func() {
		if err := server.Start(); err != nil {
			sugar.Errorw("server starting got error", "error", err)
		}
		stop <- os.Interrupt
	}()

	<-stop

	if !server.IsRunning() {
		sugar.Error("server stopped unexpected")
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

func createRepositoryService() *service.StorageService {
	gaugeStorage, counterStorage :=
		memory.NewMetricGaugeMemoryRepository(),
		memory.NewMetricCounterMemoryRepository()

	svc := service.NewStorageService(gaugeStorage, counterStorage)
	return svc
}

func createLogger() *zap.Logger {
	encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	core := zapcore.NewCore(encoder, os.Stdout, zap.DebugLevel)

	logger := zap.New(core, zap.WithCaller(false), zap.AddStacktrace(zap.ErrorLevel))
	return logger
}

func createAndPreparingServer(addr string, logger *zap.SugaredLogger) *rest.HTTPServer {
	server := rest.NewHTTPServer(addr, rest.WithLogger(logger), rest.WithCompress("application/json", "text/html"))
	svc := createRepositoryService()

	registerHandlers(server, svc, logger)
	return server
}

func registerHandlers(srv *rest.HTTPServer, svc *service.StorageService, logger *zap.SugaredLogger) {
	srv.Register("/", nil, nil, rest.DisplayAllMetrics(svc, logger))
	srv.Register("/update/", toSlice(http.MethodPost), nil, rest.UpdateMetric(svc, logger))
	srv.Register("/value/", toSlice(http.MethodPost), nil, rest.DisplayMetric(svc, logger))
	srv.Register("/value/{type}/{name}", toSlice(http.MethodGet), nil, rest.DisplayMetric(svc, logger))
	srv.Register("/update/{type}/{name}/{value}", toSlice(http.MethodPost), nil, rest.UpdateMetric(svc, logger))
}

func toSlice(it ...string) []string {
	return it
}
