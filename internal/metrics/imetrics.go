// Package metrics provides the interfaces to interact with a metrics server
package metrics

//go:generate mockgen -package mymock -destination ../../mocks/metrics_mock.go github.com/RichardoC/kube-audit-rest/internal/metrics Counter,MetricsServer

type Counter interface {
	Inc()
}

// A server that exposes an endpoint where it publishes metrics
type MetricsServer interface {
	// Start the server
	Start()
	// Stop the server gracefully
	Stop()
	// Creates the counter, registers it and returns it
	CreateAndRegisterCounter(name string, help string) Counter
}
