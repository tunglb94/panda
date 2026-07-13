package middleware

import "net/http"

// CORS allows browser-based clients (Admin Web — apps/admin, a Flutter Web
// app served from its own origin, e.g. http://localhost:8765 in dev) to call
// this API cross-origin. Every route here is either public (login) or
// Bearer-token gated (RequireAdmin/Auth) — never cookie-session gated — so a
// wildcard origin carries no CSRF risk; it does not send credentials
// (cookies), only an Authorization header the browser never attaches
// automatically to a forged request from another site.
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
