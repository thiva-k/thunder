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

// Package claims provides utilities for building OIDC claims for ID tokens and UserInfo responses.
package claims

import (
	"slices"
	"time"

	appmodel "github.com/asgardeo/thunder/internal/application/model"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
)

// BuildUserClaims generates OIDC claims based on scopes, user attributes, and application configuration.
// Used by both ID token generation and UserInfo endpoint to ensure consistent claim handling.
func BuildUserClaims(scopes []string, userAttributes map[string]interface{},
	oauthApp *appmodel.OAuthAppConfigProcessedDTO, includeAuthTime bool) map[string]interface{} {

	claims := make(map[string]interface{})

	if includeAuthTime {
		claims["auth_time"] = time.Now().Unix()
	}

	var configuredUserAttributes []string
	if oauthApp != nil && oauthApp.Token != nil && oauthApp.Token.IDToken != nil {
		configuredUserAttributes = oauthApp.Token.IDToken.UserAttributes
	}

	if oauthApp != nil && len(configuredUserAttributes) == 0 {
		return claims
	}

	for _, scope := range scopes {
		var scopeClaims []string

		if oauthApp != nil && oauthApp.Token != nil && oauthApp.Token.IDToken != nil &&
			oauthApp.Token.IDToken.ScopeClaims != nil {
			if appClaims, exists := oauthApp.Token.IDToken.ScopeClaims[scope]; exists {
				scopeClaims = appClaims
			}
		}

		if scopeClaims == nil {
			if oidcScope, exists := constants.StandardOIDCScopes[scope]; exists {
				scopeClaims = oidcScope.Claims
			}
		}

		for _, claim := range scopeClaims {
			if oauthApp != nil {
				if slices.Contains(configuredUserAttributes, claim) && userAttributes[claim] != nil {
					claims[claim] = userAttributes[claim]
				}
			} else {
				if userAttributes[claim] != nil {
					claims[claim] = userAttributes[claim]
				}
			}
		}
	}

	return claims
}
