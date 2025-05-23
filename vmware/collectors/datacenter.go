package vmwareCollectors

import (
	"context"
	"flag"
	"fmt"
	"log/slog"

	"github.com/prezhdarov/prometheus-exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
)

const (
	datacenterSubsystem = "datacenter"
)

var datacenterCollectorFlag = flag.Bool(fmt.Sprintf("collector.%s", datacenterSubsystem), collector.DefaultEnabled, fmt.Sprintf("Enable the %s collector (default: %v)", datacenterSubsystem, collector.DefaultEnabled))

type datacenterCollector struct {
	logger *slog.Logger
}

func init() {
	collector.RegisterCollector("datacenter", datacenterCollectorFlag, NewdatacenterCollector)
}

func NewdatacenterCollector(logger *slog.Logger) (collector.Collector, error) {
	return &datacenterCollector{logger}, nil
}

func (c *datacenterCollector) Update(ch chan<- prometheus.Metric, namespace string, clientAPI collector.ClientAPI, loginData map[string]interface{}, params map[string]string) error {

	var datacenters []mo.Datacenter

	err := fetchProperties(
		loginData["ctx"].(context.Context), loginData["view"].(*view.Manager), loginData["client"].(*vim25.Client),
		[]string{"Datacenter"}, []string{"name", "parent"}, &datacenters, c.logger,
	)
	if err != nil {
		return err

	}

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "vcenter", "info"),
			"This is basic vcenter info", nil,
			map[string]string{
				"version": loginData["client"].(*vim25.Client).ServiceContent.About.Version,
				"build":   loginData["client"].(*vim25.Client).ServiceContent.About.Build,
				"patch":   loginData["client"].(*vim25.Client).ServiceContent.About.PatchLevel,
				"vcenter": loginData["target"].(string)},
		), prometheus.GaugeValue, 1.0,
	)

	for _, datacenter := range datacenters {

		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				prometheus.BuildFQName(namespace, datacenterSubsystem, "info"),
				"This is basic datacenter info to be used for parent reference", nil,
				map[string]string{"dcmo": datacenter.Self.Value, "dc": datacenter.Name,
					"vcenter": loginData["target"].(string)},
			), prometheus.GaugeValue, 1.0,
		)

	}

	var folders []mo.Folder

	err = fetchProperties(
		loginData["ctx"].(context.Context), loginData["view"].(*view.Manager), loginData["client"].(*vim25.Client),
		[]string{"Folder"}, []string{"name", "parent"}, &folders, c.logger,
	)
	if err != nil {
		return err

	}

	for _, folder := range folders {

		if folder.Name == "host" || folder.Name == "datastore" {

			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "folder", "info"),
					"This is basic datacenter info to be used for parent reference", nil,
					map[string]string{"foldermo": folder.Self.Value, "dc": folder.Name, "dcmo": folder.Parent.Value,
						"vcenter": loginData["target"].(string)},
				), prometheus.GaugeValue, 1.0,
			)
		}
	}

	return nil
}
