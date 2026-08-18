//go:build windows

package lsproc

import (
	"os/exec"
	"strconv"
	"strings"
)

func powershell(script string) (string, error) {
	out, err := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", script).Output()
	return string(out), err
}

const listProcessesScript = `Get-CimInstance Win32_Process | Where-Object { $_.CommandLine -like '*language_server*' -and $_.CommandLine -like '*--standalone*' } | ForEach-Object { "$($_.ProcessId)|$($_.CommandLine)" }`

func standaloneProcesses() ([]process, error) {
	out, err := powershell(listProcessesScript)
	if err != nil {
		return nil, err
	}

	var procs []process
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 2)
		if len(parts) < 2 {
			continue
		}
		pid, err := strconv.Atoi(parts[0])
		if err != nil || !isStandaloneLS(parts[1]) {
			continue
		}
		procs = append(procs, process{pid: pid, cmdline: parts[1]})
	}
	return procs, nil
}

func listeningPorts(pid int) ([]int, error) {
	script := `Get-NetTCPConnection -State Listen -OwningProcess ` + strconv.Itoa(pid) +
		` -ErrorAction SilentlyContinue | ForEach-Object { $_.LocalPort }`

	out, err := powershell(script)
	if err != nil {
		return nil, err
	}

	var ports []int
	seen := map[int]bool{}
	for _, line := range strings.Split(out, "\n") {
		port, err := strconv.Atoi(strings.TrimSpace(line))
		if err != nil || seen[port] {
			continue
		}
		seen[port] = true
		ports = append(ports, port)
	}
	return ports, nil
}
