// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"net/http"
	"strings"

	serverconst "github.com/thunder-id/thunderid/internal/system/constants"
	"github.com/thunder-id/thunderid/internal/system/csp"
)

// cspAppPathPrefixes are the only paths CSP is emitted for: the Gate and Console apps, the only
// responses that render as documents a browser enforces CSP against. Everything else (API, OAuth2) is
// JSON or a redirect, where a CSP header would be inert.
var cspAppPathPrefixes = []string{"/gate/", "/console/"}

// SecurityHeadersMiddleware applies X-Frame-Options: DENY to every response and the deny-first
// Content-Security-Policy to Gate and Console responses. The effective CSP policy is resolved per
// request from the csp server-config section, falling back to the deny-first baseline in report-only
// mode when unset. X-Frame-Options stays global since, unlike CSP, it has no report-only mode.
func SecurityHeadersMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set(serverconst.XFrameOptionsHeaderName, serverconst.XFrameOptionsDeny)

			if isCSPAppPath(r.URL.Path) {
				policy := csp.Resolve(r.Context())
				w.Header().Set(policy.HeaderName(), policy.HeaderValue(r.URL.Path))
			}

			next.ServeHTTP(w, r)
		})
	}
}

// isCSPAppPath reports whether path belongs to the Gate or Console application.
func isCSPAppPath(path string) bool {
	for _, prefix := range cspAppPathPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}
