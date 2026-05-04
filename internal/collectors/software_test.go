// Copyright (c) 2026, Ampere Computing LLC.
// 
// SPDX-License-Identifier: BSD-3-Clause

package collectors

import "testing"

func TestParseGCCVersion(t *testing.T) {
	tests := []struct {
		name  string
		lines []string
		want  string
	}{
		{
			name: "gcc dash dash version output",
			lines: []string{
				"gcc (Ubuntu 14.3.0-1ubuntu1~24.04) 14.3.0",
			},
			want: "14.3.0",
		},
		{
			name: "gcc dash v output trims trailing parenthesis release metadata",
			lines: []string{
				"Using built-in specs.",
				"gcc version 14.3.1 20250425 (Red Hat 14.3.1-4)",
			},
			want: "14.3.1",
		},
		{
			name: "missing gcc version",
			lines: []string{
				"some unrelated output",
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseGCCVersion(tt.lines); got != tt.want {
				t.Fatalf("parseGCCVersion() = %q, want %q", got, tt.want)
			}
		})
	}
}
