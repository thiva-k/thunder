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

package flowauthn

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/suite"
)

const (
	authFlowConfigBasicAttrCollectTest1 = "auth_flow_config_basic_attr_collect_test_1"
)

// Test users with different attribute configurations
var (
	// User with no attributes stored in profile (all attributes missing)
	testUserNoAttributes = User{
		OrganizationUnit: "456e8400-e29b-41d4-a716-446655440001",
		Type:             "person",
		Attributes: json.RawMessage(`{
			"username": "noattrsuser",
			"password": "testpassword"
		}`),
	}

	// User with partial attributes (email and mobileNumber missing)
	testUserPartialAttributes = User{
		OrganizationUnit: "456e8400-e29b-41d4-a716-446655440001",
		Type:             "person",
		Attributes: json.RawMessage(`{
			"username": "partialuser",
			"password": "testpassword",
			"firstName": "Partial",
			"lastName": "User"
		}`),
	}

	// User with all required attributes present
	testUserFullAttributes = User{
		OrganizationUnit: "456e8400-e29b-41d4-a716-446655440001",
		Type:             "person",
		Attributes: json.RawMessage(`{
			"username": "fulluser",
			"password": "testpassword",
			"firstName": "Full",
			"lastName": "User",
			"email": "fulluser@example.com",
			"mobileNumber": "+1234567890"
		}`),
	}

	// Another user with no attributes stored in profile (all attributes missing)
	testUserNoAttributes2 = User{
		OrganizationUnit: "456e8400-e29b-41d4-a716-446655440001",
		Type:             "person",
		Attributes: json.RawMessage(`{
			"username": "noattrsuser2",
			"password": "testpassword"
		}`),
	}
)

type AttributeCollectTestData struct {
	name                 string
	user                 User
	expectedMissingAttrs []string
	credentials          map[string]string
	providedAttrs        map[string]string
}

type AttributeCollectFlowTestSuite struct {
	suite.Suite
	config   *TestSuiteConfig
	testData []AttributeCollectTestData
}

func TestAttributeCollectFlowTestSuite(t *testing.T) {
	suite.Run(t, new(AttributeCollectFlowTestSuite))
}

func (ts *AttributeCollectFlowTestSuite) SetupSuite() {
	// Initialize config
	ts.config = &TestSuiteConfig{}

	// Store original app config
	originalConfig, err := getAppConfig(appID)
	if err != nil {
		ts.T().Fatalf("Failed to get original app config: %v", err)
	}
	ts.config.OriginalAppConfig = originalConfig

	// Update app to use attribute collection flow
	err = updateAppConfig(appID, authFlowConfigBasicAttrCollectTest1)
	if err != nil {
		ts.T().Fatalf("Failed to update app config: %v", err)
	}

	// Setup test data
	ts.testData = []AttributeCollectTestData{
		{
			name:                 "UserWithNoAttributes",
			user:                 testUserNoAttributes,
			expectedMissingAttrs: []string{"firstName", "lastName", "email", "mobileNumber"},
			credentials: map[string]string{
				"username": "noattrsuser",
				"password": "testpassword",
			},
			providedAttrs: map[string]string{
				"firstName":    "John",
				"lastName":     "Doe",
				"email":        "john.doe@example.com",
				"mobileNumber": "+1987654321",
			},
		},
		{
			name:                 "UserWithPartialAttributes",
			user:                 testUserPartialAttributes,
			expectedMissingAttrs: []string{"email", "mobileNumber"},
			credentials: map[string]string{
				"username": "partialuser",
				"password": "testpassword",
			},
			providedAttrs: map[string]string{
				"email":        "partial@example.com",
				"mobileNumber": "+1555666777",
			},
		},
		{
			name:                 "UserWithFullAttributes",
			user:                 testUserFullAttributes,
			expectedMissingAttrs: []string{},
			credentials: map[string]string{
				"username": "fulluser",
				"password": "testpassword",
			},
			providedAttrs: map[string]string{},
		},
	}

	// Create all test users
	var users []User
	for _, testCase := range ts.testData {
		users = append(users, testCase.user)
	}
	users = append(users, testUserNoAttributes2) // Additional user for second login tests

	userIDs, err := CreateMultipleUsers(users...)
	if err != nil {
		ts.T().Fatalf("Failed to create test users during setup: %v", err)
	}
	ts.config.CreatedUserIDs = userIDs
}

func (ts *AttributeCollectFlowTestSuite) TearDownSuite() {
	// Delete all created users
	if err := CleanupUsers(ts.config.CreatedUserIDs); err != nil {
		ts.T().Logf("Failed to cleanup users during teardown: %v", err)
	}

	// Restore original app config
	if ts.config.OriginalAppConfig != nil {
		err := RestoreAppConfig(appID, ts.config.OriginalAppConfig)
		if err != nil {
			ts.T().Logf("Failed to restore original app config during teardown: %v", err)
		}
	}
}

