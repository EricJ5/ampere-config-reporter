package main

import (
	_ "embed"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"

	"github.com/AmpereComputing/ampere-config-reporter/internal/collectors"
	"github.com/AmpereComputing/ampere-config-reporter/internal/reporters"
)

//go:embed _version.txt
var embeddedVersion string

func version() string {
	v := strings.TrimSpace(embeddedVersion)
	if v == "" {
		return "dev"
	}
	return v
}

// Result struct
type Result struct {
	Name  string
	Data  map[string]interface{}
	Error error
}

func defaultOutputFile(format string) string {
	switch format {
	case "csv":
		return "acr.csv"
	case "html":
		return "acr.html"
	case "text":
		return "acr.txt"
	default:
		return "acr.json"
	}
}

func resolveOutputFile(format, outFile string, outFileSet bool) string {
	if outFileSet {
		return outFile
	}
	return defaultOutputFile(format)
}

func isFlagSet(name string) bool {
	var found bool
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

func main() {
	outFile := flag.String("o", "acr.json", "output file to write the report to (default: acr.<format>)")
	format := flag.String("format", "json", "ouput format: json, csv, text or html")
	debugFlag := flag.Bool("debug", false, "set debug logs")
	showVersion := flag.Bool("v", false, "show tool version")
	flag.BoolVar(showVersion, "version", false, "print version")
	flag.Parse()
	*outFile = resolveOutputFile(*format, *outFile, isFlagSet("o"))
	var logLevel slog.LevelVar
	if *debugFlag {
		logLevel.Set(slog.LevelDebug)
	} else {
		logLevel.Set(slog.LevelInfo)
	}

	if *showVersion {
		fmt.Println("Ampere-Config-Reporter (ACR)", version())
		return
	}

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: &logLevel,
	})

	logger := slog.New(handler)
	slog.SetDefault(logger)
	slog.Info("Ampere Config Reporter", "debug-log", *debugFlag)
	allCollectors := []collectors.Collector{
		&collectors.SystemCollector{},
		&collectors.CPUCollector{},
		&collectors.SoftwareCollector{},
		&collectors.MemoryCollector{},
		&collectors.DimmCollector{},
		&collectors.DiskCollector{},
		&collectors.NetworkCollector{},
		&collectors.PmuCollector{},
	}

	// map available reporters
	reportFormats := map[string]reporters.Reporter{
		"json": &reporters.JSONReporter{},
		"csv":  &reporters.CSVReporter{},
		"html": &reporters.HTMLReporter{},
	}

	reporter, ok := reportFormats[*format]
	if !ok {
		slog.Error("Error: Invalid format, Available formats: json, text, csv, html", "format", *format)
	}

	resultsChan := make(chan Result, len(allCollectors))
	var wg sync.WaitGroup

	//Launch a goroutine for each collector

	for _, collector := range allCollectors {
		wg.Add(1)
		go func(c collectors.Collector) {
			defer wg.Done()
			data, err := c.Collect()
			resultsChan <- Result{
				Name:  c.Name(),
				Data:  data,
				Error: err,
			}
		}(collector)
	}

	//wait for all collectors to finish, then close the channel
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	//aggregate the results
	finalReport := make(map[string]interface{})
	for result := range resultsChan {
		if result.Error != nil {
			slog.Error("Error from collector '%s': %v", result.Name, result.Error)
			finalReport[result.Name] = map[string]string{"error": result.Error.Error()}
		} else {
			finalReport[result.Name] = result.Data
		}
	}

	//print final report as JSON for now
	//TODO: extend to other formats

	output, err := reporter.Report(finalReport)
	if err != nil {
		slog.Error("Failed to generate report", "error", err)
	}

	werr := os.WriteFile(*outFile, output, 0666)
	if werr != nil {
		slog.Error("Failed to write report", "error", werr)
	}
	slog.Info("Report successfully written", "file", *outFile)

}
