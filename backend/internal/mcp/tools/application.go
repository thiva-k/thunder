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

// SPAInput represents input for creating a Single Page Application (Browser platform).
type SPAInput struct {
	Name        string `json:"name" jsonschema:"required" jsonschema_description:"Name of the application"`
	Description string `json:"description,omitempty" jsonschema_description:"Description of the application"`
	RedirectURL string `json:"redirect_url" jsonschema:"required" jsonschema_description:"Redirect URL for OAuth callbacks (e.g. http://localhost:3000)"`
	LogoURL     string `json:"logo_url,omitempty" jsonschema_description:"URL to the application logo"`
}

// MobileAppInput represents input for creating a Mobile Application.
type MobileAppInput struct {
	Name        string `json:"name" jsonschema:"required" jsonschema_description:"Name of the application"`
	Description string `json:"description,omitempty" jsonschema_description:"Description of the application"`
	RedirectURL string `json:"redirect_url" jsonschema:"required" jsonschema_description:"Redirect URL with custom scheme (e.g. myapp://callback)"`
	LogoURL     string `json:"logo_url,omitempty" jsonschema_description:"URL to the application logo"`
}

// ServerAppInput represents input for creating a Server Application (SSR/Traditional Web).
type ServerAppInput struct {
	Name        string `json:"name" jsonschema:"required" jsonschema_description:"Name of the application"`
	Description string `json:"description,omitempty" jsonschema_description:"Description of the application"`
	RedirectURL string `json:"redirect_url" jsonschema:"required" jsonschema_description:"Redirect URL for OAuth callbacks (e.g. https://example.com/callback)"`
	LogoURL     string `json:"logo_url,omitempty" jsonschema_description:"URL to the application logo"`
}

// M2MAppInput represents input for creating a Machine-to-Machine Application (Backend).
type M2MAppInput struct {
	Name        string `json:"name" jsonschema:"required" jsonschema_description:"Name of the application"`
	Description string `json:"description,omitempty" jsonschema_description:"Description of the application"`
	LogoURL     string `json:"logo_url,omitempty" jsonschema_description:"URL to the application logo"`
}

// OAuthConfigUpdateInput represents input for updating OAuth configuration.
type OAuthConfigUpdateInput struct {
	ID                      string   `json:"id" jsonschema:"required" jsonschema_description:"ID of the application to update"`
	RedirectURIs            []string `json:"redirect_uris,omitempty" jsonschema_description:"List of allowed redirect URIs"`
	GrantTypes              []string `json:"grant_types,omitempty" jsonschema_description:"Allowed OAuth grant types"`
	ResponseTypes           []string `json:"response_types,omitempty" jsonschema_description:"Allowed OAuth response types"`
	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method,omitempty" jsonschema_description:"Token endpoint authentication method"`
	PKCERequired            *bool    `json:"pkce_required,omitempty" jsonschema_description:"Whether PKCE is required"`
	PublicClient            *bool    `json:"public_client,omitempty" jsonschema_description:"Whether this is a public client (no client secret)"`
}

