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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	authncm "github.com/asgardeo/thunder/internal/authn/common"
	flowcm "github.com/asgardeo/thunder/internal/flow/common"
	flowcore "github.com/asgardeo/thunder/internal/flow/core"
)

type HTTPRequestExecutorTestSuite struct {
	suite.Suite
	executor   *httpRequestExecutor
	mockServer *httptest.Server
}

func TestHTTPRequestExecutorTestSuite(t *testing.T) {
	suite.Run(t, new(HTTPRequestExecutorTestSuite))
}

func (suite *HTTPRequestExecutorTestSuite) SetupTest() {
	flowFactory := flowcore.Initialize()
	suite.executor = newHTTPRequestExecutor(flowFactory)
}

func (suite *HTTPRequestExecutorTestSuite) TearDownTest() {
	if suite.mockServer != nil {
		suite.mockServer.Close()
		suite.mockServer = nil
	}
}

func (suite *HTTPRequestExecutorTestSuite) TestResolvePlaceholder() {
	ctx := &flowcore.NodeContext{
		FlowID: "test-flow",
		UserInputData: map[string]string{
			"username": "testuser",
			"email":    "test@example.com",
		},
		RuntimeData: map[string]string{
			"sessionId": "session-123",
		},
		AuthenticatedUser: authncm.AuthenticatedUser{
			UserID: "user-456",
		},
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Resolve from UserInputData",
			input:    "Hello {{ context.username }}",
			expected: "Hello testuser",
		},
		{
			name:     "Resolve from RuntimeData",
			input:    "Session: {{ context.sessionId }}",
			expected: "Session: session-123",
		},
		{
			name:     "Resolve userID from AuthenticatedUser",
			input:    "User {{ context.userID }} logged in",
			expected: "User user-456 logged in",
		},
		{
			name:     "RuntimeData takes precedence over UserInputData",
			input:    "{{ context.sessionId }}",
			expected: "session-123",
		},
		{
			name:     "Multiple placeholders",
			input:    "{{ context.username }} - {{ context.email }}",
			expected: "testuser - test@example.com",
		},
		{
			name:     "Placeholder with spaces",
			input:    "{{  context.username  }}",
			expected: "testuser",
		},
		{
			name:     "Non-existent placeholder remains unchanged",
			input:    "{{ context.nonexistent }}",
			expected: "{{ context.nonexistent }}",
		},
		{
			name:     "No placeholders",
			input:    "static text",
			expected: "static text",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result := suite.executor.resolvePlaceholder(ctx, tt.input)
			assert.Equal(suite.T(), tt.expected, result)
		})
	}
}

func (suite *HTTPRequestExecutorTestSuite) TestResolvePlaceholder_UserIDSpecialHandling() {
	// Test 1: userID should not resolve from UserInputData
	ctx := &flowcore.NodeContext{
		FlowID: "test-flow",
		UserInputData: map[string]string{
			"userID": "input-user-id", // This should NOT be used
		},
		AuthenticatedUser: authncm.AuthenticatedUser{},
	}

	result := suite.executor.resolvePlaceholder(ctx, "{{ context.userID }}")
	assert.Equal(suite.T(), "{{ context.userID }}", result,
		"userID should not be resolved from UserInputData")

	// Test 2: userID resolves from AuthenticatedUser
	ctx.AuthenticatedUser.UserID = "auth-user-id"
	result = suite.executor.resolvePlaceholder(ctx, "{{ context.userID }}")
	assert.Equal(suite.T(), "auth-user-id", result)

	// Test 3: userID resolves from RuntimeData
	ctx.AuthenticatedUser.UserID = ""
	ctx.RuntimeData = map[string]string{
		"userID": "runtime-user-id",
	}
	result = suite.executor.resolvePlaceholder(ctx, "{{ context.userID }}")
	assert.Equal(suite.T(), "runtime-user-id", result)

	// Test 4: AuthenticatedUser takes precedence over RuntimeData for userID
	ctx.AuthenticatedUser.UserID = "auth-user-id"
	ctx.RuntimeData["userID"] = "runtime-user-id"
	result = suite.executor.resolvePlaceholder(ctx, "{{ context.userID }}")
	assert.Equal(suite.T(), "auth-user-id", result)
}

