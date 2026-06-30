package middleware

import "net/http"

// SecurityHeaders sets conservative security headers on every response.
// default-src 'none' is correct for JSON-API responses; the SPA document gets
// its own, looser CSP from nginx.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.Set("Cache-Control", "private, no-store")
		next.ServeHTTP(w, r)
	})
}
