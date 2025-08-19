package vmwareCollectors

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/prezhdarov/prometheus-exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/vmware/govmomi/performance"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

const (
	vmSubsystem = "vm"
)

var vmCollectorFlag = flag.Bool(fmt.Sprintf("collector.%s", vmSubsystem), collector.DefaultEnabled, fmt.Sprintf("Enable the %s collector (default: %v)", vmSubsystem, collector.DefaultEnabled))

var (
	cVMCounters = []string{"cpu.usagemhz.average", "cpu.demand.average", "cpu.latency.average", "cpu.entitlement.latest",
		"cpu.ready.summation", "cpu.readiness.average", "cpu.costop.summation", "cpu.maxlimited.summation",
		"mem.entitlement.average", "mem.active.average", "mem.shared.average", "mem.vmmemctl.average",
		"mem.swapped.average", "mem.consumed.average", "sys.uptime.latest",
	} //Common or generic counters that need not be instanced
	iVMCounters = []string{"net.bytesRx.average", "net.bytesTx.average",
		"datastore.read.average", "datastore.write.average", "datastore.numberReadAveraged.average",
		"datastore.numberWriteAveraged.average", "datastore.totalReadLatency.average", "datastore.totalWriteLatency.average"} //Counters that come in multiple i
)

type vmCollector struct {
	logger *slog.Logger
}

func init() {
	collector.RegisterCollector(vmSubsystem, vmCollectorFlag, NewvmCollector)
}

// NewMeminfoCollector returns a new Collector exposing memory stats.
func NewvmCollector(logger *slog.Logger) (collector.Collector, error) {
	return &vmCollector{logger}, nil
}

func (c *vmCollector) Update(ch chan<- prometheus.Metric, namespace string, clientAPI collector.ClientAPI, loginData map[string]interface{}, params map[string]string) error {

	var (
		vms     []mo.VirtualMachine
		vmRefs  []types.ManagedObjectReference
		vmNames = make(map[string]string)
	)

	begin := time.Now()

	err := fetchProperties(
		loginData["ctx"].(context.Context), loginData["view"].(*view.Manager), loginData["client"].(*vim25.Client),
		[]string{"VirtualMachine"}, []string{"summary", "runtime", "storage"}, &vms, c.logger,
	)
	if err != nil {
		return err

	}

	wg := sync.WaitGroup{}

	for _, vm := range vms {

		if vm.Runtime.PowerState == "poweredOn" {

			vmRefs = append(vmRefs, vm.Self)

			vmNames[vm.Self.Value] = vm.Summary.Config.Name

			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, vmSubsystem, "info"),
					"This is basic vm info to be used for parent reference", nil,
					map[string]string{"vmmo": vm.Self.Value, "vm": vm.Summary.Config.Name, "hostmo": vm.Runtime.Host.Value, "vcenter": loginData["target"].(string)},
				), prometheus.GaugeValue, 1.0,
			)

			//vmLabels := map[string]string{"vmmo": vm.Self.Value, "vm": vm.Summary.Config.Name, "vcenter": loginData["target"].(string)}

			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, vmSubsystem, "cpu_corecount"),
					"Number of virtual CPUs", nil, map[string]string{"vmmo": vm.Self.Value, "vm": vm.Summary.Config.Name, "hostmo": vm.Runtime.Host.Value, "vcenter": loginData["target"].(string)},
				), prometheus.GaugeValue, float64(vm.Summary.Config.NumCpu),
			)

			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, vmSubsystem, "mem_capacity"),
					"Virtual memory configured in MB", nil, map[string]string{"vmmo": vm.Self.Value, "vm": vm.Summary.Config.Name, "hostmo": vm.Runtime.Host.Value, "vcenter": loginData["target"].(string)},
				), prometheus.GaugeValue, float64(vm.Summary.Config.MemorySizeMB),
			)

			for _, datastore := range vm.Storage.PerDatastoreUsage {

				ch <- prometheus.MustNewConstMetric(
					prometheus.NewDesc(
						prometheus.BuildFQName(namespace, vmSubsystem, "datastore_capacity_used"),
						"Virtual memory configured in MB", nil,
						map[string]string{"vmmo": vm.Self.Value, "vm": vm.Summary.Config.Name,
							"vcenter": loginData["target"].(string), "dsmo": datastore.Datastore.Value},
					), prometheus.GaugeValue, float64(datastore.Committed),
				)
			}
		}

	}

	c.logger.Debug("msg", fmt.Sprintf("Time to process PropColletor for VM: %f\n", time.Since(begin).Seconds()), nil)

	begin = time.Now()

	if len(vmRefs) > 0 {

		wg.Add(2)
		for i := 0; i < 2; i++ {
			switch {
			case i == 0:
				go func(i int) {
					scrapePerformance(loginData["ctx"].(context.Context), ch, c.logger, loginData["samples"].(int32), loginData["interval"].(int32), loginData["perf"].(*performance.Manager),
						loginData["target"].(string), "VirtualMachine", namespace, vmSubsystem, "", cVMCounters,
						loginData["counters"].(map[string]*types.PerfCounterInfo), vmRefs, vmNames)
					wg.Done()
				}(i)

			case i == 1:
				go func(i int) {
					scrapePerformance(loginData["ctx"].(context.Context), ch, c.logger, loginData["samples"].(int32), loginData["interval"].(int32), loginData["perf"].(*performance.Manager),
						loginData["target"].(string), "VirtualMachine", namespace, vmSubsystem, "*", iVMCounters,
						loginData["counters"].(map[string]*types.PerfCounterInfo), vmRefs, vmNames)
					wg.Done()
				}(i)
			}

		}

		wg.Wait()

	}

	c.logger.Debug("msg", fmt.Sprintf("Time to process PerfMan for VM: %f\n", time.Since(begin).Seconds()), nil)

	return nil
}
