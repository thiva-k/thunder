# Thunder OAuth 2.0 / OIDC Component — Threat Model

**Version:** v1  
**Date:** 2025-02-24  
**Email:** security@wso2.com  

---

## Revision History

| Version | Release Date | Contributors / Authors | Summary of Changes |
|---------|-------------|----------------------|-------------------|
| v1 | 2025-02-24 | Thunder Team | Initial version |

---

## Introduction

Thunder is a lightweight user and identity management product providing authentication and authorization capabilities. This threat model covers the **OAuth 2.0 / OpenID Connect (OIDC)** component of Thunder, which implements the authorization server functionality including token issuance, authorization code flows, client authentication, token introspection, JWKS, OIDC discovery, Dynamic Client Registration (DCR), and UserInfo endpoints.

The OAuth component is implemented in Go under `backend/internal/oauth/` and follows RFC 6749 (OAuth 2.0), RFC 7636 (PKCE), RFC 7662 (Token Introspection), RFC 8414 (Authorization Server Metadata), RFC 8693 (Token Exchange), OpenID Connect Core 1.0, and RFC 9700 (OAuth 2.0 Security Best Current Practice).

**Associated Design Documentation:**
- [OAuth2 API specification](../api/authentication.yaml)
- [Flow Execution API](../api/flow-execution.yaml)
- [Application API](../api/application.yaml)

---

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        EXTERNAL / UNTRUSTED                             │
│                                                                         │
│  ┌──────────────┐     ┌──────────────────┐     ┌─────────────────┐     │
│  │  End User     │     │  OAuth Client     │     │  Resource Server │     │
│  │  (Browser)    │     │  (Relying Party)  │     │  (API Consumer)  │     │
│  └──────┬───────┘     └───────┬──────────┘     └────────┬────────┘     │
│         │                     │                          │              │
└─────────┼─────────────────────┼──────────────────────────┼──────────────┘
          │ [B-EX-EN]           │ [B-EX-EN]                │ [B-EX-EN]
          │ [M-NT]              │ [M-NT]                   │ [M-NT]
          │ [C-High]            │ [C-High]                 │ [C-Medium]
══════════╪═════════════════════╪══════════════════════════╪═══════════════
          │              TLS BOUNDARY                       │
┌─────────┼─────────────────────┼──────────────────────────┼──────────────┐
│         ▼                     ▼                          ▼              │
│  ┌──────────────────────────────────────────────────────────────────┐   │
│  │                    THUNDER SERVER (Go)                            │   │
│  │                                                                  │   │
│  │  ┌────────────────────────────────────────────────────────────┐  │   │
│  │  │                  OAuth 2.0 / OIDC Module                   │  │   │
│  │  │                                                            │  │   │
│  │  │  ┌──────────────┐  ┌──────────────┐  ┌────────────────┐   │  │   │
│  │  │  │ Authorization │  │   Token      │  │  Client Auth   │   │  │   │
│  │  │  │  Endpoint     │  │  Endpoint    │  │  Middleware     │   │  │   │
│  │  │  │ GET /oauth2/  │  │ POST /oauth2/│  │ (Basic/Post/   │   │  │   │
│  │  │  │   authorize   │  │   token      │  │  None)         │   │  │   │
│  │  │  └──────┬───────┘  └──────┬───────┘  └────────────────┘   │  │   │
│  │  │         │                  │                                │  │   │
│  │  │  ┌──────┴───────┐  ┌──────┴───────┐  ┌────────────────┐   │  │   │
│  │  │  │ Auth Code     │  │ Grant        │  │ Token Builder/ │   │  │   │
│  │  │  │ Service       │  │ Handlers     │  │ Validator      │   │  │   │
│  │  │  │ (PKCE, State) │  │ (AuthZ Code, │  │ (JWT/RS256)    │   │  │   │
│  │  │  │               │  │  CC, Refresh,│  │                │   │  │   │
│  │  │  │               │  │  TokenExch)  │  │                │   │  │   │
│  │  │  └──────────────┘  └──────────────┘  └────────────────┘   │  │   │
│  │  │                                                            │  │   │
│  │  │  ┌──────────────┐  ┌──────────────┐  ┌────────────────┐   │  │   │
│  │  │  │ JWKS         │  │ Introspection│  │ UserInfo       │   │  │   │
│  │  │  │ GET /oauth2/ │  │ POST /oauth2/│  │ GET|POST       │   │  │   │
│  │  │  │   jwks       │  │  introspect  │  │  /oauth2/      │   │  │   │
│  │  │  │              │  │              │  │   userinfo     │   │  │   │
│  │  │  └──────────────┘  └──────────────┘  └────────────────┘   │  │   │
│  │  │                                                            │  │   │
│  │  │  ┌──────────────┐  ┌──────────────┐  ┌────────────────┐   │  │   │
│  │  │  │ Discovery    │  │ DCR          │  │ Scope          │   │  │   │
│  │  │  │ /.well-known/│  │ POST /oauth2/│  │ Validator      │   │  │   │
│  │  │  │              │  │  dcr/register│  │                │   │  │   │
│  │  │  └──────────────┘  └──────────────┘  └────────────────┘   │  │   │
│  │  └────────────────────────────────────────────────────────────┘  │   │
│  │                                                                  │   │
│  │  ┌────────────┐  ┌──────────────┐  ┌──────────────┐             │   │
│  │  │ Application │  │ Flow Engine  │  │ User Service │  [B-EX-CP] │   │
│  │  │ Service     │  │ (AuthN)      │  │              │  [M-IN]    │   │
│  │  └──────┬─────┘  └──────────────┘  └──────────────┘             │   │
│  │         │                                                        │   │
│  └─────────┼────────────────────────────────────────────────────────┘   │
│            │ [M-DB]                                                     │
│            ▼                                                            │
│  ┌──────────────────────────────────────────────────────────────────┐   │
│  │                    Database Layer                                 │   │
│  │  ┌──────────┐  ┌───────────┐  ┌───────────┐                     │   │
│  │  │ Thunder  │  │  Runtime  │  │   User    │                     │   │
│  │  │   DB     │  │    DB     │  │    DB     │                     │   │
│  │  │(Apps,    │  │(Auth Codes│  │(User      │                     │   │
│  │  │ Clients) │  │ Auth Reqs)│  │ Attrs)    │                     │   │
│  │  └──────────┘  └───────────┘  └───────────┘                     │   │
│  └──────────────────────────────────────────────────────────────────┘   │
│                                                                         │
│  ┌──────────────────────────────────────────────────────────────────┐   │
│  │                  PKI / Crypto Layer                               │   │
│  │  Signing Keys (RSA), Encryption Keys, TLS Certificates           │   │
│  │  [M-FS] [C-High]                                                 │   │
│  └──────────────────────────────────────────────────────────────────┘   │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## Data Flow / Sequence Diagrams

### DF-1: Authorization Code Grant with PKCE

