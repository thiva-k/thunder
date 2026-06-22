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

package openid4vp

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type OpenID4VPStoreTestSuite struct {
	suite.Suite
}

func TestOpenID4VPStoreTestSuite(t *testing.T) {
	suite.Run(t, new(OpenID4VPStoreTestSuite))
}

// identityCrypto is a no-op ConfigCryptoProvider for tests: it exercises the
// marshal/encrypt/decrypt/parse path without a real symmetric key.
type identityCrypto struct{}

func (identityCrypto) Encrypt(_ context.Context, content []byte) ([]byte, error) { return content, nil }
func (identityCrypto) Decrypt(_ context.Context, content []byte) ([]byte, error) { return content, nil }

// The runtime store read path reconstructs a RequestState from a result row,
// decrypting the ephemeral key and decoding the verification result.
func (suite *OpenID4VPStoreTestSuite) TestDBStateStoreReadPathRoundTrip() {
	crypto := identityCrypto{}
	store := &dbStateStore{deploymentID: "test", crypto: crypto}

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	suite.Require().NoError(err)
	pkcs8, err := x509.MarshalPKCS8PrivateKey(key)
	suite.Require().NoError(err)
	encKey, err := crypto.Encrypt(context.Background(), pkcs8)
	suite.Require().NoError(err)

	vp := &VerifiedPresentation{
		Subject: "sub-1", VCT: "urn:eudi:pid:de:1",
		Claims: map[string]interface{}{"given_name": "Erika"},
	}
	resultJSON, err := json.Marshal(vp)
	suite.Require().NoError(err)

	expiry := time.Now().Add(time.Minute).UTC()
	row := map[string]interface{}{
		"state":          "state-1",
		"definition_id":  "eudi-pid",
		"nonce":          "nonce-1",
		"ephemeral_key":  encKey,
		"client_id":      "x509_hash:abc",
		"rp_id":          "rp-1",
		"request_uri":    "https://verifier.example/openid4vp/request?state=state-1",
		"status":         "COMPLETED",
		"result":         resultJSON,
		"failure_reason": "",
		"expiry_time":    expiry,
	}

	rs, err := store.buildRequestStateFromRow(context.Background(), row)
	suite.Require().NoError(err)
	suite.Equal("state-1", rs.State)
	suite.Equal("eudi-pid", rs.DefinitionID)
	suite.Equal("nonce-1", rs.Nonce)
	suite.Equal(StatusCompleted, rs.Status)
	suite.Require().NotNil(rs.EphemeralKey)
	suite.True(key.Equal(rs.EphemeralKey))
	suite.Require().NotNil(rs.Result)
	suite.Equal("Erika", rs.Result.Claims["given_name"])
	suite.WithinDuration(expiry, rs.ExpiresAt, time.Second)
}

// A row with no ephemeral key and no result yields a state with nil fields.
func (suite *OpenID4VPStoreTestSuite) TestDBStateStoreReadPathPending() {
	store := &dbStateStore{deploymentID: "test", crypto: identityCrypto{}}
	row := map[string]interface{}{
		"state":       "state-2",
		"status":      "PENDING",
		"expiry_time": time.Now().Add(time.Minute).UTC(),
	}
	rs, err := store.buildRequestStateFromRow(context.Background(), row)
	suite.Require().NoError(err)
	suite.Equal(StatusPending, rs.Status)
	suite.Nil(rs.EphemeralKey)
	suite.Nil(rs.Result)
}

// parseStateTime handles Postgres time.Time and SQLite datetime strings.
func (suite *OpenID4VPStoreTestSuite) TestParseStateTime() {
	now := time.Now().UTC().Truncate(time.Second)

	got, err := parseStateTime(now)
	suite.Require().NoError(err)
	suite.True(now.Equal(got))

	got, err = parseStateTime(now.Format(time.RFC3339))
	suite.Require().NoError(err)
	suite.True(now.Equal(got))

	got, err = parseStateTime(now.Format("2006-01-02 15:04:05"))
	suite.Require().NoError(err)
	suite.Equal(now.Format("2006-01-02 15:04:05"), got.Format("2006-01-02 15:04:05"))

	_, err = parseStateTime([]byte(now.Format(time.RFC3339)))
	suite.Require().NoError(err)

	_, err = parseStateTime(123)
	suite.Require().Error(err)
}

func (suite *OpenID4VPStoreTestSuite) TestInMemoryStateStoreRoundTrip() {
	store := newInMemoryStateStore()
	suite.Require().NotNil(store)

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	suite.Require().NoError(err)
	st := &RequestState{
		State:        "state-mem",
		Nonce:        "nonce-mem",
		EphemeralKey: key,
		Status:       StatusPending,
		ExpiresAt:    time.Now().Add(time.Minute),
	}

	suite.Require().NoError(store.Save(context.Background(), st))

	got, ok := store.Get(context.Background(), "state-mem")
	suite.Require().True(ok)
	suite.Equal("nonce-mem", got.Nonce)
	suite.Equal(StatusPending, got.Status)
	suite.True(key.Equal(got.EphemeralKey))

	_, ok = store.Get(context.Background(), "missing")
	suite.False(ok)

	suite.Require().NoError(store.Delete(context.Background(), "state-mem"))
	_, ok = store.Get(context.Background(), "state-mem")
	suite.False(ok)

	// Deleting a missing entry is a no-op.
	suite.Require().NoError(store.Delete(context.Background(), "still-missing"))
}
