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

package context

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ContextTestSuite struct {
	suite.Suite
}

func TestContextSuite(t *testing.T) {
	suite.Run(t, new(ContextTestSuite))
}

func (s *ContextTestSuite) TestGetTraceID_WithNilContext() {
	traceID := GetTraceID(nil) //nolint:staticcheck // Testing nil context handling
	if traceID == "" {
		s.T().Error("Expected non-empty trace ID, got empty string")
	}
}

func (s *ContextTestSuite) TestGetTraceID_WithEmptyContext() {
	ctx := context.Background()
	traceID := GetTraceID(ctx)
	if traceID == "" {
		s.T().Error("Expected non-empty trace ID, got empty string")
	}
}

func (s *ContextTestSuite) TestGetTraceID_WithExistingTraceID() {
	expectedID := "test-trace-id-123"
	ctx := WithTraceID(context.Background(), expectedID)

	traceID := GetTraceID(ctx)
	if traceID != expectedID {
		s.T().Errorf("Expected trace ID %s, got %s", expectedID, traceID)
	}
}

func (s *ContextTestSuite) TestWithTraceID() {
	expectedID := "custom-trace-id"
	ctx := WithTraceID(context.Background(), expectedID)

	traceID := GetTraceID(ctx)
	if traceID != expectedID {
		s.T().Errorf("Expected trace ID %s, got %s", expectedID, traceID)
	}
}

func (s *ContextTestSuite) TestWithTraceID_NilContext() {
	expectedID := "custom-trace-id"
	ctx := WithTraceID(nil, expectedID) //nolint:staticcheck // Testing nil context handling

	traceID := GetTraceID(ctx)
	if traceID != expectedID {
		s.T().Errorf("Expected trace ID %s, got %s", expectedID, traceID)
	}
}

func (s *ContextTestSuite) TestEnsureTraceID_CreatesNew() {
	ctx := context.Background()
	ctx = EnsureTraceID(ctx)

	traceID := GetTraceID(ctx)
	if traceID == "" {
		s.T().Error("Expected non-empty trace ID after EnsureTraceID")
	}
}

func (s *ContextTestSuite) TestEnsureTraceID_PreservesExisting() {
	expectedID := "existing-trace-id"
	ctx := WithTraceID(context.Background(), expectedID)
	ctx = EnsureTraceID(ctx)

	traceID := GetTraceID(ctx)
	if traceID != expectedID {
		s.T().Errorf("Expected trace ID %s to be preserved, got %s", expectedID, traceID)
	}
}

func (s *ContextTestSuite) TestEnsureTraceID_NilContext() {
	ctx := EnsureTraceID(nil) //nolint:staticcheck // Testing nil context handling

	traceID := GetTraceID(ctx)
	if traceID == "" {
		s.T().Error("Expected non-empty trace ID after EnsureTraceID with nil context")
	}
}

func (s *ContextTestSuite) TestGetTraceID_WithNilContextExplicit() {
	// Test with explicitly nil context
	traceID := GetTraceID(nil) //nolint:staticcheck // Testing nil context handling
	if traceID == "" {
		s.T().Error("Expected non-empty trace ID with nil context, got empty string")
	}
}

func (s *ContextTestSuite) TestGetTraceID_WithEmptyTraceIDValue() {
	// Test with empty trace ID value in context
	ctx := context.WithValue(context.Background(), TraceIDKey, "")
	traceID := GetTraceID(ctx)
	if traceID == "" {
		s.T().Error("Expected non-empty trace ID when context has empty trace ID, got empty string")
	}
}

func (s *ContextTestSuite) TestGetTraceID_WithWrongType() {
	// Test with wrong type in context (not a string)
	ctx := context.WithValue(context.Background(), TraceIDKey, 12345)
	traceID := GetTraceID(ctx)
	if traceID == "" {
		s.T().Error("Expected non-empty trace ID when context has wrong type, got empty string")
	}
}

func (s *ContextTestSuite) TestWithTraceID_WithNilContextExplicit() {
	// Test with explicitly nil context
	expectedID := "test-id-nil-ctx"
	ctx := WithTraceID(nil, expectedID) //nolint:staticcheck // Testing nil context handling

	traceID := GetTraceID(ctx)
	if traceID != expectedID {
		s.T().Errorf("Expected trace ID %s with nil context, got %s", expectedID, traceID)
	}
}

