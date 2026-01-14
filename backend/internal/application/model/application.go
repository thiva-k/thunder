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

// Package model defines the data structures for the application module.
package model

import (
	"github.com/asgardeo/thunder/internal/cert"
)

// TokenConfig represents the token configuration structure for application-level (root) token configs.
type TokenConfig struct {
	Issuer         string   `json:"issuer,omitempty" yaml:"issuer,omitempty" jsonschema:"Token issuer URL"`
	ValidityPeriod int64    `json:"validity_period,omitempty" yaml:"validity_period,omitempty" jsonschema:"Token validity period in seconds"`
	UserAttributes []string `json:"user_attributes,omitempty" yaml:"user_attributes,omitempty" jsonschema:"User attributes to include in token"`
}

// AccessTokenConfig represents the access token configuration structure.
type AccessTokenConfig struct {
	ValidityPeriod int64    `json:"validity_period" yaml:"validity_period,omitempty" jsonschema:"Access token validity period in seconds"`
	UserAttributes []string `json:"user_attributes" yaml:"user_attributes,omitempty" jsonschema:"User attributes to include in access token"`
}

// IDTokenConfig represents the ID token configuration structure.
type IDTokenConfig struct {
	ValidityPeriod int64               `json:"validity_period" yaml:"validity_period,omitempty" jsonschema:"ID token validity period in seconds"`
	UserAttributes []string            `json:"user_attributes" yaml:"user_attributes,omitempty" jsonschema:"User attributes to include in ID token"`
	ScopeClaims    map[string][]string `json:"scope_claims,omitempty" yaml:"scope_claims,omitempty" jsonschema:"Map of scopes to claims"`
}

// OAuthTokenConfig represents the OAuth token configuration structure with access_token and id_token wrappers.
// The Issuer field at this level is used by both access and ID tokens.
type OAuthTokenConfig struct {
	Issuer      string             `json:"issuer,omitempty" yaml:"issuer,omitempty" jsonschema:"Token issuer URL"`
	AccessToken *AccessTokenConfig `json:"access_token,omitempty" yaml:"access_token,omitempty" jsonschema:"Access token configuration"`
	IDToken     *IDTokenConfig     `json:"id_token,omitempty" yaml:"id_token,omitempty" jsonschema:"ID token configuration"`
}

// ApplicationDTO represents the data transfer object for application service operations.
type ApplicationDTO struct {
	ID                        string `json:"id,omitempty" jsonschema:"Application ID (required for update/delete, auto-generated for create)"`
	Name                      string `json:"name" jsonschema:"Application name"`
	Description               string `json:"description,omitempty" jsonschema:"Application description"`
	AuthFlowID                string `json:"auth_flow_id,omitempty" jsonschema:"Authentication flow ID to assign to this application"`
	RegistrationFlowID        string `json:"registration_flow_id,omitempty" jsonschema:"Registration flow ID to assign to this application"`
	IsRegistrationFlowEnabled bool   `json:"is_registration_flow_enabled,omitempty" jsonschema:"Enable registration flow for this application"`
	BrandingID                string `json:"branding_id,omitempty" jsonschema:"Branding configuration ID"`
	Template                  string `json:"template,omitempty" jsonschema:"Application template"`

	URL       string   `json:"url,omitempty" jsonschema:"Application URL"`
	LogoURL   string   `json:"logo_url,omitempty" jsonschema:"Logo URL"`
	TosURI    string   `json:"tos_uri,omitempty" jsonschema:"Terms of service URI"`
	PolicyURI string   `json:"policy_uri,omitempty" jsonschema:"Privacy policy URI"`
	Contacts  []string `json:"contacts,omitempty" jsonschema:"Contact email addresses"`

	Token             *TokenConfig            `json:"token,omitempty" jsonschema:"Token configuration"`
	Certificate       *ApplicationCertificate `json:"certificate,omitempty" jsonschema:"Application certificate"`
	InboundAuthConfig []InboundAuthConfigDTO  `json:"inbound_auth_config,omitempty" jsonschema:"OAuth/OIDC inbound authentication configuration"`
	AllowedUserTypes  []string                `json:"allowed_user_types,omitempty" jsonschema:"Allowed user types for this application"`
}

// BasicApplicationDTO represents a simplified data transfer object for application service operations.
type BasicApplicationDTO struct {
	ID                        string
	Name                      string
	Description               string
	AuthFlowID                string
	RegistrationFlowID        string
	IsRegistrationFlowEnabled bool
	BrandingID                string
	Template                  string
	ClientID                  string
	LogoURL                   string
}

