// Copyright (c) 2026, Ampere Computing LLC.
// 
// SPDX-License-Identifier: BSD-3-Clause

package reporters

import (
	"strings"
	"testing"
)

func TestHTMLReporterReportRendersNestedData(t *testing.T) {
	reporter := &HTMLReporter{}
	data := map[string]interface{}{
		"system_info": map[string]interface{}{
			"hostname": "node<1>",
			"devices": []map[string]string{
				{
					"Slot":   "0000:01:00.0",
					"Vendor": "Ampere & Co",
					"Device": "Ampere PCIe Adapter",
				},
			},
		},
		"network_info": map[string]interface{}{
			"interfaces": map[string]interface{}{
				"eth0": map[string]interface{}{
					"ipv4_addresses": []string{"192.0.2.10/24"},
					"state":          "up",
				},
			},
		},
	}

	output, err := reporter.Report(data)
	if err != nil {
		t.Fatalf("Report returned error: %v", err)
	}

	htmlOutput := string(output)
	for _, want := range []string{
		"<!DOCTYPE html>",
		"Ampere System Info Report",
		"overflow: hidden;",
		`<main class="content">`,
		`<header class="hero">`,
		`<div class="section-stack">`,
		"--ampere-red: #c8102e;",
		`<dl class="definition-list">`,
		`<summary class="collection-summary">Show 1 interface</summary>`,
		`<details class="named-entry"><summary class="named-entry-summary"><span class="named-entry-title">eth0</span>`,
		`<span class="named-entry-meta">up`,
		`<summary class="collection-summary">Show 1 PCIe device</summary>`,
		`<details class="named-entry"><summary class="named-entry-summary"><span class="named-entry-title">Ampere PCIe Adapter</span><span class="named-entry-meta">0000:01:00.0 · Ampere &amp; Co</span></summary>`,
		`<aside class="sidebar" aria-label="Subsystem navigation">`,
		`href="#section-network-info"`,
		`href="#section-system-info"`,
		`<section id="section-network-info">`,
		"<h2>Network</h2>",
		`<section id="section-system-info">`,
		"<h2>System</h2>",
		"node&lt;1&gt;",
		"Ampere &amp; Co",
		"192.0.2.10/24",
	} {
		if !strings.Contains(htmlOutput, want) {
			t.Fatalf("expected output to contain %q", want)
		}
	}

	if strings.Index(htmlOutput, `href="#section-system-info"`) > strings.Index(htmlOutput, `href="#section-network-info"`) {
		t.Fatal("expected system section to appear before network section")
	}

	for _, unwanted := range []string{
		`<th>Key</th>`,
		`<th>Value</th>`,
		`>Device</dt>`,
	} {
		if strings.Contains(htmlOutput, unwanted) {
			t.Fatalf("expected output not to contain %q", unwanted)
		}
	}
}

func TestHTMLReporterReportHandlesNilAndEmptyCollections(t *testing.T) {
	reporter := &HTMLReporter{}
	data := map[string]interface{}{
		"system_info": map[string]interface{}{
			"empty_list": []string{},
			"empty_map":  map[string]string{},
			"nil_value":  nil,
		},
	}

	output, err := reporter.Report(data)
	if err != nil {
		t.Fatalf("Report returned error: %v", err)
	}

	htmlOutput := string(output)
	for _, want := range []string{
		`<span class="empty">[]</span>`,
		`<span class="empty">{}</span>`,
		`<span class="empty">null</span>`,
	} {
		if !strings.Contains(htmlOutput, want) {
			t.Fatalf("expected output to contain %q", want)
		}
	}
}

