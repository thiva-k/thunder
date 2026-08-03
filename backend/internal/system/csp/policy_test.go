// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

package csp

import (
	"testing"

	"github.com/stretchr/testify/assert"

	serverconst "github.com/thunder-id/thunderid/internal/system/constants"
)

func boolPtr(b bool) *bool { return &b }

func TestPolicyConfig_EffectiveReportOnly(t *testing.T) {
	assert.True(t, PolicyConfig{}.EffectiveReportOnly(), "defaults to report-only when unset")
	assert.True(t, PolicyConfig{ReportOnly: boolPtr(true)}.EffectiveReportOnly())
	assert.False(t, PolicyConfig{ReportOnly: boolPtr(false)}.EffectiveReportOnly())
}

func TestPolicyConfig_HeaderName(t *testing.T) {
	assert.Equal(t, serverconst.ContentSecurityPolicyReportOnlyHeaderName, PolicyConfig{}.HeaderName())
	assert.Equal(t, serverconst.ContentSecurityPolicyHeaderName,
		PolicyConfig{ReportOnly: boolPtr(false)}.HeaderName())
}

func TestPolicyConfig_HeaderValue_Baseline(t *testing.T) {
	got := PolicyConfig{}.HeaderValue("/anything")
	assert.Equal(t,
		"default-src 'none'; script-src 'self'; style-src 'self'; img-src 'self'; connect-src 'self'; "+
			"frame-ancestors 'none'; base-uri 'none'; form-action 'self'",
		got)
	assert.NotContains(t, got, "unsafe-inline")
	assert.NotContains(t, got, "nonce-")
}

func TestPolicyConfig_HeaderValue_DefaultOverride(t *testing.T) {
	got := PolicyConfig{
		Directives: map[string][]string{
			"style-src": {"'self'", "'unsafe-inline'", "https://fonts.googleapis.com"},
			"font-src":  {"'self'", "https://fonts.gstatic.com"},
		},
	}.HeaderValue("/anything")

	// An overridden baseline directive is replaced, not appended to.
	assert.Contains(t, got, "style-src 'self' 'unsafe-inline' https://fonts.googleapis.com")
	// A configured directive not in the baseline is added, in stable order after the baseline.
	assert.Contains(t, got, "font-src 'self' https://fonts.gstatic.com")
	// Unlisted baseline directives keep their default.
	assert.Contains(t, got, "default-src 'none'")
	assert.Contains(t, got, "img-src 'self'")
}

func TestPolicyConfig_HeaderValue_PerPath(t *testing.T) {
	cfg := PolicyConfig{
		Directives: map[string][]string{"img-src": {"'self'", "data:"}}, // default for all paths
		Paths: []PathPolicy{
			{Location: "/console/", Directives: map[string][]string{"worker-src": {"'self'", "blob:"}}},
			{Location: "/gate/", Directives: map[string][]string{"style-src": {"'self'", "'unsafe-inline'"}}},
		},
	}

	console := cfg.HeaderValue("/console/apps")
	gate := cfg.HeaderValue("/gate/signin")
	api := cfg.HeaderValue("/oauth2/token")

	// Default directive applies everywhere.
	for _, h := range []string{console, gate, api} {
		assert.Contains(t, h, "img-src 'self' data:")
	}
	// Path-specific directives apply only to the matching path.
	assert.Contains(t, console, "worker-src 'self' blob:")
	assert.NotContains(t, console, "'unsafe-inline'")
	assert.Contains(t, gate, "style-src 'self' 'unsafe-inline'")
	assert.NotContains(t, gate, "worker-src")
	// A non-matching path (API) gets only the baseline plus default.
	assert.NotContains(t, api, "worker-src")
	assert.Contains(t, api, "style-src 'self';")
}

func TestPolicyConfig_HeaderValue_LongestPrefixWins(t *testing.T) {
	cfg := PolicyConfig{Paths: []PathPolicy{
		{Location: "/console/", Directives: map[string][]string{"img-src": {"'self'"}}},
		{Location: "/console/design/", Directives: map[string][]string{"img-src": {"'self'", "https:"}}},
	}}
	assert.Contains(t, cfg.HeaderValue("/console/design/themes"), "img-src 'self' https:")
	assert.Contains(t, cfg.HeaderValue("/console/apps"), "img-src 'self';")
}

func TestPolicyConfig_HeaderValue_ReportURI(t *testing.T) {
	got := PolicyConfig{ReportURI: "https://collector.example.com/csp"}.HeaderValue("/x")
	assert.Contains(t, got, "report-uri https://collector.example.com/csp")
}

func TestPolicyConfig_Validate(t *testing.T) {
	t.Run("empty is valid", func(t *testing.T) {
		assert.NoError(t, PolicyConfig{}.Validate())
	})

	t.Run("any directive may be configured, including former core directives", func(t *testing.T) {
		for _, name := range []string{"connect-src", "default-src", "frame-ancestors", "base-uri", "form-action"} {
			assert.NoError(t, PolicyConfig{Directives: map[string][]string{name: {"'self'"}}}.Validate(),
				"directive %q should be configurable", name)
		}
	})

	t.Run("invalid directive name is rejected", func(t *testing.T) {
		for _, name := range []string{"Bad-Name", "img src", "img;src", ""} {
			assert.Error(t, PolicyConfig{Directives: map[string][]string{name: {"'self'"}}}.Validate())
		}
	})

	t.Run("source with a separator or whitespace is rejected", func(t *testing.T) {
		for _, source := range []string{"https://a.com; script-src 'unsafe-inline'", "a,b", "a b"} {
			assert.Error(t, PolicyConfig{
				Directives: map[string][]string{"img-src": {source}},
			}.Validate(), "source %q should be rejected", source)
		}
	})

	t.Run("valid path policy is accepted", func(t *testing.T) {
		assert.NoError(t, PolicyConfig{Paths: []PathPolicy{
			{Location: "/console/", Directives: map[string][]string{"img-src": {"'self'", "data:"}}},
		}}.Validate())
	})

	t.Run("path prefix must be rooted", func(t *testing.T) {
		assert.ErrorContains(t, PolicyConfig{Paths: []PathPolicy{
			{Location: "console", Directives: map[string][]string{"img-src": {"'self'"}}},
		}}.Validate(), "must start with")
	})

	t.Run("path with no directives is rejected", func(t *testing.T) {
		assert.ErrorContains(t, PolicyConfig{Paths: []PathPolicy{
			{Location: "/console/"},
		}}.Validate(), "no directives")
	})

	t.Run("invalid source inside a path is rejected", func(t *testing.T) {
		assert.ErrorContains(t, PolicyConfig{Paths: []PathPolicy{
			{Location: "/console/", Directives: map[string][]string{"img-src": {"a b"}}},
		}}.Validate(), "/console/")
	})

	t.Run("non-http scheme report_uri is rejected", func(t *testing.T) {
		assert.ErrorContains(t, PolicyConfig{ReportURI: "ftp://collector.example.com"}.Validate(), "report_uri")
	})

	t.Run("report_uri containing a separator character is rejected", func(t *testing.T) {
		err := PolicyConfig{ReportURI: "https://example.com/csp;style-src https://evil.example"}.Validate()
		assert.ErrorContains(t, err, "report_uri")
	})
}
