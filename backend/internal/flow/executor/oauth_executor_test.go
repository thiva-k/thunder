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
	"github.com/asgardeo/thunder/tests/mocks/authn/oauthmock"
	"github.com/asgardeo/thunder/tests/mocks/flow/coremock"
	"github.com/asgardeo/thunder/tests/mocks/idp/idpmock"
	"github.com/asgardeo/thunder/tests/mocks/usermock"
	"github.com/asgardeo/thunder/tests/mocks/userschemamock"
)

type OAuthExecutorTestSuite struct {
	suite.Suite
	mockOAuthService      *oauthmock.OAuthAuthnCoreServiceInterfaceMock
	mockIDPService        *idpmock.IDPServiceInterfaceMock
	mockFlowFactory       *coremock.FlowFactoryInterfaceMock
	mockUserService       *usermock.UserServiceInterfaceMock
	mockUserSchemaService *userschemamock.UserSchemaServiceInterfaceMock
	executor              oAuthExecutorInterface
}

const (
	testOUID     = "ou-123"
	testUserType = "employee"
)

func TestOAuthExecutorSuite(t *testing.T) {
	suite.Run(t, new(OAuthExecutorTestSuite))
}

func (suite *OAuthExecutorTestSuite) SetupTest() {
	suite.mockOAuthService = oauthmock.NewOAuthAuthnCoreServiceInterfaceMock(suite.T())
	suite.mockIDPService = idpmock.NewIDPServiceInterfaceMock(suite.T())
	suite.mockFlowFactory = coremock.NewFlowFactoryInterfaceMock(suite.T())
	suite.mockUserService = usermock.NewUserServiceInterfaceMock(suite.T())
	suite.mockUserSchemaService = userschemamock.NewUserSchemaServiceInterfaceMock(suite.T())

	defaultInputs := []flowcm.InputData{{Name: "code", Type: "string", Required: true}}
	mockExec := createMockAuthExecutor(suite.T(), ExecutorNameOAuth)
	suite.mockFlowFactory.On("CreateExecutor", ExecutorNameOAuth, flowcm.ExecutorTypeAuthentication,
		defaultInputs, []flowcm.InputData{}).Return(mockExec)

	suite.executor = newOAuthExecutor(ExecutorNameOAuth, defaultInputs, []flowcm.InputData{},
		suite.mockFlowFactory, suite.mockIDPService, suite.mockOAuthService,
		suite.mockUserService, suite.mockUserSchemaService)
}

func (suite *OAuthExecutorTestSuite) TestNewOAuthExecutor() {
	assert.NotNil(suite.T(), suite.executor)
}

func (suite *OAuthExecutorTestSuite) TestExecute_CodeNotProvided_BuildsAuthorizeURL() {
	ctx := &flowcore.NodeContext{
		FlowID:        "flow-123",
		FlowType:      flowcm.FlowTypeAuthentication,
		UserInputData: map[string]string{},
		NodeInputData: []flowcm.InputData{{Name: "code", Type: "string", Required: true}},
		NodeProperties: map[string]interface{}{
			"idpId": "idp-123",
		},
	}

	suite.mockOAuthService.On("BuildAuthorizeURL", "idp-123").
		Return("https://oauth.provider.com/authorize?client_id=abc", nil)

	suite.mockIDPService.On("GetIdentityProvider", "idp-123").
		Return(&idp.IDPDTO{ID: "idp-123", Name: "TestIDP"}, nil)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), flowcm.ExecExternalRedirection, resp.Status)
	assert.Equal(suite.T(), "https://oauth.provider.com/authorize?client_id=abc", resp.RedirectURL)
	assert.Equal(suite.T(), "TestIDP", resp.AdditionalData[flowcm.DataIDPName])
	suite.mockOAuthService.AssertExpectations(suite.T())
	suite.mockIDPService.AssertExpectations(suite.T())
}

func (suite *OAuthExecutorTestSuite) TestExecute_CodeProvided_AuthenticatesUser() {
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
		ExpiresIn:   3600,
	}

	userInfo := map[string]interface{}{
		"sub":   "user-sub-123",
		"email": "test@example.com",
		"name":  "Test User",
	}

	existingUser := &user.User{
		ID:               "user-123",
		OrganizationUnit: testOUID,
		Type:             "INTERNAL",
	}

	suite.mockOAuthService.On("ExchangeCodeForToken", "idp-123", "auth_code_123", true).
		Return(tokenResp, nil)
	suite.mockOAuthService.On("FetchUserInfo", "idp-123", "access_token_123").
		Return(userInfo, nil)
	suite.mockOAuthService.On("GetInternalUser", "user-sub-123").
		Return(existingUser, nil)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), flowcm.ExecComplete, resp.Status)
	assert.True(suite.T(), resp.AuthenticatedUser.IsAuthenticated)
	assert.Equal(suite.T(), "user-123", resp.AuthenticatedUser.UserID)
	assert.Equal(suite.T(), testOUID, resp.AuthenticatedUser.OrganizationUnitID)
	assert.Equal(suite.T(), "test@example.com", resp.RuntimeData["email"])
	suite.mockOAuthService.AssertExpectations(suite.T())
}

