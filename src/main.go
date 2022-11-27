package main

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/tidwall/gjson"
)

const responseTemplate = `{
	"apiVersion": "admission.k8s.io/v1",
	"kind": "AdmissionReview",
	"response": {
		"uid": "%s",
		"allowed": true
	}
}`

func logRequest(requestBody []byte) {
	fmt.Printf("%s\n", requestBody)
}

func logRequestHandler(w http.ResponseWriter, r *http.Request) {
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
	go logRequest(body)

	// Template the uid into our default approval and finish up
	fmt.Fprintf(w, responseTemplate, requestUid)

}

func main() {

	http.HandleFunc("/log-request", logRequestHandler)

	// TODO have these be flags to a location
	certFile := "/etc/tls/tls.crt"
	keyFile := "/etc/tls/tls.key"
	http.ListenAndServeTLS(":9090", certFile, keyFile, nil)
}
