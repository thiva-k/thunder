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
	"errors"
	"fmt"

	authncm "github.com/asgardeo/thunder/internal/authn/common"
	authnoauth "github.com/asgardeo/thunder/internal/authn/oauth"
	flowcm "github.com/asgardeo/thunder/internal/flow/common"
	flowcore "github.com/asgardeo/thunder/internal/flow/core"
	"github.com/asgardeo/thunder/internal/idp"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	systemutils "github.com/asgardeo/thunder/internal/system/utils"
)

const (
	oAuthLoggerComponentName = "OAuthExecutor"
)

// OAuthTokenResponse represents the response from a OAuth token endpoint.
type OAuthTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token"`
	ExpiresIn    int    `json:"expires_in"`
}

// oAuthExecutorInterface defines the interface for OAuth authentication executors.
type oAuthExecutorInterface interface {
	flowcore.ExecutorInterface
	BuildAuthorizeFlow(ctx *flowcore.NodeContext, execResp *flowcm.ExecutorResponse) error
	ProcessAuthFlowResponse(ctx *flowcore.NodeContext, execResp *flowcm.ExecutorResponse) error
	ExchangeCodeForToken(ctx *flowcore.NodeContext, execResp *flowcm.ExecutorResponse,
		code string) (*OAuthTokenResponse, error)
	GetUserInfo(ctx *flowcore.NodeContext, execResp *flowcm.ExecutorResponse,
		accessToken string) (map[string]string, error)
	GetIdpID(ctx *flowcore.NodeContext) (string, error)
}

// oAuthExecutor implements the OAuthExecutorInterface for handling generic OAuth authentication flows.
type oAuthExecutor struct {
	flowcore.ExecutorInterface
	authService authnoauth.OAuthAuthnCoreServiceInterface
	idpService  idp.IDPServiceInterface
	logger      *log.Logger
}

var _ flowcore.ExecutorInterface = (*oAuthExecutor)(nil)

// newOAuthExecutor creates a new instance of OAuthExecutor.
func newOAuthExecutor(
	name string,
	defaultInputs, prerequisites []flowcm.InputData,
	flowFactory flowcore.FlowFactoryInterface,
	idpService idp.IDPServiceInterface,
	authService authnoauth.OAuthAuthnCoreServiceInterface,
) oAuthExecutorInterface {
	if name == "" {
		name = ExecutorNameOAuth
	}
	if len(defaultInputs) == 0 {
		defaultInputs = []flowcm.InputData{
			{
				Name:     "code",
				Type:     "string",
				Required: true,
			},
		}
	}
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, oAuthLoggerComponentName),
		log.String(log.LoggerKeyExecutorName, name))

	base := flowFactory.CreateExecutor(name, flowcm.ExecutorTypeAuthentication,
		defaultInputs, prerequisites)

	return &oAuthExecutor{
		ExecutorInterface: base,
		authService:       authService,
		idpService:        idpService,
		logger:            logger,
	}
}

// Execute executes the OAuth authentication flow.
func (o *oAuthExecutor) Execute(ctx *flowcore.NodeContext) (*flowcm.ExecutorResponse, error) {
	logger := o.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Executing OAuth authentication executor")

	execResp := &flowcm.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	// Check if the required input data is provided
	if o.CheckInputData(ctx, execResp) {
		// If required input data is not provided, return incomplete status with redirection.
		logger.Debug("Required input data for OAuth authentication executor is not provided")
		err := o.BuildAuthorizeFlow(ctx, execResp)
		if err != nil {
			return nil, err
		}
	} else {
		err := o.ProcessAuthFlowResponse(ctx, execResp)
		if err != nil {
			return nil, err
		}
	}

	logger.Debug("OAuth authentication executor execution completed",
		log.String("status", string(execResp.Status)),
		log.Bool("isAuthenticated", execResp.AuthenticatedUser.IsAuthenticated))

	return execResp, nil
}

// BuildAuthorizeFlow constructs the redirection to the external OAuth provider for user authentication.
func (o *oAuthExecutor) BuildAuthorizeFlow(ctx *flowcore.NodeContext, execResp *flowcm.ExecutorResponse) error {
	logger := o.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Initiating OAuth authentication flow")

	idpID, err := o.GetIdpID(ctx)
	if err != nil {
		return err
	}

	authorizeURL, svcErr := o.authService.BuildAuthorizeURL(idpID)
	if svcErr != nil {
		if svcErr.Type == serviceerror.ClientErrorType {
			execResp.Status = flowcm.ExecFailure
			execResp.FailureReason = svcErr.ErrorDescription
			return nil
		}

		logger.Error("Failed to build authorize URL", log.String("errorCode", svcErr.Code),
			log.String("errorDescription", svcErr.ErrorDescription))
		return errors.New("failed to build authorize URL")
	}

	// Get the idp name for additional data
	idpName, err := o.getIDPName(idpID)
	if err != nil {
		return fmt.Errorf("failed to get idp name: %w", err)
	}

	// Set the response to redirect the user to the authorization URL.
	execResp.Status = flowcm.ExecExternalRedirection
	execResp.RedirectURL = authorizeURL
	execResp.AdditionalData = map[string]string{
		flowcm.DataIDPName: idpName,
	}

	return nil
}

