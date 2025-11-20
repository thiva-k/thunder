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

package authz

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	appmodel "github.com/asgardeo/thunder/internal/application/model"
	flowcm "github.com/asgardeo/thunder/internal/flow/common"
	"github.com/asgardeo/thunder/internal/flow/flowexec"
	"github.com/asgardeo/thunder/tests/mocks/applicationmock"
	"github.com/asgardeo/thunder/tests/mocks/flow/flowexecmock"
	"github.com/asgardeo/thunder/tests/mocks/jwtmock"

	oauth2const "github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	oauth2model "github.com/asgardeo/thunder/internal/oauth/oauth2/model"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
)

type AuthorizeHandlerTestSuite struct {
	suite.Suite
	handler             *authorizeHandler
	mockAppService      *applicationmock.ApplicationServiceInterfaceMock
	mockJWTService      *jwtmock.JWTServiceInterfaceMock
	mockAuthzCodeStore  *AuthorizationCodeStoreInterfaceMock
	mockFlowExecService *flowexecmock.FlowExecServiceInterfaceMock
}

func TestAuthorizeHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(AuthorizeHandlerTestSuite))
}

func (suite *AuthorizeHandlerTestSuite) SetupTest() {
	// Initialize Thunder Runtime config with basic test config
	testConfig := &config.Config{
		GateClient: config.GateClientConfig{
			Scheme:    "https",
			Hostname:  "localhost",
			Port:      3000,
			LoginPath: "/login",
			ErrorPath: "/error",
		},
	}
	_ = config.InitializeThunderRuntime("test", testConfig)

	// Create mocked dependencies for testing
	suite.mockAppService = applicationmock.NewApplicationServiceInterfaceMock(suite.T())
	suite.mockJWTService = jwtmock.NewJWTServiceInterfaceMock(suite.T())
	suite.mockAuthzCodeStore = NewAuthorizationCodeStoreInterfaceMock(suite.T())
	suite.mockFlowExecService = flowexecmock.NewFlowExecServiceInterfaceMock(suite.T())

	suite.handler = newAuthorizeHandler(
		suite.mockAppService, suite.mockJWTService, suite.mockAuthzCodeStore,
		suite.mockFlowExecService).(*authorizeHandler)
}

func (suite *AuthorizeHandlerTestSuite) TestnewAuthorizeHandler() {
	mockStore := NewAuthorizationCodeStoreInterfaceMock(suite.T())
	mockFlowExec := flowexecmock.NewFlowExecServiceInterfaceMock(suite.T())
	handler := newAuthorizeHandler(suite.mockAppService, suite.mockJWTService, mockStore, mockFlowExec)
	assert.NotNil(suite.T(), handler)
	assert.Implements(suite.T(), (*AuthorizeHandlerInterface)(nil), handler)
}

func (suite *AuthorizeHandlerTestSuite) TestGetOAuthMessageForGetRequest_Success() {
	req := httptest.NewRequest(http.MethodGet, "/auth?client_id=test-client&redirect_uri=https://example.com", nil)

	msg, err := suite.handler.getOAuthMessageForGetRequest(req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), msg)
	if msg != nil {
		assert.Equal(suite.T(), oauth2const.TypeInitialAuthorizationRequest, msg.RequestType)
		assert.Equal(suite.T(), "test-client", msg.RequestQueryParams["client_id"])
		assert.Equal(suite.T(), "https://example.com", msg.RequestQueryParams["redirect_uri"])
		assert.Empty(suite.T(), msg.SessionDataKey)
	}
}

func (suite *AuthorizeHandlerTestSuite) TestGetOAuthMessageForGetRequest_ParseFormError() {
	// Create a malformed URL to trigger ParseForm error
	req := httptest.NewRequest(http.MethodGet, "/auth?client_id=%ZZ", nil)

	msg, err := suite.handler.getOAuthMessageForGetRequest(req)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), msg)
	assert.Contains(suite.T(), err.Error(), "failed to parse form data")
}

