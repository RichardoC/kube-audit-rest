// Package auditwritter implements the interfaces to write events
// to some medium (disk, stdout, etc.)
package auditwritter

//go:generate mockgen -package mymock -destination ../../mocks/audit_writer_mock.go github.com/RichardoC/kube-audit-rest/internal/audit_writer AuditWritter

type AuditWritter interface {
	LogEvent(body []byte)
	Sync()
}
