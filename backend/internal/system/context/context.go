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

// Package context provides utilities for managing trace IDs (correlation IDs)
// and authentication context across the application.
package context

import (
	"context"
	"crypto/rand"
	"fmt"
)

type contextKey string

const (
	// TraceIDKey is the context key for storing the trace ID (correlation ID).
	TraceIDKey contextKey = "trace_id"

	// authenticationContextKey is the context key for storing authentication context (unexported for security).
	authenticationContextKey contextKey = "authentication_context"
)

// ============================================================================
// Trace ID Functions
// ============================================================================

// generateUUID generates a UUID v4 string.
// This is a copy of utils.GenerateUUID to avoid import cycles.
func generateUUID() string {
	var uuid [16]byte
	_, err := rand.Read(uuid[:])
	if err != nil {
		panic(fmt.Errorf("failed to generate random bytes: %w", err))
	}

	uuid[6] = (uuid[6] & 0x0f) | 0x40 // Version 4
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // Variant is 10

	return fmt.Sprintf("%x-%x-%x-%x-%x",
		uuid[0:4],
		uuid[4:6],
		uuid[6:8],
		uuid[8:10],
		uuid[10:],
	)
}

// GetTraceID retrieves the trace ID (correlation ID) from the context.
// If no trace ID exists, it generates a new UUID.
// This trace ID can be used to correlate logs, events, and operations across a request flow.
func GetTraceID(ctx context.Context) string {
	if ctx == nil {
		return generateUUID()
	}

	if traceID, ok := ctx.Value(TraceIDKey).(string); ok && traceID != "" {
		return traceID
	}

	return generateUUID()
}

// WithTraceID adds a trace ID (correlation ID) to the context.
// Use this to propagate trace IDs through your application.
func WithTraceID(ctx context.Context, traceID string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, TraceIDKey, traceID)
}

// EnsureTraceID ensures a trace ID (correlation ID) exists in the context,
// generating one if needed. This is useful at entry points where you want to
// guarantee a trace ID is present for downstream operations.
func EnsureTraceID(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	if traceID, ok := ctx.Value(TraceIDKey).(string); !ok || traceID == "" {
		ctx = WithTraceID(ctx, generateUUID())
	}

	return ctx
}

// ============================================================================
// Authentication Context
// ============================================================================

// AuthenticationContext holds immutable authenticated user information.
type AuthenticationContext struct {
	userID string
	ouID   string
	appID  string
	token  string
	claims map[string]interface{}
}

// NewAuthenticationContext creates a new immutable AuthenticationContext.
// This should only be called by the security package.
func NewAuthenticationContext(userID, ouID, appID, token string, claims map[string]interface{}) *AuthenticationContext {
	return &AuthenticationContext{
		userID: userID,
		ouID:   ouID,
		appID:  appID,
		token:  token,
		claims: claims,
	}
}

// WithAuthenticationContext adds authentication context to the request context.
// This should only be called by the security package.
func WithAuthenticationContext(ctx context.Context, authCtx *AuthenticationContext) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, authenticationContextKey, authCtx)
}

// GetUserID retrieves the authenticated user ID from the context.
// Returns empty string if no authentication context is present.
func GetUserID(ctx context.Context) string {
	authCtx := getAuthenticationContext(ctx)
	if authCtx != nil {
		return authCtx.userID
	}
	return ""
}

// GetOUID retrieves the organization unit ID from the context.
// Returns empty string if no authentication context is present.
func GetOUID(ctx context.Context) string {
	authCtx := getAuthenticationContext(ctx)
	if authCtx != nil {
		return authCtx.ouID
	}
	return ""
}

// GetAppID retrieves the application ID from the context.
// Returns empty string if no authentication context is present.
func GetAppID(ctx context.Context) string {
	authCtx := getAuthenticationContext(ctx)
	if authCtx != nil {
		return authCtx.appID
	}
	return ""
}

// GetClaim retrieves a specific claim from the authentication token.
// Returns defensive copies for mutable types (slices, maps) to prevent modification.
// Returns nil if no authentication context is present or claim doesn't exist.
func GetClaim(ctx context.Context, key string) interface{} {
	authCtx := getAuthenticationContext(ctx)
	if authCtx == nil {
		return nil
	}

	value, exists := authCtx.claims[key]
	if !exists {
		return nil
	}

	// Return defensive copies for mutable types to prevent modification
	switch v := value.(type) {
	case []string:
		result := make([]string, len(v))
		copy(result, v)
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		copy(result, v)
		return result
	case map[string]interface{}:
		result := make(map[string]interface{}, len(v))
		for k, val := range v {
			result[k] = val
		}
		return result
	default:
		// Immutable types (string, int, bool, etc.) are safe to return directly
		return value
	}
}

// GetClaimString retrieves a string claim from the authentication token.
// Returns empty string if claim doesn't exist or is not a string.
func GetClaimString(ctx context.Context, key string) string {
	value := GetClaim(ctx, key)
	if str, ok := value.(string); ok {
		return str
	}
	return ""
}

// GetClaimStrings retrieves a string slice claim from the authentication token.
// Returns nil if claim doesn't exist or is not a string slice.
// The returned slice is a defensive copy and can be safely modified.
func GetClaimStrings(ctx context.Context, key string) []string {
	value := GetClaim(ctx, key)
	if strs, ok := value.([]string); ok {
		return strs // Already a defensive copy from GetClaim
	}
	return nil
}

// getAuthenticationContext is an internal helper to retrieve the authentication context.
// This function is unexported to prevent downstream services from accessing the raw context object.
func getAuthenticationContext(ctx context.Context) *AuthenticationContext {
	if ctx == nil {
		return nil
	}
	authCtx, _ := ctx.Value(authenticationContextKey).(*AuthenticationContext)
	return authCtx
}
