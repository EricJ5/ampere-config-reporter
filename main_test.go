package main

import "testing"

func TestResolveOutputFile(t *testing.T) {
	testCases := []struct {
		name       string
		format     string
		outFile    string
		outFileSet bool
		want       string
	}{
		{name: "default json", format: "json", outFile: "acr.json", want: "acr.json"},
		{name: "default csv", format: "csv", outFile: "acr.json", want: "acr.csv"},
		{name: "default html", format: "html", outFile: "acr.json", want: "acr.html"},
		{name: "default text", format: "text", outFile: "acr.json", want: "acr.txt"},
		{name: "unknown format falls back to json", format: "unknown", outFile: "acr.json", want: "acr.json"},
		{name: "explicit output preserved", format: "html", outFile: "custom.out", outFileSet: true, want: "custom.out"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := resolveOutputFile(tc.format, tc.outFile, tc.outFileSet); got != tc.want {
				t.Fatalf("resolveOutputFile(%q, %q, %t) = %q, want %q", tc.format, tc.outFile, tc.outFileSet, got, tc.want)
			}
		})
	}
}
