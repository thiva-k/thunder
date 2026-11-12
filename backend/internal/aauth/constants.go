// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

package aauth

import "time"

// HTTP header names used by AAuth and HTTP Signature Keys.
const (
	headerSignature                   = "Signature"
	headerSignatureInput              = "Signature-Input"
	headerSignatureKey                = "Signature-Key"
	headerSignatureError              = "Signature-Error"
	headerAcceptSignatureScheme       = "Accept-Signature-Scheme"
	headerAcceptSignatureAlg          = "Accept-Signature-Alg"
	headerContentDigest               = "Content-Digest"
	headerAAuthRequirement            = "AAuth-Requirement"
	headerPrefer                      = "Prefer"
	headerRetryAfter                  = "Retry-After"
	headerLocation                    = "Location"
	headerCacheControl                = "Cache-Control"
	contentTypeJSON                   = "application/json"
	contentTypeProblemJSON            = "application/problem+json"
	signatureKeySchemeJWT             = "jwt"
	maxRequestBodyBytes         int64 = 64 * 1024
)

// JWT types and well-known document names (dwk) of the AAuth tokens.
const (
	typAgentToken    = "aa-agent+jwt"    // #nosec G101 -- token type name, not a secret
	typPersonToken   = "aa-person+jwt"   // #nosec G101 -- token type name, not a secret
	typResourceToken = "aa-resource+jwt" // #nosec G101 -- token type name, not a secret
	typAuthToken     = "aa-auth+jwt"     // #nosec G101 -- token type name, not a secret
	dwkAgent         = "aauth-agent.json"
	dwkPerson        = "aauth-person.json"
	dwkResource      = "aauth-resource.json"
	dwkAccess        = "aauth-access.json"
)

// Endpoint paths served by the person server.
const (
	pathPersonMetadata = "/.well-known/" + dwkPerson
	pathJWKS           = "/aauth/jwks"
	pathPersonToken    = "/aauth/person" // #nosec G101 -- endpoint path, not a secret
	pathAuthToken      = "/aauth/token"  // #nosec G101 -- endpoint path, not a secret
	pathMission        = "/aauth/mission"
	pathPending        = "/aauth/pending"
	pathInteraction    = "/aauth/interaction"
)

// Fully-specified JWS algorithm identifiers.
const (
	algEd25519 = "Ed25519"
	algES256   = "ES256"
	algRS256   = "RS256"
)

// Lifetimes and limits.
const (
	signatureWindow       = 60 * time.Second
	personTokenLifetime   = time.Hour
	authTokenLifetime     = time.Hour
	pendingLifetime       = 10 * time.Minute
	defaultPreferWait     = 0
	maxPreferWait         = 45 * time.Second
	pollInterval          = 1
	interactionCodeLength = 8
	maxCodeAttempts       = 5
	metadataCacheLifetime = 5 * time.Minute
	httpFetchTimeout      = 10 * time.Second
	maxMetadataBytes      = 256 * 1024
)

// Signature-Error codes from HTTP Signature Keys and the AAuth profile.
const (
	sigErrInvalidSignature      = "invalid_signature"
	sigErrInvalidInput          = "invalid_input"
	sigErrInvalidKey            = "invalid_key"
	sigErrUnknownKey            = "unknown_key"
	sigErrInvalidJWT            = "invalid_jwt"
	sigErrExpiredJWT            = "expired_jwt"
	sigErrRevokedJWT            = "revoked_jwt"
	sigErrClockSkew             = "clock_skew"
	sigErrUnsupportedScheme     = "unsupported_scheme"
	sigErrUnsupportedAlgorithm  = "unsupported_algorithm"
	sigErrIssuerMissing         = "issuer_missing"
	sigErrIssuerMismatch        = "issuer_mismatch"
	sigErrSignatureKeyNoMatch   = "invalid_key"
	acceptedSignatureAlgorithms = "Ed25519, ES256"
)
