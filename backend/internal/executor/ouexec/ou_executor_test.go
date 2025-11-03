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

package ouexec

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	flowcm "github.com/asgardeo/thunder/internal/flow/common"
	flowmodel "github.com/asgardeo/thunder/internal/flow/common/model"
	"github.com/asgardeo/thunder/internal/ou"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/tests/mocks/oumock"
)

type OUExecutorTestSuite struct {
	suite.Suite
	mockOUService *oumock.OrganizationUnitServiceInterfaceMock
	executor      *OUExecutor
}

func TestOUExecutorSuite(t *testing.T) {
	suite.Run(t, new(OUExecutorTestSuite))
}

func (suite *OUExecutorTestSuite) SetupTest() {
	suite.mockOUService = oumock.NewOrganizationUnitServiceInterfaceMock(suite.T())

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

	suite.executor = &OUExecutor{
		ExecutorInterface: flowmodel.NewExecutor("TestOUExecutor", flowcm.ExecutorTypeUtility,
			defaultInputs, []flowmodel.InputData{}),
		ouService: suite.mockOUService,
	}
}

func (suite *OUExecutorTestSuite) TestNewOUExecutor() {
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

	executor := &OUExecutor{
		ExecutorInterface: flowmodel.NewExecutor("OUExecutor", flowcm.ExecutorTypeUtility,
			defaultInputs, []flowmodel.InputData{}),
		ouService: suite.mockOUService,
	}

	assert.NotNil(suite.T(), executor)
	assert.Equal(suite.T(), "OUExecutor", executor.GetName())

	defaultInputsResult := executor.GetDefaultExecutorInputs()
	assert.Len(suite.T(), defaultInputsResult, 3)
	assert.Equal(suite.T(), userInputOuName, defaultInputsResult[0].Name)
	assert.True(suite.T(), defaultInputsResult[0].Required)
	assert.Equal(suite.T(), userInputOuHandle, defaultInputsResult[1].Name)
	assert.True(suite.T(), defaultInputsResult[1].Required)
	assert.Equal(suite.T(), userInputOuDesc, defaultInputsResult[2].Name)
	assert.False(suite.T(), defaultInputsResult[2].Required)
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
			name: "GetDefaultExecutorInputs returns three inputs",
			testFunc: func() {
				inputs := suite.executor.GetDefaultExecutorInputs()
				assert.Len(suite.T(), inputs, 3)
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
				userInputOuDesc:   "Engineering Department",
			},
			expectedOUID: "ou-123",
			expectedRequest: ou.OrganizationUnitRequest{
				Name:        "Engineering",
				Handle:      "engineering",
				Description: "Engineering Department",
			},
			expectedResponse: ou.OrganizationUnit{
				ID:          "ou-123",
				Name:        "Engineering",
				Handle:      "engineering",
				Description: "Engineering Department",
			},
		},
		{
			name: "Create OU without description",
			userInputData: map[string]string{
				userInputOuName:   "Sales",
				userInputOuHandle: "sales",
				userInputOuDesc:   "",
			},
			expectedOUID: "ou-456",
			expectedRequest: ou.OrganizationUnitRequest{
				Name:        "Sales",
				Handle:      "sales",
				Description: "",
			},
			expectedResponse: ou.OrganizationUnit{
				ID:     "ou-456",
				Name:   "Sales",
				Handle: "sales",
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			ctx := &flowmodel.NodeContext{
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

			ctx := &flowmodel.NodeContext{
				FlowID:   "flow-123",
				FlowType: tc.flowType,
			}

			result, err := suite.executor.Execute(ctx)

			assert.NoError(suite.T(), err)
			assert.NotNil(suite.T(), result)
			assert.Equal(suite.T(), flowcm.ExecComplete, result.Status)
			assert.Empty(suite.T(), result.RuntimeData[ouIDKey])
			suite.mockOUService.AssertNotCalled(suite.T(), "CreateOrganizationUnit", mock.Anything)
		})
	}
}

