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
	FlowType string `json:"flow_type,omitempty" jsonschema:"Filter by flow type: 'AUTHENTICATION' for login flows or 'REGISTRATION' for signup flows. Omit to see all flows."`
}

// FlowListOutput represents the output for list_flows tool.
type FlowListOutput struct {
	TotalCount int                           `json:"total_count"`
	Flows      []flowmgt.BasicFlowDefinition `json:"flows"`
}

// GetFlowByHandleInput represents the input for get_flow_by_handle tool.
type GetFlowByHandleInput struct {
	Handle   string `json:"handle" jsonschema:"Flow handle to search for. Example: 'basic-login', 'invite-registration'"`
	FlowType string `json:"flow_type" jsonschema:"Flow type: 'AUTHENTICATION' or 'REGISTRATION'. Required to uniquely identify the flow."`
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

Related: 
- Use returned 'id' with get_flow_by_id to see full flow definition
- Use returned 'handle' with get_flow_by_handle for easier lookup
- Use returned 'id' with update_flow to modify flows
- Assign flow 'id' to applications via create_application or update_application`,
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
- Reference node patterns (PROMPT nodes with components, TASK_EXECUTION nodes with executors)

Related: Use before update_flow or as reference when creating new flows.`,
		Annotations: &mcp.ToolAnnotations{
			Title:        "Get Flow by ID",
			ReadOnlyHint: true,
		},
	}, t.GetFlowByID)

	getFlowByHandleSchema := GenerateSchema[GetFlowByHandleInput](
		WithEnum("flow_type", []string{"AUTHENTICATION", "REGISTRATION"}),
		WithRequired("handle", "flow_type"),
	)

	mcp.AddTool(server, &mcp.Tool{
		Name: "get_flow_by_handle",
		Description: `Retrieve a flow by its handle (human-readable identifier).

More intuitive than get_flow_by_id when you know the flow's handle.

Common handles:
- 'basic-login' - Standard username/password authentication
- 'invite-registration' - Invite-based user registration
- 'otp-login' - Phone/Email OTP authentication

Related: Alternative to get_flow_by_id when you know the handle but not the ID.`,
		InputSchema: getFlowByHandleSchema,
		Annotations: &mcp.ToolAnnotations{
			Title:        "Get Flow by Handle",
			ReadOnlyHint: true,
		},
	}, t.GetFlowByHandle)

	creatFlowSchema := GenerateSchema[flowmgt.FlowDefinition](
		WithEnum("flowType", []string{string(common.FlowTypeAuthentication), string(common.FlowTypeRegistration)}),
	)

	mcp.AddTool(server, &mcp.Tool{
		Name: "create_flow",
		Description: `Create a new authentication or registration flow.

Use get_flow_by_handle on existing flows (like 'basic-login') to see node structure examples.

Prerequisites:
- For SMS/Email OTP: Create notification sender first using create_notification_sender
- Review existing flows using list_flows and get_flow_by_id to understand patterns

Key Requirements:
- Handle must be unique per flow type (lowercase, alphanumeric with dashes/underscores)
- Must include START node (entry point) and END node (exit point)
- Node types: START, END, TASK_EXECUTION (backend logic), PROMPT (user input)
- PROMPT nodes require 'meta.components' array for UI rendering
- Each node needs onSuccess/onFailure to define flow path

Common Patterns:
- Login: START → PROMPT (username/password) → TASK_EXECUTION (authenticate) → END
- OTP: START → PROMPT (phone) → TASK_EXECUTION (send OTP) → PROMPT (verify) → TASK_EXECUTION (validate) → END

Related: 
- Use get_flow_by_id or get_flow_by_handle to see examples of node structures
- Assign created flow to applications using create_application or update_application`,
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
		Description: `Update an existing flow definition with PATCH semantics.

PATCH semantics: Only provided fields are updated. Omitted fields are preserved.

Updatable fields:
- name: Display name
- nodes: Complete node array (full replacement for nodes)

Prerequisites: Use get_flow first to retrieve current structure.

Important Notes:
- Flow versions are automatically tracked (up to 50 versions retained)
- Updating creates a new version and sets it as active
- Cannot change handle or flowType after creation

Tip: Use get_flow to see current configuration before updating.`,
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
	input IDInput,
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

// UpdateFlowInput represents the input for update_flow tool.
type UpdateFlowInput struct {
	ID    string                   `json:"id" jsonschema:"The unique identifier of the flow to update. Required."`
	Name  string                   `json:"name,omitempty" jsonschema:"Display name for the flow. Optional. Provide only if updating."`
	Nodes []flowmgt.NodeDefinition `json:"nodes,omitempty" jsonschema:"Array of nodes defining the flow steps. Optional. Provide complete node list if updating flow structure."`
}

// UpdateFlow handles the update_flow tool call with PATCH semantics.
// Only provided (non-empty) fields are updated; omitted fields are preserved.
func (t *FlowTools) UpdateFlow(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input UpdateFlowInput,
) (*mcp.CallToolResult, *flowmgt.CompleteFlowDefinition, error) {

	// Get current flow
	currentFlow, svcErr := t.flowService.GetFlow(input.ID)
	if svcErr != nil {
		return nil, nil, fmt.Errorf("failed to get flow: %s", svcErr.ErrorDescription)
	}

	// Build update definition starting with current values
	updateDef := &flowmgt.FlowDefinition{
		Handle:   currentFlow.Handle, // Immutable
		Name:     currentFlow.Name,
		FlowType: currentFlow.FlowType, // Immutable
		Nodes:    currentFlow.Nodes,
	}

	// Apply updates only for provided fields (PATCH semantics)
	if input.Name != "" {
		updateDef.Name = input.Name
	}
	if len(input.Nodes) > 0 {
		updateDef.Nodes = input.Nodes
	}

	updatedFlow, svcErr := t.flowService.UpdateFlow(input.ID, updateDef)
	if svcErr != nil {
		return nil, nil, fmt.Errorf("failed to update flow: %s", svcErr.ErrorDescription)
	}

	return nil, updatedFlow, nil
}