// RegisterTools registers all application tools with the MCP server.
func (t *ApplicationTools) RegisterTools(server *mcp.Server) {
	// Generate schema for update with 'id' as required
	updateAppSchema := GenerateSchema[model.ApplicationDTO](
		WithEnum("grant_types", oauth2const.GetSupportedGrantTypes()),
		WithEnum("response_types", oauth2const.GetSupportedResponseTypes()),
		WithEnum("token_endpoint_auth_method", oauth2const.GetSupportedTokenEndpointAuthMethods()),
		WithRequired("id"),
	)

	// Generate schema for OAuth config update with enums
	oauthConfigSchema := GenerateSchema[OAuthConfigUpdateInput](
		WithEnum("grant_types", oauth2const.GetSupportedGrantTypes()),
		WithEnum("response_types", oauth2const.GetSupportedResponseTypes()),
		WithEnum("token_endpoint_auth_method", oauth2const.GetSupportedTokenEndpointAuthMethods()),
	)

	mcp.AddTool(server, &mcp.Tool{
		Name: "list_applications",
		Description: `List all registered applications.

Related: Use returned 'id' with get_application or update_application_oauth_config.`,
		Annotations: &mcp.ToolAnnotations{
			Title:        "List Applications",
			ReadOnlyHint: true,
		},
	}, t.ListApplications)

	mcp.AddTool(server, &mcp.Tool{
		Name: "get_application",
		Description: `Retrieve full details of an application including OAuth settings, branding, and flow associations.

Related: Use before update tools to review current configuration.`,
		Annotations: &mcp.ToolAnnotations{
			Title:        "Get Application",
			ReadOnlyHint: true,
		},
	}, t.GetApplication)

	mcp.AddTool(server, &mcp.Tool{
		Name: "create_single_page_app",
		Description: `Create a new Single Page Application (SPA) with pre-configured OAuth settings.

Platform: BROWSER (web applications running in browser)

This tool creates an application optimized for SPAs (React, Vue, Angular, etc.) with:
- Authorization Code grant with PKCE
- Refresh token support
- Public client (no client secret)
- Code response type

Prerequisites: None - uses default authentication flow.

Outputs: Created application with 'client_id' (no client secret for public clients).

Related: Use integrate_react_sdk after creation for SDK integration instructions.`,
		Annotations: &mcp.ToolAnnotations{
			Title:          "Create Single Page App",
			IdempotentHint: true,
		},
	}, t.CreateSinglePageApp)

	mcp.AddTool(server, &mcp.Tool{
		Name: "create_mobile_app",
		Description: `Create a new Mobile Application with pre-configured OAuth settings.

Platform: MOBILE (native mobile applications)

This tool creates an application optimized for mobile apps (iOS, Android, React Native, Flutter) with:
- Authorization Code grant with PKCE (required for mobile security)
- Refresh token support (for long-lived sessions)
- Public client (mobile apps cannot securely store secrets)
- Custom URL scheme redirect support (e.g., myapp://callback)

Prerequisites: None - uses default authentication flow.

Outputs: Created application with 'client_id' (no client secret for public clients).

Note: Ensure your redirect URL uses a custom scheme registered with your mobile app.`,
		Annotations: &mcp.ToolAnnotations{
			Title:          "Create Mobile App",
			IdempotentHint: true,
		},
	}, t.CreateMobileApp)

	mcp.AddTool(server, &mcp.Tool{
		Name: "create_server_app",
		Description: `Create a new Server Application with pre-configured OAuth settings.

Platform: SERVER (server-side applications with confidential client)

This tool creates an application optimized for server-side apps (Next.js, Nuxt, traditional web apps) with:
- Authorization Code grant with PKCE
- Refresh token support
- Confidential client (can securely store client secret)
- Client secret basic authentication
- Code response type

Prerequisites: None - uses default authentication flow.

Outputs: Created application with 'client_id' and 'client_secret'.

Note: Store the client secret securely on your server. Never expose it to the browser.`,
		Annotations: &mcp.ToolAnnotations{
			Title:          "Create Server App",
			IdempotentHint: true,
		},
	}, t.CreateServerApp)

	mcp.AddTool(server, &mcp.Tool{
		Name: "create_m2m_app",
		Description: `Create a new Machine-to-Machine (M2M) Application with pre-configured OAuth settings.

Platform: BACKEND (machine-to-machine backend services)

This tool creates an application optimized for backend services, APIs, microservices, and cron jobs with:
- Client Credentials grant (no user interaction)
- Confidential client (uses client secret)
- Client secret basic authentication
- No redirect URIs (server-to-server communication)

Prerequisites: None - M2M apps don't use authentication flows.

Outputs: Created application with 'client_id' and 'client_secret'.

Use Case: Service-to-service authentication, API access, scheduled jobs.`,
		Annotations: &mcp.ToolAnnotations{
			Title:          "Create M2M App",
			IdempotentHint: true,
		},
	}, t.CreateM2MApp)

	mcp.AddTool(server, &mcp.Tool{
		Name: "update_application",
		Description: `Update an existing application configuration (PUT semantics).

Prerequisites: Use get_application first to retrieve current configuration.

IMPORTANT: This is a full replacement. 
Missing fields will be reset to defaults/nil which is not desirable for existing fields. Provide the complete existing object.`,
		InputSchema: updateAppSchema,
		Annotations: &mcp.ToolAnnotations{
			Title:          "Update Application",
			IdempotentHint: true,
		},
	}, t.UpdateApplication)

	mcp.AddTool(server, &mcp.Tool{
		Name: "update_application_oauth_config",
		Description: `Update OAuth/OIDC configuration of an application.

This tool allows updating specific OAuth settings without affecting other application properties:
- Redirect URIs
- Grant types and response types
- Token endpoint authentication method
- PKCE and public client settings

Prerequisites: Use get_application first to see current OAuth configuration.

Behavior: Only provided fields are updated. Omitted fields remain unchanged.

Related: Use create_single_page_app for pre-configured SPA settings.`,
		InputSchema: oauthConfigSchema,
		Annotations: &mcp.ToolAnnotations{
			Title:          "Update Application OAuth Config",
			IdempotentHint: true,
		},
	}, t.UpdateApplicationOAuthConfig)
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

// GetApplication handles the get_application tool call.
func (t *ApplicationTools) GetApplication(
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

// CreateMobileApp handles the create_mobile_app tool call.
// Creates an application pre-configured for mobile apps with PKCE and refresh tokens.
func (t *ApplicationTools) CreateMobileApp(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input MobileAppInput,
) (*mcp.CallToolResult, *model.ApplicationDTO, error) {

	// Build application DTO with Mobile-optimized OAuth config (matches MOBILE platform from frontend)
	appDTO := &model.ApplicationDTO{
		Name:        input.Name,
		Description: input.Description,
		LogoURL:     input.LogoURL,
		InboundAuthConfig: []model.InboundAuthConfigDTO{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &model.OAuthAppConfigDTO{
					RedirectURIs: []string{input.RedirectURL},
					GrantTypes: []oauth2const.GrantType{
						oauth2const.GrantTypeAuthorizationCode,
						oauth2const.GrantTypeRefreshToken,
					},
					ResponseTypes: []oauth2const.ResponseType{
						oauth2const.ResponseTypeCode,
					},
					TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodNone,
					PKCERequired:            true,
					PublicClient:            true,
				},
			},
		},
	}

	createdApp, svcErr := t.appService.CreateApplication(appDTO)
	if svcErr != nil {
		return nil, nil, fmt.Errorf("failed to create mobile application: %s", svcErr.ErrorDescription)
	}

	return nil, createdApp, nil
}

// CreateServerApp handles the create_server_app tool call.
// Creates an application pre-configured for server-side apps with client secret.
func (t *ApplicationTools) CreateServerApp(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input ServerAppInput,
) (*mcp.CallToolResult, *model.ApplicationDTO, error) {

	// Build application DTO with Server-optimized OAuth config (matches SERVER platform from frontend)
	appDTO := &model.ApplicationDTO{
		Name:        input.Name,
		Description: input.Description,
		LogoURL:     input.LogoURL,
		InboundAuthConfig: []model.InboundAuthConfigDTO{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &model.OAuthAppConfigDTO{
					RedirectURIs: []string{input.RedirectURL},
					GrantTypes: []oauth2const.GrantType{
						oauth2const.GrantTypeAuthorizationCode,
						oauth2const.GrantTypeRefreshToken,
					},
					ResponseTypes: []oauth2const.ResponseType{
						oauth2const.ResponseTypeCode,
					},
					TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodClientSecretBasic,
					PKCERequired:            true,
					PublicClient:            false,
				},
			},
		},
	}

	createdApp, svcErr := t.appService.CreateApplication(appDTO)
	if svcErr != nil {
		return nil, nil, fmt.Errorf("failed to create server application: %s", svcErr.ErrorDescription)
	}

	return nil, createdApp, nil
}

