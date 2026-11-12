// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

package model

// ApplicationArtifactProfile describes the artifacts an application issues: the OAuth client they
// carry, and the longest any of them can live. It is a fact about the application rather than a plan
// for one operation on it. A revocation is one consumer, which sizes a deny-list row from the lifetime.
type ApplicationArtifactProfile struct {
	// ClientKey is the application's OAuth client id, the value already-issued artifacts carry. It is
	// empty for an application with no OAuth component, which issues no artifacts.
	ClientKey string
	// MaxLifetimeSeconds is the longest an artifact issued to this client can remain valid. Zero means
	// the application states no opinion, leaving a consumer on its own configured default.
	MaxLifetimeSeconds int64
}

// CredentialAction names what is done to an application's credential. Regeneration
// is the only action today; rotation with an overlap window and additional secrets are the shapes this
// is expected to grow, and they arrive as further actions rather than further methods.
type CredentialAction string

const (
	// CredentialActionRegenerate replaces the application's credential with a newly generated value,
	// leaving no window in which the previous one still authenticates.
	CredentialActionRegenerate CredentialAction = "regenerate"
)

// ClientSecretResponse carries a newly generated client secret. It is returned once, when the secret is
// generated, and no read returns it again.
type ClientSecretResponse struct {
	ClientSecret string `json:"clientSecret"`
	// RevocationIncomplete is set when the secret is stored but revoking the tokens issued while it was
	// being written failed. Regenerating again revokes them.
	RevocationIncomplete bool `json:"revocationIncomplete,omitempty"`
}

// RegenerateClientSecretOutput is what the regenerate_application_secret tool returns.
type RegenerateClientSecretOutput struct {
	ClientSecret string `json:"clientSecret" jsonschema:"The new client secret"`
	// RevocationIncomplete is set when the tokens issued while the secret was replaced may not be revoked;
	// regenerating again revokes them.
	RevocationIncomplete bool   `json:"revocationIncomplete,omitempty" jsonschema:"Revoking recent tokens failed"`
	Notice               string `json:"notice" jsonschema:"How to handle the new client secret"`
}
