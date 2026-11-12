// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

package aauth

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	oauth2authz "github.com/thunder-id/thunderid/internal/oauth/oauth2/authz"
	"github.com/thunder-id/thunderid/pkg/thunderidengine/providers"
	"github.com/thunder-id/thunderid/tests/mocks/crypto/cryptomock"
	"github.com/thunder-id/thunderid/tests/mocks/httpmock"
	"github.com/thunder-id/thunderid/tests/mocks/runtimestoreprovidermock"
)

const (
	testIssuer   = "https://ps.example"
	testResource = "https://resource.example"
	testUser     = "user-1"
)

type ServiceTestSuite struct {
	suite.Suite
	psPriv     ed25519.PrivateKey
	agentPub   ed25519.PublicKey
	agentPriv  ed25519.PrivateKey
	store      map[string][]byte
	mu         sync.Mutex
	mockStore  *runtimestoreprovidermock.RuntimeStoreProviderMock
	mockCrypto *cryptomock.RuntimeCryptoProviderMock
	redeemer   *authorizationCodeRedeemerMock
	mockHTTP   *httpmock.HTTPClientInterfaceMock
	docs       map[string]string
	service    *aauthService
	agent      *agentIdentity
}

func TestServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}

func (s *ServiceTestSuite) SetupTest() {
	psPub, psPriv, _ := ed25519.GenerateKey(rand.Reader)
	s.psPriv = psPriv
	s.agentPub, s.agentPriv, _ = ed25519.GenerateKey(rand.Reader)
	s.store = map[string][]byte{}
	s.mockStore = runtimestoreprovidermock.NewRuntimeStoreProviderMock(s.T())
	s.mockCrypto = cryptomock.NewRuntimeCryptoProviderMock(s.T())
	s.redeemer = newAuthorizationCodeRedeemerMock(s.T())
	s.mockHTTP = httpmock.NewHTTPClientInterfaceMock(s.T())
	s.docs = map[string]string{}
	s.mockHTTP.EXPECT().Do(mock.Anything).RunAndReturn(func(r *http.Request) (*http.Response, error) {
		body, ok := s.docs[r.URL.String()]
		if !ok {
			return nil, errors.New("offline")
		}
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body))}, nil
	}).Maybe()
	s.backStoreWithMap()

	s.mockCrypto.EXPECT().GetPublicKeys(mock.Anything, providers.PublicKeyFilter{KeyID: "aauth-key"}).
		Return([]providers.PublicKeyInfo{{KeyID: "aauth-key", Kid: "kid-1", PublicKey: psPub}}, nil)
	s.mockCrypto.EXPECT().Sign(mock.Anything, providers.KeyRef{KeyID: "aauth-key"}, "EdDSA", mock.Anything).
		RunAndReturn(func(_ context.Context, _ providers.KeyRef, _ string, content []byte) ([]byte, error) {
			return ed25519.Sign(psPriv, content), nil
		}).Maybe()
	s.mockCrypto.EXPECT().Verify(mock.Anything, mock.Anything, "EdDSA", mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, ref providers.KeyRef, _ string, content, sig []byte) error {
			x, _ := base64.RawURLEncoding.DecodeString(ref.PublicKeyJWK["x"].(string))
			if !ed25519.Verify(ed25519.PublicKey(x), content, sig) {
				return errors.New("bad signature")
			}
			return nil
		}).Maybe()

	var err error
	s.service, err = newAAuthService(context.Background(), testIssuer, "aauth-key", "AAUTH_PS", s.mockCrypto,
		newAAuthStore(s.mockStore), newIssuerKeyFetcher(s.mockHTTP), s.redeemer, nil)
	s.Require().NoError(err)
	agentJWK := map[string]any{"kty": "OKP", "crv": "Ed25519", "x": encodeB64URL(s.agentPub), "alg": "Ed25519"}
	jkt, _ := jktOf(agentJWK)
	s.agent = &agentIdentity{issuer: "https://agent.example", subject: "aauth:bot@agent.example", jwk: agentJWK,
		jkt: jkt, expiresAt: time.Now().Add(time.Hour).Unix(), name: "Bot"}
}

