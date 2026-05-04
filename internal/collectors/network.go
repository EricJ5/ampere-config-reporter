// Copyright (c) 2026, Ampere Computing LLC.
// 
// SPDX-License-Identifier: BSD-3-Clause

package collectors

import (
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

type NetworkCollector struct{}

func (c *NetworkCollector) Name() string {
	return "network_info"
}

func (c *NetworkCollector) Collect() (map[string]interface{}, error) {
	data := make(map[string]interface{})
	interfaces := make(map[string]interface{})

	ifDirs, err := os.ReadDir("/sys/class/net")
	if err != nil {
		slog.Error("Unable to read N/W Interface directories", "error", err)
		return nil, err
	}

	for _, ifDir := range ifDirs {
		ifName := ifDir.Name()

		// skip virtual interfaces like loopback (lo)
		isVirtual, _ := isVirtualInterface(ifName)
		if isVirtual {
			continue
		}
		ifData := make(map[string]interface{})

		// get basic info from sys
		getSysNetInfo(ifName, ifData)

		//get driver and firmware info from ethtool
		getEthtoolInfo(ifName, ifData)

		// get IP address
		getIPAddrInfo(ifName, ifData)

		interfaces[ifName] = ifData

	}
	data["interfaces"] = interfaces
	slog.Info("Network info collection", "status", "complete")
	slog.Debug("Network complete", "interfaces", interfaces)
	return data, nil
}

func isVirtualInterface(ifName string) (bool, error) {
	link, err := os.Readlink(filepath.Join("/sys/class/net", ifName))
	if err != nil {
		slog.Error("Unable to read interface link", "error", err)
		return false, err
	}
	return strings.Contains(link, "/virtual/"), nil
}

func getSysNetInfo(ifName string, data map[string]interface{}) {
	basePath := filepath.Join("/sys/class/net", ifName)
	readFile := func(file string) string {
		content, err := os.ReadFile(filepath.Join(basePath, file))
		if err != nil {
			slog.Error("Unable to read network properties of interface", "error", err)
			return ""
		}
		return strings.TrimSpace(string(content))
	}
	data["mac_address"] = readFile("address")
	data["speed"] = readFile("speed")
	data["state"] = readFile("operstate")
}

func getEthtoolInfo(ifName string, data map[string]interface{}) {
	cmd := exec.Command("ethtool", "-i", ifName)
	output, err := cmd.Output()
	if err != nil {
		slog.Debug("Unable to execute ethtool cmd", "error", err)
		return
	}

	for _, line := range strings.Split(string(output), "\n") {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) > 2 {
			key := parts[0]
			value := parts[1]
			if key == "driver" || key == "version" || strings.HasPrefix(key, "firmware") {
				data[strings.ReplaceAll(key, " ", "_")] = value
			}
		}
	}
}

func getIPAddrInfo(ifName string, data map[string]interface{}) {
	cmd := exec.Command("ip", "addr", "show", ifName)
	output, err := cmd.Output()
	if err != nil {
		slog.Debug("Unable to IP address", "error", err)
		return
	}
	var ipv4Addr []string
	var ipv6Addr []string

	re := regexp.MustCompile(`(?m)^\s+(inet6?)\s+([0-9a-fA-F:./]+)`)
	matches := re.FindAllStringSubmatch(string(output), -1)

	for _, match := range matches {
		if len(match) == 3 {
			switch match[1] {
			case "inet":
				ipv4Addr = append(ipv4Addr, match[2])
			case "inet6":
				ipv6Addr = append(ipv6Addr, match[2])
			}
		}
	}
	data["ipv4_addresses"] = ipv4Addr
	data["ipv6_addresses"] = ipv6Addr
}
