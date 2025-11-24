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

package executor

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	appmodel "github.com/asgardeo/thunder/internal/application/model"
	authncm "github.com/asgardeo/thunder/internal/authn/common"
	authnoauth "github.com/asgardeo/thunder/internal/authn/oauth"
	flowcm "github.com/asgardeo/thunder/internal/flow/common"
	flowcore "github.com/asgardeo/thunder/internal/flow/core"
	"github.com/asgardeo/thunder/internal/idp"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/user"
	"github.com/asgardeo/thunder/internal/userschema"
	"github.com/asgardeo/thunder/tests/mocks/authn/oidcmock"
	"github.com/asgardeo/thunder/tests/mocks/flow/coremock"
	"github.com/asgardeo/thunder/tests/mocks/idp/idpmock"
	"github.com/asgardeo/thunder/tests/mocks/usermock"
	"github.com/asgardeo/thunder/tests/mocks/userschemamock"
)

type OIDCAuthExecutorTestSuite struct {
	suite.Suite
	mockOIDCService       *oidcmock.OIDCAuthnCoreServiceInterfaceMock
	mockIDPService        *idpmock.IDPServiceInterfaceMock
	mockFlowFactory       *coremock.FlowFactoryInterfaceMock
	mockUserService       *usermock.UserServiceInterfaceMock
	mockUserSchemaService *userschemamock.UserSchemaServiceInterfaceMock
	executor              oidcAuthExecutorInterface
}

func TestOIDCAuthExecutorSuite(t *testing.T) {
	suite.Run(t, new(OIDCAuthExecutorTestSuite))
}

func (suite *OIDCAuthExecutorTestSuite) SetupTest() {
	suite.mockOIDCService = oidcmock.NewOIDCAuthnCoreServiceInterfaceMock(suite.T())
	suite.mockIDPService = idpmock.NewIDPServiceInterfaceMock(suite.T())
	suite.mockFlowFactory = coremock.NewFlowFactoryInterfaceMock(suite.T())
	suite.mockUserService = usermock.NewUserServiceInterfaceMock(suite.T())
	suite.mockUserSchemaService = userschemamock.NewUserSchemaServiceInterfaceMock(suite.T())

	defaultInputs := []flowcm.InputData{{Name: "code", Type: "string", Required: true}}
	mockExec := createMockAuthExecutor(suite.T(), ExecutorNameOIDCAuth)
	suite.mockFlowFactory.On("CreateExecutor", ExecutorNameOIDCAuth, flowcm.ExecutorTypeAuthentication,
		defaultInputs, []flowcm.InputData{}).Return(mockExec)

	suite.executor = newOIDCAuthExecutor(ExecutorNameOIDCAuth, defaultInputs, []flowcm.InputData{},
		suite.mockFlowFactory, suite.mockIDPService, suite.mockOIDCService,
		suite.mockUserService, suite.mockUserSchemaService)
}

func (suite *OIDCAuthExecutorTestSuite) TestNewOIDCAuthExecutor() {
	assert.NotNil(suite.T(), suite.executor)
}

func (suite *OIDCAuthExecutorTestSuite) TestExecute_CodeNotProvided_BuildsAuthorizeURL() {
	ctx := &flowcore.NodeContext{
		FlowID:        "flow-123",
		FlowType:      flowcm.FlowTypeAuthentication,
		UserInputData: map[string]string{},
		NodeInputData: []flowcm.InputData{{Name: "code", Type: "string", Required: true}},
		NodeProperties: map[string]interface{}{
			"idpId": "idp-123",
		},
	}

	suite.mockOIDCService.On("BuildAuthorizeURL", "idp-123").
		Return("https://oidc.provider.com/authorize?client_id=abc&scope=openid", nil)

	suite.mockIDPService.On("GetIdentityProvider", "idp-123").
		Return(&idp.IDPDTO{ID: "idp-123", Name: "TestOIDCProvider"}, nil)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), flowcm.ExecExternalRedirection, resp.Status)
	assert.Contains(suite.T(), resp.RedirectURL, "https://oidc.provider.com/authorize")
	assert.Equal(suite.T(), "TestOIDCProvider", resp.AdditionalData[flowcm.DataIDPName])
	suite.mockOIDCService.AssertExpectations(suite.T())
	suite.mockIDPService.AssertExpectations(suite.T())
}

func (suite *OIDCAuthExecutorTestSuite) TestExecute_CodeProvided_ValidIDToken_AuthenticatesUser() {
	ctx := &flowcore.NodeContext{
		FlowID:   "flow-123",
		FlowType: flowcm.FlowTypeAuthentication,
		UserInputData: map[string]string{
			"code": "auth_code_123",
		},
		NodeProperties: map[string]interface{}{
			"idpId": "idp-123",
		},
	}

	tokenResp := &authnoauth.TokenResponse{
		AccessToken: "access_token_123",
		TokenType:   "Bearer",
		Scope:       "openid profile email",
		IDToken:     "id_token_jwt_123",
		ExpiresIn:   3600,
	}

	idTokenClaims := map[string]interface{}{
		"sub":   "user-sub-123",
		"email": "test@example.com",
		"name":  "Test User",
		"iss":   "https://oidc.provider.com",
		"aud":   "client-id-123",
		"exp":   1234567890,
		"iat":   1234567800,
	}

	existingUser := &user.User{
		ID:               "user-123",
		OrganizationUnit: testOUID,
		Type:             "INTERNAL",
	}

	oauthConfig := &authnoauth.OAuthClientConfig{
		Scopes: []string{"openid"},
	}

	suite.mockOIDCService.On("ExchangeCodeForToken", "idp-123", "auth_code_123", true).
		Return(tokenResp, nil)
	suite.mockOIDCService.On("GetIDTokenClaims", "id_token_jwt_123").
		Return(idTokenClaims, nil)
	suite.mockOIDCService.On("GetInternalUser", "user-sub-123").
		Return(existingUser, nil)
	suite.mockOIDCService.On("GetOAuthClientConfig", "idp-123").
		Return(oauthConfig, nil)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), flowcm.ExecComplete, resp.Status)
	assert.True(suite.T(), resp.AuthenticatedUser.IsAuthenticated)
	assert.Equal(suite.T(), "user-123", resp.AuthenticatedUser.UserID)
	assert.Equal(suite.T(), testOUID, resp.AuthenticatedUser.OrganizationUnitID)
	assert.Equal(suite.T(), "test@example.com", resp.RuntimeData["email"])
	suite.mockOIDCService.AssertExpectations(suite.T())
}