// backStoreWithMap answers the runtime store mock from an in-memory map.
func (s *ServiceTestSuite) backStoreWithMap() {
	ns := providers.NamespaceAAuth
	s.mockStore.EXPECT().Get(mock.Anything, ns, mock.Anything).RunAndReturn(
		func(_ context.Context, _ providers.RuntimeStoreNamespace, k string) ([]byte, error) {
			s.mu.Lock()
			defer s.mu.Unlock()
			return s.store[k], nil
		}).Maybe()
	s.mockStore.EXPECT().Put(mock.Anything, ns, mock.Anything, mock.Anything, mock.Anything).RunAndReturn(
		func(_ context.Context, _ providers.RuntimeStoreNamespace, k string, v []byte, _ int64) error {
			s.mu.Lock()
			defer s.mu.Unlock()
			s.store[k] = v
			return nil
		}).Maybe()
	s.mockStore.EXPECT().PutIfNotExists(mock.Anything, ns, mock.Anything, mock.Anything, mock.Anything).RunAndReturn(
		func(_ context.Context, _ providers.RuntimeStoreNamespace, k string, v []byte, _ int64) (bool, error) {
			s.mu.Lock()
			defer s.mu.Unlock()
			if _, ok := s.store[k]; ok {
				return false, nil
			}
			s.store[k] = v
			return true, nil
		}).Maybe()
	s.mockStore.EXPECT().Take(mock.Anything, ns, mock.Anything).RunAndReturn(
		func(_ context.Context, _ providers.RuntimeStoreNamespace, k string) ([]byte, error) {
			s.mu.Lock()
			defer s.mu.Unlock()
			v := s.store[k]
			delete(s.store, k)
			return v, nil
		}).Maybe()
	s.mockStore.EXPECT().Delete(mock.Anything, ns, mock.Anything).RunAndReturn(
		func(_ context.Context, _ providers.RuntimeStoreNamespace, k string) error {
			s.mu.Lock()
			defer s.mu.Unlock()
			delete(s.store, k)
			return nil
		}).Maybe()
	s.mockStore.EXPECT().CompareFieldAndSwapBeforeExpiry(mock.Anything, ns, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything).RunAndReturn(
		func(_ context.Context, _ providers.RuntimeStoreNamespace, k, field, expected string, exp time.Time,
			v []byte) (bool, error) {
			s.mu.Lock()
			defer s.mu.Unlock()
			current := map[string]any{}
			if json.Unmarshal(s.store[k], &current) != nil || current[field] != expected || time.Now().After(exp) {
				return false, nil
			}
			s.store[k] = v
			return true, nil
		}).Maybe()
}

func (s *ServiceTestSuite) personRequest() *personTokenRequest {
	return &personTokenRequest{Resource: testResource, Capabilities: []string{"interaction"}}
}

func (s *ServiceTestSuite) TestMetadataAndJWKS() {
	meta := s.service.metadata()
	s.Equal(testIssuer, meta["issuer"])
	s.Equal(testIssuer+pathPersonToken, meta["person_token_endpoint"])
	s.Equal(testIssuer+pathAuthToken, meta["auth_token_endpoint"])
	keys := s.service.jwks()["keys"].([]map[string]any)
	s.Equal("Ed25519", keys[0]["alg"])
	s.Equal("kid-1", keys[0]["kid"])
}

func (s *ServiceTestSuite) TestPersonTokenRequestValidation() {
	cases := []struct {
		name  string
		agent func(*agentIdentity)
		req   func(*personTokenRequest)
		code  string
	}{
		{"sub-agent signer", func(a *agentIdentity) { a.parentAgent = "aauth:parent@agent.example" }, nil,
			errInvalidRequest},
		{"resource with a port", nil, func(r *personTokenRequest) { r.Resource = "https://resource.example:8443" },
			errInvalidRequest},
		{"upstream token", nil, func(r *personTokenRequest) { r.UpstreamToken = "x" }, errInvalidUpstreamToken},
		{"unknown mission", nil, func(r *personTokenRequest) { r.MissionS256 = strings.Repeat("a", 43) },
			errInvalidRequest},
		{"no interaction capability", nil, func(r *personTokenRequest) { r.Capabilities = nil }, errUserUnreachable},
	}
	for _, tc := range cases {
		agent := *s.agent
		if tc.agent != nil {
			tc.agent(&agent)
		}
		req := s.personRequest()
		if tc.req != nil {
			tc.req(req)
		}
		_, perr := s.service.requestPersonToken(context.Background(), &agent, req)
		s.Require().NotNil(perr, tc.name)
		s.Equal(tc.code, perr.code, tc.name)
	}
}

