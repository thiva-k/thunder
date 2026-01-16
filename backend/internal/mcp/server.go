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
	"github.com/asgardeo/thunder/internal/application"
	flowmgt "github.com/asgardeo/thunder/internal/flow/mgt"
	"github.com/asgardeo/thunder/internal/mcp/tools"
	"github.com/asgardeo/thunder/internal/notification"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Server wraps the MCP server and its tool registrations.
type Server struct {
	mcpServer *mcp.Server
}

// NewServer creates a new MCP server with all tools registered.
func NewServer(
	appService application.ApplicationServiceInterface,
	flowService flowmgt.FlowMgtServiceInterface,
	notifService notification.NotificationSenderMgtSvcInterface,
) *Server {
	// Create the MCP server instance.
	mcpServer := mcp.NewServer(&mcp.Implementation{
		Name:    "thunder-mcp",
		Version: "1.0.0",
	}, nil)

	// Register application tools.
	appTools := tools.NewApplicationTools(appService)
	appTools.RegisterTools(mcpServer)

	// Register flow tools.
	flowTools := tools.NewFlowTools(flowService)
	flowTools.RegisterTools(mcpServer)

	// Register notification sender tools.
	notifTools := tools.NewNotificationSenderTools(notifService)
	notifTools.RegisterTools(mcpServer)

	// Register React SDK integration tools.
	reactTools := tools.NewReactSDKTools()
	reactTools.RegisterTools(mcpServer)

	return &Server{
		mcpServer: mcpServer,
	}
}

// GetMCPServer returns the underlying MCP server instance.
func (s *Server) GetMCPServer() *mcp.Server {
	return s.mcpServer
}
