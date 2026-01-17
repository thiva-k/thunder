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

package model

import (
	"crypto/subtle"
	"fmt"
	"net/url"
	"slices"

	oauth2const "github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/internal/system/crypto/hash"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/utils"
)

// OAuthAppConfig represents the structure for OAuth application configuration.
type OAuthAppConfig struct {
	ClientID                string                              `json:"client_id"`
	RedirectURIs            []string                            `json:"redirect_uris"`
	GrantTypes              []oauth2const.GrantType             `json:"grant_types"`
	ResponseTypes           []oauth2const.ResponseType          `json:"response_types"`
	TokenEndpointAuthMethod oauth2const.TokenEndpointAuthMethod `json:"token_endpoint_auth_method"`
	PKCERequired            bool                                `json:"pkce_required"`
	PublicClient            bool                                `json:"public_client"`
	Token                   *OAuthTokenConfig                   `json:"token,omitempty"`
	Scopes                  []string                            `json:"scopes,omitempty"`
}

// OAuthAppConfigComplete represents the complete structure for OAuth application configuration.
//
//nolint:lll
type OAuthAppConfigComplete struct {
	ClientID                string                              `json:"client_id" yaml:"client_id"`
	ClientSecret            string                              `json:"client_secret,omitempty" yaml:"client_secret"`
	RedirectURIs            []string                            `json:"redirect_uris" yaml:"redirect_uris"`
	GrantTypes              []oauth2const.GrantType             `json:"grant_types" yaml:"grant_types"`
	ResponseTypes           []oauth2const.ResponseType          `json:"response_types" yaml:"response_types"`
	TokenEndpointAuthMethod oauth2const.TokenEndpointAuthMethod `json:"token_endpoint_auth_method" yaml:"token_endpoint_auth_method"`
	PKCERequired            bool                                `json:"pkce_required" yaml:"pkce_required"`
	PublicClient            bool                                `json:"public_client" yaml:"public_client"`
	Token                   *OAuthTokenConfig                   `json:"token,omitempty" yaml:"token,omitempty"`
	Scopes                  []string                            `json:"scopes,omitempty" yaml:"scopes,omitempty"`
}

// OAuthAppConfigDTO represents the data transfer object for OAuth application configuration.
type OAuthAppConfigDTO struct {
	AppID                   string                              `json:"app_id,omitempty" jsonschema:"The unique identifier of the OAuth application"`
	ClientID                string                              `json:"client_id,omitempty" jsonschema:"OAuth client ID (auto-generated if not provided)"`
	ClientSecret            string                              `json:"client_secret,omitempty" jsonschema:"OAuth client secret (auto-generated if not provided)"`
	RedirectURIs            []string                            `json:"redirect_uris,omitempty" jsonschema:"Allowed redirect URIs. PUBLIC CLIENTS (SPA/Mobile): Required. Use http://localhost:PORT for local SPAs, custom schemes like myapp://callback for mobile apps. CONFIDENTIAL CLIENTS (Server): Required. Use https:// URLs only. M2M: Not needed (omit or leave empty)."`
	GrantTypes              []oauth2const.GrantType             `json:"grant_types,omitempty" jsonschema:"OAuth grant types. PUBLIC CLIENTS (SPA/Mobile): [authorization_code, refresh_token]. CONFIDENTIAL CLIENTS (Server): [authorization_code, refresh_token]. M2M: [client_credentials]."`
	ResponseTypes           []oauth2const.ResponseType          `json:"response_types,omitempty" jsonschema:"OAuth response types. PUBLIC CLIENTS (SPA/Mobile): [code]. CONFIDENTIAL CLIENTS (Server): [code]. M2M: Not needed (omit or leave empty)."`
	TokenEndpointAuthMethod oauth2const.TokenEndpointAuthMethod `json:"token_endpoint_auth_method,omitempty" jsonschema:"Token endpoint authentication method. PUBLIC CLIENTS (SPA/Mobile): none (cannot store secrets). CONFIDENTIAL CLIENTS (Server/M2M): client_secret_basic (recommended) or client_secret_post."`
	PKCERequired            bool                                `json:"pkce_required,omitempty" jsonschema:"Whether PKCE is required. PUBLIC CLIENTS (SPA/Mobile): true (required for security). CONFIDENTIAL CLIENTS (Server): true (best practice). M2M: false (not applicable for client_credentials flow)."`
	PublicClient            bool                                `json:"public_client,omitempty" jsonschema:"Whether this is a public client (cannot securely store secrets). PUBLIC CLIENTS (SPA/Mobile): true. CONFIDENTIAL CLIENTS (Server/M2M): false."`
	Token                   *OAuthTokenConfig                   `json:"token,omitempty" jsonschema:"Token configuration for access tokens and ID tokens"`
	Scopes                  []string                            `json:"scopes,omitempty" jsonschema:"Allowed OAuth scopes. Common scopes: openid (required for OIDC), profile, email, address, phone. Add custom scopes as needed for your application."`
}

// IsAllowedGrantType checks if the provided grant type is allowed.
func (o *OAuthAppConfigDTO) IsAllowedGrantType(grantType oauth2const.GrantType) bool {
	return isAllowedGrantType(o.GrantTypes, grantType)
}

// IsAllowedResponseType checks if the provided response type is allowed.
func (o *OAuthAppConfigDTO) IsAllowedResponseType(responseType string) bool {
	return isAllowedResponseType(o.ResponseTypes, responseType)
}

