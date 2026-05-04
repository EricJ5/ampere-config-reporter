// Copyright (c) 2026, Ampere Computing LLC.
// 
// SPDX-License-Identifier: BSD-3-Clause

package collectors

import (
	"bufio"
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type CPUCollector struct{}

func (c *CPUCollector) Name() string {
	return "cpu_info"
}

// Collect gathers data from both lscpu and the /sys filesystem.
func (c *CPUCollector) Collect() (map[string]interface{}, error) {
	// 1. Get static architectural info from lscpu
	cmd := exec.Command("lscpu")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	} else {
		data := parseLSCPUOutput(output)
		// 2. Get dynamic frequency and governor info
		freqData, err := getCPUFreqInfo()
		if err != nil {
			// Log the error but don't fail the whole collector, as lscpu data is still valuable
			slog.Error("Unable to get cpu frequnecy data: %v\n", err.Error(), "cpu")
		} else {
			data["frequency_governors"] = freqData
		}

		slcData, err := getSlcInfo()
		if err != nil {
			slog.Error("SLC error", "err", err)
		}
		data["slc_info"] = slcData

		// get cpu part info
		cpuPart, err := getCpuPart()
		if err != nil || cpuPart == "" {
			data["cpu_part"] = "Not Collected"
		}
		data["cpu_part"] = cpuPart

		slog.Info("CPU info collection", "status", "complete")
		return data, nil
	}
}

func getCpuPart() (string, error) {

	dmiOutput, dmiErr := getDMI()
	if dmiErr != nil {
		return "", dmiErr
	}
	if processorList, ok := dmiOutput["Processor Information"]; ok {
		// slog.Debug("debugging multi socket", "length", len(dmiOutput))
		if len(processorList) > 0 {
			socket0 := processorList[0]
			if partNumber, ok := socket0["Part Number"].(string); ok {
				return string(partNumber), nil
			}
		}
	}
	return "", nil

}

// parse lscpu
func parseLSCPUOutput(output []byte) map[string]interface{} {
	// Using a map as a set for efficient lookups.
	desiredFields := map[string]bool{
		"Architecture":             true,
		"Model name":               true,
		"NUMA node(s)":             true,
		"Socket(s)":                true,
		"Core(s) per socket":       true,
		"Stepping":                 true,
		"L1d cache":                true,
		"L1i cache":                true,
		"L2 cache":                 true,
		"Vulnerability Spectre v1": true,
		"Vulnerability Spectre v2": true,
		"Vulnerability Meltdown":   true,
	}
	data := make(map[string]interface{})
	var currentSection string
	var sectionData map[string]interface{}

	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		if strings.HasSuffix(line, ":") && len(strings.Split(line, ":")) == 2 && strings.TrimSpace(strings.Split(line, ":")[1]) == "" {
			sectionName := strings.TrimSuffix(line, ":")
			normalizedSectionName := normalizeKey(sectionName)

			currentSection = normalizedSectionName
			sectionData = make(map[string]interface{})
			data[currentSection] = sectionData
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Only proceed if the key is in our desiredFields set
		if _, ok := desiredFields[key]; !ok {
			continue
		}

		if currentSection != "" && strings.HasPrefix(line, "  ") {
			if sectionData != nil {
				sectionData[key] = value
			}
		} else {
			currentSection = ""
			data[key] = value
		}
	}

	return data
}

// normalizeKey remains the same as before.
func normalizeKey(s string) string {
	s = strings.ToLower(s)
	re := regexp.MustCompile(`\s*\(.*?\)`)
	s = re.ReplaceAllString(s, "")
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, "-", "_")
	return s
}

// getCPUFreqInfo reads CPU frequency and governor details from /sys
// It's more efficient to read from policy0 directories than every cpu*
func getCPUFreqInfo() (map[string]interface{}, error) {
	policyPaths, err := filepath.Glob("/sys/devices/system/cpu/cpufreq/policy0")
	if err != nil {
		return nil, err
	}
	if len(policyPaths) == 0 {
		return map[string]interface{}{"status": "cpufreq interface not found in /sys"}, nil
	}

	policyData := make(map[string]interface{})
	for _, policyPath := range policyPaths {
		policyID := filepath.Base(policyPath)

		// Helper function to read a file from the policy directory
		readFile := func(name string) string {
			content, err := os.ReadFile(filepath.Join(policyPath, name))
			if err != nil {
				return "N/A"
			}
			return strings.TrimSpace(string(content))
		}
		curFreq, _ := strconv.ParseFloat(readFile("scaling_cur_freq"), 64)
		minFreq, _ := strconv.ParseFloat(readFile("scaling_min_freq"), 64)
		maxFreq, _ := strconv.ParseFloat(readFile("scaling_max_freq"), 64)

		policyData[policyID] = map[string]string{
			"scaling_governor":     readFile("scaling_governor"),
			"scaling_cur_freq_mhz": fmt.Sprint(curFreq / 1000),
			"scaling_min_freq_mhz": fmt.Sprint(minFreq / 1000),
			"scaling_max_freq_mhz": fmt.Sprint(maxFreq / 1000),
		}
	}
	return policyData, nil
}

func getSlcInfo() (string, error) {
	cacheDir := "/sys/devices/system/cpu/cpu0/cache"
	indices, err := os.ReadDir(cacheDir)
	var slcData string
	if err != nil {
		return "", err
	}

	for _, index := range indices {
		indexPath := filepath.Join(cacheDir, index.Name())
		levelBytes, err := os.ReadFile(filepath.Join(indexPath, "level"))
		if err != nil {
			continue
		}

		level, err := strconv.Atoi(strings.TrimSpace(string(levelBytes)))
		if err != nil {
			continue
		}
		typeBytes, err := os.ReadFile(filepath.Join(indexPath, "type"))
		if err != nil {
			continue
		}
		cacheType := strings.TrimSpace(string(typeBytes))
		slog.Debug("processing indexPath: %s, level: %v, cacheType: %s\n", indexPath, level, cacheType, "cpu")
		if level >= 3 && (cacheType == "Unified" || cacheType == "WBLookaside") {
			sizeBytes, err := os.ReadFile(filepath.Join(indexPath, "size"))
			if err != nil {
				return "", err
			}
			slcData = strings.TrimSpace(string(sizeBytes))

		} else {
			slog.Debug("SLC not detected", "cache level: %v", level)
			slcData = "Not Detected"
		}

	}

	return slcData, nil
}
