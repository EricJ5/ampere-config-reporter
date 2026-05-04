// Copyright (c) 2026, Ampere Computing LLC.
// 
// SPDX-License-Identifier: BSD-3-Clause

package collectors

import (
	"bufio"
	"log/slog"
	"os"
	"strings"
)

type MemoryCollector struct{}

func (c *MemoryCollector) Name() string {
	return "mem_info"
}

func (c *MemoryCollector) Collect() (map[string]interface{}, error) {
	desiredFields := map[string]bool{
		"MemAvailable":    true,
		"MemFree":         true,
		"MemTotal":        true,
		"Buffers":         true,
		"Cached":          true,
		"HugePages_Total": true,
		"Hugepagesize":    true,
		"SwapCached":      true,
		"SwapFree":        true,
		"SwapTotal":       true,
	}
	data := make(map[string]interface{})

	file, err := os.Open("/proc/meminfo")
	if err != nil {
		data["mem_info"] = "Not collected"
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			if _, ok := desiredFields[key]; !ok {
				continue
			}
			data[key] = value
		}

	}
	slog.Info("Memory info collection", "status", "complete")
	return data, scanner.Err()
}
