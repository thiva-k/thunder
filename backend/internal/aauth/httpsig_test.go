// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

package aauth

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type HTTPSigTestSuite struct {
	suite.Suite
	priv ed25519.PrivateKey
	pub  ed25519.PublicKey
	now  time.Time
}

func TestHTTPSigTestSuite(t *testing.T) {
	suite.Run(t, new(HTTPSigTestSuite))
}

func (s *HTTPSigTestSuite) SetupTest() {
	s.pub, s.priv, _ = ed25519.GenerateKey(rand.Reader)
	s.now = time.Unix(1_800_000_000, 0)
}

// signedRequest builds a request signed over components, the way an AAuth agent does.
func (s *HTTPSigTestSuite) signedRequest(method, target, body string, components []string, created int64,
	tamper func(*http.Request)) *http.Request {
	r := httptest.NewRequest(method, target, strings.NewReader(body))
	r.Host = "ps.example"
	r.Header.Set(headerSignatureKey, `sig=jwt;jwt="eyJ.x.y"`)
	if body != "" {
		r.Header.Set("Content-Type", contentTypeJSON)
		sum := sha256.Sum256([]byte(body))
		r.Header.Set(headerContentDigest, "sha-256=:"+base64.StdEncoding.EncodeToString(sum[:])+":")
	}
	quoted := make([]string, 0, len(components))
	for _, c := range components {
		quoted = append(quoted, `"`+c+`"`)
	}
	params := fmt.Sprintf("(%s);created=%d", strings.Join(quoted, " "), created)
	var base strings.Builder
	for _, c := range components {
		value, _ := componentValue(r, c)
		base.WriteString(`"` + c + `": ` + value + "\n")
	}
	base.WriteString(`"@signature-params": ` + params)
	sig := ed25519.Sign(s.priv, []byte(base.String()))
	r.Header.Set(headerSignatureInput, "sig="+params)
	r.Header.Set(headerSignature, "sig=:"+base64.StdEncoding.EncodeToString(sig)+":")
	if tamper != nil {
		tamper(r)
	}
	return r
}

var bodyComponents = []string{"@method", "@authority", "@path", "content-type", "signature-key", "content-digest"}

func (s *HTTPSigTestSuite) TestValidSignatureBaseVerifies() {
	r := s.signedRequest(http.MethodPost, "https://ps.example/aauth/person", `{"resource":"https://r.example"}`,
		bodyComponents, s.now.Unix(), nil)
	signed, sigErr := parseSignedRequest(r, s.now)
	s.Require().Nil(sigErr)
	s.True(ed25519.Verify(s.pub, signed.base, signed.signature))
	s.Equal(`{"resource":"https://r.example"}`, string(signed.body))
	s.Equal(sfToken("jwt"), signed.keyMember.item.value)
}

func (s *HTTPSigTestSuite) TestGETWithoutBodyNeedsNoDigest() {
	r := s.signedRequest(http.MethodGet, "https://ps.example/aauth/pending/abc", "",
		[]string{"@method", "@authority", "@path", "signature-key"}, s.now.Unix(), nil)
	_, sigErr := parseSignedRequest(r, s.now)
	s.Nil(sigErr)
}

func (s *HTTPSigTestSuite) TestFailures() {
	cases := []struct {
		name       string
		components []string
		created    int64
		tamper     func(*http.Request)
		code       string
	}{
		{"missing headers", bodyComponents, s.now.Unix(), func(r *http.Request) { r.Header.Del(headerSignature) },
			sigErrInvalidSignature},
		{"body not covered", []string{"@method", "@authority", "@path", "signature-key"}, s.now.Unix(), nil,
			sigErrInvalidInput},
		{"too old", bodyComponents, s.now.Unix() - 120, nil, sigErrInvalidSignature},
		{"from the future", bodyComponents, s.now.Unix() + 120, nil, sigErrClockSkew},
		{"digest mismatch", bodyComponents, s.now.Unix(), func(r *http.Request) {
			r.Header.Set(headerContentDigest, "sha-256=:"+base64.StdEncoding.EncodeToString(make([]byte, 32))+":")
		}, sigErrInvalidSignature},
		{"no Signature-Key member for the label", bodyComponents, s.now.Unix(), func(r *http.Request) {
			r.Header.Set(headerSignatureKey, `other=jwt;jwt="x"`)
		}, sigErrInvalidSignature},
		{"malformed Signature-Input", bodyComponents, s.now.Unix(), func(r *http.Request) {
			r.Header.Set(headerSignatureInput, "sig=(")
		}, sigErrInvalidSignature},
	}
	for _, tc := range cases {
		r := s.signedRequest(http.MethodPost, "https://ps.example/aauth/person", `{"a":1}`, tc.components, tc.created,
			tc.tamper)
		_, sigErr := parseSignedRequest(r, s.now)
		s.Require().NotNil(sigErr, tc.name)
		s.Equal(tc.code, sigErr.code, tc.name)
	}
}

func (s *HTTPSigTestSuite) TestTamperedBodyFailsVerification() {
	r := s.signedRequest(http.MethodPost, "https://ps.example/aauth/person", `{"a":1}`, bodyComponents, s.now.Unix(),
		func(r *http.Request) { r.Header.Set("Content-Type", "text/plain") })
	signed, sigErr := parseSignedRequest(r, s.now)
	s.Require().Nil(sigErr)
	s.False(ed25519.Verify(s.pub, signed.base, signed.signature))
}

func (s *HTTPSigTestSuite) TestComponentValues() {
	r := httptest.NewRequest(http.MethodGet, "https://PS.Example:443/a%20b?x=1", nil)
	r.Header.Add("X-Multi", " one ")
	r.Header.Add("X-Multi", "two")
	for name, want := range map[string]string{
		"@method": "GET", "@authority": "ps.example", "@path": "/a%20b", "@query": "?x=1", "@scheme": "https",
		"@target-uri": "https://ps.example/a%20b?x=1", "x-multi": "one, two",
	} {
		got, ok := componentValue(r, name)
		s.True(ok, name)
		s.Equal(want, got, name)
	}
	_, ok := componentValue(r, "@unknown")
	s.False(ok)
	_, ok = componentValue(r, "absent-header")
	s.False(ok)
}
