// Copyright (c) 2026, Ampere Computing LLC.
// 
// SPDX-License-Identifier: BSD-3-Clause

package collectors

import (
	"encoding/binary"
	"fmt"
	"log/slog"
	"os/exec"
	"regexp"
	"strings"
	"unsafe"

	"golang.org/x/sys/unix"
)

type PmuCollector struct{}

const (
	PERF_TYPE_HARDWARE                = 0
	PERF_COUNT_HW_CPU_CYCLES   uint64 = 0
	PERF_COUNT_HW_INSTRUCTIONS uint64 = 1

	PERF_EVENT_IOC_ENABLE  = 0x2400
	PERF_EVENT_IOC_DISABLE = 0x2401
	PERF_EVENT_IOC_RESET   = 0x2403
)

func (c *PmuCollector) Name() string {
	return "pmu_info"
}

func ioctl(fd int, req uintptr) error {
	_, _, errno := unix.Syscall(unix.SYS_IOCTL, uintptr(fd), req, 0)
	if errno != 0 {
		return errno
	}
	return nil
}

func readCounter(fd int) (uint64, error) {
	var b [8]byte
	n, err := unix.Read(fd, b[:])
	if err != nil {
		return 0, err
	}
	if n != 8 {
		return 0, fmt.Errorf("short read: %d", n)
	}
	return binary.LittleEndian.Uint64(b[:]), nil

}

func openHWCounter(config uint64) (int, error) {
	attr := &unix.PerfEventAttr{
		Type:   unix.PERF_TYPE_HARDWARE,
		Size:   uint32(unsafe.Sizeof(unix.PerfEventAttr{})),
		Config: config,
		Bits: unix.PerfBitDisabled |
			unix.PerfBitExcludeKernel |
			unix.PerfBitExcludeHv,
		// Flags:  attrDisabled | attrExcludeKernel | attrExcludeHV, //user only
	}
	fd, err := unix.PerfEventOpen(attr, 0, -1, -1, 0)
	if err != nil {
		slog.Debug("call to perf event open failed", "err", err)
		return -1, err
	}
	return fd, nil
}

func ReadCyclesAndInstructions() (cycles, insns uint64, err error) {
	cfd, err := openHWCounter(unix.PERF_COUNT_HW_CPU_CYCLES)
	var cerr error
	var ierr error
	if err != nil {
		slog.Debug("failed to open HW counter", "event", "cycles")
		// slog.Debug("", "",err)
		return 0, 0, fmt.Errorf("open cycles perf event: %w", err)
	}
	defer unix.Close(cfd)

	ifd, err := openHWCounter(unix.PERF_COUNT_HW_INSTRUCTIONS)
	if err != nil {
		slog.Debug("failed to open HW counter", "event", "insns")
		return 0, 0, fmt.Errorf("open instructions perf event: %w", err)
	}
	defer unix.Close(ifd)

	_ = ioctl(cfd, PERF_EVENT_IOC_RESET)
	_ = ioctl(ifd, PERF_EVENT_IOC_RESET)

	if err := ioctl(cfd, PERF_EVENT_IOC_ENABLE); err != nil {
		return 0, 0, fmt.Errorf("enable cycles: %w", err)
	}

	if err := ioctl(ifd, PERF_EVENT_IOC_ENABLE); err != nil {
		return 0, 0, fmt.Errorf("enable insns: %w", err)
	}

	_ = ioctl(cfd, PERF_EVENT_IOC_DISABLE)
	_ = ioctl(ifd, PERF_EVENT_IOC_DISABLE)

	cycles, cerr = readCounter(cfd)
	insns, ierr = readCounter(ifd)
	if cerr != nil && ierr != nil {
		return 0, 0, err
	} else if cerr != nil {
		return 0, insns, cerr
	} else if ierr != nil {
		return cycles, 0, ierr
	}
	return cycles, insns, nil

}

func (c *PmuCollector) Collect() (map[string]interface{}, error) {
	perfRegex := regexp.MustCompile(`perf version (.*)`)
	data := make(map[string]interface{})
	perfPath, err := exec.LookPath("perf")
	perfCollect := false
	if err != nil {
		slog.Debug("perf exec not available", "err", err)
		data["linux-perf"] = "Not Detected"
		return data, err
	}
	slog.Debug("perf exec detected", "perfPath", perfPath)

	versionCmd := exec.Command(perfPath, "--version")
	perfVersion, err := versionCmd.CombinedOutput()
	if err != nil {
		slog.Debug("Unable to get pet perf version", "error", err)
		data["linux-perf"] = "Not Detected"
	} else {
		data["linux-perf"] = "Available"
		versionString := strings.TrimSpace(string(perfVersion))
		matches := perfRegex.FindStringSubmatch(versionString)
		if len(matches) < 2 {
			slog.Debug("unable to parse perf version", "versionraw", versionString)
		} else {
			perfCollect = true
			data["perf-version"] = matches[1]
		}
	}

	if perfCollect {
		cycles, insns, err := ReadCyclesAndInstructions()
		if err != nil {
			slog.Debug("PMU via perf_event_open not available", "err", err)
			data["pmu_enabled"] = false
			data["cycles"] = "Not collectable"
			data["instructions"] = "Not collectable"
		} else {
			data["pmu_enabled"] = true
			data["cycles"] = "Available"
			slog.Debug("Cycles fixed PMU counter collected successfully", "cycles", cycles)
			data["instructions"] = "Available"
			slog.Debug("Instructions PMU counter collected successfully", "cycles", insns)

		}

	}

	slog.Info("PMU info collection", "status", "complete")
	return data, nil
}
