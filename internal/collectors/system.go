// Copyright (c) 2026, Ampere Computing LLC.
// 
// SPDX-License-Identifier: BSD-3-Clause

package collectors

import (
	"bufio"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

type SystemCollector struct{}

func (c *SystemCollector) Name() string {
	return "system_info"
}

func (c *SystemCollector) Collect() (map[string]interface{}, error) {
	data := make(map[string]interface{})

	//get kernel version
	cmd := exec.Command("uname", "-r")

	// Run the command and get the output.
	kernelVersion, err := cmd.Output()

	if err != nil {
		// Log any errors that occur during execution
		slog.Error("Command failed", "with error: %s", err)
		data["kernel_version"] = "Not Collected"
	} else {
		data["kernel_version"] = strings.TrimSpace(string(kernelVersion))
	}

	// get page size
	pagesize := syscall.Getpagesize()
	data["kernel_pagesize"] = fmt.Sprintf("%d%s", pagesize/1024, "K")

	// get distro information

	distroInfo, err := parseOsRelease()
	if err != nil {
		slog.Error("failed to parse", "OS distro information: %v", err.Error())
		data["os_name"] = "Not Collected"
	} else {
		data["os_name"] = distroInfo

	}

	// get bootparams
	file, err := os.Open("/proc/cmdline")
	if err != nil {
		slog.Debug("Unable to get boot params", "unable to open file: %v", file)
		slog.Debug("Unable to get boot params", "error: %v", err)
		// data[""]

	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// bootdata := make(map[string]string)

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := parts[0]
			value := parts[1]
			data[key] = value
		}
	}

	// get bios info
	dmiPath := "/sys/class/dmi/id/"

	// check if the path exists
	if _, err := os.Stat(dmiPath); os.IsNotExist(err) {
		slog.Error("Unable to get bios info", "dmi path %v doesn't exist", dmiPath)

	}
	readFile := func(fileName string) string {
		content, err := os.ReadFile(filepath.Join(dmiPath, fileName))
		if err != nil {
			slog.Debug("Unable to read", "filename %v", fileName)
			return ""
		}
		return strings.TrimSpace(string(content))
	}

	data["bios_vendor"] = readFile("bios_vendor")
	data["bios_version"] = readFile("bios_version")
	data["bios_release_date"] = readFile("bios_date")
	data["bios_release"] = readFile("bios_release")

	// get baseboard/chassis info

	data["board_vendor"] = readFile("board_vendor")
	data["chassis_serial"] = readFile("chassis_serial")

	// get host info
	hostname, err := os.Hostname()
	if err != nil {
		slog.Error("Failed to get hostname", "%v", err)
	}
	data["hostname"] = hostname
	data["time"] = time.Now().Format(time.RFC1123Z)

	// get pcie info

	pciecmd := exec.Command("lspci", "-vmm")
	out, err := pciecmd.Output()
	if err != nil {
		log.Fatalf("lspci isn't available")
		return nil, err
	}

	devices := []map[string]string{}
	currentDevice := make(map[string]string)

	for _, line := range strings.Split(string(out), "\n") {
		if line == "" && len(currentDevice) > 0 {
			devices = append(devices, currentDevice)
			currentDevice = make(map[string]string)
			continue
		}
		parts := strings.SplitN(line, ":\t", 2)
		if len(parts) == 2 {
			currentDevice[parts[0]] = parts[1]
		}
	}
	if len(currentDevice) > 0 {
		devices = append(devices, currentDevice)
	}

	data["devices"] = devices
	slog.Info("System info collection", "status", "complete")
	return data, nil

}

func parseOsRelease() (string, error) {
	file, err := os.Open("/etc/os-release")
	if err != nil {
		slog.Error("unable to parse OS release info", "unable to open file and failed with error: %v", err)
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var data string

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		key := parts[0]
		if len(parts) == 2 && key == "PRETTY_NAME" {
			value := strings.Trim(parts[1], `"`)
			data = value
		}
	}

	return data, scanner.Err()
}
