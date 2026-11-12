// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

package aauth

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/thunder-id/thunderid/tests/mocks/httpmock"
)

type JOSETestSuite struct {
	suite.Suite
}

func TestJOSETestSuite(t *testing.T) {
	suite.Run(t, new(JOSETestSuite))
}

func (s *JOSETestSuite) TestServerIdentifiers() {
	for v, want := range map[string]bool{
		"https://ps.example": true, "https://xn--nxasmq6b.example": true,
		"http://ps.example": false, "https://PS.example": false, "https://ps.example:8443": false,
		"https://ps.example/": false, "https://ps.example/v1": false, "https://ps.example?a=1": false,
		"https://user@ps.example": false, "": false,
	} {
		s.Equal(want, isServerIdentifier(v), v)
	}
}

func (s *JOSETestSuite) TestAgentIdentifiers() {
	for v, want := range map[string]bool{
		"aauth:assistant-v2@agent.example": true, "aauth:Planner.7f3c+search1@vendor.example": true,
		"aauth:@agent.example": false, "aauth:my agent@agent.example": false, "agent@agent.example": false,
		"aauth:a@http://agent.example": false, "aauth:a@Agent.example": false,
	} {
		s.Equal(want, isAgentIdentifier(v), v)
	}
}

func (s *JOSETestSuite) TestAlgorithms() {
	alg, err := cryptoAlgorithm("Ed25519")
	s.NoError(err)
	s.Equal("EdDSA", alg)
	for _, bad := range []string{"EdDSA", "HS256", "none", ""} {
		_, err := cryptoAlgorithm(bad)
		s.ErrorIs(err, errUnsupportedAlg, bad)
	}
}

func (s *JOSETestSuite) TestCheckJWKForAlg() {
	cases := []struct {
		jwk  map[string]any
		want error
	}{
		{map[string]any{"kty": "OKP", "crv": "Ed25519", "x": "x", "alg": "Ed25519"}, nil},
		{map[string]any{"kty": "EC", "crv": "P-256", "x": "x", "y": "y", "alg": "ES256"}, nil},
		{map[string]any{"kty": "OKP", "crv": "Ed25519", "x": "x"}, errUnsupportedAlg},
		{map[string]any{"kty": "OKP", "crv": "Ed25519", "x": "x", "alg": "EdDSA"}, errUnsupportedAlg},
		{map[string]any{"kty": "EC", "crv": "P-384", "x": "x", "y": "y", "alg": "ES256"}, errKeyAlgMismatch},
		{map[string]any{"kty": "OKP", "crv": "Ed25519", "x": "x", "d": "secret", "alg": "Ed25519"}, errKeyAlgMismatch},
	}
	for _, tc := range cases {
		_, err := checkJWKForAlg(tc.jwk)
		if tc.want == nil {
			s.NoError(err)
		} else {
			s.ErrorIs(err, tc.want)
		}
	}
}

func (s *JOSETestSuite) TestParseJWT() {
	p, err := parseJWT("eyJhbGciOiJFZDI1NTE5IiwidHlwIjoiYWEtYWdlbnQrand0In0." +
		"eyJpc3MiOiJodHRwczovL2EuZXhhbXBsZSIsImV4cCI6MTAsImNuZiI6eyJqd2siOnsia3R5IjoiT0tQIn19fQ.c2ln")
	s.Require().NoError(err)
	s.Equal("aa-agent+jwt", p.headerString("typ"))
	s.Equal("https://a.example", p.claimString("iss"))
	s.Equal(int64(10), p.exp())
	s.Equal("OKP", p.cnfJWK()["kty"])
	for _, bad := range []string{"a.b", "!!.e30.e30", "e30.!!.e30", "e30.e30.!!", "bm90anNvbg.e30.e30"} {
		_, err := parseJWT(bad)
		s.ErrorIs(err, errMalformedJWT, bad)
	}
}

func jsonResponse(body string) *http.Response {
	return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)),
		Header: http.Header{"Content-Type": {"application/json"}}}
}

func (s *JOSETestSuite) TestIssuerKeyFetcher() {
	client := httpmock.NewHTTPClientInterfaceMock(s.T())
	client.EXPECT().Do(mock.Anything).RunAndReturn(func(r *http.Request) (*http.Response, error) {
		switch r.URL.String() {
		case "https://good.example/.well-known/aauth-agent.json":
			return jsonResponse(
				`{"issuer":"https://good.example","jwks_uri":"https://good.example/jwks","name":"Good"}`), nil
		case "https://good.example/jwks":
			return jsonResponse(`{"keys":[{"kid":"k1","kty":"OKP","crv":"Ed25519","x":"x","alg":"Ed25519"}]}`), nil
		case "https://liar.example/.well-known/aauth-agent.json":
			return jsonResponse(`{"issuer":"https://good.example","jwks_uri":"https://good.example/jwks"}`), nil
		case "https://anon.example/.well-known/aauth-agent.json":
			return jsonResponse(`{"jwks_uri":"https://anon.example/jwks"}`), nil
		}
		return &http.Response{StatusCode: http.StatusNotFound, Body: io.NopCloser(strings.NewReader(""))}, nil
	})
	fetcher := newIssuerKeyFetcher(client)
	key, entry, err := fetcher.key(context.Background(), "https://good.example", dwkAgent, "k1")
	s.Require().NoError(err)
	s.Equal("k1", key["kid"])
	s.Equal("Good", entry.metadata.Name)

	_, _, err = fetcher.key(context.Background(), "https://good.example", dwkAgent, "missing")
	s.ErrorIs(err, errUnknownKey)
	_, err = fetcher.metadata(context.Background(), "https://liar.example", dwkAgent, false)
	s.ErrorIs(err, errIssuerMismatch)
	_, err = fetcher.metadata(context.Background(), "https://anon.example", dwkAgent, false)
	s.ErrorIs(err, errIssuerMissing)
	_, err = fetcher.metadata(context.Background(), "https://gone.example", dwkAgent, false)
	s.ErrorIs(err, errMetadataFetch)
	_, err = fetcher.metadata(context.Background(), "https://good.example:8443", dwkAgent, false)
	s.ErrorIs(err, errInvalidIdentifier)
}