func (suite *AuthorizeHandlerTestSuite) TestGetOAuthMessageForPostRequest_MissingSessionDataKey() {
	postData := AuthZPostRequest{
		SessionDataKey: "", // Missing session data key
		Assertion:      "test-assertion",
	}
	jsonData, _ := json.Marshal(postData)

	req := httptest.NewRequest(http.MethodPost, "/auth", bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")

	msg, err := suite.handler.getOAuthMessageForPostRequest(req)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), msg)
	assert.Contains(suite.T(), err.Error(), "sessionDataKey or assertion is missing")
}

func (suite *AuthorizeHandlerTestSuite) TestGetOAuthMessageForPostRequest_MissingAssertion() {
	postData := AuthZPostRequest{
		SessionDataKey: "test-session-key",
		Assertion:      "", // Missing assertion
	}
	jsonData, _ := json.Marshal(postData)

	req := httptest.NewRequest(http.MethodPost, "/auth", bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")

	msg, err := suite.handler.getOAuthMessageForPostRequest(req)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), msg)
	assert.Contains(suite.T(), err.Error(), "sessionDataKey or assertion is missing")
}

func (suite *AuthorizeHandlerTestSuite) TestGetOAuthMessage_UnsupportedMethod() {
	req := httptest.NewRequest(http.MethodPatch, "/auth", nil)
	rr := httptest.NewRecorder()

	msg := suite.handler.getOAuthMessage(req, rr)

	assert.Nil(suite.T(), msg)
	assert.Equal(suite.T(), http.StatusBadRequest, rr.Code)
}

func (suite *AuthorizeHandlerTestSuite) TestGetOAuthMessage_NilRequest() {
	rr := httptest.NewRecorder()

	msg := suite.handler.getOAuthMessage(nil, rr)

	assert.Nil(suite.T(), msg)
}

func (suite *AuthorizeHandlerTestSuite) TestGetOAuthMessage_NilResponseWriter() {
	req := httptest.NewRequest(http.MethodGet, "/auth", nil)

	msg := suite.handler.getOAuthMessage(req, nil)

	assert.Nil(suite.T(), msg)
}

func (suite *AuthorizeHandlerTestSuite) TestGetAuthorizationCode_Success() {
	// Create a valid OAuth message with session data
	sessionData := &SessionData{
		OAuthParameters: oauth2model.OAuthParameters{
			ClientID:         "test-client",
			RedirectURI:      "https://client.example.com/callback",
			StandardScopes:   []string{"openid", "profile"},
			PermissionScopes: []string{"read", "write"},
		},
		AuthTime: time.Now(),
	}

	assertionClaims := &assertionClaims{userID: "test-user"}

	result, err := createAuthorizationCode(sessionData, assertionClaims)

	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), result.CodeID)
	assert.NotEmpty(suite.T(), result.Code)
	assert.Equal(suite.T(), "test-client", result.ClientID)
	assert.Equal(suite.T(), "https://client.example.com/callback", result.RedirectURI)
	assert.Equal(suite.T(), "test-user", result.AuthorizedUserID)
	assert.Equal(suite.T(), "openid profile read write", result.Scopes)
	assert.Equal(suite.T(), AuthCodeStateActive, result.State)
	assert.NotZero(suite.T(), result.TimeCreated)
	assert.True(suite.T(), result.ExpiryTime.After(result.TimeCreated))
}

func (suite *AuthorizeHandlerTestSuite) TestGetAuthorizationCode_MissingClientID() {
	sessionData := &SessionData{
		OAuthParameters: oauth2model.OAuthParameters{
			ClientID:    "", // Empty client ID
			RedirectURI: "https://client.example.com/callback",
		},
		AuthTime: time.Now(),
	}

	assertionClaims := &assertionClaims{
		userID: "test-user",
	}

	result, err := createAuthorizationCode(sessionData, assertionClaims)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "client_id or redirect_uri is missing")
	assert.Equal(suite.T(), AuthorizationCode{}, result)
}

