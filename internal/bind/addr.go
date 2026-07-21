package bind

import (
	"fmt"
	"net"
	"strings"
)

// Check validates the listen address.
// Default policy: loopback only. Non-loopback requires allowRemote.
func Check(addr string, allowRemote bool) error {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		if strings.HasPrefix(addr, ":") {
			return fmt.Errorf("refusing bind-all address %q — use 127.0.0.1:7474 (or pass --allow-remote with an explicit host if you understand the risk)", addr)
		}
		return fmt.Errorf("invalid listen address %q: %w", addr, err)
	}
	if host == "" {
		return fmt.Errorf("refusing bind-all address %q — use 127.0.0.1:7474", addr)
	}
	if host == "localhost" {
		return nil
	}
	ip := net.ParseIP(host)
	if ip != nil && ip.IsLoopback() {
		return nil
	}
	if !allowRemote {
		return fmt.Errorf("refusing non-loopback listen address %q — dash has no auth; pass --allow-remote only if you understand kube access equals this process", addr)
	}
	return nil
}

// IsLoopback reports whether addr is localhost / loopback.
func IsLoopback(addr string) bool {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return false
	}
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
