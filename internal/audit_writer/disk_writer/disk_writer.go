package diskwriter

import (
	auditwritter "github.com/RichardoC/kube-audit-rest/internal/audit_writer"
	commonwriter "github.com/RichardoC/kube-audit-rest/internal/audit_writer/common_writer"
	"gopkg.in/natefinch/lumberjack.v2"
)

type diskWritter struct {
	lumberjackLogger *lumberjack.Logger
}

func New(filename string, loggerMaxSize int, loggerMaxBackups int) auditwritter.AuditWritter {
	lumberjackLogger := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    loggerMaxSize,
		MaxBackups: loggerMaxBackups,
	}
	return &diskWritter{lumberjackLogger: lumberjackLogger}
}

func (dw *diskWritter) LogEvent(body []byte) {
	commonwriter.LogEvent(body, dw.lumberjackLogger)
}

func (dw *diskWritter) Sync() {}
