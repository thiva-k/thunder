#!/bin/bash
# Entity Unification — End-to-End Test: App RBAC in Client Credentials Flow
# Requires: THUNDER_SKIP_SECURITY=true on the server
set -e

BASE_URL="https://localhost:8090"
CURL="curl -sk"

echo "=========================================="
echo " Entity RBAC E2E Test"
echo "=========================================="

# Get OU ID (needed for resource server and role creation)
OU_ID=$(sqlite3 /Users/thiva/Desktop/Repos/thunder3/backend/cmd/server/repository/database/userdb.db \
  "SELECT OU_ID FROM ORGANIZATION_UNIT LIMIT 1;")
echo "OU ID: $OU_ID"

# Step 1: Create Resource Server with Permissions
echo ""
echo "--- Step 1: Create Resource Server with Permissions ---"
RS_RESPONSE=$($CURL -X POST "$BASE_URL/resource-servers" \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"Data API $(date +%s)\",
    \"identifier\": \"data-api-$(date +%s)\",
    \"description\": \"API for data operations\",
    \"ouId\": \"$OU_ID\"
  }")
RS_ID=$(echo "$RS_RESPONSE" | jq -r '.id')
echo "Resource Server ID: $RS_ID"

if [ "$RS_ID" = "null" ] || [ -z "$RS_ID" ]; then
  echo "FAILED: Could not create resource server"
  echo "$RS_RESPONSE" | jq .
  exit 1
fi

$CURL -X POST "$BASE_URL/resource-servers/$RS_ID/actions" \
  -H "Content-Type: application/json" \
  -d '{"name": "Read Data", "handle": "read"}' > /dev/null
$CURL -X POST "$BASE_URL/resource-servers/$RS_ID/actions" \
  -H "Content-Type: application/json" \
  -d '{"name": "Write Data", "handle": "write"}' > /dev/null
$CURL -X POST "$BASE_URL/resource-servers/$RS_ID/actions" \
  -H "Content-Type: application/json" \
  -d '{"name": "Delete Data", "handle": "delete"}' > /dev/null
echo "Created permissions: read, write, delete"

# Step 2: Create Full-Stack Confidential Application
echo ""
echo "--- Step 2: Create Application (auth_code + client_credentials) ---"
APP_RESPONSE=$($CURL -X POST "$BASE_URL/applications" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Data Service App '$(date +%s)'",
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

APP_ID=$(echo "$APP_RESPONSE" | jq -r '.id')
CLIENT_ID=$(echo "$APP_RESPONSE" | jq -r '.inboundAuthConfig[0].config.clientId')
CLIENT_SECRET=$(echo "$APP_RESPONSE" | jq -r '.inboundAuthConfig[0].config.clientSecret')

echo "App ID (Entity ID): $APP_ID"
echo "Client ID: $CLIENT_ID"
echo "Client Secret: ${CLIENT_SECRET:0:10}..."

if [ "$APP_ID" = "null" ] || [ -z "$APP_ID" ]; then
  echo "FAILED: Could not create application"
  echo "$APP_RESPONSE" | jq .
  exit 1
fi

# Verify entity in directory
echo ""
echo "--- Verify: Entity in Directory ---"
echo "ENTITY:"
sqlite3 /Users/thiva/Desktop/Repos/thunder3/backend/cmd/server/repository/database/userdb.db \
  "SELECT ENTITY_CATEGORY, SYSTEM_ATTRIBUTES FROM ENTITY WHERE ENTITY_ID='$APP_ID';"
echo "IDENTIFIERS:"
sqlite3 /Users/thiva/Desktop/Repos/thunder3/backend/cmd/server/repository/database/userdb.db \
  "SELECT TYPE || '=' || VALUE || ' (' || SOURCE || ')' FROM ENTITY_IDENTIFIER WHERE ENTITY_ID='$APP_ID';"
HAS_CREDS=$(sqlite3 /Users/thiva/Desktop/Repos/thunder3/backend/cmd/server/repository/database/userdb.db \
  "SELECT CASE WHEN SYSTEM_CREDENTIALS != '{}' AND SYSTEM_CREDENTIALS != '' THEN 'YES' ELSE 'NO' END FROM ENTITY WHERE ENTITY_ID='$APP_ID';")
echo "Has hashed credentials: $HAS_CREDS"
echo ""
echo "--- Verify: Config in Gateway (no name/description/clientId/secret) ---"
echo "APPLICATION columns: ID, AUTH_FLOW_ID, THEME_ID"
sqlite3 /Users/thiva/Desktop/Repos/thunder3/backend/cmd/server/repository/database/configdb.db \
  "SELECT ID, AUTH_FLOW_ID, THEME_ID FROM APPLICATION WHERE ID='$APP_ID';"
echo "OAUTH columns: ENTITY_ID, substr(OAUTH_CONFIG_JSON)"
sqlite3 /Users/thiva/Desktop/Repos/thunder3/backend/cmd/server/repository/database/configdb.db \
  "SELECT ENTITY_ID, substr(OAUTH_CONFIG_JSON, 1, 80) FROM APP_OAUTH_INBOUND_CONFIG WHERE ENTITY_ID='$APP_ID';"

# Step 3: Create Role with subset of permissions
echo ""
echo "--- Step 3: Create Role (read + write only, NOT delete) ---"
ROLE_RESPONSE=$($CURL -X POST "$BASE_URL/roles" \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"Data Reader Writer $(date +%s)\",
    \"description\": \"Can read and write, not delete\",
    \"ouId\": \"$OU_ID\",
    \"permissions\": [{
      \"resourceServerId\": \"$RS_ID\",
      \"permissions\": [\"read\", \"write\"]
    }]
  }")
