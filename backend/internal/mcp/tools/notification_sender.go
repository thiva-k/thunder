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

	"github.com/asgardeo/thunder/internal/notification"
	notifcommon "github.com/asgardeo/thunder/internal/notification/common"
	"github.com/asgardeo/thunder/internal/system/cmodels"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// NotificationSenderTools provides MCP tools for managing notification senders.
type NotificationSenderTools struct {
	notifService notification.NotificationSenderMgtSvcInterface
}

// NewNotificationSenderTools creates a new NotificationSenderTools instance.
func NewNotificationSenderTools(notifService notification.NotificationSenderMgtSvcInterface) *NotificationSenderTools {
	return &NotificationSenderTools{
		notifService: notifService,
	}
}

// ListSendersInput represents the input for the list_notification_senders tool.
type ListSendersInput struct{}

// SenderListOutput represents the output for list_notification_senders tool.
type SenderListOutput struct {
	TotalCount int                                      `json:"total_count"`
	Senders    []notifcommon.NotificationSenderResponse `json:"senders"`
}

// RegisterTools registers all notification sender tools with the MCP server.
func (t *NotificationSenderTools) RegisterTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name: "list_notification_senders",
		Description: `List all configured notification senders (SMS, Email providers).

Inputs: None.

Outputs:
- Returns a list of senders with their IDs, names, and providers.

Next Steps:
- Use the returned 'id' to configure flow nodes (e.g., SMSOTPAuthExecutor) or to get detailed sender info.`,
		Annotations: &mcp.ToolAnnotations{
			Title:        "List Notification Senders",
			ReadOnlyHint: true,
		},
	}, t.ListSenders)

	mcp.AddTool(server, &mcp.Tool{
		Name: "get_notification_sender",
		Description: `Get detailed information about a specific notification sender by its ID.

Inputs:
- id (required): The unique identifier of the notification sender.

Outputs:
- Returns the full sender configuration including name, provider, and properties (masked secrets).

Next Steps:
- Use this information to review configuration or prepare for an update.`,
		Annotations: &mcp.ToolAnnotations{
			Title:        "Get Notification Sender",
			ReadOnlyHint: true,
		},
	}, t.GetSender)

	providerEnum := []string{"twilio", "vonage", "custom"}
	createSenderSchema := GenerateSchema[notifcommon.NotificationSenderRequest](
		WithEnum("provider", providerEnum),
	)

	mcp.AddTool(server, &mcp.Tool{
		Name: "create_notification_sender",
		Description: `Create a new notification sender for SMS/message delivery.

Inputs:
- name (required): Display name for the sender.
- provider (required): Type of provider.
  - "twilio": Twilio SMS service
  - "vonage": Vonage/Nexmo SMS service
  - "custom": Custom webhook endpoint
- properties (required): Configuration properties based on provider.
  - Twilio: account_sid, auth_token (secret), from_number
  - Vonage: api_key, api_secret (secret), from_number
  - Custom: url, http_method (POST), content_type (JSON)

Example (Custom Webhook):
{
  "name": "My Webhook Sender",
  "provider": "custom",
  "properties": [
    {"name": "url", "value": "https://api.example.com/send", "is_secret": false},
    {"name": "http_method", "value": "POST", "is_secret": false},
    {"name": "content_type", "value": "JSON", "is_secret": false}
  ]
}

Outputs:
- Returns the created sender with its assigned ID.

Next Steps:
- Use the returned 'id' in authentication flows (e.g., set 'senderId' property in SMSOTPAuthExecutor).`,
		InputSchema: createSenderSchema,
		Annotations: &mcp.ToolAnnotations{
			Title:          "Create Notification Sender",
			IdempotentHint: true,
		},
	}, t.CreateSender)

	updateSenderSchema := GenerateSchema[notifcommon.NotificationSenderRequestWithID](
		WithEnum("provider", providerEnum),
		WithRequired("id"),
	)

	mcp.AddTool(server, &mcp.Tool{
		Name: "update_notification_sender",
		Description: `Update an existing notification sender configuration.

Prerequisites:
- Obtain the current configuration using 'get_notification_sender'.

Inputs:
- id (required): The unique identifier of the sender to update.
- name (required): Display name.
- provider (required): Provider type (twilio, vonage, custom).
- properties (required): Full list of properties.
  Note: This is a full replacement. Include ALL current properties (and any new ones).

Outputs:
- Returns the updated sender configuration.

Next Steps:
- Verify the update by listing or getting the sender.`,
		InputSchema: updateSenderSchema,
		Annotations: &mcp.ToolAnnotations{
			Title:          "Update Notification Sender",
			IdempotentHint: true,
		},
	}, t.UpdateSender)

	mcp.AddTool(server, &mcp.Tool{
		Name: "delete_notification_sender",
		Description: `Delete a notification sender.

Inputs:
- id (required): The unique identifier of the sender to delete.

Outputs:
- Success message if deleted.

Next Steps:
- Ensure no flows are currently referencing this sender ID before deletion to avoid runtime errors.`,
		Annotations: &mcp.ToolAnnotations{
			Title:           "Delete Notification Sender",
			DestructiveHint: ptr(true),
		},
	}, t.DeleteSender)
}

