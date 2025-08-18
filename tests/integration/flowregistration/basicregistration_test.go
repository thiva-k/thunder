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
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/suite"
)

const (
	appID = "550e8400-e29b-41d4-a716-446655440000" // Default test app ID
)

type BasicRegistrationFlowTestSuite struct {
	suite.Suite
	config *TestSuiteConfig
}

func TestBasicRegistrationFlowTestSuite(t *testing.T) {
	suite.Run(t, new(BasicRegistrationFlowTestSuite))
}

func (ts *BasicRegistrationFlowTestSuite) SetupSuite() {
	// Initialize config
	ts.config = &TestSuiteConfig{}

	// Store original app config
	var err error
	ts.config.OriginalAppConfig, err = getAppConfig(appID)
	if err != nil {
		ts.T().Fatalf("Failed to get original app config during setup: %v", err)
	}

	err = updateAppConfig(appID, "auth_flow_config_basic", "registration_flow_config_basic")
	if err != nil {
		ts.T().Fatalf("Failed to update app config for basic flow: %v", err)
	}
}

func (ts *BasicRegistrationFlowTestSuite) TearDownSuite() {
	if ts.config.OriginalAppConfig != nil {
		err := RestoreAppConfig(appID, ts.config.OriginalAppConfig)
		if err != nil {
			ts.T().Logf("Failed to restore original app config during teardown: %v", err)
		}
	}

	if err := CleanupUsers(ts.config.CreatedUserIDs); err != nil {
		ts.T().Logf("Failed to cleanup users during teardown: %v", err)
	}
}