func (suite *OIDCAuthExecutorTestSuite) TestProcessAuthFlowResponse_ValidIDToken_Success() {
	ctx := &flowcore.NodeContext{
		FlowID:   "flow-123",
		FlowType: flowcm.FlowTypeAuthentication,
		UserInputData: map[string]string{
			"code": "auth_code_123",
		},
		NodeProperties: map[string]interface{}{
			"idpId": "idp-123",
		},
	}

	execResp := &flowcm.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	tokenResp := &authnoauth.TokenResponse{
		AccessToken: "access_token_123",
		TokenType:   "Bearer",
		Scope:       "openid profile",
		IDToken:     "id_token_jwt",
		ExpiresIn:   3600,
	}

	idTokenClaims := map[string]interface{}{
		"sub":   "user-sub-456",
		"email": "user@example.com",
		"iss":   "https://provider.com",
		"aud":   "client-id",
	}

	existingUser := &user.User{
		ID:               "user-456",
		OrganizationUnit: "ou-456",
		Type:             "INTERNAL",
	}

	oauthConfig := &authnoauth.OAuthClientConfig{
		Scopes: []string{"openid"},
	}

	suite.mockOIDCService.On("ExchangeCodeForToken", "idp-123", "auth_code_123", true).
		Return(tokenResp, nil)
	suite.mockOIDCService.On("GetIDTokenClaims", "id_token_jwt").
		Return(idTokenClaims, nil)
	suite.mockOIDCService.On("GetInternalUser", "user-sub-456").
		Return(existingUser, nil)
	suite.mockOIDCService.On("GetOAuthClientConfig", "idp-123").
		Return(oauthConfig, nil)

	err := suite.executor.ProcessAuthFlowResponse(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), flowcm.ExecComplete, execResp.Status)
	assert.True(suite.T(), execResp.AuthenticatedUser.IsAuthenticated)
	suite.mockOIDCService.AssertExpectations(suite.T())
}

func (suite *OIDCAuthExecutorTestSuite) TestProcessAuthFlowResponse_InvalidNonce() {
	ctx := &flowcore.NodeContext{
		FlowID:   "flow-123",
		FlowType: flowcm.FlowTypeAuthentication,
		UserInputData: map[string]string{
			"code":  "auth_code_123",
			"nonce": "expected_nonce_123",
		},
		NodeProperties: map[string]interface{}{
			"idpId": "idp-123",
		},
	}

	execResp := &flowcm.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	tokenResp := &authnoauth.TokenResponse{
		AccessToken: "access_token_123",
		TokenType:   "Bearer",
		Scope:       "openid profile",
		IDToken:     "id_token_jwt",
		ExpiresIn:   3600,
	}

	idTokenClaims := map[string]interface{}{
		"sub":   "user-sub-123",
		"nonce": "different_nonce_456",
	}

	suite.mockOIDCService.On("ExchangeCodeForToken", "idp-123", "auth_code_123", true).
		Return(tokenResp, nil)
	suite.mockOIDCService.On("GetIDTokenClaims", "id_token_jwt").
		Return(idTokenClaims, nil)

	err := suite.executor.ProcessAuthFlowResponse(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), flowcm.ExecFailure, execResp.Status)
	assert.Contains(suite.T(), execResp.FailureReason, "Nonce mismatch")
	suite.mockOIDCService.AssertExpectations(suite.T())
}

func (suite *OIDCAuthExecutorTestSuite) TestProcessAuthFlowResponse_NoSubClaim() {
	ctx := &flowcore.NodeContext{
		FlowID:   "flow-123",
		FlowType: flowcm.FlowTypeAuthentication,
		UserInputData: map[string]string{
			"code": "auth_code_123",
		},
		NodeProperties: map[string]interface{}{
			"idpId": "idp-123",
		},
	}

	execResp := &flowcm.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	tokenResp := &authnoauth.TokenResponse{
		AccessToken: "access_token_123",
		TokenType:   "Bearer",
		Scope:       "openid",
		IDToken:     "id_token_jwt",
		ExpiresIn:   3600,
	}

	idTokenClaims := map[string]interface{}{
		"email": "test@example.com",
		"name":  "Test User",
	}

	suite.mockOIDCService.On("ExchangeCodeForToken", "idp-123", "auth_code_123", true).
		Return(tokenResp, nil)
	suite.mockOIDCService.On("GetIDTokenClaims", "id_token_jwt").
		Return(idTokenClaims, nil)

	err := suite.executor.ProcessAuthFlowResponse(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), flowcm.ExecFailure, execResp.Status)
	assert.Contains(suite.T(), execResp.FailureReason, "sub claim not found")
	suite.mockOIDCService.AssertExpectations(suite.T())
}

func (suite *OIDCAuthExecutorTestSuite) TestProcessAuthFlowResponse_RegistrationFlow_UserNotFound() {
	ctx := &flowcore.NodeContext{
		FlowID:   "flow-123",
		FlowType: flowcm.FlowTypeRegistration,
		UserInputData: map[string]string{
			"code": "auth_code_123",
		},
		NodeProperties: map[string]interface{}{
			"idpId": "idp-123",
		},
	}

	execResp := &flowcm.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	tokenResp := &authnoauth.TokenResponse{
		AccessToken: "access_token_123",
		TokenType:   "Bearer",
		Scope:       "openid profile email",
		IDToken:     "id_token_jwt",
		ExpiresIn:   3600,
	}

	idTokenClaims := map[string]interface{}{
		"sub":   "new-user-sub",
		"email": "newuser@example.com",
		"name":  "New User",
	}

	oauthConfig := &authnoauth.OAuthClientConfig{
		Scopes: []string{"openid"},
	}

	suite.mockOIDCService.On("ExchangeCodeForToken", "idp-123", "auth_code_123", true).
		Return(tokenResp, nil)
	suite.mockOIDCService.On("GetIDTokenClaims", "id_token_jwt").
		Return(idTokenClaims, nil)
	suite.mockOIDCService.On("GetInternalUser", "new-user-sub").
		Return(nil, &serviceerror.ServiceError{
			Code: authncm.ErrorUserNotFound.Code,
			Type: serviceerror.ClientErrorType,
		})
	suite.mockOIDCService.On("GetOAuthClientConfig", "idp-123").
		Return(oauthConfig, nil)

	err := suite.executor.ProcessAuthFlowResponse(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), flowcm.ExecComplete, execResp.Status)
	assert.False(suite.T(), execResp.AuthenticatedUser.IsAuthenticated)
	assert.Equal(suite.T(), "new-user-sub", execResp.RuntimeData["sub"])
	suite.mockOIDCService.AssertExpectations(suite.T())
}

func (suite *OIDCAuthExecutorTestSuite) TestProcessAuthFlowResponse_AuthFlow_UserNotFound() {
	ctx := &flowcore.NodeContext{
		FlowID:   "flow-123",
		FlowType: flowcm.FlowTypeAuthentication,
		UserInputData: map[string]string{
			"code": "auth_code_123",
		},
		NodeProperties: map[string]interface{}{
			"idpId": "idp-123",
		},
	}

	execResp := &flowcm.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	tokenResp := &authnoauth.TokenResponse{
		AccessToken: "access_token_123",
		TokenType:   "Bearer",
		Scope:       "openid",
		IDToken:     "id_token_jwt",
		ExpiresIn:   3600,
	}

	idTokenClaims := map[string]interface{}{
		"sub": "unknown-user",
	}

	suite.mockOIDCService.On("ExchangeCodeForToken", "idp-123", "auth_code_123", true).
		Return(tokenResp, nil)
	suite.mockOIDCService.On("GetIDTokenClaims", "id_token_jwt").
		Return(idTokenClaims, nil)
	suite.mockOIDCService.On("GetInternalUser", "unknown-user").
		Return(nil, &serviceerror.ServiceError{
			Code: authncm.ErrorUserNotFound.Code,
			Type: serviceerror.ClientErrorType,
		})

	err := suite.executor.ProcessAuthFlowResponse(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), flowcm.ExecFailure, execResp.Status)
	assert.Equal(suite.T(), failureReasonUserNotFound, execResp.FailureReason)
	suite.mockOIDCService.AssertExpectations(suite.T())
}

