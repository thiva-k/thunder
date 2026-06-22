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
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"
)

type OpenID4VCIHandlerTestSuite struct {
	suite.Suite
}

func TestOpenID4VCIHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(OpenID4VCIHandlerTestSuite))
}

func makeToken(t *testing.T, payload map[string]any) string {
	t.Helper()
	b, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	return "e30." + base64.RawURLEncoding.EncodeToString(b) + ".sig"
}

func (s *OpenID4VCIHandlerTestSuite) TestVerifyDPoPBearerTokenSkipped() {
	h := &openID4VCIHandler{}
	token := makeToken(s.T(), map[string]any{"sub": "u1"})
	req := httptest.NewRequest(http.MethodPost, "/openid4vci/credential", nil)
	s.NoError(h.verifyDPoP(req, token), "bearer (unbound) token should skip DPoP")
}

func (s *OpenID4VCIHandlerTestSuite) TestVerifyDPoPBoundTokenRequiresProof() {
	h := &openID4VCIHandler{}
	token := makeToken(s.T(), map[string]any{"sub": "u1", "cnf": map[string]any{"jkt": "abc"}})
	req := httptest.NewRequest(http.MethodPost, "/openid4vci/credential", nil)
	s.ErrorIs(h.verifyDPoP(req, token), ErrInvalidDPoP, "DPoP-bound token without proof should fail")
}
