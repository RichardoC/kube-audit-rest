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
	fmt.Fprintf(logger, "%s", requestBody)
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
	LoggerFilename string `long:"logger-filename" description:"Location to log audit log to" default:"/tmp/kube-rest-audit.log"`
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
		Filename:   "/tmp/kube-rest-audit.log",
		MaxSize:    500, // megabytes
		MaxBackups: 3,
		MaxAge:     28,    //days
		Compress:   false, // disabled by default
	}

	http.HandleFunc("/log-request", func(w http.ResponseWriter, r *http.Request) { logRequestHandler(w, r, auditLogger) })

	// TODO have these be flags to a location
	certFile := "/etc/tls/tls.crt"
	keyFile := "/etc/tls/tls.key"
	// log out starting server
	http.ListenAndServeTLS(":9090", certFile, keyFile, nil)
}
