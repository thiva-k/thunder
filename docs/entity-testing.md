# Entity Unification — End-to-End Testing: App RBAC in Client Credentials Flow

## Prerequisites

- Thunder server running on `https://localhost:8090` with `THUNDER_SKIP_SECURITY=true`
- Fresh database (bootstrap completed)
- `jq` installed for JSON parsing

Run the automated test: `bash docs/entity-test.sh`

Or follow the manual steps below.

---

## Step 1: Create a Resource Server with Permissions

```bash
# Get OU ID
OU_ID=$(curl -sk https://localhost:8090/organization-units | jq -r '.organizationUnits[0].id')
echo "OU ID: $OU_ID"

# Create resource server
RS_RESPONSE=$(curl -sk -X POST https://localhost:8090/resource-servers \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"Data API\",
    \"identifier\": \"data-api\",
    \"description\": \"API for data operations\",
    \"ouId\": \"$OU_ID\"
  }")
RS_ID=$(echo "$RS_RESPONSE" | jq -r '.id')
echo "Resource Server ID: $RS_ID"

# Create permissions
curl -sk -X POST "https://localhost:8090/resource-servers/$RS_ID/actions" \
  -H "Content-Type: application/json" -d '{"name": "Read Data", "handle": "read"}'

curl -sk -X POST "https://localhost:8090/resource-servers/$RS_ID/actions" \
  -H "Content-Type: application/json" -d '{"name": "Write Data", "handle": "write"}'

curl -sk -X POST "https://localhost:8090/resource-servers/$RS_ID/actions" \
  -H "Content-Type: application/json" -d '{"name": "Delete Data", "handle": "delete"}'
```

---

## Step 2: Create a Full-Stack Application (auth_code + client_credentials)

```bash
APP_RESPONSE=$(curl -sk -X POST https://localhost:8090/applications \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Data Service App",
    "description": "Full-stack app with M2M and interactive auth",
    "inboundAuthConfig": [{
      "type": "oauth2",
      "config": {
        "grantTypes": ["client_credentials", "authorization_code", "refresh_token"],
        "responseTypes": ["code"],
        "redirectUris": ["http://localhost:3000/callback"],
        "tokenEndpointAuthMethod": "client_secret_basic",
        "pkceRequired": true
      }
    }]
  }')

echo "$APP_RESPONSE" | jq .
APP_ID=$(echo "$APP_RESPONSE" | jq -r '.id')
CLIENT_ID=$(echo "$APP_RESPONSE" | jq -r '.inboundAuthConfig[0].config.clientId')
CLIENT_SECRET=$(echo "$APP_RESPONSE" | jq -r '.inboundAuthConfig[0].config.clientSecret')
echo "App ID (Entity ID): $APP_ID"
echo "Client ID: $CLIENT_ID"
echo "Client Secret: $CLIENT_SECRET"
```

### Verify application was created

```bash
curl -sk "https://localhost:8090/applications/$APP_ID" | jq '{name, inboundAuthConfig}'
```

---

## Step 3: Create a Role with a Subset of Permissions

```bash
OU_ID=$(curl -sk https://localhost:8090/organization-units | jq -r '.organizationUnits[0].id')

ROLE_RESPONSE=$(curl -sk -X POST https://localhost:8090/roles \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"Data Reader Writer\",
    \"description\": \"Can read and write data, but not delete\",
    \"ouId\": \"$OU_ID\",
    \"permissions\": [{
      \"resourceServerId\": \"$RS_ID\",
      \"permissions\": [\"read\", \"write\"]
    }]
  }")

ROLE_ID=$(echo "$ROLE_RESPONSE" | jq -r '.id')
echo "Role ID: $ROLE_ID"
```

---

## Step 4: Assign the Role to the Application Entity

```bash
curl -sk -X POST "https://localhost:8090/roles/$ROLE_ID/assignments/add" \
  -H "Content-Type: application/json" \
  -d "{
    \"assignments\": [{
      \"id\": \"$APP_ID\",
      \"type\": \"app\"
    }]
  }"

# Verify assignment — should show app with display name
curl -sk "https://localhost:8090/roles/$ROLE_ID/assignments?include=display" | jq '.assignments'
```

---

## Step 5: Request Token with More Scopes Than Assigned

```bash
TOKEN_RESPONSE=$(curl -sk -X POST https://localhost:8090/oauth2/token \
  -u "$CLIENT_ID:$CLIENT_SECRET" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=client_credentials&scope=read write delete")

echo "$TOKEN_RESPONSE" | jq .
```

### Expected Result

The token should contain **only `read write`** — `delete` is filtered by RBAC:

```json
{
  "access_token": "eyJ...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "scope": "read write"
}
```

### Decode the token to verify

```bash
ACCESS_TOKEN=$(echo "$TOKEN_RESPONSE" | jq -r '.access_token')
echo "$ACCESS_TOKEN" | cut -d. -f2 | tr '_-' '/+' | base64 -d 2>/dev/null | jq .
```

The `scope` claim in the JWT should be `"read write"` (not `"read write delete"`).

---

## Step 6: Test with No Role Assignments

```bash
# Remove the role assignment
curl -sk -X POST "https://localhost:8090/roles/$ROLE_ID/assignments/remove" \
  -H "Content-Type: application/json" \
  -d "{
    \"assignments\": [{
      \"id\": \"$APP_ID\",
      \"type\": \"app\"
    }]
  }"

# Request token again — should get empty scopes
curl -sk -X POST https://localhost:8090/oauth2/token \
  -u "$CLIENT_ID:$CLIENT_SECRET" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=client_credentials&scope=read write delete" | jq .
```

### Expected Result

The token should have **empty scope** since the app has no role assignments.

---

## Summary

| Step | What | Expected |
|------|------|----------|
| 1 | Create resource server with 3 permissions (read, write, delete) | Resource server created |
| 2 | Create full-stack app (auth_code + client_credentials) | Entity + config created |
| 3 | Create role with 2 permissions (read, write) | Role with subset of permissions |
| 4 | Assign role to app entity (type: "app") | Assignment with display name |
| 5 | Token request with scope=read write delete | Token scope = "read write" (delete filtered) |
| 6 | Remove role, request token again | Token scope = "" (no permissions) |