func (suite *OIDCAuthExecutorTestSuite) TestProcessAuthFlowResponse_UserAlreadyExists_RegistrationFlow() {
	ctx := &flowcore.NodeContext{
		FlowID:   "flow-123",
		FlowType: flowcm.FlowTypeRegistration,
		UserInputData: map[string]string{
			"code": "auth_code_123",
		},
		NodeProperties: map[string]interface{}{
			"idpId": "idp-123",
		},
	}

	execResp := &flowcm.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	tokenResp := &authnoauth.TokenResponse{
		AccessToken: "access_token_123",
		TokenType:   "Bearer",
		Scope:       "openid",
		IDToken:     "id_token_jwt",
		ExpiresIn:   3600,
	}

	idTokenClaims := map[string]interface{}{
		"sub": "existing-user-sub",
	}

	existingUser := &user.User{
		ID:               "user-789",
		OrganizationUnit: "ou-789",
	}

	suite.mockOIDCService.On("ExchangeCodeForToken", "idp-123", "auth_code_123", true).
		Return(tokenResp, nil)
	suite.mockOIDCService.On("GetIDTokenClaims", "id_token_jwt").
		Return(idTokenClaims, nil)
	suite.mockOIDCService.On("GetInternalUser", "existing-user-sub").
		Return(existingUser, nil)

	err := suite.executor.ProcessAuthFlowResponse(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), flowcm.ExecFailure, execResp.Status)
	assert.Contains(suite.T(), execResp.FailureReason, "User already exists")
	suite.mockOIDCService.AssertExpectations(suite.T())
}

func (suite *OIDCAuthExecutorTestSuite) TestGetIDTokenClaims_Success() {
	execResp := &flowcm.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	expectedClaims := map[string]interface{}{
		"sub":   "user-sub-123",
		"email": "test@example.com",
		"name":  "Test User",
		"iss":   "https://provider.com",
		"aud":   "client-id",
	}

	suite.mockOIDCService.On("GetIDTokenClaims", "id_token_jwt").
		Return(expectedClaims, nil)

	claims, err := suite.executor.GetIDTokenClaims(execResp, "id_token_jwt")

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), claims)
	assert.Equal(suite.T(), "user-sub-123", claims["sub"])
	assert.Equal(suite.T(), "test@example.com", claims["email"])
	suite.mockOIDCService.AssertExpectations(suite.T())
}

func (suite *OIDCAuthExecutorTestSuite) TestGetIDTokenClaims_ClientError() {
	execResp := &flowcm.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	suite.mockOIDCService.On("GetIDTokenClaims", "invalid_token").
		Return(nil, &serviceerror.ServiceError{
			Type:             serviceerror.ClientErrorType,
			ErrorDescription: "Invalid ID token",
		})

	claims, err := suite.executor.GetIDTokenClaims(execResp, "invalid_token")

	assert.NoError(suite.T(), err)
	assert.Nil(suite.T(), claims)
	assert.Equal(suite.T(), flowcm.ExecFailure, execResp.Status)
	assert.Equal(suite.T(), "Invalid ID token", execResp.FailureReason)
	suite.mockOIDCService.AssertExpectations(suite.T())
}

func (suite *OIDCAuthExecutorTestSuite) TestGetIDTokenClaims_ServerError() {
	execResp := &flowcm.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	suite.mockOIDCService.On("GetIDTokenClaims", "id_token").
		Return(nil, &serviceerror.ServiceError{
			Type:             serviceerror.ServerErrorType,
			Code:             "OIDC-5000",
			ErrorDescription: "Failed to extract claims",
		})

	claims, err := suite.executor.GetIDTokenClaims(execResp, "id_token")

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), claims)
	assert.Contains(suite.T(), err.Error(), "failed to extract claims from the ID token")
	suite.mockOIDCService.AssertExpectations(suite.T())
}

func (suite *OIDCAuthExecutorTestSuite) TestProcessAuthFlowResponse_WithAdditionalScopes_FetchesUserInfo() {
	ctx := &flowcore.NodeContext{
		FlowID:   "flow-123",
		FlowType: flowcm.FlowTypeAuthentication,
		UserInputData: map[string]string{
			"code": "auth_code_123",
		},
		NodeProperties: map[string]interface{}{
			"idpId": "idp-123",
		},
	}

	execResp := &flowcm.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	tokenResp := &authnoauth.TokenResponse{
		AccessToken: "access_token_123",
		TokenType:   "Bearer",
		Scope:       "openid profile email",
		IDToken:     "id_token_jwt",
		ExpiresIn:   3600,
	}

	idTokenClaims := map[string]interface{}{
		"sub": "user-sub-123",
	}

	userInfo := map[string]interface{}{
		"sub":     "user-sub-123",
		"email":   "user@example.com",
		"phone":   "+1234567890",
		"address": "123 Main St",
	}

	existingUser := &user.User{
		ID:               "user-123",
		OrganizationUnit: testOUID,
		Type:             "INTERNAL",
	}

	oauthConfig := &authnoauth.OAuthClientConfig{
		Scopes: []string{"openid", "profile", "email"},
	}

	suite.mockOIDCService.On("ExchangeCodeForToken", "idp-123", "auth_code_123", true).
		Return(tokenResp, nil)
	suite.mockOIDCService.On("GetIDTokenClaims", "id_token_jwt").
		Return(idTokenClaims, nil)
	suite.mockOIDCService.On("GetInternalUser", "user-sub-123").
		Return(existingUser, nil)
	suite.mockOIDCService.On("GetOAuthClientConfig", "idp-123").
		Return(oauthConfig, nil)
	suite.mockOIDCService.On("FetchUserInfo", "idp-123", "access_token_123").
		Return(userInfo, nil)

	err := suite.executor.ProcessAuthFlowResponse(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), flowcm.ExecComplete, execResp.Status)
	assert.True(suite.T(), execResp.AuthenticatedUser.IsAuthenticated)
	assert.Contains(suite.T(), execResp.AuthenticatedUser.Attributes, "email")
	assert.Contains(suite.T(), execResp.AuthenticatedUser.Attributes, "phone")
	suite.mockOIDCService.AssertExpectations(suite.T())
}

