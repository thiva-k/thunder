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
	"encoding/json"
	"errors"
	"fmt"
	"slices"

	authncm "github.com/asgardeo/thunder/internal/authn/common"
	authnoauth "github.com/asgardeo/thunder/internal/authn/oauth"
	authnoidc "github.com/asgardeo/thunder/internal/authn/oidc"
	flowcm "github.com/asgardeo/thunder/internal/flow/common"
	flowcore "github.com/asgardeo/thunder/internal/flow/core"
	"github.com/asgardeo/thunder/internal/idp"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/user"
	"github.com/asgardeo/thunder/internal/userschema"
)

const (
	oidcAuthLoggerComponentName = "OIDCAuthExecutor"
)

// idTokenNonUserAttributes contains the list of non-user attributes that are expected in the ID token.
var idTokenNonUserAttributes = []string{"aud", "exp", "iat", "iss", "at_hash", "azp", "nonce", "sub"}

// oidcAuthExecutorInterface defines the interface for OIDC authentication executors.
type oidcAuthExecutorInterface interface {
	oAuthExecutorInterface
	GetIDTokenClaims(execResp *flowcm.ExecutorResponse, idToken string) (map[string]interface{}, error)
}

// oidcAuthExecutor implements the OIDCAuthExecutorInterface for handling generic OIDC authentication flows.
type oidcAuthExecutor struct {
	oAuthExecutorInterface
	authService       authnoidc.OIDCAuthnCoreServiceInterface
	userService       user.UserServiceInterface
	userSchemaService userschema.UserSchemaServiceInterface
	logger            *log.Logger
}

var _ flowcore.ExecutorInterface = (*oidcAuthExecutor)(nil)

// newOIDCAuthExecutor creates a new instance of OIDCAuthExecutor.
func newOIDCAuthExecutor(
	name string,
	defaultInputs, prerequisites []flowcm.InputData,
	flowFactory flowcore.FlowFactoryInterface,
	idpService idp.IDPServiceInterface,
	authService authnoidc.OIDCAuthnCoreServiceInterface,
	userService user.UserServiceInterface,
	userSchemaService userschema.UserSchemaServiceInterface,
) oidcAuthExecutorInterface {
	if name == "" {
		name = ExecutorNameOIDCAuth
	}
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, oidcAuthLoggerComponentName),
		log.String(log.LoggerKeyExecutorName, name))

	oauthSvcCast, ok := authService.(authnoauth.OAuthAuthnCoreServiceInterface)
	if !ok {
		panic("failed to cast OIDCAuthnService to OAuthAuthnCoreServiceInterface")
	}

	base := newOAuthExecutor(name, defaultInputs, prerequisites,
		flowFactory, idpService, oauthSvcCast, userService, userSchemaService)

	return &oidcAuthExecutor{
		oAuthExecutorInterface: base,
		authService:            authService,
		userService:            userService,
		userSchemaService:      userSchemaService,
		logger:                 logger,
	}
}

