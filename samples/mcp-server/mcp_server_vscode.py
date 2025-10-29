#!/usr/bin/env python3
"""
MCP Server for VS Code/Cursor Integration with Thunder Authorization

This server implements:
- MCP Protocol (JSON-RPC 2.0)
- HTTP Transport (compatible with VS Code/Cursor)
- OAuth 2.0 Authorization with Thunder
- RFC 9728 Protected Resource Metadata
"""

from flask import Flask, request, jsonify
from functools import wraps
import requests
import jwt
import os
import uuid
import json
import urllib3
import base64
import struct
from urllib.parse import urljoin
from cryptography.hazmat.primitives.asymmetric import rsa
from cryptography.hazmat.primitives import serialization
from cryptography.hazmat.backends import default_backend

# Disable SSL warnings for development (self-signed certificates)
urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

# Configure requests to not verify SSL for all requests (development only)
requests.packages.urllib3.disable_warnings()

app = Flask(__name__)

# Configuration
THUNDER_AUTH_SERVER = os.getenv("THUNDER_AUTH_SERVER", "http://localhost:8090")
MCP_SERVER_URL = os.getenv("MCP_SERVER_URL", "http://localhost:3000")
MCP_SERVER_PORT = int(os.getenv("MCP_SERVER_PORT", "3000"))

# In-memory session storage for HTTP transport
sessions = {}


def get_jwks_uri():
    """Get JWKS URI from Thunder authorization server metadata"""
    try:
        metadata_url = urljoin(THUNDER_AUTH_SERVER, "/.well-known/oauth-authorization-server")
        # Disable SSL verification for development (self-signed certs)
        response = requests.get(metadata_url, verify=False, timeout=10)
        response.raise_for_status()  # Raise exception for bad status codes
        if response.status_code == 200:
            metadata = response.json()
            return metadata.get("jwks_uri")
    except requests.exceptions.RequestException as e:
        print(f"Error fetching metadata from {metadata_url}: {e}")
        print(f"Make sure Thunder is running at {THUNDER_AUTH_SERVER}")
    except Exception as e:
        print(f"Unexpected error fetching metadata: {e}")
    return None


# Initialize JWKS URI
jwks_uri = get_jwks_uri()


def validate_bearer_token(token):
    """Extract and log Bearer token (no validation)"""
    if not token or not token.startswith("Bearer "):
        return None, "Invalid Authorization header format"
    
    token = token[7:]  # Remove "Bearer " prefix
    
    try:
        # Decode token WITHOUT verification - just for logging
        # No signature, expiration, issuer, or audience validation
        decoded = jwt.decode(
            token,
            options={"verify_signature": False}
        )
        
        # Log the full access token claims to console
        import json
        print("=" * 80)
        print("🔐 ACCESS TOKEN DECODED (NO VALIDATION):")
        print(json.dumps(decoded, indent=2))
        print("=" * 80)
        
        return decoded, None
        
    except Exception as e:
        import traceback
        print(f"❌ [TOKEN] Error decoding token: {e}")
        print(f"   Traceback: {traceback.format_exc()}")
        return None, f"Token decode error: {str(e)}"


def require_auth(f):
    """Decorator to require valid Bearer token"""
    @wraps(f)
    def decorated_function(*args, **kwargs):
        auth_header = request.headers.get("Authorization")
        
        if not auth_header:
            print(f"❌ [AUTH] Missing Authorization header for {request.method} {request.path}")
            resource_metadata_url = f"{MCP_SERVER_URL}/.well-known/oauth-protected-resource"
            return jsonify({
                "jsonrpc": "2.0",
                "error": {
                    "code": -32001,
                    "message": "Unauthorized: Missing Authorization header"
                },
                "id": request.json.get("id") if request.is_json else None
            }), 401, {
                "WWW-Authenticate": f'Bearer realm="mcp", resource_metadata="{resource_metadata_url}"'
            }
        
        # Extract token for logging (first 20 chars for security)
        token_preview = auth_header[:27] + "..." if len(auth_header) > 27 else auth_header
        print(f"🔍 [AUTH] Validating token: {token_preview}")
        
        claims, error = validate_bearer_token(auth_header)
        
        if error:
            print(f"❌ [AUTH] Token validation failed: {error}")
            print(f"   Request method: {request.method}, path: {request.path}")
            if request.is_json:
                print(f"   RPC method: {request.json.get('method', 'unknown')}")
            resource_metadata_url = f"{MCP_SERVER_URL}/.well-known/oauth-protected-resource"
            return jsonify({
                "jsonrpc": "2.0",
                "error": {
                    "code": -32001,
                    "message": f"Unauthorized: {error}"
                },
                "id": request.json.get("id") if request.is_json else None
            }), 401, {
                "WWW-Authenticate": f'Bearer error="invalid_token", error_description="{error}", resource_metadata="{resource_metadata_url}"'
            }
        
        print(f"✅ [AUTH] Token validated successfully")
        request.token_claims = claims
        return f(*args, **kwargs)
    
    return decorated_function


