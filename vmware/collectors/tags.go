package vmwareCollectors

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"time"

	"github.com/prezhdarov/prometheus-exporter/pkg/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/vmware/govmomi/vapi/tags"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
)

const (
	tagsSubsystem = "tags"
	//vCenter caps vAPI list-attached-tags-on-objects requests, so ask in batches
	tagsBatchSize = 500
)

var tagsCollectorFlag = flag.Bool(fmt.Sprintf("collector.%s", tagsSubsystem), collector.DefaultEnabled, fmt.Sprintf("Enable the %s collector (default: %v)", tagsSubsystem, collector.DefaultEnabled))

type tagsCollector struct {
	logger *slog.Logger
}

func init() {
	collector.RegisterCollector(tagsSubsystem, tagsCollectorFlag, NewtagsCollector)
}

// NewtagsCollector returns a new Collector exposing vSphere tag assignments.
func NewtagsCollector(logger *slog.Logger) (collector.Collector, error) {
	return &tagsCollector{logger}, nil
}

func (c *tagsCollector) Update(ch chan<- prometheus.Metric, namespace string, clientAPI collector.ClientAPI, loginData map[string]interface{}, params map[string]string) error {

	tagsManager, ok := loginData["tags"].(*tags.Manager)
	if !ok {
		return fmt.Errorf("no tags manager in login data - has the vAPI REST login succeeded?")
	}

	ctx := loginData["ctx"].(context.Context)

	var vms []mo.VirtualMachine

	err := fetchProperties(
		ctx, loginData["view"].(*view.Manager), loginData["client"].(*vim25.Client),
		[]string{"VirtualMachine"}, []string{"name"}, &vms, c.logger,
	)
	if err != nil {
		return err
	}

	begin := time.Now()

	categories, err := tagsManager.GetCategories(ctx)
	if err != nil {
		return fmt.Errorf("tag categories err: %s", err)
	}

	categoryNames := make(map[string]string, len(categories))
	for _, category := range categories {
		categoryNames[category.ID] = category.Name
	}

	vmNames := make(map[string]string, len(vms))
	vmRefs := make([]mo.Reference, 0, len(vms))

	for _, vm := range vms {
		vmNames[vm.Self.Value] = vm.Name
		vmRefs = append(vmRefs, vm.Self)
	}

	for start := 0; start < len(vmRefs); start += tagsBatchSize {
		end := min(start+tagsBatchSize, len(vmRefs))

		attached, err := tagsManager.GetAttachedTagsOnObjects(ctx, vmRefs[start:end])
		if err != nil {
			return fmt.Errorf("attached tags err: %s", err)
		}

		for _, object := range attached {
			objectRef := object.ObjectID.Reference()

			for _, tag := range object.Tags {

				ch <- prometheus.MustNewConstMetric(
					prometheus.NewDesc(
						prometheus.BuildFQName(namespace, "vm", "tag"),
						"vSphere tag attached to this virtual machine. Value is always 1 - join on vmmo", nil,
						map[string]string{"vmmo": objectRef.Value, "vm": vmNames[objectRef.Value],
							"category": categoryNames[tag.CategoryID], "tag": tag.Name,
							"vcenter": loginData["target"].(string)},
					), prometheus.GaugeValue, 1.0,
				)
			}
		}
	}

	c.logger.Debug("time to fetch vm tags", "vms", len(vmRefs), "duration_seconds", time.Since(begin).Seconds())

	return nil
}
