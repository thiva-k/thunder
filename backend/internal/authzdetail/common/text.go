// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

package common

import "errors"

// IsSafeText reports whether text carries no control characters other than whitespace and no bidirectional
// formatting characters, so it can be shown to a user as it is.
func IsSafeText(text string) bool {
	return checkText(text) == nil
}

// checkText refuses control characters other than whitespace, and bidirectional formatting characters.
func checkText(text string) error {
	for _, r := range text {
		switch {
		case r == '\t' || r == '\n' || r == '\r':
		case r < 0x20 || (r >= 0x7f && r <= 0x9f):
			return errors.New("value carries a control character")
		case r == 0x061c || r == 0x200e || r == 0x200f || (r >= 0x202a && r <= 0x202e) ||
			(r >= 0x2066 && r <= 0x2069):
			return errors.New("value carries a bidirectional formatting character")
		}
	}
	return nil
}
