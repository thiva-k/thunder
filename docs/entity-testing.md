# Entity Unification — End-to-End Testing: App RBAC in Client Credentials Flow

## Prerequisites

- Thunder server running on `https://localhost:8090`
- Fresh database (bootstrap completed)
- Use `-k` flag for self-signed certs

All commands use the existing OU ID from the Default organization unit.

---

## Step 1: Create a Resource Server with Permissions

```bash
# Create resource server "Data API"
RS_RESPONSE=$(curl -sk -X POST https://localhost:8090/resource-servers \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Data API",
    "handle": "data-api",
    "description": "API for data operations"
  }')

echo "$RS_RESPONSE" | jq .
RS_ID=$(echo "$RS_RESPONSE" | jq -r '.id')
echo "Resource Server ID: $RS_ID"
```

```bash
# Create actions (permissions) on the resource server
curl -sk -X POST "https://localhost:8090/resource-servers/$RS_ID/actions" \
  -H "Content-Type: application/json" \
  -d '{"name": "Read Data", "handle": "read"}'

curl -sk -X POST "https://localhost:8090/resource-servers/$RS_ID/actions" \
  -H "Content-Type: application/json" \
  -d '{"name": "Write Data", "handle": "write"}'

curl -sk -X POST "https://localhost:8090/resource-servers/$RS_ID/actions" \
  -H "Content-Type: application/json" \
  -d '{"name": "Delete Data", "handle": "delete"}'

# Verify permissions exist
curl -sk "https://localhost:8090/resource-servers/$RS_ID/actions" | jq .
```

---

## Step 2: Create a Confidential M2M Application

```bash
# Create confidential application with client_credentials grant
APP_RESPONSE=$(curl -sk -X POST https://localhost:8090/applications \
  -H "Content-Type: application/json" \
  -d '{
    "name": "M2M Data Client",
    "description": "Machine-to-machine client for Data API",
    "inboundAuthConfig": [{
      "type": "oauth2",
      "oAuthAppConfig": {
        "grantTypes": ["client_credentials"],
        "tokenEndpointAuthMethod": "client_secret_basic"
      }
    }]
  }')

echo "$APP_RESPONSE" | jq .
APP_ID=$(echo "$APP_RESPONSE" | jq -r '.id')
CLIENT_ID=$(echo "$APP_RESPONSE" | jq -r '.inboundAuthConfig[0].oAuthAppConfig.clientId')
CLIENT_SECRET=$(echo "$APP_RESPONSE" | jq -r '.inboundAuthConfig[0].oAuthAppConfig.clientSecret')
echo "App ID (Entity ID): $APP_ID"
echo "Client ID: $CLIENT_ID"
echo "Client Secret: $CLIENT_SECRET"
```

### Verify entity was created in directory

```bash
# Check ENTITY table
sqlite3 backend/cmd/server/repository/database/userdb.db \
  "SELECT ENTITY_ID, ENTITY_CATEGORY, SYSTEM_ATTRIBUTES FROM ENTITY WHERE ENTITY_ID='$APP_ID';"

# Check ENTITY_IDENTIFIER table
sqlite3 backend/cmd/server/repository/database/userdb.db \
  "SELECT * FROM ENTITY_IDENTIFIER WHERE ENTITY_ID='$APP_ID';"

# Check SYSTEM_CREDENTIALS (should have hashed clientSecret)
sqlite3 backend/cmd/server/repository/database/userdb.db \
  "SELECT SYSTEM_CREDENTIALS FROM ENTITY WHERE ENTITY_ID='$APP_ID';"
```

---

## Step 3: Create a Role with a Subset of Permissions

Get the OU ID first:
```bash
OU_ID=$(sqlite3 backend/cmd/server/repository/database/userdb.db \
  "SELECT OU_ID FROM ORGANIZATION_UNIT LIMIT 1;")
echo "OU ID: $OU_ID"
```

```bash
# Create a role with only read + write (NOT delete)
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

echo "$ROLE_RESPONSE" | jq .
ROLE_ID=$(echo "$ROLE_RESPONSE" | jq -r '.id')
echo "Role ID: $ROLE_ID"
```

---

## Step 4: Assign the Role to the Application Entity

```bash
# Assign role to app entity (type: "app")
curl -sk -X POST "https://localhost:8090/roles/$ROLE_ID/assignments/add" \
  -H "Content-Type: application/json" \
  -d "{
    \"assignments\": [{
      \"id\": \"$APP_ID\",
      \"type\": \"app\"
    }]
  }"

# Verify assignment
curl -sk "https://localhost:8090/roles/$ROLE_ID/assignments?include=display" | jq .
```

---

## Step 5: Request Token with More Scopes Than Assigned

```bash
# Request token with read + write + delete scopes
# The app only has read + write via its role — delete should be filtered out
TOKEN_RESPONSE=$(curl -sk -X POST https://localhost:8090/oauth2/token \
  -u "$CLIENT_ID:$CLIENT_SECRET" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=client_credentials&scope=read write delete")

echo "$TOKEN_RESPONSE" | jq .
```

### Expected Result

The response should contain an access token with **only `read write`** scopes — `delete` should be filtered out by RBAC:

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
# Decode JWT payload (base64)
echo "$ACCESS_TOKEN" | cut -d. -f2 | base64 -d 2>/dev/null | jq .
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
TOKEN_RESPONSE2=$(curl -sk -X POST https://localhost:8090/oauth2/token \
  -u "$CLIENT_ID:$CLIENT_SECRET" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=client_credentials&scope=read write delete")

echo "$TOKEN_RESPONSE2" | jq .
```

### Expected Result

The token should have **empty scope** since the app has no role assignments:

```json
{
  "access_token": "eyJ...",
  "scope": ""
}
```

---

## Summary

| Step | What | Expected |
|------|------|----------|
| 1 | Create resource server with 3 permissions (read, write, delete) | Resource server created |
| 2 | Create confidential M2M app | ENTITY row + APPLICATION config + OAuth config created |
| 3 | Create role with 2 permissions (read, write) | Role with subset of permissions |
| 4 | Assign role to app entity | ROLE_ASSIGNMENT with type="app" |
| 5 | Token request with scope=read write delete | Token scope = "read write" (delete filtered) |
| 6 | Remove role, request token again | Token scope = "" (no permissions) |
