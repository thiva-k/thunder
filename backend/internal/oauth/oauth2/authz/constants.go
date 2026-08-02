// Copyright 2025 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

package authz

// Authorization code states.
const (
	AuthCodeStateActive   = "ACTIVE"
	AuthCodeStateInactive = "INACTIVE"
	AuthCodeStateExpired  = "EXPIRED"
	AuthCodeStateRevoked  = "REVOKED"
)
