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

package tokenservice

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	appmodel "github.com/asgardeo/thunder/internal/application/model"
	"github.com/asgardeo/thunder/internal/system/config"
)

type UtilsTestSuite struct {
	suite.Suite
}

const (
	testTokenAud        = "https://token-aud.example.com" //nolint:gosec // Test data, not a real credential
	testDefaultAudience = "default-app"
)

func TestUtilsTestSuite(t *testing.T) {
	suite.Run(t, new(UtilsTestSuite))
}

func (suite *UtilsTestSuite) SetupTest() {
	// Initialize Thunder Runtime for tests
	testConfig := &config.Config{
		JWT: config.JWTConfig{
			Issuer:         "https://default.thunder.io",
			ValidityPeriod: 3600,
		},
	}
	_ = config.InitializeThunderRuntime("test", testConfig)
}

func (suite *UtilsTestSuite) TestGetValidIssuers_WithNilOAuthApp() {
	// When oauthApp is nil, should return default issuer from config
	validIssuers := getValidIssuers(nil)

	assert.NotNil(suite.T(), validIssuers)
	assert.Contains(suite.T(), validIssuers, "https://thunder.io")
}

func (suite *UtilsTestSuite) TestGetValidIssuers_WithOnlyDefaultIssuer() {
	oauthApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID: "test-client",
	}

	validIssuers := getValidIssuers(oauthApp)

	assert.NotNil(suite.T(), validIssuers)
	assert.Len(suite.T(), validIssuers, 1)
	assert.Contains(suite.T(), validIssuers, "https://thunder.io")
}

func (suite *UtilsTestSuite) TestGetValidIssuers_WithCustomTokenIssuer() {
	oauthApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID: "test-client",
		Token: &appmodel.OAuthTokenConfig{
			Issuer: "https://custom.thunder.io",
		},
	}

	validIssuers := getValidIssuers(oauthApp)

	assert.NotNil(suite.T(), validIssuers)
	assert.Len(suite.T(), validIssuers, 2)
	assert.Contains(suite.T(), validIssuers, "https://thunder.io")
	assert.Contains(suite.T(), validIssuers, "https://custom.thunder.io")
}

func (suite *UtilsTestSuite) TestGetValidIssuers_WithAccessTokenIssuer() {
	oauthApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID: "test-client",
		Token: &appmodel.OAuthTokenConfig{
			AccessToken: &appmodel.TokenConfig{
				Issuer: "https://access.thunder.io",
			},
		},
	}

	validIssuers := getValidIssuers(oauthApp)

	assert.NotNil(suite.T(), validIssuers)
	// ResolveTokenConfig returns the custom issuer, so only that issuer is in the map
	assert.Len(suite.T(), validIssuers, 1)
	assert.Contains(suite.T(), validIssuers, "https://access.thunder.io")
}

func (suite *UtilsTestSuite) TestGetValidIssuers_WithAllIssuers() {
	oauthApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID: "test-client",
		Token: &appmodel.OAuthTokenConfig{
			Issuer: "https://custom.thunder.io",
			AccessToken: &appmodel.TokenConfig{
				Issuer: "https://access.thunder.io",
			},
		},
	}

	validIssuers := getValidIssuers(oauthApp)

	assert.NotNil(suite.T(), validIssuers)
	// ResolveTokenConfig returns AccessToken issuer, plus Token.Issuer and AccessToken.Issuer are added
	assert.Len(suite.T(), validIssuers, 2)
	assert.Contains(suite.T(), validIssuers, "https://custom.thunder.io")
	assert.Contains(suite.T(), validIssuers, "https://access.thunder.io")
}

func (suite *UtilsTestSuite) TestGetValidIssuers_WithDuplicateIssuers() {
	// Same issuer in multiple places should only appear once
	oauthApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID: "test-client",
		Token: &appmodel.OAuthTokenConfig{
			Issuer: "https://thunder.io",
			AccessToken: &appmodel.TokenConfig{
				Issuer: "https://thunder.io",
			},
		},
	}

	validIssuers := getValidIssuers(oauthApp)

	assert.NotNil(suite.T(), validIssuers)
	// Should still work with map deduplication
	assert.True(suite.T(), validIssuers["https://thunder.io"])
}