type ExecutePrerequisitesFailureTestCase struct {
	name        string
	ctx         *flowmodel.NodeContext
	expectedMsg string
}

func (suite *OUExecutorTestSuite) TestExecute_PrerequisitesFailure() {
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

	prerequisitesExecutor := &OUExecutor{
		ExecutorInterface: flowmodel.NewExecutor("Test", flowcm.ExecutorTypeUtility,
			defaultInputs, []flowmodel.InputData{{Name: "requiredField", Required: true, Type: "string"}}),
		ouService: suite.mockOUService,
	}

	testCases := []ExecutePrerequisitesFailureTestCase{
		{
			name: "Missing prerequisite field",
			ctx: &flowmodel.NodeContext{
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
			suite.mockOUService.AssertNotCalled(suite.T(), "CreateOrganizationUnit", mock.Anything)
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

			ctx := &flowmodel.NodeContext{
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
				userInputOuDesc:   "",
			},
			expectedRequest: ou.OrganizationUnitRequest{
				Name:        "Engineering",
				Handle:      "engineering",
				Description: "",
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
				userInputOuDesc:   "",
			},
			expectedRequest: ou.OrganizationUnitRequest{
				Name:        "Engineering",
				Handle:      "engineering",
				Description: "",
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
				userInputOuDesc:   "",
			},
			expectedRequest: ou.OrganizationUnitRequest{
				Name:        "Engineering",
				Handle:      "engineering",
				Description: "",
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
				userInputOuDesc:   "",
			},
			expectedRequest: ou.OrganizationUnitRequest{
				Name:        "Engineering",
				Handle:      "engineering",
				Description: "",
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			ctx := &flowmodel.NodeContext{
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

	ctx := &flowmodel.NodeContext{
		FlowID:   "flow-123",
		FlowType: flowcm.FlowTypeRegistration,
		UserInputData: map[string]string{
			userInputOuName:   "Engineering",
			userInputOuHandle: "engineering",
			userInputOuDesc:   "",
		},
		RuntimeData: map[string]string{},
	}

	expectedRequest := ou.OrganizationUnitRequest{
		Name:        "Engineering",
		Handle:      "engineering",
		Description: "",
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
				ctx := &flowmodel.NodeContext{
					UserInputData: map[string]string{},
					RuntimeData:   map[string]string{},
				}
				execResp := &flowmodel.ExecutorResponse{
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
				ctx := &flowmodel.NodeContext{
					UserInputData: map[string]string{},
					RuntimeData:   map[string]string{},
				}
				execResp := &flowmodel.ExecutorResponse{
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
				ctx := &flowmodel.NodeContext{
					UserInputData: map[string]string{},
					RuntimeData:   map[string]string{},
				}

				userID, err := suite.executor.GetUserIDFromContext(ctx)

				assert.NoError(suite.T(), err)
				assert.Empty(suite.T(), userID)
			},
		},
		{
			name: "GetRequiredData returns three required fields",
			testFunc: func() {
				ctx := &flowmodel.NodeContext{
					UserInputData: map[string]string{},
					RuntimeData:   map[string]string{},
				}

				requiredData := suite.executor.GetRequiredData(ctx)

				assert.NotEmpty(suite.T(), requiredData)
				assert.Len(suite.T(), requiredData, 3)
			},
		},
		{
			name: "getOrganizationUnitRequest constructs request correctly",
			testFunc: func() {
				ctx := &flowmodel.NodeContext{
					UserInputData: map[string]string{
						userInputOuName:   "Engineering",
						userInputOuHandle: "engineering",
						userInputOuDesc:   "Engineering Department",
					},
				}

				request := suite.executor.getOrganizationUnitRequest(ctx)

				assert.Equal(suite.T(), "Engineering", request.Name)
				assert.Equal(suite.T(), "engineering", request.Handle)
				assert.Equal(suite.T(), "Engineering Department", request.Description)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, tc.testFunc)
	}
}

func (suite *OUExecutorTestSuite) TestOUExecutorInterface() {
	var _ flowmodel.ExecutorInterface = (*OUExecutor)(nil)
}
