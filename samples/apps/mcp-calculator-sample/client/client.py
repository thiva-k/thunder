# /// script
# requires-python = ">=3.11"
# dependencies = ["fastmcp>=3.4,<4"]
# ///
"""OAuth-authenticated client for the Calculator MCP server."""

import asyncio
from pathlib import Path

from fastmcp import Client
from fastmcp.client.auth import FileTreeStore, OAuth

MCP_URL = "http://localhost:8000/mcp"
THUNDERID_CA_CERT = "../server/thunderid.cert"
TOKEN_CACHE_DIR = Path(__file__).parent / ".mcp-oauth-cache"
CALLBACK_PORT = 52360

oauth = OAuth(
    mcp_url=MCP_URL,
    token_storage=FileTreeStore(data_directory=TOKEN_CACHE_DIR),
    callback_port=CALLBACK_PORT,
)


async def main() -> None:
    async with Client(MCP_URL, auth=oauth, verify=THUNDERID_CA_CERT) as client:
        tools = await client.list_tools()
        print("Tools:", sorted(tool.name for tool in tools))

        result = await client.call_tool("add", {"a": 7, "b": 5})
        print("add(7, 5) =", result.data)

        result = await client.call_tool("divide", {"a": 10, "b": 4})
        print("divide(10, 4) =", result.data)


if __name__ == "__main__":
    asyncio.run(main())
