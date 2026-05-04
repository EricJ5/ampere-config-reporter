package collectors

import (
	"log/slog"
	"os/exec"
	"strings"
	"sync"
)

var (
	once      sync.Once
	dmiOutput []byte
	dmiErr    error
)

func loadDMI() {
	slog.Debug("Running dmidecode to get processor info the first time")
	cmd := exec.Command("sudo", "dmidecode", "-t", "processor")
	dmiOutput, dmiErr = cmd.CombinedOutput()
	if dmiErr != nil {
		slog.Debug("Unable to run dmidecode command", "err", dmiErr)
		return
	}
}

func getDMI() (map[string][]map[string]interface{}, error) {
	// this will be run only once and all other goroutines will wait until loadDMI is finished
	once.Do(loadDMI)
	dmiOut, err := parseDMI(string(dmiOutput))
	if err != nil {
		slog.Debug("unable parse dmidecode", "error", err)
		return nil, err
	}
	return dmiOut, err
}

func parseDMI(output string) (map[string][]map[string]interface{}, error) {
	result := make(map[string][]map[string]interface{})
	records := strings.Split(output, "\n\n")

	for _, record := range records {
		lines := strings.Split(record, "\n")
		if len(lines) < 3 {
			continue
		}
		recordTitle := strings.TrimSpace(lines[1])
		if recordTitle == "" {
			continue
		}

		data := make(map[string]interface{})
		var lastKey string

		for _, line := range lines[2:] {
			trimLine := strings.TrimSpace(line)
			if strings.HasPrefix(line, "\t\t") {
				if lastKey != "" {
					if slice, ok := data[lastKey].([]string); ok {
						data[lastKey] = append(slice, trimLine)
					}
				}
				continue
			}

			if strings.HasPrefix(line, "\t") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					val := strings.TrimSpace(parts[1])
					lastKey = key

					if val == "" {
						data[key] = []string{}
					} else {
						data[key] = val
					}
				}
			}
		}

		if len(data) > 0 {
			result[recordTitle] = append(result[recordTitle], data)
		}
	}
	return result, nil
}