// CreateM2MApp handles the create_m2m_app tool call.
// Creates an application pre-configured for machine-to-machine communication.
func (t *ApplicationTools) CreateM2MApp(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input M2MAppInput,
) (*mcp.CallToolResult, *model.ApplicationDTO, error) {

	// Build application DTO with M2M-optimized OAuth config (matches BACKEND platform from frontend)
	appDTO := &model.ApplicationDTO{
		Name:        input.Name,
		Description: input.Description,
		LogoURL:     input.LogoURL,
		InboundAuthConfig: []model.InboundAuthConfigDTO{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &model.OAuthAppConfigDTO{
					GrantTypes: []oauth2const.GrantType{
						oauth2const.GrantTypeClientCredentials,
					},
					TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodClientSecretBasic,
					PKCERequired:            false,
					PublicClient:            false,
				},
			},
		},
	}

	createdApp, svcErr := t.appService.CreateApplication(appDTO)
	if svcErr != nil {
		return nil, nil, fmt.Errorf("failed to create M2M application: %s", svcErr.ErrorDescription)
	}

	return nil, createdApp, nil
}

// UpdateApplication handles the update_application tool call.
func (t *ApplicationTools) UpdateApplication(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input model.ApplicationDTO,
) (*mcp.CallToolResult, *model.ApplicationDTO, error) {

	// Verify existence
	_, svcErr := t.appService.GetApplication(input.ID)
	if svcErr != nil {
		return nil, nil, fmt.Errorf("failed to get application: %s", svcErr.ErrorDescription)
	}

	updatedApp, svcErr := t.appService.UpdateApplication(input.ID, &input)
	if svcErr != nil {
		return nil, nil, fmt.Errorf("failed to update application: %s", svcErr.ErrorDescription)
	}

	return nil, updatedApp, nil
}