```
End User          OAuth Client           Thunder Auth Endpoint       Flow Engine        Thunder Token Endpoint
(Browser)         (Relying Party)        GET /oauth2/authorize       (AuthN/MFA)        POST /oauth2/token
   │                    │                        │                       │                      │
   │  1. Click Login    │                        │                       │                      │
   │──────────────────>│                        │                       │                      │
   │                    │  2. Generate PKCE      │                       │                      │
   │                    │     code_verifier +    │                       │                      │
   │                    │     code_challenge     │                       │                      │
   │                    │                        │                       │                      │
   │                    │  3. GET /oauth2/authorize                      │                      │
   │                    │     ?response_type=code │                       │                      │
   │                    │     &client_id=...      │                       │                      │
   │                    │     &redirect_uri=...   │                       │                      │
   │                    │     &scope=openid       │                       │                      │
   │                    │     &state=...          │                       │                      │
   │                    │     &code_challenge=... │                       │                      │
   │                    │     &code_challenge_method=S256                │                      │
   │                    │──────────────────────>│                       │                      │
   │                    │                        │ 4. Validate request   │                      │
   │                    │                        │    (redirect_uri,     │                      │
   │                    │                        │     client, PKCE)     │                      │
   │                    │                        │                       │                      │
   │                    │                        │ 5. Store auth request │                      │
   │                    │                        │    context (authId)   │                      │
   │                    │                        │                       │                      │
   │  6. 302 Redirect to Login UI (gate)         │                       │                      │
   │<─────────────────────────────────────────────│                       │                      │
   │                    │                        │                       │                      │
   │  7. User authenticates via Flow Engine        │                       │                      │
   │──────────────────────────────────────────────────────────────────>│                      │
   │                    │                        │                       │                      │
   │                    │                        │ 8. Flow engine returns │                      │
   │                    │                        │    JWT assertion      │                      │
   │                    │                        │<──────────────────────│                      │
   │                    │                        │                       │                      │
   │                    │                        │ 9. Verify assertion,  │                      │
   │                    │                        │    generate auth code │                      │
   │                    │                        │    (UUIDv7), store    │                      │
   │                    │                        │    with PKCE + state  │                      │
   │                    │                        │                       │                      │
   │  10. 302 Redirect to client                 │                       │                      │
   │      ?code=...&state=...                    │                       │                      │
   │<─────────────────────────────────────────────│                       │                      │
   │──────────────────>│                        │                       │                      │
   │                    │                        │                       │                      │
   │                    │  11. POST /oauth2/token │                       │                      │
   │                    │      grant_type=authorization_code              │                      │
   │                    │      code=...           │                       │                      │
   │                    │      code_verifier=...  │                       │                      │
   │                    │      redirect_uri=...   │                       │                      │
   │                    │      + Client Auth     │                       │                      │
   │                    │──────────────────────────────────────────────────────────────────────>│
   │                    │                        │                       │                      │
   │                    │                        │                       │     12. Validate     │
   │                    │                        │                       │   client auth,       │
   │                    │                        │                       │   retrieve auth code │
   │                    │                        │                       │   (single-use),      │
   │                    │                        │                       │   verify PKCE,       │
   │                    │                        │                       │   verify redirect_uri│
   │                    │                        │                       │                      │
   │                    │  13. Token Response     │                       │                      │
   │                    │      {access_token, refresh_token, id_token}    │                      │
   │                    │<─────────────────────────────────────────────────────────────────────│
```

### DF-2: Client Credentials Grant

```
OAuth Client                 Client Auth Middleware           Thunder Token Endpoint
   │                                │                                │
   │  1. POST /oauth2/token         │                                │
   │     grant_type=client_credentials                               │
   │     + Authorization: Basic     │                                │
   │       base64(client_id:secret) │                                │
   │──────────────────────────────>│                                │
   │                                │ 2. Extract & validate          │
   │                                │    client credentials          │
   │                                │                                │
   │                                │ 3. Forward to token endpoint   │
   │                                │──────────────────────────────>│
   │                                │                                │ 4. Validate grant type
   │                                │                                │    allowed for client
   │                                │                                │ 5. Build access token (JWT)
   │                                │                                │    with client_id as sub
   │                                │                                │
   │  6. Token Response              │                                │
   │     {access_token, token_type,  │                                │
   │      expires_in}               │                                │
   │<──────────────────────────────────────────────────────────────│
```

### DF-3: Refresh Token Grant

```
OAuth Client                 Client Auth Middleware           Thunder Token Endpoint
   │                                │                                │
   │  1. POST /oauth2/token         │                                │
   │     grant_type=refresh_token    │                                │
   │     refresh_token=...           │                                │
   │     + Client Auth              │                                │
   │──────────────────────────────>│                                │
   │                                │ 2. Validate client             │
   │                                │──────────────────────────────>│
   │                                │                                │ 3. Validate & decode
   │                                │                                │    refresh token JWT
   │                                │                                │ 4. Verify client_id match
   │                                │                                │ 5. Intersect requested
   │                                │                                │    scopes with original
   │                                │                                │ 6. Build new access token
   │                                │                                │ 7. Optionally rotate
   │                                │                                │    refresh token
   │  8. Token Response              │                                │
   │     {access_token,              │                                │
   │      refresh_token (if rotated)}│                                │
   │<──────────────────────────────────────────────────────────────│
```

---

## Actors and Resources

### Actors

| Actor (Role) | Description | Roles or Permissions |
|---|---|---|
| **End User** | A resource owner who authenticates via the browser to authorize client access to their resources | Authenticate, authorize/deny consent, access own resources |
| **OAuth Client (Confidential)** | A server-side application registered with a client secret that requests tokens on behalf of users or itself | Authorization code grant, client credentials grant, refresh token grant, token exchange |
| **OAuth Client (Public)** | A client-side application (SPA, native app) that cannot securely store client secrets | Authorization code + PKCE grant, refresh token grant (sender-constrained) |
| **Resource Server** | An API server that validates access tokens to protect resources | Introspect tokens, validate JWT signatures via JWKS |
| **Admin** | A privileged user who registers/manages OAuth applications via the management API | Register applications, configure grant types, manage redirect URIs, configure PKCE requirements |
| **Malicious Actor (Web Attacker - A1)** | Per RFC 9700 Section 3: Can set up and operate arbitrary web servers, create arbitrary redirect URIs, manipulate browser-based flows | Phishing, authorization code interception, token theft |
| **Malicious Actor (Network Attacker - A2)** | Per RFC 9700 Section 3: Full control over the network (eavesdrop, intercept, modify traffic) | Man-in-the-middle, token interception, TLS downgrade |

### Entitlement Matrix

