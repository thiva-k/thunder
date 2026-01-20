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

	"github.com/asgardeo/thunder/internal/flow/common"
	"github.com/asgardeo/thunder/internal/flow/core"
	"github.com/asgardeo/thunder/tests/mocks/flow/coremock"
	"github.com/asgardeo/thunder/tests/mocks/notification/notificationmock"
	"github.com/asgardeo/thunder/tests/mocks/observabilitymock"
	"github.com/asgardeo/thunder/tests/mocks/usermock"
)

type SMSAuthExecutorTestSuite struct {
	suite.Suite
	mockUserService   *usermock.UserServiceInterfaceMock
	mockOTPService    *notificationmock.OTPServiceInterfaceMock
	mockFlowFactory   *coremock.FlowFactoryInterfaceMock
	mockObservability *observabilitymock.ObservabilityServiceInterfaceMock
	executor          *smsOTPAuthExecutor
}

func TestSMSAuthExecutorSuite(t *testing.T) {
	suite.Run(t, new(SMSAuthExecutorTestSuite))
}

func (suite *SMSAuthExecutorTestSuite) SetupTest() {
	suite.mockUserService = usermock.NewUserServiceInterfaceMock(suite.T())
	suite.mockOTPService = notificationmock.NewOTPServiceInterfaceMock(suite.T())
	suite.mockFlowFactory = coremock.NewFlowFactoryInterfaceMock(suite.T())
	suite.mockObservability = observabilitymock.NewObservabilityServiceInterfaceMock(suite.T())

	// Default behavior for observability: disabled
	suite.mockObservability.On("IsEnabled").Return(false).Maybe()

	defaultInputs := []common.Input{
		{
			Ref:        "otp_input",
			Identifier: userInputOTP,
			Type:       "OTP_INPUT",
			Required:   true,
		},
	}
	prerequisites := []common.Input{
		mobileNumberInput,
	}

	// Mock identifying executor
	identifyingMock := createMockIdentifyingExecutor(suite.T())
	suite.mockFlowFactory.On("CreateExecutor", ExecutorNameIdentifying, common.ExecutorTypeUtility,
		mock.Anything, mock.Anything).Return(identifyingMock).Maybe()

	// Mock base executor
	mockExec := coremock.NewExecutorInterfaceMock(suite.T())
	mockExec.On("GetName").Return(ExecutorNameSMSAuth).Maybe()
	mockExec.On("GetType").Return(common.ExecutorTypeAuthentication).Maybe()
	mockExec.On("GetDefaultInputs").Return(defaultInputs).Maybe()
	mockExec.On("GetRequiredInputs", mock.Anything).Return(defaultInputs).Maybe()
	mockExec.On("GetPrerequisites").Return(prerequisites).Maybe()
	mockExec.On("ValidatePrerequisites", mock.Anything, mock.Anything).Return(true).Maybe()
	mockExec.On("HasRequiredInputs", mock.Anything, mock.Anything).Return(
		func(ctx *core.NodeContext, execResp *common.ExecutorResponse) bool {
			otp, exists := ctx.UserInputs[userInputOTP]
			if !exists || otp == "" {
				execResp.Inputs = defaultInputs
				execResp.Status = common.ExecUserInputRequired
				return false
			}
			return true
		}).Maybe()

	suite.mockFlowFactory.On("CreateExecutor", ExecutorNameSMSAuth, common.ExecutorTypeAuthentication,
		defaultInputs, prerequisites).Return(mockExec)

	suite.executor = newSMSOTPAuthExecutor(suite.mockFlowFactory, suite.mockUserService,
		suite.mockOTPService, suite.mockObservability)
	// Inject the mock base executor
	suite.executor.ExecutorInterface = mockExec
}

func (suite *SMSAuthExecutorTestSuite) TestValidatePrerequisites_RegistrationFlow_PromptsMobileNumber() {
	// Create a mock that returns false for ValidatePrerequisites (prerequisites not met)
	mockExec := coremock.NewExecutorInterfaceMock(suite.T())
	mockExec.On("ValidatePrerequisites", mock.Anything, mock.Anything).Return(false)
	suite.executor.ExecutorInterface = mockExec

	ctx := &core.NodeContext{
		FlowID:      "test-flow-123",
		FlowType:    common.FlowTypeRegistration,
		UserInputs:  make(map[string]string),
		RuntimeData: make(map[string]string),
	}
	execResp := &common.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	result := suite.executor.ValidatePrerequisites(ctx, execResp)

	// Should return false (prerequisites not met)
	assert.False(suite.T(), result)

	// Should set status to ExecUserInputRequired
	assert.Equal(suite.T(), common.ExecUserInputRequired, execResp.Status)

	// Should return mobile number input
	assert.Len(suite.T(), execResp.Inputs, 1)
	assert.Equal(suite.T(), userAttributeMobileNumber, execResp.Inputs[0].Identifier)
	assert.Equal(suite.T(), "PHONE_INPUT", execResp.Inputs[0].Type)
	assert.Equal(suite.T(), "mobile_number_input", execResp.Inputs[0].Ref)
	assert.True(suite.T(), execResp.Inputs[0].Required)

	// Should include meta for UI rendering
	assert.NotNil(suite.T(), execResp.Meta)
	meta, ok := execResp.Meta.(core.MetaStructure)
	assert.True(suite.T(), ok, "Meta should be of type core.MetaStructure")
	assert.NotEmpty(suite.T(), meta.Components, "Meta should contain components")
}

