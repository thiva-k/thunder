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

package mcp

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

// ListApplicationsInput represents the input for the list_applications tool.
type ListApplicationsInput struct{}

// ApplicationListOutput represents the output for list_applications tool.
type ApplicationListOutput struct {
	TotalCount   int                              `json:"total_count"`
	Applications []model.BasicApplicationResponse `json:"applications"`
}

// RegisterTools registers all application tools with the MCP server.
func (t *ApplicationTools) RegisterTools(server *mcp.Server) {
	// Generate schema with enum support for ApplicationDTO
	appDTOSchema := GenerateSchema[model.ApplicationDTO](
		WithEnum("grant_types", oauth2const.GetSupportedGrantTypes()),
		WithEnum("response_types", oauth2const.GetSupportedResponseTypes()),
		WithEnum("token_endpoint_auth_method", oauth2const.GetSupportedTokenEndpointAuthMethods()),
	)
	// Generate schema for update with 'id' as required
	updateAppDTOSchema := GenerateSchema[model.ApplicationDTO](
		WithEnum("grant_types", oauth2const.GetSupportedGrantTypes()),
		WithEnum("response_types", oauth2const.GetSupportedResponseTypes()),
		WithEnum("token_endpoint_auth_method", oauth2const.GetSupportedTokenEndpointAuthMethods()),
		WithRequired("id"),
	)

	mcp.AddTool(server, &mcp.Tool{
		Name: "list_applications",
		Description: `List all registered OAuth applications.
		
Outputs: A list of applications containing ID, Name, ClientID, and associated Flow IDs.
Next Steps: Use the returned 'id' with get_application or update_application.`,
		Annotations: &mcp.ToolAnnotations{
			Title:        "List Applications",
			ReadOnlyHint: true,
		},
	}, t.ListApplications)

	mcp.AddTool(server, &mcp.Tool{
		Name: "get_application",
		Description: `Retrieve full details of a specific application.
		
Inputs: 'id' (UUID of the application).
Outputs: Complete application configuration including OAuth settings, branding, and flow associations.`,
		Annotations: &mcp.ToolAnnotations{
			Title:        "Get Application",
			ReadOnlyHint: true,
		},
	}, t.GetApplication)

	mcp.AddTool(server, &mcp.Tool{
		Name: "create_application",
		Description: `Create a new OAuth application.

Prerequisites:
- Authentication/Registration flows should be created first if custom flows are needed.

Inputs:
- name (Required): Application display name.
- auth_flow_id: ID of the authentication flow (default flow used if omitted).
- registration_flow_id: ID of the registration flow.
- inbound_auth_config: An array containing the OAuth configuration.
  Example Structure:
  [
    {
      "type": "oauth2",
      "config": {
        "grant_types": ["authorization_code", "refresh_token"],
        "response_types": ["code"],
        "pkce_required": true,
        "redirect_uris": ["https://myapp.com/callback"],
        "public_client": false
      }
    }
  ]

  Supported Enums:
  - grant_types: "authorization_code", "refresh_token", "client_credentials", "urn:ietf:params:oauth:grant-type:token-exchange"
  - response_types: "code"

Outputs: Created application including generated 'client_id' and 'client_secret'.

Next Steps: Configure your OIDC client (Gate/Frontend) with the returned credentials.`,
		InputSchema: appDTOSchema,
		Annotations: &mcp.ToolAnnotations{
			Title:          "Create Application",
			IdempotentHint: true,
		},
	}, t.CreateApplication)

	mcp.AddTool(server, &mcp.Tool{
		Name: "update_application",
		Description: `Update an existing OAuth application configuration.

Inputs:
- id (Required): The application UUID.
- All other fields from 'create_application' (Full Replacement).
  IMPORTANT: You must provide the COMPLETE object. Missing fields will be reset to defaults/nil.
  
  Example Update:
  {
    "id": "...",
    "name": "New Name",
    "inbound_auth_config": [ { "type": "oauth2", "config": { ... } } ]
  }

Outputs: The updated application configuration.`,
		InputSchema: updateAppDTOSchema,
		Annotations: &mcp.ToolAnnotations{
			Title:          "Update Application",
			IdempotentHint: true,
		},
	}, t.UpdateApplication)

	mcp.AddTool(server, &mcp.Tool{
		Name: "delete_application",
		Description: `Permanently delete an application.
		
Inputs: 'id' (UUID).
Impact: Invalidates all active tokens and client credentials associated with this app.`,
		Annotations: &mcp.ToolAnnotations{
			Title:           "Delete Application",
			DestructiveHint: ptr(true),
		},
	}, t.DeleteApplication)
}

// ptr returns a pointer to the given value.
func ptr[T any](v T) *T {
	return &v
}

// ListApplications handles the list_applications tool call.
func (t *ApplicationTools) ListApplications(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input ListApplicationsInput,
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

// UpdateApplication handles the update_application tool call.
// Uses ApplicationDTO directly - ID field is required for update.
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

// DeleteApplication handles the delete_application tool call.
func (t *ApplicationTools) DeleteApplication(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input IDInput,
) (*mcp.CallToolResult, DeleteOutput, error) {

	svcErr := t.appService.DeleteApplication(input.ID)
	if svcErr != nil {
		return nil, DeleteOutput{
			Success: false,
			Message: fmt.Sprintf("Failed to delete application: %s", svcErr.ErrorDescription),
		}, nil
	}

	return nil, DeleteOutput{
		Success: true,
		Message: fmt.Sprintf("Application %s deleted successfully", input.ID),
	}, nil
}