| Actor | Initiate Auth Code Flow | Exchange Auth Code for Token | Use Client Credentials | Refresh Tokens | Introspect Tokens | Register Clients (DCR) | Access JWKS/Discovery | Access UserInfo |
|---|---|---|---|---|---|---|---|---|
| End User | Yes (via browser) | No | No | No | No | No | Yes | No |
| OAuth Client (Confidential) | Yes | Yes (with client_secret) | Yes | Yes | No | Yes | Yes | Yes (with access_token) |
| OAuth Client (Public) | Yes (with PKCE) | Yes (with code_verifier) | No | Yes | No | Yes | Yes | Yes (with access_token) |
| Resource Server | No | No | No | No | Yes | No | Yes | No |
| Admin | No (uses management APIs) | No | No | No | No | Yes (via management API) | Yes | No |

### Resources

| Asset | Description |
|---|---|
| **Authorization Endpoint** (`GET /oauth2/authorize`) | Initiates the authorization code flow; validates client, redirect URI, PKCE parameters; stores auth request context in runtime DB |
| **Token Endpoint** (`POST /oauth2/token`) | Issues access tokens, refresh tokens, ID tokens; protected by client authentication middleware |
| **JWKS Endpoint** (`GET /oauth2/jwks`) | Exposes public keys (RSA, ECDSA, EdDSA) for token verification; publicly accessible |
| **Introspection Endpoint** (`POST /oauth2/introspect`) | Validates tokens and returns active/inactive status with claims |
| **UserInfo Endpoint** (`GET/POST /oauth2/userinfo`) | Returns OIDC user claims; requires valid access token with `openid` scope |
| **Discovery Endpoints** (`/.well-known/openid-configuration`, `/.well-known/oauth-authorization-server`) | Publishes server metadata per RFC 8414 / OIDC Discovery |
| **DCR Endpoint** (`POST /oauth2/dcr/register`) | Allows dynamic client registration |
| **Authorization Codes** | Short-lived (configurable), single-use, stored in runtime DB with PKCE challenge and redirect URI |
| **Access Tokens** | JWT tokens signed with RS256; contain `sub`, `aud`, `iss`, `scope`, `client_id`, `grant_type` claims |
| **Refresh Tokens** | JWT tokens containing metadata to re-derive access tokens; support rotation via `renew_on_grant` config |
| **ID Tokens** | OIDC JWT tokens with user claims (`auth_time`, profile attributes) |
| **PKI Signing Keys** | RSA private keys on filesystem for JWT signing; public keys exposed via JWKS |
| **Encryption Key** | Symmetric key for data encryption at rest |
| **Client Secrets** | Stored in application configuration; used for `client_secret_basic` and `client_secret_post` authentication |
| **Runtime Database** | SQLite/PostgreSQL database storing authorization codes, auth request contexts |
| **Identity Database** | Stores application configurations, client registrations |
| **User Database** | Stores user profiles, credentials, attributes |

### Dependencies

| Dependency | Description |
|---|---|
| **Application Service** (`internal/application`) | Retrieves OAuth application configuration, validates client credentials, redirect URIs, allowed grant types. OAuth module depends on this for all client validation. |
| **Flow Engine** (`internal/flow/flowexec`) | Handles user authentication (password, passwordless, MFA, social login). OAuth authorization endpoint delegates authentication to this service and receives a JWT assertion upon completion. |
| **User Service** (`internal/user`) | Provides user profile attributes for populating token claims (access token, ID token, UserInfo). |
| **PKI Service** (`internal/system/crypto/pki`) | Manages X.509 certificates and private keys for JWT signing and JWKS. |
| **JWT Service** (`internal/system/jose/jwt`) | Generates and verifies JWT tokens used for access tokens, refresh tokens, ID tokens, and flow engine assertions. |
| **Database Layer** (`internal/system/database`) | SQLite (default) / PostgreSQL for persistent storage. |
| **TLS Layer** | Provides transport-level encryption; minimum TLS 1.3 configured in deployment.yaml. |
| **Gate UI** (`frontend/apps/thunder-gate`) | Login, registration, recovery UI; user authenticates via this frontend before authorization code is issued. |

---

## Trust Boundaries

| ID | Interaction Type | Interaction |
|---|---|---|
| 1 | **Untrust → Trust** | End user's browser sends authorization request to `/oauth2/authorize` over the internet |
| 2 | **Untrust → Trust** | OAuth client sends token request to `/oauth2/token` over the internet |
| 3 | **Untrust → Trust** | OAuth client sends introspection request to `/oauth2/introspect` over the internet |
| 4 | **Untrust → Trust** | OAuth client sends DCR registration request to `/oauth2/dcr/register` over the internet |
| 5 | **Untrust → Trust** | Resource server or client fetches JWKS from `/oauth2/jwks` over the internet |
| 6 | **Untrust → Trust** | OAuth client accesses UserInfo endpoint with bearer token |
| 7 | **Trust → Trust** | OAuth module invokes Application Service (internal Go function call) |
| 8 | **Trust → Trust** | OAuth module invokes User Service (internal Go function call) |
| 9 | **Trust → Trust** | OAuth module invokes Flow Engine via internal interface |
| 10 | **Internal** | OAuth module reads/writes to runtime database (authorization codes, auth request contexts) |
| 11 | **Internal** | OAuth module accesses PKI signing keys from the filesystem |
| 12 | **Trust → Untrust** | Thunder server redirects user agent to client's redirect URI with authorization code in query parameter |
| 13 | **Untrust → Trust** | End user's browser submits authentication flow callbacks to `/oauth2/auth/callback` |

---

## Threats and Mitigations

### Inherited or Out-Scope Risks

- Threats related to a malicious entity being able to manipulate the local network of the user (DNS/BGP/etc) to intercept or manipulate payloads — mitigated by mandatory TLS 1.3.
- Threats to the operating system, container runtime, or physical infrastructure hosting Thunder.
- Threats to the Flow Engine (authentication mechanisms) — covered separately by the authentication threat model.
- Threats to the management API for application CRUD operations — covered separately.
- Threats related to end-of-life or end-of-service components in the Go runtime or OS.
- Denial of Service attacks at the network/infrastructure level (L3/L4 DDoS).

---

### Interactions

---

#### [I-01]: Authorization Request (End User → Authorization Endpoint)

**Description**  
The end user's browser (redirected by the OAuth client) sends a GET request to `/oauth2/authorize` with OAuth parameters (`response_type`, `client_id`, `redirect_uri`, `scope`, `state`, `code_challenge`, `code_challenge_method`). Thunder validates the request, stores the authorization request context in the runtime database, and redirects the user to the Gate login UI.

**Assets Involved**

| Initiator | Intermediate | Target |
|---|---|---|
| End User (Browser) | OAuth Client (triggers redirect) | Thunder Authorization Endpoint |

**Data Flow**  
See DF-1 steps 1-6.

**Access Control**  
No authentication required. The endpoint is publicly accessible. Validation is performed on `client_id`, `redirect_uri` (exact match against registered URIs), and supported `response_type`.

**Security Considerations**

| Area | Response | Comments |
|---|---|---|
| Data Confidentiality | High confidential [C-High] | Contains client_id, redirect_uri, scope, PKCE challenge, state — leakage enables authorization code injection or CSRF |
| Communication Medium | Network interaction [M-NT] | HTTPS from browser to server |
| Transport Security | TLS 1.3 Encryption | Configured as minimum TLS 1.3 in deployment.yaml |
| Authentication | No Authentication | Public endpoint; client is identified by `client_id` |
| Accessibility | Publicly Accessible | Must be reachable by end user browsers |