func (suite *OAuthExecutorTestSuite) TestBuildAuthorizeFlow_Success() {
	ctx := &flowcore.NodeContext{
		FlowID:   "flow-123",
		FlowType: flowcm.FlowTypeAuthentication,
		NodeProperties: map[string]interface{}{
			"idpId": "idp-123",
		},
	}

	execResp := &flowcm.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	suite.mockOAuthService.On("BuildAuthorizeURL", "idp-123").
		Return("https://oauth.provider.com/authorize", nil)
	suite.mockIDPService.On("GetIdentityProvider", "idp-123").
		Return(&idp.IDPDTO{ID: "idp-123", Name: "GoogleIDP"}, nil)

	err := suite.executor.BuildAuthorizeFlow(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), flowcm.ExecExternalRedirection, execResp.Status)
	assert.Equal(suite.T(), "https://oauth.provider.com/authorize", execResp.RedirectURL)
	assert.Equal(suite.T(), "GoogleIDP", execResp.AdditionalData[flowcm.DataIDPName])
	suite.mockOAuthService.AssertExpectations(suite.T())
	suite.mockIDPService.AssertExpectations(suite.T())
}

func (suite *OAuthExecutorTestSuite) TestBuildAuthorizeFlow_IDPNotConfigured() {
	ctx := &flowcore.NodeContext{
		FlowID:         "flow-123",
		FlowType:       flowcm.FlowTypeAuthentication,
		NodeProperties: map[string]interface{}{},
	}

	execResp := &flowcm.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	err := suite.executor.BuildAuthorizeFlow(ctx, execResp)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "idpId is not configured")
}

func (suite *OAuthExecutorTestSuite) TestBuildAuthorizeFlow_BuildURLClientError() {
	ctx := &flowcore.NodeContext{
		FlowID:   "flow-123",
		FlowType: flowcm.FlowTypeAuthentication,
		NodeProperties: map[string]interface{}{
			"idpId": "idp-123",
		},
	}

	execResp := &flowcm.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	suite.mockOAuthService.On("BuildAuthorizeURL", "idp-123").
		Return("", &serviceerror.ServiceError{
			Type:             serviceerror.ClientErrorType,
			ErrorDescription: "Invalid IDP configuration",
		})

	err := suite.executor.BuildAuthorizeFlow(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), flowcm.ExecFailure, execResp.Status)
	assert.Equal(suite.T(), "Invalid IDP configuration", execResp.FailureReason)
	suite.mockOAuthService.AssertExpectations(suite.T())
}

func (suite *OAuthExecutorTestSuite) TestBuildAuthorizeFlow_BuildURLServerError() {
	ctx := &flowcore.NodeContext{
		FlowID:   "flow-123",
		FlowType: flowcm.FlowTypeAuthentication,
		NodeProperties: map[string]interface{}{
			"idpId": "idp-123",
		},
	}

	execResp := &flowcm.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	suite.mockOAuthService.On("BuildAuthorizeURL", "idp-123").
		Return("", &serviceerror.ServiceError{
			Type:             serviceerror.ServerErrorType,
			Code:             "OAUTH-5000",
			ErrorDescription: "Internal server error",
		})

	err := suite.executor.BuildAuthorizeFlow(ctx, execResp)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to build authorize URL")
	suite.mockOAuthService.AssertExpectations(suite.T())
}

func (suite *OAuthExecutorTestSuite) TestExchangeCodeForToken_Success() {
	ctx := &flowcore.NodeContext{
		FlowID:   "flow-123",
		FlowType: flowcm.FlowTypeAuthentication,
		NodeProperties: map[string]interface{}{
			"idpId": "idp-123",
		},
	}

	execResp := &flowcm.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	tokenResp := &authnoauth.TokenResponse{
		AccessToken:  "access_token_123",
		TokenType:    "Bearer",
		Scope:        "openid profile",
		RefreshToken: "refresh_token_123",
		IDToken:      "id_token_123",
		ExpiresIn:    3600,
	}

	suite.mockOAuthService.On("ExchangeCodeForToken", "idp-123", "auth_code", true).
		Return(tokenResp, nil)

	result, err := suite.executor.ExchangeCodeForToken(ctx, execResp, "auth_code")

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "access_token_123", result.AccessToken)
	assert.Equal(suite.T(), "Bearer", result.TokenType)
	assert.Equal(suite.T(), "openid profile", result.Scope)
	assert.Equal(suite.T(), 3600, result.ExpiresIn)
	suite.mockOAuthService.AssertExpectations(suite.T())
}