func (suite *AuthorizeHandlerTestSuite) TestGetAuthorizationCode_MissingRedirectURI() {
	sessionData := &SessionData{
		OAuthParameters: oauth2model.OAuthParameters{
			ClientID:    "test-client",
			RedirectURI: "", // Missing redirect URI
		},
		AuthTime: time.Now(),
	}

	assertionClaims := &assertionClaims{
		userID: "test-user",
	}

	result, err := createAuthorizationCode(sessionData, assertionClaims)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "client_id or redirect_uri is missing")
	assert.Equal(suite.T(), AuthorizationCode{}, result)
}

func (suite *AuthorizeHandlerTestSuite) TestGetAuthorizationCode_EmptyUserID() {
	sessionData := &SessionData{
		OAuthParameters: oauth2model.OAuthParameters{
			ClientID:    "test-client-id",
			RedirectURI: "https://client.example.com/callback",
		},
		AuthTime: time.Now(),
	}

	assertionClaims := &assertionClaims{
		userID: "", // Empty user ID
	}

	result, err := createAuthorizationCode(sessionData, assertionClaims)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "authenticated user not found")
	assert.Equal(suite.T(), AuthorizationCode{}, result)
}

func (suite *AuthorizeHandlerTestSuite) TestGetAuthorizationCode_ZeroAuthTime() {
	sessionData := &SessionData{
		OAuthParameters: oauth2model.OAuthParameters{
			ClientID:    "test-client-id",
			RedirectURI: "https://client.example.com/callback",
		},
		AuthTime: time.Time{}, // Zero auth time
	}

	assertionClaims := &assertionClaims{
		userID: "test-user",
	}

	result, err := createAuthorizationCode(sessionData, assertionClaims)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "authentication time is not set")
	assert.Equal(suite.T(), AuthorizationCode{}, result)
}

func (suite *AuthorizeHandlerTestSuite) TestGetLoginPageRedirectURI_Success() {
	queryParams := map[string]string{
		"sessionDataKey": "test-key",
		"appId":          "test-app",
	}

	redirectURI, err := getLoginPageRedirectURI(queryParams)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), redirectURI, "sessionDataKey=test-key")
	assert.Contains(suite.T(), redirectURI, "appId=test-app")
}

func (suite *AuthorizeHandlerTestSuite) TestGetErrorPageRedirectURL_Success() {
	redirectURI, err := getErrorPageRedirectURL("invalid_request", "Missing parameter")
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), redirectURI, "errorCode=invalid_request")
	assert.Contains(suite.T(), redirectURI, "errorMessage=Missing+parameter")
}

// Helper function to create a valid OAuth application for testing
func (suite *AuthorizeHandlerTestSuite) createTestOAuthApp() *appmodel.OAuthAppConfigProcessedDTO {
	return &appmodel.OAuthAppConfigProcessedDTO{
		AppID:         "test-app-id",
		ClientID:      "test-client-id",
		RedirectURIs:  []string{"https://client.example.com/callback"},
		GrantTypes:    []oauth2const.GrantType{oauth2const.GrantTypeAuthorizationCode},
		ResponseTypes: []oauth2const.ResponseType{oauth2const.ResponseTypeCode},
		PKCERequired:  false, // Disable PKCE to simplify test
	}
}

// Helper function to create a test OAuth message
func (suite *AuthorizeHandlerTestSuite) createTestOAuthMessage() *OAuthMessage {
	return &OAuthMessage{
		RequestType: oauth2const.TypeInitialAuthorizationRequest,
		RequestQueryParams: map[string]string{
			"client_id":     "test-client-id",
			"redirect_uri":  "https://client.example.com/callback",
			"response_type": "code",
			"scope":         "read write",
			"state":         "test-state",
		},
	}
}

