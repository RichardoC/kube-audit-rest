package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/thought-machine/go-flags"
	"github.com/tidwall/gjson"
	"gopkg.in/natefinch/lumberjack.v2"
)

var logger = logrus.New()
var healthy int32

func init() {

	// Log as JSON instead of the default ASCII formatter.
	// Makes it easier to be parsed by elastic search etc
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetOutput(os.Stderr)
	logger.SetLevel(logrus.InfoLevel)

	// Use logrus for standard log output
	// Note that `log` here references stdlib's log
	// Not logrus imported under the name `log`.
	log.SetOutput(logger.Writer())
}

const responseTemplate = `{
	"apiVersion": "admission.k8s.io/v1",
	"kind": "AdmissionReview",
	"response": {
		"uid": "%s",
		"allowed": true
	}
}`

func logRequest(requestBody []byte, auditLogger *lumberjack.Logger) {
	_, err := fmt.Fprintf(auditLogger, "%s", requestBody)
	if err != nil {
		logger.Error(err)
	}
}

func logRequestHandler(w http.ResponseWriter, r *http.Request, auditLogger *lumberjack.Logger) {
	var body []byte
	// Don't bother with any logic if there is no request
	if r.Body != nil {
		if data, err := io.ReadAll(r.Body); err == nil {
			body = data
		} else {
			logger.WithField("body", r.Body).Debug(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// verify the content type is accurate
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		logger.WithField("contentType", contentType).Debugf("expect application/json")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !gjson.ValidBytes(body) {
		logger.WithField("body", body).Debug("invalid json")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	requestUid := gjson.GetBytes(body, "request.uid").Str
	if requestUid == "" {
		logger.Debug("failed to find request uid")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// So we don't hold up the request with any further processing
	go logRequest(body, auditLogger)

	// Template the uid into our default approval and finish up
	fmt.Fprintf(w, responseTemplate, requestUid)

}

type Options struct {
	LoggerFilename   string `long:"logger-filename" description:"Location to log audit log to" default:"/tmp/kube-audit-rest.log"`
	LoggerMaxSize    int    `long:"logger-max-size" description:"Maximum size for each log file in megabytes" default:"500"`
	LoggerMaxBackups int    `long:"logger-max-backups" description:"Maximum number of rolled log files to store" default:"3"`
	CertFilename     string `long:"cert-filename" description:"Location of certificate for TLS" default:"/etc/tls/tls.crt"`
	CertKeyFilename  string `long:"cert-key-filename" description:"Location of certificate key for TLS" default:"/etc/tls/tls.key"`
	ServerPort       int    `long:"server-port" description:"Port to run https server on" default:"9090"`
}

func main() {
	var opts Options
	parser := flags.NewParser(&opts, flags.Default)
	_, err := parser.Parse()
	if err != nil {
		logger.Fatal(err)
	}
	logger.WithField("config", opts).Info("Got config")

	auditLogger := &lumberjack.Logger{
		Filename:   opts.LoggerFilename,
		MaxSize:    opts.LoggerMaxSize,
		MaxBackups: opts.LoggerMaxBackups,
	}

	// Required soo http errors structured
	w := logger.Writer()
	defer w.Close()

	addr := fmt.Sprintf(":%d", opts.ServerPort)
	logger.WithField("addr", addr).Info("Starting server")

	router := http.NewServeMux()
	router.HandleFunc("/log-request", func(w http.ResponseWriter, r *http.Request) { logRequestHandler(w, r, auditLogger) })
	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ErrorLog:     log.New(w, "", 0),
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
		logger.Println("Server is shutting down...")
		atomic.StoreInt32(&healthy, 0)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		server.SetKeepAlivesEnabled(false)
		if err := server.Shutdown(ctx); err != nil {
			logger.Fatalf("Could not gracefully shutdown the server: %v\n", err)
		}
		close(done)
	}()

	logger.Println("Server is ready to handle requests at", addr)
	atomic.StoreInt32(&healthy, 1)
	if err := server.ListenAndServeTLS(opts.CertFilename, opts.CertKeyFilename); err != nil && err != http.ErrServerClosed {
		logger.WithField("addr", addr).Fatal(err)
	}

	<-done
	logger.Println("Server stopped")

	// err = http.ListenAndServeTLS(addr, opts.CertFilename, opts.CertKeyFilename, nil)
	// logger.Fatal(err)

}
