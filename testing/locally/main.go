/*
Copyright (C) 2023 Richard Tweed.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/thought-machine/go-flags"
)

type Options struct {
	ServerPort  int `long:"server-port" description:"Port where the http server listens to" default:"9090"`
	MetricsPort int `long:"metrics-port" description:"Port where the metrics server listens to" default:"55555"`
}

func main() {
	var opts Options
	parser := flags.NewParser(&opts, flags.Default)
	_, err := parser.Parse()
	if err != nil {
		log.Fatalf("can't parse flags: %v", err)
	}

	testFailureCount := 0

	// Set up for TLS
	caCert, err := ioutil.ReadFile("tmp/rootCA.pem")
	if err != nil {
		log.Fatal(err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// Set up custom http client so we can correctly validate TLS
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
			},
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				if addr == fmt.Sprintf("kube-audit-rest:%d", opts.ServerPort) {
					addr = fmt.Sprintf("127.0.0.1:%d", opts.ServerPort)
				}
				dialer := &net.Dialer{}
				return dialer.DialContext(ctx, network, addr)
			},
		},
	}

	logRequestAddr := fmt.Sprintf("https://kube-audit-rest:%d/log-request", opts.ServerPort)
	metricsAddr := fmt.Sprintf("http://localhost:%d/metrics", opts.MetricsPort)

	// Happy path testing
	apiLogFile, err := os.Open("testing/locally/data/kube-audit-rest.log")
	if err != nil {
		log.Fatal(err)
	}
	defer apiLogFile.Close()

	scanner := bufio.NewScanner(apiLogFile)
	for scanner.Scan() {
		line := scanner.Bytes()
		if err == io.EOF {
			log.Println(err)
			break
		}
		line = append(line, '\n')
		resp, err := client.Post(logRequestAddr, "application/json", bytes.NewBuffer(line))
		if err != nil {
			log.Println("Error while testing the happy path")
			testFailureCount += 1
			log.Println(err)
		} else {
			log.Println(resp)
		}

	}

	// Test unhappy path

	// Send a totally invalid request
	resp, err := client.Post(logRequestAddr, "application/json", bytes.NewBuffer([]byte("{a: \"b\"}")))
	if err != nil || resp.StatusCode != http.StatusBadRequest {
		log.Println("Error while executing \"Send a totally invalid request\"")
		testFailureCount += 1
		log.Println(err)
	} else {
		log.Println(resp)
	}

	// Send an almost valid request, but missing the uid
	resp, err = client.Post(logRequestAddr, "application/json", bytes.NewBuffer([]byte(`{"kind":"AdmissionReview","apiVersion":"admission.k8s.io/v1","request":{"kind":{"group":"authorization.k8s.io","version":"v1","kind":"SelfSubjectAccessReview"},"resource":{"group":"authorization.k8s.io","version":"v1","resource":"selfsubjectaccessreviews"},"requestKind":{"group":"authorization.k8s.io","version":"v1","kind":"SelfSubjectAccessReview"},"requestResource":{"group":"authorization.k8s.io","version":"v1","resource":"selfsubjectaccessreviews"},"operation":"CREATE","userInfo":{"username":"system:admin","groups":["system:masters","system:authenticated"]},"object":{"kind":"SelfSubjectAccessReview","apiVersion":"authorization.k8s.io/v1","metadata":{"creationTimestamp":null,"managedFields":[{"manager":"steveTEST","operation":"Update","apiVersion":"authorization.k8s.io/v1","time":"2022-11-30T17:46:51Z","fieldsType":"FieldsV1","fieldsV1":{"f:spec":{"f:resourceAttributes":{".":{},"f:group":{},"f:resource":{},"f:verb":{},"f:version":{}}}}}]},"spec":{"resourceAttributes":{"verb":"list","group":"batch","version":"v1","resource":"cronjobs"}},"status":{"allowed":false}},"oldObject":null,"dryRun":false,"options":{"kind":"CreateOptions","apiVersion":"meta.k8s.io/v1"}}}
	`)))
	if err != nil || resp.StatusCode != http.StatusBadRequest {
		log.Println("Error while executing \"Send an almost valid request, but missing the uid\"")
		testFailureCount += 1
		log.Println(err)
	} else {
		log.Println(resp)
	}

	// Send something that isn't json
	resp, err = client.Post(logRequestAddr, "application/json", bytes.NewBuffer([]byte("abc")))
	if err != nil || resp.StatusCode != http.StatusBadRequest {
		log.Println("Error while executing \"Send something that isn't json\"")
		testFailureCount += 1
		log.Println(err)
	} else {
		log.Println(resp)
	}

	// Don't say we're sending json
	resp, err = client.Post(logRequestAddr, "text/plain", bytes.NewBuffer([]byte("abc")))
	if err != nil || resp.StatusCode != http.StatusBadRequest {
		log.Println("Error while executing \"Don't say we're sending json\"")
		testFailureCount += 1
		log.Println(err)
	} else {
		log.Println(resp)
	}

	// Ensure metrics server running
	resp, err = client.Get(metricsAddr)
	if err != nil {
		testFailureCount += 1
	} else {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			testFailureCount += 1
			log.Println(err)
		} else if !strings.Contains(string(bodyBytes), "kube-audit-rest") {
			testFailureCount += 1
			log.Println("Failed to find any metrics")
		}
	}

	if testFailureCount > 0 {
		os.Exit(255)
	}
}
