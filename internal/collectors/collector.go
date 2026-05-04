// Copyright (c) 2026, Ampere Computing LLC.
// 
// SPDX-License-Identifier: BSD-3-Clause

package collectors

// Collector is the interface where all information collector must implement

type Collector interface {
	// Name returns the type of the collector
	Name() string

	//Collect runs the collector and returns collected data/errors if any.
	Collect() (map[string]interface{}, error)
}
