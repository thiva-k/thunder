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

package mcp

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/asgardeo/thunder/internal/application"
	flowmgt "github.com/asgardeo/thunder/internal/flow/mgt"
	apptools "github.com/asgardeo/thunder/internal/mcp/tools/application"
	flowtools "github.com/asgardeo/thunder/internal/mcp/tools/flow"
)

// Server wraps the MCP server and its tool registrations.
type server struct {
	mcpServer *mcp.Server
}

// NewServer creates a new MCP server with application tools registered.
func newServer(
	appService application.ApplicationServiceInterface,
	flowService flowmgt.FlowMgtServiceInterface,
) *server {
	// Create the MCP server instance.
	mcpServer := mcp.NewServer(&mcp.Implementation{
		Name:    "thunder-mcp",
		Version: "1.0.0",
	}, nil)

	// Register application tools.
	apptools.NewApplicationTools(appService).RegisterTools(mcpServer)

	// Register flow tools.
	flowtools.NewFlowTools(flowService).RegisterTools(mcpServer)

	return &server{
		mcpServer: mcpServer,
	}
}

// GetMCPServer returns the underlying MCP server instance.
func (s *server) getMCPServer() *mcp.Server {
	return s.mcpServer
}
