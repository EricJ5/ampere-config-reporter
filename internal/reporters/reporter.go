// Copyright (c) 2026, Ampere Computing LLC.
// 
// SPDX-License-Identifier: BSD-3-Clause

package reporters

// reporter is interface that all input formatters must implement

type Reporter interface {
	Report(data map[string]interface{}) ([]byte, error)
}
