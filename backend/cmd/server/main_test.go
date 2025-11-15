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

package main

import (
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/tests/mocks/jwtmock"
)

// CreateSecurityMiddlewareTestSuite defines the test suite for createSecurityMiddleware function
type CreateSecurityMiddlewareTestSuite struct {
	suite.Suite
	logger         *log.Logger
	mockJWTService *jwtmock.JWTServiceInterfaceMock
	mux            *http.ServeMux
}

func TestCreateSecurityMiddlewareTestSuite(t *testing.T) {
	suite.Run(t, new(CreateSecurityMiddlewareTestSuite))
}

func (suite *CreateSecurityMiddlewareTestSuite) SetupTest() {
	suite.logger = log.GetLogger()
	suite.mockJWTService = jwtmock.NewJWTServiceInterfaceMock(suite.T())
	suite.mux = http.NewServeMux()

	// Ensure environment variable is clean before each test
	_ = os.Unsetenv("THUNDER_SKIP_SECURITY")
}

func (suite *CreateSecurityMiddlewareTestSuite) TearDownTest() {
	// Clean up environment variable after each test
	_ = os.Unsetenv("THUNDER_SKIP_SECURITY")
}

// TestCreateSecurityMiddleware_WithEnvironmentVariable tests various THUNDER_SKIP_SECURITY environment variable values
func (suite *CreateSecurityMiddlewareTestSuite) TestCreateSecurityMiddleware_WithEnvironmentVariable() {
	testCases := []struct {
		name               string
		envValue           string
		setEnv             bool
		expectSecuritySkip bool
	}{
		{
			name:               "Security enabled - no env variable",
			setEnv:             false,
			expectSecuritySkip: false,
		},
		{
			name:               "Security disabled - true",
			envValue:           "true",
			setEnv:             true,
			expectSecuritySkip: true,
		},
		{
			name:               "Security enabled - false",
			envValue:           "false",
			setEnv:             true,
			expectSecuritySkip: false,
		},
		{
			name:               "Security enabled - empty string",
			envValue:           "",
			setEnv:             true,
			expectSecuritySkip: false,
		},
		{
			name:               "Security enabled - invalid value",
			envValue:           "yes",
			setEnv:             true,
			expectSecuritySkip: false,
		},
		{
			name:               "Security enabled - uppercase TRUE",
			envValue:           "TRUE",
			setEnv:             true,
			expectSecuritySkip: false,
		},
		{
			name:               "Security enabled - mixed case True",
			envValue:           "True",
			setEnv:             true,
			expectSecuritySkip: false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Setup
			if tc.setEnv {
				_ = os.Setenv("THUNDER_SKIP_SECURITY", tc.envValue)
			} else {
				_ = os.Unsetenv("THUNDER_SKIP_SECURITY")
			}

			// Execute
			handler := createSecurityMiddleware(suite.logger, suite.mux, suite.mockJWTService)

			// Assert
			if tc.expectSecuritySkip {
				assert.Nil(suite.T(), handler, "Handler should be nil when security is skipped")
			} else {
				assert.NotNil(suite.T(), handler, "Handler should not be nil when security is enabled")
			}

			// Cleanup for next iteration
			_ = os.Unsetenv("THUNDER_SKIP_SECURITY")
		})
	}
}

// TestCreateSecurityMiddleware_WithNilLogger tests behavior with nil logger (edge case)
// Note: This test documents current behavior - the function will panic if logger is nil when security is disabled
func (suite *CreateSecurityMiddlewareTestSuite) TestCreateSecurityMiddleware_WithNilLogger() {
	// This test documents that passing nil logger will cause a panic when trying to log warnings
	// In production code, this should never happen as logger is initialized at startup

	suite.Run("NilLogger_SecurityDisabled", func() {
		// Setup
		_ = os.Setenv("THUNDER_SKIP_SECURITY", "true")

		// Execute and Assert - will panic because logger.Warn is called when security is disabled
		assert.Panics(suite.T(), func() {
			createSecurityMiddleware(nil, suite.mux, suite.mockJWTService)
		}, "Should panic when logger is nil and security is disabled")
	})
}