// IsAllowedTokenEndpointAuthMethod checks if the provided token endpoint authentication method is allowed.
func (o *OAuthAppConfigDTO) IsAllowedTokenEndpointAuthMethod(method oauth2const.TokenEndpointAuthMethod) bool {
	return o.TokenEndpointAuthMethod == method
}

// ValidateRedirectURI validates the provided redirect URI against the registered redirect URIs.
func (o *OAuthAppConfigDTO) ValidateRedirectURI(redirectURI string) error {
	return validateRedirectURI(o.RedirectURIs, redirectURI)
}

// OAuthAppConfigProcessedDTO represents the processed data transfer object for OAuth application configuration.
type OAuthAppConfigProcessedDTO struct {
	AppID                   string                              `yaml:"app_id,omitempty"`
	ClientID                string                              `yaml:"client_id,omitempty"`
	HashedClientSecret      string                              `yaml:"hashed_client_secret,omitempty"`
	RedirectURIs            []string                            `yaml:"redirect_uris,omitempty"`
	GrantTypes              []oauth2const.GrantType             `yaml:"grant_types,omitempty"`
	ResponseTypes           []oauth2const.ResponseType          `yaml:"response_types,omitempty"`
	TokenEndpointAuthMethod oauth2const.TokenEndpointAuthMethod `yaml:"token_endpoint_auth_method,omitempty"`
	PKCERequired            bool                                `yaml:"pkce_required,omitempty"`
	PublicClient            bool                                `yaml:"public_client,omitempty"`
	Token                   *OAuthTokenConfig                   `yaml:"token,omitempty"`
	Scopes                  []string                            `yaml:"scopes,omitempty"`
}

// IsAllowedGrantType checks if the provided grant type is allowed.
func (o *OAuthAppConfigProcessedDTO) IsAllowedGrantType(grantType oauth2const.GrantType) bool {
	return isAllowedGrantType(o.GrantTypes, grantType)
}

// IsAllowedResponseType checks if the provided response type is allowed.
func (o *OAuthAppConfigProcessedDTO) IsAllowedResponseType(responseType string) bool {
	return isAllowedResponseType(o.ResponseTypes, responseType)
}

// IsAllowedTokenEndpointAuthMethod checks if the provided token endpoint authentication method is allowed.
func (o *OAuthAppConfigProcessedDTO) IsAllowedTokenEndpointAuthMethod(method oauth2const.TokenEndpointAuthMethod) bool {
	return o.TokenEndpointAuthMethod == method
}

// ValidateRedirectURI validates the provided redirect URI against the registered redirect URIs.
func (o *OAuthAppConfigProcessedDTO) ValidateRedirectURI(redirectURI string) error {
	return validateRedirectURI(o.RedirectURIs, redirectURI)
}

// RequiresPKCE checks if PKCE is required for this application.
func (o *OAuthAppConfigProcessedDTO) RequiresPKCE() bool {
	return o.PKCERequired || o.PublicClient
}

// ValidateCredentials validates the provided client ID and client secret against the stored values.
func (o *OAuthAppConfigProcessedDTO) ValidateCredentials(clientID, clientSecret string) bool {
	// Validate client ID
	if clientID != o.ClientID {
		return false
	}

	// Hash the provided client secret and compare with stored hashed secret using constant-time comparison
	hashedClientSecret := hash.GenerateThumbprintFromString(clientSecret)
	return subtle.ConstantTimeCompare([]byte(hashedClientSecret), []byte(o.HashedClientSecret)) == 1
}

// isAllowedGrantType checks if the provided grant type is in the allowed list.
func isAllowedGrantType(grantTypes []oauth2const.GrantType, grantType oauth2const.GrantType) bool {
	if grantType == "" {
		return false
	}
	return slices.Contains(grantTypes, grantType)
}

// isAllowedResponseType checks if the provided response type is in the allowed list.
func isAllowedResponseType(responseTypes []oauth2const.ResponseType, responseType string) bool {
	if responseType == "" {
		return false
	}
	return slices.Contains(responseTypes, oauth2const.ResponseType(responseType))
}

// validateRedirectURI checks if the provided redirect URI is valid against the registered redirect URIs.
func validateRedirectURI(redirectURIs []string, redirectURI string) error {
	logger := log.GetLogger()

	// Check if the redirect URI is empty.
	if redirectURI == "" {
		// Check if multiple redirect URIs are registered.
		if len(redirectURIs) != 1 {
			return fmt.Errorf("redirect URI is required in the authorization request")
		}
		// Check if only a part of the redirect uri is registered.
		parsed, err := url.Parse(redirectURIs[0])
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			return fmt.Errorf("registered redirect URI is not fully qualified")
		}

		// Valid scenario.
		return nil
	}

	// Check if the redirect URI is registered.
	if !slices.Contains(redirectURIs, redirectURI) {
		return fmt.Errorf("your application's redirect URL does not match with the registered redirect URLs")
	}

	// Parse the redirect URI.
	parsedRedirectURI, err := utils.ParseURL(redirectURI)
	if err != nil {
		logger.Error("Failed to parse redirect URI", log.Error(err))
		return fmt.Errorf("invalid redirect URI: %s", err.Error())
	}
	// Check if it is a fragment URI.
	if parsedRedirectURI.Fragment != "" {
		return fmt.Errorf("redirect URI must not contain a fragment component")
	}

	return nil
}
