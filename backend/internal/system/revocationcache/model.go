// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

package revocationcache

import "time"

// revokedEntry is one non-expired deny-list record returned by a syncSource and held in the cache.
type revokedEntry struct {
	// Value is the cache lookup key: the jti for a single-token entry, the tfid for a family entry.
	Value string
	// ExpiryTime is the revoked token's (or family's) original expiry; the entry is prunable once it
	// passes.
	ExpiryTime time.Time
}

// revokedSnapshot is one source read: the revoked single-token jtis and the revoked token-family ids
// for a deployment, held in separate cache dimensions so a jti is never matched against a tfid.
type revokedSnapshot struct {
	// Tokens holds the revoked single-token entries (keyed by jti).
	Tokens []revokedEntry
	// Families holds the revoked token-family entries (keyed by tfid).
	Families []revokedEntry
}
