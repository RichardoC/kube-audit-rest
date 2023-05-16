package logrequestlistener_test

import (
	"testing"

	eventprocessor "github.com/RichardoC/kube-audit-rest/internal/event_processor"
	logrequestlistener "github.com/RichardoC/kube-audit-rest/internal/http_listener/log_request_listener"
	mymock "github.com/RichardoC/kube-audit-rest/mocks"
	"github.com/golang/mock/gomock"
)

func setup(t *testing.T) eventprocessor.EventProcessor {
	ctrl := gomock.NewController(t)
	mockEvProc := mymock.NewMockEventProcessor(ctrl)
	return mockEvProc
}

func Test_WhenListenerNotStarted_ThenStopSucceeds(t *testing.T) {
	mockEvProc := setup(t)
	lrl := logrequestlistener.New(1234, "", "", mockEvProc)
	lrl.Stop()
}

// Testing the Start and Stop of the server is difficult because we need
// to generate temporary TLS self-signed certificates and the overall
// test would be much longer than the code it's trying to test