// Application represents the structure for application which returns in GetApplicationById.
type Application struct {
	ID                        string `yaml:"id,omitempty" json:"id,omitempty"`
	Name                      string `yaml:"name,omitempty" json:"name"`
	Description               string `yaml:"description,omitempty" json:"description,omitempty"`
	AuthFlowID                string `yaml:"auth_flow_id,omitempty" json:"auth_flow_id,omitempty"`
	RegistrationFlowID        string `yaml:"registration_flow_id,omitempty" json:"registration_flow_id,omitempty"`
	IsRegistrationFlowEnabled bool   `yaml:"is_registration_flow_enabled,omitempty" json:"is_registration_flow_enabled"`
	BrandingID                string `yaml:"branding_id,omitempty" json:"branding_id,omitempty"`
	Template                  string `yaml:"template,omitempty" json:"template,omitempty"`

	URL       string `yaml:"url,omitempty" json:"url,omitempty"`
	LogoURL   string `yaml:"logo_url,omitempty" json:"logo_url,omitempty"`
	TosURI    string `yaml:"tos_uri,omitempty" json:"tos_uri,omitempty"`
	PolicyURI string `yaml:"policy_uri,omitempty" json:"policy_uri,omitempty"`
	Contacts  []string

	Token             *TokenConfig                `yaml:"token,omitempty" json:"token,omitempty"`
	Certificate       *ApplicationCertificate     `yaml:"certificate,omitempty" json:"certificate,omitempty"`
	InboundAuthConfig []InboundAuthConfigComplete `yaml:"inbound_auth_config,omitempty" json:"inbound_auth_config,omitempty"`
	AllowedUserTypes  []string                    `yaml:"allowed_user_types,omitempty" json:"allowed_user_types,omitempty"`
}

// ApplicationProcessedDTO represents the processed data transfer object for application service operations.
type ApplicationProcessedDTO struct {
	ID                        string `yaml:"id,omitempty"`
	Name                      string `yaml:"name,omitempty"`
	Description               string `yaml:"description,omitempty"`
	AuthFlowID                string `yaml:"auth_flow_id,omitempty"`
	RegistrationFlowID        string `yaml:"registration_flow_id,omitempty"`
	IsRegistrationFlowEnabled bool   `yaml:"is_registration_flow_enabled,omitempty"`
	BrandingID                string `yaml:"branding_id,omitempty"`
	Template                  string `yaml:"template,omitempty"`

	URL       string `yaml:"url,omitempty"`
	LogoURL   string `yaml:"logo_url,omitempty"`
	TosURI    string `yaml:"tos_uri,omitempty"`
	PolicyURI string `yaml:"policy_uri,omitempty"`
	Contacts  []string

	Token             *TokenConfig                    `yaml:"token,omitempty"`
	Certificate       *ApplicationCertificate         `yaml:"certificate,omitempty"`
	InboundAuthConfig []InboundAuthConfigProcessedDTO `yaml:"inbound_auth_config,omitempty"`
	AllowedUserTypes  []string                        `yaml:"allowed_user_types,omitempty"`
}

// InboundAuthConfigDTO represents the data transfer object for inbound authentication configuration.
// TODO: Need to refactor when supporting other/multiple inbound auth types.
// InboundAuthConfigDTO represents the data transfer object for inbound authentication configuration.
// TODO: Need to refactor when supporting other/multiple inbound auth types.
type InboundAuthConfigDTO struct {
	Type           InboundAuthType    `json:"type" jsonschema:"Inbound authentication type (e.g., oauth2)"`
	OAuthAppConfig *OAuthAppConfigDTO `json:"config,omitempty" jsonschema:"OAuth/OIDC application configuration"`
}

// InboundAuthConfigProcessedDTO represents the processed data transfer object for inbound authentication
// configuration.
type InboundAuthConfigProcessedDTO struct {
	Type           InboundAuthType             `json:"type" yaml:"type,omitempty"`
	OAuthAppConfig *OAuthAppConfigProcessedDTO `json:"oauth_app_config,omitempty" yaml:"config,omitempty"`
}

// ApplicationCertificate represents the certificate structure in the application request response.
type ApplicationCertificate struct {
	Type  cert.CertificateType `json:"type" yaml:"type,omitempty" jsonschema:"Certificate type (PEM or JWKS)"`
	Value string               `json:"value" yaml:"value,omitempty" jsonschema:"Certificate value or JWKS URL"`
}

