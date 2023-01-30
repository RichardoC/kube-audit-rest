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
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
)

func main() {

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
				if addr == "kube-audit-rest:9090" {
					addr = "127.0.0.1:9090"
				}
				dialer := &net.Dialer{}
				return dialer.DialContext(ctx, network, addr)
			},
		},
	}

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
		resp, err := client.Post("https://kube-audit-rest:9090/log-request", "application/json", bytes.NewBuffer(line))
		if err != nil {
			testFailureCount += 1
			log.Println(err)
		} else {
			log.Println(resp)
		}

	}

	// Test unhappy path

	// Send a totally invalid request
	resp, err := client.Post("https://kube-audit-rest:9090/log-request", "application/json", bytes.NewBuffer([]byte("{a: \"b\"}")))
	if err != nil || resp.StatusCode != http.StatusBadRequest {
		testFailureCount += 1
		log.Println(err)
	} else {
		log.Println(resp)
	}

	// Send an almost valid request, but missing the uid
	resp, err = client.Post("https://kube-audit-rest:9090/log-request", "application/json", bytes.NewBuffer([]byte(`{"kind":"AdmissionReview","apiVersion":"admission.k8s.io/v1","request":{"kind":{"group":"authorization.k8s.io","version":"v1","kind":"SelfSubjectAccessReview"},"resource":{"group":"authorization.k8s.io","version":"v1","resource":"selfsubjectaccessreviews"},"requestKind":{"group":"authorization.k8s.io","version":"v1","kind":"SelfSubjectAccessReview"},"requestResource":{"group":"authorization.k8s.io","version":"v1","resource":"selfsubjectaccessreviews"},"operation":"CREATE","userInfo":{"username":"system:admin","groups":["system:masters","system:authenticated"]},"object":{"kind":"SelfSubjectAccessReview","apiVersion":"authorization.k8s.io/v1","metadata":{"creationTimestamp":null,"managedFields":[{"manager":"steveTEST","operation":"Update","apiVersion":"authorization.k8s.io/v1","time":"2022-11-30T17:46:51Z","fieldsType":"FieldsV1","fieldsV1":{"f:spec":{"f:resourceAttributes":{".":{},"f:group":{},"f:resource":{},"f:verb":{},"f:version":{}}}}}]},"spec":{"resourceAttributes":{"verb":"list","group":"batch","version":"v1","resource":"cronjobs"}},"status":{"allowed":false}},"oldObject":null,"dryRun":false,"options":{"kind":"CreateOptions","apiVersion":"meta.k8s.io/v1"}}}
	`)))
	if err != nil || resp.StatusCode != http.StatusBadRequest {
		testFailureCount += 1
		log.Println(err)
	} else {
		log.Println(resp)
	}

	// Send something that isn't json
	resp, err = client.Post("https://kube-audit-rest:9090/log-request", "application/json", bytes.NewBuffer([]byte("abc")))
	if err != nil || resp.StatusCode != http.StatusBadRequest {
		testFailureCount += 1
		log.Println(err)
	} else {
		log.Println(resp)
	}

	// Don't say we're sending json
	resp, err = client.Post("https://kube-audit-rest:9090/log-request", "text/plain", bytes.NewBuffer([]byte("abc")))
	if err != nil || resp.StatusCode != http.StatusBadRequest {
		testFailureCount += 1
		log.Println(err)
	} else {
		log.Println(resp)
	}

	// Ensure metrics server running
	resp, err = client.Get("http://localhost:55555/metrics")
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
