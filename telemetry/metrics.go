package telemetry

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/yolkhovyy/golang-grpc-demo/config"
)

type MetricsServer struct {
	server *http.Server
}

func NewMetricsServer() (*MetricsServer, error) {
	reg := prometheus.NewRegistry()

	err := reg.Register(collectors.NewGoCollector())
	if err != nil {
		return nil, err
	}

	err = reg.Register(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	if err != nil {
		return nil, err
	}

	metricsHandler := promhttp.InstrumentMetricHandler(reg,
		promhttp.HandlerFor(reg, promhttp.HandlerOpts{
			ErrorHandling: promhttp.ContinueOnError,
		}))

	serveMux := http.NewServeMux()
	serveMux.Handle("/metrics", metricsHandler)

	address := net.JoinHostPort(config.Service.Metrics.Host, strconv.Itoa(config.Service.Metrics.Port))
	server := &http.Server{
		Addr:    address,
		Handler: serveMux,
	}

	return &MetricsServer{
		server: server,
	}, nil
}

func (ms *MetricsServer) Serve() error {
	return ms.server.ListenAndServe()
}

func (ms *MetricsServer) Shutdown(shutdownTimeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	ms.server.Shutdown(ctx)
}
