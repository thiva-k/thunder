// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

package cimd

import (
	"time"

	"github.com/thunder-id/thunderid/pkg/thunderidengine/providers"
)

// documentGrantTypes are the grants a CIMD client takes from its document. Any other grant the document
// rules permit is added by an administrator.
var documentGrantTypes = []providers.GrantType{
	providers.GrantTypeAuthorizationCode,
	providers.GrantTypeRefreshToken,
}

const (
	// maxClientIdentifierURLLength bounds a Client Identifier URL.
	maxClientIdentifierURLLength = 2048
	// cimdFetchTimeout bounds the retrieval of a Client ID Metadata Document.
	cimdFetchTimeout = 5 * time.Second
	// cimdMaxDocumentSize is the largest document processed, as the CIMD draft recommends.
	cimdMaxDocumentSize = 5 * 1024
	// cimdMaxJWKSSize is the largest key set retrieved from a document's jwks_uri.
	cimdMaxJWKSSize = 64 * 1024
	// maxApplicationNameLength is the longest name an application create request accepts.
	maxApplicationNameLength = 100
	// minCertificateValueLength and maxCertificateValueLength bound an inline JWKS or a JWKS URI, as the
	// certificate store does.
	minCertificateValueLength = 10
	maxCertificateValueLength = 4096

	schemeHTTPS = "https"
	schemeHTTP  = "http"

	// applicationTypeWeb is the application_type of a client that runs on a web server, which never takes a
	// loopback redirect URI.
	applicationTypeWeb = "web"

	// propertyLogoURL, propertyURL, propertyTosURI and propertyPolicyURI are the inbound client property keys
	// the application metadata reads a client's logo and links from.
	propertyLogoURL   = "logo_url"
	propertyURL       = "url"
	propertyTosURI    = "tos_uri"
	propertyPolicyURI = "policy_uri"

	// maxBlockedClientIDs bounds the blocked Client Identifier URLs of the cimd server configuration section.
	maxBlockedClientIDs = 500
	// maxClientIDDomains bounds the Client ID domains of one application or agent.
	maxClientIDDomains = 50
	// maxDomainLength and maxDomainLabelLength bound a Client ID domain and each of its labels.
	maxDomainLength      = 253
	maxDomainLabelLength = 63
	// wildcardDomainPrefix starts a Client ID domain that matches every subdomain of the rest.
	wildcardDomainPrefix = "*."

	// documentDefaultTTL is how long a document stays cached when its response sets no max-age.
	// documentMinTTL and documentMaxTTL bound a max-age the response sets.
	documentDefaultTTL = 5 * time.Minute
	documentMinTTL     = time.Minute
	documentMaxTTL     = time.Hour
	// documentFailureTTL is how long a failed retrieval is remembered, so a client cannot make every
	// request retrieve a document that is unavailable.
	documentFailureTTL = 30 * time.Second
	// maxCachedDocuments bounds the documents kept in memory.
	maxCachedDocuments = 1000
)
