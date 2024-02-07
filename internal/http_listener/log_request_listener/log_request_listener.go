package logrequestlistener

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/RichardoC/kube-audit-rest/internal/common"
	eventprocessor "github.com/RichardoC/kube-audit-rest/internal/event_processor"
	httplistener "github.com/RichardoC/kube-audit-rest/internal/http_listener"
)

type logRequestListener struct {
	server          *http.Server
	certFilename    string
	certKeyFilename string
}

func New(port int, certFilename string, certKeyFilename string, eProc eventprocessor.EventProcessor) httplistener.HttpListener {
	router := http.NewServeMux()
	router.HandleFunc("POST /log-request", eProc.ProcessEvent)

	addr := fmt.Sprintf(":%d", port)
	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	return &logRequestListener{
		server:          server,
		certFilename:    certFilename,
		certKeyFilename: certKeyFilename,
	}
}

func (lrl *logRequestListener) Start() {
	common.Logger.Infow("Starting server", "addr", lrl.server.Addr)
	if err := lrl.server.ListenAndServeTLS(lrl.certFilename, lrl.certKeyFilename); err != nil && err != http.ErrServerClosed {
		common.Logger.Fatalw("Failed to start server", "error", err, "addr", lrl.server.Addr)
	}
}

func (lrl *logRequestListener) Stop() {
	common.Logger.Warnw("Log Request Listener is shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	lrl.server.SetKeepAlivesEnabled(false)
	if err := lrl.server.Shutdown(ctx); err != nil {
		common.Logger.Fatalf("Could not gracefully shutdown the server: %v\n", err)
	}
}
