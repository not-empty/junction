package httpserver

import (
	"net/http"
	"strings"
)

type originPolicy int

const (
	originDenied originPolicy = iota
	originAllowed
	originAny
)

func NewCorsMiddleware(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			if origin != "" {
				switch isOriginAllowed(origin, allowedOrigins) {
				case originAny:
					w.Header().Set("Access-Control-Allow-Origin", "*")
				case originAllowed:
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Access-Control-Allow-Credentials", "true")
				}
			}

			w.Header().Add("Vary", "Origin")
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("Referrer-Policy", "no-referrer")
			w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
			w.Header().Set("Content-Security-Policy", "default-src 'none'")

			if r.Method == http.MethodOptions {
				accessMethods := "POST, GET, PUT, DELETE, PATCH"
				accessHeaders := "Content-Type, Accept, Authorization"

				w.Header().Set("Access-Control-Allow-Methods", accessMethods)
				w.Header().Set("Access-Control-Allow-Headers", accessHeaders)
				w.Header().Set("Access-Control-Max-Age", "86400")

				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func isOriginAllowed(origin string, allowedOrigins []string) originPolicy {
	if len(allowedOrigins) == 1 && allowedOrigins[0] == "*" {
		return originAny
	}

	for _, allowedOrigin := range allowedOrigins {
		if checkDomain(origin, allowedOrigin) {
			return originAllowed
		}
	}

	return originDenied
}

func checkDomain(origin string, allowedOrigin string) bool {
	scheme, wildcardHost, isWildcard := strings.Cut(allowedOrigin, "://*.")

	if isWildcard {
		return strings.HasPrefix(origin, scheme+"://") &&
			strings.HasSuffix(origin, "."+wildcardHost)
	}

	return origin == allowedOrigin
}
