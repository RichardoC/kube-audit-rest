package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/thought-machine/go-flags"
	"github.com/tidwall/gjson"
	"go.uber.org/zap"
	"go.uber.org/zap/zapio"
	"gopkg.in/natefinch/lumberjack.v2"
)

var healthy int32
var lg *zap.Logger
var logger *zap.SugaredLogger

// minimum viable response
// https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/#response
const responseTemplate = `{
	"apiVersion": "admission.k8s.io/v1",
	"kind": "AdmissionReview",
	"response": {
		"uid": "%s",
		"allowed": true
	}
}`

func logRequest(requestBody []byte, auditLogger io.Writer) {
	// Compact the json for single line use regardless of request prettiness
	dst := &bytes.Buffer{}
	json.Compact(dst, requestBody)

	_, err := fmt.Fprintln(auditLogger, dst)
	if err != nil {
		logger.Error(err)
	}
}

func logRequestHandler(w http.ResponseWriter, r *http.Request, auditLogger io.Writer) {
	totalRequests.Inc()
	logger.Debugw("Got request", "request", r)
	var body []byte
	// Don't bother with any logic if there is no request
	if r.Body != nil {
		if data, err := io.ReadAll(r.Body); err == nil {
			logger.Debugw("Got this body", "body", data)
			body = data
		} else {
			logger.Debugw(err.Error(), "body", r.Body)
			w.Header().Set("error", "Failed to read body")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	} else {
		logger.Debugw("No body provided")
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("error", "No body provided")
		return
	}

	// verify the content type is accurate
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		logger.Debugw("expect application/json", "contentType", contentType)
		w.Header().Set("error", "expect contentType application/json")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !gjson.ValidBytes(body) {
		logger.Debugw("invalid json", "body", body)
		w.Header().Set("error", "invalid json")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	requestUid := gjson.GetBytes(body, "request.uid").Str
	if requestUid == "" {
		logger.Debugln("failed to find request uid")
		w.Header().Set("error", "uid not provided")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Sychronous so that slower writes *do* slow our responses
	logRequest(body, auditLogger)

	// Record we processed a valid request
	validRequestsProcessed.Inc()

	// Template the uid into our default approval and finish up
	fmt.Fprintf(w, responseTemplate, requestUid)

}

// For holding the metrics
var (
	validRequestsProcessed = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "kube_audit_rest_valid_requests_processed_total",
		Help: "Total number of valid requests processed",
	})
	totalRequests = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "kube_audit_rest_http_requests_total",
			Help: "Total number of requests to kube-audit-rest",
		},
	)
)

func runMetrics(port int) {
	prometheus.MustRegister(validRequestsProcessed)
	prometheus.MustRegister(totalRequests)
	http.Handle("/metrics", promhttp.Handler())
	addr := fmt.Sprintf(":%d", port)
	logger.Infow("Starting metrics server",
		"addr", addr)

	http.ListenAndServe(addr, nil)
}

type Options struct {
	LoggerFilename   string `long:"logger-filename" description:"Location to log audit log to" default:"/tmp/kube-audit-rest.log"`
	AuditToStdErr    bool   `long:"audit-to-std-log" description:"Not recommended - log to stderr/stdout rather than a file"`
	LoggerMaxSize    int    `long:"logger-max-size" description:"Maximum size for each log file in megabytes" default:"500"`
	LoggerMaxBackups int    `long:"logger-max-backups" description:"Maximum number of rolled log files to store" default:"3"`
	CertFilename     string `long:"cert-filename" description:"Location of certificate for TLS" default:"/etc/tls/tls.crt"`
	CertKeyFilename  string `long:"cert-key-filename" description:"Location of certificate key for TLS" default:"/etc/tls/tls.key"`
	ServerPort       int    `long:"server-port" description:"Port to run https server on" default:"9090"`
	MetricsPort      int    `long:"metrics-port" description:"Port to run http metrics server on" default:"55555"`
	Verbose          bool   `long:"verbosity" short:"v" description:"Uses zap Development default verbose mode rather than production"`
}

func main() {
	var opts Options
	parser := flags.NewParser(&opts, flags.Default)
	_, err := parser.Parse()
	if err != nil {
		log.Fatalf("can't parse flags: %v", err)
	}

	if opts.Verbose {
		lg, err = zap.NewDevelopment()
		if err != nil {
			log.Fatalf("can't initialize zap logger: %v", err)
		}
	} else {
		lg, err = zap.NewProduction()
		if err != nil {
			log.Fatalf("can't initialize zap logger: %v", err)
		}

	}
	defer lg.Sync()
	logger = lg.Sugar()
	defer logger.Sync()

	logger.Infow("Got config",
		"config", opts,
	)

	// Send standard logging to zap
	undo := zap.RedirectStdLog(lg)
	defer undo()

	var auditLogger io.Writer

	if opts.AuditToStdErr {
		writer := &zapio.Writer{Log: lg, Level: zap.InfoLevel}
		defer writer.Close()
		auditLogger = writer

	} else {
		lumberjackLogger := &lumberjack.Logger{
			Filename:   opts.LoggerFilename,
			MaxSize:    opts.LoggerMaxSize,
			MaxBackups: opts.LoggerMaxBackups,
		}
		auditLogger = lumberjackLogger
		defer lumberjackLogger.Close()
	}

	go runMetrics(opts.MetricsPort)

	addr := fmt.Sprintf(":%d", opts.ServerPort)
	logger.Infow("Starting server",
		"addr", addr)

	router := http.NewServeMux()
	router.HandleFunc("/log-request", func(w http.ResponseWriter, r *http.Request) { logRequestHandler(w, r, auditLogger) })
	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	done := make(chan bool)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	signal.Notify(quit, syscall.SIGTERM)

	go func() {
		<-quit
		logger.Warnln("Server is shutting down...")
		// Sync loggers to make them persist
		logger.Sync()
		lg.Sync()

		atomic.StoreInt32(&healthy, 0)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		server.SetKeepAlivesEnabled(false)
		if err := server.Shutdown(ctx); err != nil {
			logger.Fatalf("Could not gracefully shutdown the server: %v\n", err)
		}
		close(done)
	}()

	atomic.StoreInt32(&healthy, 1)
	if err := server.ListenAndServeTLS(opts.CertFilename, opts.CertKeyFilename); err != nil && err != http.ErrServerClosed {
		logger.Fatalw("Failed to start server", "error", err, "addr", addr)
	}

	<-done
	logger.Infow("Server stopped")

}
