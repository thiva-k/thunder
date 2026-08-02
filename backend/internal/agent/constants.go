// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

package agent

const (
	agentBasePath = "/agents"
)

// Field keys stored in the entity SystemAttributes JSON blob.
const (
	fieldName         = "name"
	fieldDescription  = "description"
	fieldOwner        = "owner"
	fieldClientID     = "clientId"
	fieldClientSecret = "clientSecret"
)

// propLogoURL is the inbound-client PROPERTIES key holding the agent logo.
const propLogoURL = "logo_url"
