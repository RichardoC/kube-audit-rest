/*
Copyright (C) 2023 Richard Tweed.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	auditwritter "github.com/RichardoC/kube-audit-rest/internal/audit_writer"
	diskwriter "github.com/RichardoC/kube-audit-rest/internal/audit_writer/disk_writer"
	stderrwriter "github.com/RichardoC/kube-audit-rest/internal/audit_writer/stderr_writer"
	"github.com/RichardoC/kube-audit-rest/internal/common"
	eventprocessorimpl "github.com/RichardoC/kube-audit-rest/internal/event_processor/event_processor_impl"
	logrequestlistener "github.com/RichardoC/kube-audit-rest/internal/http_listener/log_request_listener"
	prometheusmetrics "github.com/RichardoC/kube-audit-rest/internal/metrics/prometheus_metrics"
	"github.com/thought-machine/go-flags"

	"go.uber.org/automaxprocs/maxprocs"
)

type Options struct {
	LoggerFilename   string `long:"logger-filename" description:"Location to log audit log to" default:"/tmp/kube-audit-rest.log"`
	AuditToStdErr    bool   `long:"audit-to-std-log" description:"Not recommended - log to stderr/stdout rather than a file"`
	LoggerMaxSize    int    `long:"logger-max-size" description:"Maximum size for each log file in megabytes" default:"500"`
	LoggerMaxBackups int    `long:"logger-max-backups" description:"Maximum number of rolled log files to store, 0 means store all rolled files" default:"1"`
	CertFilename     string `long:"cert-filename" description:"Location of certificate for TLS" default:"/etc/tls/tls.crt"`
	CertKeyFilename  string `long:"cert-key-filename" description:"Location of certificate key for TLS" default:"/etc/tls/tls.key"`
	ServerPort       int    `long:"server-port" description:"Port to run https server on" default:"9090"`
	MetricsPort      int    `long:"metrics-port" description:"Port to run http metrics server on" default:"55555"`
	Verbose          bool   `long:"verbosity" short:"v" description:"Uses zap Development default verbose mode rather than production"`
}

func main() {
	// Set and parse command line options
	var opts Options
	parser := flags.NewParser(&opts, flags.Default)
	_, err := parser.Parse()
	if err != nil {
		log.Fatalf("can't parse flags: %v", err)
	}

	// Configure the logger
	if opts.Verbose {
		common.ConfigGlobalLogger(common.Dbg)
	} else {
		common.ConfigGlobalLogger(common.Prod)
	}
	defer common.Logger.Sync()


	common.Logger.Infow("Got config", "config", opts)

	// Set maxprocs and have it use our nice logger
	maxprocs.Set(maxprocs.Logger(common.Logger.Infof))

	// Create the components. In the future we can consider using containers
	metricsServer := prometheusmetrics.New(opts.MetricsPort)
	var auditWriter auditwritter.AuditWritter
	if opts.AuditToStdErr {
		auditWriter = stderrwriter.New()
	} else {
		auditWriter = diskwriter.New(opts.LoggerFilename, opts.LoggerMaxSize, opts.LoggerMaxBackups)
	}
	eventProcessor := eventprocessorimpl.New(auditWriter, metricsServer)
	httpListener := logrequestlistener.New(opts.ServerPort, opts.CertFilename, opts.CertKeyFilename, eventProcessor)

	go metricsServer.Start()
	go httpListener.Start()

	// Logic to capture SIGTERM and ctrl+c, so we can do a graceful shutdown
	done := make(chan bool)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	signal.Notify(quit, syscall.SIGTERM)

	go func() {
		<-quit
		httpListener.Stop()
		metricsServer.Stop()
		close(done)
	}()

	<-done
	common.Logger.Infow("Server stopped")
}
