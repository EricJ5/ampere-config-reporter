![ACR report](/images/acr_report.png)  

# Ampere-Config-Reporter(ACR): System Configuration Snapshot Tool

ACR is a lightweight, dependency-minimal Go tool designed to capture a comprehensive snapshot of a server's hardware and software configuration. Its primary purpose is to establish a consistent, repeatable baseline before running performance benchmarks, ensuring that results are comparable and reproducible.

## Intent and Purpose
Performance benchmarks are highly sensitive to the underlying system configuration. A minor change in kernel version, BIOS settings, or a software library can alter results, making comparisons invalid.

**ACR** solves this by creating a detailed "fingerprint" of the system at the time of a test run. This snapshot serves three main goals:

1. **Reproducibility**: Verify that benchmark environments are identical when comparing results over time or across different machines.
2. **Debugging**: Provide a comprehensive baseline to consult when performance results are unexpected.
3. **Documentation**: Archive the exact state of the System Under Test (SUT) alongside benchmark data for future reference.

## Key Features
-   **Single Static Binary**: Written in Go, the tool compiles to a single executable with no runtime dependencies, making distribution trivial.
-   **Lightweight**: Prioritizes reading from the kernel's /proc and /sys filesystems to minimize overhead and external dependencies.
-   **Flexible Output Formats**: Supports multiple machine-readable and human-readable formats for different use cases.
-   **Modular and Extensible**: A clean, interface-based design allows new "Collectors" and "Reporters" to be added with minimal effort.
-   **Resilient**: Designed to collect as much information as possible, even if certain commands or data sources are unavailable on a system.

## Information collected
The tool is organized into modular collectors that gather following information:

-   **CPU-info** : Model name, stepping, L1/L2/SLC cache, numa node, cores per socket, vulnerabilities, cpu-part.
-   **System-info**: Boot_params, bios info, chassis info, hostname, kernel, pagesize, OS, time, PCIE devices and more.
-   **Software-info**: docker containers, gcc, glibc, go, python, java, llvm, openssl, tuned-adm
-   **Memory-info**: Buffers, cached, Hugepages, memfree, Swap info
-   **DIMM-info**: detected memory channels, populated dimm info - bank locator, manufacturer, part#, speed, type
-   **Disk-info**: mapped partitions, device type, free_bytes, fstype, used_precentage, free_bytes
-   **Network-info**: NIC interfaces, ipv4/v6 addresses, speed, state

## Getting ACR
Pre-built ACR release is available in repository's [Release](https://github.com/AmpereComputing/ampere-config-reporter/releases/download/v1.0/acr). Download and start using ACR

### Usage
The tool requires root privileges for some collectors to access detailed hardware information (e.g., from `dmidecode`, `ethtool`). It's recommended to run it with `sudo`

```bash
# Get a report in the default JSON format(acr.json)
sudo ./acr

# Get a CSV report
sudo ./acr -format=csv

# Save a report as custom JSON file
sudo ./acr -format=json -o acr_out.json
```

## Installation/Dev environment setup

### Prerequisites

You need a working Go environment (version 1.18 or newer) to build the tool from the source.

### Building from Source

1.  Clone the repository

2.  Build the binary:
    ```bash
    go build -o acr .
    ```

### Cross-compiling Linux binaries

The repository includes a small `Makefile` that builds static Linux binaries with `CGO_ENABLED=0`.

```bash
# Build for the current host architecture (linux/amd64 on most x86_64 systems)
make build

# Build an Ampere target binary
make build-arm64

# Build both amd64 and arm64 outputs into ./dist
make build-all
```

### Dependencies

The tool is designed to be as lightweight as possible, but for comprehensive data collection, it relies on a few standard Linux command-line utilities. Please ensure they are installed on the target system.
-  dmidecode: detailed memory hardware and BIOS information
-  lspci: enumerating PCIe devices
-  ethtool: detailed network driver and firmware information

#### Installation on Debian/Ubuntu

```bash
    sudo apt-get update && sudo apt-get install dmidecode pciutils ethtool
```

#### Installation on RHEL/Centos/Fedora

```bash
    sudo dnf install dmidecode pciutils ethtool
```


