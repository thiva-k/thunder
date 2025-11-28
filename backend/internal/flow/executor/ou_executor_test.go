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

	flowcm "github.com/asgardeo/thunder/internal/flow/common"
	flowcore "github.com/asgardeo/thunder/internal/flow/core"
	"github.com/asgardeo/thunder/internal/ou"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/tests/mocks/flow/coremock"
	"github.com/asgardeo/thunder/tests/mocks/oumock"
)

const testOUID = "ou-123"

type OUExecutorTestSuite struct {
	suite.Suite
	mockOUService   *oumock.OrganizationUnitServiceInterfaceMock
	mockFlowFactory *coremock.FlowFactoryInterfaceMock
	executor        *ouExecutor
}

func TestOUExecutorSuite(t *testing.T) {
	suite.Run(t, new(OUExecutorTestSuite))
}

func (suite *OUExecutorTestSuite) SetupTest() {
	suite.mockOUService = oumock.NewOrganizationUnitServiceInterfaceMock(suite.T())
	suite.mockFlowFactory = coremock.NewFlowFactoryInterfaceMock(suite.T())

	defaultInputs := []flowcm.InputData{
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
	}

	// Mock the CreateExecutor method to return a base executor
	suite.mockFlowFactory.On("CreateExecutor", ExecutorNameOUCreation, flowcm.ExecutorTypeRegistration,
		defaultInputs, []flowcm.InputData{}).
		Return(newMockExecutor("TestOUExecutor", flowcm.ExecutorTypeUtility, defaultInputs, []flowcm.InputData{}))

	suite.executor = newOUExecutor(suite.mockFlowFactory, suite.mockOUService)
}

// newMockExecutor creates a mock executor for testing purposes
func newMockExecutor(name string, executorType flowcm.ExecutorType, defaultInputs []flowcm.InputData,
	prerequisites []flowcm.InputData) flowcore.ExecutorInterface {
	mockExec := coremock.NewExecutorInterfaceMock(&testing.T{})
	mockExec.On("GetName").Return(name)
	mockExec.On("GetType").Return(executorType)
	mockExec.On("GetDefaultExecutorInputs").Return(defaultInputs)
	mockExec.On("GetPrerequisites").Return(prerequisites)
	mockExec.On("GetRequiredData", mock.Anything).Return(defaultInputs)
	mockExec.On("CheckInputData", mock.Anything, mock.Anything).Return(
		func(ctx *flowcore.NodeContext, execResp *flowcm.ExecutorResponse) bool {
			requiredData := defaultInputs
			if execResp.RequiredData == nil {
				execResp.RequiredData = make([]flowcm.InputData, 0)
			}
			if len(ctx.UserInputData) == 0 && len(ctx.RuntimeData) == 0 {
				execResp.RequiredData = append(execResp.RequiredData, requiredData...)
				return true
			}
			requireData := false
			for _, inputData := range requiredData {
				if _, ok := ctx.UserInputData[inputData.Name]; !ok {
					if _, ok := ctx.RuntimeData[inputData.Name]; ok {
						continue
					}
					requireData = true
					execResp.RequiredData = append(execResp.RequiredData, inputData)
				}
			}
			return requireData
		})
	mockExec.On("ValidatePrerequisites", mock.Anything, mock.Anything).Return(true)
	mockExec.On("GetUserIDFromContext", mock.Anything).Return("")
	return mockExec
}

func (suite *OUExecutorTestSuite) TestNewOUExecutor() {
	mockFlowFactory := coremock.NewFlowFactoryInterfaceMock(suite.T())
	mockOUService := oumock.NewOrganizationUnitServiceInterfaceMock(suite.T())

	defaultInputs := []flowcm.InputData{
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
	}

	// Mock the CreateExecutor method
	mockFlowFactory.On("CreateExecutor", ExecutorNameOUCreation, flowcm.ExecutorTypeRegistration,
		defaultInputs, []flowcm.InputData{}).
		Return(newMockExecutor("OUExecutor", flowcm.ExecutorTypeRegistration, defaultInputs, []flowcm.InputData{}))

	executor := newOUExecutor(mockFlowFactory, mockOUService)

	assert.NotNil(suite.T(), executor)
	assert.Equal(suite.T(), "OUExecutor", executor.GetName())

	defaultInputsResult := executor.GetDefaultExecutorInputs()
	assert.Len(suite.T(), defaultInputsResult, 2)
	assert.Equal(suite.T(), userInputOuName, defaultInputsResult[0].Name)
	assert.True(suite.T(), defaultInputsResult[0].Required)
	assert.Equal(suite.T(), userInputOuHandle, defaultInputsResult[1].Name)
	assert.True(suite.T(), defaultInputsResult[1].Required)
}