// ListSenders handles the list_notification_senders tool call.
func (t *NotificationSenderTools) ListSenders(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input ListSendersInput,
) (*mcp.CallToolResult, SenderListOutput, error) {
	senders, svcErr := t.notifService.ListSenders()
	if svcErr != nil {
		return nil, SenderListOutput{}, fmt.Errorf("failed to list notification senders: %s", svcErr.ErrorDescription)
	}

	// Convert to response type for proper JSON serialization
	responses := make([]notifcommon.NotificationSenderResponse, 0, len(senders))
	for _, s := range senders {
		resp, err := convertDTOToResponse(&s)
		if err != nil {
			return nil, SenderListOutput{}, err
		}
		responses = append(responses, *resp)
	}

	return nil, SenderListOutput{
		TotalCount: len(responses),
		Senders:    responses,
	}, nil
}

// GetSender handles the get_notification_sender tool call.
func (t *NotificationSenderTools) GetSender(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input IDInput,
) (*mcp.CallToolResult, *notifcommon.NotificationSenderResponse, error) {

	sender, svcErr := t.notifService.GetSender(input.ID)
	if svcErr != nil {
		return nil, nil, fmt.Errorf("failed to get notification sender: %s", svcErr.ErrorDescription)
	}

	response, err := convertDTOToResponse(sender)
	if err != nil {
		return nil, nil, err
	}

	return nil, response, nil
}

// convertPropertiesToDomain converts PropertyDTO slice to Property slice.
func convertPropertiesToDomain(props []cmodels.PropertyDTO) ([]cmodels.Property, error) {
	properties := make([]cmodels.Property, 0, len(props))
	for _, p := range props {
		prop, err := cmodels.NewProperty(p.Name, p.Value, p.IsSecret)
		if err != nil {
			return nil, fmt.Errorf("failed to create property %s: %w", p.Name, err)
		}
		properties = append(properties, *prop)
	}
	return properties, nil
}

// convertDTOToResponse converts NotificationSenderDTO to NotificationSenderResponse.
// This is needed because Property has private fields and cannot be JSON-serialized.
func convertDTOToResponse(dto *notifcommon.NotificationSenderDTO) (*notifcommon.NotificationSenderResponse, error) {
	props := make([]cmodels.PropertyDTO, 0, len(dto.Properties))
	for _, p := range dto.Properties {
		propDTO, err := p.ToPropertyDTO()
		if err != nil {
			return nil, fmt.Errorf("failed to convert property: %w", err)
		}
		props = append(props, *propDTO)
	}

	return &notifcommon.NotificationSenderResponse{
		ID:          dto.ID,
		Name:        dto.Name,
		Description: dto.Description,
		Provider:    dto.Provider,
		Properties:  props,
	}, nil
}

// CreateSender handles the create_notification_sender tool call.
func (t *NotificationSenderTools) CreateSender(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input notifcommon.NotificationSenderRequest,
) (*mcp.CallToolResult, *notifcommon.NotificationSenderResponse, error) {

	// Convert properties
	properties, err := convertPropertiesToDomain(input.Properties)
	if err != nil {
		return nil, nil, err
	}

	// Convert input to DTO
	senderDTO := notifcommon.NotificationSenderDTO{
		Name:        input.Name,
		Description: input.Description,
		Provider:    notifcommon.MessageProviderType(input.Provider),
		Type:        notifcommon.NotificationSenderTypeMessage,
		Properties:  properties,
	}

	createdSender, svcErr := t.notifService.CreateSender(senderDTO)
	if svcErr != nil {
		return nil, nil, fmt.Errorf("failed to create notification sender: %s", svcErr.ErrorDescription)
	}

	response, err := convertDTOToResponse(createdSender)
	if err != nil {
		return nil, nil, err
	}

	return nil, response, nil
}

// UpdateSender handles the update_notification_sender tool call.
func (t *NotificationSenderTools) UpdateSender(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input notifcommon.NotificationSenderRequestWithID,
) (*mcp.CallToolResult, *notifcommon.NotificationSenderResponse, error) {

	// Verify existence
	_, svcErr := t.notifService.GetSender(input.ID)
	if svcErr != nil {
		return nil, nil, fmt.Errorf("failed to get notification sender: %s", svcErr.ErrorDescription)
	}

	// Convert properties
	properties, err := convertPropertiesToDomain(input.Properties)
	if err != nil {
		return nil, nil, err
	}

	// Convert input to DTO
	senderDTO := notifcommon.NotificationSenderDTO{
		ID:          input.ID,
		Name:        input.Name,
		Description: input.Description,
		Provider:    notifcommon.MessageProviderType(input.Provider),
		Type:        notifcommon.NotificationSenderTypeMessage,
		Properties:  properties,
	}

	updatedSender, svcErr := t.notifService.UpdateSender(input.ID, senderDTO)
	if svcErr != nil {
		return nil, nil, fmt.Errorf("failed to update notification sender: %s", svcErr.ErrorDescription)
	}

	response, err := convertDTOToResponse(updatedSender)
	if err != nil {
		return nil, nil, err
	}

	return nil, response, nil
}

// DeleteSender handles the delete_notification_sender tool call.
func (t *NotificationSenderTools) DeleteSender(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input IDInput,
) (*mcp.CallToolResult, DeleteOutput, error) {

	svcErr := t.notifService.DeleteSender(input.ID)
	if svcErr != nil {
		return nil, DeleteOutput{
			Success: false,
			Message: fmt.Sprintf("Failed to delete notification sender: %s", svcErr.ErrorDescription),
		}, nil
	}

	return nil, DeleteOutput{
		Success: true,
		Message: fmt.Sprintf("Notification sender %s deleted successfully", input.ID),
	}, nil
}