func (suite *OAuthExecutorTestSuite) TestExchangeCodeForToken_ClientError() {
	ctx := &flowcore.NodeContext{
		FlowID:   "flow-123",
		FlowType: flowcm.FlowTypeAuthentication,
		NodeProperties: map[string]interface{}{
			"idpId": "idp-123",
		},
	}

	execResp := &flowcm.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	suite.mockOAuthService.On("ExchangeCodeForToken", "idp-123", "invalid_code", true).
		Return(nil, &serviceerror.ServiceError{
			Type:             serviceerror.ClientErrorType,
			ErrorDescription: "Invalid authorization code",
		})

	result, err := suite.executor.ExchangeCodeForToken(ctx, execResp, "invalid_code")

	assert.NoError(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), flowcm.ExecFailure, execResp.Status)
	assert.Equal(suite.T(), "Invalid authorization code", execResp.FailureReason)
	suite.mockOAuthService.AssertExpectations(suite.T())
}

func (suite *OAuthExecutorTestSuite) TestExchangeCodeForToken_ServerError() {
	ctx := &flowcore.NodeContext{
		FlowID:   "flow-123",
		FlowType: flowcm.FlowTypeAuthentication,
		NodeProperties: map[string]interface{}{
			"idpId": "idp-123",
		},
	}

	execResp := &flowcm.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	suite.mockOAuthService.On("ExchangeCodeForToken", "idp-123", "auth_code", true).
		Return(nil, &serviceerror.ServiceError{
			Type:             serviceerror.ServerErrorType,
			Code:             "OAUTH-5000",
			ErrorDescription: "Token exchange failed",
		})

	result, err := suite.executor.ExchangeCodeForToken(ctx, execResp, "auth_code")

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "failed to exchange code for token")
	suite.mockOAuthService.AssertExpectations(suite.T())
}

func (suite *OAuthExecutorTestSuite) TestGetUserInfo_Success() {
	ctx := &flowcore.NodeContext{
		FlowID:   "flow-123",
		FlowType: flowcm.FlowTypeAuthentication,
		NodeProperties: map[string]interface{}{
			"idpId": "idp-123",
		},
	}

	execResp := &flowcm.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	userInfo := map[string]interface{}{
		"sub":   "user-sub-123",
		"email": "test@example.com",
		"name":  "Test User",
	}

	suite.mockOAuthService.On("FetchUserInfo", "idp-123", "access_token").
		Return(userInfo, nil)

	result, err := suite.executor.GetUserInfo(ctx, execResp, "access_token")

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "user-sub-123", result["sub"])
	assert.Equal(suite.T(), "test@example.com", result["email"])
	assert.Equal(suite.T(), "Test User", result["name"])
	suite.mockOAuthService.AssertExpectations(suite.T())
}

func (suite *OAuthExecutorTestSuite) TestGetUserInfo_ClientError() {
	ctx := &flowcore.NodeContext{
		FlowID:   "flow-123",
		FlowType: flowcm.FlowTypeAuthentication,
		NodeProperties: map[string]interface{}{
			"idpId": "idp-123",
		},
	}

	execResp := &flowcm.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	suite.mockOAuthService.On("FetchUserInfo", "idp-123", "invalid_token").
		Return(nil, &serviceerror.ServiceError{
			Type:             serviceerror.ClientErrorType,
			ErrorDescription: "Invalid access token",
		})

	result, err := suite.executor.GetUserInfo(ctx, execResp, "invalid_token")

	assert.NoError(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), flowcm.ExecFailure, execResp.Status)
	assert.Equal(suite.T(), "Invalid access token", execResp.FailureReason)
	suite.mockOAuthService.AssertExpectations(suite.T())
}

func (suite *OAuthExecutorTestSuite) TestGetUserInfo_ServerError() {
	ctx := &flowcore.NodeContext{
		FlowID:   "flow-123",
		FlowType: flowcm.FlowTypeAuthentication,
		NodeProperties: map[string]interface{}{
			"idpId": "idp-123",
		},
	}

	execResp := &flowcm.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	suite.mockOAuthService.On("FetchUserInfo", "idp-123", "access_token").
		Return(nil, &serviceerror.ServiceError{
			Type:             serviceerror.ServerErrorType,
			Code:             "OAUTH-5000",
			ErrorDescription: "Failed to fetch user info",
		})

	result, err := suite.executor.GetUserInfo(ctx, execResp, "access_token")

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "failed to fetch user information")
	suite.mockOAuthService.AssertExpectations(suite.T())
}

