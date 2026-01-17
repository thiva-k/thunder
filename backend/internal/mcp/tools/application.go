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

package tools

import (
	"context"
	"fmt"

	"github.com/asgardeo/thunder/internal/application"
	"github.com/asgardeo/thunder/internal/application/model"
	oauth2const "github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ApplicationTools provides MCP tools for managing Thunder applications.
type ApplicationTools struct {
	appService application.ApplicationServiceInterface
}

// NewApplicationTools creates a new ApplicationTools instance.
func NewApplicationTools(appService application.ApplicationServiceInterface) *ApplicationTools {
	return &ApplicationTools{
		appService: appService,
	}
}

// ApplicationListOutput represents the output for list_applications tool.
type ApplicationListOutput struct {
	TotalCount   int                              `json:"total_count"`
	Applications []model.BasicApplicationResponse `json:"applications"`
}

// ClientIDInput represents input for client ID-based lookups.
type ClientIDInput struct {
	ClientID string `json:"client_id" jsonschema:"OAuth client ID to search for"`
}

// RegisterTools registers all application tools with the MCP server.
func (t *ApplicationTools) RegisterTools(server *mcp.Server) {
	// Generate schema with enum support for Application
	appSchema := GenerateSchema[model.ApplicationDTO](
		WithEnum("grant_types", oauth2const.GetSupportedGrantTypes()),
		WithEnum("response_types", oauth2const.GetSupportedResponseTypes()),
		WithEnum("token_endpoint_auth_method", oauth2const.GetSupportedTokenEndpointAuthMethods()),
	)
	// Generate schema for update with 'id' as required
	updateAppSchema := GenerateSchema[model.ApplicationDTO](
		WithEnum("grant_types", oauth2const.GetSupportedGrantTypes()),
		WithEnum("response_types", oauth2const.GetSupportedResponseTypes()),
		WithEnum("token_endpoint_auth_method", oauth2const.GetSupportedTokenEndpointAuthMethods()),
		WithRequired("id"),
	)

	mcp.AddTool(server, &mcp.Tool{
		Name: "list_applications",
		Description: `List all registered applications.

Related: Use returned 'id' with get_application_by_id or update_application.`,
		Annotations: &mcp.ToolAnnotations{
			Title:        "List Applications",
			ReadOnlyHint: true,
		},
	}, t.ListApplications)

	mcp.AddTool(server, &mcp.Tool{
		Name: "get_application_by_id",
		Description: `Retrieve full details of an application by ID including OAuth settings, branding, and flow associations.

Related: Use before update_application to review current configuration.`,
		Annotations: &mcp.ToolAnnotations{
			Title:        "Get Application by ID",
			ReadOnlyHint: true,
		},
	}, t.GetApplicationByID)

	mcp.AddTool(server, &mcp.Tool{
		Name: "get_application_by_client_id",
		Description: `Retrieve application by OAuth client_id. Useful when troubleshooting OAuth errors or debugging authentication issues.

Related: Alternative to get_application_by_id when you have the client_id from OAuth logs.`,
		Annotations: &mcp.ToolAnnotations{
			Title:        "Get Application by Client ID",
			ReadOnlyHint: true,
		},
	}, t.GetApplicationByClientID)

	mcp.AddTool(server, &mcp.Tool{
		Name: "create_application",
		Description: `Create a new application optionally with OAuth configuration.

Quick Start: Use get_application_templates to get pre-configured templates for common app types (SPA, Mobile, Server, M2M).

Prerequisites: Create flows first using create_flow if custom authentication/registration flows are needed.

Behavior: If auth_flow_id is omitted, the default authentication flow is used.

Outputs: Created application with generated 'client_id' and 'client_secret'.

Related: 
- Use get_application_templates for pre-configured OAuth settings
- Use list_flows to find available flow IDs`,
		InputSchema: appSchema,
		Annotations: &mcp.ToolAnnotations{
			Title:          "Create Application",
			IdempotentHint: true,
		},
	}, t.CreateApplication)

	mcp.AddTool(server, &mcp.Tool{
		Name: "update_application",
		Description: `Update an existing application configuration with PATCH semantics.

PATCH semantics: Only provided fields are updated. Omitted/empty fields are preserved.

Updatable fields:
- Basic info: name, description, url, logo_url, tos_uri, policy_uri, contacts
- Flow associations: auth_flow_id, registration_flow_id, is_registration_flow_enabled
- Branding: branding_id
- OAuth config: redirect_uris, grant_types, response_types, token_endpoint_auth_method, pkce_required, public_client, scopes

Examples:
- Update name only: {"id": "app-123", "name": "New Name"}
- Update redirect URIs only: {"id": "app-123", "inbound_auth_config": [{"type": "oauth2", "oauth_app_config": {"redirect_uris": ["http://localhost:3000"]}}]}
- Update multiple fields: {"id": "app-123", "name": "New Name", "description": "New desc"}

Tip: Use get_application_templates to see correct OAuth configuration patterns for different app types.

Prerequisites: Use get_application_by_id first to see current configuration if needed.`,
		InputSchema: updateAppSchema,
		Annotations: &mcp.ToolAnnotations{
			Title:          "Update Application",
			IdempotentHint: true,
		},
	}, t.UpdateApplication)

	mcp.AddTool(server, &mcp.Tool{
		Name: "get_application_templates",
		Description: `Get minimal OAuth configuration templates for common application types.

Templates contain ONLY the required fields to successfully create each app type. Optional fields with service-layer defaults are omitted for simplicity.

Returns minimal configurations for:
- spa: Single Page Application (React, Vue, Angular) - Public client with PKCE
- mobile: Mobile apps (iOS, Android, React Native, Flutter) - Public client with PKCE
- server: Server-side web apps (Next.js, Nuxt) - Confidential client with PKCE
- m2m: Machine-to-Machine / Backend services - Client credentials flow

Service-layer defaults (auto-applied if omitted):
- response_types: ["code"] for authorization_code grant
- token_endpoint_auth_method: "client_secret_basic" for confidential clients
- public_client: false for confidential clients
- scopes: Can be added later as needed

Usage:
1. Call this tool to get minimal templates
2. Copy the desired template
3. Replace placeholder values (e.g., <APP_NAME>, <REDIRECT_URI>)
4. Optionally add more fields (description, scopes, etc.)
5. Use with create_application`,
		Annotations: &mcp.ToolAnnotations{
			Title:        "Get Application Templates",
			ReadOnlyHint: true,
		},
	}, t.GetApplicationTemplates)
}

// ptr returns a pointer to the given value.
func ptr[T any](v T) *T {
	return &v
}

// ListApplications handles the list_applications tool call.
func (t *ApplicationTools) ListApplications(
	ctx context.Context,
	req *mcp.CallToolRequest,
	_ any,
) (*mcp.CallToolResult, ApplicationListOutput, error) {
	listResponse, svcErr := t.appService.GetApplicationList()
	if svcErr != nil {
		return nil, ApplicationListOutput{}, fmt.Errorf("failed to list applications: %s", svcErr.ErrorDescription)
	}

	return nil, ApplicationListOutput{
		TotalCount:   listResponse.TotalResults,
		Applications: listResponse.Applications,
	}, nil
}

// GetApplicationByID handles the get_application_by_id tool call.
func (t *ApplicationTools) GetApplicationByID(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input IDInput,
) (*mcp.CallToolResult, *model.Application, error) {

	app, svcErr := t.appService.GetApplication(input.ID)
	if svcErr != nil {
		return nil, nil, fmt.Errorf("failed to get application: %s", svcErr.ErrorDescription)
	}

	return nil, app, nil
}

// GetApplicationByClientID handles the get_application_by_client_id tool call.
func (t *ApplicationTools) GetApplicationByClientID(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input ClientIDInput,
) (*mcp.CallToolResult, *model.Application, error) {

	// Get OAuth application to find app ID
	oauthApp, svcErr := t.appService.GetOAuthApplication(input.ClientID)
	if svcErr != nil {
		return nil, nil, fmt.Errorf("failed to get OAuth application: %s", svcErr.ErrorDescription)
	}

	// Get full application details
	app, svcErr := t.appService.GetApplication(oauthApp.AppID)
	if svcErr != nil {
		return nil, nil, fmt.Errorf("failed to get application: %s", svcErr.ErrorDescription)
	}

	return nil, app, nil
}

// CreateApplication handles the create_application tool call.
// Uses ApplicationDTO directly - ID field is optional for create (auto-generated).
func (t *ApplicationTools) CreateApplication(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input model.ApplicationDTO,
) (*mcp.CallToolResult, *model.ApplicationDTO, error) {

	createdApp, svcErr := t.appService.CreateApplication(&input)
	if svcErr != nil {
		return nil, nil, fmt.Errorf("failed to create application: %s", svcErr.ErrorDescription)
	}

	return nil, createdApp, nil
}

// UpdateApplication handles the update_application tool call with PATCH semantics.
// Only provided (non-empty) fields are updated; omitted fields are preserved.
func (t *ApplicationTools) UpdateApplication(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input model.ApplicationDTO,
) (*mcp.CallToolResult, *model.ApplicationDTO, error) {

	// Get current application
	currentApp, svcErr := t.appService.GetApplication(input.ID)
	if svcErr != nil {
		return nil, nil, fmt.Errorf("failed to get application: %s", svcErr.ErrorDescription)
	}

	// Build update DTO starting with all current values
	updateDTO := &model.ApplicationDTO{
		ID:                        currentApp.ID,
		Name:                      currentApp.Name,
		Description:               currentApp.Description,
		AuthFlowID:                currentApp.AuthFlowID,
		RegistrationFlowID:        currentApp.RegistrationFlowID,
		IsRegistrationFlowEnabled: currentApp.IsRegistrationFlowEnabled,
		BrandingID:                currentApp.BrandingID,
		Template:                  currentApp.Template,
		URL:                       currentApp.URL,
		LogoURL:                   currentApp.LogoURL,
		TosURI:                    currentApp.TosURI,
		PolicyURI:                 currentApp.PolicyURI,
		Contacts:                  currentApp.Contacts,
		Token:                     currentApp.Token,
		Certificate:               currentApp.Certificate,
		AllowedUserTypes:          currentApp.AllowedUserTypes,
	}

	// Apply updates only for provided fields (PATCH semantics)
	if input.Name != "" {
		updateDTO.Name = input.Name
	}
	if input.Description != "" {
		updateDTO.Description = input.Description
	}
	if input.AuthFlowID != "" {
		updateDTO.AuthFlowID = input.AuthFlowID
	}
	if input.RegistrationFlowID != "" {
		updateDTO.RegistrationFlowID = input.RegistrationFlowID
	}
	// Note: IsRegistrationFlowEnabled is a bool, so we check if it differs from default
	// In practice, LLMs will provide it if they want to change it
	if input.IsRegistrationFlowEnabled != currentApp.IsRegistrationFlowEnabled {
		updateDTO.IsRegistrationFlowEnabled = input.IsRegistrationFlowEnabled
	}
	if input.BrandingID != "" {
		updateDTO.BrandingID = input.BrandingID
	}
	if input.URL != "" {
		updateDTO.URL = input.URL
	}
	if input.LogoURL != "" {
		updateDTO.LogoURL = input.LogoURL
	}
	if input.TosURI != "" {
		updateDTO.TosURI = input.TosURI
	}
	if input.PolicyURI != "" {
		updateDTO.PolicyURI = input.PolicyURI
	}
	if len(input.Contacts) > 0 {
		updateDTO.Contacts = input.Contacts
	}
	if len(input.AllowedUserTypes) > 0 {
		updateDTO.AllowedUserTypes = input.AllowedUserTypes
	}

	// Handle OAuth config updates if provided
	if len(input.InboundAuthConfig) > 0 && input.InboundAuthConfig[0].OAuthAppConfig != nil {
		inputOAuth := input.InboundAuthConfig[0].OAuthAppConfig

		// Start with current OAuth config if it exists
		var currentOAuth *model.OAuthAppConfigComplete
		if len(currentApp.InboundAuthConfig) > 0 {
			currentOAuth = currentApp.InboundAuthConfig[0].OAuthAppConfig
		}

		// Build OAuth config DTO
		oauthConfig := &model.OAuthAppConfigDTO{}

		// Preserve or update each OAuth field
		if currentOAuth != nil {
			oauthConfig.ClientID = currentOAuth.ClientID
			oauthConfig.ClientSecret = currentOAuth.ClientSecret
			oauthConfig.RedirectURIs = currentOAuth.RedirectURIs
			oauthConfig.GrantTypes = currentOAuth.GrantTypes
			oauthConfig.ResponseTypes = currentOAuth.ResponseTypes
			oauthConfig.TokenEndpointAuthMethod = currentOAuth.TokenEndpointAuthMethod
			oauthConfig.PKCERequired = currentOAuth.PKCERequired
			oauthConfig.PublicClient = currentOAuth.PublicClient
			oauthConfig.Token = currentOAuth.Token
			oauthConfig.Scopes = currentOAuth.Scopes
		}

		// Apply updates only for provided OAuth fields
		if len(inputOAuth.RedirectURIs) > 0 {
			oauthConfig.RedirectURIs = inputOAuth.RedirectURIs
		}
		if len(inputOAuth.GrantTypes) > 0 {
			oauthConfig.GrantTypes = inputOAuth.GrantTypes
		}
		if len(inputOAuth.ResponseTypes) > 0 {
			oauthConfig.ResponseTypes = inputOAuth.ResponseTypes
		}
		if inputOAuth.TokenEndpointAuthMethod != "" {
			oauthConfig.TokenEndpointAuthMethod = inputOAuth.TokenEndpointAuthMethod
		}
		// For booleans, we update if they differ (LLM provides them explicitly)
		if currentOAuth == nil || inputOAuth.PKCERequired != currentOAuth.PKCERequired {
			oauthConfig.PKCERequired = inputOAuth.PKCERequired
		}
		if currentOAuth == nil || inputOAuth.PublicClient != currentOAuth.PublicClient {
			oauthConfig.PublicClient = inputOAuth.PublicClient
		}
		if len(inputOAuth.Scopes) > 0 {
			oauthConfig.Scopes = inputOAuth.Scopes
		}
		if inputOAuth.Token != nil {
			if oauthConfig.Token == nil {
				oauthConfig.Token = &model.OAuthTokenConfig{}
			}
			if inputOAuth.Token.Issuer != "" {
				oauthConfig.Token.Issuer = inputOAuth.Token.Issuer
			}
			if inputOAuth.Token.AccessToken != nil {
				oauthConfig.Token.AccessToken = inputOAuth.Token.AccessToken
			}
			if inputOAuth.Token.IDToken != nil {
				oauthConfig.Token.IDToken = inputOAuth.Token.IDToken
			}
		}

		updateDTO.InboundAuthConfig = []model.InboundAuthConfigDTO{
			{
				Type:           model.OAuthInboundAuthType,
				OAuthAppConfig: oauthConfig,
			},
		}
	} else if len(currentApp.InboundAuthConfig) > 0 {
		// Preserve existing OAuth config if not updating
		currentOAuth := currentApp.InboundAuthConfig[0].OAuthAppConfig
		updateDTO.InboundAuthConfig = []model.InboundAuthConfigDTO{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &model.OAuthAppConfigDTO{
					ClientID:                currentOAuth.ClientID,
					ClientSecret:            currentOAuth.ClientSecret,
					RedirectURIs:            currentOAuth.RedirectURIs,
					GrantTypes:              currentOAuth.GrantTypes,
					ResponseTypes:           currentOAuth.ResponseTypes,
					TokenEndpointAuthMethod: currentOAuth.TokenEndpointAuthMethod,
					PKCERequired:            currentOAuth.PKCERequired,
					PublicClient:            currentOAuth.PublicClient,
					Token:                   currentOAuth.Token,
					Scopes:                  currentOAuth.Scopes,
				},
			},
		}
	}

	updatedApp, svcErr := t.appService.UpdateApplication(input.ID, updateDTO)
	if svcErr != nil {
		return nil, nil, fmt.Errorf("failed to update application: %s", svcErr.ErrorDescription)
	}

	return nil, updatedApp, nil
}

