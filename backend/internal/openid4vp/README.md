# Thunder OpenID4VP Verifier

Thunder ships an embedded **OpenID for Verifiable Presentations 1.0** verifier
that consumes Verifiable Credentials (SD-JWT VC) presented by EUDI-style
wallets. It targets the HAIP profile out of the box: signed JAR with `x5c`,
`request_uri` flow, encrypted `direct_post.jwt` response, DCQL, and the
`x509_hash:` client identifier.

Two ways to integrate:

| Path | Who triggers verification | Result delivery | Use case |
|---|---|---|---|
| **Flow Engine** | Thunder's authentication flow | `AuthenticatedUser` → session/assertion | The app uses Thunder as its IdP. "Sign in with EUDI Wallet". |
| **REST API** | Partner backend (any RP) | Signed result token | Standalone verifier-as-a-service. A non-Thunder app uses Thunder purely to verify a credential. |

Both paths share the same wallet round-trip and verification pipeline.

---

## Quick Start

Just start the server. The OpenID4VP verifier is always wired in; with nothing
in `deployment.yaml`, it boots with the defaults from `default.json`:

- `client_id: "x509_hash:dev-placeholder"` (will need replacing for live wallets — see [Going live](#going-live))
- Signing key: `default-key` (the bundled dev key)
- Base URL: `https://localhost:8090`
- One presentation definition registered: `eudi-pid` (German PID SD-JWT VC)
- All HAIP defaults: `response_mode=direct_post.jwt`, `A128GCM` content encryption, ECDH-ES key agreement, `request_audience=https://self-issued.me/v2`, and a 5-minute lifetime for both the request object and the per-transaction state.

---

## Path A — Flow Engine (Sign in with EUDI Wallet)

### Wire Up

1. Make sure your application's authentication flow handle is `default-eudi-flow`
   (set in the Applications console or via the applications REST API).
2. Open the application in a browser. The Gate renders:
   - "Sign in with EUDI Wallet" button on the auth-chooser screen.
   - After click: QR code + "Open wallet on this device" link + a **"Refresh status"** button.
3. After the wallet completes, the user clicks **Refresh status** → flow advances → assertion issued.

### Flow Nodes (Already Shipped)

`backend/cmd/server/bootstrap/flows/authentication/auth_flow_eudi.json` defines:

```
start  →  choose_auth (PROMPT: "Sign in with EUDI Wallet")
       →  eudi_verify (TASK_EXECUTION: OpenID4VPVerifyExecutor, presentation_definition_id=eudi-pid)
             onSuccess     → auth_assert → end
             onIncomplete  → eudi_wait (PROMPT: QR + "Refresh status")
                                  action_poll → eudi_verify
```

### What the Executor Returns

- First entry: `service.Initiate("eudi-pid")` → adds `openid4vpClientId`,
  `openid4vpRequestUri`, `openid4vpWalletUri` to `additionalData`, status
  `ExecUserInputRequired`. Engine merges that into the `eudi_wait` VIEW.
- Re-entry (after user clicks **Refresh status**): poll branch → `service.Result(state)`.
  Returns `ExecComplete` when verification is `COMPLETED`, with disclosed claims
  on `AuthenticatedUser.Attributes`.

### Authenticated User Attributes

The verified PID attributes appear on the authenticated user's
`Attributes` map alongside two metadata keys for downstream consumers:

```json
{
  "given_name":  "Alice",
  "family_name": "Schmidt",
  "birthdate":   "1998-04-12",
  "openid4vp_issuer": "https://demo.pid-issuer.bundesdruckerei.de/c",
  "openid4vp_vct":    "urn:eudi:pid:de:1"
}
```

---

## Path B — REST API (Verifier as a Service)

For partner backends that do not use Thunder's flow engine. Two endpoints,
both public (no authentication at the protocol layer — put the verifier
behind your usual network policy if you need RP auth).

### 1. Initiate a Verification

```bash
curl -sk -X POST https://localhost:8090/openid4vp/initiate \
  -H "Content-Type: application/json" \
  -d '{"definition_id": "eudi-pid", "rp_id": "your-rp"}'
```

Response:

```json
{
  "txn_id":     "1kX05eHw…",
  "wallet_url": "openid4vp://?client_id=…&request_uri=…",
  "status_url": "https://localhost:8090/openid4vp/status/1kX05eHw…",
  "expires_at": "2026-06-02T13:12:34Z"
}
```

Render `wallet_url` to the user as a QR (cross-device) or as a tap link
(same-device). The user's wallet does the rest of the protocol round-trip
silently against Thunder.

### 2. Poll Status

```bash
curl -sk https://localhost:8090/openid4vp/status/1kX05eHw…
```

Poll at whatever cadence you choose (500ms–2s is typical for a checkout
flow). Possible responses:

| Status | Body |
|---|---|
| `PENDING` | `{ "status": "PENDING" }` |
| `COMPLETED` | `{ "status": "COMPLETED", "result_token": "<signed JWT>" }` |
| `FAILED` | `{ "status": "FAILED", "error": "<reason>" }` |
| `EXPIRED` | `{ "status": "EXPIRED" }` |
| (unknown txn) | HTTP 404 |

### 3. Validate the Result Token (on the RP Side)

Standard JWS verification against Thunder's JWKS:

```bash
curl -sk https://localhost:8090/oauth2/jwks
```

The result token's payload:

```json
{
  "iss":  "https://localhost:8090",
  "sub":  "<derived subject>",
  "aud":  "your-rp",                     // matches the rp_id you passed
  "txn":  "1kX05eHw…",                  // matches the txn_id from initiate
  "definition_id": "eudi-pid",
  "subject":       "<derived subject>",
  "verified_claims": {
    "given_name":  "Alice",
    "family_name": "Schmidt",
    "birthdate":   "1998-04-12"
  },
  "verifier": "x509_hash:…",
  "iat": …, "exp": …
}
```

Once verified, trust `verified_claims` and proceed with your application
logic.

---

## Wallet Round-Trip (Shared by Both Paths)

After either Path A or Path B has run its `initiate`, Thunder holds a
PENDING state row. The user receives a deep link
`openid4vp://?client_id=…&request_uri=…` as a QR or tap link. From
here the protocol is identical in both paths. The wallet drives it
against Thunder's public endpoints `/openid4vp/request` and
`/openid4vp/response`.

### Step W1 — Wallet Fetches the Signed Request Object (JAR)

```http
GET /openid4vp/request?state=1kX05eHw…  HTTP/1.1
Host: localhost:8090
Accept: application/oauth-authz-req+jwt
```

Thunder builds the JAR on demand from the state row's pinned definition
and signs it with the `signing_key_id`. Response:

```http
HTTP/1.1 200 OK
Content-Type: application/oauth-authz-req+jwt
Cache-Control: no-store

eyJhbGciOiJFUzI1NiIsInR5cCI6Im9hdXRoLWF1dGh6LXJlcStqd3QiLCJ4NWMiOlsiTUlJ…
```

Decoded JWS **header**:

```json
{
  "alg": "ES256",
  "typ": "oauth-authz-req+jwt",
  "kid": "<key thumbprint>",
  "x5c": ["MIID…<DER of signing cert>…"]
}
```

Decoded JWS **payload**:

```json
{
  "response_type":   "vp_token",
  "response_mode":   "direct_post.jwt",
  "client_id":       "x509_hash:dev-placeholder",
  "response_uri":    "https://localhost:8090/openid4vp/response?state=1kX05eHw…",
  "nonce":           "<32-byte URL-safe random>",
  "state":           "1kX05eHw…",
  "aud":             "https://self-issued.me/v2",
  "iat":             1748880600,
  "exp":             1748880900,
  "dcql_query": {
    "credentials": [{
      "id":     "pid-sd-jwt",
      "format": "dc+sd-jwt",
      "meta":   { "vct_values": ["urn:eudi:pid:de:1"] },
      "claims": [
        { "path": ["given_name"] },
        { "path": ["family_name"] },
        { "path": ["birthdate"] }
      ]
    }],
    "credential_sets": [{ "options": [["pid-sd-jwt"]] }]
  },
  "client_metadata": {
    "jwks": {
      "keys": [{
        "kty": "EC", "crv": "P-256", "kid": "vp-enc",
        "use": "enc", "alg": "ECDH-ES",
        "x": "…", "y": "…"
      }]
    },
    "vp_formats_supported": { "dc+sd-jwt": {} },
    "encrypted_response_enc_values_supported": ["A128GCM"]
  }
}
```

### Step W2 — Wallet Validates the Request, Gathers Consent, Builds Presentation

Wallet-side checks (a real EUDI wallet does the following):

1. JWS signature verified against the leaf cert in `x5c`.
2. `SHA-256(leaf cert DER)` base64url-encoded equals the suffix of the
   `client_id` advertised in the deep link.
3. `x5c` chain anchors to a root in the wallet's trust list.
4. `exp > now` and `iat <= now`.

The wallet then:

5. Evaluates the DCQL → picks the matching credential (the user's PID).
6. Displays a consent screen: "Acme wants `given_name`, `family_name`, `birthdate` from your EUDI Wallet PID."
7. Holder approves → the wallet:
   - Picks the disclosures for the three requested claims.
   - Builds a **Key Binding JWT** (signed with the holder's private key, the public half is `cnf.jwk` in the credential) carrying:

     ```json
     {
       "iat":     1748880620,
       "aud":     "x509_hash:dev-placeholder",      // == JAR client_id
       "nonce":   "<echo of JAR nonce>",
       "sd_hash": "<base64url(SHA-256(<sdjwt>~<disc>~<disc>~<disc>))>"
     }
     ```

   - Assembles the SD-JWT presentation:

     ```
     <issuer-jwt>~<disclosure_given_name>~<disclosure_family_name>~<disclosure_birthdate>~<kb-jwt>
     ```

   - Builds the OpenID4VP response payload (DCQL-keyed):

     ```json
     {
       "state":    "1kX05eHw…",
       "vp_token": { "pid-sd-jwt": "<sdjwt~disc~disc~disc~kbjwt>" }
     }
     ```

   - **Encrypts it as JWE** using the EC P-256 public key from
     `client_metadata.jwks` (ECDH-ES key agreement, A128GCM content encryption).
     Result is a compact JWE in the form `header.encryptedKey.iv.ciphertext.tag`
     (`encryptedKey` is empty under ECDH-ES "direct").

### Step W3 — Wallet Posts the Encrypted Response

```http
POST /openid4vp/response?state=1kX05eHw…  HTTP/1.1
Host: localhost:8090
Content-Type: application/x-www-form-urlencoded

response=eyJhbGciOiJFQ0RILUVTIiwiZW5jIjoiQTEyOEdDTSIsImtpZCI6InZwLWVuYyJ9.…
```

Thunder's verification pipeline runs:

1. Load state row (looks up by `state`); if missing/expired → 4xx.
2. JWE-decrypt with the state row's ephemeral private key.
3. Parse plaintext into `{state, vp_token}`; cross-check inner `state`.
4. Extract the presentation by DCQL credential id (`pid-sd-jwt`).
5. SD-JWT VC verification:
   - Issuer JWS signature against the pinned `trusted_issuers[*].cert_file`.
   - Recompute `_sd` digests for each disclosure.
   - KB-JWT signature against `cnf.jwk`; check `aud`, `nonce`, `sd_hash`, `iat` skew.
6. Policy:
   - `vct` matches the definition's `vct`.
   - All disclosed claim paths are in `requested_claims`.
   - All `mandatory_claims` are present.
7. Derive subject pseudonym from `subject_claims` (SHA-256 over `iss|claim=value|…`).
8. Mark state row `COMPLETED` and persist `Result.Claims`, `Result.Subject`.

Response to the wallet:

```http
HTTP/1.1 200 OK
Content-Type: application/json

{ "redirect_uri": "<configured result_redirect_uri>" }
```

If `result_redirect_uri` is unset (default), body is `{}`. The wallet
typically dismisses or follows the redirect.

### What Happens After W3 Depends on the Path

- **Path A**: the user clicks "Refresh status" in the Gate → executor's poll
  branch reads `service.Result(state)` → sees COMPLETED → flow advances to
  `auth_assert`.
- **Path B**: the next status poll by the RP at `GET /openid4vp/status/{txn_id}` sees COMPLETED →
  Thunder mints and returns the signed `result_token`.

The wallet itself doesn't poll anything — its two HTTP exchanges (W1 + W3)
are all it ever does.

---

## Configuration Reference

### Defaults (in `default.json`)

```json
"openid4vp": {
  "client_id": "x509_hash:dev-placeholder",
  "signing_key_id": "default-key",
  "base_url": "https://localhost:8090",
  "ephemeral_key_id": "vp-enc",
  "response_enc_values": ["A128GCM"],
  "request_audience": "https://self-issued.me/v2",
  "request_validity_seconds": 300,
  "state_ttl_seconds": 300,
  "leeway_seconds": 30,
  "result_token_validity_seconds": 300,
  "presentation_definitions": [
    {
      "id": "eudi-pid",
      "display_name": "EUDI Wallet PID",
      "credential_id": "pid-sd-jwt",
      "vct": "urn:eudi:pid:de:1",
      "requested_claims": ["given_name", "family_name", "birthdate"],
      "mandatory_claims": ["given_name", "family_name"],
      "subject_claims":   ["family_name", "given_name", "birthdate"],
      "trusted_issuers": [
        { "issuer":   "https://demo.pid-issuer.bundesdruckerei.de/c",
          "cert_file": "repository/resources/security/signing.cert" }
      ]
    }
  ]
}
```

### Engine-Level Fields

| Field | Meaning |
|---|---|
| `client_id` | Verifier identifier shown to the wallet. HAIP form: `x509_hash:<base64url(sha256(DER(your registered cert)))>`. |
| `signing_key_id` | Key id in `crypto.keys[]` used to sign request objects. Must be cert-backed. |
| `base_url` | Public URL of this Thunder instance. Drives `request_uri` and `response_uri`. |
| `ephemeral_key_id` | The `kid` advertised on the per-request ephemeral encryption JWK. Cosmetic. |
| `response_enc_values` | Allowed content-encryption algorithms. HAIP: `["A128GCM"]`. |
| `request_audience` | JAR `aud` claim. OpenID4VP: `https://self-issued.me/v2`. |
| `request_validity_seconds` | JAR `exp` window. |
| `state_ttl_seconds` | How long a verification transaction lives before EXPIRED. |
| `leeway_seconds` | Allowed clock skew on KB-JWT `iat`. |
| `result_token_validity_seconds` | Result-token `exp` window for Path B. |
| `registration_cert_file` | Optional. Path (relative to ServerHome) to a Registration Certificate JWT issued by the trust framework's registrar (e.g. EUDI Sandbox). When set, the JWT rides on every JAR's verifier-info attestation array (format `"registration_cert"`). Required by HAIP / EUDI. |
| `presentation_definitions[]` | One entry per credential type the verifier accepts. |

### Definition-Level Fields

| Field | Meaning |
|---|---|
| `id` | Registry key (e.g. `"eudi-pid"`). Used in `POST /openid4vp/initiate` and on flow executor nodes. |
| `display_name` | Optional human label. |
| `credential_id` | DCQL `credentials[].id` — the slot in `vp_token` the wallet must fill. |
| `vct` | SD-JWT VC type the wallet must present. |
| `requested_claims` | Claim paths to request in DCQL. Wallet may only disclose these (defense-in-depth). |
| `mandatory_claims` | Subset of `requested_claims` that MUST be present for verification to succeed. |
| `subject_claims` | Claim values folded into the derived subject pseudonym. Sorted, joined with `\|`, SHA-256. |
| `trusted_issuers[]` | List of pinned issuer certs. Each entry has an `issuer` URL and a PEM `cert_file` path relative to the server home. |

### Overriding Defaults

Anything in `default.json` can be overridden in `deployment.yaml`. For example,
to add a second definition next to `eudi-pid`:

```yaml
openid4vp:
  presentation_definitions:
    - id: "eudi-pid"
      display_name: "EUDI Wallet PID"
      credential_id: "pid-sd-jwt"
      vct: "urn:eudi:pid:de:1"
      requested_claims: ["given_name", "family_name", "birthdate"]
      mandatory_claims: ["given_name", "family_name"]
      subject_claims:   ["family_name", "given_name", "birthdate"]
      trusted_issuers:
        - issuer: "https://demo.pid-issuer.bundesdruckerei.de/c"
          cert_file: "repository/resources/security/signing.cert"
    - id: "my-org-membership"
      display_name: "My Org Membership Card"
      credential_id: "membership"
      vct: "https://creds.myorg.example/membership"
      requested_claims: ["member_id", "tier"]
      mandatory_claims: ["member_id"]
      subject_claims:   ["member_id"]
      trusted_issuers:
        - issuer: "https://creds.myorg.example"
          cert_file: "repository/resources/security/myorg-issuer.cert"
```

Adding a definition is purely a YAML edit. No Go code change is needed.

---

## Going Live

Three operational gates close before a real EUDI wallet can complete a
verification against your deployment. None of them are code:

1. **Real Access Certificate** from the trust framework's Registrar
   (e.g. the EUDI Sandbox Registrar). Install it as a `crypto.keys[]` entry
   that matches `openid4vp.signing_key_id`. Compute the hash:
   ```bash
   openssl x509 -in access-cert.pem -outform der \
     | openssl dgst -binary -sha256 \
     | basenc --base64url | tr -d '='
   ```
   Set `client_id: "x509_hash:<that hash>"` in `deployment.yaml`.
2. **Real PID issuer cert** (e.g. Bundesdruckerei) saved as PEM under
   `repository/resources/security/`. Update the relevant
   `trusted_issuers[*].cert_file` in `deployment.yaml`.
3. **Registration Certificate** from the Registrar. Save the JWT to
   `repository/resources/security/` and set
   `openid4vp.registration_cert_file` to its path (relative to ServerHome)
   in `deployment.yaml`. The engine loads it at startup and attaches it
   to every request object as a verifier-info attestation
   (`format: "registration_cert"`).

---

## Spec References

- OpenID for Verifiable Presentations 1.0: https://openid.net/specs/openid-4-verifiable-presentations-1_0.html
- HAIP (High-Assurance Interoperability Profile): https://openid.net/specs/openid4vc-high-assurance-interoperability-profile-1_0.html
- EUDI Wallet Developer Guide: https://bmi.usercontent.opencode.de/eudi-wallet/developer-guide/
- IETF SD-JWT VC: https://datatracker.ietf.org/doc/draft-ietf-oauth-sd-jwt-vc/
