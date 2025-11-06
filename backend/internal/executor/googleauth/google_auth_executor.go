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

// Package googleauth provides the Google OIDC authentication executor.
package googleauth

import (
	"errors"
	"fmt"
	"slices"

	authncm "github.com/asgardeo/thunder/internal/authn/common"
	authngoogle "github.com/asgardeo/thunder/internal/authn/google"
	authnoidc "github.com/asgardeo/thunder/internal/authn/oidc"
	"github.com/asgardeo/thunder/internal/executor/oauth"
	"github.com/asgardeo/thunder/internal/executor/oidcauth"
	flowcm "github.com/asgardeo/thunder/internal/flow/common"
	flowmodel "github.com/asgardeo/thunder/internal/flow/common/model"
	"github.com/asgardeo/thunder/internal/idp"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/jwt"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/user"
)

const (
	executorName        = authncm.AuthenticatorGoogle
	loggerComponentName = "GoogleOIDCAuthExecutor"
)

// GoogleOIDCAuthExecutor implements the OIDC authentication executor for Google.
type GoogleOIDCAuthExecutor struct {
	oidcauth.OIDCAuthExecutorInterface
	googleAuthService authngoogle.GoogleOIDCAuthnServiceInterface
}

var _ flowmodel.ExecutorInterface = (*GoogleOIDCAuthExecutor)(nil)

// NewGoogleOIDCAuthExecutor creates a new instance of GoogleOIDCAuthExecutor with the provided details.
func NewGoogleOIDCAuthExecutor() oidcauth.OIDCAuthExecutorInterface {
	defaultInputs := []flowmodel.InputData{
		{
			Name:     "code",
			Type:     "string",
			Required: true,
		},
		{
			Name:     "nonce",
			Type:     "string",
			Required: false,
		},
	}

	// TODO: Should be injected when moving executors to di pattern.
	idpSvc := idp.NewIDPService()
	userSvc := user.GetUserService()
	jwtSvc := jwt.GetJWTService()
	googleAuthSvc := authngoogle.NewGoogleOIDCAuthnService(idpSvc, userSvc, jwtSvc)

	oidcSvcCast, ok := googleAuthSvc.(authnoidc.OIDCAuthnCoreServiceInterface)
	if !ok {
		panic("failed to cast GoogleOIDCAuthnService to OIDCAuthnCoreServiceInterface")
	}

	base := oidcauth.NewOIDCAuthExecutorWithServices(executorName, defaultInputs,
		[]flowmodel.InputData{}, oidcSvcCast, idpSvc)

	return &GoogleOIDCAuthExecutor{
		OIDCAuthExecutorInterface: base,
		googleAuthService:         googleAuthSvc,
	}
}

// Execute executes the Google OIDC authentication flow.
func (g *GoogleOIDCAuthExecutor) Execute(ctx *flowmodel.NodeContext) (*flowmodel.ExecutorResponse, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Executing Google OIDC auth executor",
		log.String(log.LoggerKeyExecutorName, g.GetName()), log.String("flowID", ctx.FlowID))

	execResp := &flowmodel.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	if g.CheckInputData(ctx, execResp) {
		logger.Debug("Required input data for Google OIDC auth executor is not provided")
		err := g.BuildAuthorizeFlow(ctx, execResp)
		if err != nil {
			return nil, err
		}

		logger.Debug("Google OIDC auth executor execution completed",
			log.String("status", string(execResp.Status)))
	} else {
		err := g.ProcessAuthFlowResponse(ctx, execResp)
		if err != nil {
			return nil, err
		}

		logger.Debug("Google OIDC auth executor execution completed",
			log.String("status", string(execResp.Status)),
			log.Bool("isAuthenticated", execResp.AuthenticatedUser.IsAuthenticated))
	}

	return execResp, nil
}

// ProcessAuthFlowResponse processes the response from the Google authentication flow and authenticates the user.
// This method has been overridden to handle Google-specific logic in id token validation.
func (g *GoogleOIDCAuthExecutor) ProcessAuthFlowResponse(ctx *flowmodel.NodeContext,
	execResp *flowmodel.ExecutorResponse) error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName),
		log.String(log.LoggerKeyExecutorName, g.GetName()),
		log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Processing Google OIDC authentication response")

	code, ok := ctx.UserInputData["code"]
	if ok && code != "" {
		tokenResp, err := g.ExchangeCodeForToken(ctx, execResp, code)
		if err != nil {
			logger.Error("Failed to exchange code for a token", log.Error(err))
			return fmt.Errorf("failed to exchange code for token: %w", err)
		}
		if execResp.Status == flowcm.ExecFailure {
			return nil
		}

		idTokenClaims, err := g.GetIDTokenClaims(execResp, tokenResp.IDToken)
		if err != nil {
			return errors.New("failed to extract ID token claims: " + err.Error())
		}
		if execResp.Status == flowcm.ExecFailure {
			return nil
		}

		// Validate nonce if configured.
		if nonce, ok := ctx.UserInputData["nonce"]; ok && nonce != "" {
			if idTokenClaims["nonce"] != nonce {
				execResp.Status = flowcm.ExecFailure
				execResp.FailureReason = "Nonce mismatch in ID token claims."
				return nil
			}
		}

		// Resolve user with the sub claim.
		// TODO: For now assume `sub` is the unique identifier for the user always.
		parsedSub := ""
		sub, ok := idTokenClaims["sub"]
		if ok && sub != "" {
			if subStr, ok := sub.(string); ok && subStr != "" {
				parsedSub = subStr
			}
		}
		if parsedSub == "" {
			execResp.Status = flowcm.ExecFailure
			execResp.FailureReason = "sub claim not found in the ID token."
			return nil
		}

		user, err := g.resolveUser(parsedSub, ctx, execResp)
		if err != nil {
			return err
		}
		if execResp.Status == flowcm.ExecFailure {
			return nil
		}

		authenticatedUser, err := g.getAuthenticatedUserWithAttributes(ctx, execResp,
			tokenResp.AccessToken, idTokenClaims, user)
		if err != nil {
			return err
		}
		if execResp.Status == flowcm.ExecFailure || authenticatedUser == nil {
			return nil
		}
		execResp.AuthenticatedUser = *authenticatedUser
	} else {
		execResp.AuthenticatedUser = authncm.AuthenticatedUser{
			IsAuthenticated: false,
		}
	}

	if execResp.AuthenticatedUser.IsAuthenticated {
		execResp.Status = flowcm.ExecComplete
	} else if ctx.FlowType != flowcm.FlowTypeRegistration {
		execResp.Status = flowcm.ExecFailure
		execResp.FailureReason = "Authentication failed. Authorization code not provided or invalid."
		return nil
	}

	return nil
}