// CreateSinglePageApp handles the create_single_page_app tool call.
// Creates an application pre-configured for SPAs with PKCE and public client settings.
func (t *ApplicationTools) CreateSinglePageApp(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input SPAInput,
) (*mcp.CallToolResult, *model.ApplicationDTO, error) {

	// Build application DTO with SPA-optimized OAuth config (matches BROWSER platform from frontend)
	appDTO := &model.ApplicationDTO{
		Name:        input.Name,
		Description: input.Description,
		LogoURL:     input.LogoURL,
		InboundAuthConfig: []model.InboundAuthConfigDTO{
			{
				Type: model.OAuthInboundAuthType,
				OAuthAppConfig: &model.OAuthAppConfigDTO{
					RedirectURIs: []string{input.RedirectURL},
					GrantTypes: []oauth2const.GrantType{
						oauth2const.GrantTypeAuthorizationCode,
						oauth2const.GrantTypeRefreshToken,
					},
					ResponseTypes: []oauth2const.ResponseType{
						oauth2const.ResponseTypeCode,
					},
					TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodNone,
					PKCERequired:            true,
					PublicClient:            true,
				},
			},
		},
	}

	createdApp, svcErr := t.appService.CreateApplication(appDTO)
	if svcErr != nil {
		return nil, nil, fmt.Errorf("failed to create single page application: %s", svcErr.ErrorDescription)
	}

	return nil, createdApp, nil
}

// UpdateApplicationOAuthConfig handles the update_application_oauth_config tool call.
// Updates only OAuth configuration without affecting other application properties.
func (t *ApplicationTools) UpdateApplicationOAuthConfig(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input OAuthConfigUpdateInput,
) (*mcp.CallToolResult, *model.ApplicationDTO, error) {

	// Get current application
	currentApp, svcErr := t.appService.GetApplication(input.ID)
	if svcErr != nil {
		return nil, nil, fmt.Errorf("failed to get application: %s", svcErr.ErrorDescription)
	}

	// Ensure OAuth config exists
	if len(currentApp.InboundAuthConfig) == 0 || currentApp.InboundAuthConfig[0].OAuthAppConfig == nil {
		return nil, nil, fmt.Errorf("application does not have OAuth configuration")
	}

	currentOAuth := currentApp.InboundAuthConfig[0].OAuthAppConfig

	// Build update DTO preserving all existing fields
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

	// Update OAuth config with new values or keep existing
	oauthConfig := &model.OAuthAppConfigDTO{
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
	}

	// Apply updates only for provided fields
	if len(input.RedirectURIs) > 0 {
		oauthConfig.RedirectURIs = input.RedirectURIs
	}
	if len(input.GrantTypes) > 0 {
		grantTypes := make([]oauth2const.GrantType, len(input.GrantTypes))
		for i, gt := range input.GrantTypes {
			grantTypes[i] = oauth2const.GrantType(gt)
		}
		oauthConfig.GrantTypes = grantTypes
	}
	if len(input.ResponseTypes) > 0 {
		responseTypes := make([]oauth2const.ResponseType, len(input.ResponseTypes))
		for i, rt := range input.ResponseTypes {
			responseTypes[i] = oauth2const.ResponseType(rt)
		}
		oauthConfig.ResponseTypes = responseTypes
	}
	if input.TokenEndpointAuthMethod != "" {
		oauthConfig.TokenEndpointAuthMethod = oauth2const.TokenEndpointAuthMethod(input.TokenEndpointAuthMethod)
	}
	if input.PKCERequired != nil {
		oauthConfig.PKCERequired = *input.PKCERequired
	}
	if input.PublicClient != nil {
		oauthConfig.PublicClient = *input.PublicClient
	}

	updateDTO.InboundAuthConfig = []model.InboundAuthConfigDTO{
		{
			Type:           model.OAuthInboundAuthType,
			OAuthAppConfig: oauthConfig,
		},
	}

	updatedApp, svcErr := t.appService.UpdateApplication(input.ID, updateDTO)
	if svcErr != nil {
		return nil, nil, fmt.Errorf("failed to update application OAuth config: %s", svcErr.ErrorDescription)
	}

	return nil, updatedApp, nil
}