func (suite *AuthorizeHandlerTestSuite) TestHandleInitialAuthorizationRequest_InitiateFlowSuccess() {
	// Create a valid OAuth application with proper grant types and response types
	app := suite.createTestOAuthApp()
	suite.mockAppService.EXPECT().GetOAuthApplication("test-client-id").Return(app, nil)

	// Mock flow exec service to return success
	expectedFlowInitCtx := &flowexec.FlowInitContext{
		ApplicationID: "test-app-id",
		FlowType:      string(flowcm.FlowTypeAuthentication),
		RuntimeData: map[string]string{
			"requested_permissions": "read write",
		},
	}
	suite.mockFlowExecService.EXPECT().InitiateFlow(expectedFlowInitCtx).Return("test-session-key", nil)

	// Create OAuth message for initial authorization request
	msg := suite.createTestOAuthMessage()

	// Create HTTP request and response recorder
	req := httptest.NewRequest(http.MethodGet, "/auth", nil)
	rr := httptest.NewRecorder()

	// Execute the method under test
	suite.handler.handleInitialAuthorizationRequest(msg, rr, req)

	// Assert that it redirects to login page
	assert.Equal(suite.T(), http.StatusFound, rr.Code)

	// Check the redirect location contains login page and flow information
	location := rr.Header().Get("Location")
	assert.Contains(suite.T(), location, "/login")
	assert.Contains(suite.T(), location, "flowId=test-session-key")
	assert.Contains(suite.T(), location, "sessionDataKey=")
}

func (suite *AuthorizeHandlerTestSuite) TestHandleInitialAuthorizationRequest_InitiateFlowError() {
	// Create a valid OAuth application with proper grant types and response types
	app := suite.createTestOAuthApp()
	suite.mockAppService.EXPECT().GetOAuthApplication("test-client-id").Return(app, nil)

	// Mock flow exec service to return an error
	expectedFlowInitCtx := &flowexec.FlowInitContext{
		ApplicationID: "test-app-id",
		FlowType:      string(flowcm.FlowTypeAuthentication),
		RuntimeData: map[string]string{
			"requested_permissions": "read write",
		},
	}
	mockError := &serviceerror.InternalServerError
	suite.mockFlowExecService.EXPECT().InitiateFlow(expectedFlowInitCtx).Return("", mockError)

	// Create OAuth message for initial authorization request
	msg := suite.createTestOAuthMessage()

	// Create HTTP request and response recorder
	req := httptest.NewRequest(http.MethodGet, "/auth", nil)
	rr := httptest.NewRecorder()

	// Execute the method under test
	suite.handler.handleInitialAuthorizationRequest(msg, rr, req)

	// Assert that it redirects to error page
	assert.Equal(suite.T(), http.StatusFound, rr.Code)

	// Check the redirect location contains error information
	location := rr.Header().Get("Location")
	assert.Contains(suite.T(), location, "/error")
	assert.Contains(suite.T(), location, "errorCode=server_error")
	assert.Contains(suite.T(), location, "errorMessage=Failed+to+initiate+authentication+flow")
}

func (suite *AuthorizeHandlerTestSuite) TestHandleInitialAuthorizationRequest_WithOIDCAndNonOIDCScopes() {
	// Create a valid OAuth application
	app := suite.createTestOAuthApp()
	suite.mockAppService.EXPECT().GetOAuthApplication("test-client-id").Return(app, nil)

	// Mock flow exec service - only non-OIDC scopes should be in RuntimeData
	expectedFlowInitCtx := &flowexec.FlowInitContext{
		ApplicationID: "test-app-id",
		FlowType:      string(flowcm.FlowTypeAuthentication),
		RuntimeData: map[string]string{
			"requested_permissions": "read write", // Only non-OIDC scopes
		},
	}
	suite.mockFlowExecService.EXPECT().InitiateFlow(expectedFlowInitCtx).Return("test-session-key", nil)

	// Create OAuth message with both OIDC (openid, profile) and non-OIDC scopes (read, write)
	msg := &OAuthMessage{
		RequestType: oauth2const.TypeInitialAuthorizationRequest,
		RequestQueryParams: map[string]string{
			"client_id":     "test-client-id",
			"redirect_uri":  "https://client.example.com/callback",
			"response_type": "code",
			"scope":         "openid profile read write", // Mixed scopes
			"state":         "test-state",
		},
	}

	// Create HTTP request and response recorder
	req := httptest.NewRequest(http.MethodGet, "/auth", nil)
	rr := httptest.NewRecorder()

	// Execute the method under test
	suite.handler.handleInitialAuthorizationRequest(msg, rr, req)

	// Assert that it redirects to login page
	assert.Equal(suite.T(), http.StatusFound, rr.Code)

	// Check the redirect location
	location := rr.Header().Get("Location")
	assert.Contains(suite.T(), location, "/login")
	assert.Contains(suite.T(), location, "flowId=test-session-key")
}

