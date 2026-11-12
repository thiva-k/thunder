// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

// Package aauth implements an AAuth person server (draft-hardt-oauth-aauth-protocol): it issues
// person tokens and auth tokens to agents that sign their requests with HTTP Message Signatures,
// approves missions, and asks the person for consent through an interaction page.
package aauth

import (
	"context"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	oauth2authz "github.com/thunder-id/thunderid/internal/oauth/oauth2/authz"
	"github.com/thunder-id/thunderid/internal/system/log"
	"github.com/thunder-id/thunderid/pkg/thunderidengine/providers"
)

// authorizationCodeRedeemer redeems an authorization code issued to the interaction client.
type authorizationCodeRedeemer interface {
	GetAuthorizationCodeDetails(ctx context.Context, clientID string, code string) (
		*oauth2authz.AuthorizationCode, error)
}

// Pending request kinds and states.
const (
	kindPersonToken       = "person_token"
	kindAuthToken         = "auth_token"
	kindMission           = "mission"
	kindMissionCompletion = "mission_completion"

	statePending       = "pending"
	stateInteracting   = "interacting"
	stateAuthenticated = "authenticated"
	stateApproved      = "approved"
	stateDenied        = "denied"

	missionActive     = "active"
	missionTerminated = "terminated"
)

// identityScopeClaims maps identity scopes to the user attributes released as auth token claims.
var identityScopeClaims = map[string][]string{
	"email":              {"email", "email_verified"},
	"profile":            {"name", "given_name", "family_name", "preferred_username", "nickname", "picture"},
	"name":               {"name"},
	"given_name":         {"given_name"},
	"family_name":        {"family_name"},
	"preferred_username": {"preferred_username"},
	"nickname":           {"nickname"},
	"picture":            {"picture"},
	"phone":              {"phone_number", "phone_number_verified"},
}

// crockfordAlphabet is the Crockford base32 symbol set interaction codes are drawn from.
const crockfordAlphabet = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"

// tokenOutcome is the result of a token or mission request: a 200 body, or a pending request.
type tokenOutcome struct {
	result  any
	pending *pendingRequest
}

// aauthService is the person server.
type aauthService struct {
	issuer              string
	signingKeyID        string
	interactionClientID string
	alg                 string
	cryptoAlg           string
	kid                 string
	publicJWK           map[string]any
	crypto              providers.RuntimeCryptoProvider
	store               *aauthStore
	fetcher             *issuerKeyFetcher
	codeRedeemer        authorizationCodeRedeemer
	actorProvider       providers.ActorProvider
	now                 func() time.Time
	logger              *log.Logger
}

func newAAuthService(ctx context.Context, issuer, signingKeyID, interactionClientID string,
	crypto providers.RuntimeCryptoProvider, store *aauthStore, fetcher *issuerKeyFetcher,
	codeRedeemer authorizationCodeRedeemer, actorProvider providers.ActorProvider) (*aauthService, error) {
	s := &aauthService{
		issuer:              issuer,
		signingKeyID:        signingKeyID,
		interactionClientID: interactionClientID,
		crypto:              crypto,
		store:               store,
		fetcher:             fetcher,
		codeRedeemer:        codeRedeemer,
		actorProvider:       actorProvider,
		now:                 time.Now,
		logger:              log.GetLogger().With(log.String(log.LoggerKeyComponentName, "AAuthService")),
	}
	keys, err := crypto.GetPublicKeys(ctx, providers.PublicKeyFilter{KeyID: signingKeyID})
	if err != nil || len(keys) == 0 {
		return nil, fmt.Errorf("aauth signing key %q is not configured", signingKeyID)
	}
	key := keys[0]
	jwk, alg, err := publicJWKFor(key.PublicKey)
	if err != nil {
		return nil, err
	}
	s.alg = alg
	s.cryptoAlg, _ = cryptoAlgorithm(alg)
	s.kid = key.Kid
	jwk["kid"] = key.Kid
	jwk["alg"] = alg
	jwk["use"] = "sig"
	s.publicJWK = jwk
	return s, nil
}

// publicJWKFor builds the public JWK and the fully-specified algorithm for a signing key.
func publicJWKFor(key any) (map[string]any, string, error) {
	switch k := key.(type) {
	case ed25519.PublicKey:
		return map[string]any{"kty": "OKP", "crv": "Ed25519", "x": encodeB64URL(k)}, algEd25519, nil
	case *ecdsa.PublicKey:
		if k.Curve != elliptic.P256() {
			return nil, "", errors.New("aauth signing key must be Ed25519 or P-256")
		}
		point, err := k.Bytes()
		if err != nil {
			return nil, "", err
		}
		// An uncompressed point is 0x04 || X || Y.
		return map[string]any{"kty": "EC", "crv": "P-256", "x": encodeB64URL(point[1:33]),
			"y": encodeB64URL(point[33:])}, algES256, nil
	case *rsa.PublicKey:
		return map[string]any{"kty": "RSA", "n": encodeB64URL(k.N.Bytes()),
			"e": encodeB64URL(big.NewInt(int64(k.E)).Bytes())}, algRS256, nil
	default:
		return nil, "", errors.New("unsupported aauth signing key type")
	}
}

// metadata returns the person server metadata document.
func (s *aauthService) metadata() map[string]any {
	return map[string]any{
		"issuer":                s.issuer,
		"jwks_uri":              s.issuer + pathJWKS,
		"name":                  "ThunderID",
		"description":           "**ThunderID** person server: manage which agents act for you.",
		"person_token_endpoint": s.issuer + pathPersonToken,
		"auth_token_endpoint":   s.issuer + pathAuthToken,
		"mission_endpoint":      s.issuer + pathMission,
		"accept_signature_algs": []string{algEd25519, algES256, "ES384", "ES512", algRS256, "RS512", "PS256"},
		"scopes_supported":      []string{"openid", "profile", "email", "name", "phone"},
		"claims_supported":      []string{"sub", "email", "email_verified", "name", "given_name", "family_name"},
	}
}