func (s *ServiceTestSuite) TestFirstPersonTokenNeedsThePerson() {
	outcome, perr := s.service.requestPersonToken(context.Background(), s.agent, s.personRequest())
	s.Require().Nil(perr)
	s.Require().NotNil(outcome.pending)
	p := outcome.pending
	s.Equal(kindPersonToken, p.Kind)
	s.Len(p.Code, interactionCodeLength)
	s.Equal(p.Code, normalizeCode(displayCode(p.Code)))
	s.NotNil(s.store[keyCode+p.Code])
}

// approve walks a pending request through sign-in and the person's approval.
func (s *ServiceTestSuite) approve(p *pendingRequest, userID string) *interactionView {
	redirect, view := s.service.startInteraction(context.Background(), strings.ToLower(displayCode(p.Code)))
	s.Require().Nil(view)
	u, _ := url.Parse(redirect)
	s.Equal("AAUTH_PS", u.Query().Get("client_id"))
	s.Equal(testIssuer+"/aauth/interaction/callback", u.Query().Get("redirect_uri"))
	stored, _ := s.service.store.getPending(context.Background(), p.ID)
	s.redeemer.EXPECT().GetAuthorizationCodeDetails(mock.Anything, "AAUTH_PS", "code-1").Return(
		&oauth2authz.AuthorizationCode{RedirectURI: testIssuer + "/aauth/interaction/callback",
			CodeChallenge: pkceChallenge(stored.PKCEVerifier), AuthorizedUserID: userID}, nil).Once()
	consent := s.service.completeSignIn(context.Background(), u.Query().Get("state"), "code-1", "")
	if consent.Error != "" {
		return consent
	}
	s.NotEmpty(consent.DecisionToken)
	return s.service.decide(context.Background(), p.ID, consent.DecisionToken, true)
}

func (s *ServiceTestSuite) TestPersonTokenApprovedThenRemembered() {
	outcome, _ := s.service.requestPersonToken(context.Background(), s.agent, s.personRequest())
	done := s.approve(outcome.pending, testUser)
	s.Empty(done.Error)

	polled, perr := s.service.poll(context.Background(), s.agent, outcome.pending.ID)
	s.Require().Nil(perr)
	body := map[string]any{}
	s.Require().NoError(json.Unmarshal(polled.result.(json.RawMessage), &body))
	claims := s.decode(body["person_token"].(string), typPersonToken)
	s.Equal(testResource, claims["aud"])
	s.Equal(testIssuer, claims["iss"])
	s.Equal(dwkPerson, claims["dwk"])
	s.Nil(claims["scope"])
	s.Equal(s.agent.jwk["x"], claims["cnf"].(map[string]any)["jwk"].(map[string]any)["x"])

	_, perr = s.service.poll(context.Background(), s.agent, outcome.pending.ID)
	s.Equal(http.StatusNotFound, perr.status)

	again, perr := s.service.requestPersonToken(context.Background(), s.agent, s.personRequest())
	s.Require().Nil(perr)
	s.Nil(again.pending)
	second := s.decode(again.result.(map[string]any)["person_token"].(string), typPersonToken)
	s.Equal(claims["sub"], second["sub"])
}

func (s *ServiceTestSuite) TestBoundAgentRejectsAnotherPerson() {
	outcome, _ := s.service.requestPersonToken(context.Background(), s.agent, s.personRequest())
	s.approve(outcome.pending, testUser)
	other := s.personRequest()
	other.Resource = "https://other.example"
	next, _ := s.service.requestPersonToken(context.Background(), s.agent, other)
	view := s.approve(next.pending, "someone-else")
	s.Contains(view.Error, "another person")
	_, perr := s.service.poll(context.Background(), s.agent, next.pending.ID)
	s.Equal(errDenied, perr.code)
}

