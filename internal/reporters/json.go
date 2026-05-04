// Copyright (c) 2026, Ampere Computing LLC.
// 
// SPDX-License-Identifier: BSD-3-Clause

package reporters

import "encoding/json"

type JSONReporter struct{}

func (r *JSONReporter) Report(data map[string]interface{}) ([]byte, error) {
	return json.MarshalIndent(data, "", " ")
}