func (suite *HTTPRequestExecutorTestSuite) TestResolveMapPlaceholders() {
	ctx := &flowcore.NodeContext{
		FlowID: "test-flow",
		UserInputData: map[string]string{
			"username": "testuser",
			"email":    "test@example.com",
		},
		RuntimeData: map[string]string{
			"orgId": "org-123",
		},
	}

	input := map[string]interface{}{
		"user": map[string]interface{}{
			"name":  "{{ context.username }}",
			"email": "{{ context.email }}",
			"metadata": map[string]interface{}{
				"orgId":  "{{ context.orgId }}",
				"static": "value",
			},
		},
		"items": []interface{}{
			"{{ context.username }}",
			"static",
			map[string]interface{}{
				"nested": "{{ context.email }}",
			},
		},
	}

	result := suite.executor.resolveMapPlaceholders(ctx, input)

	expected := map[string]interface{}{
		"user": map[string]interface{}{
			"name":  "testuser",
			"email": "test@example.com",
			"metadata": map[string]interface{}{
				"orgId":  "org-123",
				"static": "value",
			},
		},
		"items": []interface{}{
			"testuser",
			"static",
			map[string]interface{}{
				"nested": "test@example.com",
			},
		},
	}

	assert.Equal(suite.T(), expected, result)
}

func (suite *HTTPRequestExecutorTestSuite) TestExecute_SuccessfulGETRequest() {
	// Setup mock server
	suite.mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(suite.T(), "GET", r.Method)
		assert.Equal(suite.T(), "/api/users/123", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(map[string]interface{}{
			"id":   "123",
			"name": "Test User",
		})
		assert.NoError(suite.T(), err, "Failed to encode mock response")
	}))

	responseMappingJSON := `{"id": "response.data.id", "name": "response.data.name"}`

	ctx := &flowcore.NodeContext{
		FlowID: "test-flow",
		NodeProperties: map[string]interface{}{
			"url":             suite.mockServer.URL + "/api/users/123",
			"method":          "GET",
			"responseMapping": responseMappingJSON,
		},
		UserInputData: make(map[string]string),
		RuntimeData:   make(map[string]string),
	}

	execResp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), flowcm.ExecComplete, execResp.Status)
	assert.Equal(suite.T(), "123", execResp.RuntimeData["id"])
	assert.Equal(suite.T(), "Test User", execResp.RuntimeData["name"])
}

func (suite *HTTPRequestExecutorTestSuite) TestExecute_SuccessfulPOSTRequest() {
	var receivedBody map[string]interface{}
	var receivedHeaders http.Header

	suite.mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(suite.T(), "POST", r.Method)

		receivedHeaders = r.Header
		err := json.NewDecoder(r.Body).Decode(&receivedBody)
		assert.NoError(suite.T(), err, "Failed to decode request body")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		err = json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "created",
			"userId": "new-user-123",
		})
		assert.NoError(suite.T(), err, "Failed to encode mock response")
	}))

	bodyJSON := `{"username": "{{ context.username }}", "email": "{{ context.email }}"}`
	headersJSON := `{"Authorization": "Bearer token123", "X-Custom-Header": "{{ context.customValue }}"}`
	responseMappingJSON := `{"status": "response.data.status", "userId": "response.data.userId"}`

	ctx := &flowcore.NodeContext{
		FlowID: "test-flow",
		NodeProperties: map[string]interface{}{
			"url":             suite.mockServer.URL + "/api/users",
			"method":          "POST",
			"body":            bodyJSON,
			"headers":         headersJSON,
			"responseMapping": responseMappingJSON,
		},
		UserInputData: map[string]string{
			"username":    "newuser",
			"email":       "newuser@example.com",
			"customValue": "custom123",
		},
		RuntimeData: make(map[string]string),
	}

	execResp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), flowcm.ExecComplete, execResp.Status)
	assert.Equal(suite.T(), "created", execResp.RuntimeData["status"])
	assert.Equal(suite.T(), "new-user-123", execResp.RuntimeData["userId"])

	// Verify received body
	assert.Equal(suite.T(), "newuser", receivedBody["username"])
	assert.Equal(suite.T(), "newuser@example.com", receivedBody["email"])

	// Verify headers
	assert.Equal(suite.T(), "Bearer token123", receivedHeaders.Get("Authorization"))
	assert.Equal(suite.T(), "custom123", receivedHeaders.Get("X-Custom-Header"))
}

