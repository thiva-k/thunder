// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

package cli

import "testing"

func TestDefaultAdminPassword_UsesPresetEnv(t *testing.T) {
	t.Setenv("THUNDERID_ADMIN_PASSWORD", "Preset#1Pass")
	if got := defaultAdminPassword(); got != "Preset#1Pass" {
		t.Fatalf("expected pre-set password to be reused, got %q", got)
	}
}

func TestDefaultAdminPassword_GeneratesWhenUnset(t *testing.T) {
	t.Setenv("THUNDERID_ADMIN_PASSWORD", "")
	if got := defaultAdminPassword(); len(got) != 12 {
		t.Fatalf("expected a generated 12-char password, got %q (len %d)", got, len(got))
	}
}

func TestCollectAdminCredentials_SkipsWhenBothPreset(t *testing.T) {
	t.Setenv("THUNDERID_ADMIN_USERNAME", "operator")
	t.Setenv("THUNDERID_ADMIN_PASSWORD", "Preset#1Pass")
	if creds := collectAdminCredentials(); creds != nil {
		t.Fatalf("expected nil when both env vars are preset, got %+v", creds)
	}
}

func TestCollectAdminCredentials_NonInteractiveReturnsNil(t *testing.T) {
	// Under `go test` stdin is not a character device, so the interactive prompt
	// is skipped and the function falls through to setup's own defaults.
	t.Setenv("THUNDERID_ADMIN_USERNAME", "")
	t.Setenv("THUNDERID_ADMIN_PASSWORD", "")
	if creds := collectAdminCredentials(); creds != nil {
		t.Fatalf("expected nil on non-interactive stdin, got %+v", creds)
	}
}