func (s *ServiceTestSuite) TestDenyAndPollStates() {
	outcome, _ := s.service.requestPersonToken(context.Background(), s.agent, s.personRequest())
	stranger := &agentIdentity{issuer: "https://x.example", subject: "aauth:x@x.example"}
	_, perr := s.service.poll(context.Background(), stranger, outcome.pending.ID)
	s.Equal(http.StatusNotFound, perr.status)

	waiting, perr := s.service.poll(context.Background(), s.agent, outcome.pending.ID)
	s.Nil(perr)
	s.NotNil(waiting.pending)

	redirect, _ := s.service.startInteraction(context.Background(), outcome.pending.Code)
	u, _ := url.Parse(redirect)
	view := s.service.completeSignIn(context.Background(), u.Query().Get("state"), "", "access_denied")
	s.NotEmpty(view.Error)
	_, perr = s.service.poll(context.Background(), s.agent, outcome.pending.ID)
	s.Equal(errAbandoned, perr.code)

	_, view = s.service.startInteraction(context.Background(), outcome.pending.Code)
	s.NotEmpty(view.Error, "a code is single use")
}

func (s *ServiceTestSuite) TestMissionLifecycle() {
	outcome, perr := s.service.proposeMission(context.Background(), s.agent, &missionProposal{
		Description: "# Plan", Resources: []string{testResource}})
	s.Require().Nil(perr)
	s.approve(outcome.pending, testUser)
	polled, _ := s.service.poll(context.Background(), s.agent, outcome.pending.ID)
	approval := map[string]any{}
	s.Require().NoError(json.Unmarshal(polled.result.(json.RawMessage), &approval))
	s256 := approval["s256"].(string)
	blob, _ := base64.RawURLEncoding.DecodeString(approval["mission"].(string))
	s.Equal(s256, pkceChallenge(string(blob)))
	pt := s.decode(approval["person_tokens"].(map[string]any)[testResource].(string), typPersonToken)
	s.Equal(s256, pt["mission_s256"])

	req := s.personRequest()
	req.MissionS256, req.Capabilities = s256, nil
	direct, perr := s.service.requestPersonToken(context.Background(), s.agent, req)
	s.Require().Nil(perr)
	s.Nil(direct.pending)

	_, perr = s.service.missionRequest(context.Background(), s.agent, "short", &missionAction{Action: "update"})
	s.Equal(errInvalidRequest, perr.code)
	_, perr = s.service.missionRequest(context.Background(), s.agent, s256, &missionAction{Action: "nope"})
	s.Equal(errInvalidRequest, perr.code)
	stranger := &agentIdentity{issuer: "https://x.example", subject: "aauth:x@x.example"}
	_, perr = s.service.missionRequest(context.Background(), stranger, s256, &missionAction{Action: "update",
		Description: "x"})
	s.Equal(errMissionNotFound, perr.code)

	update, perr := s.service.missionRequest(context.Background(), s.agent, s256, &missionAction{Action: "update",
		Description: "more"})
	s.Require().Nil(perr)
	s.Len(update.result.(map[string]any)["s256"], 43)

	completion, _ := s.service.missionRequest(context.Background(), s.agent, s256, &missionAction{
		Action: "completion", Summary: "done"})
	s.approve(completion.pending, testUser)
	_, perr = s.service.requestPersonToken(context.Background(), s.agent, req)
	s.Equal(errMissionTerminated, perr.code)
	s.Equal("completed", perr.extra["termination_reason"])
}

