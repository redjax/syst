package pingCommand

import (
	"errors"
	"net"
	"strings"
)

func isDNSError(err error) bool {
	if err == nil {
		return false
	}
	// Look for common DNS error strings, customize as needed
	errStr := err.Error()
	if strings.Contains(errStr, "no such host") || strings.Contains(errStr, "no such IP address") {
		return true
	}

	var dnsErr *net.DNSError
	if ok := errors.As(err, &dnsErr); ok {
		return true
	}
	return false
}