// ProcessAuthFlowResponse processes the response from the OAuth authentication flow and authenticates the user.
func (o *oAuthExecutor) ProcessAuthFlowResponse(ctx *flowcore.NodeContext,
	execResp *flowcm.ExecutorResponse) error {
	logger := o.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Processing OAuth authentication response")

	code, ok := ctx.UserInputData["code"]
	if ok && code != "" {
		tokenResp, err := o.ExchangeCodeForToken(ctx, execResp, code)
		if err != nil {
			return err
		}
		if execResp.Status == flowcm.ExecFailure {
			return nil
		}

		if tokenResp.Scope == "" {
			logger.Error("Scopes are empty in the token response")
			execResp.AuthenticatedUser = authncm.AuthenticatedUser{
				IsAuthenticated: false,
			}
		} else {
			authenticatedUser, err := o.getAuthenticatedUserWithAttributes(ctx, execResp, tokenResp.AccessToken)
			if err != nil {
				return err
			}
			if authenticatedUser == nil {
				return nil
			}
			execResp.AuthenticatedUser = *authenticatedUser
		}
	} else {
		execResp.AuthenticatedUser = authncm.AuthenticatedUser{
			IsAuthenticated: false,
		}
	}

	if execResp.AuthenticatedUser.IsAuthenticated {
		execResp.Status = flowcm.ExecComplete
	} else if ctx.FlowType != flowcm.FlowTypeRegistration {
		execResp.Status = flowcm.ExecFailure
		execResp.FailureReason = failureReasonInvalidAuthorizationCode
		return nil
	}

	return nil
}

// CheckInputData checks if the required input data is provided in the context.
func (o *oAuthExecutor) CheckInputData(ctx *flowcore.NodeContext, execResp *flowcm.ExecutorResponse) bool {
	if code, ok := ctx.UserInputData["code"]; ok && code != "" {
		return false
	}

	return o.ExecutorInterface.CheckInputData(ctx, execResp)
}

// ExchangeCodeForToken exchanges the authorization code for an access token.
func (o *oAuthExecutor) ExchangeCodeForToken(ctx *flowcore.NodeContext, execResp *flowcm.ExecutorResponse,
	code string) (*OAuthTokenResponse, error) {
	logger := o.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Exchanging authorization code for a token")

	idpID, err := o.GetIdpID(ctx)
	if err != nil {
		return nil, err
	}

	tokenResp, svcErr := o.authService.ExchangeCodeForToken(idpID, code, true)
	if svcErr != nil {
		if svcErr.Type == serviceerror.ClientErrorType {
			execResp.Status = flowcm.ExecFailure
			execResp.FailureReason = svcErr.ErrorDescription
			return nil, nil
		}

		logger.Error("Failed to exchange code for a token", log.String("errorCode", svcErr.Code),
			log.String("errorDescription", svcErr.ErrorDescription))
		return nil, errors.New("failed to exchange code for token")
	}

	return &OAuthTokenResponse{
		AccessToken:  tokenResp.AccessToken,
		TokenType:    tokenResp.TokenType,
		Scope:        tokenResp.Scope,
		RefreshToken: tokenResp.RefreshToken,
		IDToken:      tokenResp.IDToken,
		ExpiresIn:    tokenResp.ExpiresIn,
	}, nil
}

// GetUserInfo fetches user information from the OAuth provider using the access token.
func (o *oAuthExecutor) GetUserInfo(ctx *flowcore.NodeContext, execResp *flowcm.ExecutorResponse,
	accessToken string) (map[string]string, error) {
	logger := o.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Fetching user info from OAuth provider")

	idpID, err := o.GetIdpID(ctx)
	if err != nil {
		return nil, err
	}

	userInfo, svcErr := o.authService.FetchUserInfo(idpID, accessToken)
	if svcErr != nil {
		if svcErr.Type == serviceerror.ClientErrorType {
			execResp.Status = flowcm.ExecFailure
			execResp.FailureReason = svcErr.ErrorDescription
			return nil, nil
		}

		logger.Error("Failed to fetch user info", log.String("errorCode", svcErr.Code),
			log.String("errorDescription", svcErr.ErrorDescription))
		return nil, errors.New("failed to fetch user information")
	}

	return systemutils.ConvertInterfaceMapToStringMap(userInfo), nil
}