ROLE_ID=$(echo "$ROLE_RESPONSE" | jq -r '.id')
echo "Role ID: $ROLE_ID"

if [ "$ROLE_ID" = "null" ] || [ -z "$ROLE_ID" ]; then
  echo "FAILED: Could not create role"
  echo "$ROLE_RESPONSE" | jq .
  exit 1
fi

# Step 4: Assign Role to App Entity
echo ""
echo "--- Step 4: Assign Role to App Entity ---"
ASSIGN_RESPONSE=$($CURL -X POST "$BASE_URL/roles/$ROLE_ID/assignments/add" \
  -H "Content-Type: application/json" \
  -d "{
    \"assignments\": [{
      \"id\": \"$APP_ID\",
      \"type\": \"app\"
    }]
  }")

echo "Assignments:"
$CURL "$BASE_URL/roles/$ROLE_ID/assignments?include=display" | jq '.assignments // .'

echo ""
echo "--- Verify: ROLE_ASSIGNMENT in DB ---"
sqlite3 /Users/thiva/Desktop/Repos/thunder3/backend/cmd/server/repository/database/configdb.db \
  "SELECT ROLE_ID, ASSIGNEE_TYPE, ASSIGNEE_ID FROM ROLE_ASSIGNMENT WHERE ASSIGNEE_ID='$APP_ID';"

# Step 5: Token Request — RBAC filters scopes
echo ""
echo "=========================================="
echo " Step 5: Client Credentials Token Request"
echo " Requesting: read write delete"
echo " Expected:   read write (delete filtered)"
echo "=========================================="

TOKEN_RESPONSE=$($CURL -X POST "$BASE_URL/oauth2/token" \
  -u "$CLIENT_ID:$CLIENT_SECRET" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=client_credentials&scope=read write delete")

echo "Token Response:"
echo "$TOKEN_RESPONSE" | jq .

SCOPE=$(echo "$TOKEN_RESPONSE" | jq -r '.scope // empty')
ACCESS_TOKEN=$(echo "$TOKEN_RESPONSE" | jq -r '.access_token // empty')

if [ -z "$ACCESS_TOKEN" ] || [ "$ACCESS_TOKEN" = "null" ]; then
  echo "FAIL: No access token returned"
  exit 1
fi

echo ""
echo "--- JWT Payload ---"
echo "$ACCESS_TOKEN" | cut -d. -f2 | tr '_-' '/+' | base64 -d 2>/dev/null | jq . 2>/dev/null || echo "(decode failed)"

echo ""
PASS=true
if echo "$SCOPE" | grep -q "delete"; then
  echo "FAIL: 'delete' was NOT filtered by RBAC"
  PASS=false
else
  echo "PASS: 'delete' filtered by RBAC"
fi
if echo "$SCOPE" | grep -q "read" && echo "$SCOPE" | grep -q "write"; then
  echo "PASS: 'read' and 'write' present"
else
  echo "WARN: Expected 'read write' but got: '$SCOPE'"
fi

# Step 6: Remove role → empty scopes
echo ""
echo "=========================================="
echo " Step 6: Remove Role → Token with No Scopes"
echo "=========================================="

$CURL -X POST "$BASE_URL/roles/$ROLE_ID/assignments/remove" \
  -H "Content-Type: application/json" \
  -d "{\"assignments\": [{\"id\": \"$APP_ID\", \"type\": \"app\"}]}" > /dev/null

TOKEN_RESPONSE2=$($CURL -X POST "$BASE_URL/oauth2/token" \
  -u "$CLIENT_ID:$CLIENT_SECRET" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=client_credentials&scope=read write delete")

SCOPE2=$(echo "$TOKEN_RESPONSE2" | jq -r '.scope // empty')
echo "Scope after role removal: '$SCOPE2'"

if [ -z "$SCOPE2" ] || [ "$SCOPE2" = "" ]; then
  echo "PASS: Empty scopes when no roles assigned"
else
  echo "WARN: Expected empty but got: '$SCOPE2'"
fi

# Cleanup
echo ""
echo "--- Cleanup ---"
$CURL -X DELETE "$BASE_URL/roles/$ROLE_ID" > /dev/null 2>&1
$CURL -X DELETE "$BASE_URL/applications/$APP_ID" > /dev/null 2>&1
$CURL -X DELETE "$BASE_URL/resource-servers/$RS_ID" > /dev/null 2>&1
echo "Cleaned up: role, app, resource server"

echo ""
echo "=========================================="
if [ "$PASS" = true ]; then
  echo " ALL TESTS PASSED ✓"
else
  echo " SOME TESTS FAILED ✗"
fi
echo "=========================================="