**Threat Assessment**

| ID | Category | Threat | Materializable | Mitigations / Comment |
|---|---|---|---|---|
| 1 | Spoofing | **Insufficient Redirect URI Validation** (RFC 9700 §4.1): Attacker crafts an authorization request with a manipulated `redirect_uri` pointing to attacker-controlled server to steal authorization code. | No | Thunder uses **exact string matching** for redirect URI validation against pre-registered URIs per RFC 9700 §2.1. No pattern matching or wildcards are supported. Redirect URI is stored at authorization time and re-verified at token exchange. |
| 2 | Tampering | **Authorization Code Injection** (RFC 9700 §4.5): Attacker obtains an authorization code and injects it into a legitimate client's redirect URI callback to impersonate the victim. | No | **PKCE (RFC 7636)** is supported and enforced for public clients. PKCE `code_challenge` is stored with the authorization code and `code_verifier` is validated at token exchange. For confidential clients, `client_secret` provides additional binding. Per RFC 9700 §2.1.1, authorization servers MUST support PKCE. |
| 3 | Information Disclosure | **Authorization Code Leakage via Referer Headers** (RFC 9700 §4.2): If the client page loads external resources after receiving the authorization code, the code may leak via Referer headers. | No | Authorization codes are **single-use** — deactivated immediately upon retrieval at the token endpoint. Codes have a short, configurable validity period. Even if leaked via Referer, PKCE binding prevents the attacker from exchanging the code without the `code_verifier`. |
| 4 | Information Disclosure | **Authorization Code Leakage via Browser History** (RFC 9700 §4.3): The authorization code appears in the redirect URL query parameters and is stored in browser history. | No | Authorization codes are single-use and short-lived. PKCE ensures the code cannot be exchanged without the verifier, which is never exposed in the URL. Per RFC 9700 §4.3.1. |
| 5 | Spoofing | **Cross-Site Request Forgery (CSRF)** (RFC 9700 §4.7): Attacker tricks the user into completing an authorization flow initiated by the attacker, linking the attacker's resource to the victim's client session. | No | The `state` parameter is preserved throughout the authorization flow. Clients that support PKCE can rely on PKCE for CSRF protection per RFC 9700 §2.1. The `state` parameter is stored in the authorization code and returned in the redirect, allowing the client to verify it. |
| 6 | Spoofing | **PKCE Downgrade Attack** (RFC 9700 §4.8): Attacker strips `code_challenge` from authorization request or `code_verifier` from token request to bypass PKCE protection. | No | When PKCE is required for an application (e.g., public clients), Thunder enforces that `code_challenge` must be present in the authorization request. At the token endpoint, if the stored authorization code contains a PKCE challenge, `code_verifier` must be provided and validated. Per RFC 9700 §4.8.2. |
| 7 | Information Disclosure | **Credential Leakage in Authorization Request URL** (RFC 9700 §4.12): If the authorization server issues a 307 redirect, user credentials submitted to the authorization endpoint could be forwarded to the redirect URI. | No | Thunder uses HTTP 302 redirects (not 307) for authorization responses, preventing credential forwarding. Per RFC 9700 §4.12. |
| 8 | Tampering | **Mix-Up Attack** (RFC 9700 §4.4): When a client interacts with multiple authorization servers, an attacker-controlled AS can cause the client to send the authorization code to the attacker's token endpoint. | No | Thunder publishes **OAuth Authorization Server Metadata** at `/.well-known/oauth-authorization-server` per RFC 8414, enabling clients to verify server identity. The `iss` (issuer) claim in tokens also allows validation. Per RFC 9700 §2.1 and §4.4.2. |
| 9 | Elevation of Privilege | **Open Redirector via Authorization Endpoint** (RFC 9700 §4.11.2): If the authorization endpoint redirects to unvalidated URIs, it becomes an open redirector that attackers can abuse for phishing. | No | Redirect URIs are validated against the exact pre-registered URIs. If the `redirect_uri` is invalid, an error page is shown instead of redirecting. Per RFC 9700 §4.11.2. |
| 10 | Information Disclosure | **Authorization Request Context Leakage**: Authorization request context (containing all OAuth parameters including PKCE challenge) is stored in the runtime database. | No | Authorization request contexts are stored with a **10-minute expiry** and are deleted immediately after retrieval (single-use). The `authId` used to reference the context is a non-guessable UUID. The database is a local resource within the trust boundary. |

---

#### [I-02]: Authorization Callback (Flow Engine → Auth Callback Endpoint)

**Description**  
After the user completes authentication via the Flow Engine, a JWT assertion is sent to `POST /oauth2/auth/callback`. Thunder verifies the assertion, retrieves the stored authorization request context, generates an authorization code (UUIDv7), stores it with PKCE/state/scope data, and redirects the user back to the client's redirect URI with the code and state.

**Assets Involved**

| Initiator | Intermediate | Target |
|---|---|---|
| Gate UI (Browser) | Flow Engine (JWT assertion) | Thunder Authorization Callback Endpoint |

**Data Flow**  
See DF-1 steps 7-10.

**Access Control**  
The callback verifies a JWT assertion generated by the Flow Engine. The assertion is signed by the server's PKI keys and verified using the JWT service.

**Security Considerations**

| Area | Response | Comments |
|---|---|---|
| Data Confidentiality | High confidential [C-High] | Contains JWT assertion with user identity; generates authorization code |
| Communication Medium | Network interaction [M-NT] | HTTPS POST from browser to server |
| Transport Security | TLS 1.3 Encryption | |
| Authentication | JWT Assertion Verification | Flow engine assertion signed and verified by internal PKI |
| Accessibility | Publicly Accessible | Must be reachable by browser after authentication flow |

**Threat Assessment**

| ID | Category | Threat | Materializable | Mitigations / Comment |
|---|---|---|---|---|
| 1 | Spoofing | **Forged JWT Assertion**: Attacker crafts a fake JWT assertion to obtain an authorization code without actual authentication. | No | Assertions are **signed with the server's private RSA key** and verified using `jwtService.VerifyJWT()`. The private key is stored on the server filesystem and never exposed. |
| 2 | Replay | **JWT Assertion Replay**: Attacker captures a valid JWT assertion and replays it to obtain additional authorization codes. | No | Authorization request context is **single-use** (deleted after callback retrieval). The `authId` parameter links the assertion to a specific authorization request context, which can only be consumed once. |
| 3 | Tampering | **Sub Claim Constraint Bypass**: Attacker manipulates the assertion to change the authenticated user's identity. | No | When OIDC `claims` parameter contains a `sub` constraint, Thunder validates that the authenticated user's subject matches the requested constraint. The JWT assertion is cryptographically signed and cannot be modified. |

---