func (suite *OAuthExecutorTestSuite) TestGetIdpID_Success() {
	ctx := &flowcore.NodeContext{
		FlowID:   "flow-123",
		FlowType: flowcm.FlowTypeAuthentication,
		NodeProperties: map[string]interface{}{
			"idpId": "idp-123",
		},
	}

	idpID, err := suite.executor.GetIdpID(ctx)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "idp-123", idpID)
}

func (suite *OAuthExecutorTestSuite) TestGetIdpID_NotConfigured() {
	ctx := &flowcore.NodeContext{
		FlowID:         "flow-123",
		FlowType:       flowcm.FlowTypeAuthentication,
		NodeProperties: map[string]interface{}{},
	}

	idpID, err := suite.executor.GetIdpID(ctx)

	assert.Error(suite.T(), err)
	assert.Empty(suite.T(), idpID)
	assert.Contains(suite.T(), err.Error(), "idpId is not configured")
}

func (suite *OAuthExecutorTestSuite) TestProcessAuthFlowResponse_RegistrationFlow_UserNotFound() {
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
		ExpiresIn:   3600,
	}

	userInfo := map[string]interface{}{
		"sub":   "new-user-sub",
		"email": "newuser@example.com",
		"name":  "New User",
	}

	suite.mockOAuthService.On("ExchangeCodeForToken", "idp-123", "auth_code_123", true).
		Return(tokenResp, nil)
	suite.mockOAuthService.On("FetchUserInfo", "idp-123", "access_token_123").
		Return(userInfo, nil)
	suite.mockOAuthService.On("GetInternalUser", "new-user-sub").
		Return(nil, &serviceerror.ServiceError{
			Code: authncm.ErrorUserNotFound.Code,
			Type: serviceerror.ClientErrorType,
		})

	err := suite.executor.ProcessAuthFlowResponse(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), flowcm.ExecComplete, execResp.Status)
	assert.False(suite.T(), execResp.AuthenticatedUser.IsAuthenticated)
	assert.Equal(suite.T(), "new-user-sub", execResp.RuntimeData["sub"])
	assert.NotNil(suite.T(), execResp.AuthenticatedUser.Attributes)
	suite.mockOAuthService.AssertExpectations(suite.T())
}

func (suite *OAuthExecutorTestSuite) TestProcessAuthFlowResponse_AuthFlow_UserNotFound() {
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
		ExpiresIn:   3600,
	}

	userInfo := map[string]interface{}{
		"sub":   "unknown-user",
		"email": "unknown@example.com",
	}

	suite.mockOAuthService.On("ExchangeCodeForToken", "idp-123", "auth_code_123", true).
		Return(tokenResp, nil)
	suite.mockOAuthService.On("FetchUserInfo", "idp-123", "access_token_123").
		Return(userInfo, nil)
	suite.mockOAuthService.On("GetInternalUser", "unknown-user").
		Return(nil, &serviceerror.ServiceError{
			Code: authncm.ErrorUserNotFound.Code,
			Type: serviceerror.ClientErrorType,
		})

	err := suite.executor.ProcessAuthFlowResponse(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), flowcm.ExecFailure, execResp.Status)
	assert.Equal(suite.T(), failureReasonUserNotFound, execResp.FailureReason)
	suite.mockOAuthService.AssertExpectations(suite.T())
}

func (suite *OAuthExecutorTestSuite) TestProcessAuthFlowResponse_UserAlreadyExists_RegistrationFlow() {
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
		Scope:       "openid profile",
		ExpiresIn:   3600,
	}

	userInfo := map[string]interface{}{
		"sub":   "existing-user-sub",
		"email": "existing@example.com",
	}

	existingUser := &user.User{
		ID:               "user-456",
		OrganizationUnit: "ou-456",
	}

	suite.mockOAuthService.On("ExchangeCodeForToken", "idp-123", "auth_code_123", true).
		Return(tokenResp, nil)
	suite.mockOAuthService.On("FetchUserInfo", "idp-123", "access_token_123").
		Return(userInfo, nil)
	suite.mockOAuthService.On("GetInternalUser", "existing-user-sub").
		Return(existingUser, nil)

	err := suite.executor.ProcessAuthFlowResponse(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), flowcm.ExecFailure, execResp.Status)
	assert.Contains(suite.T(), execResp.FailureReason, "User already exists")
	suite.mockOAuthService.AssertExpectations(suite.T())
}