func (suite *OIDCAuthExecutorTestSuite) TestProcessAuthFlowResponse_NoCodeProvided() {
	ctx := &flowcore.NodeContext{
		FlowID:        "flow-123",
		FlowType:      flowcm.FlowTypeAuthentication,
		UserInputData: map[string]string{},
		NodeProperties: map[string]interface{}{
			"idpId": "idp-123",
		},
	}

	execResp := &flowcm.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	err := suite.executor.ProcessAuthFlowResponse(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), flowcm.ExecFailure, execResp.Status)
	assert.Equal(suite.T(), failureReasonInvalidAuthorizationCode, execResp.FailureReason)
}

func (suite *OIDCAuthExecutorTestSuite) TestProcessAuthFlowResponse_FiltersNonUserClaimsFromIDToken() {
	ctx := &flowcore.NodeContext{
		FlowID:   "flow-123",
		FlowType: flowcm.FlowTypeAuthentication,
		UserInputData: map[string]string{
			"code": "auth_code_123",
		},
		NodeProperties: map[string]interface{}{
			"idpId": "idp-123",
		},
	}

	execResp := &flowcm.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	tokenResp := &authnoauth.TokenResponse{
		AccessToken: "access_token_123",
		TokenType:   "Bearer",
		Scope:       "openid profile",
		IDToken:     "id_token_jwt",
		ExpiresIn:   3600,
	}

	idTokenClaims := map[string]interface{}{
		"sub":     "user-sub-123",
		"email":   "user@example.com",
		"name":    "User Name",
		"iss":     "https://provider.com",
		"aud":     "client-id",
		"exp":     1234567890,
		"iat":     1234567800,
		"at_hash": "hash_value",
		"nonce":   "nonce_value",
	}

	existingUser := &user.User{
		ID:               "user-123",
		OrganizationUnit: testOUID,
		Type:             "INTERNAL",
	}

	oauthConfig := &authnoauth.OAuthClientConfig{
		Scopes: []string{"openid"},
	}

	suite.mockOIDCService.On("ExchangeCodeForToken", "idp-123", "auth_code_123", true).
		Return(tokenResp, nil)
	suite.mockOIDCService.On("GetIDTokenClaims", "id_token_jwt").
		Return(idTokenClaims, nil)
	suite.mockOIDCService.On("GetInternalUser", "user-sub-123").
		Return(existingUser, nil)
	suite.mockOIDCService.On("GetOAuthClientConfig", "idp-123").
		Return(oauthConfig, nil)

	err := suite.executor.ProcessAuthFlowResponse(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), flowcm.ExecComplete, execResp.Status)
	assert.Contains(suite.T(), execResp.AuthenticatedUser.Attributes, "email")
	assert.Contains(suite.T(), execResp.AuthenticatedUser.Attributes, "name")
	assert.NotContains(suite.T(), execResp.AuthenticatedUser.Attributes, "iss")
	assert.NotContains(suite.T(), execResp.AuthenticatedUser.Attributes, "aud")
	assert.NotContains(suite.T(), execResp.AuthenticatedUser.Attributes, "exp")
	assert.NotContains(suite.T(), execResp.AuthenticatedUser.Attributes, "iat")
	assert.NotContains(suite.T(), execResp.AuthenticatedUser.Attributes, "at_hash")
	assert.NotContains(suite.T(), execResp.AuthenticatedUser.Attributes, "nonce")
	assert.NotContains(suite.T(), execResp.AuthenticatedUser.Attributes, "sub")
	suite.mockOIDCService.AssertExpectations(suite.T())
}

func (suite *OIDCAuthExecutorTestSuite) TestProcessAuthFlowResponse_EmailInIDToken() {
	ctx := &flowcore.NodeContext{
		FlowID:   "flow-123",
		FlowType: flowcm.FlowTypeAuthentication,
		UserInputData: map[string]string{
			"code": "auth_code_123",
		},
		NodeProperties: map[string]interface{}{
			"idpId": "idp-123",
		},
	}

	execResp := &flowcm.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	tokenResp := &authnoauth.TokenResponse{
		AccessToken: "access_token_123",
		TokenType:   "Bearer",
		Scope:       "openid email",
		IDToken:     "id_token_jwt",
		ExpiresIn:   3600,
	}

	idTokenClaims := map[string]interface{}{
		"sub":   "user-sub-789",
		"email": "user@test.com",
		"iss":   "https://provider.com",
		"aud":   "client-id",
	}

	existingUser := &user.User{
		ID:               "user-789",
		OrganizationUnit: "ou-789",
		Type:             "INTERNAL",
	}

	oauthConfig := &authnoauth.OAuthClientConfig{
		Scopes: []string{"openid"},
	}

	suite.mockOIDCService.On("ExchangeCodeForToken", "idp-123", "auth_code_123", true).
		Return(tokenResp, nil)
	suite.mockOIDCService.On("GetIDTokenClaims", "id_token_jwt").
		Return(idTokenClaims, nil)
	suite.mockOIDCService.On("GetInternalUser", "user-sub-789").
		Return(existingUser, nil)
	suite.mockOIDCService.On("GetOAuthClientConfig", "idp-123").
		Return(oauthConfig, nil)

	err := suite.executor.ProcessAuthFlowResponse(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), flowcm.ExecComplete, execResp.Status)
	assert.True(suite.T(), execResp.AuthenticatedUser.IsAuthenticated)
	assert.Equal(suite.T(), "user@test.com", execResp.RuntimeData["email"])
	assert.Equal(suite.T(), "user@test.com", execResp.AuthenticatedUser.Attributes["email"])
	suite.mockOIDCService.AssertExpectations(suite.T())
}

func (suite *OIDCAuthExecutorTestSuite) TestProcessAuthFlowResponse_NoEmailInIDToken() {
	ctx := &flowcore.NodeContext{
		FlowID:   "flow-123",
		FlowType: flowcm.FlowTypeAuthentication,
		UserInputData: map[string]string{
			"code": "auth_code_123",
		},
		NodeProperties: map[string]interface{}{
			"idpId": "idp-123",
		},
	}

	execResp := &flowcm.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	tokenResp := &authnoauth.TokenResponse{
		AccessToken: "access_token_123",
		TokenType:   "Bearer",
		Scope:       "openid profile",
		IDToken:     "id_token_jwt",
		ExpiresIn:   3600,
	}

	idTokenClaims := map[string]interface{}{
		"sub":  "user-sub-789",
		"name": "Test User",
		"iss":  "https://provider.com",
		"aud":  "client-id",
	}

	existingUser := &user.User{
		ID:               "user-789",
		OrganizationUnit: "ou-789",
		Type:             "INTERNAL",
	}

	oauthConfig := &authnoauth.OAuthClientConfig{
		Scopes: []string{"openid"},
	}

	suite.mockOIDCService.On("ExchangeCodeForToken", "idp-123", "auth_code_123", true).
		Return(tokenResp, nil)
	suite.mockOIDCService.On("GetIDTokenClaims", "id_token_jwt").
		Return(idTokenClaims, nil)
	suite.mockOIDCService.On("GetInternalUser", "user-sub-789").
		Return(existingUser, nil)
	suite.mockOIDCService.On("GetOAuthClientConfig", "idp-123").
		Return(oauthConfig, nil)

	err := suite.executor.ProcessAuthFlowResponse(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), flowcm.ExecComplete, execResp.Status)
	assert.True(suite.T(), execResp.AuthenticatedUser.IsAuthenticated)
	assert.NotContains(suite.T(), execResp.RuntimeData, "email")
	assert.NotContains(suite.T(), execResp.AuthenticatedUser.Attributes, "email")
	suite.mockOIDCService.AssertExpectations(suite.T())
}