// jwks returns the person server's JSON Web Key Set.
func (s *aauthService) jwks() map[string]any {
	return map[string]any{"keys": []map[string]any{s.publicJWK}}
}

// authenticate verifies the request's HTTP Message Signature and the agent token that carries its
// key. requireBody requests content-digest and content-type coverage on a body.
func (s *aauthService) authenticate(r *http.Request) (*agentIdentity, []byte, *sigError) {
	signed, sigErr := parseSignedRequest(r, s.now())
	if sigErr != nil {
		return nil, nil, sigErr
	}
	if signed.keyMember.item == nil {
		return nil, nil, &sigError{code: sigErrInvalidKey, detail: "Signature-Key member is not an item"}
	}
	if scheme, ok := signed.keyMember.item.value.(sfToken); !ok || scheme != signatureKeySchemeJWT {
		return nil, nil, &sigError{code: sigErrUnsupportedScheme, detail: "the jwt scheme is required"}
	}
	token, _ := param(signed.keyMember.item.params, "jwt")
	jwt, ok := token.(string)
	if !ok || jwt == "" {
		return nil, nil, &sigError{code: sigErrInvalidKey, detail: "the jwt parameter is required"}
	}
	agent, sigErr := s.verifyAgentToken(r.Context(), jwt)
	if sigErr != nil {
		return nil, nil, sigErr
	}
	if err := verifyWithJWK(r.Context(), s.crypto, agent.jwk, "", signed.base, signed.signature); err != nil {
		if errors.Is(err, errUnsupportedAlg) {
			return nil, nil, &sigError{code: sigErrUnsupportedAlgorithm, detail: "unsupported key algorithm"}
		}
		return nil, nil, &sigError{code: sigErrInvalidSignature, detail: "the HTTP message signature does not verify"}
	}
	return agent, signed.body, nil
}

// verifyAgentToken verifies an aa-agent+jwt against its agent provider's keys.
func (s *aauthService) verifyAgentToken(ctx context.Context, token string) (*agentIdentity, *sigError) {
	p, err := parseJWT(token)
	if err != nil {
		return nil, &sigError{code: sigErrInvalidJWT, detail: "malformed agent token"}
	}
	if p.headerString("typ") != typAgentToken {
		return nil, &sigError{code: sigErrInvalidJWT, detail: "an agent token is required"}
	}
	exp := p.exp()
	if exp == 0 {
		return nil, &sigError{code: sigErrInvalidJWT, detail: "the agent token has no exp"}
	}
	if exp <= s.now().Unix() {
		return nil, &sigError{code: sigErrExpiredJWT, detail: "the agent token has expired"}
	}
	if p.claimString("dwk") != dwkAgent {
		return nil, &sigError{code: sigErrInvalidJWT, detail: "dwk must be " + dwkAgent}
	}
	subject := p.claimString("sub")
	if !isAgentIdentifier(subject) {
		return nil, &sigError{code: sigErrInvalidJWT, detail: "sub is not an agent identifier"}
	}
	jwk := p.cnfJWK()
	if jwk == nil {
		return nil, &sigError{code: sigErrInvalidJWT, detail: "cnf.jwk is required"}
	}
	if _, err := checkJWKForAlg(jwk); err != nil {
		if errors.Is(err, errUnsupportedAlg) {
			return nil, &sigError{code: sigErrUnsupportedAlgorithm, detail: "cnf.jwk needs a supported alg"}
		}
		return nil, &sigError{code: sigErrInvalidKey, detail: "cnf.jwk does not match its alg"}
	}
	if ps := p.claimString("ps"); ps != "" && !isServerIdentifier(ps) {
		return nil, &sigError{code: sigErrInvalidJWT, detail: "ps is not a server identifier"}
	}
	if parent := p.claimString("parent_agent"); parent != "" && !isAgentIdentifier(parent) {
		return nil, &sigError{code: sigErrInvalidJWT, detail: "parent_agent is not an agent identifier"}
	}
	meta, err := verifyIssuerJWT(ctx, s.crypto, s.fetcher, p, dwkAgent)
	if err != nil {
		s.logger.Debug(ctx, "Agent token verification failed", log.Error(err))
		return nil, sigErrorFor(err)
	}
	issuer := p.claimString("iss")
	jkt, err := jktOf(jwk)
	if err != nil {
		return nil, &sigError{code: sigErrInvalidKey, detail: "cnf.jwk cannot be thumbprinted"}
	}
	name := meta.Name
	if name == "" {
		name = meta.ClientName
	}
	return &agentIdentity{
		issuer:      issuer,
		subject:     subject,
		parentAgent: p.claimString("parent_agent"),
		personSrv:   p.claimString("ps"),
		jwk:         jwk,
		jkt:         jkt,
		expiresAt:   exp,
		name:        name,
	}, nil
}

// sigErrorFor maps a key discovery or verification error to a Signature-Error code.
func sigErrorFor(err error) *sigError {
	switch {
	case errors.Is(err, errIssuerMissing):
		return &sigError{code: sigErrIssuerMissing, detail: "issuer metadata has no issuer"}
	case errors.Is(err, errIssuerMismatch):
		return &sigError{code: sigErrIssuerMismatch, detail: "issuer metadata names another issuer"}
	case errors.Is(err, errUnknownKey):
		return &sigError{code: sigErrUnknownKey, detail: "the token's kid is not in the issuer's JWKS"}
	case errors.Is(err, errUnsupportedAlg):
		return &sigError{code: sigErrUnsupportedAlgorithm, detail: "unsupported token algorithm"}
	default:
		return &sigError{code: sigErrInvalidJWT, detail: "the token does not verify"}
	}
}

