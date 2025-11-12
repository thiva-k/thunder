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

package token

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	applicationmodel "github.com/asgardeo/thunder/internal/application/model"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/clientauth"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/tests/mocks/applicationmock"
	"github.com/asgardeo/thunder/tests/mocks/jwtmock"
	"github.com/asgardeo/thunder/tests/mocks/oauth/oauth2/granthandlersmock"
	"github.com/asgardeo/thunder/tests/mocks/oauth/scopemock"
	"github.com/asgardeo/thunder/tests/mocks/usermock"
)

type TokenHandlerTestSuite struct {
	suite.Suite
	mockJWTService     *jwtmock.JWTServiceInterfaceMock
	mockUserService    *usermock.UserServiceInterfaceMock
	mockAppService     *applicationmock.ApplicationServiceInterfaceMock
	mockGrantProvider  *granthandlersmock.GrantHandlerProviderInterfaceMock
	mockScopeValidator *scopemock.ScopeValidatorInterfaceMock
	mockGrantHandler   *granthandlersmock.GrantHandlerInterfaceMock
}

func TestTokenHandlerSuite(t *testing.T) {
	suite.Run(t, new(TokenHandlerTestSuite))
}

func (suite *TokenHandlerTestSuite) SetupTest() {
	// Initialize Thunder Runtime config with basic test config
	testConfig := &config.Config{
		JWT: config.JWTConfig{
			ValidityPeriod: 3600,
		},
	}
	_ = config.InitializeThunderRuntime("test", testConfig)
	suite.mockJWTService = &jwtmock.JWTServiceInterfaceMock{}
	suite.mockUserService = usermock.NewUserServiceInterfaceMock(suite.T())
	suite.mockGrantProvider = granthandlersmock.NewGrantHandlerProviderInterfaceMock(suite.T())
	suite.mockAppService = applicationmock.NewApplicationServiceInterfaceMock(suite.T())
	suite.mockScopeValidator = scopemock.NewScopeValidatorInterfaceMock(suite.T())
	suite.mockGrantHandler = granthandlersmock.NewGrantHandlerInterfaceMock(suite.T())

	// Setup common mock for GetGrantHandler that can be used across tests
	// Using Maybe() allows tests to override this if needed
	suite.mockGrantProvider.On("GetGrantHandler", constants.GrantTypeAuthorizationCode).
		Return(suite.mockGrantHandler, nil).Maybe()
}

func (suite *TokenHandlerTestSuite) TestnewTokenHandler() {
	handler := newTokenHandler(suite.mockAppService, suite.mockGrantProvider, suite.mockScopeValidator)
	assert.NotNil(suite.T(), handler)
	assert.Implements(suite.T(), (*TokenHandlerInterface)(nil), handler)
}

func (suite *TokenHandlerTestSuite) TestHandleTokenRequest_InvalidFormData() {
	handler := newTokenHandler(suite.mockAppService, suite.mockGrantProvider, suite.mockScopeValidator)
	req, _ := http.NewRequest("POST", "/token", strings.NewReader("invalid-form-data%"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()

	handler.HandleTokenRequest(rr, req)

	assert.Equal(suite.T(), http.StatusBadRequest, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "invalid_request", response["error"])
	assert.Equal(suite.T(), "Failed to parse request body", response["error_description"])
}

func (suite *TokenHandlerTestSuite) TestHandleTokenRequest_MissingGrantType() {
	formData := url.Values{}
	formData.Set("client_id", "test-client-id")
	formData.Set("client_secret", "test-secret")

	suite.testTokenRequestError(formData, http.StatusBadRequest, "invalid_request",
		"Missing grant_type parameter")
}

// Helper function to test token request error scenarios
func (suite *TokenHandlerTestSuite) testTokenRequestError(formData url.Values,
	expectedStatusCode int, expectedError, expectedErrorDescription string) {
	handler := newTokenHandler(suite.mockAppService, suite.mockGrantProvider, suite.mockScopeValidator)

	req, _ := http.NewRequest("POST", "/token", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Set up OAuth client in context (simulating middleware)
	mockApp := &applicationmodel.OAuthAppConfigProcessedDTO{
		ClientID:                "test-client-id",
		HashedClientSecret:      "hashed-secret",
		TokenEndpointAuthMethod: constants.TokenEndpointAuthMethodClientSecretPost,
		GrantTypes:              []constants.GrantType{constants.GrantTypeAuthorizationCode},
	}
	clientInfo := &clientauth.OAuthClientInfo{
		ClientID:     "test-client-id",
		ClientSecret: "test-secret",
		OAuthApp:     mockApp,
	}
	ctx := context.WithValue(req.Context(), clientauth.OAuthClientKey, clientInfo)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()

	handler.HandleTokenRequest(rr, req)

	assert.Equal(suite.T(), expectedStatusCode, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedError, response["error"])
	assert.Equal(suite.T(), expectedErrorDescription, response["error_description"])
}

func (suite *TokenHandlerTestSuite) TestHandleTokenRequest_InvalidGrantType() {
	formData := url.Values{}
	formData.Set("grant_type", "invalid_grant")
	formData.Set("client_id", "test-client-id")
	formData.Set("client_secret", "test-secret")

	suite.testTokenRequestError(formData, http.StatusBadRequest, "unsupported_grant_type",
		"Invalid grant_type parameter")
}

func (suite *TokenHandlerTestSuite) TestHandleTokenRequest_MissingClientID() {
	handler := newTokenHandler(suite.mockAppService, suite.mockGrantProvider, suite.mockScopeValidator)
	formData := url.Values{}
	formData.Set("grant_type", "authorization_code")

	req, _ := http.NewRequest("POST", "/token", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()

	handler.HandleTokenRequest(rr, req)

	// Handler should return server error when context is missing
	assert.Equal(suite.T(), http.StatusInternalServerError, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "server_error", response["error"])
}
