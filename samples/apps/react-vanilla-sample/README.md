# React Vanilla Sample Application

This sample React application demonstrates integrating authentication and registration into your application using the app-native flow orchestration API.

### Supported Authentication Methods

This sample supports the following authentication and registration methods:

- **Basic authentication** — username and password
- **Social login** — Google and GitHub OAuth
- **SMS OTP** — one-time password via SMS
- **Passkeys** — FIDO2/WebAuthn passkey authentication

To try these out, configure the corresponding flows in your ThunderID instance and assign them to your application. The UI automatically adapts to the options returned by the flow.

## Prerequisites

- Node.js 20+
- A running ThunderID server instance (default: `https://localhost:8090`)
- An application registered in the server

## Configuration

Open `public/runtime.json` and set your values:

```json
{
    "flowEndpoint": "https://localhost:8090/flow",
    "applicationID": "{your-application-id}"
}
```

| Property | Description |
|----------|-------------|
| `flowEndpoint` | Flow orchestration endpoint |
| `applicationID` | The application ID registered in the server |

### Passkey Configuration

WebAuthn requires the server to validate that the credential was created from a trusted origin. By default, ThunderID only allows `https://localhost:8090` (the server itself). When running this sample app at `https://localhost:3000`, you must add that origin to the allowed list in the server's `deployment.yaml`:

```yaml
passkey:
  allowed_origins:
    - "https://localhost:8090"
    - "https://localhost:3000"
```

If the sample is hosted at a different address, add that origin instead. Without this, passkey registration will fail with an origin validation error.

## Quick Start

### 1. Install Dependencies

```bash
npm install
```

### 2. Set Up SSL Certificates

Copy the SSL certificates from your server distribution to the project root:

```bash
# From distribution
cp /path/to/thunderid/repository/resources/security/server.key .
cp /path/to/thunderid/repository/resources/security/server.cert .

# Or from build output (if building from source)
cp ../../target/out/.cert/server.key .
cp ../../target/out/.cert/server.cert .
```

Or generate self-signed certificates:

```bash
openssl req -nodes -new -x509 -keyout server.key -out server.cert
```

### 3. Start Development Server

```bash
npm run dev
```

The application will be available at [https://localhost:3000](https://localhost:3000).

## Additional Info

### UI Rendering and Action Ref Convention

Action button labels and styles are driven by the `ref` value of each action. For actions with no special keyword, the `ref` value itself is used as the button label.

The following keywords in the `ref` trigger special rendering:

| Keyword in `ref` | Rendered as |
|------------------|-------------|
| `basic_auth` | Username and password form |
| `google` | "Continue with Google" with Google icon |
| `github` | "Continue with GitHub" with GitHub icon |
| `sms` or `mobile` | "Continue with SMS OTP" |
| `passkey` | "Sign in with Passkey" with fingerprint icon |
| `signin` or `sign_in` | "Sign In" (submit button label) |
| `signup` or `sign_up` | "Create Account" (submit button label) |

## Available Scripts

| Command | Description |
|---------|-------------|
| `npm run dev` | Start development server with hot reloading |
| `npm run build` | Build for production (outputs to `dist/`) |
| `npm run preview` | Preview the production build locally |
| `npm run lint` | Run ESLint to check code quality |
| `npm start` | Build and preview the production application |

## Hosting Options

### Using the Provided Node Server

```bash
npm install && npm run build
cd server && npm start
```

### Using Your Own Web Server

Host the contents of the `dist/` folder on any HTTPS-capable web server. Ensure `runtime.json` is served and accessible for configuration.

## License

Licensed under the Apache License, Version 2.0.

---------------------------------------------------------------------------
(c) Copyright 2025 WSO2 LLC.