// requestPersonToken handles a person token request.
func (s *aauthService) requestPersonToken(ctx context.Context, agent *agentIdentity,
	req *personTokenRequest) (*tokenOutcome, *protocolError) {
	if agent.parentAgent != "" {
		return nil, badRequest(errInvalidRequest, "a sub-agent's parent requests its person tokens")
	}
	if !isServerIdentifier(req.Resource) {
		return nil, badRequest(errInvalidRequest, "resource must be a server identifier")
	}
	if req.UpstreamToken != "" {
		return nil, badRequest(errInvalidUpstreamToken, "call chaining is not supported by this person server")
	}
	keyJWK, notAfter, subagentKey := agent.jwk, agent.expiresAt, ""
	if req.SubagentToken != "" {
		sub, perr := s.verifySubagentToken(ctx, agent, req.SubagentToken)
		if perr != nil {
			return nil, perr
		}
		keyJWK, notAfter, subagentKey = sub.jwk, min(notAfter, sub.expiresAt), sub.key()
	}

	userID, err := s.store.getBinding(ctx, agent.key())
	if err != nil {
		return nil, serverError()
	}
	var mission *missionRecord
	if req.MissionS256 != "" {
		var perr *protocolError
		mission, perr = s.activeMission(ctx, req.MissionS256, agent.key())
		if perr != nil {
			return nil, perr
		}
		notAfter = min(notAfter, mission.ExpiresAt.Unix())
	}

	approved := false
	if userID != "" {
		if mission != nil && slices.Contains(mission.ApprovedResources, req.Resource) {
			approved = true
		} else if c, err := s.store.getConsent(ctx, userID, agent.key(), req.Resource); err != nil {
			return nil, serverError()
		} else if c != nil && c.Identity {
			approved = true
		}
	}
	if approved {
		token, expiresIn, perr := s.mintPersonToken(ctx, userID, agent.key(), req.Resource, keyJWK, notAfter,
			req.MissionS256, req.Tenant)
		if perr != nil {
			return nil, perr
		}
		return &tokenOutcome{result: map[string]any{"person_token": token, "expires_in": expiresIn}}, nil
	}

	caps := req.Capabilities
	if caps == nil && mission != nil {
		caps = mission.Capabilities
	}
	if !slices.Contains(caps, "interaction") {
		return nil, &protocolError{status: http.StatusForbidden, code: errUserUnreachable,
			detail: "the person must approve this agent and the agent cannot direct them to the person server"}
	}
	p := s.newPending(kindPersonToken, agent)
	p.KeyJWK, p.TokenNotAfter, p.SubagentKey = keyJWK, notAfter, subagentKey
	p.Resource, p.MissionS256, p.Tenant = req.Resource, req.MissionS256, req.Tenant
	p.Justification, p.Platform, p.Device, p.Capabilities = req.Justification, req.Platform, req.Device, caps
	p.UserID = userID
	if err := s.store.addPending(ctx, p); err != nil {
		return nil, serverError()
	}
	return &tokenOutcome{pending: p}, nil
}

// requestAuthToken handles an auth token request in three-party mode.
func (s *aauthService) requestAuthToken(ctx context.Context, agent *agentIdentity,
	req *authTokenRequest) (*tokenOutcome, *protocolError) {
	if agent.parentAgent != "" {
		return nil, badRequest(errInvalidRequest, "a sub-agent's parent requests its auth tokens")
	}
	if req.ResourceToken == "" || req.PresentedToken == "" {
		return nil, badRequest(errInvalidRequest, "resource_token and presented_token are required")
	}
	if req.UpstreamToken != "" {
		return nil, badRequest(errInvalidUpstreamToken, "call chaining is not supported by this person server")
	}
	keyJWK, keyJKT, notAfter, subagentKey := agent.jwk, agent.jkt, agent.expiresAt, ""
	if req.SubagentToken != "" {
		sub, perr := s.verifySubagentToken(ctx, agent, req.SubagentToken)
		if perr != nil {
			return nil, perr
		}
		keyJWK, keyJKT, notAfter, subagentKey = sub.jwk, sub.jkt, min(notAfter, sub.expiresAt), sub.key()
	}

	rt, perr := s.verifyResourceToken(ctx, req.ResourceToken, keyJKT)
	if perr != nil {
		return nil, perr
	}
	presentedExp, perr := s.verifyPresentedToken(ctx, req.PresentedToken, rt)
	if perr != nil {
		return nil, perr
	}
	notAfter = min(notAfter, presentedExp)
	if rt.MissionS256 != "" {
		mission, perr := s.activeMission(ctx, rt.MissionS256, agent.key())
		if perr != nil {
			return nil, perr
		}
		notAfter = min(notAfter, mission.ExpiresAt.Unix())
	}

	userID, err := s.store.subjectOwner(ctx, rt.Issuer, rt.Subject)
	if err != nil {
		return nil, serverError()
	}
	bound, err := s.store.getBinding(ctx, agent.key())
	if err != nil {
		return nil, serverError()
	}
	if userID == "" || bound != userID {
		return nil, badRequest(errInvalidPresentedToken, "the presented token was not issued for this agent's person")
	}

	scopes := strings.Fields(rt.Scope)
	consent, err := s.store.getConsent(ctx, userID, agent.key(), rt.Issuer)
	if err != nil {
		return nil, serverError()
	}
	if consent != nil && coversScopes(consent.Scopes, scopes) {
		token, expiresIn, perr := s.mintAuthToken(ctx, userID, agent.key(), rt, keyJWK, notAfter)
		if perr != nil {
			return nil, perr
		}
		return &tokenOutcome{result: map[string]any{"auth_token": token, "expires_in": expiresIn}}, nil
	}

	caps := req.Capabilities
	if !slices.Contains(caps, "interaction") {
		return nil, &protocolError{status: http.StatusForbidden, code: errUserUnreachable,
			detail: "the person must consent and the agent cannot direct them to the person server"}
	}
	p := s.newPending(kindAuthToken, agent)
	p.KeyJWK, p.TokenNotAfter, p.SubagentKey = keyJWK, notAfter, subagentKey
	p.Resource, p.ResourceToken, p.UserID = rt.Issuer, rt, userID
	p.Justification, p.Platform, p.Device, p.Capabilities = req.Justification, req.Platform, req.Device, caps
	if err := s.store.addPending(ctx, p); err != nil {
		return nil, serverError()
	}
	return &tokenOutcome{pending: p}, nil
}

