# MCP Server Example for Testing Thunder Authorization

This is a simple MCP server implementation that uses Thunder as its authorization server, demonstrating how to integrate with Thunder for OAuth 2.0 authorization per MCP specifications.

## Overview

This MCP server implements:
- OAuth 2.0 Protected Resource Metadata (RFC 9728)
- Bearer token validation
- Audience claim validation
- Integration with Thunder authorization server

## Available Pre-built MCP Servers

### Important Note
Most official MCP servers use **STDIO transport** (for VS Code/Cursor integration), which doesn't support HTTP-based OAuth flows. To test Thunder as an authorization server, you need an **HTTP-based MCP server**.

### Options:

1. **Official MCP TypeScript SDK Examples** (GitHub: `modelcontextprotocol/servers`)
   - Need to modify to use HTTP transport and add OAuth
   - Examples: filesystem, git, etc.
   
2. **Custom Implementation (This Example)**
   - ✅ Already configured for HTTP + OAuth
   - ✅ Ready to test Thunder authorization
   - ✅ Simple and focused on authorization testing

### Why Use This Example?

The official MCP servers from Anthropic are designed for STDIO (command-line) usage with VS Code/Cursor, which doesn't support:
- OAuth 2.0 authorization flows (requires browser redirects)
- HTTP resource endpoints for protected resources
- RFC 9728 Protected Resource Metadata

This example server fills that gap for **testing authorization server functionality**.

## Prerequisites

- Python 3.8+
- Thunder authorization server running (default: `http://localhost:8090`)
- pip packages: `flask`, `requests`, `pyjwt`, `cryptography`
- VS Code or Cursor with MCP extension (for integration testing)

> **Note**: This directory contains only `mcp_server_vscode.py` which is the main server for VS Code/Cursor integration. All other unnecessary files have been removed.

## Server

### `mcp_server_vscode.py` (VS Code/Cursor MCP Server)
- ✅ Implements MCP JSON-RPC 2.0 protocol
- ✅ HTTP transport compatible with VS Code/Cursor
- ✅ OAuth 2.0 authorization with Thunder
- ✅ Example tools (add, multiply)
- ✅ RFC 9728 Protected Resource Metadata
- **Use this for VS Code/Cursor integration**

## Setup

1. Install dependencies:
```bash
pip install flask requests pyjwt cryptography
```

2. Configure Thunder authorization server URL (default: `http://localhost:8090`):
```bash
export THUNDER_AUTH_SERVER="http://localhost:8090"
```

3. Run the MCP server:
```bash
python mcp_server_vscode.py
```

The server will run on `http://localhost:3000`

## Quick Start for VS Code/Cursor

### Using `mcp_server_vscode.py` (Recommended for VS Code/Cursor)

This server implements the full MCP protocol with HTTP transport, compatible with VS Code/Cursor:

1. **Start Thunder server**:
   ```bash
   ./start.sh  # or run Thunder in another terminal
   ```

2. **Start the MCP server**:
   ```bash
   cd samples/mcp-server
   pip install -r requirements.txt
   python mcp_server_vscode.py
   ```

3. **Add to VS Code/Cursor**:
   - Press `Cmd+Shift+P` (or `Ctrl+Shift+P` on Windows/Linux)
   - Select **"MCP: Add server..."**
   - Choose **"HTTP"**
   - Enter URL: `http://localhost:3000`
   - Give it a name (e.g., "Thunder MCP Server")

4. **Authenticate**:
   - VS Code/Cursor will open a browser for OAuth flow
   - Complete authentication with Thunder
   - Grant permissions for `mcp:tools` scope

5. **Use the tools**:
   - Once connected, you'll see tools like "add" and "multiply"
   - Use `#add` or `#multiply` in the chat to invoke them

## Testing Flow

1. **Register an OAuth client** using Thunder's DCR endpoint (VS Code/Cursor does this automatically, or manually):
   ```bash
   curl -X POST http://localhost:8090/oauth2/dcr/register \
     -H "Content-Type: application/json" \
     -d '{
       "redirect_uris": ["http://localhost:3000/callback"],
       "grant_types": ["authorization_code", "refresh_token"],
       "response_types": ["code"],
       "token_endpoint_auth_method": "none",
       "scope": "mcp:tools"
     }'
   ```

2. **VS Code/Cursor will automatically**:
   - Open browser for OAuth flow
   - Complete authentication with Thunder
   - Exchange authorization code for access token
   - Connect to MCP server with Bearer token

3. **Manual testing** (optional):
   ```bash
   # Get authorization code from browser callback
   # Exchange for token
   curl -X POST http://localhost:8090/oauth2/token \
     -H "Content-Type: application/x-www-form-urlencoded" \
     -d "grant_type=authorization_code&code=<AUTH_CODE>&client_id=<CLIENT_ID>&code_verifier=<VERIFIER>&resource=http://localhost:3000"
   
   # Access protected MCP endpoint
   curl -X POST http://localhost:3000 \
     -H "Authorization: Bearer <ACCESS_TOKEN>" \
     -H "Content-Type: application/json" \
     -d '{"jsonrpc":"2.0","method":"tools/list","id":1}'
   ```

## Endpoints

- `GET /.well-known/oauth-protected-resource` - Protected Resource Metadata (RFC 9728)
- `POST /` - MCP JSON-RPC 2.0 endpoint (requires valid Bearer token)
  - `initialize` - Initialize MCP session
  - `tools/list` - List available tools
  - `tools/call` - Call a tool

## Security Notes

- The server decodes and logs access tokens (for debugging)
- Token signature validation can be enabled by uncommenting verification code
- Configure proper TLS certificates for production use

