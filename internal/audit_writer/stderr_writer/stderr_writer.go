package stderrwriter

import (
	"log"

	auditwritter "github.com/RichardoC/kube-audit-rest/internal/audit_writer"
	commonwriter "github.com/RichardoC/kube-audit-rest/internal/audit_writer/common_writer"
	"go.uber.org/zap"
	"go.uber.org/zap/zapio"
)

type stderrWritter struct {
	writer *zapio.Writer
}

func New() auditwritter.AuditWritter {
	lg, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}
	writer := &zapio.Writer{Log: lg, Level: zap.InfoLevel}
	return &stderrWritter{writer: writer}
}

func (w *stderrWritter) LogEvent(body []byte) {
	commonwriter.LogEvent(body, w.writer)
}

func (w *stderrWritter) Sync() {
	w.writer.Sync()
}