// TestAttributeCollectionFlow tests the complete attribute collection flow including first and second login
func (ts *AttributeCollectFlowTestSuite) TestAttributeCollectionFlow() {
	for _, testCase := range ts.testData {
		ts.Run(testCase.name, func() {
			// Test First Login - should prompt for missing attributes
			ts.Run("FirstLogin", func() {
				// Step 1: Initialize the flow - should prompt for username/password
				flowStep, err := initiateAuthFlow(appID, nil)
				ts.Require().NoError(err, "Failed to initiate authentication flow")
				ts.Require().Equal("INCOMPLETE", flowStep.FlowStatus, "Expected flow status to be INCOMPLETE")
				ts.Require().Equal("VIEW", flowStep.Type, "Expected flow type to be VIEW")

				// Validate that username and password are required
				ts.validateRequiredInputs(flowStep.Data.Inputs, []string{"username", "password"})

				// Step 2: Provide credentials - should authenticate and proceed to attribute collection
				credentialStep, err := completeAuthFlow(flowStep.FlowID, "", testCase.credentials)
				ts.Require().NoError(err, "Failed to complete basic authentication")

				if len(testCase.expectedMissingAttrs) == 0 {
					// User has all attributes - should complete authentication
					ts.Require().Equal("COMPLETE", credentialStep.FlowStatus,
						"Expected flow to complete for user with all attributes")
					ts.Require().NotEmpty(credentialStep.Assertion, "Expected assertion for completed flow")
				} else {
					// User missing attributes - should prompt for them
					ts.Require().Equal("INCOMPLETE", credentialStep.FlowStatus,
						"Expected flow status to be INCOMPLETE")
					ts.Require().Equal("VIEW", credentialStep.Type, "Expected flow type to be VIEW")

					// Validate that the missing attributes are prompted
					ts.validateRequiredInputs(credentialStep.Data.Inputs, testCase.expectedMissingAttrs)

					// Step 3: Provide missing attributes
					if len(testCase.providedAttrs) > 0 {
						finalStep, err := completeAuthFlow(credentialStep.FlowID, "", testCase.providedAttrs)
						ts.Require().NoError(err, "Failed to complete attribute collection")
						ts.Require().Equal("COMPLETE", finalStep.FlowStatus, "Expected flow status to be COMPLETE")
						ts.Require().NotEmpty(finalStep.Assertion, "Expected assertion after attribute collection")
					}
				}
			})

			// Test Second Login - should not prompt for attributes again
			if len(testCase.expectedMissingAttrs) > 0 {
				ts.Run("SecondLogin", func() {
					// Now perform second login - should not prompt for attributes
					flowStep, err := initiateAuthFlow(appID, nil)
					ts.Require().NoError(err, "Failed to initiate second authentication flow")
					ts.Require().Equal("INCOMPLETE", flowStep.FlowStatus, "Expected flow status to be INCOMPLETE")
					ts.Require().Equal("VIEW", flowStep.Type, "Expected flow type to be VIEW")

					// Provide credentials
					credentialStep, err := completeAuthFlow(flowStep.FlowID, "", testCase.credentials)
					ts.Require().NoError(err, "Failed to complete second authentication")
					ts.Require().Equal("COMPLETE", credentialStep.FlowStatus,
						"Expected flow to complete on second login")
					ts.Require().NotEmpty(credentialStep.Assertion, "Expected assertion on second login")
				})
			}
		})
	}
}

func (ts *AttributeCollectFlowTestSuite) TestSingleRequestLogin_WithAllInputs() {
	flowStep, err := initiateAuthFlow(appID, nil)
	ts.Require().NoError(err, "Failed to initiate authentication flow")
	ts.Require().Equal("INCOMPLETE", flowStep.FlowStatus, "Expected flow status to be INCOMPLETE")
	ts.Require().Equal("VIEW", flowStep.Type, "Expected flow type to be VIEW")
	ts.validateRequiredInputs(flowStep.Data.Inputs, []string{"username", "password"})

	// Provide all required inputs in a single request
	allInputs := map[string]string{
		"username":     "fulluser",
		"password":     "testpassword",
		"firstName":    "Full",
		"lastName":     "User",
		"email":        "john.doe2@example.com",
		"mobileNumber": "+1987654345",
	}
	finalStep, err := completeAuthFlow(flowStep.FlowID, "", allInputs)
	ts.Require().NoError(err, "Failed to complete authentication with all inputs")
	ts.Require().Equal("COMPLETE", finalStep.FlowStatus, "Expected flow status to be COMPLETE")
	ts.Require().NotEmpty(finalStep.Assertion, "Expected assertion after completing flow with all inputs")
}

func (ts *AttributeCollectFlowTestSuite) TestInvalidCredentials() {
	invalidCredentials := map[string]string{
		"username": "invaliduser",
		"password": "wrongpassword",
	}

	flowStep, err := initiateAuthFlow(appID, nil)
	ts.Require().NoError(err, "Failed to initiate authentication flow")

	errorResp, err := completeAuthFlow(flowStep.FlowID, "", invalidCredentials)
	ts.Require().NoError(err, "Expected error response for invalid credentials")
	ts.Require().NotEmpty(errorResp.FailureReason, "Expected failure reason for invalid credentials")
	ts.Require().Equal("User not found", errorResp.FailureReason,
		"Expected failure reason to indicate user not found")
}

func (ts *AttributeCollectFlowTestSuite) validateRequiredInputs(actualInputs []InputData,
	expectedInputNames []string) {
	// Use utility function for basic validation
	ts.Require().True(ValidateRequiredInputs(actualInputs, expectedInputNames),
		"Expected inputs should be present")

	// Additional validation specific to attribute collection
	ts.Require().Len(actualInputs, len(expectedInputNames),
		"Expected %d inputs, got %d", len(expectedInputNames), len(actualInputs))

	actualInputMap := make(map[string]InputData)
	for _, input := range actualInputs {
		actualInputMap[input.Name] = input
	}

	for _, expectedName := range expectedInputNames {
		input, exists := actualInputMap[expectedName]
		ts.Require().True(exists, "Expected input '%s' not found", expectedName)
		ts.Require().Equal("string", input.Type, "Expected input '%s' to be of type string", expectedName)

		// Check if required field is set correctly based on the flow definition
		if expectedName == "mobileNumber" {
			ts.Require().False(input.Required, "Expected input '%s' to be optional", expectedName)
		} else {
			ts.Require().True(input.Required, "Expected input '%s' to be required", expectedName)
		}
	}
}
