// Package netinfo discovers the addresses a phone can use to reach this host.
package netinfo

import (
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

// Info holds the addresses found for this host.
type Info struct {
	LAN       string
	Tailscale string
	Public    string
}

// Primary is the address to advertise, preferring Tailscale because it works
// from anywhere and is already encrypted.
func (i Info) Primary() string {
	if i.Tailscale != "" {
		return i.Tailscale
	}
	if i.LAN != "" {
		return i.LAN
	}
	return "127.0.0.1"
}

// Local inspects the network interfaces for usable IPv4 addresses.
func Local() Info {
	info := Info{}

	ifaces, err := net.Interfaces()
	if err != nil {
		return info
	}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			ipnet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}
			ip := ipnet.IP.To4()
			if ip == nil {
				continue
			}

			if isTailscale(ip) {
				if info.Tailscale == "" {
					info.Tailscale = ip.String()
				}
				continue
			}
			if info.LAN == "" {
				info.LAN = ip.String()
			}
		}
	}

	return info
}

// isTailscale matches the 100.64.0.0/10 range Tailscale assigns.
func isTailscale(ip net.IP) bool {
	return ip[0] == 100 && ip[1] >= 64 && ip[1] <= 127
}

// PublicIP asks an external service for this host's public address. It returns
// "" on any failure, since the address is only a convenience.
func PublicIP(timeout time.Duration) string {
	client := &http.Client{Timeout: timeout}

	resp, err := client.Get("https://api.ipify.org")
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 64))
	if err != nil {
		return ""
	}

	ip := strings.TrimSpace(string(body))
	if net.ParseIP(ip) == nil {
		return ""
	}
	return ip
}