func (suite *OIDCAuthExecutorTestSuite) TestProcessAuthFlowResponse_EmptyEmailInIDToken() {
	ctx := &flowcore.NodeContext{
		FlowID:   "flow-123",
		FlowType: flowcm.FlowTypeAuthentication,
		UserInputData: map[string]string{
			"code": "auth_code_123",
		},
		NodeProperties: map[string]interface{}{
			"idpId": "idp-123",
		},
	}

	execResp := &flowcm.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	tokenResp := &authnoauth.TokenResponse{
		AccessToken: "access_token_123",
		TokenType:   "Bearer",
		Scope:       "openid email",
		IDToken:     "id_token_jwt",
		ExpiresIn:   3600,
	}

	idTokenClaims := map[string]interface{}{
		"sub":   "user-sub-789",
		"email": "",
		"iss":   "https://provider.com",
		"aud":   "client-id",
	}

	existingUser := &user.User{
		ID:               "user-789",
		OrganizationUnit: "ou-789",
		Type:             "INTERNAL",
	}

	oauthConfig := &authnoauth.OAuthClientConfig{
		Scopes: []string{"openid"},
	}

	suite.mockOIDCService.On("ExchangeCodeForToken", "idp-123", "auth_code_123", true).
		Return(tokenResp, nil)
	suite.mockOIDCService.On("GetIDTokenClaims", "id_token_jwt").
		Return(idTokenClaims, nil)
	suite.mockOIDCService.On("GetInternalUser", "user-sub-789").
		Return(existingUser, nil)
	suite.mockOIDCService.On("GetOAuthClientConfig", "idp-123").
		Return(oauthConfig, nil)

	err := suite.executor.ProcessAuthFlowResponse(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), flowcm.ExecComplete, execResp.Status)
	assert.True(suite.T(), execResp.AuthenticatedUser.IsAuthenticated)
	assert.NotContains(suite.T(), execResp.RuntimeData, "email")
	assert.Equal(suite.T(), "", execResp.AuthenticatedUser.Attributes["email"])
	suite.mockOIDCService.AssertExpectations(suite.T())
}

func (suite *OIDCAuthExecutorTestSuite) TestProcessAuthFlowResponse_RegistrationFlow_WithEmail() {
	ctx := &flowcore.NodeContext{
		FlowID:   "flow-123",
		FlowType: flowcm.FlowTypeRegistration,
		UserInputData: map[string]string{
			"code": "auth_code_123",
		},
		NodeProperties: map[string]interface{}{
			"idpId": "idp-123",
		},
	}

	execResp := &flowcm.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	tokenResp := &authnoauth.TokenResponse{
		AccessToken: "access_token_123",
		TokenType:   "Bearer",
		Scope:       "openid email",
		IDToken:     "id_token_jwt",
		ExpiresIn:   3600,
	}

	idTokenClaims := map[string]interface{}{
		"sub":   "new-user-sub",
		"email": "newuser@example.com",
		"name":  "New User",
		"iss":   "https://provider.com",
		"aud":   "client-id",
	}

	oauthConfig := &authnoauth.OAuthClientConfig{
		Scopes: []string{"openid"},
	}

	suite.mockOIDCService.On("ExchangeCodeForToken", "idp-123", "auth_code_123", true).
		Return(tokenResp, nil)
	suite.mockOIDCService.On("GetIDTokenClaims", "id_token_jwt").
		Return(idTokenClaims, nil)
	suite.mockOIDCService.On("GetInternalUser", "new-user-sub").
		Return(nil, &serviceerror.ServiceError{
			Code: authncm.ErrorUserNotFound.Code,
			Type: serviceerror.ClientErrorType,
		})
	suite.mockOIDCService.On("GetOAuthClientConfig", "idp-123").
		Return(oauthConfig, nil)

	err := suite.executor.ProcessAuthFlowResponse(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), flowcm.ExecComplete, execResp.Status)
	assert.False(suite.T(), execResp.AuthenticatedUser.IsAuthenticated)
	assert.Equal(suite.T(), "new-user-sub", execResp.RuntimeData["sub"])
	assert.Equal(suite.T(), "newuser@example.com", execResp.RuntimeData["email"])
	assert.Equal(suite.T(), "newuser@example.com", execResp.AuthenticatedUser.Attributes["email"])
	suite.mockOIDCService.AssertExpectations(suite.T())
}

func (suite *OIDCAuthExecutorTestSuite) TestProcessAuthFlowResponse_EmailFromUserInfo() {
	ctx := &flowcore.NodeContext{
		FlowID:   "flow-123",
		FlowType: flowcm.FlowTypeAuthentication,
		UserInputData: map[string]string{
			"code": "auth_code_123",
		},
		NodeProperties: map[string]interface{}{
			"idpId": "idp-123",
		},
	}

	execResp := &flowcm.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	tokenResp := &authnoauth.TokenResponse{
		AccessToken: "access_token_123",
		TokenType:   "Bearer",
		Scope:       "openid profile email",
		IDToken:     "id_token_jwt",
		ExpiresIn:   3600,
	}

	idTokenClaims := map[string]interface{}{
		"sub":  "user-sub-789",
		"name": "Test User",
		"iss":  "https://provider.com",
		"aud":  "client-id",
	}

	userInfo := map[string]interface{}{
		"sub":   "user-sub-789",
		"email": "fromUserInfo@example.com",
		"name":  "Test User",
	}

	existingUser := &user.User{
		ID:               "user-789",
		OrganizationUnit: "ou-789",
		Type:             "INTERNAL",
	}

	oauthConfig := &authnoauth.OAuthClientConfig{
		Scopes: []string{"openid", "profile", "email"},
	}

	suite.mockOIDCService.On("ExchangeCodeForToken", "idp-123", "auth_code_123", true).
		Return(tokenResp, nil)
	suite.mockOIDCService.On("GetIDTokenClaims", "id_token_jwt").
		Return(idTokenClaims, nil)
	suite.mockOIDCService.On("GetInternalUser", "user-sub-789").
		Return(existingUser, nil)
	suite.mockOIDCService.On("GetOAuthClientConfig", "idp-123").
		Return(oauthConfig, nil)
	suite.mockOIDCService.On("FetchUserInfo", "idp-123", "access_token_123").
		Return(userInfo, nil)

	err := suite.executor.ProcessAuthFlowResponse(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), flowcm.ExecComplete, execResp.Status)
	assert.True(suite.T(), execResp.AuthenticatedUser.IsAuthenticated)
	assert.Equal(suite.T(), "fromUserInfo@example.com", execResp.RuntimeData["email"])
	assert.Equal(suite.T(), "fromUserInfo@example.com", execResp.AuthenticatedUser.Attributes["email"])
	suite.mockOIDCService.AssertExpectations(suite.T())
}

