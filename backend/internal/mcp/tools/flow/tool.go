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

// Package flow provides flow management tool models.
//
//nolint:lll
package flow

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/asgardeo/thunder/internal/flow/common"
	flowmgt "github.com/asgardeo/thunder/internal/flow/mgt"
	"github.com/asgardeo/thunder/internal/mcp/tools"
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

// RegisterTools registers all flow tools with the MCP server.
func (t *FlowTools) RegisterTools(server *mcp.Server) {
	listFlowsSchema := tools.GenerateSchema[ListFlowsInput](
		tools.WithEnum("", "flow_type", []string{"AUTHENTICATION", "REGISTRATION"}),
		tools.WithDefaults(map[string]any{"limit": 30, "offset": 0}),
	)

	getFlowByHandleSchema := tools.GenerateSchema[GetFlowByHandleInput](
		tools.WithEnum("", "flow_type", []string{"AUTHENTICATION", "REGISTRATION"}),
		tools.WithRequired("handle", "flow_type"),
	)

	createFlowSchema := tools.GenerateSchema[flowmgt.FlowDefinition](
		tools.WithEnum("", "flowType", []string{string(common.FlowTypeAuthentication), string(common.FlowTypeRegistration)}),
	)

	updateFlowSchema := tools.GenerateSchema[UpdateFlowInput](
		tools.WithRequired("id", "name", "nodes"),
	)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_flows",
		Description: `List available flows. Supports optional filtering by flow_type.`,
		InputSchema: listFlowsSchema,
		Annotations: &mcp.ToolAnnotations{
			Title:        "List Flows",
			ReadOnlyHint: true,
		},
	}, t.ListFlows)

	mcp.AddTool(server, &mcp.Tool{
		Name: "get_flow_by_id",
		Description: `Retrieve the complete definition of a flow by ID including nodes, executors, and UI metadata.
Use Cases:
- Understand flow structure before creating similar flows
- Review current configuration before updating
- Reference node patterns (e.g. PROMPT nodes with components, TASK_EXECUTION nodes with executors)`,
		Annotations: &mcp.ToolAnnotations{
			Title:        "Get Flow by ID",
			ReadOnlyHint: true,
		},
	}, t.GetFlowByID)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_flow_by_handle",
		Description: `Retrieve a complete definition of a flow by its handle (human-readable identifier).`,
		InputSchema: getFlowByHandleSchema,
		Annotations: &mcp.ToolAnnotations{
			Title:        "Get Flow by Handle",
			ReadOnlyHint: true,
		},
	}, t.GetFlowByHandle)

	mcp.AddTool(server, &mcp.Tool{
		Name: "create_flow",
		Description: `Create a new authentication or registration flow.

Prerequisites:
- For SMS/Email OTP: Create notification sender first using create_notification_sender
- Review existing similar flows by get_flow_by_handle to understand patterns and node structures

Key Requirements:
- Handle: Lowercase alphanumeric, dashes/underscores allowed (not at start/end). Unique per flow type.
- Structure: Must include START and END nodes with at least one functional node in between.
- Node types: START, END, TASK_EXECUTION, PROMPT.
- PROMPT nodes: Require 'meta.components' array for UI rendering.
- Transitions: Use onSuccess/onFailure node IDs to define the path.`,
		InputSchema: createFlowSchema,
		Annotations: &mcp.ToolAnnotations{
			Title:          "Create Flow",
			IdempotentHint: false,
		},
	}, t.CreateFlow)

	mcp.AddTool(server, &mcp.Tool{
		Name: "update_flow",
		Description: `Update an existing flow definition (full replacement for updateable fields).

This is a PUT operation for updateable fields. Use get_flow_by_id first to get current state.

Workflow:
1. Use get_flow_by_id to get current flow
2. Modify name and/or nodes as needed
3. Send the complete update back

Flow versions are automatically tracked. Updating creates a new version.`,
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

// GetFlowByID handles the get_flow_by_id tool call.
func (t *FlowTools) GetFlowByID(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input tools.IDInput,
) (*mcp.CallToolResult, *flowmgt.CompleteFlowDefinition, error) {
	flow, svcErr := t.flowService.GetFlow(input.ID)
	if svcErr != nil {
		return nil, nil, fmt.Errorf("failed to get flow: %s", svcErr.ErrorDescription)
	}

	return nil, flow, nil
}

// GetFlowByHandle handles the get_flow_by_handle tool call.
func (t *FlowTools) GetFlowByHandle(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input GetFlowByHandleInput,
) (*mcp.CallToolResult, *flowmgt.CompleteFlowDefinition, error) {
	flowType := common.FlowType(input.FlowType)

	flow, svcErr := t.flowService.GetFlowByHandle(input.Handle, flowType)
	if svcErr != nil {
		return nil, nil, fmt.Errorf("failed to get flow by handle: %s", svcErr.ErrorDescription)
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

// UpdateFlow handles the update_flow tool call.
func (t *FlowTools) UpdateFlow(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input UpdateFlowInput,
) (*mcp.CallToolResult, *flowmgt.CompleteFlowDefinition, error) {
	// Get current flow to retrieve immutable fields (handle, flowType)
	currentFlow, svcErr := t.flowService.GetFlow(input.ID)
	if svcErr != nil {
		return nil, nil, fmt.Errorf("failed to get flow: %s", svcErr.ErrorDescription)
	}

	// Build update definition with immutable fields preserved and input fields replaced
	updateDef := &flowmgt.FlowDefinition{
		Handle:   currentFlow.Handle,
		FlowType: currentFlow.FlowType,
		Name:     input.Name,
		Nodes:    input.Nodes,
	}

	updatedFlow, svcErr := t.flowService.UpdateFlow(input.ID, updateDef)
	if svcErr != nil {
		return nil, nil, fmt.Errorf("failed to update flow: %s", svcErr.ErrorDescription)
	}

	return nil, updatedFlow, nil
}
