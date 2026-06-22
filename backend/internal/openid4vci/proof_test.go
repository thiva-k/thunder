/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 *
 * WSO2 LLC. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package openid4vci

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/thunder-id/thunderid/internal/system/cryptolib"
)

const testIssuer = "https://issuer.example"

type ProofTestSuite struct {
	suite.Suite
}

func TestProofTestSuite(t *testing.T) {
	suite.Run(t, new(ProofTestSuite))
}

// signProofJWT builds a signed OpenID4VCI holder proof JWT (typ
// openid4vci-proof+jwt) carrying key's public JWK in the header and the given
// audience/nonce/iat in the payload, signed ES256 in JWS P1363 form.
func signProofJWT(t *testing.T, key *ecdsa.PrivateKey, aud, nonce string, iat time.Time) string {
	t.Helper()
	jwk := map[string]interface{}{
		"kty": "EC",
		"crv": "P-256",
		"x":   base64.RawURLEncoding.EncodeToString(key.PublicKey.X.FillBytes(make([]byte, 32))),
		"y":   base64.RawURLEncoding.EncodeToString(key.PublicKey.Y.FillBytes(make([]byte, 32))),
	}
	header := map[string]interface{}{"alg": "ES256", "typ": proofType, "jwk": jwk}
	payload := map[string]interface{}{"aud": aud, "nonce": nonce, "iat": iat.Unix()}

	hb, _ := json.Marshal(header)
	pb, _ := json.Marshal(payload)
	signingInput := base64.RawURLEncoding.EncodeToString(hb) + "." + base64.RawURLEncoding.EncodeToString(pb)

	p1363, err := cryptolib.Generate([]byte(signingInput), cryptolib.ECDSASHA256, key)
	if err != nil {
		t.Fatalf("sign proof: %v", err)
	}

	return signingInput + "." + base64.RawURLEncoding.EncodeToString(p1363)
}

func newTestService(t *testing.T, store nonceStore) *service {
	t.Helper()
	return &service{
		cfg:    serviceConfig{CredentialIssuer: testIssuer, ProofMaxAge: time.Minute, BatchSize: defaultBatchSize},
		nonces: store,
	}
}

// A batch of proofs (one per holder key) sharing a single c_nonce must yield one
// confirmation JWK per proof, and consume the shared nonce exactly once.
func (s *ProofTestSuite) TestVerifyProofsBatchConsumesSharedNonceOnce() {
	ctx := context.Background()
	store := newInMemoryNonceStore()
	nonce := "shared-nonce"
	s.Require().NoError(store.Save(ctx, &nonceRecord{Nonce: nonce, ExpiresAt: time.Now().Add(time.Minute)}))
	svc := newTestService(s.T(), store)

	proofs := make([]Proof, 0, 3)
	for i := 0; i < 3; i++ {
		key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		s.Require().NoError(err)
		proofs = append(proofs, Proof{
			ProofType: "jwt",
			JWT:       signProofJWT(s.T(), key, testIssuer, nonce, time.Now()),
		})
	}

	jwks, err := svc.verifyProofs(ctx, proofs)
	s.Require().NoError(err)
	s.Require().Len(jwks, 3)
	for i, jwk := range jwks {
		_, ok := jwk["x"].(string)
		s.True(ok, "proof %d: confirmation JWK missing x coordinate", i)
	}
	_, ok := store.Get(ctx, nonce)
	s.False(ok, "shared c_nonce should have been consumed")
}

// An unknown c_nonce is rejected for the whole batch.
func (s *ProofTestSuite) TestVerifyProofsRejectsUnknownNonce() {
	ctx := context.Background()
	svc := newTestService(s.T(), newInMemoryNonceStore())

	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	proofs := []Proof{{ProofType: "jwt", JWT: signProofJWT(s.T(), key, testIssuer, "never-issued", time.Now())}}

	_, err := svc.verifyProofs(ctx, proofs)
	s.Error(err, "expected error for unknown c_nonce")
}

func (s *ProofTestSuite) TestHolderProofsPrefersBatchThenSingle() {
	tests := []struct {
		name string
		req  CredentialRequest
		want int
	}{
		{"batch", CredentialRequest{Proofs: &Proofs{JWT: []string{"a", "b", "c"}}}, 3},
		{"single", CredentialRequest{Proof: Proof{ProofType: "jwt", JWT: "a"}}, 1},
		{"batch over single", CredentialRequest{
			Proof:  Proof{ProofType: "jwt", JWT: "single"},
			Proofs: &Proofs{JWT: []string{"a", "b"}},
		}, 2},
		{"empty", CredentialRequest{}, 0},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.Equal(tc.want, len(tc.req.holderProofs()))
		})
	}
}
