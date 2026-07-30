# MCP Calculator Sample

A scope-protected MCP server and an OAuth-authenticated MCP client, both secured with ThunderID. The server exposes four calculator tools, each gated behind its own OAuth scope, and the client signs in, registers itself, and calls the tools with a token that only carries the scopes the signed-in user's role grants.

This sample is the companion code for the ThunderID MCP quickstarts:

- [Secure Your MCP Server](https://thunderid.dev/docs/next/getting-started/connect-your-mcp/server/python/): build `server.py` and register the resource server, scopes, and role in ThunderID.
- [MCP Inspector](https://thunderid.dev/docs/next/getting-started/connect-your-mcp/client/connect/mcp-inspector/): connect MCP Inspector to the server.
- [Claude Code](https://thunderid.dev/docs/next/getting-started/connect-your-mcp/client/connect/claude-code/): connect Claude Code to the server using Dynamic Client Registration.
- [Build an MCP Client](https://thunderid.dev/docs/next/getting-started/connect-your-mcp/client/build-a-client/python/): build `client.py`, the OAuth-authenticated client in this sample.

## Project Structure

```text
mcp-calculator-sample/
├── README.md              This file
├── .gitignore             Ignores local secrets and caches
├── server/
│   ├── server.py          FastMCP resource server: add, subtract, multiply, divide
│   ├── .env.example       Template for the server's environment variables
│   ├── .env               Created locally, gitignored: your copy of .env.example
│   └── thunderid.cert     Created locally, gitignored: exported ThunderID certificate
└── client/
    ├── client.py          OAuth-authenticated FastMCP client
    └── .mcp-oauth-cache/  Created locally, gitignored: cached DCR registration and tokens
```

## Prerequisites

- A running ThunderID instance, reachable at `https://localhost:8090`.
- [`uv`](https://docs.astral.sh/uv/), which manages Python and dependencies for both `server.py` and `client.py` automatically.
- The Calculator MCP resource server, its four tool scopes, and a role that grants them to your test user, registered in ThunderID Console. Follow **Register the MCP Resource and Scopes** and **Create a Role for the Calculator Permissions** in [Secure Your MCP Server](https://thunderid.dev/docs/next/getting-started/connect-your-mcp/server/python/).

## Run the Server

From `server/`:

1. Copy the environment template:

   ```bash
   cp .env.example .env
   ```

2. Export ThunderID's self-signed certificate so the server can verify the JWKS endpoint over HTTPS. With ThunderID running, run this from the `server/` directory:

   ```bash
   openssl s_client -connect localhost:8090 -showcerts </dev/null 2>/dev/null | openssl x509 > thunderid.cert
   ```

3. Start the server:

   ```bash
   uv run server.py
   ```

The server listens at `http://localhost:8000/mcp` and rejects any request without a valid, appropriately scoped ThunderID access token. See [Configure the Environment](https://thunderid.dev/docs/next/getting-started/connect-your-mcp/server/python/#configure-the-environment) for what each `.env` variable does.

## Run the Client

From `client/`, with the server already running:

```bash
uv run client.py
```

`client.py` registers itself with ThunderID via Dynamic Client Registration (DCR) the first time it runs, so DCR must be enabled in your ThunderID deployment first. See [Enable Dynamic Client Registration](https://thunderid.dev/docs/next/getting-started/connect-your-mcp/client/build-a-client/python/#enable-dynamic-client-registration) for the `deployment.yaml` block to add.

On the first run, a browser tab opens with ThunderID's sign-in page. Sign in with your test user, and the client completes the OAuth flow and calls the tools:

```text
Tools: ['add', 'divide', 'multiply', 'subtract']
add(7, 5) = 12.0
divide(10, 4) = 2.5
```

The client persists its DCR registration and issued tokens under `client/.mcp-oauth-cache/` (gitignored). On later runs, it reuses that cache: no re-registration and no browser, as long as the cached token is still valid.

## Troubleshooting

- **`Registration failed: 400 invalid_client_metadata "An application with the same name already exists"`**: this happens when `client/.mcp-oauth-cache/` is deleted but the "FastMCP Client" application still exists in ThunderID. Delete that application in the Console, or restore the cache, then run again.
- **`npm error code EBADDEVENGINES`** when launching MCP Inspector with `NODE_EXTRA_CA_CERTS=./thunderid.cert npx @modelcontextprotocol/inspector` from `server/`: npm walked up to this repository's root `package.json`, which pins `pnpm` as the package manager, and refuses to run under plain `npm`. Launch Inspector from a directory outside the repository instead, pointing `NODE_EXTRA_CA_CERTS` at the absolute path of `thunderid.cert`.
- **Redirect to `/gate/error?errorCode=invalid_request&errorMessage=Invalid+redirect+URI`**: the registered redirect URI no longer matches the client's callback port. Delete the "FastMCP Client" application in the Console and run again so the client re-registers with the current port.

## Learn More

- [Secure Your MCP Server](https://thunderid.dev/docs/next/getting-started/connect-your-mcp/server/python/)
- [MCP Inspector](https://thunderid.dev/docs/next/getting-started/connect-your-mcp/client/connect/mcp-inspector/)
- [Claude Code](https://thunderid.dev/docs/next/getting-started/connect-your-mcp/client/connect/claude-code/)
- [Build an MCP Client](https://thunderid.dev/docs/next/getting-started/connect-your-mcp/client/build-a-client/python/)
- [Securing MCP](https://thunderid.dev/docs/next/use-cases/ai-agents/mcp-authorization/overview/)
