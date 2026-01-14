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

// Package mcp provides MCP (Model Context Protocol) server functionality for Thunder.
package mcp

import (
	"net/http"

	"github.com/asgardeo/thunder/internal/application"
	flowmgt "github.com/asgardeo/thunder/internal/flow/mgt"
	"github.com/asgardeo/thunder/internal/notification"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// Initialize initializes the MCP server and registers its routes with the provided mux.
func Initialize(
	mux *http.ServeMux,
	appService application.ApplicationServiceInterface,
	flowService flowmgt.FlowMgtServiceInterface,
	notifService notification.NotificationSenderMgtSvcInterface,
) {
	// Create the MCP server with application, flow, and notification sender tools.
	mcpServer := NewServer(appService, flowService, notifService)

	// Create HTTP handler for MCP using Streamable HTTP transport.
	// Streamable HTTP supports both HTTP POST requests and optionally SSE for streaming.
	httpHandler := mcpsdk.NewStreamableHTTPHandler(func(*http.Request) *mcpsdk.Server {
		return mcpServer.GetMCPServer()
	}, nil)

	// Register MCP routes.
	mux.Handle("/mcp", httpHandler)
	mux.Handle("/mcp/", httpHandler)
}
