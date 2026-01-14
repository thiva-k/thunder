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
	FlowType string `json:"flow_type,omitempty" jsonschema:"Filter by flow type: AUTHENTICATION or REGISTRATION"`
	Limit    int    `json:"limit,omitempty" jsonschema:"Maximum number of flows to return (default: 20)"`
	Offset   int    `json:"offset,omitempty" jsonschema:"Offset for pagination (default: 0)"`
}

// FlowIDInput represents an input that requires only a flow ID.
type FlowIDInput struct {
	ID string `json:"id" jsonschema:"The unique identifier of the flow"`
}

// FlowListOutput represents the output for list_flows tool.
type FlowListOutput struct {
	TotalCount int                           `json:"total_count"`
	Flows      []flowmgt.BasicFlowDefinition `json:"flows"`
}

// RegisterTools registers all flow tools with the MCP server.
func (t *FlowTools) RegisterTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_flows",
		Description: "List all authentication and registration flows in Thunder",
		Annotations: &mcp.ToolAnnotations{
			Title:        "List Flows",
			ReadOnlyHint: true,
		},
	}, t.ListFlows)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_flow",
		Description: "Get detailed information about a specific flow by its ID, including all nodes and configuration",
		Annotations: &mcp.ToolAnnotations{
			Title:        "Get Flow",
			ReadOnlyHint: true,
		},
	}, t.GetFlow)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_flow",
		Description: "Create a new authentication or registration flow with nodes defining the flow logic. Nodes can include authenticators like BasicAuthenticator, SMSOTPAuthenticator, etc.",
		Annotations: &mcp.ToolAnnotations{
			Title:          "Create Flow",
			IdempotentHint: true,
		},
	}, t.CreateFlow)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_flow",
		Description: "Update an existing flow. Requires the flow ID. Updates create a new version of the flow.",
		Annotations: &mcp.ToolAnnotations{
			Title:          "Update Flow",
			IdempotentHint: true,
		},
	}, t.UpdateFlow)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_flow",
		Description: "Delete a flow from Thunder. This action is irreversible.",
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
	input FlowIDInput,
) (*mcp.CallToolResult, *flowmgt.CompleteFlowDefinition, error) {
	if input.ID == "" {
		return nil, nil, fmt.Errorf("flow ID is required")
	}

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
	if input.Handle == "" {
		return nil, nil, fmt.Errorf("flow handle is required")
	}
	if input.Name == "" {
		return nil, nil, fmt.Errorf("flow name is required")
	}
	if input.FlowType == "" {
		return nil, nil, fmt.Errorf("flow type is required (AUTHENTICATION or REGISTRATION)")
	}
	if len(input.Nodes) == 0 {
		return nil, nil, fmt.Errorf("at least one node is required")
	}

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
	if input.ID == "" {
		return nil, nil, fmt.Errorf("flow ID is required for update")
	}

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
	input FlowIDInput,
) (*mcp.CallToolResult, DeleteOutput, error) {
	if input.ID == "" {
		return nil, DeleteOutput{}, fmt.Errorf("flow ID is required")
	}

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
