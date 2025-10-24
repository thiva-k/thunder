/*
 * Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
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

package discovery

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type DiscoveryTestSuite struct {
	suite.Suite
	discoveryService DiscoveryServiceInterface
	handler          DiscoveryHandlerInterface
}

func TestDiscoverySuite(t *testing.T) {
	suite.Run(t, new(DiscoveryTestSuite))
}

func (suite *DiscoveryTestSuite) SetupTest() {
	suite.discoveryService = NewDiscoveryService()
	suite.handler = NewDiscoveryHandler(suite.discoveryService)
}

func (suite *DiscoveryTestSuite) TestOAuth2AuthorizationServerMetadata() {
	req := httptest.NewRequest("GET", "/.well-known/oauth-authorization-server", nil)
	w := httptest.NewRecorder()
	
	suite.handler.HandleOAuth2AuthorizationServerMetadata(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	assert.Equal(suite.T(), "application/json", w.Header().Get("Content-Type"))

	var metadata OAuth2AuthorizationServerMetadata
	err := json.NewDecoder(w.Body).Decode(&metadata)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), metadata.Issuer)
	assert.NotEmpty(suite.T(), metadata.AuthorizationEndpoint)
	assert.NotEmpty(suite.T(), metadata.TokenEndpoint)
	assert.NotEmpty(suite.T(), metadata.JWKSUri)
	assert.NotEmpty(suite.T(), metadata.IntrospectionEndpoint)

	// Verify only implemented endpoints are present
	assert.Empty(suite.T(), metadata.UserInfoEndpoint)   // Not implemented
	assert.Empty(suite.T(), metadata.RevocationEndpoint) // Not implemented

	// Verify only implemented grant types are present
	assert.Contains(suite.T(), metadata.GrantTypesSupported, "authorization_code")
	assert.Contains(suite.T(), metadata.GrantTypesSupported, "client_credentials")
	assert.Contains(suite.T(), metadata.GrantTypesSupported, "refresh_token")
	assert.NotContains(suite.T(), metadata.GrantTypesSupported, "password") // Not implemented
	assert.NotContains(suite.T(), metadata.GrantTypesSupported, "implicit") // Not implemented

	// Verify only implemented response types are present
	assert.Equal(suite.T(), []string{"code"}, metadata.ResponseTypesSupported)
}

func (suite *DiscoveryTestSuite) TestOIDCDiscovery() {
	req := httptest.NewRequest("GET", "/.well-known/openid-configuration", nil)
	w := httptest.NewRecorder()

	suite.handler.HandleOIDCDiscovery(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	assert.Equal(suite.T(), "application/json", w.Header().Get("Content-Type"))

	var metadata OIDCProviderMetadata
	err := json.NewDecoder(w.Body).Decode(&metadata)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), metadata.Issuer)
	assert.NotEmpty(suite.T(), metadata.SubjectTypesSupported)
	assert.NotEmpty(suite.T(), metadata.ClaimsSupported)
	assert.NotEmpty(suite.T(), metadata.IDTokenSigningAlgValuesSupported)

	// Verify OIDC-specific fields
	assert.Contains(suite.T(), metadata.SubjectTypesSupported, "public")
	assert.Contains(suite.T(), metadata.IDTokenSigningAlgValuesSupported, "RS256")
	assert.Contains(suite.T(), metadata.ClaimsSupported, "sub")
	assert.Contains(suite.T(), metadata.ClaimsSupported, "iss")
	assert.Contains(suite.T(), metadata.ClaimsSupported, "aud")
}

func (suite *DiscoveryTestSuite) TestMethodNotAllowed() {
	req := httptest.NewRequest("POST", "/.well-known/oauth-authorization-server", nil)
	w := httptest.NewRecorder()
	
	suite.handler.HandleOAuth2AuthorizationServerMetadata(w, req)

	assert.Equal(suite.T(), http.StatusMethodNotAllowed, w.Code)
}
