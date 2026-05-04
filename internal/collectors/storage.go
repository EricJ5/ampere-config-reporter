// Copyright (c) 2026, Ampere Computing LLC.
// 
// SPDX-License-Identifier: BSD-3-Clause

package collectors

import (
	"log/slog"

	"github.com/shirou/gopsutil/v3/disk"
)

type DiskCollector struct{}

func (c *DiskCollector) Name() string {
	return "disk_info"
}

func (c *DiskCollector) Collect() (map[string]interface{}, error) {
	data := make(map[string]interface{})
	partitions, err := disk.Partitions(false)
	if err != nil {
		data["partitions"] = "Not Detected"
		slog.Info("Unable to get disk info", "error", err)
		return data, err
	}
	usageData := make(map[string]interface{})
	for _, p := range partitions {
		usage, err := disk.Usage(p.Mountpoint)
		if err == nil {
			usageData[p.Mountpoint] = map[string]interface{}{
				"device":       p.Device,
				"fstype":       p.Fstype,
				"total_bytes":  usage.Total,
				"free_bytes":   usage.Free,
				"used_percent": usage.UsedPercent,
			}
		}
	}
	data["partitions"] = usageData
	return data, nil
}
