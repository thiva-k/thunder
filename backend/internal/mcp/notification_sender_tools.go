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

// SenderIDInput represents an input that requires only a sender ID.
type SenderIDInput struct {
	ID string `json:"id" jsonschema:"The unique identifier of the notification sender"`
}

// SenderListOutput represents the output for list_notification_senders tool.
type SenderListOutput struct {
	TotalCount int                                 `json:"total_count"`
	Senders    []notifcommon.NotificationSenderDTO `json:"senders"`
}

// RegisterTools registers all notification sender tools with the MCP server.
func (t *NotificationSenderTools) RegisterTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_notification_senders",
		Description: "List all configured notification senders (SMS, Email providers)",
		Annotations: &mcp.ToolAnnotations{
			Title:        "List Notification Senders",
			ReadOnlyHint: true,
		},
	}, t.ListSenders)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_notification_sender",
		Description: "Get detailed information about a specific notification sender by its ID",
		Annotations: &mcp.ToolAnnotations{
			Title:        "Get Notification Sender",
			ReadOnlyHint: true,
		},
	}, t.GetSender)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_notification_sender",
		Description: "Create a new notification sender (SMS/message provider). Supports providers: twilio, vonage, or custom webhook. The returned ID can be used in flow nodes with SMSOTPAuthExecutor.",
		InputSchema: generateNotificationSenderRequestSchema(),
		Annotations: &mcp.ToolAnnotations{
			Title:          "Create Notification Sender",
			IdempotentHint: true,
		},
	}, t.CreateSender)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_notification_sender",
		Description: "Update an existing notification sender configuration",
		InputSchema: generateNotificationSenderUpdateSchema(),
		Annotations: &mcp.ToolAnnotations{
			Title:          "Update Notification Sender",
			IdempotentHint: true,
		},
	}, t.UpdateSender)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_notification_sender",
		Description: "Delete a notification sender. This action is irreversible.",
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

	return nil, SenderListOutput{
		TotalCount: len(senders),
		Senders:    senders,
	}, nil
}

// GetSender handles the get_notification_sender tool call.
func (t *NotificationSenderTools) GetSender(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input SenderIDInput,
) (*mcp.CallToolResult, *notifcommon.NotificationSenderDTO, error) {
	if input.ID == "" {
		return nil, nil, fmt.Errorf("notification sender ID is required")
	}

	sender, svcErr := t.notifService.GetSender(input.ID)
	if svcErr != nil {
		return nil, nil, fmt.Errorf("failed to get notification sender: %s", svcErr.ErrorDescription)
	}

	return nil, sender, nil
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

// CreateSender handles the create_notification_sender tool call.
func (t *NotificationSenderTools) CreateSender(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input notifcommon.NotificationSenderRequest,
) (*mcp.CallToolResult, *notifcommon.NotificationSenderDTO, error) {
	if input.Name == "" {
		return nil, nil, fmt.Errorf("notification sender name is required")
	}
	if input.Provider == "" {
		return nil, nil, fmt.Errorf("provider is required (twilio, vonage, or custom)")
	}

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

	return nil, createdSender, nil
}

// UpdateSender handles the update_notification_sender tool call.
func (t *NotificationSenderTools) UpdateSender(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input notifcommon.NotificationSenderRequestWithID,
) (*mcp.CallToolResult, *notifcommon.NotificationSenderDTO, error) {
	if input.ID == "" {
		return nil, nil, fmt.Errorf("notification sender ID is required for update")
	}

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

	return nil, updatedSender, nil
}

// DeleteSender handles the delete_notification_sender tool call.
func (t *NotificationSenderTools) DeleteSender(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input SenderIDInput,
) (*mcp.CallToolResult, DeleteOutput, error) {
	if input.ID == "" {
		return nil, DeleteOutput{}, fmt.Errorf("notification sender ID is required")
	}

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