// GetApplicationTemplates handles the get_application_templates tool call.
// Returns pre-configured templates with placeholder values for common application types.
func (t *ApplicationTools) GetApplicationTemplates(
	ctx context.Context,
	req *mcp.CallToolRequest,
	_ any,
) (*mcp.CallToolResult, map[string]interface{}, error) {

	templates := map[string]interface{}{
		"spa": map[string]interface{}{
			"name": "<APP_NAME>",
			"inbound_auth_config": []map[string]interface{}{
				{
					"type": "oauth2",
					"oauth_app_config": map[string]interface{}{
						"redirect_uris":              []string{"<REDIRECT_URI>"},
						"grant_types":                []string{"authorization_code", "refresh_token"},
						"token_endpoint_auth_method": "none",
						"pkce_required":              true,
						"public_client":              true,
					},
				},
			},
		},
		"mobile": map[string]interface{}{
			"name": "<APP_NAME>",
			"inbound_auth_config": []map[string]interface{}{
				{
					"type": "oauth2",
					"oauth_app_config": map[string]interface{}{
						"redirect_uris":              []string{"<CUSTOM_SCHEME>://callback"},
						"grant_types":                []string{"authorization_code", "refresh_token"},
						"token_endpoint_auth_method": "none",
						"pkce_required":              true,
						"public_client":              true,
					},
				},
			},
		},
		"server": map[string]interface{}{
			"name": "<APP_NAME>",
			"inbound_auth_config": []map[string]interface{}{
				{
					"type": "oauth2",
					"oauth_app_config": map[string]interface{}{
						"redirect_uris": []string{"<REDIRECT_URI>"},
						"grant_types":   []string{"authorization_code", "refresh_token"},
						"pkce_required": true,
					},
				},
			},
		},
		"m2m": map[string]interface{}{
			"name": "<APP_NAME>",
			"inbound_auth_config": []map[string]interface{}{
				{
					"type": "oauth2",
					"oauth_app_config": map[string]interface{}{
						"grant_types": []string{"client_credentials"},
					},
				},
			},
		},
	}

	return nil, templates, nil
}
