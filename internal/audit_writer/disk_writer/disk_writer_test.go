package diskwriter_test

import (
	"encoding/json"
	"log"
	"os"
	"path"
	"testing"

	diskwriter "github.com/RichardoC/kube-audit-rest/internal/audit_writer/disk_writer"
	"github.com/stretchr/testify/assert"
)

func Test_WhenWritingEvent_ThenEventWritten(t *testing.T) {
	// Create a tmp dir where we'll put the log file for this test
	// This directory is automatically destroyed after the test has finished
	tmpDir := t.TempDir()
	fileLog := path.Join(tmpDir, "test_file.log")
	dw := diskwriter.New(fileLog, 1, 1)

	event := "{\"testEvent\": \"test\"}"
	dw.LogEvent([]byte(event))

	// Check we can read the event we've just written
	byteContent, err := os.ReadFile(fileLog)
	if err != nil {
		log.Fatal(err)
	}
	var content map[string]string
	json.Unmarshal(byteContent, &content)
	assert.Equal(t, content["testEvent"], "test")
}