#### [I-03]: Token Request (OAuth Client → Token Endpoint)

**Description**  
The OAuth client sends a POST request to `/oauth2/token` to exchange an authorization code, client credentials, or refresh token for access/refresh/ID tokens. The request passes through the Client Authentication Middleware before reaching the Token Service.

**Assets Involved**

| Initiator | Intermediate | Target |
|---|---|---|
| OAuth Client | Client Auth Middleware | Thunder Token Endpoint |

**Data Flow**  
See DF-1 steps 11-13, DF-2, and DF-3.

**Access Control**  
Client authentication is enforced via the `ClientAuthMiddleware`:
- `client_secret_basic`: HTTP Basic authentication header
- `client_secret_post`: `client_id` and `client_secret` in the request body
- `none`: Only `client_id` required (for public clients)

The middleware validates that the authentication method is allowed for the specific client and that credentials are valid. Dual credential submission (both header and body) is rejected.

**Security Considerations**

| Area | Response | Comments |
|---|---|---|
| Data Confidentiality | High confidential [C-High] | Contains client credentials, authorization codes, refresh tokens; returns access/refresh/ID tokens |
| Communication Medium | Network interaction [M-NT] | HTTPS POST from client server to Thunder |
| Transport Security | TLS 1.3 Encryption | Client secrets transmitted over TLS only |
| Authentication | Client Secret (Basic/Post) or None | Per-application configurable authentication method |
| Accessibility | Publicly Accessible | Must be reachable by OAuth clients |

**Threat Assessment**

| ID | Category | Threat | Materializable | Mitigations / Comment |
|---|---|---|---|---|
| 1 | Spoofing | **Client Impersonation**: Attacker obtains or guesses client credentials to impersonate a legitimate client and exchange stolen authorization codes. | No | Client secrets are validated against stored credentials. For public clients, PKCE provides client binding via `code_verifier`. Dual credential submission is rejected to prevent confused deputy attacks. Per RFC 9700 §2.5, asymmetric cryptography is recommended for client authentication — currently `client_secret_basic/post` is supported; `private_key_jwt` should be considered. |
| 2 | Tampering | **Authorization Code Replay**: Attacker reuses a previously valid authorization code at the token endpoint. | No | Authorization codes are **single-use**: they are deactivated (set to INACTIVE) immediately upon first retrieval. Expired codes are rejected based on configurable validity period. |
| 3 | Spoofing | **Redirect URI Mismatch at Token Endpoint**: Attacker uses a different redirect URI at the token endpoint than was used during authorization to redirect tokens to an attacker-controlled URI. | No | The `redirect_uri` provided at the token endpoint is compared **exactly** against the `redirect_uri` stored with the authorization code. Mismatch results in rejection. Per RFC 9700 §4.1.3. |
| 4 | Spoofing | **PKCE Code Verifier Brute Force**: Attacker attempts to guess the `code_verifier` for a stolen authorization code. | No | PKCE `code_verifier` must be 43-128 characters from the ASCII unreserved character set. With S256, the verifier is SHA-256 hashed and base64url-encoded, making brute force computationally infeasible. Per RFC 9700 §2.1.1, S256 is the recommended method. |
| 5 | Information Disclosure | **Token Response Caching**: Intermediary caches or proxies may cache token responses, exposing tokens. | No | Token responses include `Cache-Control: no-store` and `Pragma: no-cache` headers, preventing caching. Per RFC 6749 §5.1. |
| 6 | Elevation of Privilege | **Scope Escalation on Refresh**: Attacker uses a refresh token to request broader scopes than originally granted. | No | Refresh token grant performs **scope intersection** — requested scopes are intersected with the scopes in the original refresh token. Broader scopes are silently reduced. |
| 7 | Denial of Service | **Token Endpoint Abuse**: Attacker sends large volumes of token requests to exhaust server resources or bruteforce client secrets. | Yes | Currently no rate limiting is explicitly implemented on the token endpoint. **Recommendation**: Implement rate limiting on the token endpoint, especially for failed authentication attempts. Consider using `internal/system/middleware` or an external rate limiter. |
| 8 | Information Disclosure | **Client Secret Leakage via Logs**: Client secrets could be accidentally logged in server logs. | No | Thunder follows secure logging practices per project guidelines: PII and secrets are masked using `MaskString` from `internal/system/log`. |
| 9 | Spoofing | **Resource Parameter Mismatch**: Attacker changes the `resource` parameter between authorization and token requests to target a different resource server. | No | The `resource` parameter is validated at both the authorization and token endpoints. At the token endpoint, the resource value must match what was stored in the authorization code. The resource value must be an absolute URI without a fragment. |

---

#### [I-04]: Token Introspection (Resource Server → Introspection Endpoint)

**Description**  
A resource server sends a POST request to `/oauth2/introspect` with a token to validate its active status and retrieve associated claims.

**Assets Involved**

| Initiator | Intermediate | Target |
|---|---|---|
| Resource Server | — | Thunder Introspection Endpoint |

**Data Flow**  
Resource server POSTs the token. Thunder validates the JWT signature, checks expiry, and returns `active: true/false` with claims.

**Access Control**  
Currently, the introspection endpoint does not enforce caller authentication.

**Security Considerations**

| Area | Response | Comments |
|---|---|---|
| Data Confidentiality | Medium confidential [C-Medium] | Accepts tokens; returns claim metadata |
| Communication Medium | Network interaction [M-NT] | HTTPS POST |
| Transport Security | TLS 1.3 Encryption | |
| Authentication | No Authentication | **Gap**: Endpoint not authenticated |
| Accessibility | Publicly Accessible | |

**Threat Assessment**

| ID | Category | Threat | Materializable | Mitigations / Comment |
|---|---|---|---|---|
| 1 | Information Disclosure | **Unauthenticated Token Introspection**: Any party with a token can introspect it and learn its claims, audience, and expiry. Per RFC 7662 §2.1, the introspection endpoint SHOULD be protected. | Yes | The introspection endpoint currently does not require caller authentication. **Recommendation**: Require client authentication (e.g., `client_secret_basic`) or restrict access to known resource servers. This should be tracked as a risk registry item. |
| 2 | Spoofing | **Token Fishing**: Attacker probes the introspection endpoint with guessed token values to discover valid tokens. | No | Access tokens are JWTs with high entropy (signed with RS256). Guessing a valid token is computationally infeasible. However, protecting the endpoint with authentication (Threat 1 mitigation) eliminates this risk entirely. |

---

#### [I-05]: UserInfo Request (OAuth Client → UserInfo Endpoint)

**Description**  
An OAuth client sends a GET or POST request to `/oauth2/userinfo` with a Bearer access token to retrieve the authenticated user's OIDC claims.

**Assets Involved**

| Initiator | Intermediate | Target |
|---|---|---|
| OAuth Client | — | Thunder UserInfo Endpoint |

**Data Flow**  
Client sends access token (Bearer). Thunder validates the token, extracts user attributes based on token claims and application configuration, and returns the user's OIDC profile.

