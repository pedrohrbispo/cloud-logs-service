package middleware

import (
	"net"
	"net/http"
)

// ClientIP returns the originating client IP. It trusts X-Real-IP, which the
// nginx reverse proxy OVERWRITES (proxy_set_header X-Real-IP $remote_addr) on
// every request, so a client cannot spoof it through the proxy. It falls back
// to the TCP peer when the header is absent (direct access).
func ClientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
