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

package flowregistration

import (
	"fmt"
	"testing"
	"time"

	"github.com/asgardeo/thunder/tests/integration/testutils"
	"github.com/stretchr/testify/suite"
)

const (
	mockNotificationServerPort = 8098
	customSenderName           = "Custom SMS Sender"
)

type SMSRegistrationFlowTestSuite struct {
	suite.Suite
	config     *TestSuiteConfig
	mockServer *testutils.MockNotificationServer
}

func TestSMSRegistrationFlowTestSuite(t *testing.T) {
	suite.Run(t, new(SMSRegistrationFlowTestSuite))
}

func (ts *SMSRegistrationFlowTestSuite) SetupSuite() {
	// Initialize config
	ts.config = &TestSuiteConfig{}

	// Start mock notification server
	ts.mockServer = testutils.NewMockNotificationServer(mockNotificationServerPort)
	err := ts.mockServer.Start()
	if err != nil {
		ts.T().Fatalf("Failed to start mock notification server: %v", err)
	}
	time.Sleep(100 * time.Millisecond)
	ts.T().Log("Mock notification server started successfully")

	// Create custom notification sender
	senderID, err := CreateNotificationSender(mockNotificationServerPort, customSenderName)
	if err != nil {
		ts.T().Fatalf("Failed to create notification sender during setup: %v", err)
	}
	ts.config.CreatedSenderID = senderID
	ts.T().Logf("Notification sender created with ID: %s", ts.config.CreatedSenderID)

	// Store original app config
	ts.config.OriginalAppConfig, err = getAppConfig(appID)
	if err != nil {
		ts.T().Fatalf("Failed to get original app config during setup: %v", err)
	}
}

func (ts *SMSRegistrationFlowTestSuite) TearDownSuite() {
	// Restore original app config
	if ts.config.OriginalAppConfig != nil {
		err := RestoreAppConfig(appID, ts.config.OriginalAppConfig)
		if err != nil {
			ts.T().Logf("Failed to restore original app config during teardown: %v", err)
		}
	}

	// Delete notification sender
	if ts.config.CreatedSenderID != "" {
		err := DeleteNotificationSender(ts.config.CreatedSenderID)
		if err != nil {
			ts.T().Logf("Failed to delete notification sender during teardown: %v", err)
		}
	}

	// Delete test users
	if err := CleanupUsers(ts.config.CreatedUserIDs); err != nil {
		ts.T().Logf("Failed to cleanup users during teardown: %v", err)
	}

	// Stop mock server
	if ts.mockServer != nil {
		err := ts.mockServer.Stop()
		if err != nil {
			ts.T().Logf("Failed to stop mock notification server during teardown: %v", err)
		}
	}
}

