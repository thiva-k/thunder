# Authentication Demo - Postman Collection

This Postman collection demonstrates the main authentication/registration capabilities in ThunderID. It showcases three different approaches to authenticate users and includes resources setup for running the demos. Additionally, it includes one scenario for user registration.

## Demo Scenarios

### 1. Authenticate with Direct API endpoints
Individual authentication endpoints that can be used independently for specific authentication mechanisms:
- **Credential Login** - Username/password and email/password authentication
- **SMS OTP Login** - Send and verify SMS OTP
- **Google Login** - Social login with Google
- **GitHub Login** - Social login with GitHub
- **ThunderID Login** - OAuth-based login with ThunderID

### 2. Authenticate with Flow Native APIs
Orchestrated authentication flows using flow execution engine:
- **Login with Username & Password** - Basic username/password flow
- **Login with Username & Password (Verbose)** - Detailed flow execution with frontend metadata
- **Login with SMS OTP** - SMS OTP-based authentication flow
- **Login with MFA** - Multi-factor authentication (username/password + SMS OTP)
- **Login with Multi Options** - Multiple authentication options (username/password, Google, etc.)

### 3. Authenticate with OAuth (Standard Based)
Standard OAuth 2.0 authorization code flow with PKCE:
- Start authorization request
- Execute flow-based authentication
- Complete authorization and exchange code for tokens

### 4. Registration with Flow Native APIs
User self-registration flows:
- **Registration with Username & Password** - Basic user registration flow

## Collection Structure

```
├── 01 - Set Token              # Obtain access token for management APIs
├── 02 - Setup Resources        # Create demo resources (OU, schemas, integrations, users, apps, flows)
├── 03 - Authenticate with Direct API endpoints
│   ├── 03.01 - Credential Login
│   ├── 03.02 - SMS OTP Login
│   ├── 03.03 - Google Login
│   ├── 03.04 - GitHub Login
│   └── 03.05 - ThunderID Login
├── 04 - Authenticate with Flow Native APIs
│   ├── 04.01 - Login With Username & Password
│   ├── 04.02 - Login With Username & Password (Verbose)
│   ├── 04.03 - Login With SMS OTP
│   ├── 04.04 - Login With MFA
│   └── 04.05 - Login With Multi Options
├── 05 - Registration with Flow Native APIs
│   └── 05.01 - Registration With Username & Password
└── 06 - Authenticate with OAuth (Standard Based)
```

## Prerequisites

1. A running ThunderID server
2. Postman desktop app or web version
3. External service credentials (for social login demos):
   - Google OAuth credentials
   - GitHub OAuth credentials
   - ThunderID OAuth credentials (optional)
   - SMS notification sender webhook URL (for SMS OTP demos)

## Environment Setup

Import the `environment.json` file into Postman and fill in the required values. Or create a new environment manually with the following variables.

### Server Configuration

| Variable | Description | Example |
|----------|-------------|---------|
| `baseUrl` | ThunderID management API base URL | `https://localhost:8090` |
| `ADMIN_USERNAME` | Admin user username | `admin` |
| `ADMIN_PASSWORD` | Admin user password | `admin` |

### Service Integration Credentials (Setup Phase)

#### Google IDP (Optional)

| Variable | Description |
|----------|-------------|
| `GOOGLE_CLIENT_ID` | Google OAuth client ID |
| `GOOGLE_CLIENT_SECRET` | Google OAuth client secret |

#### GitHub IDP (Optional)

| Variable | Description |
|----------|-------------|
| `GITHUB_CLIENT_ID` | GitHub OAuth client ID |
| `GITHUB_CLIENT_SECRET` | GitHub OAuth client secret |

#### ThunderID IDP (Optional)