// GetIdpID retrieves the identity provider ID from the node properties.
func (o *oAuthExecutor) GetIdpID(ctx *flowcore.NodeContext) (string, error) {
	if len(ctx.NodeProperties) > 0 {
		if val, ok := ctx.NodeProperties["idpId"]; ok {
			if idpID, valid := val.(string); valid && idpID != "" {
				return idpID, nil
			}
		}
	}
	return "", errors.New("idpId is not configured in node properties")
}

// getIDPName retrieves the name of the identity provider using its ID.
func (o *oAuthExecutor) getIDPName(idpID string) (string, error) {
	logger := o.logger
	logger.Debug("Retrieving IDP name for the given IDP ID")

	idp, svcErr := o.idpService.GetIdentityProvider(idpID)
	if svcErr != nil {
		if svcErr.Type == serviceerror.ClientErrorType {
			return "", fmt.Errorf("failed to get identity provider: %s", svcErr.ErrorDescription)
		}

		logger.Error("Error while retrieving identity provider", log.String("errorCode", svcErr.Code),
			log.String("errorDescription", svcErr.ErrorDescription))
		return "", errors.New("error while retrieving identity provider")
	}

	return idp.Name, nil
}

// getAuthenticatedUserWithAttributes retrieves the authenticated user information with additional attributes
// from the OAuth provider using the access token.
func (o *oAuthExecutor) getAuthenticatedUserWithAttributes(ctx *flowcore.NodeContext,
	execResp *flowcm.ExecutorResponse, accessToken string) (*authncm.AuthenticatedUser, error) {
	logger := o.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))

	// Get user info using the access token
	userInfo, err := o.GetUserInfo(ctx, execResp, accessToken)
	if err != nil {
		return nil, err
	}
	if execResp.Status == flowcm.ExecFailure {
		return nil, nil
	}

	// Resolve user with the sub claim.
	sub, ok := userInfo["sub"]
	if !ok || sub == "" {
		execResp.Status = flowcm.ExecFailure
		execResp.FailureReason = "sub claim not found in the response."
		return nil, nil
	}

	user, svcErr := o.authService.GetInternalUser(sub)
	if svcErr != nil {
		if svcErr.Code == authncm.ErrorUserNotFound.Code {
			if ctx.FlowType == flowcm.FlowTypeRegistration {
				logger.Debug("User not found for the provided sub claim. Proceeding with registration flow.")
				execResp.Status = flowcm.ExecComplete
				execResp.FailureReason = ""

				if execResp.RuntimeData == nil {
					execResp.RuntimeData = make(map[string]string)
				}
				execResp.RuntimeData["sub"] = sub

				return &authncm.AuthenticatedUser{
					IsAuthenticated: false,
					Attributes:      o.getUserAttributes(userInfo, "", execResp),
				}, nil
			} else {
				execResp.Status = flowcm.ExecFailure
				execResp.FailureReason = "User not found"
				return nil, nil
			}
		} else {
			if svcErr.Type == serviceerror.ClientErrorType {
				execResp.Status = flowcm.ExecFailure
				execResp.FailureReason = svcErr.ErrorDescription
				return nil, nil
			}
			logger.Error("Error while retrieving internal user", log.String("errorCode", svcErr.Code),
				log.String("description", svcErr.ErrorDescription))
			return nil, errors.New("error while retrieving internal user")
		}
	}

	if ctx.FlowType == flowcm.FlowTypeRegistration {
		// At this point, a unique user is found in the system. Hence fail the execution.
		execResp.Status = flowcm.ExecFailure
		execResp.FailureReason = "User already exists with the provided sub claim."
		return nil, nil
	}

	if user == nil || user.ID == "" {
		return nil, errors.New("retrieved user is nil or has an empty ID")
	}
	userID := user.ID

	if execResp.Status == flowcm.ExecFailure {
		return nil, nil
	}

	authenticatedUser := authncm.AuthenticatedUser{
		IsAuthenticated:    true,
		UserID:             userID,
		OrganizationUnitID: user.OrganizationUnit,
		UserType:           user.Type,
		Attributes:         o.getUserAttributes(userInfo, userID, execResp),
	}

	return &authenticatedUser, nil
}

// getUserAttributes extracts user attributes from the user info map, excluding certain keys.
// TODO: Need to convert attributes as per the IDP to local attribute mapping when the support is implemented.
func (o *oAuthExecutor) getUserAttributes(userInfo map[string]string, userID string,
	execResp *flowcm.ExecutorResponse) map[string]interface{} {
	attributes := make(map[string]interface{})
	for key, value := range userInfo {
		if key != "username" && key != "sub" {
			attributes[key] = value
		}
	}
	if userID != "" {
		attributes["user_id"] = userID
	}

	// Append email to runtime data if available.
	if email, ok := attributes["email"]; ok {
		if emailStr, ok := email.(string); ok && emailStr != "" {
			if execResp.RuntimeData == nil {
				execResp.RuntimeData = make(map[string]string)
			}
			execResp.RuntimeData["email"] = emailStr
		}
	}

	return attributes
}