func (s *ContextTestSuite) TestWithTraceID_EmptyString() {
	// Test setting empty trace ID
	ctx := WithTraceID(context.Background(), "")

	// Should still set it in context even if empty
	if ctx == nil {
		s.T().Error("Expected non-nil context")
	}

	// GetTraceID should generate a new one since it's empty
	traceID := GetTraceID(ctx)
	if traceID == "" {
		s.T().Error("Expected GetTraceID to generate new ID when stored value is empty")
	}
}

func (s *ContextTestSuite) TestEnsureTraceID_WithNilContextExplicit() {
	// Test with explicitly nil context
	ctx := EnsureTraceID(nil) //nolint:staticcheck // Testing nil context handling

	if ctx == nil {
		s.T().Error("Expected non-nil context")
	}

	traceID := GetTraceID(ctx)
	if traceID == "" {
		s.T().Error("Expected non-empty trace ID after EnsureTraceID with nil context")
	}
}

func (s *ContextTestSuite) TestEnsureTraceID_WithEmptyTraceID() {
	// Test with context containing empty trace ID
	ctx := context.WithValue(context.Background(), TraceIDKey, "")
	ctx = EnsureTraceID(ctx)

	traceID := GetTraceID(ctx)
	if traceID == "" {
		s.T().Error("Expected EnsureTraceID to generate new ID when empty")
	}
}

func (s *ContextTestSuite) TestEnsureTraceID_WithWrongType() {
	// Test with context containing wrong type for trace ID
	ctx := context.WithValue(context.Background(), TraceIDKey, 999)
	ctx = EnsureTraceID(ctx)

	traceID := GetTraceID(ctx)
	if traceID == "" {
		s.T().Error("Expected EnsureTraceID to generate new ID when type is wrong")
	}

	// Verify it's now a string in the context
	if _, ok := ctx.Value(TraceIDKey).(string); !ok {
		s.T().Error("Expected trace ID to be a string after EnsureTraceID")
	}
}

func (s *ContextTestSuite) TestGenerateUUID_Format() {
	uuid := generateUUID()

	// UUID should be in format: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	if len(uuid) != 36 {
		s.T().Errorf("Expected UUID length of 36, got %d", len(uuid))
	}

	// Check for dashes at correct positions
	if uuid[8] != '-' || uuid[13] != '-' || uuid[18] != '-' || uuid[23] != '-' {
		s.T().Errorf("UUID format incorrect: %s", uuid)
	}
}

func (s *ContextTestSuite) TestGenerateUUID_Uniqueness() {
	uuid1 := generateUUID()
	uuid2 := generateUUID()

	if uuid1 == uuid2 {
		s.T().Error("Expected different UUIDs, got same value")
	}
}

func (s *ContextTestSuite) TestGenerateUUID_Version4() {
	uuid := generateUUID()

	// Check version bits (should be 0100 for version 4)
	// The version is in the 7th byte (after 3rd dash at position 14-15)
	// Format: xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx where 4 is version
	if uuid[14] != '4' {
		s.T().Errorf("Expected UUID version 4, got version %c", uuid[14])
	}

	// Check variant bits (should be 10xx for RFC 4122)
	// The variant is in the 9th byte (after 4th dash at position 19)
	// Valid values are 8, 9, a, b (binary 10xx)
	variantChar := uuid[19]
	validVariants := map[byte]bool{'8': true, '9': true, 'a': true, 'b': true}
	if !validVariants[variantChar] {
		s.T().Errorf("Expected UUID variant to be 8, 9, a, or b, got %c", variantChar)
	}
}

func (s *ContextTestSuite) TestGenerateUUID_MultipleGenerations() {
	// Generate many UUIDs to ensure no collisions
	seen := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		uuid := generateUUID()
		if seen[uuid] {
			s.T().Errorf("Duplicate UUID generated: %s", uuid)
		}
		seen[uuid] = true
	}
}

// ============================================================================
// Authentication Context Tests
// ============================================================================

func (s *ContextTestSuite) TestNewAuthenticationContext() {
	userID := "user123"
	ouID := "ou456"
	appID := "app789"
	token := "test-token-123"
	claims := map[string]interface{}{
		"sub":   userID,
		"scope": "users:read users:write",
		"roles": []string{"admin", "user"},
	}

	authCtx := NewAuthenticationContext(userID, ouID, appID, token, claims)

	if authCtx == nil {
		s.T().Fatal("Expected non-nil AuthenticationContext")
	}

	// Access the context through the getter methods
	ctx := WithAuthenticationContext(context.Background(), authCtx)

	if GetUserID(ctx) != userID {
		s.T().Errorf("Expected userID %s, got %s", userID, GetUserID(ctx))
	}

	if GetOUID(ctx) != ouID {
		s.T().Errorf("Expected ouID %s, got %s", ouID, GetOUID(ctx))
	}

	if GetAppID(ctx) != appID {
		s.T().Errorf("Expected appID %s, got %s", appID, GetAppID(ctx))
	}
}

