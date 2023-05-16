package stderrwriter_test

import (
	"testing"

	stderrwriter "github.com/RichardoC/kube-audit-rest/internal/audit_writer/stderr_writer"
)

func Test_WhenWritingEvent_ThenSucceeds(t *testing.T) {
	writer := stderrwriter.New()
	event := "{\"testEvent\": \"test\"}"
	writer.LogEvent([]byte(event))
}