func (suite *SMSAuthExecutorTestSuite) TestValidatePrerequisites_RegistrationFlow_PrerequisitesMet() {
	// Create a mock that returns true for ValidatePrerequisites (prerequisites met)
	mockExec := coremock.NewExecutorInterfaceMock(suite.T())
	mockExec.On("ValidatePrerequisites", mock.Anything, mock.Anything).Return(true)
	suite.executor.ExecutorInterface = mockExec

	ctx := &core.NodeContext{
		FlowID:   "test-flow-123",
		FlowType: common.FlowTypeRegistration,
		UserInputs: map[string]string{
			userAttributeMobileNumber: "+1234567890",
		},
		RuntimeData: make(map[string]string),
	}
	execResp := &common.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	result := suite.executor.ValidatePrerequisites(ctx, execResp)

	// Should return true (prerequisites met)
	assert.True(suite.T(), result)

	// Status should NOT be set to ExecUserInputRequired
	assert.NotEqual(suite.T(), common.ExecUserInputRequired, execResp.Status)
}

func (suite *SMSAuthExecutorTestSuite) TestValidatePrerequisites_AuthenticationFlow_DoesNotPromptMobile() {
	// Create a mock that returns false initially (prerequisites not met)
	// and also mock additional methods that satisfyPrerequisites might call
	mockExec := coremock.NewExecutorInterfaceMock(suite.T())
	mockExec.On("ValidatePrerequisites", mock.Anything, mock.Anything).Return(false)
	mockExec.On("GetUserIDFromContext", mock.Anything).Return("").Maybe()
	suite.executor.ExecutorInterface = mockExec

	ctx := &core.NodeContext{
		FlowID:      "test-flow-123",
		FlowType:    common.FlowTypeAuthentication, // Authentication flow, NOT registration
		UserInputs:  make(map[string]string),
		RuntimeData: make(map[string]string),
	}
	execResp := &common.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	result := suite.executor.ValidatePrerequisites(ctx, execResp)

	assert.False(suite.T(), result, "Should return false when prerequisites not met")
	assert.NotEqual(suite.T(), common.ExecUserInputRequired, execResp.Status,
		"Authentication flows should not prompt for mobile number directly")
}

func (suite *SMSAuthExecutorTestSuite) TestGetMobileInputMeta() {
	meta := suite.executor.getMobileInputMeta()

	// Should return MetaStructure
	metaStruct, ok := meta.(core.MetaStructure)
	assert.True(suite.T(), ok, "getMobileInputMeta should return core.MetaStructure")
	assert.NotEmpty(suite.T(), metaStruct.Components, "Meta should contain components")

	// Verify components structure
	var hasHeading, hasBlock bool
	for _, comp := range metaStruct.Components {
		if comp.Type == "TEXT" && comp.Variant == "HEADING_2" {
			hasHeading = true
			assert.Equal(suite.T(), "{{ t(signup:heading) }}", comp.Label)
		}
		if comp.Type == "BLOCK" {
			hasBlock = true
			// Block should contain input and action
			assert.GreaterOrEqual(suite.T(), len(comp.Components), 1)

			// Find the input and action within block
			var hasInput, hasAction bool
			for _, blockComp := range comp.Components {
				if blockComp.Type == "PHONE_INPUT" {
					hasInput = true
					assert.Equal(suite.T(), userAttributeMobileNumber, blockComp.Ref)
					assert.Equal(suite.T(), "{{ t(elements:fields.mobile.label) }}", blockComp.Label)
				}
				if blockComp.Type == "ACTION" {
					hasAction = true
					assert.Equal(suite.T(), "{{ t(elements:buttons.submit.text) }}", blockComp.Label)
					assert.Equal(suite.T(), "SUBMIT", blockComp.EventType)
				}
			}
			assert.True(suite.T(), hasInput, "Block should contain PHONE_INPUT component")
			assert.True(suite.T(), hasAction, "Block should contain ACTION component")
		}
	}

	assert.True(suite.T(), hasHeading, "Meta should contain heading")
	assert.True(suite.T(), hasBlock, "Meta should contain block with inputs")
}