// ApplicationRequest represents the request structure for creating or updating an application.
//
//nolint:lll
type ApplicationRequest struct {
	Name                      string                      `json:"name" yaml:"name" jsonschema:"required" jsonschema_description:"The name of the application"`
	Description               string                      `json:"description,omitempty" yaml:"description,omitempty" jsonschema_description:"A description of the application"`
	AuthFlowID                string                      `json:"auth_flow_id,omitempty" yaml:"auth_flow_id,omitempty" jsonschema_description:"Authentication flow ID to use"`
	RegistrationFlowID        string                      `json:"registration_flow_id,omitempty" yaml:"registration_flow_id,omitempty" jsonschema_description:"Registration flow ID to use"`
	IsRegistrationFlowEnabled bool                        `json:"is_registration_flow_enabled,omitempty" yaml:"is_registration_flow_enabled,omitempty" jsonschema_description:"Whether registration flow is enabled"`
	BrandingID                string                      `json:"branding_id,omitempty" yaml:"branding_id,omitempty" jsonschema_description:"Branding configuration ID"`
	Template                  string                      `json:"template,omitempty" yaml:"template,omitempty" jsonschema_description:"Application template"`
	URL                       string                      `json:"url,omitempty" yaml:"url,omitempty" jsonschema_description:"Application URL"`
	LogoURL                   string                      `json:"logo_url,omitempty" yaml:"logo_url,omitempty" jsonschema_description:"Application logo URL"`
	Token                     *TokenConfig                `json:"token,omitempty" yaml:"token,omitempty" jsonschema_description:"Token configuration"`
	Certificate               *ApplicationCertificate     `json:"certificate,omitempty" yaml:"certificate,omitempty" jsonschema_description:"Certificate configuration"`
	TosURI                    string                      `json:"tos_uri,omitempty" yaml:"tos_uri,omitempty" jsonschema_description:"Terms of service URI"`
	PolicyURI                 string                      `json:"policy_uri,omitempty" yaml:"policy_uri,omitempty" jsonschema_description:"Privacy policy URI"`
	Contacts                  []string                    `json:"contacts,omitempty" yaml:"contacts,omitempty" jsonschema_description:"Contact email addresses"`
	InboundAuthConfig         []InboundAuthConfigComplete `json:"inbound_auth_config,omitempty" yaml:"inbound_auth_config,omitempty" jsonschema_description:"Inbound authentication configuration (OAuth settings)"`
	AllowedUserTypes          []string                    `json:"allowed_user_types,omitempty" yaml:"allowed_user_types,omitempty" jsonschema_description:"Allowed user types for this application"`
}

// ApplicationRequestWithID represents the request structure for importing an application using file based runtime.
//
//nolint:lll
type ApplicationRequestWithID struct {
	ID                        string                      `json:"id" yaml:"id"`
	Name                      string                      `json:"name" yaml:"name"`
	Description               string                      `json:"description" yaml:"description"`
	AuthFlowID                string                      `json:"auth_flow_id,omitempty" yaml:"auth_flow_id,omitempty"`
	RegistrationFlowID        string                      `json:"registration_flow_id,omitempty" yaml:"registration_flow_id,omitempty"`
	IsRegistrationFlowEnabled bool                        `json:"is_registration_flow_enabled" yaml:"is_registration_flow_enabled"`
	BrandingID                string                      `json:"branding_id,omitempty" yaml:"branding_id,omitempty"`
	Template                  string                      `json:"template,omitempty" yaml:"template,omitempty"`
	URL                       string                      `json:"url,omitempty" yaml:"url,omitempty"`
	LogoURL                   string                      `json:"logo_url,omitempty" yaml:"logo_url,omitempty"`
	Token                     *TokenConfig                `json:"token,omitempty" yaml:"token,omitempty"`
	Certificate               *ApplicationCertificate     `json:"certificate,omitempty" yaml:"certificate,omitempty"`
	TosURI                    string                      `json:"tos_uri,omitempty" yaml:"tos_uri,omitempty"`
	PolicyURI                 string                      `json:"policy_uri,omitempty" yaml:"policy_uri,omitempty"`
	Contacts                  []string                    `json:"contacts,omitempty" yaml:"contacts,omitempty"`
	InboundAuthConfig         []InboundAuthConfigComplete `json:"inbound_auth_config,omitempty" yaml:"inbound_auth_config,omitempty"`
	AllowedUserTypes          []string                    `json:"allowed_user_types,omitempty" yaml:"allowed_user_types,omitempty"`
}