**Access Control**  
Requires a valid Bearer access token. Client credentials grant tokens are rejected (as there is no resource owner in that flow).

**Security Considerations**

| Area | Response | Comments |
|---|---|---|
| Data Confidentiality | High confidential [C-High] | Returns user PII (name, email, phone, address) based on granted scopes |
| Communication Medium | Network interaction [M-NT] | HTTPS |
| Transport Security | TLS 1.3 Encryption | |
| Authentication | Bearer Token | Access token validated via JWT signature verification |
| Accessibility | Publicly Accessible | Requires valid access token |

**Threat Assessment**

| ID | Category | Threat | Materializable | Mitigations / Comment |
|---|---|---|---|---|
| 1 | Information Disclosure | **Access Token Theft for UserInfo Access** (RFC 9700 §4.10): Stolen access token used to retrieve user's PII from UserInfo. | No | Access tokens are short-lived JWTs with configurable expiry. TLS 1.3 protects tokens in transit. Per RFC 9700 §2.2.1, sender-constraining (mTLS or DPoP) is recommended to prevent misuse of stolen tokens. **Note**: Thunder currently issues bearer tokens without sender-constraining. See Security Considerations section. |
| 2 | Information Disclosure | **Excessive Data Exposure**: UserInfo returns more claims than the client needs. | No | Claims are filtered based on the granted OAuth scopes and per-application user attribute configuration. Only OIDC standard scope-to-claim mappings are applied. |
| 3 | Elevation of Privilege | **Client Credentials Token at UserInfo**: A client attempts to use a client credentials token to access UserInfo. | No | Thunder explicitly rejects client credentials grant tokens at the UserInfo endpoint, as there is no resource owner. |

---

#### [I-06]: JWKS Request (External → JWKS Endpoint)

**Description**  
Any party (resource server, client) fetches the server's public keys from `GET /oauth2/jwks` to verify JWT token signatures.

**Assets Involved**

| Initiator | Intermediate | Target |
|---|---|---|
| Resource Server / OAuth Client | — | Thunder JWKS Endpoint |

**Data Flow**  
GET request returns a JWKS JSON document containing public keys (RSA, ECDSA, EdDSA) with key IDs, X.509 certificate chains, and algorithm metadata.

**Access Control**  
No authentication required. Public keys are intended to be publicly accessible.

**Security Considerations**

| Area | Response | Comments |
|---|---|---|
| Data Confidentiality | Low confidential [C-Low] | Contains only public keys; private keys never exposed |
| Communication Medium | Network interaction [M-NT] | HTTPS |
| Transport Security | TLS 1.3 Encryption | |
| Authentication | No Authentication | Public endpoint by design |
| Accessibility | Publicly Accessible | |

**Threat Assessment**

| ID | Category | Threat | Materializable | Mitigations / Comment |
|---|---|---|---|---|
| 1 | Tampering | **JWKS Response Manipulation**: Network attacker intercepts and modifies JWKS response to inject attacker's public key, enabling token forgery. | No | TLS 1.3 prevents man-in-the-middle modification. JWKS includes `x5c` (certificate chain) and `x5t#S256` (SHA-256 thumbprint) for certificate pinning validation. |
| 2 | Information Disclosure | **Key Algorithm Enumeration**: Attacker uses JWKS to determine signing algorithms and target cryptographic weaknesses. | No | Thunder uses RS256 (RSA with SHA-256) which is widely considered secure. Keys include full certificate metadata for validation. Algorithm downgrade is not possible as the server enforces RS256. |

---

#### [I-07]: Discovery Request (External → Discovery Endpoints)

**Description**  
Clients or libraries fetch OAuth/OIDC server metadata from `/.well-known/openid-configuration` or `/.well-known/oauth-authorization-server`.

**Assets Involved**

| Initiator | Intermediate | Target |
|---|---|---|
| OAuth Client / Library | — | Thunder Discovery Endpoints |

**Data Flow**  
GET request returns a JSON document with endpoint URLs, supported grant types, response types, scopes, token endpoint auth methods, JWKS URI, etc.

**Access Control**  
No authentication required. Metadata is publicly accessible by design per RFC 8414.

**Security Considerations**

| Area | Response | Comments |
|---|---|---|
| Data Confidentiality | Low confidential [C-Low] | Contains only public server metadata |
| Communication Medium | Network interaction [M-NT] | HTTPS |
| Transport Security | TLS 1.3 Encryption | |
| Authentication | No Authentication | Public endpoint per RFC 8414 |
| Accessibility | Publicly Accessible | |

**Threat Assessment**

| ID | Category | Threat | Materializable | Mitigations / Comment |
|---|---|---|---|---|
| 1 | Tampering | **Metadata Manipulation for Endpoint Confusion**: Attacker modifies discovery response to point client to malicious endpoint. | No | TLS 1.3 prevents man-in-the-middle modification. Per RFC 9700 §2.6, publishing OAuth Authorization Server Metadata helps prevent misconfiguration. |
| 2 | Information Disclosure | **Server Fingerprinting**: Attacker uses discovery metadata to enumerate server capabilities and find weaknesses. | No | Discovery metadata is public by design. Thunder only advertises supported and secure features (e.g., no implicit grant, PKCE supported). |

---

#### [I-08]: Dynamic Client Registration (External → DCR Endpoint)

**Description**  
An OAuth client (or admin) registers a new client application by sending a POST request to `/oauth2/dcr/register` with application metadata (redirect URIs, grant types, token endpoint auth method, etc.).

**Assets Involved**

| Initiator | Intermediate | Target |
|---|---|---|
| OAuth Client / Admin | — | Thunder DCR Endpoint |

**Data Flow**  
POST with client metadata. Thunder validates and creates the application, returning `client_id`, `client_secret` (for confidential clients), and the configuration.

**Access Control**  
Currently, the DCR endpoint does not enforce caller authentication.

**Security Considerations**

| Area | Response | Comments |
|---|---|---|
| Data Confidentiality | High confidential [C-High] | Returns client_id and client_secret for newly registered applications |
| Communication Medium | Network interaction [M-NT] | HTTPS |
| Transport Security | TLS 1.3 Encryption | |
| Authentication | No Authentication | **Gap**: Endpoint not authenticated |
| Accessibility | Publicly Accessible | |

**Threat Assessment**

| ID | Category | Threat | Materializable | Mitigations / Comment |
|---|---|---|---|---|
| 1 | Spoofing | **Unauthorized Client Registration**: Any attacker can register arbitrary OAuth clients, potentially impersonating legitimate services or creating malicious clients for phishing. | Yes | DCR endpoint currently does not require authentication. **Recommendation**: Implement initial access token or admin authentication for client registration per RFC 7591 §3. This should be tracked as a risk registry item. |
| 2 | Denial of Service | **Registration Flooding**: Attacker registers a large number of clients to exhaust storage or processing resources. | Yes | No rate limiting on DCR endpoint. **Recommendation**: Implement rate limiting and optionally require authentication. |
| 3 | Elevation of Privilege | **Privilege Escalation via DCR**: Attacker registers a client with excessive grant types or scopes. | No | Thunder validates the requested grant types and token endpoint auth methods against the server's supported capabilities. Public clients (auth method `none`) automatically have PKCE enforced. |

