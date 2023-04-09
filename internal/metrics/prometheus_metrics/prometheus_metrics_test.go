package prometheusmetrics_test

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	prometheusmetrics "github.com/RichardoC/kube-audit-rest/internal/metrics/prometheus_metrics"
	"github.com/stretchr/testify/assert"
)

func getFreePort() int {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()
	return port
}

func Test_WhenMetricsNotStarted_ThenStopSucceeds(t *testing.T) {
	ms := prometheusmetrics.New(1234)
	ms.Stop()
}

func Test_WhenCounterCreated_ThenItCanBeIncremented(t *testing.T) {
	ms := prometheusmetrics.New(1234)
	counter := ms.CreateAndRegisterCounter("test_counter", "This counter is for test purposes")
	counter.Inc()
}

func Test_WhenServerStarted_ThenServesRequests(t *testing.T) {
	port := getFreePort()
	ms := prometheusmetrics.New(port)
	go ms.Start()
	// Wait 100ms so the Prometheus server will be ready to receive the request that
	// we'll do in the next lines
	time.Sleep(100 * time.Millisecond)
	counter := ms.CreateAndRegisterCounter("test_counter", "This counter is for test purposes")
	counter.Inc()

	// Make request and check we get an answer
	requestURL := fmt.Sprintf("http://localhost:%d/metrics", port)
	res, err := http.Get(requestURL)
	if err != nil {
		log.Fatalf("Error making http request: %s\n", err)
	}
	// Read http body
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("Failed reading the response. %v", err)
	}
	strBody := string(body)

	assert.Equal(t, res.StatusCode, 200)
	assert.True(t, strings.Contains(strBody, "test_counter 1"))

	ms.Stop()
}
