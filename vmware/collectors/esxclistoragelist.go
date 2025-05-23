package vmwareCollectors

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/prezhdarov/vmware-exporter/vmware/esxcli"

	"github.com/prezhdarov/prometheus-exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
)

type StorageInfo struct {
	Vendor   string `xml:"Vendor"`
	Model    string `xml:"Model"`
	Revision string `xml:"Revision"`
}

type StorageResponse struct {
	DataObject []StorageInfo `xml:"DataObject"`
}

var (
	esxclistoragelistSubsystem = "esxcli_storage"
)

var esxclistoragelistCollectorFlag = flag.Bool("collector.esxcli.storage", collector.DefaultDisabled, fmt.Sprintf("Enable the %s collector (default: %v)", esxclistoragelistSubsystem, collector.DefaultDisabled))

type esxclistoragelistCollector struct {
	logger *slog.Logger
}

func init() {
	collector.RegisterCollector("esxcli.storage", esxclistoragelistCollectorFlag, NewesxcliStorageListCCollector)
}

func NewesxcliStorageListCCollector(logger *slog.Logger) (collector.Collector, error) {
	return &esxclistoragelistCollector{logger}, nil
}

func (c *esxclistoragelistCollector) Update(ch chan<- prometheus.Metric, namespace string, clientAPI collector.ClientAPI, loginData map[string]interface{}, params map[string]string) error {

	var (
		hosts []mo.HostSystem
	)

	err := fetchProperties(
		loginData["ctx"].(context.Context), loginData["view"].(*view.Manager), loginData["client"].(*vim25.Client),
		[]string{"HostSystem"}, []string{"runtime", "name"}, &hosts, c.logger,
	)
	if err != nil {
		return err

	}

	dCounter := 0

	wg := sync.WaitGroup{}

	for _, host := range hosts {

		if host.Runtime.PowerState == "poweredOn" && host.Runtime.ConnectionState == "connected" && !host.Runtime.InMaintenanceMode {

			dCounter++
			wg.Add(1)

			go func(host mo.HostSystem) {

				defer wg.Done()
				esxcliStorageDriverInfo(ch, c.logger, loginData["ctx"].(context.Context), loginData["client"].(*vim25.Client),
					host, &namespace, &esxclistoragelistSubsystem)
			}(host)

		}
	}

	c.logger.Debug("msg", fmt.Sprintf("dispatched %d StorageDriver routines", dCounter), nil)

	wg.Wait()

	return nil
}

func esxcliStorageDriverInfo(ch chan<- prometheus.Metric, logger *slog.Logger, ctx context.Context, client *vim25.Client,
	host mo.HostSystem, namespace, subsystem *string) {

	var (
		data        StorageResponse
		driverMutex = sync.Mutex{}
		driverMap   = make(map[string][]string)
	)

	mme, err := esxcli.GetHostMME(ctx, client, &host.Self)
	if err != nil {
		logger.Error("msg", "error retrieving host MME", fmt.Sprintf("error: %s", err), "host", host.Name)
		return
	}

	request := esxcli.ExecuteSoapRequest{
		This:    *mme,
		Moid:    "ha-cli-handler-storage-core-device",
		Method:  "vim.EsxCLI.storage.core.device.list",
		Version: "urn:vim25/5.0",
	}
	/*
		res, err := esxcli.ExecuteSoap(ctx, client, &request)
		if err != nil {
			//errchan <- err
			return
		}

		if res.Returnval != nil {
			if res.Returnval.Fault != nil {
				level.Error(logger).Log("msg", "error retrieving host nic info", "err", err)
				return
			}

		}

		err = xml.Unmarshal([]byte(res.Returnval.Response), &data)
		if err != nil {
			level.Error(logger).Log("msg", "error unmarshalling host nic info", "err", err)
			return
		}
	*/
	err = esxcli.GetSOAP(ctx, client, &request, &data)
	if err != nil {
		logger.Error("msg", "error fetching soap data", fmt.Sprintf("error: %s", err))
		return
	}

	// level.Debug(logger).Log("msg", fmt.Sprintf("we have SOAP from %s", request.This))

	for _, storage := range data.DataObject {

		// level.Debug(logger).Log("msg", fmt.Sprintf("procesing entry for %s with storage vendor %s, model %s (revision: %s", host.Name, strings.TrimSpace(storage.Vendor), storage.Model, storage.Revision))

		addEntry := false

		if _, exists := driverMap[storage.Model]; exists {

			if !inSlice(driverMap[storage.Model], &storage.Revision) {

				driverMutex.Lock()
				driverMap[storage.Model] = append(driverMap[storage.Model], storage.Revision)
				driverMutex.Unlock()

				addEntry = true

			}

		} else {

			driverMutex.Lock()
			driverMap[storage.Model] = []string{storage.Revision}
			driverMutex.Unlock()

			addEntry = true

		}

		if addEntry {

			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(*namespace, *subsystem, "driver"),
					"NIC Info", nil, map[string]string{"mo": host.Self.Value, "host": host.Name, "vendor": strings.TrimSpace(storage.Vendor), "model": strings.TrimSpace(storage.Model), "revision": strings.TrimSpace(storage.Revision)},
				), prometheus.GaugeValue, float64(1),
			)

		}
	}

}