func (suite *HTTPRequestExecutorTestSuite) TestExecute_ResponseMapping() {
	suite.mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"userId":     "user-789",
				"profileUrl": "https://example.com/profile",
			},
			"metadata": map[string]interface{}{
				"timestamp": "2025-11-12T10:00:00Z",
			},
		})
		assert.NoError(suite.T(), err, "Failed to encode mock response")
	}))

	responseMappingJSON := `{"externalUserId": "response.data.data.userId", 
	"profileUrl": "response.data.data.profileUrl", "timestamp": "response.data.metadata.timestamp"}`

	ctx := &flowcore.NodeContext{
		FlowID: "test-flow",
		NodeProperties: map[string]interface{}{
			"url":             suite.mockServer.URL + "/api/data",
			"responseMapping": responseMappingJSON,
		},
		UserInputData: make(map[string]string),
		RuntimeData:   make(map[string]string),
	}

	execResp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), flowcm.ExecComplete, execResp.Status)
	assert.Equal(suite.T(), "user-789", execResp.RuntimeData["externalUserId"])
	assert.Equal(suite.T(), "https://example.com/profile", execResp.RuntimeData["profileUrl"])
	assert.Equal(suite.T(), "2025-11-12T10:00:00Z", execResp.RuntimeData["timestamp"])
	// Original keys should not be present when mapping is specified
	assert.Empty(suite.T(), execResp.RuntimeData["data"])
}

func (suite *HTTPRequestExecutorTestSuite) TestExecute_DefaultMethod() {
	suite.mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(suite.T(), "GET", r.Method)
		w.WriteHeader(http.StatusOK)
	}))

	ctx := &flowcore.NodeContext{
		FlowID: "test-flow",
		NodeProperties: map[string]interface{}{
			"url": suite.mockServer.URL + "/api/test",
			// method not specified, should default to GET
		},
		UserInputData: make(map[string]string),
		RuntimeData:   make(map[string]string),
	}

	execResp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), flowcm.ExecComplete, execResp.Status)
}

func (suite *HTTPRequestExecutorTestSuite) TestExecute_ErrorHandling_FailOnErrorFalse() {
	suite.mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte("Internal Server Error"))
		assert.NoError(suite.T(), err, "Failed to write mock error response")
	}))

	ctx := &flowcore.NodeContext{
		FlowID: "test-flow",
		NodeProperties: map[string]interface{}{
			"url": suite.mockServer.URL + "/api/error",
		},
		UserInputData: make(map[string]string),
		RuntimeData:   make(map[string]string),
	}

	execResp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	// Should complete without failure when failOnError defaults to false
	assert.Equal(suite.T(), flowcm.ExecComplete, execResp.Status)
}

func (suite *HTTPRequestExecutorTestSuite) TestExecute_ErrorHandling_FailOnErrorTrue() {
	suite.mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte("Bad Request"))
		assert.NoError(suite.T(), err, "Failed to write mock error response")
	}))

	errorHandlingJSON := `{"failOnError": true}`

	ctx := &flowcore.NodeContext{
		FlowID: "test-flow",
		NodeProperties: map[string]interface{}{
			"url":           suite.mockServer.URL + "/api/error",
			"errorHandling": errorHandlingJSON,
		},
		UserInputData: make(map[string]string),
		RuntimeData:   make(map[string]string),
	}

	execResp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), flowcm.ExecFailure, execResp.Status)
	assert.Contains(suite.T(), execResp.FailureReason, "HTTP request failed with status 400")
}

func (suite *HTTPRequestExecutorTestSuite) TestExecute_MissingURL() {
	ctx := &flowcore.NodeContext{
		FlowID: "test-flow",
		NodeProperties: map[string]interface{}{
			// URL is missing
			"method": "GET",
		},
		UserInputData: make(map[string]string),
		RuntimeData:   make(map[string]string),
	}

	execResp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	// Configuration errors always fail the flow regardless of failOnError setting
	assert.Equal(suite.T(), flowcm.ExecFailure, execResp.Status)
	assert.Contains(suite.T(), execResp.FailureReason, "url is required")
}

func (suite *HTTPRequestExecutorTestSuite) TestExecute_InvalidHTTPMethod() {
	ctx := &flowcore.NodeContext{
		FlowID: "test-flow",
		NodeProperties: map[string]interface{}{
			"url":    "https://example.com/api/test",
			"method": "INVALID",
		},
		UserInputData: make(map[string]string),
		RuntimeData:   make(map[string]string),
	}

	execResp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	// Configuration errors always fail the flow regardless of failOnError setting
	assert.Equal(suite.T(), flowcm.ExecFailure, execResp.Status)
	assert.Contains(suite.T(), execResp.FailureReason, "invalid HTTP method")
}

