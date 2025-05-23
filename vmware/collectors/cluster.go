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
	clusterSubsystem = "cluster"
)

var clusterCollectorFlag = flag.Bool(fmt.Sprintf("collector.%s", clusterSubsystem), collector.DefaultEnabled, fmt.Sprintf("Enable the %s collector (default: %v)", clusterSubsystem, collector.DefaultEnabled))

type clusterCollector struct {
	logger *slog.Logger
}

func init() {
	collector.RegisterCollector("cluster", clusterCollectorFlag, NewClusterCollector)
}

func NewClusterCollector(logger *slog.Logger) (collector.Collector, error) {
	return &clusterCollector{logger}, nil
}

func (c *clusterCollector) Update(ch chan<- prometheus.Metric, namespace string, clientAPI collector.ClientAPI, loginData map[string]interface{}, params map[string]string) error {

	var clusters []mo.ClusterComputeResource

	err := fetchProperties(
		loginData["ctx"].(context.Context), loginData["view"].(*view.Manager), loginData["client"].(*vim25.Client),
		[]string{"ClusterComputeResource"}, []string{"name", "summary", "datastore", "parent"}, &clusters, c.logger,
	)
	if err != nil {
		return err

	}

	for _, cluster := range clusters {

		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				prometheus.BuildFQName(namespace, clusterSubsystem, "info"),
				"This is basic cluster info to be used for parent reference", nil,
				map[string]string{"cmo": cluster.ComputeResource.Self.Value, "vmwcluster": cluster.Name, "foldermo": cluster.Parent.Value,
					"vcenter": loginData["target"].(string)},
			), prometheus.GaugeValue, 1.0,
		)

		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				prometheus.BuildFQName(namespace, clusterSubsystem, "datastores"),
				"This is basic cluster info to be used for parent reference", nil,
				map[string]string{"cmo": cluster.ComputeResource.Self.Value, "vmwcluster": cluster.Name, "datastores": *moSliceToString(cluster.ComputeResource.Datastore),
					"vcenter": loginData["target"].(string)},
			), prometheus.GaugeValue, 1.0,
		)
	}

	if len(clusters) == 0 {

		var compute []mo.ComputeResource

		err = fetchProperties(
			loginData["ctx"].(context.Context), loginData["view"].(*view.Manager), loginData["client"].(*vim25.Client),
			[]string{"ComputeResource"}, []string{"name", "summary", "datastore", "parent"}, &compute, c.logger,
		)
		if err != nil {
			return err

		}

		for _, cr := range compute {

			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "compute", "info"),
					"This is basic cluster info to be used for parent reference", nil,
					map[string]string{"cmo": cr.Self.Value, "host": cr.Name, "foldermo": cr.Parent.Value,
						"vcenter": loginData["target"].(string)},
				), prometheus.GaugeValue, 1.0,
			)

			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "compute", "datastores"),
					"This is basic cluster info to be used for parent reference", nil,
					map[string]string{"cmo": cr.Self.Value, "host": cr.Name, "datastores": *moSliceToString(cr.Datastore),
						"vcenter": loginData["target"].(string)},
				), prometheus.GaugeValue, 1.0,
			)
		}
	}

	return nil
}
