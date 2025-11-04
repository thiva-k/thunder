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
	"fmt"
	"strings"

	appmodel "github.com/asgardeo/thunder/internal/application/model"
	"github.com/asgardeo/thunder/internal/system/config"
)

// ParseScopes parses a space-separated scope string into a slice of scope strings.
func ParseScopes(scopeString string) []string {
	trimmed := strings.TrimSpace(scopeString)
	if trimmed == "" {
		return []string{}
	}

	// Split by space and filter out empty strings
	parts := strings.Split(trimmed, " ")
	scopes := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			scopes = append(scopes, part)
		}
	}
	return scopes
}

// JoinScopes joins a slice of scope strings into a space-separated string.
func JoinScopes(scopes []string) string {
	return strings.Join(scopes, " ")
}

// resolveTokenConfig resolves the token configuration from the OAuth app or falls back to global config.
func resolveTokenConfig(oauthApp *appmodel.OAuthAppConfigProcessedDTO, tokenType TokenType) *TokenConfig {
	conf := config.GetThunderRuntime().Config

	tokenConfig := &TokenConfig{
		Issuer:         conf.JWT.Issuer,
		ValidityPeriod: conf.JWT.ValidityPeriod,
	}

	if oauthApp == nil || oauthApp.Token == nil {
		return tokenConfig
	}

	// Override with app-specific configuration if available
	switch tokenType {
	case TokenTypeAccess:
		if oauthApp.Token.AccessToken != nil {
			if oauthApp.Token.AccessToken.Issuer != "" {
				tokenConfig.Issuer = oauthApp.Token.AccessToken.Issuer
			}
			if oauthApp.Token.AccessToken.ValidityPeriod > 0 {
				tokenConfig.ValidityPeriod = oauthApp.Token.AccessToken.ValidityPeriod
			}
		}
	case TokenTypeRefresh:
		if oauthApp.Token.AccessToken != nil {
			if oauthApp.Token.AccessToken.Issuer != "" {
				tokenConfig.Issuer = oauthApp.Token.AccessToken.Issuer
			}
		}
	case TokenTypeID:
		if oauthApp.Token.IDToken != nil {
			if oauthApp.Token.IDToken.ValidityPeriod > 0 {
				tokenConfig.ValidityPeriod = oauthApp.Token.IDToken.ValidityPeriod
			}
		}
	}

	return tokenConfig
}

// extractStringClaim safely extracts a string claim from a claims map.
func extractStringClaim(claims map[string]interface{}, key string) (string, error) {
	value, ok := claims[key]
	if !ok {
		return "", fmt.Errorf("missing claim: %s", key)
	}

	strValue, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("claim %s is not a string", key)
	}

	return strValue, nil
}

// extractInt64Claim safely extracts an int64 claim from a claims map.
func extractInt64Claim(claims map[string]interface{}, key string) (int64, error) {
	value, ok := claims[key]
	if !ok {
		return 0, fmt.Errorf("missing claim: %s", key)
	}

	// JSON numbers are decoded as float64
	switch v := value.(type) {
	case float64:
		return int64(v), nil
	case int64:
		return v, nil
	case int:
		return int64(v), nil
	default:
		return 0, fmt.Errorf("claim %s is not a number", key)
	}
}

// extractScopesFromClaims extracts and parses scopes from a claims map.
func extractScopesFromClaims(claims map[string]interface{}) []string {
	scopeValue, ok := claims["scope"]
	if !ok {
		return []string{}
	}

	scopeString, ok := scopeValue.(string)
	if !ok {
		return []string{}
	}

	return ParseScopes(scopeString)
}

// DetermineAudience determines the audience for a token based on priority.
func DetermineAudience(audience, resource, tokenAud, defaultAudience string) string {
	if audience != "" {
		return audience
	}
	if resource != "" {
		return resource
	}
	if tokenAud != "" {
		return tokenAud
	}
	return defaultAudience
}

// getStandardJWTClaims returns the standard JWT claims that should be excluded from user attributes.
func getStandardJWTClaims() map[string]bool {
	return map[string]bool{
		"sub":       true,
		"iss":       true,
		"aud":       true,
		"exp":       true,
		"nbf":       true,
		"iat":       true,
		"jti":       true,
		"scope":     true,
		"client_id": true,
		"act":       true,
	}
}

// ExtractUserAttributes extracts user attributes from JWT claims by filtering out standard claims.
func ExtractUserAttributes(claims map[string]interface{}) map[string]interface{} {
	standardClaims := getStandardJWTClaims()

	userAttributes := make(map[string]interface{})
	for key, value := range claims {
		if !standardClaims[key] {
			userAttributes[key] = value
		}
	}

	return userAttributes
}

// getValidIssuers collects all valid/trusted issuers for the given OAuth application.
func getValidIssuers(oauthApp *appmodel.OAuthAppConfigProcessedDTO) map[string]bool {
	validIssuers := make(map[string]bool)

	tokenConfig := resolveTokenConfig(oauthApp, TokenTypeAccess)
	validIssuers[tokenConfig.Issuer] = true

	if oauthApp != nil && oauthApp.Token != nil && oauthApp.Token.Issuer != "" {
		validIssuers[oauthApp.Token.Issuer] = true
	}

	// TODO: Add support for external issuers
	return validIssuers
}

// validateIssuer validates that a token issuer is trusted by checking against configured issuers.
func validateIssuer(issuer string, oauthApp *appmodel.OAuthAppConfigProcessedDTO) error {
	validIssuers := getValidIssuers(oauthApp)
	if !validIssuers[issuer] {
		return fmt.Errorf("token issuer '%s' is not supported", issuer)
	}
	return nil
}