// proposeMission handles a mission proposal. Every proposal goes to the person.
func (s *aauthService) proposeMission(ctx context.Context, agent *agentIdentity,
	proposal *missionProposal) (*tokenOutcome, *protocolError) {
	if agent.parentAgent != "" {
		return nil, badRequest(errInvalidRequest, "a sub-agent cannot own a mission")
	}
	if strings.TrimSpace(proposal.Description) == "" {
		return nil, badRequest(errInvalidRequest, "description is required")
	}
	for _, r := range proposal.Resources {
		if !isServerIdentifier(r) {
			return nil, badRequest(errInvalidRequest, "resources must be server identifiers")
		}
	}
	userID, err := s.store.getBinding(ctx, agent.key())
	if err != nil {
		return nil, serverError()
	}
	p := s.newPending(kindMission, agent)
	p.KeyJWK, p.TokenNotAfter, p.Mission, p.UserID = agent.jwk, agent.expiresAt, proposal, userID
	p.Capabilities = proposal.Capabilities
	if err := s.store.addPending(ctx, p); err != nil {
		return nil, serverError()
	}
	return &tokenOutcome{pending: p}, nil
}

// missionRequest handles a request to a mission's own URL.
func (s *aauthService) missionRequest(ctx context.Context, agent *agentIdentity, s256 string,
	action *missionAction) (*tokenOutcome, *protocolError) {
	if len(s256) != 43 {
		return nil, badRequest(errInvalidRequest, "malformed mission reference")
	}
	if action.Action != "update" && action.Action != "completion" {
		return nil, badRequest(errInvalidRequest, "action must be update or completion")
	}
	mission, err := s.store.getMission(ctx, s256)
	if err != nil {
		return nil, serverError()
	}
	// The same answer whether the mission is absent or another agent's.
	if mission == nil || subtle.ConstantTimeCompare([]byte(mission.AgentKey), []byte(agent.key())) != 1 {
		return nil, &protocolError{status: http.StatusNotFound, code: errMissionNotFound}
	}
	if perr := s.checkMissionActive(ctx, mission); perr != nil {
		return nil, perr
	}
	switch action.Action {
	case "update":
		if strings.TrimSpace(action.Description) == "" {
			return nil, badRequest(errInvalidRequest, "description is required")
		}
		update, _ := json.Marshal(map[string]string{
			"description": action.Description, "accepted_at": s.now().UTC().Format(time.RFC3339)})
		mission.Updates = append(mission.Updates, update)
		if err := s.store.putMission(ctx, mission); err != nil {
			return nil, serverError()
		}
		sum := sha256.Sum256(update)
		return &tokenOutcome{result: map[string]any{"s256": encodeB64URL(sum[:])}}, nil
	default:
		if strings.TrimSpace(action.Summary) == "" {
			return nil, badRequest(errInvalidRequest, "summary is required")
		}
		p := s.newPending(kindMissionCompletion, agent)
		p.MissionS256, p.Summary, p.UserID = s256, action.Summary, mission.UserID
		if err := s.store.addPending(ctx, p); err != nil {
			return nil, serverError()
		}
		return &tokenOutcome{pending: p}, nil
	}
}

// poll returns the state of a pending request to the agent that started it.
func (s *aauthService) poll(ctx context.Context, agent *agentIdentity, id string) (*tokenOutcome, *protocolError) {
	p, err := s.store.getPending(ctx, id)
	if err != nil {
		return nil, serverError()
	}
	if p == nil || p.AgentKey != agent.key() {
		return nil, &protocolError{status: http.StatusNotFound, code: errNotFound}
	}
	switch p.State {
	case stateApproved:
		_ = s.store.deletePending(ctx, id)
		return &tokenOutcome{result: p.Result}, nil
	case stateDenied:
		_ = s.store.deletePending(ctx, id)
		return nil, &protocolError{status: p.ErrorStatus, code: p.ErrorCode, detail: p.ErrorDetail}
	}
	if s.now().After(p.ExpiresAt) {
		_ = s.store.deletePending(ctx, id)
		if p.State == statePending {
			return nil, &protocolError{status: http.StatusRequestTimeout, code: errExpired}
		}
		return nil, &protocolError{status: http.StatusForbidden, code: errAbandoned}
	}
	return &tokenOutcome{pending: p}, nil
}

// startInteraction consumes an interaction code and returns the sign-in URL the person goes to.
func (s *aauthService) startInteraction(ctx context.Context, rawCode string) (string, *interactionView) {
	code := normalizeCode(rawCode)
	if len(code) < interactionCodeLength {
		return "", &interactionView{Error: "Enter the code your agent showed you."}
	}
	id, err := s.store.takeCode(ctx, code)
	if err != nil || id == "" {
		return "", &interactionView{Error: "This code is not valid, or it has already been used."}
	}
	p, err := s.store.getPending(ctx, id)
	if err != nil || p == nil || p.State != statePending || s.now().After(p.ExpiresAt) {
		return "", &interactionView{Error: "This request has expired. Ask your agent to try again."}
	}
	verifier := randomToken(32)
	loginState := randomToken(24)
	p.State, p.PKCEVerifier, p.LoginState = stateInteracting, verifier, loginState
	if err := s.store.swapPending(ctx, p, statePending); err != nil {
		return "", &interactionView{Error: "This request is no longer waiting for you."}
	}
	if err := s.store.putLoginState(ctx, loginState, p.ID, p.ExpiresAt); err != nil {
		return "", &interactionView{Error: "Something went wrong. Try again."}
	}
	challenge := sha256.Sum256([]byte(verifier))
	query := url.Values{
		"client_id":             {s.interactionClientID},
		"redirect_uri":          {s.callbackURL()},
		"response_type":         {"code"},
		"scope":                 {"openid"},
		"state":                 {loginState},
		"code_challenge":        {encodeB64URL(challenge[:])},
		"code_challenge_method": {"S256"},
	}
	return s.issuer + "/oauth2/authorize?" + query.Encode(), nil
}

