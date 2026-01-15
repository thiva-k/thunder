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

	"github.com/asgardeo/thunder/internal/flow/common"
	flowmgt "github.com/asgardeo/thunder/internal/flow/mgt"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var flowTypeValues = []string{string(common.FlowTypeAuthentication), string(common.FlowTypeRegistration)}

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
		WithDefaults(map[string]any{"limit": 20, "offset": 0}),
	)

	mcp.AddTool(server, &mcp.Tool{
		Name: "list_flows",
		Description: `List all available flows.

Inputs:
- flow_type (Enum): "AUTHENTICATION" (Login) or "REGISTRATION" (Signup).
- limit/offset: For pagination.

Outputs: List of flows with ID, Name, and Type.`,
		InputSchema: listFlowsSchema,
		Annotations: &mcp.ToolAnnotations{
			Title:        "List Flows",
			ReadOnlyHint: true,
		},
	}, t.ListFlows)

	mcp.AddTool(server, &mcp.Tool{
		Name: "get_flow",
		Description: `Retrieve the complete definition of a flow.

Inputs: 'id' (UUID).
Outputs: Full flow structure including all Nodes, Executors, and UI metadata.`,
		Annotations: &mcp.ToolAnnotations{
			Title:        "Get Flow",
			ReadOnlyHint: true,
		},
	}, t.GetFlow)

	creatFlowSchema := GenerateSchema[flowmgt.FlowDefinition](
		WithEnum("flowType", flowTypeValues), // FlowDefinition uses json:"flowType"
	)

	mcp.AddTool(server, &mcp.Tool{
		Name: "create_flow",
		Description: `Create a new authentication or registration flow.

Refer similar existing flows if needed. Add meta field if needed based on this.

Prerequisites:
- Notification Senders: Required if using SMS/Email OTP executors.

Inputs:
- name (Required): Unique flow name.
- flow_type (Enum): "AUTHENTICATION", "REGISTRATION".
- nodes: Array of flow nodes (START, END, TASK_EXECUTION, PROMPT, DECISION).

Node Types:
- TASK_EXECUTION: Runs backend logic (e.g., "SMSOTPAuthExecutor").
- PROMPT: Renders UI to user (requires 'meta.components').
- DECISION: Branching logic.

Outputs: Created flow definition with ID.

Next Steps: Assign this flow to an Application via create_application or update_application.`,
		InputSchema: creatFlowSchema,
		Annotations: &mcp.ToolAnnotations{
			Title:          "Create Flow",
			IdempotentHint: true,
		},
	}, t.CreateFlow)

	updateFlowSchema := GenerateSchema[UpdateFlowInput](
		WithEnum("flow_type", flowTypeValues), // UpdateFlowInput uses json:"flow_type"
		WithRequired("id"),
	)

	mcp.AddTool(server, &mcp.Tool{
		Name: "update_flow",
		Description: `Update an existing flow definition.

Inputs:
- id (Required): Flow UUID.
- Complete flow definition (replaces existingNodes).

Outputs: Updated flow definition.`,
		InputSchema: updateFlowSchema,
		Annotations: &mcp.ToolAnnotations{
			Title:          "Update Flow",
			IdempotentHint: true,
		},
	}, t.UpdateFlow)

	mcp.AddTool(server, &mcp.Tool{
		Name: "delete_flow",
		Description: `Permanently delete a flow.

Inputs: 'id' (UUID).
Prerequisites: Ensure this flow is not assigned to any Application.`,
		Annotations: &mcp.ToolAnnotations{
			Title:           "Delete Flow",
			DestructiveHint: ptr(true),
		},
	}, t.DeleteFlow)
}

// ListFlows handles the list_flows tool call.
func (t *FlowTools) ListFlows(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input ListFlowsInput,
) (*mcp.CallToolResult, FlowListOutput, error) {
	limit := input.Limit
	if limit <= 0 {
		limit = 20
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

// DeleteFlow handles the delete_flow tool call.
func (t *FlowTools) DeleteFlow(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input IDInput,
) (*mcp.CallToolResult, DeleteOutput, error) {

	svcErr := t.flowService.DeleteFlow(input.ID)
	if svcErr != nil {
		return nil, DeleteOutput{
			Success: false,
			Message: fmt.Sprintf("Failed to delete flow: %s", svcErr.ErrorDescription),
		}, nil
	}

	return nil, DeleteOutput{
		Success: true,
		Message: fmt.Sprintf("Flow %s deleted successfully", input.ID),
	}, nil
}
