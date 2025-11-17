# Exporting Configurations

This guide explains how to export Thunder configurations as YAML files for version control, backup, and deployment automation.

## Overview

Thunder provides an export functionality that allows you to export resource configurations as parameterized YAML files. These exported files can be:

- Version controlled in your repository
- Used with Thunder's immutable configuration mode for GitOps workflows
- Backed up for disaster recovery

## Current Support

**Currently Supported Resources:**
- ✅ **Applications** - Full support with parameterization


## Export API

### Export Applications

You can export application configurations using the `/export` API endpoint.

#### Export as YAML (Recommended)

```bash
curl -X POST https://localhost:8090/export \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <access-token>" \
  -d '{
    "applications": ["<application-id>"]
  }'
```

**Response:**
```yaml
# File: My_Application.yaml
# Resource Type: application
# Resource ID: 550e8400-e29b-41d4-a716-446655440000

name: My Application
description: Production application
url: https://myapp.example.com
logo_url: https://myapp.example.com/logo.png
auth_flow_graph_id: auth_flow_config_basic
registration_flow_graph_id: registration_flow_config_basic
is_registration_flow_enabled: true
inbound_auth_config:
  - type: oauth2
    config:
      client_id: {{.MY_APPLICATION_CLIENT_ID}}
      redirect_uris:
        {{- range .MY_APPLICATION_REDIRECT_URIS}}
        - {{.}}
        {{- end}}
      grant_types:
        - authorization_code
        - refresh_token
      response_types:
        - code
      token_endpoint_auth_method: client_secret_basic
      pkce_required: false
      public_client: false
      token:
        issuer: thunder
        access_token:
          validity_period: 3600
          user_attributes:
            - email
            - name
            - groups
        id_token:
          validity_period: 3600
          user_attributes:
            - email
            - name
            - groups
```

#### Export Multiple Applications

Export all applications or specific ones:

```bash
# Export all applications
curl -X POST https://localhost:8090/export \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <access-token>" \
  -d '{
    "applications": ["*"]
  }'

# Export specific applications
curl -X POST https://localhost:8090/export \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <access-token>" \
  -d '{
    "applications": [
      "app-id-1",
      "app-id-2",
      "app-id-3"
    ]
  }'
```

#### Export as ZIP Archive

For downloading multiple files at once:

```bash
curl -X POST https://localhost:8090/export/zip \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <access-token>" \
  -d '{
    "applications": ["*"]
  }' \
  --output thunder-export.zip
```

This creates a ZIP file with the following structure:
```
thunder-export.zip
└── applications/
    ├── My_Application.yaml
    ├── Mobile_App.yaml
    └── Web_Portal.yaml
```

## Parameterized Variables

Thunder automatically parameterizes sensitive and environment-specific values during export. This allows you to use the same configuration across different deployments by simply changing the variable values.

### Application Parameterization

The following fields are automatically parameterized in application exports:

#### Variables (Simple Values)

| Field | Parameter Format | Example |
|-------|-----------------|----------|
| `client_id` | `{{.APP_NAME_CLIENT_ID}}` | `{{.MY_APP_CLIENT_ID}}` |
| `client_secret` | `{{.APP_NAME_CLIENT_SECRET}}` | `{{.MY_APP_CLIENT_SECRET}}` |

#### Array Variables

| Field | Parameter Format | Example |
|-------|-----------------|----------|
| `redirect_uris` | `{{- range .APP_NAME_REDIRECT_URIS}}` | See below |

**Array Variable Example:**
```yaml
redirect_uris:
  {{- range .MY_APPLICATION_REDIRECT_URIS}}
  - {{.}}
  {{- end}}
```

### Variable Naming Convention

Variables are automatically generated from the application name:

1. Application name is converted to uppercase
2. Spaces and special characters are replaced with underscores
3. Field name is appended

**Examples:**
- "My Application" + `client_id` → `MY_APPLICATION_CLIENT_ID`
- "Web-Portal" + `client_secret` → `WEB_PORTAL_CLIENT_SECRET`
- "Mobile App" + `redirect_uris` → `MOBILE_APP_REDIRECT_URIS`

### Non-Parameterized Fields

The following fields are **NOT** parameterized and are exported as-is:

- Application name and description
- Flow graph IDs
- URLs (application URL, logo URL, ToS, policy)
- Token configurations
- Grant types and response types
- User attributes
- Scope claims

These can be customized in the YAML file directly if needed.

## Providing Variable Values

Exported YAML files contain parameterized variables that need to be provided when running Thunder with immutable configuration mode. This section explains how to inject these values using environment variables.

### Simple Variables

Simple variables use the `{{.VARIABLE_NAME}}` template syntax and are replaced with environment variable values.

**Exported YAML:**
```yaml
inbound_auth_config:
  - type: oauth2
    config:
      client_id: {{.MY_APP_CLIENT_ID}}
```

**Environment Variable:**
```bash
export MY_APP_CLIENT_ID=prod-client-id-12345
```