func (s *ContextTestSuite) TestWithAuthenticationContext_NilContext() {
	authCtx := NewAuthenticationContext("user123", "ou456", "app789", "token", map[string]interface{}{
		"sub": "user123",
	})

	ctx := WithAuthenticationContext(nil, authCtx) //nolint:staticcheck // Testing nil context handling

	if ctx == nil {
		s.T().Fatal("Expected non-nil context")
	}

	if GetUserID(ctx) != "user123" {
		s.T().Error("Expected userID to be accessible from context created with nil base")
	}
}

func (s *ContextTestSuite) TestWithAuthenticationContext_NilAuthContext() {
	ctx := WithAuthenticationContext(context.Background(), nil)

	if ctx == nil {
		s.T().Fatal("Expected non-nil context even with nil AuthenticationContext")
	}

	// Should return empty values when no auth context is set
	if GetUserID(ctx) != "" {
		s.T().Error("Expected empty userID when AuthenticationContext is nil")
	}
}

func (s *ContextTestSuite) TestGetUserID() {
	tests := []struct {
		name     string
		setup    func() context.Context
		expected string
	}{
		{
			name: "Valid authentication context",
			setup: func() context.Context {
				authCtx := NewAuthenticationContext("user123", "ou456", "app789", "token", nil)
				return WithAuthenticationContext(context.Background(), authCtx)
			},
			expected: "user123",
		},
		{
			name: "Nil context",
			setup: func() context.Context {
				return nil
			},
			expected: "",
		},
		{
			name:     "Context without authentication",
			setup:    context.Background,
			expected: "",
		},
		{
			name: "Context with nil authentication context",
			setup: func() context.Context {
				return WithAuthenticationContext(context.Background(), nil)
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			ctx := tt.setup()
			userID := GetUserID(ctx)
			if userID != tt.expected {
				t.Errorf("Expected userID %s, got %s", tt.expected, userID)
			}
		})
	}
}

func (s *ContextTestSuite) TestGetOUID() {
	s.testContextGetter("ouID", "ou456", GetOUID)
}

func (s *ContextTestSuite) TestGetAppID() {
	s.testContextGetter("appID", "app789", GetAppID)
}

// testContextGetter is a helper function to test context getter functions
func (s *ContextTestSuite) testContextGetter(fieldName, expectedValue string, getter func(context.Context) string) {
	tests := []struct {
		name     string
		setup    func() context.Context
		expected string
	}{
		{
			name: "Valid authentication context",
			setup: func() context.Context {
				authCtx := NewAuthenticationContext("user123", "ou456", "app789", "token", nil)
				return WithAuthenticationContext(context.Background(), authCtx)
			},
			expected: expectedValue,
		},
		{
			name:     "Nil context",
			setup:    func() context.Context { return nil },
			expected: "",
		},
		{
			name:     "Context without authentication",
			setup:    context.Background,
			expected: "",
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			ctx := tt.setup()
			result := getter(ctx)
			if result != tt.expected {
				t.Errorf("Expected %s %s, got %s", fieldName, tt.expected, result)
			}
		})
	}
}

const (
	testModifiedValue = "modified"
)