func (suite *AuthorizeHandlerTestSuite) TestHandleInitialAuthorizationRequest_OnlyOIDCScopes() {
	// Create a valid OAuth application
	app := suite.createTestOAuthApp()
	suite.mockAppService.EXPECT().GetOAuthApplication("test-client-id").Return(app, nil)

	// Mock flow exec service - empty RuntimeData since no non-OIDC scopes
	expectedFlowInitCtx := &flowexec.FlowInitContext{
		ApplicationID: "test-app-id",
		FlowType:      string(flowcm.FlowTypeAuthentication),
		RuntimeData: map[string]string{
			"requested_permissions": "", // Empty, only OIDC scopes
		},
	}
	suite.mockFlowExecService.EXPECT().InitiateFlow(expectedFlowInitCtx).Return("test-session-key", nil)

	// Create OAuth message with only OIDC scopes
	msg := &OAuthMessage{
		RequestType: oauth2const.TypeInitialAuthorizationRequest,
		RequestQueryParams: map[string]string{
			"client_id":     "test-client-id",
			"redirect_uri":  "https://client.example.com/callback",
			"response_type": "code",
			"scope":         "openid profile email", // Only OIDC scopes
			"state":         "test-state",
		},
	}

	// Create HTTP request and response recorder
	req := httptest.NewRequest(http.MethodGet, "/auth", nil)
	rr := httptest.NewRecorder()

	// Execute the method under test
	suite.handler.handleInitialAuthorizationRequest(msg, rr, req)

	// Assert that it redirects to login page
	assert.Equal(suite.T(), http.StatusFound, rr.Code)
}

func (suite *AuthorizeHandlerTestSuite) TestHandleAuthorizePostRequest_ConsentType() {
	postData := AuthZPostRequest{
		SessionDataKey: "test-key",
		Assertion:      "test-assertion",
	}
	jsonData, _ := json.Marshal(postData)

	// Mock to make it look like consent response (unhandled type)
	req := httptest.NewRequest(http.MethodPost, "/auth", bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	// This will fail to decode or will be treated as invalid
	suite.handler.HandleAuthorizePostRequest(rr, req)

	// Should return some response (either error or redirect)
	assert.NotEqual(suite.T(), 0, rr.Code)
}

func (suite *AuthorizeHandlerTestSuite) TestHandleAuthorizePostRequest_InvalidRequestType() {
	req := httptest.NewRequest(http.MethodPost, "/auth", nil)
	rr := httptest.NewRecorder()

	suite.handler.HandleAuthorizePostRequest(rr, req)

	assert.Equal(suite.T(), http.StatusBadRequest, rr.Code)
}

func (suite *AuthorizeHandlerTestSuite) TestHandleAuthorizationResponseFromEngine_InvalidSessionDataKey() {
	msg := &OAuthMessage{
		RequestType:    oauth2const.TypeAuthorizationResponseFromEngine,
		SessionDataKey: "invalid-key",
		RequestBodyParams: map[string]string{
			oauth2const.Assertion: "test-assertion",
		},
	}

	rr := httptest.NewRecorder()
	suite.handler.handleAuthorizationResponseFromEngine(msg, rr)

	assert.Equal(suite.T(), http.StatusOK, rr.Code)
	var resp AuthZPostResponse
	err := json.NewDecoder(rr.Body).Decode(&resp)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), resp.RedirectURI, "/error")
}