| Variable | Description |
|----------|-------------|
| `THUNDERID_CLIENT_ID` | ThunderID OAuth client ID |
| `THUNDERID_CLIENT_SECRET` | ThunderID OAuth client secret |
| `THUNDERID_BASE_URI` | ThunderID organization base URI (Ex: https://localhost:8090) |

Note: Replace `your-org` with your actual organization name.

#### Federated IDP Configuration (Common for all IDPs)

| Variable | Description | Example |
|----------|-------------|---------|
| `FED_IDP_REDIRECT_URI` | Redirect URI for federated IDPs | `https://localhost:8090/callback` |

### Notification Configuration

| Variable | Description | Example |
|----------|-------------|---------|
| `MESSAGE_NOTIFICATION_SENDER_URL` | Webhook URL for SMS sending | `https://webhook.site/xxxx` |

### Demo Application Configuration

| Variable | Description | Example |
|----------|-------------|---------|
| `LOGIN_APPLICATION_CLIENT_ID` | Client ID for the demo login app | `DEMO_CLIENT_ID` |
| `LOGIN_APPLICATION_REDIRECT_URI` | Redirect URI for the demo app | `https://localhost:3000` |

### Demo User Configuration

#### Local User

| Variable | Description | Example |
|----------|-------------|---------|
| `demoLoginUser1Username` | Demo user username | `jane` |
| `demoLoginUser1Password` | Demo user password | `password123` |
| `demoLoginUser1Email` | Demo user email | `jane@example.com` |
| `demoLoginUser1Mobile` | Demo user mobile number | `+1234567890` |

#### Google User

| Variable | Description |
|----------|-------------|
| `demoGoogleUserSub` | Google user subject identifier (unique id from Google) |
| `demoGoogleUserEmail` | Google user email |

#### GitHub User

| Variable | Description |
|----------|-------------|
| `demoGithubUserSub` | GitHub user subject identifier (unique id from GitHub) |
| `demoGithubUserEmail` | GitHub user email |

#### ThunderID User

| Variable | Description |
|----------|-------------|
| `demoThunderIDUserSub` | ThunderID user subject identifier (unique id from ThunderID) |
| `demoThunderIDUserEmail` | ThunderID user email |

### Resource IDs (for the Created Resources)

These variables will be auto-populated during the resource setup phase:

| Variable | Description |
|----------|-------------|
| `demoOuId` | Created demo organization unit ID |
| `demoOuHandle` | Created demo organization unit handle |
| `demoSchemaId` | Created demo user schema ID |
| `demoSchemaName` | Created demo user schema name |
| `googleIDPId` | Created Google IDP ID |
| `githubIDPId` | Created GitHub IDP ID |
| `thunderidIDPId` | Created ThunderID IDP ID |
| `messageNotificationSenderId` | Created SMS notification sender ID |
| `loginApplicationId` | Created demo application ID |
| `basicAuthFlowGraphId` | Created basic authentication flow graph ID |
| `smsAuthFlowGraphId` | Created SMS OTP authentication flow graph ID |
| `mfaAuthFlowGraphId` | Created MFA authentication flow graph ID |
| `multiOptionAuthFlowGraphId` | Created multi-option authentication flow graph ID |
| `basicRegistrationFlowGraphId` | Created basic registration flow graph ID |

### Token Management

These variables are used to manage tokens and authentication state. These will be auto-populated during the demo execution:

| Variable | Description |
|----------|-------------|
| `accessToken` | Access token for management API calls |
| `refreshToken` | Refresh token for token renewal |
| `expiresAt` | Token expiry timestamp |

## Collection Variables (Auto-populated)

These variables are automatically populated during the demo execution:

### Authentication Session (Demo Execution)

| Variable | Description |
|----------|-------------|
| `sms_otp_session_token` | SMS OTP session token |
| `first_factor_assertion` | First factor authentication assertion for MFA flows |
| `google_session_token` | Google authentication session token |
| `github_session_token` | GitHub authentication session token |
| `thunderid_session_token` | ThunderID authentication session token |
| `exec_flow_id` | Flow execution ID for flow-native APIs |
| `exec_challenge_token` | Challenge token for flow execution |
| `auth_std_auth_id` | OAuth standard flow auth ID |
| `auth_std_flow_id` | OAuth standard flow flow exec ID |
| `auth_std_assertion` | OAuth standard flow assertion |
| `auth_std_auth_code` | OAuth standard flow authorization code |

## Usage

### Step 1: Import Collection and Set Up Environment

1. Import the `authentication_demo.json` collection into Postman
2. Import the `environment.json` file or create a new environment with the required variables listed above
3. Fill in the environment variable values (server config, credentials, demo user details, etc)
4. Select the environment before running requests

### Step 2: Get Management Token

Run the requests in `01 - Set Token` folder sequentially:
1. **01 - Start Authorization** - Starts the OAuth flow
2. **02 - Init Flow** - Initializes the authentication flow
3. **03 - Execute Flow** - Completes authentication with admin credentials
4. **04 - Complete Authorization** - Completes the authorization
5. **05 - Exchange Code for Token** - Exchanges code for access token

### Step 3: Setup Demo Resources

Run the requests in `02 - Setup Resources` folder to create:
- Demo organization unit
- User schema
- Notification sender (for SMS OTP)
- Identity providers (Google, GitHub, ThunderID)
- Demo users
- Demo application
- Authentication and registration flows

### Step 4: Run Authentication Demos

Choose any of the authentication demo folders:
- `03 - Authenticate with Direct API endpoints` - For individual authentication mechanism demos
- `04 - Authenticate with Flow Native APIs` - For flow-based authentication demos
- `05 - Registration with Flow Native APIs` - For user registration demos
- `06 - Authenticate with OAuth (Standard Based)` - For OAuth 2.0 flow demo

## Notes

- The collection includes a pre-request script that automatically refreshes the access token when it's about to expire.
- For social login demos (Google, GitHub, ThunderID), you'll need to complete the OAuth flow in a browser and provide the authorization code manually.
- SMS OTP demos require a webhook endpoint to receive OTP messages. You can use services like [webhook.site](https://webhook.site/) for testing.
- Some requests may return `201` (created) or `409` (already exists) depending on whether resources already exist