func (s *ServiceTestSuite) TestPresentedTokenChecks() {
	token, _, perr := s.service.mintPersonToken(context.Background(), testUser, s.agent.key(), testResource,
		s.agent.jwk, time.Now().Add(time.Hour).Unix(), "", "")
	s.Require().Nil(perr)
	claims := s.decode(token, typPersonToken)
	rt := &resourceTokenClaims{Issuer: testResource, PersonServer: testIssuer, Subject: claims["sub"].(string),
		PresentedJTI: claims["jti"].(string), AgentJKT: s.agent.jkt}
	_, perr = s.service.verifyPresentedToken(context.Background(), token, rt)
	s.Nil(perr)

	for name, mutate := range map[string]func(*resourceTokenClaims){
		"jti":     func(r *resourceTokenClaims) { r.PresentedJTI = "other" },
		"sub":     func(r *resourceTokenClaims) { r.Subject = "other" },
		"mission": func(r *resourceTokenClaims) { r.MissionS256 = "m" },
		"tenant":  func(r *resourceTokenClaims) { r.Tenant = "t" },
	} {
		bad := *rt
		mutate(&bad)
		_, perr = s.service.verifyPresentedToken(context.Background(), token, &bad)
		s.Equal(errInvalidResourceToken, perr.code, name)
	}
	other := *rt
	other.Issuer = "https://elsewhere.example"
	_, perr = s.service.verifyPresentedToken(context.Background(), token, &other)
	s.Equal(errInvalidPresentedToken, perr.code)
	other = *rt
	other.AgentJKT = "nope"
	_, perr = s.service.verifyPresentedToken(context.Background(), token, &other)
	s.Equal(errInvalidPresentedToken, perr.code)
	_, perr = s.service.verifyPresentedToken(context.Background(), token[:len(token)-4]+"AAAA", rt)
	s.Equal(errInvalidPresentedToken, perr.code)
}

func sha256Sum(v string) []byte {
	sum := sha256.Sum256([]byte(v))
	return sum[:]
}

func (s *ServiceTestSuite) TestHelpers() {
	for i := 0; i < 50; i++ {
		code := newInteractionCode()
		s.Len(code, interactionCodeLength)
		for _, c := range code {
			s.Contains(crockfordAlphabet, string(c))
		}
	}
	s.Equal("A1B2C3D4", normalizeCode(" a1b2-c3d4 "))
	s.Equal("0110ABCD", normalizeCode("OIL0-abcd"))
	s.Equal("ABCD-EFGH", displayCode("ABCDEFGH"))
	s.True(coversScopes([]string{"a", "b"}, []string{"b"}))
	s.False(coversScopes([]string{"a"}, []string{"a", "c"}))
	_, _, err := publicJWKFor("not a key")
	s.Error(err)
}

// decode verifies a token minted by the service and returns its claims.
func (s *ServiceTestSuite) decode(token, typ string) map[string]any {
	p, err := parseJWT(token)
	s.Require().NoError(err)
	s.Equal(typ, p.headerString("typ"))
	s.Equal("Ed25519", p.headerString("alg"))
	s.True(ed25519.Verify(s.psPriv.Public().(ed25519.PublicKey), p.signingInput, p.signature))
	return p.claims
}

// publishIssuer serves metadata and a JWKS for an issuer whose tokens priv signs.
func (s *ServiceTestSuite) publishIssuer(issuer, dwk string, priv ed25519.PrivateKey) {
	pub := priv.Public().(ed25519.PublicKey)
	s.docs[issuer+"/.well-known/"+dwk] = fmt.Sprintf(`{"issuer":%q,"jwks_uri":%q,"name":"Issuer"}`,
		issuer, issuer+"/jwks")
	s.docs[issuer+"/jwks"] = fmt.Sprintf(`{"keys":[{"kid":"k1","kty":"OKP","crv":"Ed25519","x":%q,"alg":"Ed25519"}]}`,
		encodeB64URL(pub))
}

func signTestJWT(priv ed25519.PrivateKey, typ string, claims map[string]any) string {
	header, _ := json.Marshal(map[string]string{"alg": "Ed25519", "typ": typ, "kid": "k1"})
	payload, _ := json.Marshal(claims)
	input := encodeB64URL(header) + "." + encodeB64URL(payload)
	return input + "." + encodeB64URL(ed25519.Sign(priv, []byte(input)))
}

// agentToken mints an agent token from the test agent provider for the given key.
func (s *ServiceTestSuite) agentToken(providerPriv ed25519.PrivateKey, sub, parent string, key ed25519.PublicKey,
	exp int64) string {
	claims := map[string]any{"iss": "https://agent.example", "dwk": dwkAgent, "sub": sub, "jti": randomToken(8),
		"cnf": map[string]any{"jwk": map[string]any{"kty": "OKP", "crv": "Ed25519", "x": encodeB64URL(key),
			"alg": "Ed25519"}}, "iat": time.Now().Unix(), "exp": exp, "ps": testIssuer}
	if parent != "" {
		claims["parent_agent"] = parent
	}
	return signTestJWT(providerPriv, typAgentToken, claims)
}