func (suite *AuthorizeHandlerTestSuite) TestHandleAuthorizationResponseFromEngine_MissingAssertion() {
	sessionData := SessionData{
		OAuthParameters: oauth2model.OAuthParameters{
			ClientID:    "test-client",
			RedirectURI: "https://client.example.com/callback",
		},
		AuthTime: time.Now(),
	}
	sessionKey := suite.handler.sessionStore.AddSession(sessionData)

	msg := &OAuthMessage{
		RequestType:       oauth2const.TypeAuthorizationResponseFromEngine,
		SessionDataKey:    sessionKey,
		RequestBodyParams: map[string]string{},
	}

	rr := httptest.NewRecorder()
	suite.handler.handleAuthorizationResponseFromEngine(msg, rr)

	assert.Equal(suite.T(), http.StatusOK, rr.Code)
}

func (suite *AuthorizeHandlerTestSuite) TestHandleAuthorizationResponseFromEngine_InvalidAssertionSignature() {
	sessionData := SessionData{
		OAuthParameters: oauth2model.OAuthParameters{
			ClientID:    "test-client",
			RedirectURI: "https://client.example.com/callback",
		},
		AuthTime: time.Now(),
	}
	sessionKey := suite.handler.sessionStore.AddSession(sessionData)

	assertion := "invalid.jwt.token"
	suite.mockJWTService.EXPECT().VerifyJWT(assertion, "", "").Return(errors.New("invalid signature"))

	msg := &OAuthMessage{
		RequestType:    oauth2const.TypeAuthorizationResponseFromEngine,
		SessionDataKey: sessionKey,
		RequestBodyParams: map[string]string{
			oauth2const.Assertion: assertion,
		},
	}

	rr := httptest.NewRecorder()
	suite.handler.handleAuthorizationResponseFromEngine(msg, rr)

	assert.Equal(suite.T(), http.StatusOK, rr.Code)
}

func (suite *AuthorizeHandlerTestSuite) TestHandleAuthorizationResponseFromEngine_FailedToDecodeAssertion() {
	sessionData := SessionData{
		OAuthParameters: oauth2model.OAuthParameters{
			ClientID:    "test-client",
			RedirectURI: "https://client.example.com/callback",
		},
		AuthTime: time.Now(),
	}
	sessionKey := suite.handler.sessionStore.AddSession(sessionData)

	assertion := "invalid-jwt-format"
	suite.mockJWTService.EXPECT().VerifyJWT(assertion, "", "").Return(nil)

	msg := &OAuthMessage{
		RequestType:    oauth2const.TypeAuthorizationResponseFromEngine,
		SessionDataKey: sessionKey,
		RequestBodyParams: map[string]string{
			oauth2const.Assertion: assertion,
		},
	}

	rr := httptest.NewRecorder()
	suite.handler.handleAuthorizationResponseFromEngine(msg, rr)

	assert.Equal(suite.T(), http.StatusOK, rr.Code)
}

func (suite *AuthorizeHandlerTestSuite) TestHandleAuthorizationResponseFromEngine_EmptyUserID() {
	sessionData := SessionData{
		OAuthParameters: oauth2model.OAuthParameters{
			ClientID:    "test-client",
			RedirectURI: "https://client.example.com/callback",
		},
		AuthTime: time.Now(),
	}
	sessionKey := suite.handler.sessionStore.AddSession(sessionData)

	// This will fail during decode since it's not a valid JWT
	assertion := "not.a.valid.jwt"
	suite.mockJWTService.EXPECT().VerifyJWT(assertion, "", "").Return(nil)

	msg := &OAuthMessage{
		RequestType:    oauth2const.TypeAuthorizationResponseFromEngine,
		SessionDataKey: sessionKey,
		RequestBodyParams: map[string]string{
			oauth2const.Assertion: assertion,
		},
	}

	rr := httptest.NewRecorder()
	suite.handler.handleAuthorizationResponseFromEngine(msg, rr)

	assert.Equal(suite.T(), http.StatusOK, rr.Code)
	var resp AuthZPostResponse
	err := json.NewDecoder(rr.Body).Decode(&resp)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), resp.RedirectURI, "/error")
}

