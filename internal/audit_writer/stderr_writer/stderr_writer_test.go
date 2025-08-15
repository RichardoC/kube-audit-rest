package stderrwriter_test

import (
	"testing"

	stderrwriter "github.com/RichardoC/kube-audit-rest/internal/audit_writer/stderr_writer"
)

func Test_WhenWritingEventWithJsonFormat_ThenSucceeds(t *testing.T) {
	writer := stderrwriter.New("json")
	event := "{\"testEvent\": \"test\"}"
	writer.LogEvent([]byte(event))
}

func Test_WhenWritingEventWithConsoleFormat_ThenSucceeds(t *testing.T) {
	writer := stderrwriter.New("console")
	event := "{\"testEvent\": \"test\"}"
	writer.LogEvent([]byte(event))
}