---

#### [I-09]: Token Exchange (OAuth Client → Token Endpoint)

**Description**  
An OAuth client uses the Token Exchange grant (RFC 8693) to exchange an existing token for a new token with different attributes, audience, or scope. This is used for delegation and impersonation scenarios.

**Assets Involved**

| Initiator | Intermediate | Target |
|---|---|---|
| OAuth Client | Client Auth Middleware | Thunder Token Endpoint |

**Data Flow**  
POST with `grant_type=urn:ietf:params:oauth:grant-type:token-exchange`, `subject_token`, `subject_token_type`, optionally `actor_token`, `actor_token_type`, `audience`, `scope`, `requested_token_type`.

**Access Control**  
Client authentication via middleware (same as standard token requests).

**Security Considerations**

| Area | Response | Comments |
|---|---|---|
| Data Confidentiality | High confidential [C-High] | Contains subject/actor tokens; issues new tokens |
| Communication Medium | Network interaction [M-NT] | HTTPS |
| Transport Security | TLS 1.3 Encryption | |
| Authentication | Client Secret (Basic/Post) | Client must be authenticated |
| Accessibility | Publicly Accessible | |

**Threat Assessment**

| ID | Category | Threat | Materializable | Mitigations / Comment |
|---|---|---|---|---|
| 1 | Elevation of Privilege | **Token Exchange for Privilege Escalation**: Attacker exchanges a low-privilege token for a higher-privilege token by specifying a different audience or scope. | No | The exchanged token inherits the subject from the original token. Scope can only be equal to or narrower than the subject token's scope. The `act` claim is included to record the delegation chain. |
| 2 | Spoofing | **Stolen Subject Token Reuse**: Attacker uses a stolen access token as a `subject_token` for token exchange. | No | Client authentication is required. The client must be authorized for the token exchange grant type. Tokens are validated for signature and expiry. **Note**: Sender-constraining would further mitigate this risk. |

---

#### [I-10]: Database Operations (OAuth Module → Database)

**Description**  
The OAuth module reads and writes authorization codes, authorization request contexts, and application data to/from the database.

**Assets Involved**

| Initiator | Intermediate | Target |
|---|---|---|
| OAuth Module | DB Client | Runtime DB / Identity DB / User DB |

**Data Flow**  
SQL queries via `DBClient` for CRUD operations on authorization codes, auth request contexts.

**Access Control**  
Database access is controlled by the application's internal database client. Multi-deployment scenarios are supported via `DEPLOYMENT_ID` in all queries.

**Security Considerations**

| Area | Response | Comments |
|---|---|---|
| Data Confidentiality | High confidential [C-High] | Contains authorization codes, PKCE challenges, user attributes, client configurations |
| Communication Medium | Database interaction [M-DB] | Local SQLite file or PostgreSQL connection |
| Transport Security | Local file (SQLite) or TLS (PostgreSQL) | |
| Authentication | Internal | Application-level database access |
| Accessibility | Internal | Not accessible externally |

**Threat Assessment**

| ID | Category | Threat | Materializable | Mitigations / Comment |
|---|---|---|---|---|
| 1 | Tampering | **SQL Injection**: Attacker injects SQL via OAuth parameters (client_id, scope, etc.) to manipulate database queries. | No | Thunder uses parameterized queries with `DBQuery` from `internal/system/database/model`. All user inputs are bound as parameters, not concatenated into SQL strings. |
| 2 | Information Disclosure | **Authorization Code Theft from Database**: An attacker with database access can steal active authorization codes. | No | Authorization codes are single-use and short-lived. Database files are local to the server (SQLite) or secured via authentication (PostgreSQL). `DEPLOYMENT_ID` provides multi-tenant isolation. |

---

## Review Checklist

### Security Considerations

| Security Consideration | State | Comments |
|---|---|---|
| Are all inputs and outputs validated? | **Yes** | OAuth parameters (response_type, grant_type, redirect_uri, scope, code_challenge, code_verifier, resource) are validated at the handler and service layers. Error descriptions are validated against RFC-allowed character sets. |
| Are rate limits in place where necessary? | **Partial** | Rate limiting is not explicitly implemented on OAuth endpoints (token, introspect, DCR). Consider adding rate limiting especially for the token endpoint (failed auth attempts) and DCR endpoint. |
| Are proper authentication and authorizations in place before granting access to resources based on least privilege and business needs? | **Partial** | Token endpoint has client auth middleware. UserInfo requires Bearer token. However, introspection and DCR endpoints lack caller authentication. |
| Are permissions, roles, and entitlements defined (based on least privilege) and validated in both the front end and back end? | **Yes** | Allowed grant types, token endpoint auth methods, and scopes are configured per-application and enforced at the backend. Public clients are restricted from using `client_secret_basic/post`. PKCE is auto-enforced for public clients. |
| Have any default credentials been changed, and are the default superuser/root accounts not in use? | **N/A** | OAuth module does not have default credentials. Client secrets are generated at registration time. |
| Is the source code kept private? | **No** | Thunder is an open-source project. The source code is publicly available. Security relies on the soundness of the protocol implementation, not code secrecy. |
| Is the source code or IaC code review being conducted, and have the findings been addressed? | **Yes** | Code reviews via pull requests on GitHub. |
| Is Static/IaC scanning conducted on the source code, and are findings addressed? | **Yes** | Go static analysis tools (gosec, staticcheck) are used. `#nosec` annotations are used where false positives are identified (e.g., token type constants). |
| Is Software Composition Analysis being conducted or integrated into the source code repository, and are findings addressed? | **Yes** | Go modules with known CVE scanning (go.mod, go.sum). |
| Is Dynamic scanning conducted on the non-production setup, and are findings addressed? | **To Check** | Recommend integrating DAST scanning (e.g., OWASP ZAP) for OAuth endpoints. |
| Are audit logs generated for critical functionalities and made available to administrators to track critical events? | **Partial** | Server-side error logging is in place. Recommend adding structured audit logs for token issuance events, client authentication failures, and authorization code generation/consumption. Observability events are published for token issuance. |
| Do audit logs for critical configuration changes include a record of the differences between the old and new versions? | **N/A** | OAuth runtime configuration is not changed at runtime; it is set at deployment time via configuration files. |
| What aspects of resilience are considered, such as RPO/RTO, MTTD, high availability, backups, disaster recovery options, health check endpoints, and end-user messaging? | **Partial** | Health check endpoints are available. Database connections have configurable pool sizes and timeouts. Recommend documenting RPO/RTO for authorization code storage and client registration data. |
| Are data in transit and data at rest encrypted? | **Yes** | TLS 1.3 for all network communication. Database encryption depends on deployment configuration. JWTs are cryptographically signed (RS256). Encryption key exists for data at rest (`crypto.encryption.key` in deployment.yaml). |
| Are sensitive data, such as credentials and keys, stored in secret stores like key vaults? | **Partial** | Signing keys and encryption key are stored on the local filesystem. The encryption key path supports `file://` URI scheme. Recommend supporting external secret stores (Vault, KMS) for production deployments. |
| Have you ensured that personal, sensitive, or confidential data is not logged in the logs? | **Yes** | Per project guidelines, `MaskString` from `internal/system/log` is used to mask sensitive data. Client secrets, tokens, and PII are not logged in cleartext. |

