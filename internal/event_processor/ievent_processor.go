// Package eventprocessor provides the interfaces to process and
// reply to http requests
package eventprocessor

//go:generate mockgen -package mymock -destination ../../mocks/event_processor_mock.go github.com/RichardoC/kube-audit-rest/internal/event_processor EventProcessor

import "net/http"

type EventProcessor interface {
	ProcessEvent(http.ResponseWriter, *http.Request)
}