func (suite *OUExecutorTestSuite) TestExecutorMetadata() {
	testCases := []struct {
		name     string
		testFunc func()
	}{
		{
			name: "GetName returns correct executor name",
			testFunc: func() {
				assert.Equal(suite.T(), "TestOUExecutor", suite.executor.GetName())
			},
		},
		{
			name: "GetDefaultExecutorInputs returns two inputs",
			testFunc: func() {
				inputs := suite.executor.GetDefaultExecutorInputs()
				assert.Len(suite.T(), inputs, 2)
			},
		},
		{
			name: "GetPrerequisites returns empty list",
			testFunc: func() {
				prerequisites := suite.executor.GetPrerequisites()
				assert.Empty(suite.T(), prerequisites)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, tc.testFunc)
	}
}

type ExecuteSuccessTestCase struct {
	name             string
	userInputData    map[string]string
	expectedOUID     string
	expectedRequest  ou.OrganizationUnitRequest
	expectedResponse ou.OrganizationUnit
}

func (suite *OUExecutorTestSuite) TestExecute_Success() {
	testCases := []ExecuteSuccessTestCase{
		{
			name: "Create OU with all fields",
			userInputData: map[string]string{
				userInputOuName:   "Engineering",
				userInputOuHandle: "engineering",
			},
			expectedOUID: testOUID,
			expectedRequest: ou.OrganizationUnitRequest{
				Name:   "Engineering",
				Handle: "engineering",
			},
			expectedResponse: ou.OrganizationUnit{
				ID:     testOUID,
				Name:   "Engineering",
				Handle: "engineering",
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			ctx := &flowcore.NodeContext{
				FlowID:        "flow-123",
				FlowType:      flowcm.FlowTypeRegistration,
				UserInputData: tc.userInputData,
				RuntimeData:   map[string]string{},
			}

			suite.mockOUService.On("CreateOrganizationUnit", tc.expectedRequest).
				Return(tc.expectedResponse, nil)

			result, err := suite.executor.Execute(ctx)

			assert.NoError(suite.T(), err)
			assert.NotNil(suite.T(), result)
			assert.Equal(suite.T(), flowcm.ExecComplete, result.Status)
			assert.Equal(suite.T(), tc.expectedOUID, result.RuntimeData[ouIDKey])
			suite.mockOUService.AssertExpectations(suite.T())
		})
	}
}

type ExecuteNonRegistrationFlowTestCase struct {
	name     string
	flowType flowcm.FlowType
}

func (suite *OUExecutorTestSuite) TestExecute_NonRegistrationFlow() {
	testCases := []ExecuteNonRegistrationFlowTestCase{
		{
			name:     "Authentication flow",
			flowType: flowcm.FlowTypeAuthentication,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			ctx := &flowcore.NodeContext{
				FlowID:   "flow-123",
				FlowType: tc.flowType,
			}

			result, err := suite.executor.Execute(ctx)

			assert.NoError(suite.T(), err)
			assert.NotNil(suite.T(), result)
			assert.Equal(suite.T(), flowcm.ExecUserInputRequired, result.Status)
			assert.Empty(suite.T(), result.RuntimeData[ouIDKey])
		})
	}
}

type ExecutePrerequisitesFailureTestCase struct {
	name        string
	ctx         *flowcore.NodeContext
	expectedMsg string
}

func (suite *OUExecutorTestSuite) TestExecute_PrerequisitesFailure() {
	mockOUService := oumock.NewOrganizationUnitServiceInterfaceMock(suite.T())

	defaultInputs := []flowcm.InputData{
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
	}

	prerequisites := []flowcm.InputData{{Name: "requiredField", Required: true, Type: "string"}}

	// Create a mock executor with prerequisites
	mockExec := coremock.NewExecutorInterfaceMock(suite.T())
	mockExec.On("GetName").Return("Test").Maybe()
	mockExec.On("GetType").Return(flowcm.ExecutorTypeUtility).Maybe()
	mockExec.On("GetDefaultExecutorInputs").Return(defaultInputs).Maybe()
	mockExec.On("GetPrerequisites").Return(prerequisites).Maybe()
	mockExec.On("ValidatePrerequisites", mock.Anything, mock.Anything).Return(
		func(ctx *flowcore.NodeContext, execResp *flowcm.ExecutorResponse) bool {
			for _, prerequisite := range prerequisites {
				if _, ok := ctx.UserInputData[prerequisite.Name]; !ok {
					if _, ok := ctx.RuntimeData[prerequisite.Name]; !ok {
						execResp.Status = flowcm.ExecFailure
						execResp.FailureReason = "Prerequisite not met: " + prerequisite.Name
						return false
					}
				}
			}
			return true
		}).Maybe()

	// Create a prerequisitesExecutor with the mock interface directly
	prerequisitesExecutor := &ouExecutor{
		ExecutorInterface: mockExec,
		ouService:         mockOUService,
		logger: log.GetLogger().With(
			log.String(log.LoggerKeyComponentName, ouExecLoggerComponentName)),
	}

	testCases := []ExecutePrerequisitesFailureTestCase{
		{
			name: "Missing prerequisite field",
			ctx: &flowcore.NodeContext{
				FlowID:        "flow-123",
				FlowType:      flowcm.FlowTypeRegistration,
				UserInputData: map[string]string{},
				RuntimeData:   map[string]string{},
			},
			expectedMsg: "Prerequisites validation failed for OU creation",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			result, err := prerequisitesExecutor.Execute(tc.ctx)

			assert.NoError(suite.T(), err)
			assert.NotNil(suite.T(), result)
			assert.Equal(suite.T(), flowcm.ExecFailure, result.Status)
			assert.Equal(suite.T(), tc.expectedMsg, result.FailureReason)
			mockOUService.AssertNotCalled(suite.T(), "CreateOrganizationUnit", mock.Anything)
		})
	}
}

