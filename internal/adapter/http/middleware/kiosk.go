package middleware

import "net/http"

// KioskTokenHeader carries the token from a kiosk link (K-016).
const KioskTokenHeader = "X-Kiosk-Token"

// RequireKioskToken rejects requests without a token that valid accepts.
func RequireKioskToken(valid func(r *http.Request, token string) bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !valid(r, r.Header.Get(KioskTokenHeader)) {
				http.Error(w, `{"error":"valid kiosk link required"}`, http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