// TestCreateSecurityMiddleware_WithNilJWTService tests behavior with nil JWT service (edge case)
func (suite *CreateSecurityMiddlewareTestSuite) TestCreateSecurityMiddleware_WithNilJWTService() {
	suite.Run("NilJWTService_SecurityDisabled", func() {
		// Setup
		_ = os.Setenv("THUNDER_SKIP_SECURITY", "true")

		// Execute - should not panic since security.Initialize is not called
		handler := createSecurityMiddleware(suite.logger, suite.mux, nil)

		// Assert
		assert.Nil(suite.T(), handler)
	})

	suite.Run("NilJWTService_SecurityEnabled", func() {
		// Setup
		_ = os.Unsetenv("THUNDER_SKIP_SECURITY")

		// Execute - security.Initialize should succeed with nil JWT service
		// The function will create a JWT authenticator but it may fail later during actual authentication
		handler := createSecurityMiddleware(suite.logger, suite.mux, nil)

		// Assert - handler is created successfully
		assert.NotNil(suite.T(), handler)
	})
}

// TestCreateSecurityMiddleware_WithNilMux tests behavior with nil mux (edge case)
func (suite *CreateSecurityMiddlewareTestSuite) TestCreateSecurityMiddleware_WithNilMux() {
	suite.Run("NilMux_SecurityDisabled", func() {
		// Setup
		_ = os.Setenv("THUNDER_SKIP_SECURITY", "true")

		// Execute - should not panic since mux is not used when security is disabled
		handler := createSecurityMiddleware(suite.logger, nil, suite.mockJWTService)

		// Assert
		assert.Nil(suite.T(), handler)
	})

	suite.Run("NilMux_SecurityEnabled", func() {
		// Setup
		_ = os.Unsetenv("THUNDER_SKIP_SECURITY")

		// Execute - security.Initialize should succeed even with nil mux
		// The middleware function wraps the handler, not the mux
		handler := createSecurityMiddleware(suite.logger, nil, suite.mockJWTService)

		// Assert
		assert.NotNil(suite.T(), handler)
	})
}

// TestCreateSecurityMiddleware_MultipleInvocations tests that multiple calls work correctly
func (suite *CreateSecurityMiddlewareTestSuite) TestCreateSecurityMiddleware_MultipleInvocations() {
	// Execute multiple times
	handler1 := createSecurityMiddleware(suite.logger, suite.mux, suite.mockJWTService)
	handler2 := createSecurityMiddleware(suite.logger, suite.mux, suite.mockJWTService)
	handler3 := createSecurityMiddleware(suite.logger, suite.mux, suite.mockJWTService)

	// Assert - each call should return a new handler instance
	assert.NotNil(suite.T(), handler1)
	assert.NotNil(suite.T(), handler2)
	assert.NotNil(suite.T(), handler3)
}

// TestCreateSecurityMiddleware_RuntimeToggle tests toggling security at runtime by changing environment variable
func (suite *CreateSecurityMiddlewareTestSuite) TestCreateSecurityMiddleware_RuntimeToggle() {
	// First call with security enabled
	handler1 := createSecurityMiddleware(suite.logger, suite.mux, suite.mockJWTService)
	assert.NotNil(suite.T(), handler1, "First handler should not be nil")

	// Disable security
	_ = os.Setenv("THUNDER_SKIP_SECURITY", "true")
	handler2 := createSecurityMiddleware(suite.logger, suite.mux, suite.mockJWTService)
	assert.Nil(suite.T(), handler2, "Second handler should be nil when security is disabled")

	// Re-enable security
	_ = os.Unsetenv("THUNDER_SKIP_SECURITY")
	handler3 := createSecurityMiddleware(suite.logger, suite.mux, suite.mockJWTService)
	assert.NotNil(suite.T(), handler3, "Third handler should not be nil after re-enabling security")
}
