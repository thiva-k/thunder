// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

package setup

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseAdminCredentials_ParsesBlock(t *testing.T) {
	output := "some noise\n" +
		"Admin credentials:\n" +
		"  Username: admin\n" +
		"  Password: abc123\n" +
		"  Sign in to the Console with these credentials.\n" +
		"\n" +
		"trailing noise\n"

	creds := parseAdminCredentials(output)

	assert.NotNil(t, creds)
	assert.Equal(t, "admin", creds.Username)
	assert.Equal(t, "abc123", creds.Password)
}

func TestParseAdminCredentials_CRLF(t *testing.T) {
	output := "some noise\r\n" +
		"Admin credentials:\r\n" +
		"  Username: admin\r\n" +
		"  Password: abc123\r\n" +
		"  Sign in to the Console with these credentials.\r\n" +
		"\r\n" +
		"trailing noise\r\n"

	creds := parseAdminCredentials(output)

	assert.NotNil(t, creds)
	assert.Equal(t, "admin", creds.Username)
	assert.Equal(t, "abc123", creds.Password)
}

func TestParseAdminCredentials_NoBlockReturnsNil(t *testing.T) {
	assert.Nil(t, parseAdminCredentials("no credentials here at all"))
}

func TestGenerateAdminPassword(t *testing.T) {
	const special = "@#%+=_.?-"
	for i := 0; i < 100; i++ {
		pw := GenerateAdminPassword()
		assert.Len(t, pw, 12)
		assert.True(t, strings.ContainsAny(pw, "0123456789"), "must contain a digit: %q", pw)
		assert.True(t, strings.ContainsAny(pw, special), "must contain a special char: %q", pw)
	}
	assert.NotEqual(t, GenerateAdminPassword(), GenerateAdminPassword())
}