When Thunder starts, `{{.MY_APP_CLIENT_ID}}` is replaced with `prod-client-id-12345`.

### Array Variables

Array variables use the `{{- range .ARRAY_NAME}}` template syntax and are populated from indexed environment variables.

**Exported YAML:**
```yaml
redirect_uris:
  {{- range .MY_APP_REDIRECT_URIS}}
  - {{.}}
  {{- end}}
```

**Environment Variables (Indexed Pattern):**
```bash
export MY_APP_REDIRECT_URIS_0=https://app.example.com/callback
export MY_APP_REDIRECT_URIS_1=https://app.example.com/silent-callback
export MY_APP_REDIRECT_URIS_2=https://app.example.com/logout
```

Thunder reads variables with `_0`, `_1`, `_2`, etc. suffixes sequentially until it finds an empty or non-existent variable. The above variables will produce:

```yaml
redirect_uris:
  - https://app.example.com/callback
  - https://app.example.com/silent-callback
  - https://app.example.com/logout
```

#### Important Array Rules

1. **Start at index 0**: Always begin with `VARIABLE_NAME_0`
2. **Sequential indices**: No gaps allowed (0, 1, 2, not 0, 2, 4)
3. **Stop at first empty**: Thunder stops reading when it encounters an empty value or missing variable

**Example - Incorrect (will only get first 2 items):**
```bash
export MY_APP_REDIRECT_URIS_0=https://first.com/callback
export MY_APP_REDIRECT_URIS_1=https://second.com/callback
# Missing _2, so Thunder stops here
export MY_APP_REDIRECT_URIS_3=https://third.com/callback  # This will be ignored
```

**Example - Correct:**
```bash
export MY_APP_REDIRECT_URIS_0=https://first.com/callback
export MY_APP_REDIRECT_URIS_1=https://second.com/callback
export MY_APP_REDIRECT_URIS_2=https://third.com/callback
export MY_APP_REDIRECT_URIS_3=https://fourth.com/callback
```

### Complete Application Example

Here's a complete example showing how to provide all variables for an exported application:

**Exported YAML** (`My_Application.yaml`):
```yaml
name: My Application
description: Production application
url: https://myapp.example.com
auth_flow_graph_id: auth_flow_config_basic
inbound_auth_config:
  - type: oauth2
    config:
      client_id: {{.MY_APPLICATION_CLIENT_ID}}
      redirect_uris:
        {{- range .MY_APPLICATION_REDIRECT_URIS}}
        - {{.}}
        {{- end}}
      grant_types:
        - authorization_code
        - refresh_token
```

**Environment Variables:**
```bash
# Simple variable for client_id
export MY_APPLICATION_CLIENT_ID=my-prod-client-id

# Array variables for redirect_uris (indexed from 0)
export MY_APPLICATION_REDIRECT_URIS_0=https://myapp.example.com/callback
export MY_APPLICATION_REDIRECT_URIS_1=https://myapp.example.com/oauth2/callback
export MY_APPLICATION_REDIRECT_URIS_2=https://myapp.example.com/silent-callback
```

**Resulting Configuration** (what Thunder loads):
```yaml
name: My Application
description: Production application
url: https://myapp.example.com
auth_flow_graph_id: auth_flow_config_basic
inbound_auth_config:
  - type: oauth2
    config:
      client_id: my-prod-client-id
      redirect_uris:
        - https://myapp.example.com/callback
        - https://myapp.example.com/oauth2/callback
        - https://myapp.example.com/silent-callback
      grant_types:
        - authorization_code
        - refresh_token
```

### Multiple Applications Example

When managing multiple applications, organize variables by application name:

```bash
# Application 1: Web Application
export WEB_APP_CLIENT_ID=web-client-123
export WEB_APP_REDIRECT_URIS_0=https://web.example.com/callback
export WEB_APP_REDIRECT_URIS_1=https://web.example.com/logout

# Application 2: Mobile Application  
export MOBILE_APP_CLIENT_ID=mobile-client-456
export MOBILE_APP_REDIRECT_URIS_0=myapp://callback
export MOBILE_APP_REDIRECT_URIS_1=myapp://logout

# Application 3: Admin Portal
export ADMIN_PORTAL_CLIENT_ID=admin-client-789
export ADMIN_PORTAL_REDIRECT_URIS_0=https://admin.example.com/callback
```

### Using Environment Files

For easier management, store variables in environment files:

**production.env:**
```bash
# Web Application
WEB_APP_CLIENT_ID=web-prod-client-id
WEB_APP_CLIENT_SECRET=web-prod-secret-xyz
WEB_APP_REDIRECT_URIS_0=https://app.example.com/callback
WEB_APP_REDIRECT_URIS_1=https://app.example.com/silent-callback
WEB_APP_REDIRECT_URIS_2=https://app.example.com/logout

# Mobile Application
MOBILE_APP_CLIENT_ID=mobile-prod-client-id
MOBILE_APP_CLIENT_SECRET=mobile-prod-secret-abc
MOBILE_APP_REDIRECT_URIS_0=myapp://callback
MOBILE_APP_REDIRECT_URIS_1=myapp://logout
```

