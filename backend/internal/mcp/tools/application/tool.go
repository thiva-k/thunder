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

// Package application provides application management tool models.
//
//nolint:lll
package application

import (
	"context"
	"fmt"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	appsvc "github.com/asgardeo/thunder/internal/application"
	"github.com/asgardeo/thunder/internal/application/model"
	"github.com/asgardeo/thunder/internal/mcp/tools"
	oauth2const "github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
)

// ApplicationTools provides MCP tools for managing Thunder applications.
type ApplicationTools struct {
	appService appsvc.ApplicationServiceInterface
}

// NewApplicationTools creates a new ApplicationTools instance.
func NewApplicationTools(appService appsvc.ApplicationServiceInterface) *ApplicationTools {
	return &ApplicationTools{
		appService: appService,
	}
}

// RegisterTools registers all application tools with the MCP server.
func (t *ApplicationTools) RegisterTools(server *mcp.Server) {
	// Common schema modifiers for ApplicationDTO
	appSchemaModifiers := []func(*jsonschema.Schema){
		tools.WithEnum("config", "grant_types", oauth2const.GetSupportedGrantTypes()),
		tools.WithEnum("config", "response_types", oauth2const.GetSupportedResponseTypes()),
		tools.WithEnum("config", "token_endpoint_auth_method", oauth2const.GetSupportedTokenEndpointAuthMethods()),
		tools.WithEnum("inbound_auth_config", "type", []string{string(model.OAuthInboundAuthType)}),
	}

	createAppSchema := tools.GenerateSchema[model.ApplicationDTO](
		append(appSchemaModifiers, tools.WithRemove("id"))...,
	)

	updateAppSchema := tools.GenerateSchema[model.ApplicationDTO](
		append(appSchemaModifiers, tools.WithRequired("id"))...,
	)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_applications",
		Description: `List all registered applications.`,
		Annotations: &mcp.ToolAnnotations{
			Title:        "List Applications",
			ReadOnlyHint: true,
		},
	}, t.ListApplications)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_application_by_id",
		Description: `Retrieve full details of an application by ID including OAuth settings, branding, and flow associations.`,
		Annotations: &mcp.ToolAnnotations{
			Title:        "Get Application by ID",
			ReadOnlyHint: true,
		},
	}, t.GetApplicationByID)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_application_by_client_id",
		Description: `Retrieve full details of an application by client_id including OAuth settings, branding, and flow associations`,
		Annotations: &mcp.ToolAnnotations{
			Title:        "Get Application by Client ID",
			ReadOnlyHint: true,
		},
	}, t.GetApplicationByClientID)

	mcp.AddTool(server, &mcp.Tool{
		Name: "create_application",
		Description: `Create a new application optionally with OAuth configuration.

Use get_application_templates to get pre-configured minimal templates for common app types (SPA, Mobile, Server, M2M).

Prerequisites: Create flows first using create_flow if custom authentication/registration flows are needed.

Behavior: If auth_flow_id is omitted, the default authentication flow is used.`,
		InputSchema: createAppSchema,
		Annotations: &mcp.ToolAnnotations{
			Title:          "Create Application",
			IdempotentHint: false,
		},
	}, t.CreateApplication)

	mcp.AddTool(server, &mcp.Tool{
		Name: "update_application",
		Description: `Update an existing application (full replacement).

This is a PUT operation - you must provide the COMPLETE application object.

Workflow:
1. Use get_application_by_id to get current state
2. Modify the fields you want to change
3. Send the complete object back

Any field not provided will be reset to empty/default.`,
		InputSchema: updateAppSchema,
		Annotations: &mcp.ToolAnnotations{
			Title:          "Update Application",
			IdempotentHint: true,
		},
	}, t.UpdateApplication)

	mcp.AddTool(server, &mcp.Tool{
		Name: "get_application_templates",
		Description: `Get minimal OAuth configuration templates for common application types.

Templates contain ONLY the required fields to create each app type. Optional fields with service-layer defaults are omitted.`,
		Annotations: &mcp.ToolAnnotations{
			Title:        "Get Application Templates",
			ReadOnlyHint: true,
		},
	}, t.GetApplicationTemplates)
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
	input tools.IDInput,
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

// UpdateApplication handles the update_application tool call with complete replacement.
func (t *ApplicationTools) UpdateApplication(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input model.ApplicationDTO,
) (*mcp.CallToolResult, *model.ApplicationDTO, error) {
	updatedApp, svcErr := t.appService.UpdateApplication(input.ID, &input)
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
					"config": map[string]interface{}{
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
					"config": map[string]interface{}{
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
					"config": map[string]interface{}{
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
					"config": map[string]interface{}{
						"grant_types": []string{"client_credentials"},
					},
				},
			},
		},
	}

	return nil, templates, nil
}
