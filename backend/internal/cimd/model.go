// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

package cimd

import (
	"encoding/json"

	"github.com/thunder-id/thunderid/pkg/thunderidengine/providers"
)

// cimdDocument holds the Client ID Metadata Document fields ThunderID reads.
type cimdDocument struct {
	ClientID                string          `json:"client_id"`
	ClientName              string          `json:"client_name"`
	ClientURI               string          `json:"client_uri"`
	LogoURI                 string          `json:"logo_uri"`
	TosURI                  string          `json:"tos_uri"`
	PolicyURI               string          `json:"policy_uri"`
	Contacts                []string        `json:"contacts"`
	RedirectURIs            []string        `json:"redirect_uris"`
	PostLogoutRedirectURIs  []string        `json:"post_logout_redirect_uris"`
	GrantTypes              []string        `json:"grant_types"`
	ResponseTypes           []string        `json:"response_types"`
	Scope                   string          `json:"scope"`
	ApplicationType         string          `json:"application_type"`
	TokenEndpointAuthMethod string          `json:"token_endpoint_auth_method"`
	JWKSURI                 string          `json:"jwks_uri"`
	JWKS                    json.RawMessage `json:"jwks"`
	ClientSecret            string          `json:"client_secret"`
	// resolvedJWKS is the key set retrieved from JWKSURI when the document is retrieved for a client a Client ID
	// domain admits, so the client authenticates with keys retrieved under the document retrieval's protections.
	resolvedJWKS string
}

// PreviewRequest is the request body of a Client ID Metadata Document preview.
type PreviewRequest struct {
	ClientID string `json:"clientId"`
}

// PreviewResponse holds the values taken from a Client ID Metadata Document, in the shape of the
// fields a create request takes.
type PreviewResponse struct {
	Name              string                                  `json:"name"`
	URL               string                                  `json:"url,omitempty"`
	TosURI            string                                  `json:"tosUri,omitempty"`
	PolicyURI         string                                  `json:"policyUri,omitempty"`
	Contacts          []string                                `json:"contacts,omitempty"`
	InboundAuthConfig []providers.InboundAuthConfigWithSecret `json:"inboundAuthConfig"`
}

// unregisteredClientSettings is the value of the cimd server configuration section: whether a client that
// publishes a Client ID Metadata Document may sign in through the Client ID domains of an application or agent,
// and the Client Identifier URLs refused even then.
type unregisteredClientSettings struct {
	Enabled          bool     `json:"enabled"`
	BlockedClientIDs []string `json:"blockedClientIds"`
}

// DocumentClient is a client that a Client ID domain admits, as its current Client ID Metadata Document
// describes it.
type DocumentClient struct {
	// OAuthConfig is the client's OAuth configuration, checked against the document rules.
	OAuthConfig *providers.OAuthConfigWithSecret
	// DeclaredGrantTypes are the grants the document declares, authorization_code when it declares none.
	DeclaredGrantTypes []string
	// DeclaredScopes are the scopes the document's scope declares the client may request, none when it declares
	// none.
	DeclaredScopes []string
	// PostLogoutRedirectURIs are the document's post-logout redirect URIs that pass the redirect URI rules.
	PostLogoutRedirectURIs []string
	// Name is the document's client_name, or the Client Identifier URL's host when it gives none.
	Name string
	// Properties holds the document's logo and links under the inbound client property keys the application
	// metadata reads.
	Properties map[string]interface{}
}