// signedRequest builds an agent request to the person server signed with key.
func signedAgentRequest(method, path, body, agentToken string, key ed25519.PrivateKey) *http.Request {
	r := httptest.NewRequest(method, testIssuer+path, strings.NewReader(body))
	r.Host = "ps.example"
	r.Header.Set(headerSignatureKey, `sig=jwt;jwt="`+agentToken+`"`)
	components := []string{"@method", "@authority", "@path", "signature-key"}
	if body != "" {
		r.Header.Set("Content-Type", contentTypeJSON)
		r.Header.Set(headerContentDigest, "sha-256=:"+base64.StdEncoding.EncodeToString(sha256Sum(body))+":")
		components = append(components, "content-type", "content-digest")
	}
	quoted := make([]string, 0, len(components))
	var base strings.Builder
	for _, c := range components {
		quoted = append(quoted, `"`+c+`"`)
		v, _ := componentValue(r, c)
		base.WriteString(`"` + c + `": ` + v + "\n")
	}
	params := fmt.Sprintf("(%s);created=%d", strings.Join(quoted, " "), time.Now().Unix())
	base.WriteString(`"@signature-params": ` + params)
	r.Header.Set(headerSignatureInput, "sig="+params)
	signature := ed25519.Sign(key, []byte(base.String()))
	r.Header.Set(headerSignature, "sig=:"+base64.StdEncoding.EncodeToString(signature)+":")
	return r
}

func (s *ServiceTestSuite) TestAuthenticate() {
	_, providerPriv, _ := ed25519.GenerateKey(rand.Reader)
	s.publishIssuer("https://agent.example", dwkAgent, providerPriv)
	exp := time.Now().Add(time.Hour).Unix()
	token := s.agentToken(providerPriv, "aauth:bot@agent.example", "", s.agentPub, exp)

	agent, body, sigErr := s.service.authenticate(signedAgentRequest(http.MethodPost, pathPersonToken, `{"a":1}`, token,
		s.agentPriv))
	s.Require().Nil(sigErr)
	s.Equal("aauth:bot@agent.example", agent.subject)
	s.Equal("Issuer", agent.name)
	s.Equal(`{"a":1}`, string(body))

	_, otherKey, _ := ed25519.GenerateKey(rand.Reader)
	_, _, sigErr = s.service.authenticate(signedAgentRequest(http.MethodPost, pathPersonToken, `{}`, token, otherKey))
	s.Equal(sigErrInvalidSignature, sigErr.code)

	expired := s.agentToken(providerPriv, "aauth:bot@agent.example", "", s.agentPub, time.Now().Unix()-1)
	_, _, sigErr = s.service.authenticate(signedAgentRequest(http.MethodGet, "/x", "", expired, s.agentPriv))
	s.Equal(sigErrExpiredJWT, sigErr.code)

	_, rogue, _ := ed25519.GenerateKey(rand.Reader)
	forged := s.agentToken(rogue, "aauth:bot@agent.example", "", s.agentPub, exp)
	_, _, sigErr = s.service.authenticate(signedAgentRequest(http.MethodGet, "/x", "", forged, s.agentPriv))
	s.Equal(sigErrInvalidJWT, sigErr.code)

	notAgent := signTestJWT(providerPriv, typPersonToken, map[string]any{"exp": exp})
	_, _, sigErr = s.service.authenticate(signedAgentRequest(http.MethodGet, "/x", "", notAgent, s.agentPriv))
	s.Equal(sigErrInvalidJWT, sigErr.code)

	r := signedAgentRequest(http.MethodGet, "/x", "", token, s.agentPriv)
	r.Header.Set(headerSignatureKey, `sig=hwk;kty="OKP"`)
	_, _, sigErr = s.service.authenticate(r)
	s.Equal(sigErrUnsupportedScheme, sigErr.code)
}