func (ts *BasicRegistrationFlowTestSuite) TestBasicRegistrationFlowSuccess() {
	// Generate unique username for this test
	username := generateUniqueUsername("reguser")

	// Step 1: Initialize the registration flow
	flowStep, err := initiateRegistrationFlow(appID, nil)
	if err != nil {
		ts.T().Fatalf("Failed to initiate registration flow: %v", err)
	}

	ts.Require().Equal("INCOMPLETE", flowStep.FlowStatus, "Expected flow status to be INCOMPLETE")
	ts.Require().Equal("VIEW", flowStep.Type, "Expected flow type to be VIEW")
	ts.Require().NotEmpty(flowStep.FlowID, "Flow ID should not be empty")
	ts.Require().NotEmpty(flowStep.Data, "Flow data should not be empty")
	ts.Require().NotEmpty(flowStep.Data.Inputs, "Flow should require inputs")

	// Verify username and password are required inputs
	ts.Require().True(ValidateRequiredInputs(flowStep.Data.Inputs, []string{"username", "password"}),
		"Username and password inputs should be required")
	ts.Require().True(HasInput(flowStep.Data.Inputs, "username"), "Username input should be present")
	ts.Require().True(HasInput(flowStep.Data.Inputs, "password"), "Password input should be present")

	// Step 2: Continue the flow with registration credentials
	inputs := map[string]string{
		"username": username,
		"password": "testpassword123",
	}

	completeFlowStep, err := completeRegistrationFlow(flowStep.FlowID, "", inputs)
	if err != nil {
		ts.T().Fatalf("Failed to complete registration flow: %v", err)
	}

	// Step 3: Continue the flow with additional attributes
	ts.Require().Equal("INCOMPLETE", completeFlowStep.FlowStatus,
		"Expected flow status to be INCOMPLETE after first step")
	ts.Require().NotEmpty(completeFlowStep.Data, "Flow data should not be empty after first step")
	ts.Require().NotEmpty(completeFlowStep.Data.Inputs, "Flow should require additional inputs after first step")
	ts.Require().True(ValidateRequiredInputs(completeFlowStep.Data.Inputs,
		[]string{"email", "firstName", "lastName"}),
		"Email, first name, and last name should be required inputs after first step")

	inputs = map[string]string{
		"email":     username + "@example.com",
		"firstName": "Test",
		"lastName":  "User",
	}
	completeFlowStep, err = completeRegistrationFlow(completeFlowStep.FlowID, "", inputs)
	if err != nil {
		ts.T().Fatalf("Failed to complete registration flow with additional attributes: %v", err)
	}

	// Step 4: Verify successful registration
	ts.Require().Equal("COMPLETE", completeFlowStep.FlowStatus, "Expected flow status to be COMPLETE")
	ts.Require().NotEmpty(completeFlowStep.Assertion,
		"JWT assertion should be returned after successful registration")
	ts.Require().Empty(completeFlowStep.FailureReason, "Failure reason should be empty for successful registration")

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

func (ts *BasicRegistrationFlowTestSuite) TestBasicRegistrationFlowDuplicateUser() {
	// Create a test user first
	testUser := User{
		OrganizationUnit: "456e8400-e29b-41d4-a716-446655440001",
		Type:             "person",
		Attributes: json.RawMessage(`{
			"username": "duplicateuser",
			"password": "testpassword",
			"email": "duplicate@example.com",
			"firstName": "Duplicate",
			"lastName": "User"
		}`),
	}

	userIDs, err := CreateMultipleUsers(testUser)
	if err != nil {
		ts.T().Fatalf("Failed to create test user for duplicate test: %v", err)
	}
	ts.config.CreatedUserIDs = append(ts.config.CreatedUserIDs, userIDs...)

	// Step 1: Initialize the registration flow
	flowStep, err := initiateRegistrationFlow(appID, nil)
	if err != nil {
		ts.T().Fatalf("Failed to initiate registration flow: %v", err)
	}

	// Step 2: Try to register with existing username
	inputs := map[string]string{
		"username": "duplicateuser",
		"password": "newpassword123",
	}

	completeFlowStep, err := completeRegistrationFlow(flowStep.FlowID, "", inputs)
	if err != nil {
		ts.T().Fatalf("Failed to complete registration flow: %v", err)
	}

	// Step 3: Verify registration failure due to duplicate username
	ts.Require().Equal("ERROR", completeFlowStep.FlowStatus, "Expected flow status to be ERROR")
	ts.Require().Empty(completeFlowStep.Assertion, "No JWT assertion should be returned for failed registration")
	ts.Require().NotEmpty(completeFlowStep.FailureReason, "Failure reason should be provided for duplicate user")
	ts.Equal("User already exists with the provided username.", completeFlowStep.FailureReason,
		"Failure reason should indicate duplicate username")
}

func (ts *BasicRegistrationFlowTestSuite) TestBasicRegistrationFlowInitialInvalidInput() {
	// Step 1: Initialize the registration flow
	flowStep, err := initiateRegistrationFlow(appID, nil)
	if err != nil {
		ts.T().Fatalf("Failed to initiate registration flow: %v", err)
	}

	// Step 2: Try to register with only the username
	username := generateUniqueUsername("newuser")
	inputs := map[string]string{
		"username": username,
	}
	completeFlowStep, err := completeRegistrationFlow(flowStep.FlowID, "", inputs)
	if err != nil {
		ts.T().Fatalf("Failed to complete registration flow: %v", err)
	}

	// Step 3: Verify flow prompt for username again
	ts.Require().Equal("INCOMPLETE", completeFlowStep.FlowStatus, "Expected flow status to be INCOMPLETE")
	ts.Require().Empty(completeFlowStep.Assertion, "No JWT assertion should be returned for incomplete registration")
	ts.Require().Empty(completeFlowStep.FailureReason, "Failure reason should be empty for incomplete registration")
	ts.Require().True(HasInput(completeFlowStep.Data.Inputs, "password"),
		"Flow should prompt for password after invalid input")

	// Step 4: Continue with the password input
	inputs = map[string]string{
		"password": "testpassword123",
	}
	completeFlowStep, err = completeRegistrationFlow(completeFlowStep.FlowID, "", inputs)
	if err != nil {
		ts.T().Fatalf("Failed to complete registration flow with username input: %v", err)
	}

	// Step 5: Continue the flow with additional attributes
	ts.Require().Equal("INCOMPLETE", completeFlowStep.FlowStatus,
		"Expected flow status to be INCOMPLETE after first step")
	ts.Require().NotEmpty(completeFlowStep.Data, "Flow data should not be empty after first step")
	ts.Require().NotEmpty(completeFlowStep.Data.Inputs, "Flow should require additional inputs after first step")
	ts.Require().True(ValidateRequiredInputs(completeFlowStep.Data.Inputs,
		[]string{"email", "firstName", "lastName"}),
		"Email, first name, and last name should be required inputs after first step")

	inputs = map[string]string{
		"email":     username + "@example.com",
		"firstName": "Test",
		"lastName":  "User",
	}
	completeFlowStep, err = completeRegistrationFlow(completeFlowStep.FlowID, "", inputs)
	if err != nil {
		ts.T().Fatalf("Failed to complete registration flow with additional attributes: %v", err)
	}

	// Step 6: Verify successful registration
	ts.Require().Equal("COMPLETE", completeFlowStep.FlowStatus, "Expected flow status to be COMPLETE")
	ts.Require().NotEmpty(completeFlowStep.Assertion,
		"JWT assertion should be returned after successful registration")
	ts.Require().Empty(completeFlowStep.FailureReason, "Failure reason should be empty for successful registration")

	// Step 7: Verify the user was created by searching via the user API
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

func (ts *BasicRegistrationFlowTestSuite) TestBasicRegistrationFlowSingleRequest() {
	// Generate unique username for this test
	username := generateUniqueUsername("singlereguser")

	// Step 1: Initialize the registration flow with credentials in one request
	inputs := map[string]string{
		"username":  username,
		"password":  "testpassword123",
		"email":     username + "@example.com",
		"firstName": "Single",
		"lastName":  "Request",
	}

	flowStep, err := initiateRegistrationFlow(appID, inputs)
	if err != nil {
		ts.T().Fatalf("Failed to initiate registration flow with inputs: %v", err)
	}

	// Step 2: Verify successful registration in a single request
	ts.Require().Equal("COMPLETE", flowStep.FlowStatus, "Expected flow status to be COMPLETE")
	ts.Require().NotEmpty(flowStep.Assertion,
		"JWT assertion should be returned after successful registration")
	ts.Require().Empty(flowStep.FailureReason, "Failure reason should be empty for successful registration")

	// Step 3: Verify the user was created by searching via the user API
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
