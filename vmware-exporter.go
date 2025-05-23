package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/prezhdarov/prometheus-exporter/config"
	"github.com/prezhdarov/prometheus-exporter/exporter"
	vmware "github.com/prezhdarov/vmware-exporter/vmware/api"
	vmwareCollectors "github.com/prezhdarov/vmware-exporter/vmware/collectors"

	"github.com/prometheus/common/promslog"
	"github.com/prometheus/exporter-toolkit/web"
)

const (
	exporterName = "VMware vSphere Exporter"
	namespace    = "vmware"
)

var (
	listenAddress          = flag.String("http.address", ":9169", "Address and port to listen for http connections")
	maxRequests            = flag.Int("prom.maxRequests", 20, "Maximum number of parallel scrape requests. Use 0 to disable.")
	disableExporterTarget  = flag.Bool("disable.exporter.target", false, "Disable default target for /metrics path.")
	disableExporterMetrics = flag.Bool("disable.exporter.metrics", true, "Disable exporter metrics in /metrics path. Always enabled if /metrics target disabled")

	logLevel  = flag.String("log.level", "debug", "Log Level minimums. Available options are: debug,info,warn and error")
	logFormat = flag.String("log.format", "logfmt", "Log output format. Available options are: logfmt and json")
)

func usage() {
	const s = `
vmware-exporter collects metrics data from VMware vCenter. 
`
	config.Usage(s)
}

func webConfig(listenAddress *string) *web.FlagConfig {

	listenAddresses := []string{*listenAddress}
	systemSocket := false
	configFile := ""

	return &web.FlagConfig{WebListenAddresses: &listenAddresses, WebSystemdSocket: &systemSocket, WebConfigFile: &configFile}
}

func main() {

	flag.CommandLine.SetOutput(os.Stdout)
	flag.Usage = usage
	config.Parse()

	logger := promslog.New(config.SetLogger(logFormat, logLevel))

	logger.Debug("disable exporter target is", fmt.Sprintf("%t", *disableExporterTarget), nil)

	vmware.Load(logger)

	vmwareCollectors.Load(logger)

	http.Handle("/metrics", exporter.CreateHandler(!*disableExporterMetrics, *disableExporterTarget, *maxRequests, namespace, logger))
	http.HandleFunc("/probe", func(w http.ResponseWriter, r *http.Request) {
		exporter.CreateHandleFunc(w, r, namespace, "", logger)
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`
			<head><title>VMware vSphere Exporter</title></head>
			<body>
			<h1>VMware vSphere Exporter</h1>
			<p><a href="/metrics">Metrics</a></p>
			<p><a href="/probe">Probe</a></p>
			</body>
			</html>`))
	})

	logger.Info("msg", "listening on", "address", *listenAddress, nil)

	server := &http.Server{}

	if err := web.ListenAndServe(server, webConfig(listenAddress), logger); err != nil {
		logger.Error(fmt.Sprintf("error: %s", err))
		os.Exit(1)
	}

}