func (ts *SMSRegistrationFlowTestSuite) TestSMSRegistrationFlowWithMobileNumber() {
	// Update app to use SMS flow
	err := updateAppConfig(appID, "auth_flow_config_sms", "registration_flow_config_sms")
	if err != nil {
		ts.T().Fatalf("Failed to update app config for SMS flow: %v", err)
	}

	// Generate unique mobile number for registration
	mobileNumber := generateUniqueMobileNumber()

	// Step 1: Initialize the registration flow by calling the flow execution API
	flowStep, err := initiateRegistrationFlow(appID, nil)
	if err != nil {
		ts.T().Fatalf("Failed to initiate registration flow: %v", err)
	}

	ts.Require().Equal("INCOMPLETE", flowStep.FlowStatus, "Expected flow status to be INCOMPLETE")
	ts.Require().Equal("VIEW", flowStep.Type, "Expected flow type to be VIEW")
	ts.Require().NotEmpty(flowStep.FlowID, "Flow ID should not be empty")

	// Validate that mobile number input is required
	ts.Require().NotEmpty(flowStep.Data, "Flow data should not be empty")
	ts.Require().NotEmpty(flowStep.Data.Inputs, "Flow should require inputs")
	ts.Require().True(HasInput(flowStep.Data.Inputs, "mobileNumber"),
		"Mobile number input should be required")

	// Clear any previous messages
	ts.mockServer.ClearMessages()

	// Step 2: Continue the flow with mobile number
	inputs := map[string]string{
		"mobileNumber": mobileNumber,
	}

	otpFlowStep, err := completeRegistrationFlow(flowStep.FlowID, "", inputs)
	if err != nil {
		ts.T().Fatalf("Failed to complete registration flow with mobile number: %v", err)
	}

	// Verify OTP input is now required
	ts.Require().Equal("INCOMPLETE", otpFlowStep.FlowStatus, "Expected flow status to be INCOMPLETE")
	ts.Require().Equal("VIEW", otpFlowStep.Type, "Expected flow type to be VIEW")
	ts.Require().NotEmpty(otpFlowStep.Data, "Flow data should not be empty")
	ts.Require().NotEmpty(otpFlowStep.Data.Inputs, "Flow should require inputs")
	ts.Require().True(HasInput(otpFlowStep.Data.Inputs, "otp"), "OTP input should be required")

	// Wait for SMS to be sent
	time.Sleep(1000 * time.Millisecond)

	// Verify SMS was sent
	lastMessage := ts.mockServer.GetLastMessage()
	ts.Require().NotNil(lastMessage, "Last message should not be nil")
	ts.Require().NotEmpty(lastMessage.OTP, "OTP should be extracted from message")

	// Step 3: Complete registration with OTP
	otpInputs := map[string]string{
		"otp": lastMessage.OTP,
	}

	completeFlowStep, err := completeRegistrationFlow(flowStep.FlowID, "", otpInputs)
	if err != nil {
		ts.T().Fatalf("Failed to complete registration flow with OTP: %v", err)
	}

	ts.Require().Equal("INCOMPLETE", completeFlowStep.FlowStatus,
		"Expected flow status to be INCOMPLETE after OTP input")
	ts.Require().Equal("VIEW", completeFlowStep.Type, "Expected flow type to be VIEW after OTP input")
	ts.Require().NotEmpty(completeFlowStep.Data, "Flow data should not be empty after OTP input")
	ts.Require().NotEmpty(completeFlowStep.Data.Inputs, "Flow should require inputs after OTP input")
	ts.Require().True(ValidateRequiredInputs(completeFlowStep.Data.Inputs, []string{"email"}),
		"Email should be a required inputs after first step")

	// Step 4: Provide additional attributes
	fillInputs := []InputData{
		{
			Name:     "firstName",
			Type:     "string",
			Required: true,
		},
		{
			Name:     "lastName",
			Type:     "string",
			Required: true,
		},
	}
	fillInputs = append(fillInputs, completeFlowStep.Data.Inputs...)
	attrInputs := fillRequiredRegistrationAttributes(fillInputs, mobileNumber)
	completeFlowStep, err = completeRegistrationFlow(flowStep.FlowID, "", attrInputs)
	if err != nil {
		ts.T().Fatalf("Failed to complete registration flow with attributes: %v", err)
	}

	// Verify successful registration
	ts.Require().Equal("COMPLETE", completeFlowStep.FlowStatus, "Expected flow status to be COMPLETE")
	ts.Require().NotEmpty(completeFlowStep.Assertion,
		"JWT assertion should be returned after successful registration")
	ts.Require().Empty(completeFlowStep.FailureReason,
		"Failure reason should be empty for successful registration")

	// Step 5: Verify the user was created by searching via the user API
	user, err := FindUserByAttribute("mobileNumber", mobileNumber)
	if err != nil {
		ts.T().Fatalf("Failed to retrieve user by mobile number: %v", err)
	}
	ts.Require().NotNil(user, "User should be found in user list after registration")

	// Store the created user for cleanup
	if user != nil {
		ts.config.CreatedUserIDs = append(ts.config.CreatedUserIDs, user.Id)
	}
}