func (suite *OAuthExecutorTestSuite) TestProcessAuthFlowResponse_NoCodeProvided() {
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

func (suite *OAuthExecutorTestSuite) TestProcessAuthFlowResponse_EmptyScope() {
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
		Scope:       "",
		ExpiresIn:   3600,
	}

	suite.mockOAuthService.On("ExchangeCodeForToken", "idp-123", "auth_code_123", true).
		Return(tokenResp, nil)

	err := suite.executor.ProcessAuthFlowResponse(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), flowcm.ExecFailure, execResp.Status)
	assert.False(suite.T(), execResp.AuthenticatedUser.IsAuthenticated)
	suite.mockOAuthService.AssertExpectations(suite.T())
}

func (suite *OAuthExecutorTestSuite) TestProcessAuthFlowResponse_NoSubClaim() {
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
		ExpiresIn:   3600,
	}

	userInfo := map[string]interface{}{
		"email": "test@example.com",
		"name":  "Test User",
	}

	suite.mockOAuthService.On("ExchangeCodeForToken", "idp-123", "auth_code_123", true).
		Return(tokenResp, nil)
	suite.mockOAuthService.On("FetchUserInfo", "idp-123", "access_token_123").
		Return(userInfo, nil)

	err := suite.executor.ProcessAuthFlowResponse(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), flowcm.ExecFailure, execResp.Status)
	assert.Contains(suite.T(), execResp.FailureReason, "sub claim not found")
	suite.mockOAuthService.AssertExpectations(suite.T())
}

func (suite *OAuthExecutorTestSuite) TestCheckInputData_CodeProvided() {
	ctx := &flowcore.NodeContext{
		FlowID:   "flow-123",
		FlowType: flowcm.FlowTypeAuthentication,
		UserInputData: map[string]string{
			"code": "auth_code_123",
		},
	}

	execResp := &flowcm.ExecutorResponse{
		RequiredData: []flowcm.InputData{},
	}

	result := suite.executor.CheckInputData(ctx, execResp)

	assert.False(suite.T(), result)
}

func (suite *OAuthExecutorTestSuite) TestCheckInputData_CodeNotProvided() {
	ctx := &flowcore.NodeContext{
		FlowID:        "flow-123",
		FlowType:      flowcm.FlowTypeAuthentication,
		UserInputData: map[string]string{},
		NodeInputData: []flowcm.InputData{{Name: "code", Type: "string", Required: true}},
	}

	execResp := &flowcm.ExecutorResponse{
		RequiredData: []flowcm.InputData{},
	}

	result := suite.executor.CheckInputData(ctx, execResp)

	assert.True(suite.T(), result)
	assert.NotEmpty(suite.T(), execResp.RequiredData)
}

func (suite *OAuthExecutorTestSuite) TestGetUserAttributes_WithEmail() {
	userInfo := map[string]string{
		"sub":      "user-sub-123",
		"email":    "test@example.com",
		"name":     "Test User",
		"username": "testuser",
	}

	execResp := &flowcm.ExecutorResponse{
		RuntimeData: make(map[string]string),
	}

	attributes := suite.executor.(*oAuthExecutor).getUserAttributes(userInfo, "user-123", execResp)

	assert.NotNil(suite.T(), attributes)
	assert.Equal(suite.T(), "test@example.com", attributes["email"])
	assert.Equal(suite.T(), "Test User", attributes["name"])
	assert.Equal(suite.T(), "user-123", attributes["user_id"])
	assert.NotContains(suite.T(), attributes, "sub")
	assert.NotContains(suite.T(), attributes, "username")
	assert.Equal(suite.T(), "test@example.com", execResp.RuntimeData["email"])
}

func (suite *OAuthExecutorTestSuite) TestGetUserAttributes_WithoutEmail() {
	userInfo := map[string]string{
		"sub":  "user-sub-123",
		"name": "Test User",
	}

	execResp := &flowcm.ExecutorResponse{
		RuntimeData: make(map[string]string),
	}

	attributes := suite.executor.(*oAuthExecutor).getUserAttributes(userInfo, "user-123", execResp)

	assert.NotNil(suite.T(), attributes)
	assert.Equal(suite.T(), "Test User", attributes["name"])
	assert.Equal(suite.T(), "user-123", attributes["user_id"])
	assert.NotContains(suite.T(), attributes, "email")
	assert.NotContains(suite.T(), execResp.RuntimeData, "email")
}

func (suite *OAuthExecutorTestSuite) TestGetUserAttributes_WithEmptyEmail() {
	userInfo := map[string]string{
		"sub":   "user-sub-123",
		"email": "",
		"name":  "Test User",
	}

	execResp := &flowcm.ExecutorResponse{
		RuntimeData: make(map[string]string),
	}

	attributes := suite.executor.(*oAuthExecutor).getUserAttributes(userInfo, "user-123", execResp)

	assert.NotNil(suite.T(), attributes)
	assert.Equal(suite.T(), "", attributes["email"])
	assert.NotContains(suite.T(), execResp.RuntimeData, "email")
}