// completeSignIn redeems the person's authorization code and returns the consent page.
func (s *aauthService) completeSignIn(ctx context.Context, loginState, code, oauthError string) *interactionView {
	id, err := s.store.takeLoginState(ctx, loginState)
	if err != nil || id == "" {
		return &interactionView{Error: "This sign-in is not valid, or it has already been used."}
	}
	p, err := s.store.getPending(ctx, id)
	if err != nil || p == nil || p.State != stateInteracting {
		return &interactionView{Error: "This request is no longer waiting for you."}
	}
	if oauthError != "" || code == "" {
		s.deny(ctx, p, stateInteracting, http.StatusForbidden, errAbandoned, "the person did not sign in")
		return &interactionView{Error: "Sign-in did not complete. Ask your agent to try again."}
	}
	record, err := s.codeRedeemer.GetAuthorizationCodeDetails(ctx, s.interactionClientID, code)
	if err != nil || record == nil || record.RedirectURI != s.callbackURL() ||
		record.CodeChallenge != pkceChallenge(p.PKCEVerifier) || record.AuthorizedUserID == "" {
		s.deny(ctx, p, stateInteracting, http.StatusForbidden, errAbandoned, "sign-in could not be verified")
		return &interactionView{Error: "Sign-in could not be verified. Ask your agent to try again."}
	}
	userID := record.AuthorizedUserID
	bound, err := s.store.getBinding(ctx, p.AgentKey)
	if err != nil {
		return &interactionView{Error: "Something went wrong. Try again."}
	}
	if (bound != "" && bound != userID) || (p.UserID != "" && p.UserID != userID) {
		s.deny(ctx, p, stateInteracting, http.StatusForbidden, errDenied, "this agent acts for another person")
		return &interactionView{Error: "This agent acts for another person. Sign in as that person."}
	}
	p.State, p.UserID, p.DecisionToken = stateAuthenticated, userID, randomToken(24)
	if err := s.store.swapPending(ctx, p, stateInteracting); err != nil {
		return &interactionView{Error: "This request is no longer waiting for you."}
	}
	return s.consentView(ctx, p, bound == "")
}

// decide records the person's decision on a pending request.
func (s *aauthService) decide(ctx context.Context, id, decisionToken string, approve bool) *interactionView {
	p, err := s.store.getPending(ctx, id)
	if err != nil || p == nil || p.State != stateAuthenticated || p.DecisionToken == "" ||
		subtle.ConstantTimeCompare([]byte(p.DecisionToken), []byte(decisionToken)) != 1 {
		return &interactionView{Error: "This request is no longer waiting for you."}
	}
	if !approve {
		s.deny(ctx, p, stateAuthenticated, http.StatusForbidden, errDenied, "the person declined the request")
		return &interactionView{Message: "You declined the request. You can close this window."}
	}
	result, perr := s.resolveApproved(ctx, p)
	if perr != nil {
		s.deny(ctx, p, stateAuthenticated, perr.status, perr.code, perr.detail)
		return &interactionView{Error: "The request could not be completed: " + perr.detail}
	}
	data, _ := json.Marshal(result)
	p.State, p.Result, p.DecisionToken = stateApproved, data, ""
	if err := s.store.swapPending(ctx, p, stateAuthenticated); err != nil {
		return &interactionView{Error: "This request is no longer waiting for you."}
	}
	return &interactionView{Message: "Approved. Your agent can continue. You can close this window."}
}

// resolveApproved performs what the person approved and returns the agent's 200 body.
func (s *aauthService) resolveApproved(ctx context.Context, p *pendingRequest) (any, *protocolError) {
	if p.Kind != kindMissionCompletion {
		if err := s.store.bindAgent(ctx, p.AgentKey, p.UserID); err != nil {
			return nil, &protocolError{status: http.StatusForbidden, code: errDenied,
				detail: "this agent acts for another person"}
		}
	}
	switch p.Kind {
	case kindPersonToken:
		if err := s.addConsent(ctx, p.UserID, p.AgentKey, p.Resource, true, nil); err != nil {
			return nil, serverError()
		}
		token, expiresIn, perr := s.mintPersonToken(ctx, p.UserID, p.AgentKey, p.Resource, p.KeyJWK,
			p.TokenNotAfter, p.MissionS256, p.Tenant)
		if perr != nil {
			return nil, perr
		}
		return map[string]any{"person_token": token, "expires_in": expiresIn}, nil
	case kindAuthToken:
		if err := s.addConsent(ctx, p.UserID, p.AgentKey, p.Resource, true,
			strings.Fields(p.ResourceToken.Scope)); err != nil {
			return nil, serverError()
		}
		token, expiresIn, perr := s.mintAuthToken(ctx, p.UserID, p.AgentKey, p.ResourceToken, p.KeyJWK,
			p.TokenNotAfter)
		if perr != nil {
			return nil, perr
		}
		return map[string]any{"auth_token": token, "expires_in": expiresIn}, nil
	case kindMission:
		return s.approveMission(ctx, p)
	default:
		mission, err := s.store.getMission(ctx, p.MissionS256)
		if err != nil || mission == nil {
			return nil, serverError()
		}
		mission.Status, mission.TerminationReason = missionTerminated, "completed"
		if err := s.store.putMission(ctx, mission); err != nil {
			return nil, serverError()
		}
		return map[string]any{"s256": mission.S256, "mission_status": missionTerminated,
			"termination_reason": "completed"}, nil
	}
}