func TestHTMLReporterReportUsesConfiguredSubsystemOrder(t *testing.T) {
	reporter := &HTMLReporter{}
	data := map[string]interface{}{
		"pmu_info":      map[string]interface{}{"status": "ok"},
		"software_info": map[string]interface{}{"docker_version": "26.1"},
		"network_info":  map[string]interface{}{"interfaces": map[string]interface{}{}},
		"disk_info":     map[string]interface{}{"partitions": map[string]interface{}{}},
		"mem_info":      map[string]interface{}{"MemTotal": "1 TB"},
		"dimm_info":     map[string]interface{}{"detected_channel_names": []string{"A"}},
		"cpu_info":      map[string]interface{}{"Model name": "AmpereOne"},
		"system_info":   map[string]interface{}{"hostname": "node-1"},
	}

	output, err := reporter.Report(data)
	if err != nil {
		t.Fatalf("Report returned error: %v", err)
	}

	htmlOutput := string(output)
	orderedLinks := []string{
		`href="#section-system-info"`,
		`href="#section-cpu-info"`,
		`href="#section-memory-info"`,
		`href="#section-disk-info"`,
		`href="#section-network-info"`,
		`href="#section-software-info"`,
		`href="#section-pmu-info"`,
	}

	lastIndex := -1
	for _, want := range orderedLinks {
		idx := strings.Index(htmlOutput, want)
		if idx < 0 {
			t.Fatalf("expected output to contain %q", want)
		}
		if idx < lastIndex {
			t.Fatalf("expected output to keep configured subsystem order, found %q out of place", want)
		}
		lastIndex = idx
	}

	for _, unwanted := range []string{
		`href="#section-dimm-info"`,
		`href="#section-mem-info"`,
	} {
		if strings.Contains(htmlOutput, unwanted) {
			t.Fatalf("expected merged memory subsystem to omit %q", unwanted)
		}
	}
}

func TestHTMLReporterReportMergesMemoryAndDIMMSections(t *testing.T) {
	reporter := &HTMLReporter{}
	data := map[string]interface{}{
		"mem_info": map[string]interface{}{
			"MemTotal": "1 TB",
		},
		"dimm_info": map[string]interface{}{
			"detected_channel_names":       []string{"A"},
			"detected_memory_channels":     1,
			"total_memory_slots_populated": 2,
		},
	}

	output, err := reporter.Report(data)
	if err != nil {
		t.Fatalf("Report returned error: %v", err)
	}

	htmlOutput := string(output)
	for _, want := range []string{
		`href="#section-memory-info"`,
		`<section id="section-memory-info">`,
		`<h2>Memory</h2>`,
		`<h3>Memory Info</h3>`,
		`<h3>DIMM Info</h3>`,
		"MemTotal",
		`<p class="channel-title">Populated Channels</p>`,
	} {
		if !strings.Contains(htmlOutput, want) {
			t.Fatalf("expected output to contain %q", want)
		}
	}

	if strings.Contains(htmlOutput, `section-dimm-info`) || strings.Contains(htmlOutput, `section-mem-info`) {
		t.Fatal("expected separate DIMM and memory sections to be merged into one memory section")
	}

	if strings.Index(htmlOutput, `<h3>Memory Info</h3>`) > strings.Index(htmlOutput, `<h3>DIMM Info</h3>`) {
		t.Fatal("expected memory info to render before DIMM info inside the merged memory section")
	}
}

func TestHTMLReporterReportCollapsesDiskPartitions(t *testing.T) {
	reporter := &HTMLReporter{}
	data := map[string]interface{}{
		"disk_info": map[string]interface{}{
			"partitions": map[string]interface{}{
				"/": map[string]interface{}{
					"device":       "/dev/nvme0n1p2",
					"fstype":       "xfs",
					"free_bytes":   1000,
					"total_bytes":  2000,
					"used_percent": 50.0,
				},
				"/boot": map[string]interface{}{
					"device":       "/dev/nvme0n1p1",
					"fstype":       "vfat",
					"free_bytes":   100,
					"total_bytes":  200,
					"used_percent": 50.0,
				},
			},
		},
	}

	output, err := reporter.Report(data)
	if err != nil {
		t.Fatalf("Report returned error: %v", err)
	}

	htmlOutput := string(output)
	for _, want := range []string{
		`<summary class="collection-summary">Show 2 partitions</summary>`,
		`<details class="named-entry"><summary class="named-entry-summary"><span class="named-entry-title">/</span><span class="named-entry-meta">/dev/nvme0n1p2 · xfs</span></summary>`,
		`<details class="named-entry"><summary class="named-entry-summary"><span class="named-entry-title">/boot</span><span class="named-entry-meta">/dev/nvme0n1p1 · vfat</span></summary>`,
		"/dev/nvme0n1p2",
		"/dev/nvme0n1p1",
	} {
		if !strings.Contains(htmlOutput, want) {
			t.Fatalf("expected output to contain %q", want)
		}
	}
}

