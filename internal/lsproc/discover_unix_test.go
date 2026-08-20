//go:build !windows

package lsproc

import (
	"net"
	"os"
	"testing"
)

func TestPortsViaProcWithLocalListener(t *testing.T) {
	// If /proc is not available (e.g. macOS), skip
	if _, err := os.Stat("/proc"); os.IsNotExist(err) {
		t.Skip("/proc is not available on this platform")
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to create test listener: %v", err)
	}
	defer ln.Close()

	expectedPort := ln.Addr().(*net.TCPAddr).Port

	ports, err := portsViaProc(os.Getpid())
	if err != nil {
		t.Fatalf("portsViaProc failed: %v", err)
	}

	found := false
	for _, p := range ports {
		if p == expectedPort {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("portsViaProc did not find expected port %d in %v", expectedPort, ports)
	}
}

func TestListeningPortsFallbackChain(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to create test listener: %v", err)
	}
	defer ln.Close()

	expectedPort := ln.Addr().(*net.TCPAddr).Port

	ports, err := listeningPorts(os.Getpid())
	if err != nil {
		t.Fatalf("listeningPorts failed: %v", err)
	}

	found := false
	for _, p := range ports {
		if p == expectedPort {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("listeningPorts did not find expected port %d in %v", expectedPort, ports)
	}
}