func (suite *OIDCAuthExecutorTestSuite) TestProcessAuthFlowResponse_EmailInIDToken_NilRuntimeData() {
	ctx := &flowcore.NodeContext{
		FlowID:   "flow-123",
		FlowType: flowcm.FlowTypeAuthentication,
		UserInputData: map[string]string{
			"code": "auth_code_123",
		},
		NodeProperties: map[string]interface{}{
			"idpId": "idp-123",
		},
	}

	execResp := &flowcm.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    nil, // Explicitly nil
	}

	tokenResp := &authnoauth.TokenResponse{
		AccessToken: "access_token_123",
		TokenType:   "Bearer",
		Scope:       "openid email",
		IDToken:     "id_token_jwt",
		ExpiresIn:   3600,
	}

	idTokenClaims := map[string]interface{}{
		"sub":   "user-sub-999",
		"email": "niltest@example.com",
		"iss":   "https://provider.com",
		"aud":   "client-id",
	}

	existingUser := &user.User{
		ID:               "user-999",
		OrganizationUnit: "ou-999",
		Type:             "INTERNAL",
	}

	oauthConfig := &authnoauth.OAuthClientConfig{
		Scopes: []string{"openid"},
	}

	suite.mockOIDCService.On("ExchangeCodeForToken", "idp-123", "auth_code_123", true).
		Return(tokenResp, nil)
	suite.mockOIDCService.On("GetIDTokenClaims", "id_token_jwt").
		Return(idTokenClaims, nil)
	suite.mockOIDCService.On("GetInternalUser", "user-sub-999").
		Return(existingUser, nil)
	suite.mockOIDCService.On("GetOAuthClientConfig", "idp-123").
		Return(oauthConfig, nil)

	err := suite.executor.ProcessAuthFlowResponse(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), flowcm.ExecComplete, execResp.Status)
	assert.True(suite.T(), execResp.AuthenticatedUser.IsAuthenticated)
	assert.NotNil(suite.T(), execResp.RuntimeData, "RuntimeData should be initialized")
	assert.Equal(suite.T(), "niltest@example.com", execResp.RuntimeData["email"])
	assert.Equal(suite.T(), "niltest@example.com", execResp.AuthenticatedUser.Attributes["email"])
	suite.mockOIDCService.AssertExpectations(suite.T())
}

// Test provisioning functionality

func (suite *OIDCAuthExecutorTestSuite) TestProcessAuthFlowResponse_AuthFlow_UserNotFound_ProvisioningSucceeds() {
	ctx := &flowcore.NodeContext{
		FlowID:   "flow-123",
		FlowType: flowcm.FlowTypeAuthentication,
		Application: appmodel.Application{
			AllowedUserTypes: []string{"employee"},
		},
		UserInputData: map[string]string{
			"code": "auth_code_123",
		},
		NodeProperties: map[string]interface{}{
			"idpId": "idp-123",
		},
	}

	execResp := &flowcm.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	tokenResp := &authnoauth.TokenResponse{
		AccessToken: "access_token_123",
		TokenType:   "Bearer",
		Scope:       "openid",
		IDToken:     "id_token_jwt",
		ExpiresIn:   3600,
	}

	idTokenClaims := map[string]interface{}{
		"sub":   "new-user-sub",
		"email": "newuser@example.com",
		"name":  "New User",
		"iss":   "https://provider.com",
		"aud":   "client-id",
	}

	userSchema := &userschema.UserSchema{
		ID:                    "schema-123",
		Name:                  "employee",
		OrganizationUnitID:    testOUID,
		AllowSelfRegistration: true,
	}

	provisionedUser := &user.User{
		ID:               "user-provisioned-123",
		OrganizationUnit: testOUID,
		Type:             "employee",
	}

	oauthConfig := &authnoauth.OAuthClientConfig{
		Scopes: []string{"openid"},
	}

	suite.mockOIDCService.On("ExchangeCodeForToken", "idp-123", "auth_code_123", true).
		Return(tokenResp, nil)
	suite.mockOIDCService.On("GetIDTokenClaims", "id_token_jwt").
		Return(idTokenClaims, nil)
	suite.mockOIDCService.On("GetInternalUser", "new-user-sub").
		Return(nil, &serviceerror.ServiceError{
			Code: authncm.ErrorUserNotFound.Code,
			Type: serviceerror.ClientErrorType,
		})
	suite.mockOIDCService.On("GetOAuthClientConfig", "idp-123").
		Return(oauthConfig, nil)
	suite.mockUserSchemaService.On("GetUserSchemaByName", "employee").
		Return(userSchema, nil)
	suite.mockUserService.On("CreateUser", mock.MatchedBy(func(u *user.User) bool {
		return u.Type == testUserType && u.OrganizationUnit == testOUID
	})).Return(provisionedUser, nil)

	err := suite.executor.ProcessAuthFlowResponse(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), flowcm.ExecComplete, execResp.Status)
	assert.True(suite.T(), execResp.AuthenticatedUser.IsAuthenticated)
	assert.Equal(suite.T(), "user-provisioned-123", execResp.AuthenticatedUser.UserID)
	assert.Equal(suite.T(), testOUID, execResp.AuthenticatedUser.OrganizationUnitID)
	assert.Equal(suite.T(), "employee", execResp.AuthenticatedUser.UserType)
	suite.mockOIDCService.AssertExpectations(suite.T())
	suite.mockUserSchemaService.AssertExpectations(suite.T())
	suite.mockUserService.AssertExpectations(suite.T())
}

func (suite *OIDCAuthExecutorTestSuite) TestProcessAuthFlowResponse_AuthFlow_UserNotFound_NoAllowedUserTypes() {
	ctx := &flowcore.NodeContext{
		FlowID:   "flow-123",
		FlowType: flowcm.FlowTypeAuthentication,
		Application: appmodel.Application{
			AllowedUserTypes: []string{},
		},
		UserInputData: map[string]string{
			"code": "auth_code_123",
		},
		NodeProperties: map[string]interface{}{
			"idpId": "idp-123",
		},
	}

	execResp := &flowcm.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	tokenResp := &authnoauth.TokenResponse{
		AccessToken: "access_token_123",
		TokenType:   "Bearer",
		Scope:       "openid",
		IDToken:     "id_token_jwt",
		ExpiresIn:   3600,
	}

	idTokenClaims := map[string]interface{}{
		"sub": "new-user-sub",
	}

	suite.mockOIDCService.On("ExchangeCodeForToken", "idp-123", "auth_code_123", true).
		Return(tokenResp, nil)
	suite.mockOIDCService.On("GetIDTokenClaims", "id_token_jwt").
		Return(idTokenClaims, nil)
	suite.mockOIDCService.On("GetInternalUser", "new-user-sub").
		Return(nil, &serviceerror.ServiceError{
			Code: authncm.ErrorUserNotFound.Code,
			Type: serviceerror.ClientErrorType,
		})

	err := suite.executor.ProcessAuthFlowResponse(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), flowcm.ExecFailure, execResp.Status)
	assert.Equal(suite.T(), failureReasonUserNotFound, execResp.FailureReason)
	suite.mockOIDCService.AssertExpectations(suite.T())
	suite.mockUserSchemaService.AssertNotCalled(suite.T(), "GetUserSchemaByName")
	suite.mockUserService.AssertNotCalled(suite.T(), "CreateUser")
}

