// Copyright (c) 2026, Ampere Computing LLC.
// 
// SPDX-License-Identifier: BSD-3-Clause

package collectors

import (
	"reflect"
	"testing"
)

func TestParseDmidecodeMemoryCapturesLocatorAndChannels(t *testing.T) {
	output := `
Memory Device
	Size: 64 GB
	Locator: DIMM_P0_A0
	Bank Locator: P1-CH A
	Type: DDR5
	Speed: 4800 MT/s
	Configured Memory Speed: 4800 MT/s
	Manufacturer: Ampere
	Part Number: PN-A0

Memory Device
	Size: 64 GB
	Locator: DIMM_P0_C1
	Bank Locator: P1-CH C
	Type: DDR5
	Speed: 4800 MT/s
	Configured Memory Speed: 4800 MT/s
	Manufacturer: Ampere
	Part Number: PN-C1

Memory Device
	Size: No Module Installed
	Locator: DIMM_P0_B0
	Bank Locator: P1-CH B
`

	data := parseDmidecodeMemory(output)

	if got, want := data["total_memory_slots_populated"], 2; got != want {
		t.Fatalf("total_memory_slots_populated = %v, want %v", got, want)
	}

	if got, want := data["detected_memory_channels"], 2; got != want {
		t.Fatalf("detected_memory_channels = %v, want %v", got, want)
	}

	gotChannels, ok := data["detected_channel_names"].([]string)
	if !ok {
		t.Fatalf("detected_channel_names has unexpected type %T", data["detected_channel_names"])
	}
	if want := []string{"A", "C"}; !reflect.DeepEqual(gotChannels, want) {
		t.Fatalf("detected_channel_names = %v, want %v", gotChannels, want)
	}

	gotDIMMs, ok := data["populated_dimm_details"].([]map[string]string)
	if !ok {
		t.Fatalf("populated_dimm_details has unexpected type %T", data["populated_dimm_details"])
	}
	if len(gotDIMMs) != 2 {
		t.Fatalf("len(populated_dimm_details) = %d, want 2", len(gotDIMMs))
	}

	if gotDIMMs[0]["Locator"] != "DIMM_P0_A0" {
		t.Fatalf("first DIMM Locator = %q, want %q", gotDIMMs[0]["Locator"], "DIMM_P0_A0")
	}
	if gotDIMMs[1]["Bank Locator"] != "P1-CH C" {
		t.Fatalf("second DIMM Bank Locator = %q, want %q", gotDIMMs[1]["Bank Locator"], "P1-CH C")
	}
}