@app.route("/.well-known/oauth-protected-resource", methods=["GET"])
@app.route("/.well-known/oauth-resource-metadata", methods=["GET"])  # Alias for compatibility
def resource_metadata():
    """
    OAuth 2.0 Protected Resource Metadata (RFC 9728)
    Required for MCP authorization discovery
    
    RFC 9728 specifies: /.well-known/oauth-protected-resource
    """
    metadata = {
        "resource": MCP_SERVER_URL,
        "authorization_servers": [THUNDER_AUTH_SERVER],
        "scopes_supported": ["mcp:tools"],
        "bearer_methods_supported": ["header"]
    }
    
    return jsonify(metadata), 200, {
        "Content-Type": "application/json",
        "Access-Control-Allow-Origin": "*"
    }


def handle_initialize(params):
    """Handle MCP initialize request"""
    session_id = str(uuid.uuid4())
    # Use the protocol version from client, or default to latest
    client_protocol_version = params.get("protocolVersion", "2024-11-05")
    sessions[session_id] = {
        "initialized": True,
        "protocol_version": client_protocol_version,
        "capabilities": params.get("capabilities", {}),
        "client_info": params.get("clientInfo", {})
    }
    
    # Return capabilities matching what client requested
    client_capabilities = params.get("capabilities", {})
    
    return {
        "protocolVersion": client_protocol_version,  # Match client's version
        "capabilities": {
            "tools": {} if client_capabilities.get("tools") else None,  # Support tools if client requests it
            "prompts": {} if client_capabilities.get("prompts") else None,
            "resources": {} if client_capabilities.get("resources") else None
        },
        "serverInfo": {
            "name": "thunder-mcp-server",
            "version": "1.0.0"
        }
    }, session_id


def handle_tools_list():
    """Handle tools/list request"""
    return {
        "tools": [
            {
                "name": "add",
                "description": "Add two numbers together",
                "inputSchema": {
                    "type": "object",
                    "properties": {
                        "a": {
                            "type": "number",
                            "description": "First number to add"
                        },
                        "b": {
                            "type": "number",
                            "description": "Second number to add"
                        }
                    },
                    "required": ["a", "b"]
                }
            },
            {
                "name": "multiply",
                "description": "Multiply two numbers together",
                "inputSchema": {
                    "type": "object",
                    "properties": {
                        "x": {
                            "type": "number",
                            "description": "First number to multiply"
                        },
                        "y": {
                            "type": "number",
                            "description": "Second number to multiply"
                        }
                    },
                    "required": ["x", "y"]
                }
            }
        ]
    }


def handle_tools_call(tool_name, arguments):
    """Handle tools/call request"""
    if tool_name == "add":
        result = arguments.get("a", 0) + arguments.get("b", 0)
        return {
            "content": [
                {
                    "type": "text",
                    "text": f"{arguments.get('a')} + {arguments.get('b')} = {result}"
                }
            ]
        }
    elif tool_name == "multiply":
        result = arguments.get("x", 0) * arguments.get("y", 0)
        return {
            "content": [
                {
                    "type": "text",
                    "text": f"{arguments.get('x')} × {arguments.get('y')} = {result}"
                }
            ]
        }
    else:
        raise ValueError(f"Unknown tool: {tool_name}")