func (s *ServiceTestSuite) TestAuthTokenFlowWithSubagent() {
	_, providerPriv, _ := ed25519.GenerateKey(rand.Reader)
	s.publishIssuer("https://agent.example", dwkAgent, providerPriv)
	_, resourcePriv, _ := ed25519.GenerateKey(rand.Reader)
	s.publishIssuer(testResource, dwkResource, resourcePriv)

	// Bind the agent and approve the resource.
	outcome, _ := s.service.requestPersonToken(context.Background(), s.agent, s.personRequest())
	s.approve(outcome.pending, testUser)
	_, _ = s.service.poll(context.Background(), s.agent, outcome.pending.ID)

	subPub, _, _ := ed25519.GenerateKey(rand.Reader)
	subToken := s.agentToken(providerPriv, "aauth:bot+sub@agent.example", s.agent.subject, subPub,
		time.Now().Add(time.Hour).Unix())
	req := s.personRequest()
	req.SubagentToken = subToken
	ptOutcome, perr := s.service.requestPersonToken(context.Background(), s.agent, req)
	s.Require().Nil(perr)
	pt := ptOutcome.result.(map[string]any)["person_token"].(string)
	ptClaims := s.decode(pt, typPersonToken)
	subJWK := map[string]any{"kty": "OKP", "crv": "Ed25519", "x": encodeB64URL(subPub)}
	subJKT, _ := jktOf(subJWK)

	resourceToken := func(jkt, presentedJTI string) string {
		return signTestJWT(resourcePriv, typResourceToken, map[string]any{
			"iss": testResource, "dwk": dwkResource, "aud": testIssuer, "jti": randomToken(8), "ps": testIssuer,
			"sub": ptClaims["sub"], "presented_jti": presentedJTI, "agent_jkt": jkt, "scope": "files.read",
			"iat": time.Now().Unix(), "exp": time.Now().Add(5 * time.Minute).Unix()})
	}
	atReq := &authTokenRequest{ResourceToken: resourceToken(subJKT, ptClaims["jti"].(string)), PresentedToken: pt,
		SubagentToken: subToken, Capabilities: []string{"interaction"}}
	atOutcome, perr := s.service.requestAuthToken(context.Background(), s.agent, atReq)
	s.Require().Nil(perr)
	s.Require().NotNil(atOutcome.pending, "new scopes need consent")
	s.approve(atOutcome.pending, testUser)
	polled, perr := s.service.poll(context.Background(), s.agent, atOutcome.pending.ID)
	s.Require().Nil(perr)
	body := map[string]any{}
	s.Require().NoError(json.Unmarshal(polled.result.(json.RawMessage), &body))
	at := s.decode(body["auth_token"].(string), typAuthToken)
	s.Equal(testIssuer, at["ps"])
	s.Equal("files.read", at["scope"])
	s.Equal(encodeB64URL(subPub), at["cnf"].(map[string]any)["jwk"].(map[string]any)["x"])

	again, perr := s.service.requestAuthToken(context.Background(), s.agent, atReq)
	s.Require().Nil(perr)
	s.Nil(again.pending, "consent is remembered")

	for name, bad := range map[string]*authTokenRequest{
		"wrong key": {ResourceToken: resourceToken("nope", ptClaims["jti"].(string)), PresentedToken: pt,
			SubagentToken: subToken},
		"no presented": {ResourceToken: atReq.ResourceToken},
		"garbage rt":   {ResourceToken: "a.b.c", PresentedToken: pt},
		"upstream":     {ResourceToken: atReq.ResourceToken, PresentedToken: pt, UpstreamToken: "x"},
		"foreign subagent": {ResourceToken: atReq.ResourceToken, PresentedToken: pt,
			SubagentToken: s.agentToken(providerPriv, "aauth:x@agent.example", "aauth:other@agent.example", subPub,
				time.Now().Add(time.Hour).Unix())},
		"expired subagent": {ResourceToken: atReq.ResourceToken, PresentedToken: pt,
			SubagentToken: s.agentToken(providerPriv, "aauth:bot+sub@agent.example", s.agent.subject, subPub,
				time.Now().Unix()-1)},
	} {
		_, perr := s.service.requestAuthToken(context.Background(), s.agent, bad)
		s.NotNil(perr, name)
	}
}

