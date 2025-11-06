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

// Package basicauth provides the basic authentication executor for handling username and password authentication.
package basicauth

import (
	"encoding/json"
	"errors"

	authncm "github.com/asgardeo/thunder/internal/authn/common"
	authncreds "github.com/asgardeo/thunder/internal/authn/credentials"
	"github.com/asgardeo/thunder/internal/executor/identify"
	flowcm "github.com/asgardeo/thunder/internal/flow/common"
	flowmodel "github.com/asgardeo/thunder/internal/flow/common/model"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
)

const (
	executorName          = authncm.AuthenticatorCredentials
	loggerComponentName   = "BasicAuthExecutor"
	userAttributeUsername = "username"
	userAttributePassword = "password"
)

// BasicAuthExecutor implements the ExecutorInterface for basic authentication.
type BasicAuthExecutor struct {
	flowmodel.ExecutorInterface
	identify.IdentifyingExecutorInterface
	credsAuthSvc authncreds.CredentialsAuthnServiceInterface
}

var _ flowmodel.ExecutorInterface = (*BasicAuthExecutor)(nil)
var _ identify.IdentifyingExecutorInterface = (*BasicAuthExecutor)(nil)

// NewBasicAuthExecutor creates a new instance of BasicAuthExecutor.
func NewBasicAuthExecutor() *BasicAuthExecutor {
	defaultInputs := []flowmodel.InputData{
		{
			Name:     userAttributeUsername,
			Type:     "string",
			Required: true,
		},
		{
			Name:     userAttributePassword,
			Type:     "string",
			Required: true,
		},
	}

	identifyExec := identify.NewIdentifyingExecutor()
	base := flowmodel.NewExecutor(executorName, flowcm.ExecutorTypeAuthentication,
		defaultInputs, []flowmodel.InputData{})

	return &BasicAuthExecutor{
		ExecutorInterface:            base,
		IdentifyingExecutorInterface: identifyExec,
		credsAuthSvc:                 authncreds.NewCredentialsAuthnService(nil),
	}
}

// Execute executes the basic authentication logic.
func (b *BasicAuthExecutor) Execute(ctx *flowmodel.NodeContext) (*flowmodel.ExecutorResponse, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName),
		log.String(log.LoggerKeyExecutorName, b.GetName()),
		log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Executing basic authentication executor")

	execResp := &flowmodel.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	// Validate for the required input data.
	if b.CheckInputData(ctx, execResp) {
		// If required input data is not provided, return incomplete status.
		logger.Debug("Required input data for basic authentication executor is not provided")
		execResp.Status = flowcm.ExecUserInputRequired
		return execResp, nil
	}

	// TODO: Should handle client errors here. Service should return a ServiceError and
	//  client errors should be appended as a failure.
	//  For the moment handling returned error as a authentication failure.
	authenticatedUser, err := b.getAuthenticatedUser(ctx, execResp)
	if err != nil {
		execResp.Status = flowcm.ExecFailure
		execResp.FailureReason = "Failed to authenticate user: " + err.Error()
		return execResp, nil
	}
	if execResp.Status == flowcm.ExecFailure {
		return execResp, nil
	}
	if authenticatedUser == nil {
		execResp.Status = flowcm.ExecFailure
		execResp.FailureReason = "Authenticated user not found."
		return execResp, nil
	}
	if !authenticatedUser.IsAuthenticated && ctx.FlowType != flowcm.FlowTypeRegistration {
		execResp.Status = flowcm.ExecFailure
		execResp.FailureReason = "User authentication failed."
		return execResp, nil
	}

	execResp.AuthenticatedUser = *authenticatedUser
	execResp.Status = flowcm.ExecComplete

	logger.Debug("Basic authentication executor execution completed",
		log.String("status", string(execResp.Status)),
		log.Bool("isAuthenticated", execResp.AuthenticatedUser.IsAuthenticated))

	return execResp, nil
}

// getAuthenticatedUser perform authentication based on the provided username and password and return
// authenticated user details.
func (b *BasicAuthExecutor) getAuthenticatedUser(ctx *flowmodel.NodeContext,
	execResp *flowmodel.ExecutorResponse) (*authncm.AuthenticatedUser, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName),
		log.String(log.LoggerKeyExecutorName, b.GetName()))

	username := ctx.UserInputData[userAttributeUsername]
	filters := map[string]interface{}{userAttributeUsername: username}
	userID, err := b.IdentifyUser(filters, execResp)
	if err != nil {
		return nil, err
	}

	// Handle registration flows.
	if ctx.FlowType == flowcm.FlowTypeRegistration {
		if execResp.Status == flowcm.ExecFailure {
			if execResp.FailureReason == "User not found" {
				logger.Debug("User not found for the provided username. Proceeding with registration flow.")
				execResp.Status = flowcm.ExecComplete

				return &authncm.AuthenticatedUser{
					IsAuthenticated: false,
					Attributes: map[string]interface{}{
						userAttributeUsername: username,
					},
				}, nil
			}
			return nil, err
		}

		// At this point, a unique user is found in the system. Hence fail the execution.
		execResp.Status = flowcm.ExecFailure
		execResp.FailureReason = "User already exists with the provided username."
		return nil, nil
	}

	if execResp.Status == flowcm.ExecFailure {
		return nil, nil
	}

	// Prepare authentication attributes with user identifier and credentials.
	authAttributes := map[string]interface{}{
		userAttributeUsername: username,
		userAttributePassword: ctx.UserInputData[userAttributePassword],
	}

	user, svcErr := b.credsAuthSvc.Authenticate(authAttributes)
	if svcErr != nil {
		if svcErr.Type == serviceerror.ClientErrorType {
			execResp.Status = flowcm.ExecFailure
			execResp.FailureReason = "Failed to authenticate user: " + svcErr.ErrorDescription
			return nil, nil
		}
		logger.Error("Failed to authenticate user", log.String("userID", *userID),
			log.String("errorCode", svcErr.Code), log.String("errorDescription", svcErr.ErrorDescription))
		return nil, errors.New("failed to authenticate user")
	}

	var attrs map[string]interface{}
	if err := json.Unmarshal(user.Attributes, &attrs); err != nil {
		logger.Error("Failed to unmarshal user attributes", log.Error(err))
		return nil, err
	}

	authenticatedUser := authncm.AuthenticatedUser{
		IsAuthenticated:    true,
		UserID:             user.ID,
		OrganizationUnitID: user.OrganizationUnit,
		UserType:           user.Type,
		Attributes:         attrs,
	}
	return &authenticatedUser, nil
}
