package main

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/thought-machine/go-flags"
	"github.com/tidwall/gjson"
	"gopkg.in/natefinch/lumberjack.v2"
)

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
		fmt.Println(err)
	}
}

func logRequestHandler(w http.ResponseWriter, r *http.Request, logger *lumberjack.Logger) {
	var body []byte
	// Don't bother with any logic if there is no request
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// verify the content type is accurate
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		fmt.Printf("contentType=%s, expect application/json", contentType)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !gjson.ValidBytes(body) {
		fmt.Printf("invalid json")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	requestUid := gjson.GetBytes(body, "request.uid").Str
	if requestUid == "" {
		fmt.Printf("failed to find request uid")
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
}

func main() {
	var opts Options
	parser := flags.NewParser(&opts, flags.Default)
	_, err := parser.Parse()
	if err != nil {
		panic(err)
	}
	fmt.Println(opts.LoggerFilename)
	// Log out log location, size etc

	auditLogger := &lumberjack.Logger{
		Filename:   opts.LoggerFilename,
		MaxSize:    opts.LoggerMaxSize, // megabytes
		MaxBackups: opts.LoggerMaxBackups,
	}

	http.HandleFunc("/log-request", func(w http.ResponseWriter, r *http.Request) { logRequestHandler(w, r, auditLogger) })

	// log out starting server
	err = http.ListenAndServeTLS(":9090", opts.CertFilename, opts.CertKeyFilename, nil)
	panic(err)
}