func (s *ContextTestSuite) TestGetClaim() {
	claims := map[string]interface{}{
		"string_claim":    "string_value",
		"int_claim":       42,
		"bool_claim":      true,
		"string_slice":    []string{"a", "b", "c"},
		"interface_slice": []interface{}{"x", "y", "z"},
		"map_claim": map[string]interface{}{
			"nested": "value",
		},
		"nil_claim": nil,
	}

	authCtx := NewAuthenticationContext("user", "ou", "app", "token", claims)
	ctx := WithAuthenticationContext(context.Background(), authCtx)

	tests := []struct {
		name     string
		key      string
		validate func(t *testing.T, value interface{})
	}{
		{
			name: "String claim",
			key:  "string_claim",
			validate: func(t *testing.T, value interface{}) {
				if str, ok := value.(string); !ok || str != "string_value" {
					t.Errorf("Expected string 'string_value', got %v", value)
				}
			},
		},
		{
			name: "Int claim",
			key:  "int_claim",
			validate: func(t *testing.T, value interface{}) {
				if num, ok := value.(int); !ok || num != 42 {
					t.Errorf("Expected int 42, got %v", value)
				}
			},
		},
		{
			name: "Bool claim",
			key:  "bool_claim",
			validate: func(t *testing.T, value interface{}) {
				if b, ok := value.(bool); !ok || b != true {
					t.Errorf("Expected bool true, got %v", value)
				}
			},
		},
		{
			name: "String slice (defensive copy)",
			key:  "string_slice",
			validate: func(t *testing.T, value interface{}) {
				slice, ok := value.([]string)
				if !ok {
					t.Errorf("Expected []string, got %T", value)
					return
				}
				if len(slice) != 3 || slice[0] != "a" || slice[1] != "b" || slice[2] != "c" {
					t.Errorf("Expected [a b c], got %v", slice)
					return
				}
				// Verify it's a defensive copy by modifying it
				slice[0] = testModifiedValue
				// Get the claim again and verify it wasn't modified
				freshClaim := GetClaim(ctx, "string_slice")
				freshSlice := freshClaim.([]string)
				if freshSlice[0] != "a" {
					t.Error("Defensive copy failed - original slice was modified")
				}
			},
		},
		{
			name: "Interface slice (defensive copy)",
			key:  "interface_slice",
			validate: func(t *testing.T, value interface{}) {
				slice, ok := value.([]interface{})
				if !ok {
					t.Errorf("Expected []interface{}, got %T", value)
					return
				}
				if len(slice) != 3 {
					t.Errorf("Expected length 3, got %d", len(slice))
					return
				}
				// Verify it's a defensive copy
				slice[0] = testModifiedValue
				freshClaim := GetClaim(ctx, "interface_slice")
				freshSlice := freshClaim.([]interface{})
				if freshSlice[0] != "x" {
					t.Error("Defensive copy failed - original slice was modified")
				}
			},
		},
		{
			name: "Map claim (defensive copy)",
			key:  "map_claim",
			validate: func(t *testing.T, value interface{}) {
				m, ok := value.(map[string]interface{})
				if !ok {
					t.Errorf("Expected map[string]interface{}, got %T", value)
					return
				}
				if m["nested"] != "value" {
					t.Errorf("Expected nested value, got %v", m["nested"])
					return
				}
				// Verify it's a defensive copy
				m["modified"] = "new_value"
				freshClaim := GetClaim(ctx, "map_claim")
				freshMap := freshClaim.(map[string]interface{})
				if _, exists := freshMap["modified"]; exists {
					t.Error("Defensive copy failed - original map was modified")
				}
			},
		},
		{
			name: "Nil claim",
			key:  "nil_claim",
			validate: func(t *testing.T, value interface{}) {
				if value != nil {
					t.Errorf("Expected nil, got %v", value)
				}
			},
		},
		{
			name: "Non-existent claim",
			key:  "does_not_exist",
			validate: func(t *testing.T, value interface{}) {
				if value != nil {
					t.Errorf("Expected nil for non-existent claim, got %v", value)
				}
			},
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			value := GetClaim(ctx, tt.key)
			tt.validate(t, value)
		})
	}

	// Test with nil context
	s.T().Run("Nil context", func(t *testing.T) {
		value := GetClaim(nil, "any_key") //nolint:staticcheck // Testing nil context handling
		if value != nil {
			t.Errorf("Expected nil for nil context, got %v", value)
		}
	})

	// Test with context without auth
	s.T().Run("Context without auth", func(t *testing.T) {
		value := GetClaim(context.Background(), "any_key")
		if value != nil {
			t.Errorf("Expected nil for context without auth, got %v", value)
		}
	})
}

func (s *ContextTestSuite) TestGetClaimString() {
	claims := map[string]interface{}{
		"string_claim": "string_value",
		"int_claim":    42,
		"bool_claim":   true,
		"nil_claim":    nil,
	}

	authCtx := NewAuthenticationContext("user", "ou", "app", "token", claims)
	ctx := WithAuthenticationContext(context.Background(), authCtx)

	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{
			name:     "Valid string claim",
			key:      "string_claim",
			expected: "string_value",
		},
		{
			name:     "Non-string claim",
			key:      "int_claim",
			expected: "",
		},
		{
			name:     "Bool claim",
			key:      "bool_claim",
			expected: "",
		},
		{
			name:     "Nil claim",
			key:      "nil_claim",
			expected: "",
		},
		{
			name:     "Non-existent claim",
			key:      "does_not_exist",
			expected: "",
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			result := GetClaimString(ctx, tt.key)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}

	// Test with nil context
	s.T().Run("Nil context", func(t *testing.T) {
		result := GetClaimString(nil, "any_key") //nolint:staticcheck // Testing nil context handling
		if result != "" {
			t.Errorf("Expected empty string for nil context, got %s", result)
		}
	})

	// Test with context without auth
	s.T().Run("Context without auth", func(t *testing.T) {
		result := GetClaimString(context.Background(), "any_key")
		if result != "" {
			t.Errorf("Expected empty string for context without auth, got %s", result)
		}
	})
}