func (s *ServiceTestSuite) TestHandlersEndToEnd() {
	_, providerPriv, _ := ed25519.GenerateKey(rand.Reader)
	s.publishIssuer("https://agent.example", dwkAgent, providerPriv)
	token := s.agentToken(providerPriv, "aauth:bot@agent.example", "", s.agentPub, time.Now().Add(time.Hour).Unix())
	mux := http.NewServeMux()
	registerRoutes(mux, newAAuthHandler(s.service))

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest(http.MethodGet, pathPersonMetadata, nil))
	s.Equal(http.StatusOK, w.Code)
	s.Contains(w.Body.String(), `"person_token_endpoint"`)

	w = httptest.NewRecorder()
	mux.ServeHTTP(w, signedAgentRequest(http.MethodPost, pathPersonToken,
		`{"resource":"https://resource.example","capabilities":["interaction"]}`, token, s.agentPriv))
	s.Equal(http.StatusAccepted, w.Code)
	location := w.Header().Get("Location")
	code := strings.Split(w.Header().Get("AAuth-Requirement"), `code="`)[1]
	code = strings.TrimSuffix(code, `"`)

	w = httptest.NewRecorder()
	pendingPath := strings.TrimPrefix(location, testIssuer)
	mux.ServeHTTP(w, signedAgentRequest(http.MethodGet, pendingPath, "", token, s.agentPriv))
	s.Equal(http.StatusAccepted, w.Code)

	w = httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest(http.MethodGet, pathInteraction, nil))
	s.Contains(w.Body.String(), "Code from your agent")
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest(http.MethodGet, pathInteraction+"?code=ZZZZ-ZZZZ", nil))
	s.Equal(http.StatusGone, w.Code)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest(http.MethodGet, pathInteraction+"?code="+code, nil))
	s.Equal(http.StatusFound, w.Code)
	authorize, _ := url.Parse(w.Header().Get("Location"))

	id := strings.TrimPrefix(location, testIssuer+pathPending+"/")
	stored, _ := s.service.store.getPending(context.Background(), id)
	s.redeemer.EXPECT().GetAuthorizationCodeDetails(mock.Anything, "AAUTH_PS", "c").Return(
		&oauth2authz.AuthorizationCode{RedirectURI: testIssuer + "/aauth/interaction/callback",
			CodeChallenge: pkceChallenge(stored.PKCEVerifier), AuthorizedUserID: testUser}, nil).Once()
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest(http.MethodGet,
		pathInteraction+"/callback?code=c&state="+url.QueryEscape(authorize.Query().Get("state")), nil))
	s.Equal(http.StatusOK, w.Code)
	cookie := w.Result().Cookies()[0]
	form := url.Values{"pending_id": {id}, "decision_token": {cookie.Value}, "decision": {"approve"}}
	r := httptest.NewRequest(http.MethodPost, pathInteraction+"/decision", strings.NewReader(form.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.AddCookie(cookie)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, r)
	s.Contains(w.Body.String(), "Approved")

	w = httptest.NewRecorder()
	mux.ServeHTTP(w, signedAgentRequest(http.MethodGet, pendingPath, "", token, s.agentPriv))
	s.Equal(http.StatusOK, w.Code)
	s.Contains(w.Body.String(), `"person_token"`)

	w = httptest.NewRecorder()
	mux.ServeHTTP(w, signedAgentRequest(http.MethodPost, pathMission, `{"description":""}`, token, s.agentPriv))
	s.Equal(http.StatusBadRequest, w.Code)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, signedAgentRequest(http.MethodPost, pathAuthToken, `{}`, token, s.agentPriv))
	s.Equal(http.StatusBadRequest, w.Code)
	w = httptest.NewRecorder()
	missionPath := pathMission + "/" + strings.Repeat("a", 43)
	mux.ServeHTTP(w, signedAgentRequest(http.MethodPost, missionPath, `{"action":"update"}`, token, s.agentPriv))
	s.Equal(http.StatusNotFound, w.Code)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest(http.MethodPost, pathPersonToken, strings.NewReader("{}")))
	s.Equal(http.StatusUnauthorized, w.Code)
}
