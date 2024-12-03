package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	service "github.com/vilasle/metrics/internal/service/server"

	"github.com/vilasle/metrics/internal/repository/memory"
	rest "github.com/vilasle/metrics/internal/transport/rest/server"
)

type runConfig struct {
	address string
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

	conf := getConfig()

	gaugeStorage := memory.NewMetricGaugeMemoryRepository()
	counterStorage := memory.NewMetricCounterMemoryRepository()

	svc := service.NewStorageService(gaugeStorage, counterStorage)

	server := rest.NewHTTPServer(conf.address)

	server.Register("/", methods(), contentTypes(), rest.DisplayAllMetrics(svc))
	server.Register("/value/", methods(http.MethodGet), contentTypes(), rest.DisplayMetric(svc))
	server.Register("/update/", methods(http.MethodPost), contentTypes(), rest.UpdateMetric(svc))

	stop := make(chan os.Signal, 1)
	defer close(stop)

	signal.Notify(stop, os.Interrupt)

	fmt.Printf("run server on %s\n", conf)
	go func() {
		if err := server.Start(); err != nil {
			fmt.Printf("server starting got error, %v", err)
		}
		stop <- os.Interrupt
	}()

	<-stop

	if !server.IsRunning() {
		os.Exit(0)
	}

	stopErr := make(chan error)
	defer close(stopErr)

	tickForce := time.NewTicker(time.Second * 5)
	tickKill := time.NewTicker(time.Second * 10)

	go func() { stopErr <- server.Stop() }()

	for {
		select {
		case err := <-stopErr:
			if err != nil {
				fmt.Println("server stopped with error", err)
				server.ForceStop()
			} else {
				os.Exit(0)
			}
		case <-tickForce.C:
			go server.ForceStop()
		case <-tickKill.C:
			fmt.Println("server did not stop during expected time")
			os.Exit(1)
		}
	}

}
func contentTypes(contentType ...string) []string {
	return contentType
}

func methods(method ...string) []string {
	return method
}
