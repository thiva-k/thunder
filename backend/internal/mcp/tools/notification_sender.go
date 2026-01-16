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

// SenderListOutput represents the output for list_notification_senders tool.
type SenderListOutput struct {
	TotalCount int                                      `json:"total_count"`
	Senders    []notifcommon.NotificationSenderResponse `json:"senders"`
}

// RegisterTools registers all notification sender tools with the MCP server.
func (t *NotificationSenderTools) RegisterTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name: "list_notification_senders",
		Description: `List all configured notification senders (SMS/Email providers).

Related: Use returned 'id' with get_notification_sender for details, or in flow executor configurations.`,
		Annotations: &mcp.ToolAnnotations{
			Title:        "List Notification Senders",
			ReadOnlyHint: true,
		},
	}, t.ListSenders)

	mcp.AddTool(server, &mcp.Tool{
		Name: "get_notification_sender",
		Description: `Get detailed configuration of a notification sender (secrets are masked).

Related: Use before update_notification_sender to review current configuration.`,
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

Prerequisites: None - this is typically the first step before creating OTP-based flows.

Provider-specific properties:
- twilio: account_sid, auth_token (secret), from_number
- vonage: api_key, api_secret (secret), from_number  
- custom: url, http_method, content_type

Related: Use returned 'id' as 'senderId' property in SMSOTPAuthExecutor flow nodes.
Prompt for any user credentials if not provided. Do not make assumptions about the values.`,
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

Prerequisites: Use get_notification_sender first to retrieve current properties.

IMPORTANT: This is a full replacement.
Include ALL properties including unchanged ones. Missing properties will be removed.`,
		InputSchema: updateSenderSchema,
		Annotations: &mcp.ToolAnnotations{
			Title:          "Update Notification Sender",
			IdempotentHint: true,
		},
	}, t.UpdateSender)

	mcp.AddTool(server, &mcp.Tool{
		Name: "delete_notification_sender",
		Description: `Permanently delete a notification sender.

Prerequisites: Ensure no flows reference this sender ID to avoid runtime errors.

Impact: Flows using this sender will fail to send messages after deletion.`,
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
	_ any,
) (*mcp.CallToolResult, SenderListOutput, error) {
	senders, svcErr := t.notifService.ListSenders()
	if svcErr != nil {
		return nil, SenderListOutput{}, fmt.Errorf("failed to list notification senders: %s", svcErr.ErrorDescription)
	}

	// Convert to response type for proper JSON serialization
	responses := make([]notifcommon.NotificationSenderResponse, 0, len(senders))
	for _, s := range senders {
		resp, err := senderDTOToResponse(&s)
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

	response, err := senderDTOToResponse(sender)
	if err != nil {
		return nil, nil, err
	}

	return nil, response, nil
}

// CreateSender handles the create_notification_sender tool call.
func (t *NotificationSenderTools) CreateSender(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input notifcommon.NotificationSenderRequest,
) (*mcp.CallToolResult, *notifcommon.NotificationSenderResponse, error) {

	// Convert properties
	properties, err := propertyDTOsToProperties(input.Properties)
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

	response, err := senderDTOToResponse(createdSender)
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
	properties, err := propertyDTOsToProperties(input.Properties)
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

	response, err := senderDTOToResponse(updatedSender)
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

// propertyDTOsToProperties converts a slice of PropertyDTO to domain Property objects.
func propertyDTOsToProperties(props []cmodels.PropertyDTO) ([]cmodels.Property, error) {
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

// senderDTOToResponse converts NotificationSenderDTO to NotificationSenderResponse.
func senderDTOToResponse(dto *notifcommon.NotificationSenderDTO) (*notifcommon.NotificationSenderResponse, error) {
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