### Vulnerability Management

| Question | Response |
|---|---|
| How are we planning to address product vulnerabilities, and what's the frequency of patching? | Go dependencies are managed via `go.mod`. Vulnerabilities are tracked via `go vuln` and dependency scanning. Patches are applied upon release cycles. |
| How are we planning to address deployment vulnerabilities, and what's the frequency of patching? | Deployment-level patches (OS, container runtime) are managed by the deployment team. Docker images should be kept up to date. |
| Are there any End of Life or End of Service components being used? | No. Thunder uses Go latest stable, SQLite (actively maintained), PostgreSQL (actively maintained). |

### Privacy Considerations

| Privacy Consideration | State | Comments |
|---|---|---|
| Is the purpose and legal basis for the processing of personal data clearly defined? | **Yes** | Personal data is processed to fulfill the OAuth/OIDC protocol requirements (user authentication, token issuance, UserInfo responses). Processing is based on user consent (authorization grant). |
| Is personal data being stored securely? | **Yes** | User data is stored in the User DB with encryption at rest support. Access tokens containing user claims are signed JWTs that expire. |
| Are privacy policies updated to reflect any new personal data processing or changes to purpose and legal basis? | **N/A** | Privacy policies are managed by the deploying organization. |
| Is access to personal data being granted based on the need to know? | **Yes** | User attributes in tokens and UserInfo are filtered based on granted OAuth scopes and per-application attribute configuration. Only requested and consented claims are returned. |
| Are data retention requirements considered? | **Yes** | Authorization codes have configurable, short validity periods. Authorization request contexts expire in 10 minutes. Tokens have configurable expiry. |
| Is there a process for disposing of personal data collected upon request in a timely manner while meeting retention requirements? | **Partial** | Authorization codes and request contexts are auto-cleaned (single-use + expiry). User account deletion processes should cascade to revoke associated tokens. |
| Have you added relevant records in the WSO2 Data Inventory? | **To Check** | Ensure OAuth-related data processing is recorded in the WSO2 Data Inventory. |

---

## Threat Model Consultation Sessions

Session 1:
- Date: TBD
- Participants: Thunder Team, WSO2 Security Team
- Session recording: [TBD]
- Notes:
- Action Items:

---

## Risk Registry Entries

| ID | Threat | Risk Level | Status | Tracking |
|---|---|---|---|---|
| RR-01 | Introspection endpoint lacks caller authentication (I-04, Threat 1) | Medium | Open | To be tracked via GitHub issue |
| RR-02 | DCR endpoint lacks caller authentication (I-08, Threat 1) | High | Open | To be tracked via GitHub issue |
| RR-03 | DCR endpoint lacks rate limiting (I-08, Threat 2) | Medium | Open | To be tracked via GitHub issue |
| RR-04 | Token endpoint lacks rate limiting (I-03, Threat 7) | Medium | Open | To be tracked via GitHub issue |
| RR-05 | Access tokens are Bearer tokens without sender-constraining (I-05, Threat 1) | Low | Open | Consider implementing DPoP (RFC 9449) or mTLS (RFC 8705) per RFC 9700 §2.2.1 |

---

## Document Lifecycle

- Threat model created: 2025-02-24
- Threat model to be moved to Security Review Documents
- Threat model to be reviewed by security team and leads: TBD
- Created GitHub issues for tracking threats that need to be addressed: TBD
- Risk registry entities to be updated (if applicable)

---

## Appendix

### RFC 9700 Compliance Summary

| RFC 9700 Requirement | Section | Thunder Status | Notes |
|---|---|---|---|
| Exact string matching for redirect URIs | §2.1, §4.1.3 | **Compliant** | Exact match against pre-registered URIs |
| No open redirectors | §2.1, §4.11 | **Compliant** | Invalid redirect URIs show error page |
| CSRF protection (state/PKCE) | §2.1, §4.7 | **Compliant** | State parameter preserved; PKCE provides additional CSRF protection |
| Support for PKCE | §2.1.1 | **Compliant** | Full RFC 7636 support; S256 and plain methods; enforced for public clients |
| PKCE downgrade prevention | §2.1.1, §4.8.2 | **Compliant** | code_verifier required when code_challenge was stored |
| Implicit grant deprecated | §2.1.2 | **Compliant** | Only `code` response type supported; implicit grant not implemented |
| Resource owner password grant deprecated | §2.4 | **Compliant** | Not implemented |
| Client authentication recommended | §2.5 | **Compliant** | Supports client_secret_basic, client_secret_post; asymmetric (private_key_jwt) not yet supported |
| OAuth Server Metadata published | §2.6 | **Compliant** | Both OIDC and OAuth2 discovery endpoints available |
| Authorization responses over HTTPS only | §2.6 | **Compliant** | TLS 1.3 minimum; insecure redirect_uri produces warning |
| Sender-constrained access tokens | §2.2.1 | **Partial** | Not implemented; consider DPoP or mTLS |
| Refresh token rotation or sender-constraining | §2.2.2 | **Compliant** | Refresh token rotation supported via `renew_on_grant` config |
| Access token audience restriction | §2.3 | **Compliant** | `aud` claim in access tokens; `resource` parameter supported |
| Access token scope restriction | §2.3 | **Compliant** | Scope validation and intersection on refresh |
| Mix-up attack prevention | §2.1, §4.4 | **Compliant** | Server metadata published; `iss` claim in tokens |
| Authorization code single-use | §4.5 | **Compliant** | Codes deactivated immediately upon retrieval |
| 307 redirect prevention | §4.12 | **Compliant** | 302 redirects used |
| TLS best practices | §2.6 | **Compliant** | TLS 1.3 minimum per deployment.yaml |
| Cache-Control headers on token response | RFC 6749 §5.1 | **Compliant** | `no-store` and `Pragma: no-cache` set |

### Sample OAuth Configuration

```yaml
server:
  hostname: "localhost"
  port: 8090

tls:
  min_version: "1.3"
  cert_file: "repository/resources/security/server.cert"
  key_file: "repository/resources/security/server.key"

crypto:
  encryption:
    key: "file://repository/resources/security/crypto.key"
  keys:
    - id: "default-key"
      cert_file: "repository/resources/security/signing.cert"
      key_file: "repository/resources/security/signing.key"

jwt:
  preferred_key_id: "default-key"

cors:
  allowed_origins:
    - "https://localhost:3000"
    - "https://localhost:5190"
```