func (suite *UtilsTestSuite) TestGetValidIssuers_WithEmptyIssuerStrings() {
	// Empty issuer strings should not be added
	oauthApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID: "test-client",
		Token: &appmodel.OAuthTokenConfig{
			Issuer: "",
			AccessToken: &appmodel.TokenConfig{
				Issuer: "",
			},
		},
	}

	validIssuers := getValidIssuers(oauthApp)

	assert.NotNil(suite.T(), validIssuers)
	// Only default issuer from config should be present
	assert.Contains(suite.T(), validIssuers, "https://thunder.io")
	assert.NotContains(suite.T(), validIssuers, "")
}

// ============================================================================
// validateIssuer Tests
// ============================================================================

func (suite *UtilsTestSuite) TestvalidateIssuer_WithValidDefaultIssuer() {
	oauthApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID: "test-client",
	}

	err := validateIssuer("https://thunder.io", oauthApp)

	assert.NoError(suite.T(), err)
}

func (suite *UtilsTestSuite) TestvalidateIssuer_WithValidCustomIssuer() {
	oauthApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID: "test-client",
		Token: &appmodel.OAuthTokenConfig{
			Issuer: "https://custom.thunder.io",
		},
	}

	err := validateIssuer("https://custom.thunder.io", oauthApp)

	assert.NoError(suite.T(), err)
}

func (suite *UtilsTestSuite) TestvalidateIssuer_WithValidAccessTokenIssuer() {
	oauthApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID: "test-client",
		Token: &appmodel.OAuthTokenConfig{
			AccessToken: &appmodel.TokenConfig{
				Issuer: "https://access.thunder.io",
			},
		},
	}

	err := validateIssuer("https://access.thunder.io", oauthApp)

	assert.NoError(suite.T(), err)
}

func (suite *UtilsTestSuite) TestvalidateIssuer_WithInvalidIssuer() {
	oauthApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID: "test-client",
		Token: &appmodel.OAuthTokenConfig{
			Issuer: "https://custom.thunder.io",
		},
	}

	err := validateIssuer("https://evil.example.com", oauthApp)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "not supported")
	assert.Contains(suite.T(), err.Error(), "https://evil.example.com")
}

func (suite *UtilsTestSuite) TestvalidateIssuer_WithEmptyIssuer() {
	oauthApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID: "test-client",
	}

	err := validateIssuer("", oauthApp)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "not supported")
}

func (suite *UtilsTestSuite) TestvalidateIssuer_WithNilOAuthApp() {
	// Should still validate against default issuer from config
	err := validateIssuer("https://thunder.io", nil)

	assert.NoError(suite.T(), err)
}

func (suite *UtilsTestSuite) TestvalidateIssuer_WithNilOAuthAppInvalidIssuer() {
	err := validateIssuer("https://invalid.com", nil)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "not supported")
}

func (suite *UtilsTestSuite) TestFederationScenario_MultipleThunderIssuers() {
	// Simulates a scenario where an organization has multiple Thunder instances
	oauthApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID: "test-client",
		Token: &appmodel.OAuthTokenConfig{
			Issuer: "https://thunder-prod.company.com",
			AccessToken: &appmodel.TokenConfig{
				Issuer: "https://thunder-staging.company.com",
			},
		},
	}

	validIssuers := getValidIssuers(oauthApp)

	// Should accept tokens from all configured Thunder instances
	assert.Contains(suite.T(), validIssuers, "https://thunder-prod.company.com")
	assert.Contains(suite.T(), validIssuers, "https://thunder-staging.company.com")

	// Validate each configured issuer
	assert.NoError(suite.T(), validateIssuer("https://thunder-prod.company.com", oauthApp))
	assert.NoError(suite.T(), validateIssuer("https://thunder-staging.company.com", oauthApp))

	// Should reject unknown issuers
	assert.Error(suite.T(), validateIssuer("https://unknown.company.com", oauthApp))
}

func (suite *UtilsTestSuite) TestFederationScenario_FutureExternalIssuerSupport() {
	// This test documents the intended behavior for future external issuer support
	// TODO: When external issuer support is added, update GetValidIssuers to include
	// external federated issuers from configuration

	oauthApp := &appmodel.OAuthAppConfigProcessedDTO{
		ClientID: "test-client",
		Token: &appmodel.OAuthTokenConfig{
			Issuer: "https://thunder.company.com",
			// In the future, add field for external issuers:
			// ExternalIssuers: []ExternalIssuerConfig{
			//     {Issuer: "https://external-idp.com", JWKSEndpoint: "..."},
			// }
		},
	}

	validIssuers := getValidIssuers(oauthApp)

	// Currently only Thunder issuers are returned
	assert.Contains(suite.T(), validIssuers, "https://thunder.company.com")

	// In the future, external issuers should also be included
	// assert.Contains(suite.T(), validIssuers, "https://external-idp.com")
}

