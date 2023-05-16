// Package httplistener provides the interfaces to create an http server
// that listens on a specific endpoint
package httplistener

//go:generate mockgen -package mymock -destination ../../mocks/http_listener_mock.go github.com/RichardoC/kube-audit-rest/internal/http_listener HttpListener

type HttpListener interface {
	Start()
	Stop()
}
