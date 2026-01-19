/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
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
	flowmgt "github.com/asgardeo/thunder/internal/flow/mgt"
	"github.com/asgardeo/thunder/internal/mcp/tools/common"
)

// listFlowsInput represents the input for the list_flows tool.
type listFlowsInput struct {
	common.PaginationInput
	FlowType string `json:"flow_type,omitempty" jsonschema:"Filter by flow type: 'AUTHENTICATION' for login flows or 'REGISTRATION' for signup flows. Omit to see all flows."`
}

// flowListOutput represents the output for list_flows tool.
type flowListOutput struct {
	TotalCount int                           `json:"total_count" jsonschema:"Total number of flows available."`
	Flows      []flowmgt.BasicFlowDefinition `json:"flows" jsonschema:"List of flow definitions."`
}

// getFlowByHandleInput represents the input for get_flow_by_handle tool.
type getFlowByHandleInput struct {
	Handle   string `json:"handle" jsonschema:"Flow handle to search for."`
	FlowType string `json:"flow_type" jsonschema:"Flow type: 'AUTHENTICATION' or 'REGISTRATION'. Required to uniquely identify the flow."`
}

// updateFlowInput represents the input for update_flow tool.
type updateFlowInput struct {
	ID    string                   `json:"id" jsonschema:"The unique identifier of the flow to update. Required."`
	Name  string                   `json:"name" jsonschema:"Display name for the flow. Required for PUT."`
	Nodes []flowmgt.NodeDefinition `json:"nodes" jsonschema:"Array of nodes defining the flow steps. Required for PUT."`
}