type ExecuteUserInputRequiredTestCase struct {
	name          string
	userInputData map[string]string
}

func (suite *OUExecutorTestSuite) TestExecute_UserInputRequired() {
	testCases := []ExecuteUserInputRequiredTestCase{
		{
			name:          "No input data provided",
			userInputData: map[string]string{},
		},
		{
			name: "Missing OU name",
			userInputData: map[string]string{
				userInputOuHandle: "engineering",
			},
		},
		{
			name: "Missing OU handle",
			userInputData: map[string]string{
				userInputOuName: "Engineering",
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			ctx := &flowcore.NodeContext{
				FlowID:        "flow-123",
				FlowType:      flowcm.FlowTypeRegistration,
				UserInputData: tc.userInputData,
			}

			result, err := suite.executor.Execute(ctx)

			assert.NoError(suite.T(), err)
			assert.NotNil(suite.T(), result)
			assert.Equal(suite.T(), flowcm.ExecUserInputRequired, result.Status)
			assert.NotEmpty(suite.T(), result.RequiredData)
			suite.mockOUService.AssertNotCalled(suite.T(), "CreateOrganizationUnit", mock.Anything)
		})
	}
}

func (suite *OUExecutorTestSuite) TestExecute_ErrorScenarios() {
	testCases := []struct {
		name            string
		serviceError    serviceerror.ServiceError
		expectedFailure string
		expectError     bool
		expectNilResult bool
		userInputData   map[string]string
		expectedRequest ou.OrganizationUnitRequest
	}{
		{
			name:            "OU name conflict",
			serviceError:    ou.ErrorOrganizationUnitNameConflict,
			expectedFailure: "An organization unit with the same name already exists.",
			expectError:     false,
			expectNilResult: false,
			userInputData: map[string]string{
				userInputOuName:   "Engineering",
				userInputOuHandle: "engineering",
			},
			expectedRequest: ou.OrganizationUnitRequest{
				Name:   "Engineering",
				Handle: "engineering",
			},
		},
		{
			name:            "OU handle conflict",
			serviceError:    ou.ErrorOrganizationUnitHandleConflict,
			expectedFailure: "An organization unit with the same handle already exists.",
			expectError:     false,
			expectNilResult: false,
			userInputData: map[string]string{
				userInputOuName:   "Engineering",
				userInputOuHandle: "engineering",
			},
			expectedRequest: ou.OrganizationUnitRequest{
				Name:   "Engineering",
				Handle: "engineering",
			},
		},
		{
			name: "Other client error",
			serviceError: serviceerror.ServiceError{
				Type:             serviceerror.ClientErrorType,
				Code:             "OU-9999",
				Error:            "Test Error",
				ErrorDescription: "Test error description",
			},
			expectedFailure: "Failed to create organization unit: Test error description",
			expectError:     false,
			expectNilResult: false,
			userInputData: map[string]string{
				userInputOuName:   "Engineering",
				userInputOuHandle: "engineering",
			},
			expectedRequest: ou.OrganizationUnitRequest{
				Name:   "Engineering",
				Handle: "engineering",
			},
		},
		{
			name:            "Internal server error",
			serviceError:    ou.ErrorInternalServerError,
			expectedFailure: "failed to create organization unit",
			expectError:     true,
			expectNilResult: true,
			userInputData: map[string]string{
				userInputOuName:   "Engineering",
				userInputOuHandle: "engineering",
			},
			expectedRequest: ou.OrganizationUnitRequest{
				Name:   "Engineering",
				Handle: "engineering",
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			ctx := &flowcore.NodeContext{
				FlowID:        "flow-123",
				FlowType:      flowcm.FlowTypeRegistration,
				UserInputData: tc.userInputData,
				RuntimeData:   map[string]string{},
			}

			suite.mockOUService.On("CreateOrganizationUnit", tc.expectedRequest).
				Return(ou.OrganizationUnit{}, &tc.serviceError)

			result, err := suite.executor.Execute(ctx)

			if tc.expectError {
				assert.Error(suite.T(), err)
				assert.Equal(suite.T(), tc.expectedFailure, err.Error())
			} else {
				assert.NoError(suite.T(), err)
				assert.Equal(suite.T(), flowcm.ExecFailure, result.Status)
				assert.Equal(suite.T(), tc.expectedFailure, result.FailureReason)
			}

			if tc.expectNilResult {
				assert.Nil(suite.T(), result)
			} else {
				assert.NotNil(suite.T(), result)
			}

			suite.mockOUService.AssertExpectations(suite.T())
		})
	}
}

