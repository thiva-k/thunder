// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

package actorprovider

// clientIdentifierURLPrefix starts the Client Identifier URL the login page names a client admitted by Client ID
// domain with.
const clientIdentifierURLPrefix = "https://"

// documentPropertyKeys are the inbound client property keys of a client's logo and links, which a client admitted
// by Client ID domain takes from its own document.
var documentPropertyKeys = []string{"logo_url", "url", "tos_uri", "policy_uri"}