func (suite *AuthorizeHandlerTestSuite) TestRedirectToLoginPage_NilResponseWriter() {
	req := httptest.NewRequest(http.MethodGet, "/auth", nil)
	queryParams := map[string]string{"sessionDataKey": "test-key"}

	suite.handler.redirectToLoginPage(nil, req, queryParams)
	// Should not panic and should log error
}

func (suite *AuthorizeHandlerTestSuite) TestRedirectToLoginPage_NilRequest() {
	rr := httptest.NewRecorder()
	queryParams := map[string]string{"sessionDataKey": "test-key"}

	suite.handler.redirectToLoginPage(rr, nil, queryParams)
	// Should not panic and should log error
}

func (suite *AuthorizeHandlerTestSuite) TestRedirectToErrorPage_NilResponseWriter() {
	req := httptest.NewRequest(http.MethodGet, "/auth", nil)

	suite.handler.redirectToErrorPage(nil, req, "error_code", "error message")
	// Should not panic and should log error
}

func (suite *AuthorizeHandlerTestSuite) TestRedirectToErrorPage_NilRequest() {
	rr := httptest.NewRecorder()

	suite.handler.redirectToErrorPage(rr, nil, "error_code", "error message")
	// Should not panic and should log error
}

func (suite *AuthorizeHandlerTestSuite) TestWriteAuthZResponseToErrorPage_WithState() {
	sessionData := &SessionData{
		OAuthParameters: oauth2model.OAuthParameters{
			State: "test-state",
		},
	}

	rr := httptest.NewRecorder()
	suite.handler.writeAuthZResponseToErrorPage(rr, "error_code", "error message", sessionData)

	assert.Equal(suite.T(), http.StatusOK, rr.Code)
	var resp AuthZPostResponse
	err := json.NewDecoder(rr.Body).Decode(&resp)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), resp.RedirectURI, "state=test-state")
}

func (suite *AuthorizeHandlerTestSuite) TestWriteAuthZResponseToErrorPage_NilSessionData() {
	rr := httptest.NewRecorder()
	suite.handler.writeAuthZResponseToErrorPage(rr, "error_code", "error message", nil)

	assert.Equal(suite.T(), http.StatusOK, rr.Code)
	var resp AuthZPostResponse
	err := json.NewDecoder(rr.Body).Decode(&resp)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), resp.RedirectURI)
}

