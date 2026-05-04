// Copyright (c) 2026, Ampere Computing LLC.
// 
// SPDX-License-Identifier: BSD-3-Clause

package reporters

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"sort"
)

type CSVReporter struct{}

func (r *CSVReporter) Report(data map[string]interface{}) ([]byte, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	//write header
	if err := writer.Write([]string{"Category", "Key", "Value"}); err != nil {
		return nil, err
	}

	// sort categories for consistent output
	categories := make([]string, 0, len(data))
	for k := range data {
		categories = append(categories, k)
	}
	sort.Strings(categories)

	for _, category := range categories {
		flattenWrite(writer, category, "", data[category])
	}

	writer.Flush()
	return buf.Bytes(), writer.Error()
}

func flattenWrite(writer *csv.Writer, category, prefix string, val interface{}) {
	switch v := val.(type) {
	case map[string]interface{}:
		for key, value := range v {
			newPrefix := prefix + key + "."
			flattenWrite(writer, category, newPrefix, value)
		}
	case map[string]string:
		for key, value := range v {
			fullKey := prefix + key
			writer.Write([]string{category, fullKey, value})
		}
	case []map[string]string: // For PCIe devices
		for i, item := range v {
			for key, value := range item {
				newPrefix := fmt.Sprintf("%s%d.%s", prefix, i, key)
				writer.Write([]string{category, newPrefix, value})
			}
		}
	default:
		writer.Write([]string{category, prefix, fmt.Sprintf("%v", v)})
	}
}
