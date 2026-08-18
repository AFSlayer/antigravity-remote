//go:build !windows

package lsproc

import (
	"errors"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

func standaloneProcesses() ([]process, error) {
	out, err := exec.Command("ps", "ax", "-o", "pid=,command=").Output()
	if err != nil {
		return nil, err
	}

	var procs []process
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || !isStandaloneLS(line) {
			continue
		}

		fields := strings.SplitN(line, " ", 2)
		if len(fields) < 2 {
			continue
		}
		pid, err := strconv.Atoi(fields[0])
		if err != nil {
			continue
		}
		procs = append(procs, process{pid: pid, cmdline: fields[1]})
	}
	return procs, nil
}

func listeningPorts(pid int) ([]int, error) {
	if ports, err := portsViaLsof(pid); err == nil && len(ports) > 0 {
		return ports, nil
	}
	if ports, err := portsViaSS(pid); err == nil && len(ports) > 0 {
		return ports, nil
	}
	return nil, errors.New("could not determine listening ports")
}

var lsofPortRe = regexp.MustCompile(`:(\d+)\s+\(LISTEN\)`)

func portsViaLsof(pid int) ([]int, error) {
	out, err := exec.Command("lsof", "-a", "-p", strconv.Itoa(pid), "-iTCP", "-sTCP:LISTEN", "-P", "-n").Output()
	if err != nil {
		return nil, err
	}
	return collectPorts(string(out), lsofPortRe, ""), nil
}

var ssPortRe = regexp.MustCompile(`[:\]](\d+)\s`)

func portsViaSS(pid int) ([]int, error) {
	out, err := exec.Command("ss", "-tlnpH").Output()
	if err != nil {
		return nil, err
	}
	return collectPorts(string(out), ssPortRe, "pid="+strconv.Itoa(pid)+","), nil
}

func collectPorts(out string, re *regexp.Regexp, mustContain string) []int {
	var ports []int
	seen := map[int]bool{}

	for _, line := range strings.Split(out, "\n") {
		if mustContain != "" && !strings.Contains(line, mustContain) {
			continue
		}
		for _, m := range re.FindAllStringSubmatch(line, -1) {
			port, err := strconv.Atoi(m[1])
			if err != nil || port <= 0 || port > 65535 || seen[port] {
				continue
			}
			seen[port] = true
			ports = append(ports, port)
		}
	}
	return ports
}