// Execute executes the OIDC authentication logic.
func (o *oidcAuthExecutor) Execute(ctx *flowcore.NodeContext) (*flowcm.ExecutorResponse, error) {
	logger := o.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Executing OIDC authentication executor")

	execResp := &flowcm.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	// Check if the required input data is provided
	if o.CheckInputData(ctx, execResp) {
		// If required input data is not provided, return incomplete status with redirection to OIDC provider.
		logger.Debug("Required input data for OIDC authentication executor is not provided")
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

	logger.Debug("OIDC authentication executor execution completed",
		log.String("status", string(execResp.Status)),
		log.Bool("isAuthenticated", execResp.AuthenticatedUser.IsAuthenticated))

	return execResp, nil
}

// ProcessAuthFlowResponse processes the response from the OIDC authentication flow and authenticates the user.
func (o *oidcAuthExecutor) ProcessAuthFlowResponse(ctx *flowcore.NodeContext,
	execResp *flowcm.ExecutorResponse) error {
	logger := o.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Processing OIDC authentication response")

	code, ok := ctx.UserInputData["code"]
	if ok && code != "" {
		tokenResp, err := o.ExchangeCodeForToken(ctx, execResp, code)
		if err != nil {
			logger.Error("Failed to exchange code for a token", log.Error(err))
			return fmt.Errorf("failed to exchange code for token: %w", err)
		}
		if execResp.Status == flowcm.ExecFailure {
			return nil
		}

		idTokenClaims, err := o.GetIDTokenClaims(execResp, tokenResp.IDToken)
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

		user, err := o.resolveUserForOIDC(ctx, execResp, parsedSub, idTokenClaims)
		if err != nil {
			return err
		}
		if execResp.Status == flowcm.ExecFailure {
			return nil
		}

		authenticatedUser, err := getAuthenticatedUserForOIDC(o, o.authService, logger,
			ctx, execResp, tokenResp.AccessToken, idTokenClaims, user)
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
		execResp.FailureReason = failureReasonInvalidAuthorizationCode
		return nil
	}

	return nil
}

// GetIDTokenClaims extracts the ID token claims from the provided ID token.
func (o *oidcAuthExecutor) GetIDTokenClaims(execResp *flowcm.ExecutorResponse,
	idToken string) (map[string]interface{}, error) {
	logger := o.logger
	logger.Debug("Extracting claims from the ID token")

	claims, svcErr := o.authService.GetIDTokenClaims(idToken)
	if svcErr != nil {
		if svcErr.Type == serviceerror.ClientErrorType {
			execResp.Status = flowcm.ExecFailure
			execResp.FailureReason = svcErr.ErrorDescription
			return nil, nil
		}

		logger.Error("Failed to extract claims from the ID token", log.String("errorCode", svcErr.Code),
			log.String("errorDescription", svcErr.ErrorDescription))
		return nil, errors.New("failed to extract claims from the ID token")
	}

	return claims, nil
}

// resolveUserForOIDC resolves the internal user based on the sub claim.
func (o *oidcAuthExecutor) resolveUserForOIDC(ctx *flowcore.NodeContext, execResp *flowcm.ExecutorResponse,
	sub string, idTokenClaims map[string]interface{}) (*user.User, error) {
	logger := o.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))

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

				return nil, nil
			} else {
				// Provision the user automatically if allowed user types are configured
				provisionedUser, provisionErr := o.provisionUserOIDC(ctx, execResp, sub, idTokenClaims)
				if provisionErr != nil {
					logger.Error("Automatic user provisioning failed", log.Error(provisionErr), log.String("sub", sub))
					execResp.Status = flowcm.ExecFailure
					execResp.FailureReason = "User not found and automatic provisioning is not available"
					return nil, nil
				}
				if provisionedUser != nil {
					return provisionedUser, nil
				}
				// Provisioning not possible, return failure
				execResp.Status = flowcm.ExecFailure
				execResp.FailureReason = failureReasonUserNotFound
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

// getAuthenticatedUserForOIDC constructs the authenticated user object with attributes from the
// ID token and user info.
func getAuthenticatedUserForOIDC(o oidcAuthExecutorInterface, authService authnoidc.OIDCAuthnCoreServiceInterface,
	logger *log.Logger, ctx *flowcore.NodeContext, execResp *flowcm.ExecutorResponse, accessToken string,
	idTokenClaims map[string]interface{}, user *user.User) (*authncm.AuthenticatedUser, error) {
	userClaims := make(map[string]interface{})
	if len(idTokenClaims) != 0 {
		// Filter non-user claims from the ID token claims.
		for attr, val := range idTokenClaims {
			if !slices.Contains(idTokenNonUserAttributes, attr) {
				userClaims[attr] = val
			}
		}
		logger.Debug("Extracted ID token claims", log.Int("noOfClaims", len(idTokenClaims)))
	}

	idpID, err := o.GetIdpID(ctx)
	if err != nil {
		return nil, err
	}

	oauthConfigs, svcErr := authService.GetOAuthClientConfig(idpID)
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
		userInfo, err := o.GetUserInfo(ctx, execResp, accessToken)
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

	// Append email to runtime data if available.
	if email, ok := userClaims["email"]; ok {
		if emailStr, ok := email.(string); ok && emailStr != "" {
			if execResp.RuntimeData == nil {
				execResp.RuntimeData = make(map[string]string)
			}
			execResp.RuntimeData["email"] = emailStr
		}
	}

	return &authenticatedUser, nil
}

// provisionUserOIDC attempts to automatically provision a user if allowed user types are configured.
func (o *oidcAuthExecutor) provisionUserOIDC(ctx *flowcore.NodeContext, execResp *flowcm.ExecutorResponse,
	sub string, idTokenClaims map[string]interface{}) (*user.User, error) {
	logger := o.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))

	allowedUserTypes := ctx.Application.AllowedUserTypes
	if len(allowedUserTypes) == 0 {
		logger.Debug("No allowed user types configured, cannot provision user automatically",
			log.String("sub", sub))
		return nil, nil
	}

	if o.userService == nil || o.userSchemaService == nil {
		return nil, fmt.Errorf("required services are not available for provisioning")
	}

	// Filter allowed user types to only those with self-registration enabled
	selfRegEnabledUserTypes := make([]string, 0)
	var selectedUserSchema *userschema.UserSchema
	for _, userType := range allowedUserTypes {
		userSchema, svcErr := o.userSchemaService.GetUserSchemaByName(userType)
		if svcErr != nil {
			logger.Debug("Failed to get user schema for user type, skipping",
				log.String("userType", userType), log.String("errorCode", svcErr.Code))
			continue
		}
		if userSchema.AllowSelfRegistration {
			selfRegEnabledUserTypes = append(selfRegEnabledUserTypes, userType)
			if selectedUserSchema == nil {
				selectedUserSchema = userSchema
			}
		}
	}

	// Fail if no user types have self-registration enabled
	if len(selfRegEnabledUserTypes) == 0 {
		logger.Debug("No user types with self-registration enabled for automatic provisioning",
			log.String("sub", sub))
		return nil, fmt.Errorf("no user types with self-registration enabled")
	}

	// Fail if multiple user types have self-registration enabled (ambiguous)
	if len(selfRegEnabledUserTypes) > 1 {
		logger.Debug("Multiple user types with self-registration enabled, cannot automatically provision",
			log.String("sub", sub), log.Any("userTypes", selfRegEnabledUserTypes))
		return nil, fmt.Errorf("multiple user types with self-registration enabled: %v", selfRegEnabledUserTypes)
	}

	// Exactly one user type with self-registration enabled
	userType := selfRegEnabledUserTypes[0]
	userSchema := selectedUserSchema

	if userSchema.OrganizationUnitID == "" {
		logger.Error("No organization unit found for user type", log.String("userType", userType))
		return nil, fmt.Errorf("no organization unit found for user type: %s", userType)
	}

	// Extract user attributes from ID token claims
	// We need to include 'sub' in attributes for user identification
	userAttributes := make(map[string]interface{})
	if len(idTokenClaims) != 0 {
		for attr, val := range idTokenClaims {
			if attr == "sub" || !slices.Contains(idTokenNonUserAttributes, attr) {
				userAttributes[attr] = val
			}
		}
	}

	attributesJSON, err := json.Marshal(userAttributes)
	if err != nil {
		logger.Error("Failed to marshal user attributes", log.Error(err))
		return nil, fmt.Errorf("failed to marshal user attributes: %w", err)
	}

	// Create the user
	newUser := &user.User{
		OrganizationUnit: userSchema.OrganizationUnitID,
		Type:             userType,
		Attributes:       attributesJSON,
	}

	createdUser, svcErr := o.userService.CreateUser(newUser)
	if svcErr != nil {
		if svcErr.Type == serviceerror.ClientErrorType {
			return nil, fmt.Errorf("failed to create user: %s", svcErr.ErrorDescription)
		}
		logger.Error("Failed to create user during provisioning",
			log.String("errorCode", svcErr.Code), log.String("errorDescription", svcErr.ErrorDescription),
			log.String("userType", userType), log.String("sub", sub))
		return nil, fmt.Errorf("failed to create user: %s", svcErr.ErrorDescription)
	}

	// Set runtime data for user provisioning
	if execResp.RuntimeData == nil {
		execResp.RuntimeData = make(map[string]string)
	}
	execResp.RuntimeData[userAutoProvisionedKey] = "true"
	execResp.RuntimeData[userTypeKey] = userType
	execResp.RuntimeData[defaultOUIDKey] = userSchema.OrganizationUnitID

	return createdUser, nil
}