func (suite *OAuthExecutorTestSuite) TestGetUserAttributes_WithoutUserID() {
	userInfo := map[string]string{
		"sub":   "user-sub-123",
		"email": "test@example.com",
		"name":  "Test User",
	}

	execResp := &flowcm.ExecutorResponse{
		RuntimeData: make(map[string]string),
	}

	attributes := suite.executor.(*oAuthExecutor).getUserAttributes(userInfo, "", execResp)

	assert.NotNil(suite.T(), attributes)
	assert.Equal(suite.T(), "test@example.com", attributes["email"])
	assert.NotContains(suite.T(), attributes, "user_id")
	assert.Equal(suite.T(), "test@example.com", execResp.RuntimeData["email"])
}

func (suite *OAuthExecutorTestSuite) TestProcessAuthFlowResponse_RegistrationFlow_WithEmail() {
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
		ExpiresIn:   3600,
	}

	userInfo := map[string]interface{}{
		"sub":   "new-user-sub",
		"email": "newuser@example.com",
		"name":  "New User",
	}

	suite.mockOAuthService.On("ExchangeCodeForToken", "idp-123", "auth_code_123", true).
		Return(tokenResp, nil)
	suite.mockOAuthService.On("FetchUserInfo", "idp-123", "access_token_123").
		Return(userInfo, nil)
	suite.mockOAuthService.On("GetInternalUser", "new-user-sub").
		Return(nil, &serviceerror.ServiceError{
			Code: authncm.ErrorUserNotFound.Code,
			Type: serviceerror.ClientErrorType,
		})

	err := suite.executor.ProcessAuthFlowResponse(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), flowcm.ExecComplete, execResp.Status)
	assert.False(suite.T(), execResp.AuthenticatedUser.IsAuthenticated)
	assert.Equal(suite.T(), "new-user-sub", execResp.RuntimeData["sub"])
	assert.Equal(suite.T(), "newuser@example.com", execResp.RuntimeData["email"])
	assert.Equal(suite.T(), "newuser@example.com", execResp.AuthenticatedUser.Attributes["email"])
	suite.mockOAuthService.AssertExpectations(suite.T())
}

func (suite *OAuthExecutorTestSuite) TestGetUserAttributes_WithEmail_NilRuntimeData() {
	userInfo := map[string]string{
		"sub":   "user-sub-123",
		"email": "test@example.com",
		"name":  "Test User",
	}

	execResp := &flowcm.ExecutorResponse{
		RuntimeData: nil, // Explicitly nil
	}

	attributes := suite.executor.(*oAuthExecutor).getUserAttributes(userInfo, "user-123", execResp)

	assert.NotNil(suite.T(), attributes)
	assert.Equal(suite.T(), "test@example.com", attributes["email"])
	assert.Equal(suite.T(), "Test User", attributes["name"])
	assert.Equal(suite.T(), "user-123", attributes["user_id"])
	assert.NotNil(suite.T(), execResp.RuntimeData, "RuntimeData should be initialized")
	assert.Equal(suite.T(), "test@example.com", execResp.RuntimeData["email"])
}

// Test provisioning functionality

