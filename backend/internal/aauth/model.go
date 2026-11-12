// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

package aauth

import (
	"encoding/json"
	"time"
)

// sigError is a signature failure, answered with 401 and a Signature-Error header.
type sigError struct {
	code          string
	detail        string
	requiredInput []string
}

// protocolError is an AAuth error answered with an application/problem+json body.
type protocolError struct {
	status int
	code   string
	detail string
	extra  map[string]any
}

// parsedJWT is a compact JWS split into its parts. Nothing in it has been verified.
type parsedJWT struct {
	header       map[string]any
	claims       map[string]any
	signingInput []byte
	signature    []byte
}

// agentIdentity is the verified agent behind a signed request.
type agentIdentity struct {
	issuer      string
	subject     string
	parentAgent string
	personSrv   string
	jwk         map[string]any
	jkt         string
	expiresAt   int64
	name        string
}

// key returns the (iss, sub) pair the person server recognizes the agent by.
func (a *agentIdentity) key() string {
	return a.issuer + "|" + a.subject
}

// personTokenRequest is the body of a person token request.
type personTokenRequest struct {
	Resource      string          `json:"resource"`
	MissionS256   string          `json:"mission_s256"`
	SubagentToken string          `json:"subagent_token"`
	UpstreamToken string          `json:"upstream_token"`
	Capabilities  []string        `json:"capabilities"`
	LoginHint     string          `json:"login_hint"`
	Tenant        string          `json:"tenant"`
	DomainHint    string          `json:"domain_hint"`
	Prompt        string          `json:"prompt"`
	Justification string          `json:"justification"`
	Platform      string          `json:"platform"`
	Device        string          `json:"device"`
	Extra         json.RawMessage `json:"-"`
}

// authTokenRequest is the body of an auth token request.
type authTokenRequest struct {
	ResourceToken  string   `json:"resource_token"`
	PresentedToken string   `json:"presented_token"`
	UpstreamToken  string   `json:"upstream_token"`
	SubagentToken  string   `json:"subagent_token"`
	Justification  string   `json:"justification"`
	LoginHint      string   `json:"login_hint"`
	Tenant         string   `json:"tenant"`
	DomainHint     string   `json:"domain_hint"`
	Prompt         string   `json:"prompt"`
	Platform       string   `json:"platform"`
	Device         string   `json:"device"`
	Capabilities   []string `json:"capabilities"`
}

// missionTool is a tool named in a mission proposal or blob.
type missionTool struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// missionProposal is the body of a mission proposal.
type missionProposal struct {
	Description  string        `json:"description"`
	Tools        []missionTool `json:"tools,omitempty"`
	Resources    []string      `json:"resources,omitempty"`
	Capabilities []string      `json:"capabilities,omitempty"`
}

// missionAction is the body of a request to a mission's own URL.
type missionAction struct {
	Action      string `json:"action"`
	Description string `json:"description"`
	Summary     string `json:"summary"`
}

// missionBlob is the approved mission. Its members are in the order they are serialized and hashed.
type missionBlob struct {
	Agent             string        `json:"agent"`
	ApprovedAt        string        `json:"approved_at"`
	ExpiresAt         string        `json:"expires_at,omitempty"`
	Description       string        `json:"description"`
	ApprovedTools     []missionTool `json:"approved_tools,omitempty"`
	ApprovedResources []string      `json:"approved_resources,omitempty"`
}

// missionRecord is a mission as the person server stores it.
type missionRecord struct {
	S256              string    `json:"s256"`
	Blob              []byte    `json:"blob"`
	AgentKey          string    `json:"agentKey"`
	UserID            string    `json:"userId"`
	Status            string    `json:"status"`
	TerminationReason string    `json:"terminationReason,omitempty"`
	ExpiresAt         time.Time `json:"expiresAt"`
	ApprovedResources []string  `json:"approvedResources,omitempty"`
	Capabilities      []string  `json:"capabilities,omitempty"`
	Updates           [][]byte  `json:"updates,omitempty"`
}

// resourceTokenClaims are the verified claims of a resource token that an auth token is issued for.
type resourceTokenClaims struct {
	Issuer       string `json:"iss"`
	JTI          string `json:"jti"`
	PersonServer string `json:"ps"`
	Subject      string `json:"sub"`
	PresentedJTI string `json:"presented_jti"`
	AgentJKT     string `json:"agent_jkt"`
	Scope        string `json:"scope,omitempty"`
	Account      string `json:"account,omitempty"`
	MissionS256  string `json:"mission_s256,omitempty"`
	Tenant       string `json:"tenant,omitempty"`
	ExpiresAt    int64  `json:"exp"`
}

// pendingRequest is a request waiting on the person, as stored in the runtime store. State is the
// field compare-and-swap transitions key on.
type pendingRequest struct {
	ID    string `json:"id"`
	Kind  string `json:"kind"`
	State string `json:"State"`

	Code string `json:"code"`

	AgentKey      string         `json:"agentKey"`
	AgentName     string         `json:"agentName,omitempty"`
	AgentIssuer   string         `json:"agentIssuer"`
	AgentSubject  string         `json:"agentSubject"`
	KeyJWK        map[string]any `json:"keyJwk"`
	TokenNotAfter int64          `json:"tokenNotAfter"`
	SubagentKey   string         `json:"subagentKey,omitempty"`
	Platform      string         `json:"platform,omitempty"`
	Device        string         `json:"device,omitempty"`
	Justification string         `json:"justification,omitempty"`
	Capabilities  []string       `json:"capabilities,omitempty"`

	Resource      string               `json:"resource,omitempty"`
	MissionS256   string               `json:"missionS256,omitempty"`
	Tenant        string               `json:"tenant,omitempty"`
	ResourceToken *resourceTokenClaims `json:"resourceToken,omitempty"`
	UserID        string               `json:"userId,omitempty"`
	Mission       *missionProposal     `json:"mission,omitempty"`
	Summary       string               `json:"summary,omitempty"`

	LoginState    string `json:"loginState,omitempty"`
	PKCEVerifier  string `json:"pkceVerifier,omitempty"`
	DecisionToken string `json:"decisionToken,omitempty"`

	Result      json.RawMessage `json:"result,omitempty"`
	ErrorStatus int             `json:"errorStatus,omitempty"`
	ErrorCode   string          `json:"errorCode,omitempty"`
	ErrorDetail string          `json:"errorDetail,omitempty"`

	ExpiresAt time.Time `json:"expiresAt"`
}

// consentRecord is what a person has approved for an agent at one resource.
type consentRecord struct {
	Identity bool     `json:"identity"`
	Scopes   []string `json:"scopes,omitempty"`
}

// issuedToken is the retention record for a person or auth token the person server issued.
type issuedToken struct {
	Type      string `json:"type"`
	Audience  string `json:"aud"`
	AgentKey  string `json:"agentKey"`
	UserID    string `json:"userId"`
	ExpiresAt int64  `json:"exp"`
}

// interactionView is what the consent page shows.
type interactionView struct {
	Kind          string
	Code          string
	AgentName     string
	AgentID       string
	AgentProvider string
	NewAgent      bool
	Platform      string
	Device        string
	Justification string
	Resource      string
	ResourceName  string
	ResourceDesc  string
	AccessMode    string
	Scopes        []string
	Mission       *missionProposal
	MissionDesc   string
	Summary       string
	PendingID     string
	DecisionToken string
	Message       string
	Error         string
}