func (suite *AuthorizeHandlerTestSuite) TestHandleInitialAuthorizationRequest_MissingClientID() {
	msg := &OAuthMessage{
		RequestType: oauth2const.TypeInitialAuthorizationRequest,
		RequestQueryParams: map[string]string{
			"redirect_uri":  "https://client.example.com/callback",
			"response_type": "code",
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/auth", nil)
	rr := httptest.NewRecorder()

	suite.handler.handleInitialAuthorizationRequest(msg, rr, req)

	assert.Equal(suite.T(), http.StatusFound, rr.Code)
	location := rr.Header().Get("Location")
	assert.Contains(suite.T(), location, "/error")
}

func (suite *AuthorizeHandlerTestSuite) TestHandleInitialAuthorizationRequest_InvalidClient() {
	suite.mockAppService.EXPECT().GetOAuthApplication("invalid-client").Return(nil, &serviceerror.ServiceError{
		Code: "CLIENT_NOT_FOUND",
	})

	msg := &OAuthMessage{
		RequestType: oauth2const.TypeInitialAuthorizationRequest,
		RequestQueryParams: map[string]string{
			"client_id":     "invalid-client",
			"redirect_uri":  "https://client.example.com/callback",
			"response_type": "code",
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/auth", nil)
	rr := httptest.NewRecorder()

	suite.handler.handleInitialAuthorizationRequest(msg, rr, req)

	assert.Equal(suite.T(), http.StatusFound, rr.Code)
	location := rr.Header().Get("Location")
	assert.Contains(suite.T(), location, "/error")
}

func (suite *AuthorizeHandlerTestSuite) TestHandleInitialAuthorizationRequest_ValidationError_RedirectToApp() {
	app := suite.createTestOAuthApp()
	suite.mockAppService.EXPECT().GetOAuthApplication("test-client-id").Return(app, nil)

	msg := &OAuthMessage{
		RequestType: oauth2const.TypeInitialAuthorizationRequest,
		RequestQueryParams: map[string]string{
			"client_id":     "test-client-id",
			"redirect_uri":  "https://client.example.com/callback",
			"response_type": "invalid_type", // Invalid response type
			"state":         "test-state",
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/auth", nil)
	rr := httptest.NewRecorder()

	suite.handler.handleInitialAuthorizationRequest(msg, rr, req)

	assert.Equal(suite.T(), http.StatusFound, rr.Code)
}

func (suite *AuthorizeHandlerTestSuite) TestHandleInitialAuthorizationRequest_InsecureRedirectURI() {
	app := suite.createTestOAuthApp()
	app.RedirectURIs = []string{"http://client.example.com/callback"} // HTTP instead of HTTPS
	suite.mockAppService.EXPECT().GetOAuthApplication("test-client-id").Return(app, nil)

	expectedFlowInitCtx := &flowexec.FlowInitContext{
		ApplicationID: "test-app-id",
		FlowType:      string(flowcm.FlowTypeAuthentication),
		RuntimeData: map[string]string{
			"requested_permissions": "read write",
		},
	}
	suite.mockFlowExecService.EXPECT().InitiateFlow(expectedFlowInitCtx).Return("test-flow-id", nil)

	msg := &OAuthMessage{
		RequestType: oauth2const.TypeInitialAuthorizationRequest,
		RequestQueryParams: map[string]string{
			"client_id":     "test-client-id",
			"redirect_uri":  "http://client.example.com/callback",
			"response_type": "code",
			"scope":         "read write",
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/auth", nil)
	rr := httptest.NewRecorder()

	suite.handler.handleInitialAuthorizationRequest(msg, rr, req)

	assert.Equal(suite.T(), http.StatusFound, rr.Code)
	location := rr.Header().Get("Location")
	assert.Contains(suite.T(), location, "showInsecureWarning=true")
}

func (suite *AuthorizeHandlerTestSuite) TestHandleInitialAuthorizationRequest_EmptyRedirectURI() {
	app := suite.createTestOAuthApp()
	suite.mockAppService.EXPECT().GetOAuthApplication("test-client-id").Return(app, nil)

	expectedFlowInitCtx := &flowexec.FlowInitContext{
		ApplicationID: "test-app-id",
		FlowType:      string(flowcm.FlowTypeAuthentication),
		RuntimeData: map[string]string{
			"requested_permissions": "read write",
		},
	}
	suite.mockFlowExecService.EXPECT().InitiateFlow(expectedFlowInitCtx).Return("test-flow-id", nil)

	msg := &OAuthMessage{
		RequestType: oauth2const.TypeInitialAuthorizationRequest,
		RequestQueryParams: map[string]string{
			"client_id":     "test-client-id",
			"response_type": "code",
			"scope":         "read write",
			// redirect_uri is empty, should use app's default
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/auth", nil)
	rr := httptest.NewRecorder()

	suite.handler.handleInitialAuthorizationRequest(msg, rr, req)

	assert.Equal(suite.T(), http.StatusFound, rr.Code)
}
