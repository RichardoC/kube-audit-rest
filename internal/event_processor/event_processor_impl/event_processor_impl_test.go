package eventprocessorimpl_test

import (
	"net/http"
	"strings"
	"testing"

	eventprocessor "github.com/RichardoC/kube-audit-rest/internal/event_processor"
	eventprocessorimpl "github.com/RichardoC/kube-audit-rest/internal/event_processor/event_processor_impl"
	mymock "github.com/RichardoC/kube-audit-rest/mocks"
	"github.com/golang/mock/gomock"
)

type mockResponseWriter struct{}

func (mrw *mockResponseWriter) Header() http.Header {
	return http.Header{}
}

func (mrw *mockResponseWriter) Write([]byte) (int, error) {
	return 1, nil
}

func (mrw *mockResponseWriter) WriteHeader(statusCode int) {}

var correctBodyRequest string = `
{
	"request": {
		"uid": "test-uid"
	}
}
`

func setup(t *testing.T) (*mymock.MockAuditWritter, *mymock.MockMetricsServer) {
	ctrl := gomock.NewController(t)
	aw := mymock.NewMockAuditWritter(ctrl)
	ms := mymock.NewMockMetricsServer(ctrl)
	counter := mymock.NewMockCounter(ctrl)
	counter.EXPECT().Inc().AnyTimes()
	ms.EXPECT().CreateAndRegisterCounter(gomock.Any(), gomock.Any()).Return(counter).Times(2)
	return aw, ms
}

func sendRequest(ep eventprocessor.EventProcessor, header map[string][]string, body string) {
	reader := strings.NewReader(body)
	req, _ := http.NewRequest("POST", "localhost:80", reader)
	if header != nil {
		req.Header = header
	}
	respWriter := &mockResponseWriter{}
	ep.ProcessEvent(respWriter, req)
}

func Test_WhenRequestWellFormatted_ThenResponseSent(t *testing.T) {
	header := make(map[string][]string)
	header["Content-Type"] = []string{"application/json"}

	aw, ms := setup(t)
	aw.EXPECT().LogEvent([]byte(correctBodyRequest))
	ep := eventprocessorimpl.New(aw, ms)

	sendRequest(ep, header, correctBodyRequest)
}

func Test_WhenBadHeader_ThenNoEventLogged(t *testing.T) {
	header := make(map[string][]string)

	aw, ms := setup(t)
	ep := eventprocessorimpl.New(aw, ms)

	sendRequest(ep, header, correctBodyRequest)
}

func Test_WhenInvalidJsonBody_ThenNoEventLogged(t *testing.T) {
	header := make(map[string][]string)
	header["Content-Type"] = []string{"application/json"}

	aw, ms := setup(t)
	ep := eventprocessorimpl.New(aw, ms)

	sendRequest(ep, header, "")
}