func (suite *OIDCAuthExecutorTestSuite) TestProcessAuthFlowResponse_AuthFlow_UserNotFound_UserSchemaNotFound() {
	ctx := &flowcore.NodeContext{
		FlowID:   "flow-123",
		FlowType: flowcm.FlowTypeAuthentication,
		Application: appmodel.Application{
			AllowedUserTypes: []string{"employee"},
		},
		UserInputData: map[string]string{
			"code": "auth_code_123",
		},
		NodeProperties: map[string]interface{}{
			"idpId": "idp-123",
		},
	}

	execResp := &flowcm.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	tokenResp := &authnoauth.TokenResponse{
		AccessToken: "access_token_123",
		TokenType:   "Bearer",
		Scope:       "openid",
		IDToken:     "id_token_jwt",
		ExpiresIn:   3600,
	}

	idTokenClaims := map[string]interface{}{
		"sub": "new-user-sub",
	}

	suite.mockOIDCService.On("ExchangeCodeForToken", "idp-123", "auth_code_123", true).
		Return(tokenResp, nil)
	suite.mockOIDCService.On("GetIDTokenClaims", "id_token_jwt").
		Return(idTokenClaims, nil)
	suite.mockOIDCService.On("GetInternalUser", "new-user-sub").
		Return(nil, &serviceerror.ServiceError{
			Code: authncm.ErrorUserNotFound.Code,
			Type: serviceerror.ClientErrorType,
		})
	suite.mockUserSchemaService.On("GetUserSchemaByName", "employee").
		Return(nil, &serviceerror.ServiceError{
			Type:             serviceerror.ClientErrorType,
			Code:             "USRS-1002",
			ErrorDescription: "User schema not found",
		})

	err := suite.executor.ProcessAuthFlowResponse(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), flowcm.ExecFailure, execResp.Status)
	assert.Equal(suite.T(), "User not found and automatic provisioning is not available", execResp.FailureReason)
	suite.mockOIDCService.AssertExpectations(suite.T())
	suite.mockUserSchemaService.AssertExpectations(suite.T())
	suite.mockUserService.AssertNotCalled(suite.T(), "CreateUser")
}

func (suite *OIDCAuthExecutorTestSuite) TestProcessAuthFlowResponse_AuthFlow_UserNotFound_UserCreationFails() {
	ctx := &flowcore.NodeContext{
		FlowID:   "flow-123",
		FlowType: flowcm.FlowTypeAuthentication,
		Application: appmodel.Application{
			AllowedUserTypes: []string{"employee"},
		},
		UserInputData: map[string]string{
			"code": "auth_code_123",
		},
		NodeProperties: map[string]interface{}{
			"idpId": "idp-123",
		},
	}

	execResp := &flowcm.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	tokenResp := &authnoauth.TokenResponse{
		AccessToken: "access_token_123",
		TokenType:   "Bearer",
		Scope:       "openid",
		IDToken:     "id_token_jwt",
		ExpiresIn:   3600,
	}

	idTokenClaims := map[string]interface{}{
		"sub":   "new-user-sub",
		"email": "newuser@example.com",
	}

	userSchema := &userschema.UserSchema{
		ID:                    "schema-123",
		Name:                  "employee",
		OrganizationUnitID:    testOUID,
		AllowSelfRegistration: true,
	}

	suite.mockOIDCService.On("ExchangeCodeForToken", "idp-123", "auth_code_123", true).
		Return(tokenResp, nil)
	suite.mockOIDCService.On("GetIDTokenClaims", "id_token_jwt").
		Return(idTokenClaims, nil)
	suite.mockOIDCService.On("GetInternalUser", "new-user-sub").
		Return(nil, &serviceerror.ServiceError{
			Code: authncm.ErrorUserNotFound.Code,
			Type: serviceerror.ClientErrorType,
		})
	suite.mockUserSchemaService.On("GetUserSchemaByName", "employee").
		Return(userSchema, nil)
	suite.mockUserService.On("CreateUser", mock.MatchedBy(func(u *user.User) bool {
		return u.Type == testUserType && u.OrganizationUnit == testOUID
	})).Return(nil, &serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "USR-1014",
		ErrorDescription: "Attribute conflict",
	})

	err := suite.executor.ProcessAuthFlowResponse(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), flowcm.ExecFailure, execResp.Status)
	assert.Equal(suite.T(), "User not found and automatic provisioning is not available", execResp.FailureReason)
	suite.mockOIDCService.AssertExpectations(suite.T())
	suite.mockUserSchemaService.AssertExpectations(suite.T())
	suite.mockUserService.AssertExpectations(suite.T())
}

func (suite *OIDCAuthExecutorTestSuite) TestProcessAuthFlowResponse_AuthFlow_UserNotFound_MultipleAllowedUserTypes() {
	ctx := &flowcore.NodeContext{
		FlowID:   "flow-123",
		FlowType: flowcm.FlowTypeAuthentication,
		Application: appmodel.Application{
			AllowedUserTypes: []string{"employee", "customer", "partner"},
		},
		UserInputData: map[string]string{
			"code": "auth_code_123",
		},
		NodeProperties: map[string]interface{}{
			"idpId": "idp-123",
		},
	}

	execResp := &flowcm.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	tokenResp := &authnoauth.TokenResponse{
		AccessToken: "access_token_123",
		TokenType:   "Bearer",
		Scope:       "openid",
		IDToken:     "id_token_jwt",
		ExpiresIn:   3600,
	}

	idTokenClaims := map[string]interface{}{
		"sub":   "new-user-sub",
		"email": "newuser@example.com",
	}

	employeeSchema := &userschema.UserSchema{
		ID:                    "schema-employee",
		Name:                  "employee",
		OrganizationUnitID:    testOUID,
		AllowSelfRegistration: true,
	}
	customerSchema := &userschema.UserSchema{
		ID:                    "schema-customer",
		Name:                  "customer",
		OrganizationUnitID:    testOUID,
		AllowSelfRegistration: true,
	}
	partnerSchema := &userschema.UserSchema{
		ID:                    "schema-partner",
		Name:                  "partner",
		OrganizationUnitID:    testOUID,
		AllowSelfRegistration: true,
	}

	suite.mockOIDCService.On("ExchangeCodeForToken", "idp-123", "auth_code_123", true).
		Return(tokenResp, nil)
	suite.mockOIDCService.On("GetIDTokenClaims", "id_token_jwt").
		Return(idTokenClaims, nil)
	suite.mockOIDCService.On("GetInternalUser", "new-user-sub").
		Return(nil, &serviceerror.ServiceError{
			Code: authncm.ErrorUserNotFound.Code,
			Type: serviceerror.ClientErrorType,
		})
	// All three user types have self-registration enabled, so should fail
	suite.mockUserSchemaService.On("GetUserSchemaByName", "employee").
		Return(employeeSchema, nil)
	suite.mockUserSchemaService.On("GetUserSchemaByName", "customer").
		Return(customerSchema, nil)
	suite.mockUserSchemaService.On("GetUserSchemaByName", "partner").
		Return(partnerSchema, nil)

	err := suite.executor.ProcessAuthFlowResponse(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), flowcm.ExecFailure, execResp.Status)
	assert.Equal(suite.T(), "User not found and automatic provisioning is not available", execResp.FailureReason)
	suite.mockOIDCService.AssertExpectations(suite.T())
	suite.mockUserSchemaService.AssertExpectations(suite.T())
	suite.mockUserService.AssertNotCalled(suite.T(), "CreateUser")
}

