package eventprocessorimpl

import (
	"html/template"
	"io"
	"net/http"

	auditwriter "github.com/RichardoC/kube-audit-rest/internal/audit_writer"
	"github.com/RichardoC/kube-audit-rest/internal/common"
	eventprocessor "github.com/RichardoC/kube-audit-rest/internal/event_processor"
	"github.com/RichardoC/kube-audit-rest/internal/metrics"
	"github.com/tidwall/gjson"
)

// minimum viable response
// https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/#response
const responseTemplate = `{
	"apiVersion": "admission.k8s.io/v1",
	"kind": "AdmissionReview",
	"response": {
		"uid": "{{.}}",
		"allowed": true
	}
}`

type eventProcImpl struct {
	validReqProc     metrics.Counter
	totalReq         metrics.Counter
	eventWritter     auditwriter.AuditWritter
	responseTemplate template.Template
}

func New(eventWritter auditwriter.AuditWritter, metricsServer metrics.MetricsServer) (eventprocessor.EventProcessor, error) {
	validReqProc := metricsServer.CreateAndRegisterCounter(
		"kube_audit_rest_valid_requests_processed_total",
		"Total number of valid requests processed",
	)
	totalReq := metricsServer.CreateAndRegisterCounter(
		"kube_audit_rest_http_requests_total",
		"Total number of requests to kube-audit-rest",
	)
	tmpl, err := template.New("name").Parse(responseTemplate)

	if err != nil {
		return &eventProcImpl{}, err
	}
	
	return &eventProcImpl{
		validReqProc:     validReqProc,
		totalReq:         totalReq,
		eventWritter:     eventWritter,
		responseTemplate: *tmpl,
	}, nil
}

func (ep *eventProcImpl) ProcessEvent(w http.ResponseWriter, r *http.Request) {
	ep.totalReq.Inc()
	common.Logger.Debugw("Got request", "request", r)
	var body []byte
	// Don't bother with any logic if there is no request
	if r.Body != nil {
		if data, err := io.ReadAll(r.Body); err == nil {
			common.Logger.Debugw("Got this body", "body", string(data))
			body = data
		} else {
			common.Logger.Debugw(err.Error(), "body", r.Body)
			w.Header().Set("error", "Failed to read body")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	} else {
		common.Logger.Debugw("No body provided")
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("error", "No body provided")
		return
	}

	// verify the content type is accurate
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		common.Logger.Debugw("expect application/json", "contentType", contentType)
		w.Header().Set("error", "expect contentType application/json")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !gjson.ValidBytes(body) {
		common.Logger.Debugw("invalid json", "body", body)
		w.Header().Set("error", "invalid json")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	requestUid := gjson.GetBytes(body, "request.uid").Str
	if requestUid == "" {
		common.Logger.Debugln("failed to find request uid")
		w.Header().Set("error", "uid not provided")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Sychronous so that slower writes *do* slow our responses
	ep.eventWritter.LogEvent(body)

	// Record we processed a valid request
	ep.validReqProc.Inc()

	// Template the uid into our default approval and finish up

	ep.responseTemplate.Execute(w, requestUid)

	// fmt.Fprintf(w, responseTemplate, requestUid)
}