func (suite *OUExecutorTestSuite) TestExecute_EmptyOUID() {
	suite.SetupTest()

	ctx := &flowcore.NodeContext{
		FlowID:   "flow-123",
		FlowType: flowcm.FlowTypeRegistration,
		UserInputData: map[string]string{
			userInputOuName:   "Engineering",
			userInputOuHandle: "engineering",
		},
		RuntimeData: map[string]string{},
	}

	expectedRequest := ou.OrganizationUnitRequest{
		Name:   "Engineering",
		Handle: "engineering",
	}

	suite.mockOUService.On("CreateOrganizationUnit", expectedRequest).
		Return(ou.OrganizationUnit{ID: ""}, nil)

	result, err := suite.executor.Execute(ctx)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), "failed to create organization unit", err.Error())
	suite.mockOUService.AssertExpectations(suite.T())
}

func (suite *OUExecutorTestSuite) TestExecutorHelperMethods() {
	testCases := []struct {
		name     string
		testFunc func()
	}{
		{
			name: "CheckInputData with empty inputs returns true and sets required data",
			testFunc: func() {
				ctx := &flowcore.NodeContext{
					UserInputData: map[string]string{},
					RuntimeData:   map[string]string{},
				}
				execResp := &flowcm.ExecutorResponse{
					AdditionalData: make(map[string]string),
					RuntimeData:    make(map[string]string),
				}

				result := suite.executor.CheckInputData(ctx, execResp)

				assert.True(suite.T(), result)
				assert.NotEmpty(suite.T(), execResp.RequiredData)
			},
		},
		{
			name: "ValidatePrerequisites with no prerequisites returns true",
			testFunc: func() {
				ctx := &flowcore.NodeContext{
					UserInputData: map[string]string{},
					RuntimeData:   map[string]string{},
				}
				execResp := &flowcm.ExecutorResponse{
					AdditionalData: make(map[string]string),
					RuntimeData:    make(map[string]string),
				}

				result := suite.executor.ValidatePrerequisites(ctx, execResp)

				assert.True(suite.T(), result)
			},
		},
		{
			name: "GetUserIDFromContext with empty context returns empty string",
			testFunc: func() {
				ctx := &flowcore.NodeContext{
					UserInputData: map[string]string{},
					RuntimeData:   map[string]string{},
				}

				userID := suite.executor.GetUserIDFromContext(ctx)
				assert.Empty(suite.T(), userID)
			},
		},
		{
			name: "GetRequiredData returns three required fields",
			testFunc: func() {
				ctx := &flowcore.NodeContext{
					UserInputData: map[string]string{},
					RuntimeData:   map[string]string{},
				}

				requiredData := suite.executor.GetRequiredData(ctx)

				assert.NotEmpty(suite.T(), requiredData)
				assert.Len(suite.T(), requiredData, 2)
			},
		},
		{
			name: "getOrganizationUnitRequest constructs request correctly",
			testFunc: func() {
				ctx := &flowcore.NodeContext{
					UserInputData: map[string]string{
						userInputOuName:   "Engineering",
						userInputOuHandle: "engineering",
					},
				}

				request := suite.executor.getOrganizationUnitRequest(ctx)

				assert.Equal(suite.T(), "Engineering", request.Name)
				assert.Equal(suite.T(), "engineering", request.Handle)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, tc.testFunc)
	}
}

func (suite *OUExecutorTestSuite) TestOUExecutorInterface() {
	var _ flowcore.ExecutorInterface = (*ouExecutor)(nil)
}
