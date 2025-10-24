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

package discovery

// Discovery endpoint paths
const (
	OAuth2AuthorizationServerMetadataEndpoint = "/.well-known/oauth-authorization-server"
	OIDCDiscoveryEndpoint                    = "/.well-known/openid-configuration"
)

// Discovery response field names
const (
	FieldIssuer                            = "issuer"
	FieldAuthorizationEndpoint             = "authorization_endpoint"
	FieldTokenEndpoint                     = "token_endpoint"
	FieldUserInfoEndpoint                  = "userinfo_endpoint"
	FieldJWKSUri                           = "jwks_uri"
	FieldRevocationEndpoint                = "revocation_endpoint"
	FieldIntrospectionEndpoint             = "introspection_endpoint"
	FieldScopesSupported                   = "scopes_supported"
	FieldResponseTypesSupported            = "response_types_supported"
	FieldGrantTypesSupported               = "grant_types_supported"
	FieldTokenEndpointAuthMethodsSupported = "token_endpoint_auth_methods_supported"
	FieldCodeChallengeMethodsSupported     = "code_challenge_methods_supported"

	// OIDC-specific fields
	FieldSubjectTypesSupported            = "subject_types_supported"
	FieldIDTokenSigningAlgValuesSupported = "id_token_signing_alg_values_supported"
	FieldClaimsSupported                  = "claims_supported"
	FieldEndSessionEndpoint               = "end_session_endpoint"
)
