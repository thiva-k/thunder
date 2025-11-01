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

// Package ouexec implements the organizational unit (OU) executor for creating OUs.
package ouexec

import (
	"errors"

	flowconst "github.com/asgardeo/thunder/internal/flow/common/constants"
	flowmodel "github.com/asgardeo/thunder/internal/flow/common/model"
	"github.com/asgardeo/thunder/internal/ou"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
)

const (
	loggerComponentName = "OUExecutor"
	userInputOuName     = "ouName"
	userInputOuHandle   = "ouHandle"
	userInputOuDesc     = "ouDescription"
	ouIDKey             = "ouId"
)

// OUExecutor is responsible for creating organizational units (OUs) within the system.
type OUExecutor struct {
	internal  flowmodel.Executor
	ouService ou.OrganizationUnitServiceInterface
}

var _ flowmodel.ExecutorInterface = (*OUExecutor)(nil)

// NewOUExecutor creates a new instance of OUExecutor with the given parameters.
func NewOUExecutor(id, name string, properties map[string]string) *OUExecutor {
	defaultInputs := []flowmodel.InputData{
		{
			Name:     userInputOuName,
			Required: true,
			Type:     "string",
		},
		{
			Name:     userInputOuHandle,
			Required: true,
			Type:     "string",
		},
		{
			Name:     userInputOuDesc,
			Required: false,
			Type:     "string",
		},
	}

	return &OUExecutor{
		internal: *flowmodel.NewExecutor(id, name, flowconst.ExecutorTypeUtility,
			defaultInputs, []flowmodel.InputData{}, properties),
		ouService: ou.NewOrganizationUnitService(),
	}
}

// GetID returns the ID of the OUExecutor.
func (o *OUExecutor) GetID() string {
	return o.internal.GetID()
}

// GetName returns the name of the OUExecutor.
func (o *OUExecutor) GetName() string {
	return o.internal.GetName()
}

// GetProperties returns the properties of the OUExecutor.
func (o *OUExecutor) GetProperties() flowmodel.ExecutorProperties {
	return o.internal.GetProperties()
}

// Execute executes the ou creation logic.
func (o *OUExecutor) Execute(ctx *flowmodel.NodeContext) (*flowmodel.ExecutorResponse, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName),
		log.String(log.LoggerKeyExecutorID, o.GetID()),
		log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Executing OU creation executor")

	execResp := &flowmodel.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	if ctx.FlowType != flowconst.FlowTypeRegistration {
		logger.Debug("Flow type is not registration, skipping OU creation")
		execResp.Status = flowconst.ExecComplete
		return execResp, nil
	}

	if !o.ValidatePrerequisites(ctx, execResp) {
		logger.Debug("Prerequisites validation failed for OU creation")
		execResp.Status = flowconst.ExecFailure
		execResp.FailureReason = "Prerequisites validation failed for OU creation"
		return execResp, nil
	}

	if o.CheckInputData(ctx, execResp) {
		logger.Debug("Required input data for OU creation is not provided")
		execResp.Status = flowconst.ExecUserInputRequired
		return execResp, nil
	}

	// Create the OU using the OU service.
	ouRequest := o.getOrganizationUnitRequest(ctx)
	createdOU, svcErr := o.ouService.CreateOrganizationUnit(ouRequest)
	if svcErr != nil {
		if svcErr.Type == serviceerror.ClientErrorType {
			execResp.Status = flowconst.ExecFailure

			switch svcErr.Code {
			case ou.ErrorOrganizationUnitNameConflict.Code:
				execResp.FailureReason = "An organization unit with the same name already exists."
			case ou.ErrorOrganizationUnitHandleConflict.Code:
				execResp.FailureReason = "An organization unit with the same handle already exists."
			default:
				execResp.FailureReason = "Failed to create organization unit: " + svcErr.ErrorDescription
			}

			return execResp, nil
		}

		logger.Error("Error occurred while creating organization unit: ", log.String("errorCode", svcErr.Code),
			log.String("errorDescription", svcErr.ErrorDescription))
		return nil, errors.New("failed to create organization unit")
	}

	if createdOU.ID == "" {
		logger.Error("Organization unit creation failed: received empty OU ID")
		return nil, errors.New("failed to create organization unit")
	}

	// Set the created OU ID in the runtime data for further use in the flow.
	execResp.RuntimeData[ouIDKey] = createdOU.ID

	logger.Debug("Organization unit created successfully", log.String(ouIDKey, createdOU.ID))
	execResp.Status = flowconst.ExecComplete
	return execResp, nil
}

// GetDefaultExecutorInputs returns the default inputs required by the OUExecutor.
func (o *OUExecutor) GetDefaultExecutorInputs() []flowmodel.InputData {
	return o.internal.GetDefaultExecutorInputs()
}

// GetPrerequisites returns the prerequisites for the OUExecutor.
func (o *OUExecutor) GetPrerequisites() []flowmodel.InputData {
	return o.internal.GetPrerequisites()
}

// CheckInputData checks if the required input data is provided.
func (o *OUExecutor) CheckInputData(ctx *flowmodel.NodeContext, execResp *flowmodel.ExecutorResponse) bool {
	return o.internal.CheckInputData(ctx, execResp)
}

// ValidatePrerequisites validates prerequisites for the OUExecutor.
func (o *OUExecutor) ValidatePrerequisites(ctx *flowmodel.NodeContext,
	execResp *flowmodel.ExecutorResponse) bool {
	return o.internal.ValidatePrerequisites(ctx, execResp)
}

// GetUserIDFromContext retrieves user ID from the context.
func (o *OUExecutor) GetUserIDFromContext(ctx *flowmodel.NodeContext) (string, error) {
	return o.internal.GetUserIDFromContext(ctx)
}

// GetRequiredData returns the required input data for the OUExecutor.
func (o *OUExecutor) GetRequiredData(ctx *flowmodel.NodeContext) []flowmodel.InputData {
	return o.internal.GetRequiredData(ctx)
}

// getOrganizationUnitRequest constructs an OrganizationUnitRequest from the NodeContext.
func (o *OUExecutor) getOrganizationUnitRequest(ctx *flowmodel.NodeContext) ou.OrganizationUnitRequest {
	return ou.OrganizationUnitRequest{
		Name:        ctx.UserInputData[userInputOuName],
		Handle:      ctx.UserInputData[userInputOuHandle],
		Description: ctx.UserInputData[userInputOuDesc],
	}
}
