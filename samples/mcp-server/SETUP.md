# Setting Up MCP Server with Thunder Authorization

## Question: Can I use a pre-built MCP server?

**Short answer**: Most official MCP servers don't support HTTP + OAuth testing out of the box.

## Details

### Official MCP Servers
The official MCP servers from Anthropic (available on GitHub at `modelcontextprotocol/servers`) are:
- ✅ Great for VS Code/Cursor integration
- ✅ Use STDIO transport (standard input/output)
- ❌ **Don't support HTTP-based OAuth flows** (required for authorization testing)

Examples:
- `filesystem` - File system operations
- `git` - Git repository operations  
- `github` - GitHub API access
- `postgres` - Database queries

### Why STDIO is an issue?
STDIO-based MCP servers communicate via:
- Standard input/output pipes
- No HTTP endpoints
- No browser redirect support (needed for OAuth)

### For Authorization Testing, You Need:
1. **HTTP transport** - To serve protected resources
2. **HTTP endpoints** - For OAuth redirects
3. **RFC 9728 metadata** - Protected Resource Metadata endpoint

## Recommended Approach

### Option 1: Use This Example Server ✅ (Recommended for Testing)
- Already configured for HTTP + OAuth
- Simple and focused
- Ready to test Thunder immediately

```bash
cd samples/mcp-server
pip install -r requirements.txt
python mcp_server_vscode.py
```

### Option 2: Modify Official MCP Server (Advanced)
1. Clone from `modelcontextprotocol/servers`
2. Convert STDIO to HTTP transport
3. Add OAuth middleware
4. Implement RFC 9728 metadata endpoint

**This requires significant development work.**

### Option 3: Build Custom Server
- Use MCP TypeScript/Python SDK
- Implement HTTP transport
- Add OAuth integration
- More control, more work

## Quick Setup

```bash
# 1. Start Thunder (in another terminal)
./start.sh

# 2. Start MCP server (in this directory)
cd samples/mcp-server
pip install -r requirements.txt
python mcp_server_vscode.py

# 3. Add to VS Code/Cursor
# - Press Cmd+Shift+P (or Ctrl+Shift+P)
# - Select "MCP: Add server..."
# - Choose "HTTP"
# - Enter URL: http://localhost:3000
```

## Configuration

Set these environment variables (optional, defaults shown):

```bash
export THUNDER_AUTH_SERVER="http://localhost:8090"
export MCP_SERVER_URL="http://localhost:3000"
export MCP_SERVER_PORT="3000"
```

## Summary

For **testing Thunder's authorization capabilities**, this example server is the best option because:
- ✅ It's ready to use immediately
- ✅ Implements all MCP authorization requirements
- ✅ Works with Thunder out of the box
- ✅ Simple and focused on authorization testing

Official MCP servers are great for actual MCP functionality but require significant modification for OAuth testing.