@app.route("/", methods=["POST"])
@require_auth
def handle_mcp_request():
    """
    Main MCP protocol endpoint
    Handles JSON-RPC 2.0 requests
    """
    session_id = request.headers.get("mcp-session-id")
    
    if not request.is_json:
        return jsonify({
            "jsonrpc": "2.0",
            "error": {
                "code": -32700,
                "message": "Parse error"
            },
            "id": None
        }), 400
    
    rpc_request = request.json
    
    # Validate JSON-RPC 2.0 format
    if rpc_request.get("jsonrpc") != "2.0":
        return jsonify({
            "jsonrpc": "2.0",
            "error": {
                "code": -32600,
                "message": "Invalid Request"
            },
            "id": rpc_request.get("id")
        }), 400
    
    method = rpc_request.get("method")
    params = rpc_request.get("params", {})
    request_id = rpc_request.get("id")
    
    print(f"📨 [MCP] Incoming request - Method: {method}, ID: {request_id}")
    
    try:
        # Handle notifications first (they have id: null and don't require a response)
        if request_id is None:
            if method == "notifications/initialized":
                print("✅ [MCP] notifications/initialized received (no response needed)")
                # Notifications don't return a response in JSON-RPC 2.0
                # Return empty 200 OK response
                return jsonify({}), 200
            else:
                print(f"⚠️  [MCP] Unknown notification: {method}")
                # Still return success for unknown notifications
                return jsonify({}), 200
        
        # Handle initialize (creates new session)
        if method == "initialize":
            print("🚀 [MCP] initialize called")
            print(f"   Params: {json.dumps(params, indent=2)}")
            result, new_session_id = handle_initialize(params)
            print(f"✅ [MCP] Initialize successful, session: {new_session_id}")
            print(f"   Result: {json.dumps(result, indent=2)}")
            response = jsonify({
                "jsonrpc": "2.0",
                "result": result,
                "id": request_id
            })
            response.headers["mcp-session-id"] = new_session_id
            return response
        
        # For other methods, require existing session
        if not session_id or session_id not in sessions:
            return jsonify({
                "jsonrpc": "2.0",
                "error": {
                    "code": -32000,
                    "message": "Server error: No valid session. Call initialize first."
                },
                "id": request_id
            }), 400
        
        # Route to appropriate handler for requests
        if method == "tools/list":
            print("📋 [MCP] tools/list called - returning available tools")
            result = handle_tools_list()
            print(f"✅ [MCP] Returning {len(result.get('tools', []))} tools")
        elif method == "tools/call":
            tool_name = params.get("name")
            arguments = params.get("arguments", {})
            print(f"🔧 [MCP] tools/call - Tool: {tool_name}, Arguments: {arguments}")
            result = handle_tools_call(tool_name, arguments)
            print(f"✅ [MCP] Tool execution result: {result}")
        else:
            print(f"❌ [MCP] Unknown method: {method}")
            return jsonify({
                "jsonrpc": "2.0",
                "error": {
                    "code": -32601,
                    "message": f"Method not found: {method}"
                },
                "id": request_id
            }), 400
        
        response = jsonify({
            "jsonrpc": "2.0",
            "result": result,
            "id": request_id
        })
        response.headers["mcp-session-id"] = session_id
        return response
        
    except Exception as e:
        return jsonify({
            "jsonrpc": "2.0",
            "error": {
                "code": -32603,
                "message": f"Internal error: {str(e)}"
            },
            "id": request_id
        }), 500


@app.route("/", methods=["GET", "DELETE"])
@require_auth
def handle_session_request():
    """Handle GET/DELETE for session management"""
    session_id = request.headers.get("mcp-session-id")
    
    if request.method == "DELETE":
        if session_id and session_id in sessions:
            del sessions[session_id]
            return "", 204
        return "", 404
    
    # GET request - could be used for SSE in future
    return jsonify({
        "status": "ok",
        "session_id": session_id
    }), 200


@app.route("/health", methods=["GET"])
def health():
    """Health check endpoint"""
    return jsonify({
        "status": "healthy",
        "authorization_server": THUNDER_AUTH_SERVER,
        "mcp_server_url": MCP_SERVER_URL,
        "jwks_uri": jwks_uri
    }), 200


if __name__ == "__main__":
    print("🚀 MCP Server starting...")
    print(f"📡 Authorization Server: {THUNDER_AUTH_SERVER}")
    print(f"📡 MCP Server URL: {MCP_SERVER_URL}")
    print(f"🔐 OAuth metadata: {MCP_SERVER_URL}/.well-known/oauth-protected-resource")
    print(f"🔑 JWKS URI: {jwks_uri}")
    print("\nTo add to VS Code/Cursor:")
    print(f"1. Open Command Palette (Cmd+Shift+P)")
    print(f"2. Select 'MCP: Add server...'")
    print(f"3. Choose 'HTTP'")
    print(f"4. Enter URL: {MCP_SERVER_URL}")
    print("\nServer starting...")
    
    app.run(host="0.0.0.0", port=MCP_SERVER_PORT, debug=True)