func (ts *SMSRegistrationFlowTestSuite) TestSMSRegistrationFlowWithUsername() {
	// Update app to use SMS flow with username
	err := updateAppConfig(appID, "auth_flow_config_sms_with_username", "registration_flow_config_sms_with_username")
	if err != nil {
		ts.T().Fatalf("Failed to update app config for SMS flow with username: %v", err)
	}

	// Generate unique username
	username := generateUniqueUsername("smsreguser")

	// Step 1: Initialize the registration flow
	flowStep, err := initiateRegistrationFlow(appID, nil)
	if err != nil {
		ts.T().Fatalf("Failed to initiate registration flow: %v", err)
	}

	// Validate that username input is required
	ts.Require().Equal("INCOMPLETE", flowStep.FlowStatus, "Expected flow status to be INCOMPLETE")
	ts.Require().Equal("VIEW", flowStep.Type, "Expected flow type to be VIEW")
	ts.Require().NotEmpty(flowStep.FlowID, "Flow ID should not be empty")
	ts.Require().NotEmpty(flowStep.Data, "Flow data should not be empty")
	ts.Require().NotEmpty(flowStep.Data.Inputs, "Flow should require inputs")
	ts.Require().True(HasInput(flowStep.Data.Inputs, "username"), "Username input should be required")

	// Clear any previous messages
	ts.mockServer.ClearMessages()

	// Step 2: Continue the flow with username
	inputs := map[string]string{
		"username": username,
	}
	otpFlowStep, err := completeRegistrationFlow(flowStep.FlowID, "", inputs)
	if err != nil {
		ts.T().Fatalf("Failed to continue registration flow with username: %v", err)
	}

	// Verify that mobile number input is now required
	ts.Require().Equal("INCOMPLETE", otpFlowStep.FlowStatus,
		"Expected flow status to be INCOMPLETE after username input")
	ts.Require().Equal("VIEW", otpFlowStep.Type, "Expected flow type to be VIEW after username input")
	ts.Require().NotEmpty(otpFlowStep.Data, "Flow data should not be empty after username input")
	ts.Require().NotEmpty(otpFlowStep.Data.Inputs, "Flow should require inputs after username input")
	ts.Require().True(HasInput(otpFlowStep.Data.Inputs, "mobileNumber"),
		"Mobile number input should be required after username input")

	// Step 3: Continue the flow with mobile number
	mobileNumber := generateUniqueMobileNumber()
	inputs = map[string]string{
		"mobileNumber": mobileNumber,
	}
	otpFlowStep, err = completeRegistrationFlow(otpFlowStep.FlowID, "", inputs)
	if err != nil {
		ts.T().Fatalf("Failed to continue registration flow with mobile number: %v", err)
	}

	// Verify OTP input is now required
	ts.Require().Equal("INCOMPLETE", otpFlowStep.FlowStatus,
		"Expected flow status to be INCOMPLETE after username input")
	ts.Require().Equal("VIEW", otpFlowStep.Type, "Expected flow type to be VIEW after username input")
	ts.Require().NotEmpty(otpFlowStep.Data, "Flow data should not be empty after username input")
	ts.Require().NotEmpty(otpFlowStep.Data.Inputs, "Flow should require inputs after username input")
	ts.Require().True(HasInput(otpFlowStep.Data.Inputs, "otp"), "OTP input should be required after username input")

	// Wait for SMS to be sent
	time.Sleep(500 * time.Millisecond)

	// Verify SMS was sent
	lastMessage := ts.mockServer.GetLastMessage()
	ts.Require().NotNil(lastMessage, "SMS should have been sent")
	ts.Require().NotEmpty(lastMessage.OTP, "OTP should be available")

	// Step 4: Complete registration with OTP
	otpInputs := map[string]string{
		"otp": lastMessage.OTP,
	}
	completeFlowStep, err := completeRegistrationFlow(flowStep.FlowID, "", otpInputs)
	if err != nil {
		ts.T().Fatalf("Failed to complete registration flow with OTP: %v", err)
	}

	// Verify successful registration
	ts.Require().Equal("COMPLETE", completeFlowStep.FlowStatus, "Expected flow status to be COMPLETE")
	ts.Require().NotEmpty(completeFlowStep.Assertion,
		"JWT assertion should be returned after successful registration")
	ts.Require().Empty(completeFlowStep.FailureReason,
		"Failure reason should be empty for successful registration")

	// Step 5: Verify the user was created by searching via the user API
	user, err := FindUserByAttribute("username", username)
	if err != nil {
		ts.T().Fatalf("Failed to retrieve user by username: %v", err)
	}
	ts.Require().NotNil(user, "User should be found in user list after registration")

	// Store the created user for cleanup
	if user != nil {
		ts.config.CreatedUserIDs = append(ts.config.CreatedUserIDs, user.Id)
	}
}