func (suite *HTTPRequestExecutorTestSuite) TestParseAndValidateConfig_TimeoutLimits() {
	// Test timeout exceeding maximum
	properties := map[string]interface{}{
		"url":     "https://example.com/api/test",
		"timeout": "60", // Exceeds max of 30
	}

	config, err := suite.executor.parseAndValidateConfig(properties)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), maxHTTPRequestTimeout, config.Timeout, "Timeout should be capped at maximum")

	// Test default timeout
	properties2 := map[string]interface{}{
		"url": "https://example.com/api/test",
	}

	config2, err := suite.executor.parseAndValidateConfig(properties2)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), defaultHTTPTimeout, config2.Timeout)
}

func (suite *HTTPRequestExecutorTestSuite) TestParseAndValidateConfig_RetryLimits() {
	errorHandlingJSON := `{"retryCount": 10, "retryDelay": 10000}`

	properties := map[string]interface{}{
		"url":           "https://example.com/api/test",
		"errorHandling": errorHandlingJSON,
	}

	config, err := suite.executor.parseAndValidateConfig(properties)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), maxHTTPRequestRetryCount, config.ErrorHandling.RetryCount)
	assert.Equal(suite.T(), maxHTTPRequestRetryDelay, config.ErrorHandling.RetryDelay)
}

func (suite *HTTPRequestExecutorTestSuite) TestExecute_AllHTTPMethods() {
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}

	for _, method := range methods {
		suite.Run("Method_"+method, func() {
			suite.mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(suite.T(), method, r.Method)
				w.WriteHeader(http.StatusOK)
			}))
			defer suite.mockServer.Close()

			ctx := &flowcore.NodeContext{
				FlowID: "test-flow",
				NodeProperties: map[string]interface{}{
					"url":    suite.mockServer.URL + "/api/test",
					"method": method,
				},
				UserInputData: make(map[string]string),
				RuntimeData:   make(map[string]string),
			}

			execResp, err := suite.executor.Execute(ctx)

			assert.NoError(suite.T(), err)
			assert.Equal(suite.T(), flowcm.ExecComplete, execResp.Status)
		})
	}
}

func (suite *HTTPRequestExecutorTestSuite) TestExtractValueFromPath() {
	data := map[string]interface{}{
		"user": map[string]interface{}{
			"id":   "123",
			"name": "Test User",
			"profile": map[string]interface{}{
				"email": "test@example.com",
			},
		},
		"count": 42,
	}

	tests := []struct {
		name     string
		path     string
		expected interface{}
	}{
		{
			name:     "Top level string",
			path:     "count",
			expected: 42,
		},
		{
			name:     "Nested string",
			path:     "user.id",
			expected: "123",
		},
		{
			name:     "Deeply nested string",
			path:     "user.profile.email",
			expected: "test@example.com",
		},
		{
			name:     "Non-existent path",
			path:     "user.nonexistent",
			expected: nil,
		},
		{
			name:     "Empty path",
			path:     "",
			expected: data,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result := suite.executor.extractValueFromPath(data, tt.path)
			assert.Equal(suite.T(), tt.expected, result)
		})
	}
}

func (suite *HTTPRequestExecutorTestSuite) TestExecute_NonJSONResponse() {
	suite.mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("Plain text response"))
		assert.NoError(suite.T(), err, "Failed to write mock plain text response")
	}))

	responseMappingJSON := `{"raw": "response.data.raw"}`

	ctx := &flowcore.NodeContext{
		FlowID: "test-flow",
		NodeProperties: map[string]interface{}{
			"url":             suite.mockServer.URL + "/api/text",
			"responseMapping": responseMappingJSON,
		},
		UserInputData: make(map[string]string),
		RuntimeData:   make(map[string]string),
	}

	execResp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), flowcm.ExecComplete, execResp.Status)
	assert.Equal(suite.T(), "Plain text response", execResp.RuntimeData["raw"])
}

func (suite *HTTPRequestExecutorTestSuite) TestExecute_ResponseStatusExtraction() {
	suite.mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		err := json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      "123",
			"message": "Resource created",
		})
		assert.NoError(suite.T(), err, "Failed to encode mock response")
	}))

	responseMappingJSON := `{"resourceId": "response.data.id", "statusCode": "response.status"}`

	ctx := &flowcore.NodeContext{
		FlowID: "test-flow",
		NodeProperties: map[string]interface{}{
			"url":             suite.mockServer.URL + "/api/resource",
			"method":          "POST",
			"responseMapping": responseMappingJSON,
		},
		UserInputData: make(map[string]string),
		RuntimeData:   make(map[string]string),
	}

	execResp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), flowcm.ExecComplete, execResp.Status)
	assert.Equal(suite.T(), "123", execResp.RuntimeData["resourceId"])
	assert.Equal(suite.T(), "201", execResp.RuntimeData["statusCode"])
}