**Load and run:**
```bash
# Load environment variables
source production.env

# Or use export with cat
export $(cat production.env | xargs)

# Start Thunder with immutable configuration
./start.sh
```

### Docker Example

Inject variables via Docker:

```bash
docker run \
  -e MY_APP_CLIENT_ID=prod-client-id \
  -e MY_APP_CLIENT_SECRET=prod-secret \
  -e MY_APP_REDIRECT_URIS_0=https://app.example.com/callback \
  -e MY_APP_REDIRECT_URIS_1=https://app.example.com/logout \
  -v $(pwd)/configs:/app/repository/conf/immutable_resources \
  thunder:latest
```

Or use an environment file:

```bash
docker run --env-file production.env \
  -v $(pwd)/configs:/app/repository/conf/immutable_resources \
  thunder:latest
```

### Kubernetes Example

Use Kubernetes Secrets to inject variables:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: thunder-app-secrets
type: Opaque
stringData:
  # Simple variables
  MY_APP_CLIENT_ID: prod-client-id
  MY_APP_CLIENT_SECRET: prod-secret
  
  # Array variables - indexed
  MY_APP_REDIRECT_URIS_0: https://app.example.com/callback
  MY_APP_REDIRECT_URIS_1: https://app.example.com/silent-callback
  MY_APP_REDIRECT_URIS_2: https://app.example.com/logout
```

Reference in Deployment:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: thunder
spec:
  template:
    spec:
      containers:
      - name: thunder
        envFrom:
        - secretRef:
            name: thunder-app-secrets
```

## Best Practices

### 1. Version Control

Store exported YAML configurations in version control, but **never commit environment files with actual secrets**:

```bash
# Create a configs directory
mkdir -p configs/applications

# Export all applications
curl -X POST https://localhost:8090/export/zip \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <access-token>" \
  -d '{"applications": ["*"]}' \
  --output configs.zip

# Extract to configs directory
unzip configs.zip -d configs/

# Commit YAML files (these contain template variables, not secrets)
git add configs/
git commit -m "Export Thunder application configurations"

# IMPORTANT: Add .env files to .gitignore
echo "*.env" >> .gitignore
echo ".env.*" >> .gitignore
git add .gitignore
git commit -m "Ignore environment files"
```

### 2. Environment-Specific Configuration

Create separate environment files for each deployment:

**Structure:**
```
project/
├── configs/
│   └── applications/
│       ├── web-app.yaml
│       └── mobile-app.yaml
└── environments/
    ├── dev.env
    ├── staging.env
    └── prod.env
```

**dev.env:**
```bash
WEB_APP_CLIENT_ID=dev-web-client
WEB_APP_CLIENT_SECRET=dev-secret
WEB_APP_REDIRECT_URIS_0=http://localhost:3000/callback
WEB_APP_REDIRECT_URIS_1=http://localhost:3000/logout
```

**prod.env:**
```bash
WEB_APP_CLIENT_ID=prod-web-client
WEB_APP_CLIENT_SECRET=prod-secret-xyz
WEB_APP_REDIRECT_URIS_0=https://app.example.com/callback
WEB_APP_REDIRECT_URIS_1=https://app.example.com/silent-callback
WEB_APP_REDIRECT_URIS_2=https://app.example.com/logout
```


## Export Options

### Include Metadata

To include resource IDs and timestamps:

```json
{
  "applications": ["app-id"],
  "options": {
    "include_metadata": true
  }
}
```

### Custom Folder Structure

Organize exports by type:

```json
{
  "applications": ["*"],
  "options": {
    "folder_structure": {
      "group_by_type": true
    }
  }
}
```

### File Naming Patterns

Customize file naming:

```json
{
  "applications": ["*"],
  "options": {
    "folder_structure": {
      "file_naming_pattern": "${type}_${name}_${id}"
    }
  }
}
```

## Next Steps

- Learn how to [use exported configurations with Immutable Configuration Mode](./immutable-configuration.md)
- Review the [Application Management API](/api/application.yaml)
- Review the [Export API](/api/WIP/export.yaml)

## Troubleshooting

### Export Returns Empty Response

**Cause:** No applications found or invalid application IDs.

**Solution:**
1. Verify application IDs using `GET /applications`
2. Check that applications exist in the system
3. Review error messages in the export summary

### Parameterization Not Working

**Cause:** Field is not configured for parameterization.

**Solution:**
Currently, only OAuth `client_id`, `client_secret`, and `redirect_uris` are parameterized. Other fields will be exported as-is.

### Export API Returns 500 Error

**Cause:** Internal server error during export.

**Solution:**
1. Check Thunder server logs for detailed error messages
2. Verify the application data is valid
3. Ensure the application has all required fields

## Support

For issues or questions:
- **GitHub Issues:** [Report a bug](https://github.com/asgardeo/thunder/issues)
- **Documentation:** [Thunder Guides](/docs/guides/README.md)