func (ts *SMSRegistrationFlowTestSuite) TestSMSRegistrationFlowInvalidOTP() {
	// Update app to use SMS flow
	err := updateAppConfig(appID, "auth_flow_config_sms", "registration_flow_config_sms")
	if err != nil {
		ts.T().Fatalf("Failed to update app config for SMS flow: %v", err)
	}

	// Generate unique mobile number
	mobileNumber := generateUniqueMobileNumber()

	// Step 1: Initialize the registration flow and provide mobile number
	inputs := map[string]string{
		"mobileNumber": mobileNumber,
	}

	flowStep, err := initiateRegistrationFlow(appID, inputs)
	if err != nil {
		ts.T().Fatalf("Failed to initiate registration flow: %v", err)
	}

	// Clear any previous messages
	ts.mockServer.ClearMessages()

	// Continue flow to trigger OTP sending
	otpFlowStep, err := completeRegistrationFlow(flowStep.FlowID, "", inputs)
	if err != nil {
		ts.T().Fatalf("Failed to complete registration flow with mobile number: %v", err)
	}

	ts.Require().Equal("INCOMPLETE", otpFlowStep.FlowStatus, "Expected flow status to be INCOMPLETE")

	// Wait for SMS to be sent
	time.Sleep(500 * time.Millisecond)

	// Step 2: Try with invalid OTP
	invalidOTPInputs := map[string]string{
		"otp": "000000",
	}

	completeFlowStep, err := completeRegistrationFlow(flowStep.FlowID, "", invalidOTPInputs)
	if err != nil {
		ts.T().Fatalf("Failed to complete registration flow with invalid OTP: %v", err)
	}

	// Verify registration failure
	ts.Require().Equal("ERROR", completeFlowStep.FlowStatus, "Expected flow status to be ERROR")
	ts.Require().Empty(completeFlowStep.Assertion, "No JWT assertion should be returned for failed registration")
	ts.Require().NotEmpty(completeFlowStep.FailureReason, "Failure reason should be provided for invalid OTP")
	ts.Equal("invalid OTP provided", completeFlowStep.FailureReason,
		"Expected failure reason to indicate invalid OTP")
}

func (ts *SMSRegistrationFlowTestSuite) TestSMSRegistrationFlowSingleRequestWithMobileNumber() {
	// Update app to use SMS flow
	err := updateAppConfig(appID, "auth_flow_config_sms", "registration_flow_config_sms")
	if err != nil {
		ts.T().Fatalf("Failed to update app config for SMS flow: %v", err)
	}

	// Clear any previous messages
	ts.mockServer.ClearMessages()

	// Generate unique mobile number
	mobileNumber := generateUniqueMobileNumber()

	// Step 1: Initialize the registration flow with mobile number
	inputs := map[string]string{
		"mobileNumber": mobileNumber,
		"firstName":    "Test",
		"lastName":     "User",
		"email":        fmt.Sprintf("%s@example.com", mobileNumber),
	}

	flowStep, err := initiateRegistrationFlow(appID, inputs)
	if err != nil {
		ts.T().Fatalf("Failed to initiate registration flow: %v", err)
	}

	// Should require OTP input now
	ts.Require().Equal("INCOMPLETE", flowStep.FlowStatus, "Expected flow status to be INCOMPLETE")
	ts.Require().Equal("VIEW", flowStep.Type, "Expected flow type to be VIEW")

	// Wait for SMS to be sent
	time.Sleep(500 * time.Millisecond)

	// Get the OTP from mock server
	lastMessage := ts.mockServer.GetLastMessage()
	ts.Require().NotNil(lastMessage, "SMS should have been sent")
	ts.Require().NotEmpty(lastMessage.OTP, "OTP should be available")

	// Step 2: Complete with OTP
	otpInputs := map[string]string{
		"otp": lastMessage.OTP,
	}

	completeFlowStep, err := completeRegistrationFlow(flowStep.FlowID, "", otpInputs)
	if err != nil {
		ts.T().Fatalf("Failed to complete registration flow with OTP: %v", err)
	}

	// Verify successful registration
	ts.Require().Equal("COMPLETE", completeFlowStep.FlowStatus, "Expected flow status to be COMPLETE")
	ts.Require().NotEmpty(completeFlowStep.Assertion,
		"JWT assertion should be returned after successful registration")
	ts.Require().Empty(completeFlowStep.FailureReason,
		"Failure reason should be empty for successful registration")

	// Step 3: Verify the user was created by searching via the user API
	user, err := FindUserByAttribute("mobileNumber", mobileNumber)
	if err != nil {
		ts.T().Fatalf("Failed to retrieve user by mobile number: %v", err)
	}
	ts.Require().NotNil(user, "User should be found in user list after registration")

	// Store the created user for cleanup
	if user != nil {
		ts.config.CreatedUserIDs = append(ts.config.CreatedUserIDs, user.Id)
	}
}

// Helper function to generate unique mobile numbers
func generateUniqueMobileNumber() string {
	return fmt.Sprintf("+1234567%d", time.Now().UnixNano()%10000)
}

// Helper to fill required attributes for registration
func fillRequiredRegistrationAttributes(inputs []InputData, mobile string) map[string]string {
	attrInputs := map[string]string{}
	for _, input := range inputs {
		if input.Required {
			switch input.Name {
			case "firstName":
				attrInputs["firstName"] = "Test"
			case "lastName":
				attrInputs["lastName"] = "User"
			case "email":
				attrInputs["email"] = fmt.Sprintf("%s@example.com", mobile)
			default:
				attrInputs[input.Name] = "dummy"
			}
		}
	}
	return attrInputs
}