func (suite *OAuthExecutorTestSuite) TestProcessAuthFlowResponse_AuthFlow_UserNotFound_ProvisioningSucceeds() {
	ctx := &flowcore.NodeContext{
		FlowID:   "flow-123",
		FlowType: flowcm.FlowTypeAuthentication,
		Application: appmodel.Application{
			AllowedUserTypes: []string{testUserType},
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
		Scope:       "openid profile",
		ExpiresIn:   3600,
	}

	userInfo := map[string]interface{}{
		"sub":   "new-user-sub",
		"email": "newuser@example.com",
		"name":  "New User",
	}

	userSchema := &userschema.UserSchema{
		ID:                    "schema-123",
		Name:                  testUserType,
		OrganizationUnitID:    testOUID,
		AllowSelfRegistration: true,
	}

	provisionedUser := &user.User{
		ID:               "user-provisioned-123",
		OrganizationUnit: testOUID,
		Type:             testUserType,
	}

	suite.mockOAuthService.On("ExchangeCodeForToken", "idp-123", "auth_code_123", true).
		Return(tokenResp, nil)
	suite.mockOAuthService.On("FetchUserInfo", "idp-123", "access_token_123").
		Return(userInfo, nil)
	suite.mockOAuthService.On("GetInternalUser", "new-user-sub").
		Return(nil, &serviceerror.ServiceError{
			Code: authncm.ErrorUserNotFound.Code,
			Type: serviceerror.ClientErrorType,
		})
	suite.mockUserSchemaService.On("GetUserSchemaByName", testUserType).
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
	assert.Equal(suite.T(), testUserType, execResp.AuthenticatedUser.UserType)
	suite.mockOAuthService.AssertExpectations(suite.T())
	suite.mockUserSchemaService.AssertExpectations(suite.T())
	suite.mockUserService.AssertExpectations(suite.T())
}

func (suite *OAuthExecutorTestSuite) TestProcessAuthFlowResponse_AuthFlow_UserNotFound_NoAllowedUserTypes() {
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
		Scope:       "openid profile",
		ExpiresIn:   3600,
	}

	userInfo := map[string]interface{}{
		"sub":   "new-user-sub",
		"email": "newuser@example.com",
	}

	suite.mockOAuthService.On("ExchangeCodeForToken", "idp-123", "auth_code_123", true).
		Return(tokenResp, nil)
	suite.mockOAuthService.On("FetchUserInfo", "idp-123", "access_token_123").
		Return(userInfo, nil)
	suite.mockOAuthService.On("GetInternalUser", "new-user-sub").
		Return(nil, &serviceerror.ServiceError{
			Code: authncm.ErrorUserNotFound.Code,
			Type: serviceerror.ClientErrorType,
		})

	err := suite.executor.ProcessAuthFlowResponse(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), flowcm.ExecFailure, execResp.Status)
	assert.Equal(suite.T(), failureReasonUserNotFound, execResp.FailureReason)
	suite.mockOAuthService.AssertExpectations(suite.T())
	suite.mockUserSchemaService.AssertNotCalled(suite.T(), "GetUserSchemaByName")
	suite.mockUserService.AssertNotCalled(suite.T(), "CreateUser")
}

func (suite *OAuthExecutorTestSuite) testUserNotFoundWithSchemaError(schemaError *serviceerror.ServiceError) {
	ctx := &flowcore.NodeContext{
		FlowID:   "flow-123",
		FlowType: flowcm.FlowTypeAuthentication,
		Application: appmodel.Application{
			AllowedUserTypes: []string{testUserType},
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
		Scope:       "openid profile",
		ExpiresIn:   3600,
	}

	userInfo := map[string]interface{}{
		"sub":   "new-user-sub",
		"email": "newuser@example.com",
	}

	suite.mockOAuthService.On("ExchangeCodeForToken", "idp-123", "auth_code_123", true).
		Return(tokenResp, nil)
	suite.mockOAuthService.On("FetchUserInfo", "idp-123", "access_token_123").
		Return(userInfo, nil)
	suite.mockOAuthService.On("GetInternalUser", "new-user-sub").
		Return(nil, &serviceerror.ServiceError{
			Code: authncm.ErrorUserNotFound.Code,
			Type: serviceerror.ClientErrorType,
		})
	suite.mockUserSchemaService.On("GetUserSchemaByName", testUserType).
		Return(nil, schemaError)

	err := suite.executor.ProcessAuthFlowResponse(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), flowcm.ExecFailure, execResp.Status)
	assert.Equal(suite.T(), "User not found and automatic provisioning is not available", execResp.FailureReason)
	suite.mockOAuthService.AssertExpectations(suite.T())
	suite.mockUserSchemaService.AssertExpectations(suite.T())
	suite.mockUserService.AssertNotCalled(suite.T(), "CreateUser")
}

func (suite *OAuthExecutorTestSuite) TestProcessAuthFlowResponse_AuthFlow_UserNotFound_UserSchemaNotFound() {
	suite.testUserNotFoundWithSchemaError(&serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "USRS-1002",
		ErrorDescription: "User schema not found",
	})
}

func (suite *OAuthExecutorTestSuite) TestProcessAuthFlowResponse_AuthFlow_UserNotFound_UserSchemaServiceError() {
	suite.testUserNotFoundWithSchemaError(&serviceerror.ServiceError{
		Type:             serviceerror.ServerErrorType,
		Code:             "USRS-5000",
		ErrorDescription: "Internal server error",
	})
}

