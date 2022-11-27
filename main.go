package main

import (
	"fmt"
	"io"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/thought-machine/go-flags"
	"github.com/tidwall/gjson"
	"gopkg.in/natefinch/lumberjack.v2"
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	// Makes it easier to be parsed by elastic search etc
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stderr)
	log.SetLevel(log.InfoLevel)
}

const responseTemplate = `{
	"apiVersion": "admission.k8s.io/v1",
	"kind": "AdmissionReview",
	"response": {
		"uid": "%s",
		"allowed": true
	}
}`

func logRequest(requestBody []byte, logger *lumberjack.Logger) {
	_, err := fmt.Fprintf(logger, "%s", requestBody)
	if err != nil {
		log.Error(err)
	}
}

func logRequestHandler(w http.ResponseWriter, r *http.Request, logger *lumberjack.Logger) {
	var body []byte
	// Don't bother with any logic if there is no request
	if r.Body != nil {
		if data, err := io.ReadAll(r.Body); err == nil {
			body = data
		} else {
			log.WithField("body", r.Body).Debug(err)
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
		log.WithField("contentType", contentType).Debugf("expect application/json")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !gjson.ValidBytes(body) {
		log.WithField("body", body).Debug("invalid json")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	requestUid := gjson.GetBytes(body, "request.uid").Str
	if requestUid == "" {
		log.Debug("failed to find request uid")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// So we don't hold up the request with any further processing
	go logRequest(body, logger)

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
		log.Fatal(err)
	}
	log.WithField("config", opts).Info("Got config")

	auditLogger := &lumberjack.Logger{
		Filename:   opts.LoggerFilename,
		MaxSize:    opts.LoggerMaxSize,
		MaxBackups: opts.LoggerMaxBackups,
	}

	addr := fmt.Sprintf(":%d", opts.ServerPort)
	log.WithField("addr", addr).Info("Starting server")
	http.HandleFunc("/log-request", func(w http.ResponseWriter, r *http.Request) { logRequestHandler(w, r, auditLogger) })

	err = http.ListenAndServeTLS(addr, opts.CertFilename, opts.CertKeyFilename, nil)
	log.Fatal(err)
}
