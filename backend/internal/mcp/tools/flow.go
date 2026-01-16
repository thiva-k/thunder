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

	"github.com/asgardeo/thunder/internal/flow/common"
	flowmgt "github.com/asgardeo/thunder/internal/flow/mgt"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// FlowTools provides MCP tools for managing Thunder authentication flows.
type FlowTools struct {
	flowService flowmgt.FlowMgtServiceInterface
}

// NewFlowTools creates a new FlowTools instance.
func NewFlowTools(flowService flowmgt.FlowMgtServiceInterface) *FlowTools {
	return &FlowTools{
		flowService: flowService,
	}
}

// ListFlowsInput represents the input for the list_flows tool.
type ListFlowsInput struct {
	PaginationInput
	FlowType string `json:"flow_type,omitempty" jsonschema:"Filter by flow type: AUTHENTICATION or REGISTRATION"`
}

// FlowListOutput represents the output for list_flows tool.
type FlowListOutput struct {
	TotalCount int                           `json:"total_count"`
	Flows      []flowmgt.BasicFlowDefinition `json:"flows"`
}

// RegisterTools registers all flow tools with the MCP server.
func (t *FlowTools) RegisterTools(server *mcp.Server) {
	listFlowsSchema := GenerateSchema[ListFlowsInput](
		WithEnum("flow_type", []string{"AUTHENTICATION", "REGISTRATION"}),
		WithDefaults(map[string]any{"limit": 30, "offset": 0}),
	)

	mcp.AddTool(server, &mcp.Tool{
		Name: "list_flows",
		Description: `List all available authentication and registration flows.

Behavior: Returns paginated results. Filter by flow_type to see only login or signup flows.

Related: Use returned 'id' with get_flow, update_flow, or assign to applications.`,
		InputSchema: listFlowsSchema,
		Annotations: &mcp.ToolAnnotations{
			Title:        "List Flows",
			ReadOnlyHint: true,
		},
	}, t.ListFlows)

	mcp.AddTool(server, &mcp.Tool{
		Name: "get_flow",
		Description: `Retrieve the complete definition of a flow including nodes, executors, and UI metadata.

Related: Use before update_flow to understand current structure. Reference existing flows when creating new ones.`,
		Annotations: &mcp.ToolAnnotations{
			Title:        "Get Flow",
			ReadOnlyHint: true,
		},
	}, t.GetFlow)

	creatFlowSchema := GenerateSchema[flowmgt.FlowDefinition](
		WithEnum("flowType", []string{string(common.FlowTypeAuthentication), string(common.FlowTypeRegistration)}),
	)

	mcp.AddTool(server, &mcp.Tool{
		Name: "create_flow",
		Description: `Create a new authentication or registration flow.

Prerequisites:
- For SMS/Email OTP: Create notification sender first using create_notification_sender.
- Reference existing flows using get_flow to understand node structure.

Behavior:
- Handle must be unique per flow type (lowercase, alphanumeric with dashes/underscores).
- Node types: START, END, TASK_EXECUTION, PROMPT.
- PROMPT nodes require 'meta.components' for UI rendering.

Related: Assign flow to application using create_application or update_application.`,
		InputSchema: creatFlowSchema,
		Annotations: &mcp.ToolAnnotations{
			Title:          "Create Flow",
			IdempotentHint: true,
		},
	}, t.CreateFlow)

	updateFlowSchema := GenerateSchema[UpdateFlowInput](
		WithEnum("flow_type", []string{string(common.FlowTypeAuthentication), string(common.FlowTypeRegistration)}),
		WithRequired("id"),
	)

	mcp.AddTool(server, &mcp.Tool{
		Name: "update_flow",
		Description: `Update an existing flow definition.

Prerequisites: Use get_flow first to retrieve current structure.

IMPORTANT: This is a full replacement. Provide complete node array.
Flow versions are automatically tracked (up to 50 versions retained).`,
		InputSchema: updateFlowSchema,
		Annotations: &mcp.ToolAnnotations{
			Title:          "Update Flow",
			IdempotentHint: true,
		},
	}, t.UpdateFlow)
}

// ListFlows handles the list_flows tool call.
func (t *FlowTools) ListFlows(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input ListFlowsInput,
) (*mcp.CallToolResult, FlowListOutput, error) {
	limit := input.Limit
	if limit <= 0 {
		limit = 30
	}

	flowType := common.FlowType(input.FlowType)

	listResponse, svcErr := t.flowService.ListFlows(limit, input.Offset, flowType)
	if svcErr != nil {
		return nil, FlowListOutput{}, fmt.Errorf("failed to list flows: %s", svcErr.ErrorDescription)
	}

	return nil, FlowListOutput{
		TotalCount: listResponse.TotalResults,
		Flows:      listResponse.Flows,
	}, nil
}

// GetFlow handles the get_flow tool call.
func (t *FlowTools) GetFlow(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input IDInput,
) (*mcp.CallToolResult, *flowmgt.CompleteFlowDefinition, error) {

	flow, svcErr := t.flowService.GetFlow(input.ID)
	if svcErr != nil {
		return nil, nil, fmt.Errorf("failed to get flow: %s", svcErr.ErrorDescription)
	}

	return nil, flow, nil
}

// CreateFlow handles the create_flow tool call.
func (t *FlowTools) CreateFlow(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input flowmgt.FlowDefinition,
) (*mcp.CallToolResult, *flowmgt.CompleteFlowDefinition, error) {
	createdFlow, svcErr := t.flowService.CreateFlow(&input)
	if svcErr != nil {
		return nil, nil, fmt.Errorf("failed to create flow: %s", svcErr.ErrorDescription)
	}

	return nil, createdFlow, nil
}

// UpdateFlowInput represents the input for update_flow tool.
type UpdateFlowInput struct {
	ID string `json:"id" jsonschema:"The unique identifier of the flow to update"`
	flowmgt.FlowDefinition
}

// UpdateFlow handles the update_flow tool call.
func (t *FlowTools) UpdateFlow(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input UpdateFlowInput,
) (*mcp.CallToolResult, *flowmgt.CompleteFlowDefinition, error) {

	// Verify existence
	_, svcErr := t.flowService.GetFlow(input.ID)
	if svcErr != nil {
		return nil, nil, fmt.Errorf("failed to get flow: %s", svcErr.ErrorDescription)
	}

	updatedFlow, svcErr := t.flowService.UpdateFlow(input.ID, &input.FlowDefinition)
	if svcErr != nil {
		return nil, nil, fmt.Errorf("failed to update flow: %s", svcErr.ErrorDescription)
	}

	return nil, updatedFlow, nil
}