func (suite *OAuthExecutorTestSuite) TestProcessAuthFlowResponse_AuthFlow_UserNotFound_UserCreationFails() {
	ctx := &flowcore.NodeContext{
		FlowID:   "flow-123",
		FlowType: flowcm.FlowTypeAuthentication,
		Application: appmodel.Application{
			AllowedUserTypes: []string{testUserType},
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
		Scope:       "openid profile",
		ExpiresIn:   3600,
	}

	userInfo := map[string]interface{}{
		"sub":   "new-user-sub",
		"email": "newuser@example.com",
	}

	userSchema := &userschema.UserSchema{
		ID:                    "schema-123",
		Name:                  testUserType,
		OrganizationUnitID:    testOUID,
		AllowSelfRegistration: true,
	}

	suite.mockOAuthService.On("ExchangeCodeForToken", "idp-123", "auth_code_123", true).
		Return(tokenResp, nil)
	suite.mockOAuthService.On("FetchUserInfo", "idp-123", "access_token_123").
		Return(userInfo, nil)
	suite.mockOAuthService.On("GetInternalUser", "new-user-sub").
		Return(nil, &serviceerror.ServiceError{
			Code: authncm.ErrorUserNotFound.Code,
			Type: serviceerror.ClientErrorType,
		})
	suite.mockUserSchemaService.On("GetUserSchemaByName", testUserType).
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
	suite.mockOAuthService.AssertExpectations(suite.T())
	suite.mockUserSchemaService.AssertExpectations(suite.T())
	suite.mockUserService.AssertExpectations(suite.T())
}

func (suite *OAuthExecutorTestSuite) TestProcessAuthFlowResponse_AuthFlow_UserNotFound_MultipleAllowedUserTypes() {
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
		Scope:       "openid profile",
		ExpiresIn:   3600,
	}

	userInfo := map[string]interface{}{
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

	suite.mockOAuthService.On("ExchangeCodeForToken", "idp-123", "auth_code_123", true).
		Return(tokenResp, nil)
	suite.mockOAuthService.On("FetchUserInfo", "idp-123", "access_token_123").
		Return(userInfo, nil)
	suite.mockOAuthService.On("GetInternalUser", "new-user-sub").
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
	suite.mockOAuthService.AssertExpectations(suite.T())
	suite.mockUserSchemaService.AssertExpectations(suite.T())
	suite.mockUserService.AssertNotCalled(suite.T(), "CreateUser")
}

func (suite *OAuthExecutorTestSuite) TestProcessAuthFlowResponse_AuthFlow_UserNotFound_SelfRegistrationDisabled() {
	ctx := &flowcore.NodeContext{
		FlowID:   "flow-123",
		FlowType: flowcm.FlowTypeAuthentication,
		Application: appmodel.Application{
			AllowedUserTypes: []string{testUserType},
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
		Scope:       "openid profile",
		ExpiresIn:   3600,
	}

	userInfo := map[string]interface{}{
		"sub":   "new-user-sub",
		"email": "newuser@example.com",
	}

	userSchema := &userschema.UserSchema{
		ID:                    "schema-123",
		Name:                  testUserType,
		OrganizationUnitID:    testOUID,
		AllowSelfRegistration: false,
	}

	suite.mockOAuthService.On("ExchangeCodeForToken", "idp-123", "auth_code_123", true).
		Return(tokenResp, nil)
	suite.mockOAuthService.On("FetchUserInfo", "idp-123", "access_token_123").
		Return(userInfo, nil)
	suite.mockOAuthService.On("GetInternalUser", "new-user-sub").
		Return(nil, &serviceerror.ServiceError{
			Code: authncm.ErrorUserNotFound.Code,
			Type: serviceerror.ClientErrorType,
		})
	suite.mockUserSchemaService.On("GetUserSchemaByName", testUserType).
		Return(userSchema, nil)

	err := suite.executor.ProcessAuthFlowResponse(ctx, execResp)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), flowcm.ExecFailure, execResp.Status)
	assert.Equal(suite.T(), "User not found and automatic provisioning is not available", execResp.FailureReason)
	suite.mockOAuthService.AssertExpectations(suite.T())
	suite.mockUserSchemaService.AssertExpectations(suite.T())
	suite.mockUserService.AssertNotCalled(suite.T(), "CreateUser")
}

func (suite *OAuthExecutorTestSuite) TestProcessAuthFlowResponse_AuthFlow_UserNotFound_MultipleTypes_OneWithSelfReg() {
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
		Scope:       "openid profile",
		ExpiresIn:   3600,
	}

	userInfo := map[string]interface{}{
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

	suite.mockOAuthService.On("ExchangeCodeForToken", "idp-123", "auth_code_123", true).
		Return(tokenResp, nil)
	suite.mockOAuthService.On("FetchUserInfo", "idp-123", "access_token_123").
		Return(userInfo, nil)
	suite.mockOAuthService.On("GetInternalUser", "new-user-sub").
		Return(nil, &serviceerror.ServiceError{
			Code: authncm.ErrorUserNotFound.Code,
			Type: serviceerror.ClientErrorType,
		})
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
	suite.mockOAuthService.AssertExpectations(suite.T())
	suite.mockUserSchemaService.AssertExpectations(suite.T())
	suite.mockUserService.AssertExpectations(suite.T())
}