// approveMission creates the approved mission and its person tokens.
func (s *aauthService) approveMission(ctx context.Context, p *pendingRequest) (any, *protocolError) {
	approvedAt := s.now().UTC()
	blob := missionBlob{
		Agent:             p.AgentSubject,
		ApprovedAt:        approvedAt.Format(time.RFC3339),
		Description:       p.Mission.Description,
		ApprovedTools:     p.Mission.Tools,
		ApprovedResources: p.Mission.Resources,
	}
	blobBytes, err := json.Marshal(blob)
	if err != nil {
		return nil, serverError()
	}
	sum := sha256.Sum256(blobBytes)
	s256 := encodeB64URL(sum[:])
	caps := []string{"interaction"}
	record := &missionRecord{
		S256: s256, Blob: blobBytes, AgentKey: p.AgentKey, UserID: p.UserID, Status: missionActive,
		ExpiresAt: approvedAt.Add(time.Duration(longLivedTTL) * time.Second), ApprovedResources: p.Mission.Resources,
		Capabilities: caps,
	}
	if err := s.store.putMission(ctx, record); err != nil {
		return nil, serverError()
	}
	response := map[string]any{"s256": s256, "mission": encodeB64URL(blobBytes), "capabilities": caps}
	if len(p.Mission.Resources) > 0 {
		tokens := map[string]string{}
		for _, resource := range p.Mission.Resources {
			if err := s.addConsent(ctx, p.UserID, p.AgentKey, resource, true, nil); err != nil {
				return nil, serverError()
			}
			token, _, perr := s.mintPersonToken(ctx, p.UserID, p.AgentKey, resource, p.KeyJWK,
				p.TokenNotAfter, s256, "")
			if perr != nil {
				return nil, perr
			}
			tokens[resource] = token
		}
		response["person_tokens"] = tokens
	}
	return response, nil
}

// mintPersonToken issues an aa-person+jwt.
func (s *aauthService) mintPersonToken(ctx context.Context, userID, agentKey, resource string,
	keyJWK map[string]any, notAfter int64, missionS256, tenant string) (string, int64, *protocolError) {
	subject, err := s.store.directedSubject(ctx, userID, resource, func() string { return randomHex(16) })
	if err != nil {
		return "", 0, serverError()
	}
	now := s.now().Unix()
	exp := min(now+int64(personTokenLifetime.Seconds()), notAfter)
	if exp <= now {
		return "", 0, badRequest(errInvalidRequest, "the agent token expires too soon")
	}
	jti := randomToken(16)
	claims := map[string]any{
		"iss": s.issuer, "dwk": dwkPerson, "aud": resource, "sub": subject,
		"cnf": map[string]any{"jwk": keyJWK}, "jti": jti, "iat": now, "exp": exp,
	}
	if missionS256 != "" {
		claims["mission_s256"] = missionS256
	}
	if tenant != "" {
		claims["tenant"] = tenant
	}
	token, err := s.sign(ctx, typPersonToken, claims)
	if err != nil {
		return "", 0, serverError()
	}
	_ = s.store.recordIssuedToken(ctx, jti, &issuedToken{Type: typPersonToken, Audience: resource,
		AgentKey: agentKey, UserID: userID, ExpiresAt: exp})
	return token, exp - now, nil
}

// mintAuthToken issues an aa-auth+jwt for the resource token's resource.
func (s *aauthService) mintAuthToken(ctx context.Context, userID, agentKey string, rt *resourceTokenClaims,
	keyJWK map[string]any, notAfter int64) (string, int64, *protocolError) {
	now := s.now().Unix()
	exp := min(now+int64(authTokenLifetime.Seconds()), notAfter)
	if exp <= now {
		return "", 0, badRequest(errExpiredPresentedToken, "the presented token has expired")
	}
	jti := randomToken(16)
	claims := s.identityClaims(ctx, userID, strings.Fields(rt.Scope))
	for k, v := range map[string]any{
		"iss": s.issuer, "dwk": dwkPerson, "aud": rt.Issuer, "ps": s.issuer, "sub": rt.Subject,
		"cnf": map[string]any{"jwk": keyJWK}, "jti": jti, "iat": now, "exp": exp,
	} {
		claims[k] = v
	}
	for k, v := range map[string]string{"scope": rt.Scope, "account": rt.Account,
		"mission_s256": rt.MissionS256, "tenant": rt.Tenant} {
		if v != "" {
			claims[k] = v
		}
	}
	token, err := s.sign(ctx, typAuthToken, claims)
	if err != nil {
		return "", 0, serverError()
	}
	_ = s.store.recordIssuedToken(ctx, jti, &issuedToken{Type: typAuthToken, Audience: rt.Issuer,
		AgentKey: agentKey, UserID: userID, ExpiresAt: exp})
	return token, exp - now, nil
}

// identityClaims returns the person's attributes released by the identity scopes.
func (s *aauthService) identityClaims(ctx context.Context, userID string, scopes []string) map[string]any {
	claims := map[string]any{}
	if s.actorProvider == nil {
		return claims
	}
	var names []string
	for _, scope := range scopes {
		names = append(names, identityScopeClaims[scope]...)
	}
	if len(names) == 0 {
		return claims
	}
	entity, svcErr := s.actorProvider.GetActor(ctx, userID)
	if svcErr != nil || entity == nil {
		return claims
	}
	attrs := map[string]any{}
	_ = json.Unmarshal(entity.Attributes, &attrs)
	for _, name := range names {
		if v, ok := attrs[name]; ok {
			claims[name] = v
		}
	}
	return claims
}

