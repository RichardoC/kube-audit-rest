package prometheusmetrics

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/RichardoC/kube-audit-rest/internal/common"
	"github.com/RichardoC/kube-audit-rest/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type prometheusMetricsServer struct {
	reg    prometheus.Registerer
	server *http.Server
}

func New(port int) metrics.MetricsServer {
	reg := prometheus.NewRegistry()

	router := http.NewServeMux()
	router.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))

	addr := fmt.Sprintf(":%d", port)
	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	metricsServer := &prometheusMetricsServer{reg: reg, server: server}
	return metricsServer
}

func (ms *prometheusMetricsServer) CreateAndRegisterCounter(name string, help string) metrics.Counter {
	counter := prometheus.NewCounter(prometheus.CounterOpts{Name: name, Help: help})
	ms.reg.MustRegister(counter)
	return counter
}

func (ms *prometheusMetricsServer) Start() {
	common.Logger.Infow("Starting server", "addr", ms.server.Addr)
	if err := ms.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		common.Logger.Fatalw("Failed to start the Prometheus metrics server", "error", err, "addr", ms.server.Addr)
	}
}

func (ms *prometheusMetricsServer) Stop() {
	defer common.Logger.Sync()
	common.Logger.Warnw("Prometheus Metrics Server is shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	ms.server.SetKeepAlivesEnabled(false)
	if err := ms.server.Shutdown(ctx); err != nil {
		common.Logger.Fatalf("Could not gracefully shutdown the server: %v\n", err)
	}
}
