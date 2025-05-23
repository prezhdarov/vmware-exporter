package vmwareCollectors

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/vmware/govmomi/performance"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/types"
)

func Load(logger *slog.Logger) {
	logger.Info("msg", "Loading VMware vSphere collector set", nil)
}

func inSlice(slice []string, val *string) bool {
	for _, item := range slice {
		if item == *val {
			return true
		}
	}
	return false
}

func moSliceToString(moSlice []types.ManagedObjectReference) *string {

	var stringList string
	if len(moSlice) > 0 {

		stringList = moSlice[0].Value

		if len(moSlice) > 1 {

			for _, item := range moSlice[1:] {

				stringList = stringList + "," + item.Value
			}
		}
	}

	return &stringList
}

func fetchProperties(ctx context.Context, viewManager *view.Manager, vmwClient *vim25.Client, moTypes, propSpec []string, dataContainer interface{}, logger *slog.Logger) error {

	view, err := viewManager.CreateContainerView(
		ctx, vmwClient.ServiceContent.RootFolder,
		moTypes, true,
	)
	if err != nil {
		return err

	}

	defer view.Destroy(ctx)

	begin := time.Now()

	err = view.Retrieve(ctx, moTypes, propSpec, dataContainer)
	if err != nil {
		return err
	}

	logger.Debug("msg", fmt.Sprintf("Time to fetch PropColletor for %s: %f\n", moTypes, time.Since(begin).Seconds()), nil)

	return nil

}

func scrapePerformance(ctx context.Context, ch chan<- prometheus.Metric, logger *slog.Logger, sampleCount, sampleInterval int32,
	perfManager *performance.Manager, vcenter, moType, namespace, subsystem, instance string,
	counters []string, countersSpec map[string]*types.PerfCounterInfo,
	targetRefs []types.ManagedObjectReference, targetNames map[string]string) {

	logger.Debug("msg", fmt.Sprintf("gathering perfman metrics for hostRef %s\n", targetRefs[0]), nil)

	begin := time.Now()

	spec := types.PerfQuerySpec{
		MaxSample:  sampleCount,                                // Number of samples to fetch - if samples are fetched every 20s only one is needed.
		MetricId:   []types.PerfMetricId{{Instance: instance}}, //Instance takes either null string or * (or in fact any name of an performance manager metric instance)
		IntervalId: sampleInterval,                             // 20 seconds
	}

	sample, err := perfManager.SampleByName(ctx, spec, counters, targetRefs)
	if err != nil {
		logger.Error("msg", "error sampling the metrics and targtes", fmt.Sprintf("error: %s", err))
	}

	metrics, err := perfManager.ToMetricSeries(ctx, sample)
	if err != nil {
		logger.Error("msg", "error fetching metrics", fmt.Sprintf("error: %s", err))
	}

	logger.Debug("msg", fmt.Sprintf("Time to fetch Perfman for %s: %f\n", moType, time.Since(begin).Seconds()), nil)

	begin = time.Now()

	for _, metric := range metrics {

		labelMap := map[string]string{"vcenter": vcenter}

		switch {
		case moType == "HostSystem":
			labelMap["host"] = targetNames[metric.Entity.Value]
			labelMap["hostmo"] = metric.Entity.Value
		case moType == "VirtualMachine":
			labelMap["vm"] = targetNames[metric.Entity.Value]
			labelMap["vmmo"] = metric.Entity.Value
		case moType == "Datastore":
			labelMap["ds"] = targetNames[metric.Entity.Value]
			labelMap["dsmo"] = metric.Entity.Value
		}

		for _, value := range metric.Value {

			if value.Instance != "" {

				labelMap["pfinstance"] = value.Instance

			} else if instance != "" {
				continue //labels["instance"] = "-"
			}

			if len(value.Value) != 0 {

				if len(value.Value) == len(metric.SampleInfo) {

					avg := 0

					for _, subvalue := range value.Value {

						avg += int(subvalue)

					}

					avg = avg / len(value.Value)

					ch <- prometheus.MustNewConstMetric(
						prometheus.NewDesc(
							prometheus.BuildFQName(namespace, subsystem, strings.Replace(value.Name, ".", "_", -1)),
							fmt.Sprintf("%s in %s ", countersSpec[value.Name].UnitInfo.GetElementDescription().Label, countersSpec[value.Name].NameInfo.GetElementDescription().Summary),
							nil, labelMap,
						), prometheus.GaugeValue, float64(avg),
					)

				}
			}

		}
	}

	logger.Debug("msg", fmt.Sprintf("Time to process Perfman for %s: %f\n", moType, time.Since(begin).Seconds()), nil)

}