func (suite *UtilsTestSuite) TestJoinScopes_WithMultipleScopes() {
	scopes := []string{"read", "write", "admin"}
	result := JoinScopes(scopes)

	assert.Equal(suite.T(), "read write admin", result)
}

func (suite *UtilsTestSuite) TestJoinScopes_WithSingleScope() {
	scopes := []string{"read"}
	result := JoinScopes(scopes)

	assert.Equal(suite.T(), "read", result)
}

func (suite *UtilsTestSuite) TestJoinScopes_WithEmptySlice() {
	scopes := []string{}
	result := JoinScopes(scopes)

	assert.Equal(suite.T(), "", result)
}

func (suite *UtilsTestSuite) TestJoinScopes_WithNilSlice() {
	scopes := []string(nil)
	result := JoinScopes(scopes)

	assert.Equal(suite.T(), "", result)
}

// ============================================================================
// DetermineAudience Tests
// ============================================================================

func (suite *UtilsTestSuite) TestDetermineAudience_WithAudience() {
	audience := "https://api.example.com"
	resource := "https://other-api.com"
	tokenAud := testTokenAud
	defaultAudience := testDefaultAudience

	result := DetermineAudience(audience, resource, tokenAud, defaultAudience)

	assert.Equal(suite.T(), audience, result)
}

func (suite *UtilsTestSuite) TestDetermineAudience_WithResource() {
	audience := ""
	resource := "https://api.example.com"
	tokenAud := testTokenAud
	defaultAudience := testDefaultAudience

	result := DetermineAudience(audience, resource, tokenAud, defaultAudience)

	assert.Equal(suite.T(), resource, result)
}

func (suite *UtilsTestSuite) TestDetermineAudience_WithTokenAud() {
	audience := ""
	resource := ""
	tokenAud := testTokenAud
	defaultAudience := testDefaultAudience

	result := DetermineAudience(audience, resource, tokenAud, defaultAudience)

	assert.Equal(suite.T(), tokenAud, result)
}

func (suite *UtilsTestSuite) TestDetermineAudience_WithoutResource() {
	audience := ""
	resource := ""
	tokenAud := ""
	defaultAudience := testDefaultAudience

	result := DetermineAudience(audience, resource, tokenAud, defaultAudience)

	assert.Equal(suite.T(), defaultAudience, result)
}

func (suite *UtilsTestSuite) TestDetermineAudience_EmptyDefault() {
	audience := ""
	resource := ""
	tokenAud := ""
	defaultAudience := ""

	result := DetermineAudience(audience, resource, tokenAud, defaultAudience)

	assert.Equal(suite.T(), "", result)
}

// ============================================================================
// getStandardJWTClaims Tests
// ============================================================================

func (suite *UtilsTestSuite) TestgetStandardJWTClaims_ContainsAllStandardClaims() {
	claims := getStandardJWTClaims()

	assert.True(suite.T(), claims["sub"])
	assert.True(suite.T(), claims["iss"])
	assert.True(suite.T(), claims["aud"])
	assert.True(suite.T(), claims["exp"])
	assert.True(suite.T(), claims["nbf"])
	assert.True(suite.T(), claims["iat"])
	assert.True(suite.T(), claims["jti"])
	assert.True(suite.T(), claims["scope"])
	assert.True(suite.T(), claims["client_id"])
	assert.True(suite.T(), claims["act"])
}

func (suite *UtilsTestSuite) TestgetStandardJWTClaims_ReturnsNewMap() {
	claims1 := getStandardJWTClaims()
	claims2 := getStandardJWTClaims()

	// Should be independent - modifying one shouldn't affect the other
	claims1["test"] = true
	assert.NotContains(suite.T(), claims2, "test")
}

func (suite *UtilsTestSuite) TestExtractUserAttributes_WithStandardClaimsOnly() {
	claims := map[string]interface{}{
		"sub":   "user123",
		"iss":   "https://thunder.io",
		"aud":   "app123",
		"exp":   1234567890,
		"scope": "read write",
	}

	result := ExtractUserAttributes(claims)

	assert.Empty(suite.T(), result)
}

