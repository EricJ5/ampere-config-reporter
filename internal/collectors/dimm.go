// Copyright (c) 2026, Ampere Computing LLC.
// 
// SPDX-License-Identifier: BSD-3-Clause

package collectors

import (
	"log/slog"
	"os/exec"
	"regexp"
	"sort"
	"strings"
)

// DimmCollector gathers detailed hardware info about memory from dmidecode.
type DimmCollector struct{}

func (c *DimmCollector) Name() string {
	return "dimm_info"
}

func (c *DimmCollector) Collect() (map[string]interface{}, error) {
	cmd := exec.Command("dmidecode", "-t", "memory")
	out, err := cmd.Output()
	if err != nil {
		// This can happen if dmidecode is not installed or user doesn't have permissions
		return nil, err
	}

	return parseDmidecodeMemory(string(out)), nil
}

// parseDmidecodeMemory extracts memory channel and DIMM info from dmidecode output.
func parseDmidecodeMemory(output string) map[string]interface{} {
	data := make(map[string]interface{})
	populatedDIMMs := []map[string]string{}
	channels := make(map[string]bool)

	// Looks for "Channel" or "CH" followed by a letter or number.
	// e.g., "CPU0-Channel0-DIMM0", "P1-CH A", "ChannelA"
	explicitChannelRegex := regexp.MustCompile(`(?i)(?:channel|ch)\s*([0-9a-z]+)`)
	structuredLocatorRegex := regexp.MustCompile(`(?i)DIMM_P\d+_([A-Z])\d+`)
	simpleLocatorRegex := regexp.MustCompile(`(?i)DIMM([A-Z])\d+`)

	// Split output into blocks for each "Memory Device"
	deviceBlocks := strings.Split(output, "Memory Device")

	for _, block := range deviceBlocks[1:] { // Skip the first part before any device
		dimmInfo := make(map[string]string)
		var size string
		var locator string
		var banklocator string

		lines := strings.Split(block, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			parts := strings.SplitN(line, ":", 2)
			if len(parts) < 2 {
				continue
			}
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			switch key {
			case "Size":
				size = value
				dimmInfo[key] = value
			case "Locator":
				locator = value
				dimmInfo[key] = value
			case "Bank Locator":
				banklocator = value
				dimmInfo[key] = value
			case "Type":
				dimmInfo[key] = value
			case "Speed":
				dimmInfo[key] = value
			case "Configured Memory Speed":
				dimmInfo[key] = value
			case "Manufacturer":
				dimmInfo[key] = value
			case "Part Number":
				dimmInfo[key] = value
			}
		}

		// Only process populated DIMMs
		if size != "No Module Installed" && size != "" {
			populatedDIMMs = append(populatedDIMMs, dimmInfo)
			var channelID string
			// Try to find channel info in the locator string
			matches := explicitChannelRegex.FindStringSubmatch(banklocator)
			if len(matches) > 1 {
				channelID = strings.ToUpper(matches[1])
				// channels[channelID] = true
			}

			if channelID == "" {
				matches = structuredLocatorRegex.FindStringSubmatch(locator)
				if len(matches) > 1 {
					channelID = matches[1]
				}
			}

			if channelID == "" {
				matches = simpleLocatorRegex.FindStringSubmatch(locator)
				if len(matches) > 1 {
					channelID = matches[1]
				}
			}

			if channelID != "" {
				channels[strings.ToUpper(channelID)] = true
			}
		}
	}

	channelList := make([]string, 0, len(channels))
	for ch := range channels {
		channelList = append(channelList, ch)
	}
	sort.Strings(channelList)

	data["total_memory_slots_populated"] = len(populatedDIMMs)
	data["detected_memory_channels"] = len(channels)
	data["populated_dimm_details"] = populatedDIMMs
	if len(channels) > 0 {
		data["detected_channel_names"] = channelList
	}
	slog.Info("DIMM info collection", "status", "complete")
	return data

}
