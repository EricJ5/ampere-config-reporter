// Copyright (c) 2026, Ampere Computing LLC.
// 
// SPDX-License-Identifier: BSD-3-Clause

package collectors

import (
	"fmt"
	"log"
	"log/slog"
	"os/exec"
	"regexp"
	"strings"

	docker "github.com/fsouza/go-dockerclient"
	// "github.com/moby/moby/client"
)

type SoftwareCollector struct{}

var gccVersionPatterns = []*regexp.Regexp{
	regexp.MustCompile(`^gcc version\s+([^\s)]+)`),
	regexp.MustCompile(`^gcc\s+\([^)]*\)\s+([^\s)]+)`),
}

func (c *SoftwareCollector) Name() string {
	return "software_info"
}

func (c *SoftwareCollector) Collect() (map[string]interface{}, error) {
	data := make(map[string]interface{})
	// GCC info
	cmd := exec.Command("gcc", "--version")
	var gccVersion string
	gccOutput, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error running gcc command: %v\n", err)
		// Try -v flag if --version fails (some systems might prefer it for detailed info)
		cmd = exec.Command("gcc", "-v")
		gccOutput, err = cmd.CombinedOutput()
		if err != nil {
			slog.Error("Error running gcc -v command:", "error", err)
			gccVersion = "Not Detected"
		}
	}
	lines := strings.Split(string(gccOutput), "\n")
	gccVersion = parseGCCVersion(lines)
	if gccVersion == "" {
		gccVersion = "Not Detected"
	}
	data["gcc_version"] = gccVersion
	// glibc
	lddCmd := exec.Command("ldd", "--version")
	var glibcVersion string
	output, err := lddCmd.CombinedOutput()
	if err != nil {
		slog.Error("glibc unavailable", "error", err)
		glibcVersion = "Not Detected"
	} else {
		// The output is typically on the first line
		glibc := strings.Split(string(output), "\n")[0]
		glibc_version := strings.Split(glibc, " ")
		glibcVersion = glibc_version[len(glibc_version)-1]
	}
	data["glibc_version"] = glibcVersion

	// openssl
	sslCmd := exec.Command("openssl", "version")
	var sslVersion string
	sslout, err := sslCmd.CombinedOutput()
	if err != nil {
		slog.Info("openssl isn't available", "error", err)
		sslVersion = "Not Detected"
	} else {
		sslVersion = strings.Split(string(sslout), " ")[1]
	}
	data["openssl_version"] = sslVersion

	// golang
	goCmd := exec.Command("go", "env", "GOVERSION")
	goOut, err := goCmd.CombinedOutput()
	var goVersion string
	if err != nil {
		slog.Info("golang not available", "error", err)
		goVersion = "Not Detected"
	} else {
		goVersion = strings.TrimSpace(string(goOut))
	}
	data["go_version"] = goVersion
	// llvm
	llvm := exec.Command("llvm-config", "--version")
	var llvmVersion string
	llvmOut, err := llvm.CombinedOutput()
	if err != nil {
		slog.Info("llvm-config unavailable", "error", err)
		llvmVersion = "Not Detected"
	} else {
		llvmVersion = strings.TrimSpace(string(llvmOut))
	}
	data["llvm_version"] = llvmVersion

	// Java
	java := exec.Command("java", "--version")
	var javaVersion string
	javaOut, err := java.CombinedOutput()
	if err != nil {
		log.Printf("Java not installed: %v", err)
		javaVersion = "Not Detected"
	} else {
		javaVersion = parseJavaVersion(javaOut)
	}
	data["java_version"] = javaVersion

	// Tuned-adm
	adm := exec.Command("sudo", "tuned-adm", "active")
	var admStatus string
	admOp, err := adm.Output()
	if err != nil {
		slog.Info("unable to collect tuned-adm", "error", err)
		admStatus = "Not Detected"

	} else {
		opStr := strings.TrimSpace(string(admOp))
		parts := strings.SplitN(opStr, ":", 2)
		if len(parts) != 2 {
			slog.Error("unable to parse tuned-adm profile status", "profile-string", parts)
			admStatus = "Not Detected"
		}
		admStatus = strings.TrimSpace(parts[1])

	}
	data["tuned-adm"] = admStatus

	// python version
	version, err := getPythonVersion("python")
	if err != nil {
		slog.Debug("Couldn't get version for python", "error", err)
		// try python3
		version, err = getPythonVersion("python3")
		if err != nil {
			slog.Info("Unable to detect either python or python3 version", "err", err)
			version = "Not Detected"
		}
	}
	data["python_version"] = version

	// docker info & container list

	dockerVersion, containerList := getDockerInfo()
	data["docker_version"] = dockerVersion
	if len(containerList) > 0 {
		data["containers"] = containerList
	}

	slog.Info("Software info collection", "status", "complete")
	return data, nil
}

func parseJavaVersion(input []byte) string {
	version := "Not Detected"
	lines := strings.Split(string(input), "\n")
	if len(lines) > 0 {
		firstLine := lines[0]
		parts := strings.Split(firstLine, " ")
		if len(parts) >= 2 {
			version = parts[1]
		}
	} else {
		slog.Error("Could not parse version", "from output: ", string(input))
	}
	return version
}

func parseGCCVersion(lines []string) string {
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		for _, pattern := range gccVersionPatterns {
			matches := pattern.FindStringSubmatch(line)
			if len(matches) == 2 {
				return strings.TrimRight(matches[1], ")")
			}
		}
	}

	return ""
}

func getPythonVersion(cmdName string) (string, error) {
	versionCmd := exec.Command(cmdName, "-V")
	output, err := versionCmd.Output()
	if err != nil {
		slog.Debug("unable to execute command", "cmd", cmdName)
		return "", err
	}

	versionString := strings.TrimSpace(string(output))
	parts := strings.Fields(versionString)
	if len(parts) >= 2 {
		return parts[1], nil
	}

	return versionString, nil

}

func getDockerInfo() (string, []map[string]interface{}) {
	versionString := "Not Detected"
	cli, err := docker.NewClientFromEnv()
	if err != nil {
		slog.Info("Failed to create docker client", "error", err)
		return "Not Detected", nil
	}
	env, err := cli.Version()
	if err == nil {
		versionString = env.Get("Version")
	}
	opts := docker.ListContainersOptions{All: true}
	containers, err := cli.ListContainers(opts)
	if err != nil {
		slog.Error("Failed to list containers", "error", err)
		return versionString, nil
	}
	var allContainerInfo []map[string]interface{}

	for _, container := range containers {
		containerInfo := make(map[string]interface{})
		containerInfo["Image"] = container.Image
		containerInfo["ID"] = container.ID[:10]
		containerInfo["Status"] = container.Status
		if len(container.Names) > 0 {
			containerInfo["Name"] = strings.TrimPrefix(container.Names[0], "/")
		}
		allContainerInfo = append(allContainerInfo, containerInfo)
	}

	return versionString, allContainerInfo

}