// ApplicationCompleteResponse represents the complete response structure for an application.
type ApplicationCompleteResponse struct {
	ID                        string                      `json:"id,omitempty"`
	Name                      string                      `json:"name"`
	Description               string                      `json:"description,omitempty"`
	ClientID                  string                      `json:"client_id,omitempty"`
	AuthFlowID                string                      `json:"auth_flow_id,omitempty"`
	RegistrationFlowID        string                      `json:"registration_flow_id,omitempty"`
	IsRegistrationFlowEnabled bool                        `json:"is_registration_flow_enabled"`
	BrandingID                string                      `json:"branding_id,omitempty"`
	Template                  string                      `json:"template,omitempty"`
	URL                       string                      `json:"url,omitempty"`
	LogoURL                   string                      `json:"logo_url,omitempty"`
	Token                     *TokenConfig                `json:"token,omitempty"`
	Certificate               *ApplicationCertificate     `json:"certificate,omitempty"`
	TosURI                    string                      `json:"tos_uri,omitempty"`
	PolicyURI                 string                      `json:"policy_uri,omitempty"`
	Contacts                  []string                    `json:"contacts,omitempty"`
	InboundAuthConfig         []InboundAuthConfigComplete `json:"inbound_auth_config,omitempty"`
	AllowedUserTypes          []string                    `json:"allowed_user_types,omitempty"`
}

// ApplicationGetResponse represents the response structure for getting an application.
type ApplicationGetResponse struct {
	ID                        string                  `json:"id,omitempty"`
	Name                      string                  `json:"name"`
	Description               string                  `json:"description,omitempty"`
	ClientID                  string                  `json:"client_id,omitempty"`
	AuthFlowID                string                  `json:"auth_flow_id,omitempty"`
	RegistrationFlowID        string                  `json:"registration_flow_id,omitempty"`
	IsRegistrationFlowEnabled bool                    `json:"is_registration_flow_enabled"`
	BrandingID                string                  `json:"branding_id,omitempty"`
	Template                  string                  `json:"template,omitempty"`
	URL                       string                  `json:"url,omitempty"`
	LogoURL                   string                  `json:"logo_url,omitempty"`
	Token                     *TokenConfig            `json:"token,omitempty"`
	Certificate               *ApplicationCertificate `json:"certificate,omitempty"`
	TosURI                    string                  `json:"tos_uri,omitempty"`
	PolicyURI                 string                  `json:"policy_uri,omitempty"`
	Contacts                  []string                `json:"contacts,omitempty"`
	InboundAuthConfig         []InboundAuthConfig     `json:"inbound_auth_config,omitempty"`
	AllowedUserTypes          []string                `json:"allowed_user_types,omitempty"`
}

// BasicApplicationResponse represents a simplified response structure for an application.
type BasicApplicationResponse struct {
	ID                        string `json:"id,omitempty"`
	Name                      string `json:"name"`
	Description               string `json:"description,omitempty"`
	ClientID                  string `json:"client_id,omitempty"`
	LogoURL                   string `json:"logo_url,omitempty"`
	AuthFlowID                string `json:"auth_flow_id,omitempty"`
	RegistrationFlowID        string `json:"registration_flow_id,omitempty"`
	IsRegistrationFlowEnabled bool   `json:"is_registration_flow_enabled"`
	BrandingID                string `json:"branding_id,omitempty"`
	Template                  string `json:"template,omitempty"`
}

// ApplicationListResponse represents the response structure for listing applications.
type ApplicationListResponse struct {
	TotalResults int                        `json:"totalResults"`
	Count        int                        `json:"count"`
	Applications []BasicApplicationResponse `json:"applications"`
}

// InboundAuthConfig represents the structure for inbound authentication configuration.
type InboundAuthConfig struct {
	Type           InboundAuthType `json:"type"`
	OAuthAppConfig *OAuthAppConfig `json:"config,omitempty"`
}

// InboundAuthConfigComplete represents the complete structure for inbound authentication configuration.
type InboundAuthConfigComplete struct {
	Type           InboundAuthType         `json:"type" yaml:"type"`
	OAuthAppConfig *OAuthAppConfigComplete `json:"config,omitempty" yaml:"config,omitempty"`
}
