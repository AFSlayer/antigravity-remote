package auth

import (
	"net"
	"net/http"
	"strings"
)

func isTrusted(ip net.IP, trusted []*net.IPNet) bool {
	for _, n := range trusted {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}

func remoteIP(r *http.Request) net.IP {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	return net.ParseIP(strings.TrimSpace(host))
}

// ClientIP returns the address to rate-limit against. X-Forwarded-For is only
// consulted when the direct peer is a trusted proxy, and then the rightmost
// untrusted hop wins, so a client cannot spoof its way out of a lockout by
// prepending addresses.
func ClientIP(r *http.Request, trusted []*net.IPNet) string {
	direct := remoteIP(r)
	if direct == nil {
		return "unknown"
	}
	if len(trusted) == 0 || !isTrusted(direct, trusted) {
		return direct.String()
	}

	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded == "" {
		return direct.String()
	}

	parts := strings.Split(forwarded, ",")
	for i := len(parts) - 1; i >= 0; i-- {
		ip := net.ParseIP(strings.TrimSpace(parts[i]))
		if ip == nil {
			continue
		}
		if !isTrusted(ip, trusted) {
			return ip.String()
		}
	}
	return direct.String()
}

// RequestIsSecure reports whether the browser's connection was HTTPS, which
// decides if the session cookie may be marked Secure. X-Forwarded-Proto is
// believed only from a trusted proxy.
func RequestIsSecure(r *http.Request, trusted []*net.IPNet) bool {
	if r.TLS != nil {
		return true
	}

	direct := remoteIP(r)
	if direct == nil || !isTrusted(direct, trusted) {
		return false
	}
	return strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")
}
