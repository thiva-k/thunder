// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	serverconst "github.com/thunder-id/thunderid/internal/system/constants"
	"github.com/thunder-id/thunderid/internal/system/csp"
	"github.com/thunder-id/thunderid/pkg/thunderidengine/common"
)

// fakeCSPReader serves a fixed merged csp policy to the resolver.
type fakeCSPReader struct {
	cfg csp.PolicyConfig
}

func (f fakeCSPReader) GetMergedConfig(_ context.Context, _ string) (any, *common.ServiceError) {
	return f.cfg, nil
}

// serve runs the middleware against a single GET request for path and returns the recorder.
func serve(path string) *httptest.ResponseRecorder {
	handler := SecurityHeadersMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, path, nil))
	return rr
}

func boolPtr(b bool) *bool { return &b }

func TestSecurityHeadersMiddleware_DefaultReportOnlyBaseline(t *testing.T) {
	csp.InitializeConfigReader(nil)
	t.Cleanup(func() { csp.InitializeConfigReader(nil) })

	rr := serve("/console/apps")

	assert.Equal(t, http.StatusOK, rr.Code)
	// With no configured section the default is report-only.
	policy := rr.Header().Get(serverconst.ContentSecurityPolicyReportOnlyHeaderName)
	assert.NotEmpty(t, policy)
	assert.Empty(t, rr.Header().Get(serverconst.ContentSecurityPolicyHeaderName))

	for _, directive := range []string{
		"default-src 'none'",
		"script-src 'self'",
		"style-src 'self'",
		"img-src 'self'",
		"connect-src 'self'",
		"frame-ancestors 'none'",
		"base-uri 'none'",
		"form-action 'self'",
	} {
		assert.Contains(t, policy, directive)
	}
	assert.NotContains(t, policy, "unsafe-inline")
	assert.NotContains(t, policy, "unsafe-eval")
	assert.NotContains(t, policy, "nonce-")

	// X-Frame-Options is always enforced.
	assert.Equal(t, serverconst.XFrameOptionsDeny, rr.Header().Get(serverconst.XFrameOptionsHeaderName))
}

func TestSecurityHeadersMiddleware_Enforcing(t *testing.T) {
	csp.InitializeConfigReader(fakeCSPReader{cfg: csp.PolicyConfig{ReportOnly: boolPtr(false)}})
	t.Cleanup(func() { csp.InitializeConfigReader(nil) })

	rr := serve("/console/apps")

	assert.NotEmpty(t, rr.Header().Get(serverconst.ContentSecurityPolicyHeaderName))
	assert.Empty(t, rr.Header().Get(serverconst.ContentSecurityPolicyReportOnlyHeaderName))
}

func TestSecurityHeadersMiddleware_OverrideAndReportURI(t *testing.T) {
	csp.InitializeConfigReader(fakeCSPReader{cfg: csp.PolicyConfig{
		ReportURI:  "/csp-report",
		Directives: map[string][]string{"img-src": {"'self'", "https://avatars.example.com"}},
	}})
	t.Cleanup(func() { csp.InitializeConfigReader(nil) })

	rr := serve("/console/apps")

	policy := rr.Header().Get(serverconst.ContentSecurityPolicyReportOnlyHeaderName)
	assert.Contains(t, policy, "img-src 'self' https://avatars.example.com")
	assert.Contains(t, policy, "report-uri /csp-report")
}

func TestSecurityHeadersMiddleware_PerPath(t *testing.T) {
	csp.InitializeConfigReader(fakeCSPReader{cfg: csp.PolicyConfig{
		ReportOnly: boolPtr(false),
		Paths: []csp.PathPolicy{
			{Location: "/gate/", Directives: map[string][]string{"style-src": {"'self'", "'unsafe-inline'"}}},
		},
	}})
	t.Cleanup(func() { csp.InitializeConfigReader(nil) })

	// The matching path gets its relaxation.
	gate := serve("/gate/signin").Header().Get(serverconst.ContentSecurityPolicyHeaderName)
	assert.Contains(t, gate, "style-src 'self' 'unsafe-inline'")

	// Console still gets a policy (the default, unrelaxed here), just not the gate-specific override.
	console := serve("/console/apps").Header().Get(serverconst.ContentSecurityPolicyHeaderName)
	assert.NotContains(t, console, "unsafe-inline")
	assert.Contains(t, console, "style-src 'self';")
}

func TestSecurityHeadersMiddleware_ScopedToConsoleAndGate(t *testing.T) {
	csp.InitializeConfigReader(fakeCSPReader{cfg: csp.PolicyConfig{ReportOnly: boolPtr(false)}})
	t.Cleanup(func() { csp.InitializeConfigReader(nil) })

	for _, path := range []string{"/console/", "/console/apps", "/gate/", "/gate/signin"} {
		rr := serve(path)
		assert.NotEmpty(t, rr.Header().Get(serverconst.ContentSecurityPolicyHeaderName), "path %q should get CSP", path)
	}

	// API, OAuth2, and other non-app responses get no CSP header at all: they are JSON or redirects,
	// never a rendered document, so CSP on them would be inert.
	for _, path := range []string{"/oauth2/token", "/oauth2/authorize", "/server-config/cors", "/health", "/"} {
		rr := serve(path)
		assert.Empty(t, rr.Header().Get(serverconst.ContentSecurityPolicyHeaderName),
			"path %q should not get CSP", path)
		assert.Empty(t, rr.Header().Get(serverconst.ContentSecurityPolicyReportOnlyHeaderName),
			"path %q should not get report-only CSP either", path)
		// X-Frame-Options is unaffected by the CSP scoping.
		assert.Equal(t, serverconst.XFrameOptionsDeny, rr.Header().Get(serverconst.XFrameOptionsHeaderName),
			"path %q should still get X-Frame-Options", path)
	}
}