func TestHTMLReporterReportCollapsesLongNestedCollections(t *testing.T) {
	reporter := &HTMLReporter{}
	data := map[string]interface{}{
		"software_info": map[string]interface{}{
			"containers": []map[string]string{
				{"Name": "api", "Image": "example/api:1", "Status": "Up 2 hours"},
				{"Name": "worker", "Image": "example/worker:1", "Status": "Up 2 hours"},
				{"Name": "db", "Image": "postgres:16", "Status": "Up 1 day"},
				{"Name": "cache", "Image": "redis:7", "Status": "Exited (0) 1 hour ago"},
			},
		},
	}

	output, err := reporter.Report(data)
	if err != nil {
		t.Fatalf("Report returned error: %v", err)
	}

	htmlOutput := string(output)
	for _, want := range []string{
		`<details class="collection-toggle">`,
		`<summary class="collection-summary">Show 4 containers</summary>`,
		`<details class="named-entry"><summary class="named-entry-summary"><span class="named-entry-title">worker</span><span class="named-entry-meta">example/worker:1 · Up 2 hours</span></summary>`,
		"example/worker:1",
	} {
		if !strings.Contains(htmlOutput, want) {
			t.Fatalf("expected output to contain %q", want)
		}
	}

	if strings.Contains(htmlOutput, `>Name</dt>`) {
		t.Fatal("expected container name to be shown in the nested summary instead of repeated in the detail body")
	}
}

func TestHTMLReporterReportKeepsScalarCollectionsExpanded(t *testing.T) {
	reporter := &HTMLReporter{}
	data := map[string]interface{}{
		"network_info": map[string]interface{}{
			"ipv4_addresses": []string{
				"192.0.2.10/24",
				"192.0.2.11/24",
				"192.0.2.12/24",
				"192.0.2.13/24",
			},
		},
	}

	output, err := reporter.Report(data)
	if err != nil {
		t.Fatalf("Report returned error: %v", err)
	}

	htmlOutput := string(output)
	if strings.Contains(htmlOutput, `<details class="collection-toggle">`) {
		t.Fatal("expected scalar collections to remain expanded")
	}

	for _, want := range []string{
		`<ul>`,
		"192.0.2.13/24",
	} {
		if !strings.Contains(htmlOutput, want) {
			t.Fatalf("expected output to contain %q", want)
		}
	}
}

func TestHTMLReporterReportRendersDIMMChannelSummary(t *testing.T) {
	reporter := &HTMLReporter{}
	data := map[string]interface{}{
		"dimm_info": map[string]interface{}{
			"total_memory_slots_populated": 4,
			"detected_memory_channels":     2,
			"detected_channel_names":       []string{"A", "C"},
			"populated_dimm_details": []map[string]string{
				{"Bank Locator": "P1-CH A", "Locator": "DIMM_P0_A0", "Size": "64 GB"},
				{"Bank Locator": "P1-CH A", "Locator": "DIMM_P0_A1", "Size": "64 GB"},
				{"Bank Locator": "P1-CH C", "Locator": "DIMM_P0_C0", "Size": "64 GB"},
				{"Bank Locator": "P1-CH C", "Locator": "DIMM_P0_C1", "Size": "64 GB"},
			},
		},
	}

	output, err := reporter.Report(data)
	if err != nil {
		t.Fatalf("Report returned error: %v", err)
	}

	htmlOutput := string(output)
	for _, want := range []string{
		`<p class="channel-title">Populated Channels</p>`,
		`<span class="dimm-stat-label">Populated Slots</span><span class="dimm-stat-value">4</span>`,
		`<span class="dimm-stat-label">Populated Channels</span><span class="dimm-stat-value">2</span>`,
		`<span class="channel-chip">A</span>`,
		`<span class="channel-chip">C</span>`,
		`<span class="channel-card-count">2 DIMMs</span>`,
		`<span class="channel-slot">DIMM_P0_A0 · 64 GB</span>`,
		`<span class="channel-slot">DIMM_P0_C1 · 64 GB</span>`,
	} {
		if !strings.Contains(htmlOutput, want) {
			t.Fatalf("expected output to contain %q", want)
		}
	}

	if strings.Contains(htmlOutput, ">detected_channel_names</td>") {
		t.Fatal("expected DIMM channel names to be rendered as the visual summary instead of a raw table row")
	}
}