// ExchangeCodeForToken exchanges the authorization code for an access token.
// This method has been overridden to handle Google-specific logic in id token validation.
func (g *GoogleOIDCAuthExecutor) ExchangeCodeForToken(ctx *flowmodel.NodeContext,
	execResp *flowmodel.ExecutorResponse, code string) (*oauth.TokenResponse, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName),
		log.String(log.LoggerKeyExecutorName, g.GetName()),
		log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Exchanging authorization code for a token")

	idpID, err := g.GetIdpID(ctx)
	if err != nil {
		return nil, err
	}

	tokenResp, svcErr := g.googleAuthService.ExchangeCodeForToken(idpID, code, true)
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

	return &oauth.TokenResponse{
		AccessToken:  tokenResp.AccessToken,
		TokenType:    tokenResp.TokenType,
		Scope:        tokenResp.Scope,
		RefreshToken: tokenResp.RefreshToken,
		IDToken:      tokenResp.IDToken,
		ExpiresIn:    tokenResp.ExpiresIn,
	}, nil
}

// getAuthenticatedUserWithAttributes constructs the authenticated user object with attributes from the
// ID token and user info.
func (g *GoogleOIDCAuthExecutor) getAuthenticatedUserWithAttributes(ctx *flowmodel.NodeContext,
	execResp *flowmodel.ExecutorResponse, accessToken string, idTokenClaims map[string]interface{},
	user *user.User) (*authncm.AuthenticatedUser, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName),
		log.String(log.LoggerKeyExecutorName, g.GetName()),
		log.String(log.LoggerKeyFlowID, ctx.FlowID))

	userClaims := make(map[string]interface{})
	if len(idTokenClaims) != 0 {
		// Filter non-user claims from the ID token claims.
		for attr, val := range idTokenClaims {
			if !slices.Contains(oidcauth.IDTokenNonUserAttributes, attr) {
				userClaims[attr] = val
			}
		}
		logger.Debug("Extracted ID token claims", log.Int("noOfClaims", len(idTokenClaims)))
	}

	idpID, err := g.GetIdpID(ctx)
	if err != nil {
		return nil, err
	}

	oauthConfigs, svcErr := g.googleAuthService.GetOAuthClientConfig(idpID)
	if svcErr != nil {
		if svcErr.Type == serviceerror.ClientErrorType {
			execResp.Status = flowcm.ExecFailure
			execResp.FailureReason = fmt.Sprintf("failed to retrieve OAuth client configuration: %s",
				svcErr.ErrorDescription)
			return nil, nil
		}

		logger.Error("Failed to retrieve OAuth client configuration", log.String("errorCode", svcErr.Code),
			log.String("errorDescription", svcErr.ErrorDescription))
		return nil, errors.New("failed to retrieve OAuth client configuration")
	}

	if len(oauthConfigs.Scopes) == 1 {
		logger.Debug("No additional scopes configured.")
	} else {
		// Get user info using the access token
		userInfo, err := g.GetUserInfo(ctx, execResp, accessToken)
		if err != nil {
			return nil, fmt.Errorf("failed to get user info: %w", err)
		}
		if execResp.Status == flowcm.ExecFailure {
			return nil, nil
		}

		for key, value := range userInfo {
			if key != "username" && key != "sub" && key != "id" {
				userClaims[key] = value
			}
		}
	}

	authenticatedUser := authncm.AuthenticatedUser{}
	if ctx.FlowType == flowcm.FlowTypeRegistration {
		authenticatedUser.IsAuthenticated = false
	} else {
		authenticatedUser.IsAuthenticated = true
		authenticatedUser.UserID = user.ID
		authenticatedUser.OrganizationUnitID = user.OrganizationUnit
		authenticatedUser.UserType = user.Type
	}

	// TODO: Need to convert attributes as per the IDP to local attribute mapping
	//  when the support is implemented.
	authenticatedUser.Attributes = userClaims

	return &authenticatedUser, nil
}

// resolveUser resolves the internal user based on the sub claim.
func (g *GoogleOIDCAuthExecutor) resolveUser(sub string, ctx *flowmodel.NodeContext,
	execResp *flowmodel.ExecutorResponse) (*user.User, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName),
		log.String(log.LoggerKeyExecutorName, g.GetName()),
		log.String(log.LoggerKeyFlowID, ctx.FlowID))

	user, svcErr := g.googleAuthService.GetInternalUser(sub)
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

				return nil, nil
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

	return user, nil
}