// sign signs claims as a JWT of the given type with the person server's key.
func (s *aauthService) sign(ctx context.Context, typ string, claims map[string]any) (string, error) {
	header, err := json.Marshal(map[string]string{"alg": s.alg, "typ": typ, "kid": s.kid})
	if err != nil {
		return "", err
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	input := encodeB64URL(header) + "." + encodeB64URL(payload)
	signature, err := s.crypto.Sign(ctx, providers.KeyRef{KeyID: s.signingKeyID}, s.cryptoAlg, []byte(input))
	if err != nil {
		return "", err
	}
	return input + "." + encodeB64URL(signature), nil
}

// verifySubagentToken verifies a sub-agent's agent token presented by its parent.
func (s *aauthService) verifySubagentToken(ctx context.Context, parent *agentIdentity, token string) (
	*agentIdentity, *protocolError) {
	sub, sigErr := s.verifyAgentToken(ctx, token)
	if sigErr != nil {
		if sigErr.code == sigErrExpiredJWT {
			return nil, badRequest(errExpiredSubagentToken, sigErr.detail)
		}
		return nil, badRequest(errInvalidSubagentToken, sigErr.detail)
	}
	if sub.parentAgent != parent.subject || sub.issuer != parent.issuer {
		return nil, badRequest(errInvalidSubagentToken, "the sub-agent token does not name the signing agent")
	}
	return sub, nil
}

// verifyResourceToken verifies a resource token addressed to this person server.
func (s *aauthService) verifyResourceToken(ctx context.Context, token, keyJKT string) (
	*resourceTokenClaims, *protocolError) {
	p, err := parseJWT(token)
	if err != nil || p.headerString("typ") != typResourceToken || p.claimString("dwk") != dwkResource {
		return nil, badRequest(errInvalidResourceToken, "a resource token is required")
	}
	exp := p.exp()
	if exp == 0 {
		return nil, badRequest(errInvalidResourceToken, "the resource token has no exp")
	}
	if exp <= s.now().Unix() {
		return nil, badRequest(errExpiredResourceToken, "the resource token has expired")
	}
	if _, err := verifyIssuerJWT(ctx, s.crypto, s.fetcher, p, dwkResource); err != nil {
		s.logger.Debug(ctx, "Resource token verification failed", log.Error(err))
		return nil, badRequest(errInvalidResourceToken, "the resource token does not verify")
	}
	var rt resourceTokenClaims
	raw, _ := json.Marshal(p.claims)
	_ = json.Unmarshal(raw, &rt)
	if rt.PresentedJTI == "" {
		rt.PresentedJTI = p.claimString("person_token_jti")
	}
	if p.claimString("aud") != s.issuer {
		return nil, badRequest(errInvalidResourceToken,
			"the resource token is not addressed to this person server (four-party federation is not supported)")
	}
	if rt.PersonServer != s.issuer {
		return nil, badRequest(errInvalidResourceToken, "the resource token names another person server")
	}
	if rt.AgentJKT == "" || rt.AgentJKT != keyJKT {
		return nil, badRequest(errInvalidResourceToken, "agent_jkt does not match the agent's key")
	}
	if rt.Subject == "" || rt.PresentedJTI == "" {
		return nil, badRequest(errInvalidResourceToken, "sub and presented_jti are required")
	}
	return &rt, nil
}

// verifyPresentedToken verifies the person or auth token the agent presented to the resource, and
// matches it against the resource token. It returns the presented token's exp.
func (s *aauthService) verifyPresentedToken(ctx context.Context, token string, rt *resourceTokenClaims) (
	int64, *protocolError) {
	p, err := parseJWT(token)
	if err != nil {
		return 0, badRequest(errInvalidPresentedToken, "malformed presented token")
	}
	typ := p.headerString("typ")
	if typ != typPersonToken && typ != typAuthToken {
		return 0, badRequest(errInvalidPresentedToken, "the presented token must be a person token or an auth token")
	}
	if p.claimString("iss") != s.issuer || p.claimString("dwk") != dwkPerson || p.headerString("alg") != s.alg {
		return 0, badRequest(errInvalidPresentedToken, "the presented token was not issued by this person server")
	}
	if err := s.crypto.Verify(ctx, providers.KeyRef{PublicKeyJWK: stripAlg(s.publicJWK)}, s.cryptoAlg,
		p.signingInput, p.signature); err != nil {
		return 0, badRequest(errInvalidPresentedToken, "the presented token does not verify")
	}
	exp := p.exp()
	if exp <= s.now().Unix() {
		return 0, badRequest(errExpiredPresentedToken, "the presented token has expired")
	}
	if p.claimString("aud") != rt.Issuer {
		return 0, badRequest(errInvalidPresentedToken, "the presented token is for another resource")
	}
	jwk := p.cnfJWK()
	if jwk == nil {
		return 0, badRequest(errInvalidPresentedToken, "the presented token has no cnf.jwk")
	}
	if jkt, err := jktOf(jwk); err != nil || jkt != rt.AgentJKT {
		return 0, badRequest(errInvalidPresentedToken, "the presented token is bound to another key")
	}
	ps := p.claimString("iss")
	if typ == typAuthToken {
		ps = p.claimString("ps")
	}
	if p.claimString("jti") != rt.PresentedJTI || ps != rt.PersonServer || p.claimString("sub") != rt.Subject ||
		p.claimString("mission_s256") != rt.MissionS256 || p.claimString("tenant") != rt.Tenant {
		return 0, badRequest(errInvalidResourceToken, "the resource token does not match the presented token")
	}
	return exp, nil
}

// activeMission returns the agent's mission when it exists and is active.
func (s *aauthService) activeMission(ctx context.Context, s256, agentKey string) (*missionRecord, *protocolError) {
	mission, err := s.store.getMission(ctx, s256)
	if err != nil {
		return nil, serverError()
	}
	if mission == nil || mission.AgentKey != agentKey {
		return nil, badRequest(errInvalidRequest, "mission_s256 does not name a mission of this agent")
	}
	if perr := s.checkMissionActive(ctx, mission); perr != nil {
		return nil, perr
	}
	return mission, nil
}

// checkMissionActive rejects a terminated mission, terminating one past its expiry.
func (s *aauthService) checkMissionActive(ctx context.Context, mission *missionRecord) *protocolError {
	if mission.Status == missionActive && s.now().After(mission.ExpiresAt) {
		mission.Status, mission.TerminationReason = missionTerminated, "expired"
		_ = s.store.putMission(ctx, mission)
	}
	if mission.Status != missionActive {
		return &protocolError{status: http.StatusForbidden, code: errMissionTerminated,
			extra: map[string]any{"mission_status": missionTerminated, "termination_reason": mission.TerminationReason}}
	}
	return nil
}

func (s *aauthService) addConsent(ctx context.Context, userID, agentKey, resource string, identity bool,
	scopes []string) error {
	c, err := s.store.getConsent(ctx, userID, agentKey, resource)
	if err != nil {
		return err
	}
	if c == nil {
		c = &consentRecord{}
	}
	c.Identity = c.Identity || identity
	for _, scope := range scopes {
		if !slices.Contains(c.Scopes, scope) {
			c.Scopes = append(c.Scopes, scope)
		}
	}
	return s.store.putConsent(ctx, userID, agentKey, resource, c)
}

// deny ends a pending request with an error the agent sees on its next poll.
func (s *aauthService) deny(ctx context.Context, p *pendingRequest, expected string, status int, code, detail string) {
	p.State, p.ErrorStatus, p.ErrorCode, p.ErrorDetail, p.DecisionToken = stateDenied, status, code, detail, ""
	if err := s.store.swapPending(ctx, p, expected); err != nil {
		s.logger.Debug(ctx, "Pending request changed before it could be denied", log.String("id", p.ID))
	}
}

func (s *aauthService) newPending(kind string, agent *agentIdentity) *pendingRequest {
	return &pendingRequest{
		ID:           randomToken(18),
		Kind:         kind,
		State:        statePending,
		Code:         newInteractionCode(),
		AgentKey:     agent.key(),
		AgentName:    agent.name,
		AgentIssuer:  agent.issuer,
		AgentSubject: agent.subject,
		ExpiresAt:    s.now().Add(pendingLifetime),
	}
}

// consentView builds the consent page for an authenticated pending request. Resource-asserted
// content comes from the resource's metadata and resource token; agent-asserted content is shown
// apart from it.
func (s *aauthService) consentView(ctx context.Context, p *pendingRequest, newAgent bool) *interactionView {
	view := &interactionView{
		Kind: p.Kind, AgentName: p.AgentName, AgentID: p.AgentSubject, AgentProvider: p.AgentIssuer,
		NewAgent: newAgent, Platform: p.Platform, Device: p.Device, Justification: p.Justification,
		Resource: p.Resource, Mission: p.Mission, Summary: p.Summary, PendingID: p.ID,
		DecisionToken: p.DecisionToken,
	}
	if p.ResourceToken != nil {
		view.Scopes = strings.Fields(p.ResourceToken.Scope)
	}
	if p.Resource != "" {
		if entry, err := s.fetcher.metadata(ctx, p.Resource, dwkResource, false); err == nil {
			view.ResourceName = entry.metadata.Name
			view.ResourceDesc = entry.metadata.Description
			view.AccessMode = entry.metadata.AccessMode
		}
	}
	if p.Kind == kindMissionCompletion {
		if m, err := s.store.getMission(ctx, p.MissionS256); err == nil && m != nil {
			var blob missionBlob
			_ = json.Unmarshal(m.Blob, &blob)
			view.MissionDesc = blob.Description
		}
	}
	return view
}

func (s *aauthService) callbackURL() string {
	return s.issuer + pathInteraction + "/callback"
}

// interactionURL is the URL the agent directs the person to.
func (s *aauthService) interactionURL() string {
	return s.issuer + pathInteraction
}

func (s *aauthService) pendingURL(id string) string {
	return s.issuer + pathPending + "/" + id
}

func coversScopes(granted, requested []string) bool {
	for _, scope := range requested {
		if !slices.Contains(granted, scope) {
			return false
		}
	}
	return true
}

func stripAlg(jwk map[string]any) map[string]any {
	out := make(map[string]any, len(jwk))
	for k, v := range jwk {
		if k != "alg" && k != "kid" && k != "use" {
			out[k] = v
		}
	}
	return out
}

func pkceChallenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return encodeB64URL(sum[:])
}

// newInteractionCode returns an 8-symbol Crockford base32 code carrying 40 bits of entropy.
func newInteractionCode() string {
	buf := make([]byte, interactionCodeLength)
	_, _ = rand.Read(buf)
	code := make([]byte, interactionCodeLength)
	for i, b := range buf {
		code[i] = crockfordAlphabet[int(b)%len(crockfordAlphabet)]
	}
	return string(code)
}

// displayCode groups a code for display, as in A1B2-C3D4.
func displayCode(code string) string {
	if len(code) != interactionCodeLength {
		return code
	}
	return code[:4] + "-" + code[4:]
}

// normalizeCode strips hyphens, uppercases, and folds the Crockford decode aliases.
func normalizeCode(code string) string {
	code = strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(code), "-", ""))
	return strings.NewReplacer("I", "1", "L", "1", "O", "0").Replace(code)
}

func randomToken(n int) string {
	buf := make([]byte, n)
	_, _ = rand.Read(buf)
	return base64.RawURLEncoding.EncodeToString(buf)
}

func randomHex(n int) string {
	buf := make([]byte, n)
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}
