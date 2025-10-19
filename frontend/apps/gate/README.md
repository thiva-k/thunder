# Thunder - Gate App

This is the gate app for project thunder. Which serves UIs for Login, Registration and Recovery.

### ✅ Prerequisites

- Node.js 20+
- PNPM 10+

---

### 🛠 Step 1: Install Dependencies

```bash
pnpm i
```

### 🔐 Step 2: Generate SSL Certificates

```bash
openssl req -x509 -newkey rsa:4096 -keyout server.key -out server.cert -days 365 -nodes -subj "/C=US/ST=State/L=City/O=Organization/CN=localhost"
```

### ▶️ Step 3: Run the Application

```bash
pnpm --filter gate dev
```

### ⚙️ Step 4: Configuration

The application supports environment variables for configuration that can be changed after build time.

#### Environment Variables

Copy `.env.example` to `.env.local` for local development:

```bash
cp .env.example .env.local
```

Available environment variables:

- `NEXT_PUBLIC_PRODUCT_NAME`: Product name displayed in the UI (default: "WSO2 Thunder")
- `NEXT_PUBLIC_THUNDER_HOST`: Thunder backend host URL (default: `https://localhost:8090`)
- `NEXT_PUBLIC_FLOW_EXECUTION_ENDPOINT`: Backend flow execution API endpoint (defaults to `{THUNDER_HOST}/flow/execute`)
- `NEXT_PUBLIC_AUTHORIZATION_ENDPOINT`: Backend authorization API endpoint (defaults to `{THUNDER_HOST}/oauth2/authorize`)

**Note**: If you set `NEXT_PUBLIC_THUNDER_HOST`, you don't need to set the individual endpoint URLs unless you want to override them specifically.