func (s *ContextTestSuite) TestGetClaimStrings() {
	claims := map[string]interface{}{
		"string_slice":    []string{"a", "b", "c"},
		"interface_slice": []interface{}{"x", "y", "z"},
		"string_claim":    "not_a_slice",
		"int_slice":       []int{1, 2, 3},
		"nil_claim":       nil,
	}

	authCtx := NewAuthenticationContext("user", "ou", "app", "token", claims)
	ctx := WithAuthenticationContext(context.Background(), authCtx)

	tests := []struct {
		name     string
		key      string
		expected []string
	}{
		{
			name:     "Valid string slice",
			key:      "string_slice",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "Interface slice (should return nil)",
			key:      "interface_slice",
			expected: nil,
		},
		{
			name:     "Non-slice claim",
			key:      "string_claim",
			expected: nil,
		},
		{
			name:     "Int slice (should return nil)",
			key:      "int_slice",
			expected: nil,
		},
		{
			name:     "Nil claim",
			key:      "nil_claim",
			expected: nil,
		},
		{
			name:     "Non-existent claim",
			key:      "does_not_exist",
			expected: nil,
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			result := GetClaimStrings(ctx, tt.key)

			if tt.expected == nil {
				if result != nil {
					t.Errorf("Expected nil, got %v", result)
				}
			} else {
				if result == nil {
					t.Errorf("Expected %v, got nil", tt.expected)
					return
				}
				if len(result) != len(tt.expected) {
					t.Errorf("Expected length %d, got %d", len(tt.expected), len(result))
					return
				}
				for i, expected := range tt.expected {
					if result[i] != expected {
						t.Errorf("Expected element %d to be %s, got %s", i, expected, result[i])
					}
				}

				// Verify it's a defensive copy - modify returned slice
				if len(result) > 0 {
					result[0] = testModifiedValue
					freshResult := GetClaimStrings(ctx, tt.key)
					if len(freshResult) > 0 && freshResult[0] != tt.expected[0] {
						t.Error("Defensive copy failed - original slice was modified")
					}
				}
			}
		})
	}

	// Test with nil context
	s.T().Run("Nil context", func(t *testing.T) {
		result := GetClaimStrings(nil, "any_key") //nolint:staticcheck // Testing nil context handling
		if result != nil {
			t.Errorf("Expected nil for nil context, got %v", result)
		}
	})

	// Test with context without auth
	s.T().Run("Context without auth", func(t *testing.T) {
		result := GetClaimStrings(context.Background(), "any_key")
		if result != nil {
			t.Errorf("Expected nil for context without auth, got %v", result)
		}
	})
}

func (s *ContextTestSuite) TestGetAuthenticationContext() {
	s.T().Run("Valid authentication context", func(t *testing.T) {
		authCtx := NewAuthenticationContext("user", "ou", "app", "token", nil)
		ctx := WithAuthenticationContext(context.Background(), authCtx)

		retrievedCtx := getAuthenticationContext(ctx)
		if retrievedCtx == nil {
			t.Error("Expected non-nil authentication context")
		}

		// Verify it's the same context by checking user ID
		if GetUserID(ctx) != "user" {
			t.Error("Retrieved context doesn't match original")
		}
	})

	s.T().Run("Nil context", func(t *testing.T) {
		retrievedCtx := getAuthenticationContext(nil) //nolint:staticcheck // Testing nil context handling
		if retrievedCtx != nil {
			t.Error("Expected nil authentication context for nil context")
		}
	})

	s.T().Run("Context without authentication", func(t *testing.T) {
		ctx := context.Background()
		retrievedCtx := getAuthenticationContext(ctx)
		if retrievedCtx != nil {
			t.Error("Expected nil authentication context for context without auth")
		}
	})

	s.T().Run("Context with wrong type", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), authenticationContextKey, "not_auth_context")
		retrievedCtx := getAuthenticationContext(ctx)
		if retrievedCtx != nil {
			t.Error("Expected nil authentication context for context with wrong type")
		}
	})
}