func (suite *UtilsTestSuite) TestExtractUserAttributes_WithCustomClaims() {
	claims := map[string]interface{}{
		"sub":    "user123",
		"iss":    "https://thunder.io",
		"aud":    "app123",
		"exp":    1234567890,
		"scope":  "read write",
		"name":   "John Doe",
		"email":  "john@example.com",
		"groups": []string{"admin", "user"},
	}

	result := ExtractUserAttributes(claims)

	assert.Equal(suite.T(), "John Doe", result["name"])
	assert.Equal(suite.T(), "john@example.com", result["email"])
	assert.Equal(suite.T(), []string{"admin", "user"}, result["groups"])
	assert.NotContains(suite.T(), result, "sub")
	assert.NotContains(suite.T(), result, "iss")
	assert.NotContains(suite.T(), result, "aud")
	assert.NotContains(suite.T(), result, "exp")
	assert.NotContains(suite.T(), result, "scope")
}

func (suite *UtilsTestSuite) TestExtractUserAttributes_WithRefreshTokenSpecificClaims() {
	claims := map[string]interface{}{
		"sub":                          "user123",
		"iss":                          "https://thunder.io",
		"aud":                          "app123",
		"exp":                          1234567890,
		"scope":                        "read write",
		"grant_type":                   "authorization_code",
		"access_token_sub":             "user123",
		"access_token_aud":             "app123",
		"access_token_user_attributes": map[string]interface{}{"name": "John"},
		"name":                         "John Doe",
		"email":                        "john@example.com",
	}

	result := ExtractUserAttributes(claims)

	// Should include refresh token specific claims as they're not standard JWT claims
	assert.Equal(suite.T(), "John Doe", result["name"])
	assert.Equal(suite.T(), "john@example.com", result["email"])
	assert.Equal(suite.T(), "authorization_code", result["grant_type"])
	assert.Equal(suite.T(), "user123", result["access_token_sub"])
	assert.Equal(suite.T(), "app123", result["access_token_aud"])
}

func (suite *UtilsTestSuite) TestExtractUserAttributes_EmptyClaims() {
	claims := map[string]interface{}{}

	result := ExtractUserAttributes(claims)

	assert.Empty(suite.T(), result)
}

func (suite *UtilsTestSuite) TestExtractUserAttributes_NilClaims() {
	claims := map[string]interface{}(nil)

	result := ExtractUserAttributes(claims)

	assert.Empty(suite.T(), result)
}

func (suite *UtilsTestSuite) TestextractInt64Claim_WithIntType() {
	claims := map[string]interface{}{
		"iat": int(1234567890),
	}

	result, err := extractInt64Claim(claims, "iat")

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1234567890), result)
}

func (suite *UtilsTestSuite) TestextractInt64Claim_WithInt64Type() {
	claims := map[string]interface{}{
		"iat": int64(1234567890),
	}

	result, err := extractInt64Claim(claims, "iat")

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1234567890), result)
}

func (suite *UtilsTestSuite) TestextractInt64Claim_WithInvalidType() {
	claims := map[string]interface{}{
		"iat": "not-a-number",
	}

	result, err := extractInt64Claim(claims, "iat")

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), int64(0), result)
	assert.Contains(suite.T(), err.Error(), "not a number")
}

func (suite *UtilsTestSuite) TestParseScopes_WithMultipleSpaces() {
	scopeString := "read  write   admin"
	result := ParseScopes(scopeString)

	assert.Equal(suite.T(), []string{"read", "write", "admin"}, result)
}

func (suite *UtilsTestSuite) TestParseScopes_WithLeadingTrailingSpaces() {
	scopeString := "  read write  "
	result := ParseScopes(scopeString)

	assert.Equal(suite.T(), []string{"read", "write"}, result)
}

func (suite *UtilsTestSuite) TestParseScopes_WithSingleScope() {
	scopeString := "read"
	result := ParseScopes(scopeString)

	assert.Equal(suite.T(), []string{"read"}, result)
}

func (suite *UtilsTestSuite) TestextractScopesFromClaims_WithInvalidScopeType() {
	claims := map[string]interface{}{
		"scope": 12345, // Invalid type (not string)
	}

	result := extractScopesFromClaims(claims)

	assert.Empty(suite.T(), result)
}