func (suite *OIDCAuthExecutorTestSuite) TestProcessAuthFlowResponse_AuthFlow_UserNotFound_SelfRegistrationDisabled() {
	ctx := &flowcore.NodeContext{
		FlowID:   "flow-123",
		FlowType: flowcm.FlowTypeAuthentication,
		Application: appmodel.Application{
			AllowedUserTypes: []string{"employee"},
		},
		UserInputData: map[string]string{
			"code": "auth_code_123",
		},
		NodeProperties: map[string]interface{}{
			"idpId": "idp-123",
		},
	}

	execResp := &flowcm.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	tokenResp := &authnoauth.TokenResponse{
		AccessToken: "access_token_123",
		TokenType:   "Bearer",
		Scope:       "openid",
		IDToken:     "id_token_jwt",
		ExpiresIn:   3600,
	}

	idTokenClaims := map[string]interface{}{
		"sub": "new-user-sub",
	}

	userSchema := &userschema.UserSchema{
		ID:                    "schema-123",
		Name:                  "employee",
		OrganizationUnitID:    testOUID,
		AllowSelfRegistration: false,
	}

	suite.mockOIDCService.On("ExchangeCodeForToken", "idp-123", "auth_code_123", true).
		Return(tokenResp, nil)
	suite.mockOIDCService.On("GetIDTokenClaims", "id_token_jwt").
		Return(idTokenClaims, nil)
	suite.mockOIDCService.On("GetInternalUser", "new-user-sub").
		Return(nil, &serviceerror.ServiceError{
			Code: authncm.ErrorUserNotFound.Code,
			Type: serviceerror.ClientErrorType,
		})
	suite.mockUserSchemaService.On("GetUserSchemaByName", "employee").
		Return(userSchema, nil)

	err := suite.executor.ProcessAuthFlowResponse(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), flowcm.ExecFailure, execResp.Status)
	assert.Equal(suite.T(), "User not found and automatic provisioning is not available", execResp.FailureReason)
	suite.mockOIDCService.AssertExpectations(suite.T())
	suite.mockUserSchemaService.AssertExpectations(suite.T())
	suite.mockUserService.AssertNotCalled(suite.T(), "CreateUser")
}

func (suite *OIDCAuthExecutorTestSuite) TestProcessAuthFlowResponse_AuthFlow_UserNotFound_MultiTypes_OneWithSelfReg() {
	ctx := &flowcore.NodeContext{
		FlowID:   "flow-123",
		FlowType: flowcm.FlowTypeAuthentication,
		Application: appmodel.Application{
			AllowedUserTypes: []string{"employee", "customer", "partner"},
		},
		UserInputData: map[string]string{
			"code": "auth_code_123",
		},
		NodeProperties: map[string]interface{}{
			"idpId": "idp-123",
		},
	}

	execResp := &flowcm.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	tokenResp := &authnoauth.TokenResponse{
		AccessToken: "access_token_123",
		TokenType:   "Bearer",
		Scope:       "openid",
		IDToken:     "id_token_jwt",
		ExpiresIn:   3600,
	}

	idTokenClaims := map[string]interface{}{
		"sub":   "new-user-sub",
		"email": "newuser@example.com",
	}

	employeeSchema := &userschema.UserSchema{
		ID:                    "schema-employee",
		Name:                  "employee",
		OrganizationUnitID:    testOUID,
		AllowSelfRegistration: true,
	}
	customerSchema := &userschema.UserSchema{
		ID:                    "schema-customer",
		Name:                  "customer",
		OrganizationUnitID:    testOUID,
		AllowSelfRegistration: false,
	}
	partnerSchema := &userschema.UserSchema{
		ID:                    "schema-partner",
		Name:                  "partner",
		OrganizationUnitID:    testOUID,
		AllowSelfRegistration: false,
	}

	provisionedUser := &user.User{
		ID:               "user-provisioned-123",
		OrganizationUnit: testOUID,
		Type:             "employee",
	}

	oauthConfig := &authnoauth.OAuthClientConfig{
		Scopes: []string{"openid"},
	}

	suite.mockOIDCService.On("ExchangeCodeForToken", "idp-123", "auth_code_123", true).
		Return(tokenResp, nil)
	suite.mockOIDCService.On("GetIDTokenClaims", "id_token_jwt").
		Return(idTokenClaims, nil)
	suite.mockOIDCService.On("GetInternalUser", "new-user-sub").
		Return(nil, &serviceerror.ServiceError{
			Code: authncm.ErrorUserNotFound.Code,
			Type: serviceerror.ClientErrorType,
		})
	suite.mockOIDCService.On("GetOAuthClientConfig", "idp-123").
		Return(oauthConfig, nil)
	// Only employee has self-registration enabled
	suite.mockUserSchemaService.On("GetUserSchemaByName", "employee").
		Return(employeeSchema, nil)
	suite.mockUserSchemaService.On("GetUserSchemaByName", "customer").
		Return(customerSchema, nil)
	suite.mockUserSchemaService.On("GetUserSchemaByName", "partner").
		Return(partnerSchema, nil)
	suite.mockUserService.On("CreateUser", mock.MatchedBy(func(u *user.User) bool {
		return u.Type == "employee" && u.OrganizationUnit == testOUID
	})).Return(provisionedUser, nil)

	err := suite.executor.ProcessAuthFlowResponse(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), flowcm.ExecComplete, execResp.Status)
	assert.True(suite.T(), execResp.AuthenticatedUser.IsAuthenticated)
	assert.Equal(suite.T(), "employee", execResp.AuthenticatedUser.UserType)
	suite.mockOIDCService.AssertExpectations(suite.T())
	suite.mockUserSchemaService.AssertExpectations(suite.T())
	suite.mockUserService.AssertExpectations(suite.T())
}
