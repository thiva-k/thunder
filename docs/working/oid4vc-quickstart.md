# Thunder as OID4VC Issuer + Verifier — Quickstart

Bring up Thunder as **both** an OpenID4VCI **issuer** and an OpenID4VP **verifier**, issue a PID
credential to a real EUDI wallet, and verify it back — end to end, over `ngrok`, with a self‑signed
verifier cert. Tested end to end against the **Heidi / Funke** and **Lissi** wallets.

Everything below is relative to the repo root.

---

## 0. How it fits together

```
        ┌──────────── ngrok (one public HTTPS origin) ─────────────┐
 wallet │  https://<NGROK_HOST>                                    │
 (phone)│    /openid4vci/*  /openid4vp/*  /oauth2/*  /gate/*  ──▶ backend :8090
        └──────────────────────────────────────────────────────────┘
```

- One Go backend on `:8090` is the issuer, verifier, OAuth server, and login engine.
- The backend serves the **gate** (login UI) as static files at `/gate/`, so the whole demo is a
  **single public origin** — which is what the wallet needs.
- Config layering: `repository/resources/conf/default.json` holds **localhost defaults and stays
  untouched**; `repository/conf/deployment.yaml` **deep‑merges over it** with all the ngrok‑specific
  values. You only ever edit `deployment.yaml` (and one cert) per ngrok host.
- The ngrok request inspector at `http://localhost:4040` is your primary debugging tool.

Two paths get exercised:

1. **Issue (OID4VCI):** `GET /openid4vci/offer` → QR → wallet runs PAR → authorize → login → token →
   nonce → credential. The wallet stores a small **batch** of single‑use copies.
2. **Verify (OID4VP):** create a presentation definition → `POST /openid4vp/initiate` → QR → wallet
   posts an encrypted `vp_token` → `GET /openid4vp/status/{txn}` → `COMPLETED` + result token. Each
   verification **consumes one credential copy**.

---

## 1. Prerequisites

- Go + `pnpm` (to build the gate), `make`, `openssl`, `curl`, `jq`.
- `ngrok`, authenticated: `ngrok config add-authtoken <token>`.
- A Python venv with `requests` + `qrcode` for the helper scripts:
  ```bash
  python3 -m venv .venv && .venv/bin/pip install requests qrcode
  ```
- A phone with an EUDI wallet — **Heidi / Funke** or **Lissi** both work end to end with the
  self‑signed setup here (see §12 for their strictness differences). Only the **EUDI DE Sandbox**
  wallet needs a CA‑registered cert.

---

## 2. Start ngrok and capture the host

```bash
ngrok http https://localhost:8090 --host-header=localhost
```
From its output, export the host (no scheme) — used in every later step:
```bash
export NGROK_HOST="abc123.ngrok-free.app"     # <- from ngrok
export NGROK_URL="https://${NGROK_HOST}"
```

> Free ngrok hosts change on every restart. Each new host = regenerate the cert SAN (§3), update the
> host in `deployment.yaml` (§4) and in the gate's `config.js` (§5 — just the config, no rebuild),
> then restart. To avoid the churn, claim a **reserved ngrok domain** (free) and
> run `ngrok http https://localhost:8090 --domain=<your>.ngrok-free.app` — then the host never
> changes and you do §3–§5 only once.

---

## 3. Generate the verifier cert for this host

The verifier signs its request object with the `vp-verifier` key, and `client_id =
x509_san_dns:<host>` tells the wallet to check the cert's **DNS SAN matches the host**. So the cert
must carry the current ngrok host as a SAN — **and be a proper end-entity cert** (`CA:FALSE` +
`keyUsage=digitalSignature`). A plain `openssl req -x509` produces a `CA:TRUE` cert with no key
usage; lenient wallets (Heidi) accept it, but strict ones (Lissi) reject the request object's
signature → "unable to verify presentation request". Always include the two extensions below.

```bash
cd backend/cmd/server/repository/resources/security
openssl ecparam -name prime256v1 -genkey -noout -out vp-verifier.key
openssl req -new -x509 -key vp-verifier.key -out vp-verifier.cert -days 365 \
  -subj "/CN=Thunder VP Verifier" \
  -addext "subjectAltName=DNS:${NGROK_HOST}" \
  -addext "basicConstraints=critical,CA:FALSE" \
  -addext "keyUsage=critical,digitalSignature"
# confirm SAN == NGROK_HOST, CA:FALSE, Digital Signature:
openssl x509 -in vp-verifier.cert -noout -ext subjectAltName,basicConstraints,keyUsage
cd -
```

The issuer key `vci-signing.cert/.key` (EC, no SAN) does **not** need regenerating — issuance has no
host‑binding gate. `make run_backend` auto‑creates `server.*` and `signing.*` if missing; create
`vci-signing` once if absent:
```bash
cd backend/cmd/server/repository/resources/security
[ -f vci-signing.key ] || { openssl ecparam -name prime256v1 -genkey -noout -out vci-signing.key; \
  openssl req -new -x509 -key vci-signing.key -out vci-signing.cert -days 365 -subj "/CN=Thunder VCI Issuer"; }
cd -
```

---

## 4. `deployment.yaml` — the only config you edit per host

Edit `backend/cmd/server/repository/conf/deployment.yaml`. Add/replace these blocks (leave the rest
— tls, database, cors — as shipped). **Do not touch `default.json`.**

```yaml
server:
  hostname: "localhost"
  port: 8090
  # Public base URL — builds the OAuth/OIDC discovery endpoints (issuer, PAR, authorize, token)
  # the wallet reaches through the tunnel.
  public_url: "https://<NGROK_HOST>"

gate_client:
  hostname: "<NGROK_HOST>"     # gate login links use the public origin
  port: 443
  scheme: "https"
  login_path: "gate/signin"
  error_path: "gate/error"

crypto:
  keys:
    - id: "default-key"
      cert_file: "repository/resources/security/signing.cert"
      key_file: "repository/resources/security/signing.key"
    - id: "vci-signing"                                   # OID4VCI issuer signing key
      cert_file: "repository/resources/security/vci-signing.cert"
      key_file: "repository/resources/security/vci-signing.key"
    - id: "vp-verifier"                                   # OID4VP request-object signing key (from §3)
      cert_file: "repository/resources/security/vp-verifier.cert"
      key_file: "repository/resources/security/vp-verifier.key"

# Wallet-facing verifier endpoints (overrides the localhost defaults in default.json).
openid4vp:
  client_id: "x509_san_dns:<NGROK_HOST>"                  # must match the cert SAN from §3
  base_url: "https://<NGROK_HOST>"
  trusted_issuers:
    # Trust the German reference PID issuer (optional) ...
    - issuer: "https://demo.pid-issuer.bundesdruckerei.de/c"
      cert_file: "repository/resources/security/signing.cert"
    # ... and trust credentials this very instance issues (so the round-trip verifies).
    - issuer: "https://<NGROK_HOST>"
      cert_file: "repository/resources/security/vci-signing.cert"

# Wallet-facing issuer endpoints.
openid4vci:
  credential_issuer: "https://<NGROK_HOST>"
  base_url: "https://<NGROK_HOST>"
  authorization_servers:
    - "https://<NGROK_HOST>"

oauth:
  par:
    expires_in: 300        # interactive login is slow; give PAR a generous window
```

> When the ngrok host changes, the fastest update is a search‑replace of the old host in this one
> file plus regenerating the cert in §3.

> **`trusted_issuers` matters for the round-trip:** the verifier rejects a presented credential whose
> `iss` isn't trusted. Because this instance issues with `iss = https://<NGROK_HOST>`, that entry must
> be present or verification fails with a trust error. If you re-issue after an ngrok host change, the
> credential's `iss` is the *old* host — re-issue at the current host (§8).

---

## 5. Build and stage the gate (login UI)

The backend serves the gate from `backend/cmd/server/apps/gate/` (gitignored build output). Build it
with base `/gate`, copy it in, and point its runtime config at the public origin.

```bash
# build (base defaults to /gate)
( cd frontend/apps/gate && pnpm install && pnpm build )

# stage into the backend's static dir
rm -rf backend/cmd/server/apps/gate
mkdir -p backend/cmd/server/apps/gate
cp -R frontend/apps/gate/dist/. backend/cmd/server/apps/gate/

# point the gate's runtime config at the public origin
#   in backend/cmd/server/apps/gate/config.js set:
#     server: { hostname: '<NGROK_HOST>', port: 443, http_only: false }
```

> Alternative (dev loop): run the gate's vite server instead —
> `cd frontend/apps/gate && HOST=127.0.0.1 PORT=5190 pnpm dev` — but the static copy above keeps
> everything on one origin and is what this guide assumes.

---

## 6. Start the backend

Run with security disabled so the setup APIs (§7, §8) need no admin token:

```bash
SKIP_SECURITY=true make run_backend
```
Wait for readiness, then confirm the **discovery metadata advertises the ngrok host**:
```bash
curl -s "${NGROK_URL}/.well-known/openid-credential-issuer" | jq '.credential_issuer'   # == NGROK_URL
curl -s "${NGROK_URL}/.well-known/oauth-authorization-server" | jq '.pushed_authorization_request_endpoint'
curl -sk -o /dev/null -w 'gate %{http_code}\n' "${NGROK_URL}/gate/"                      # 200
```

---

## 7. Import the demo resources (user type, user, apps, login flow)

The issuer/verifier need: a **PID‑holder user type**, the subject **user** (`erika`), the **wallet
OAuth client** (issuance), the **demo app** + **EUDI login flow** (verification). These live in a
declarative file imported in one call.

`thunderid-config.yaml` (resource definitions, with `{{.VAR}}` placeholders) and `thunderid.env`
(the values) are at the repo root. Import them via `POST /import`:

```bash
CONTENT=$(jq -Rs . < thunderid-config.yaml)
VARS=$(python3 - thunderid.env <<'PY'
import sys, json
d={}
for line in open(sys.argv[1]):
    line=line.rstrip()
    if '=' in line and not line.startswith('#'):
        k,_,v=line.partition('=')
        try: d[k.strip()]=json.loads(v.strip())
        except Exception: d[k.strip()]=v.strip()
print(json.dumps(d))
PY
)
curl -sk -X POST "${NGROK_URL}/import" -H 'Content-Type: application/json' \
  -d "{\"content\": ${CONTENT}, \"variables\": ${VARS}, \"options\": {\"upsert\": true}}" | jq
```

`thunderid.env` must define at least:
```ini
ERIKA_PASSWORD=erika123
ISSUER_WALLET_CLIENT_ID=<client_id the wallet presents at issuance>   # e.g. Heidi's
ISSUER_WALLET_REDIRECT_URIS=["<wallet redirect, e.g. ch.ubique.funke://issuance>"]
EUDI_DEMO_CLIENT_ID=EUDI-DEMO-CLIENT
EUDI_DEMO_REDIRECT_URIS=["http://localhost:3000","https://oauth.pstmn.io/v1/callback"]
```

> The wallet's `ISSUER_WALLET_CLIENT_ID` / redirect must match what your wallet actually sends at the
> PAR/authorize step, or issuance fails client validation. Find them in the ngrok inspector's first
> `POST /oauth2/par` after a scan.

---

## 8. Create the presentation definition (verifier)

Presentation definitions are API‑managed (not seeded). The `handle` you pick is also the DCQL
credential id and the value you pass as `definition_id` when initiating.

```bash
curl -sk -X POST "${NGROK_URL}/openid4vp/presentation-definitions" \
  -H 'Content-Type: application/json' \
  -d '{
        "handle": "eudi-pid",
        "display_name": "EUDI Wallet PID",
        "vct": "urn:eudi:pid:de:1",
        "format": "dc+sd-jwt",
        "mandatory_claims": ["given_name", "family_name"],
        "optional_claims": ["birthdate"]
      }' | jq
curl -sk "${NGROK_URL}/openid4vp/presentation-definitions" | jq '.[].handle'   # confirm "eudi-pid"
```

You can also create/edit it in the **Console** UI (Verifiable Presentations) — start the console with
`cd frontend/apps/console && pnpm dev`, open `https://localhost:5191/console`.

---

## 9. Issue a credential to the wallet

Generate the offer QR and scan it with the wallet:
```bash
.venv/bin/python3 test_openid4vci.py --base-url "${NGROK_URL}"
```
The wallet runs PAR → authorize → **gate login as `erika` / `erika123`** → token → nonce → credential.
The PID then appears in the wallet. Confirm on the wire (look for `POST /openid4vci/credential → 200`):
```bash
curl -s "http://localhost:4040/api/requests/http?limit=10" | python3 -c "
import sys,json
for r in reversed(json.load(sys.stdin)['requests']):
    u=r['request']['uri']
    if '/openid4vci/' in u or '/oauth2/token' in u:
        print(r['request']['method'], u.split('?')[0][-38:], '->', r['response']['status_code'])"
```

> **Batch / one‑time‑use:** EUDI wallets treat each PID copy as single‑use (unlinkability). The issuer
> advertises `batch_credential_issuance` and mints one copy per holder proof, so the wallet stores a
> small batch (Heidi takes ~2). Each verification (§10) consumes one. When you run out you get
> "no matching credentials" in the wallet — just re‑issue (re‑run this step).

---

## 10. Verify the credential back

Initiate a verifier transaction, render the QR, scan and share in the wallet, and read the result:
```bash
.venv/bin/python3 test_openid4vp.py --base-url "${NGROK_URL}" --definition-id eudi-pid --rp-id demo-rp
```
On **Share/Confirm** the wallet posts an encrypted `vp_token`; the script prints `✓ COMPLETED` with the
decoded **result token** containing the verified `given_name` / `family_name` / `birthdate` and the
confirmed holder key binding.

To see verification drive a real **login** (and JIT‑provision a user from the credential), point a
sample app (e.g. the React SDK sample) at `client_id = EUDI-DEMO-CLIENT` and run its flow — the issued
claims land in the access token.

---

## 11. Troubleshooting

| Symptom | Cause | Fix |
|---|---|---|
| Setup API (`/import`, presentation-definitions) returns `401` | backend not actually started with `SKIP_SECURITY=true` (env not inherited by the server process) | start it as `SKIP_SECURITY=true make run_backend` in the same shell; confirm via the import call in §7 |
| Discovery/PAR points at `localhost` | `server.public_url` not set | §4 `server.public_url`, restart |
| `no signing key found for key id vp-verifier` | key missing from `crypto.keys` | add the three keys in §4 |
| Wallet "no matching credentials" | all single‑use copies consumed | re‑issue (§9) |
| Wallet shows `access_denied` **after** you tap Share/Confirm (claims were displayed) | the wallet's JOSE lib can't build the encrypted response with the one enc method advertised (e.g. Lissi can't do `ECDH-ES`+`A128GCM`) | advertise more `response_enc_values` so the wallet can pick one it supports — `default.json` ships `["A128GCM","A256GCM","A128CBC-HS256","A256CBC-HS512"]` for this reason; keep at least one of `A128GCM`/`A256GCM` |
| Wallet "unable to verify / process presentation request" (fails **before** showing claims) | request-object signature can't be validated: `vp-verifier` cert is a CA cert (`CA:TRUE`) or carries a stray `kid` | regenerate the cert as end-entity (§3 — `CA:FALSE`+`keyUsage`); the request object must sign with `x5c` only (no `kid`) |
| Verify fails with a trust error | credential `iss` not in `trusted_issuers` | add `https://<NGROK_HOST>` (§4); re‑issue at the current host |
| Gate screen errors / 502 | gate not staged or config points at a dead host | redo §5; set `config.js` server to `<NGROK_HOST>` |
| `credential offer expired` / inspector 404 | ngrok tunnel died | restart ngrok, redo §2–§6 |
| PAR client validation fails | `ISSUER_WALLET_CLIENT_ID`/redirect ≠ what wallet sends | read the wallet's `POST /oauth2/par` in the inspector, fix `thunderid.env`, re‑import |
| Phone can't reach `*.ngrok-free.app` | carrier/Wi‑Fi blocks ngrok | switch networks, or use a reserved domain |

---

## 12. Wallet compatibility notes

Both **Heidi / Funke** and **Lissi** complete the full round trip with the **self‑signed
`x509_san_dns`** setup in this guide — self‑signed is valid for that scheme (the host↔cert‑SAN binding
*is* the trust; no CA chain required). The two wallets just differ in strictness, and this guide's
defaults already satisfy both:

- **Heidi / Funke** (Rust) — lenient: accepts a `CA:TRUE` cert, a stray `kid`, and `A128GCM`
  response encryption.
- **Lissi** (iOS / EUDI Swift lib) — strict, and surfaced three gotchas that are now baked into the
  defaults: the cert **must be end‑entity** (`CA:FALSE`+`keyUsage`, §3); the request object **must not
  carry a `kid`** (x5c only — already the case in code); and it can't do `ECDH‑ES`+`A128GCM`, so the
  verifier must **advertise `A256GCM`** too (already in `default.json` `response_enc_values`). With
  those, Lissi shows the claims and shares successfully.
- **EUDI DE Sandbox wallet** — the one case that genuinely needs registered infrastructure: an Access
  Certificate + Registration Certificate from the German Sandbox Registrar. Switch `client_id` to
  `x509_hash:<sha256-of-access-cert-DER>`, point `signing_key_id` at the CA‑chained key, and set a
  registration cert. That's infrastructure, not code.

> Debugging a wallet that stops mid‑flow: the error *stage* tells you the cause. Fails **before**
> showing claims → request‑object signature/cert (§3). Shows claims, fails **after** Share → response
> encryption (advertise more `response_enc_values`). Posts `access_denied` but the enum also has a
> distinct `user_cancelled` — so `access_denied` means a processing failure, not your tap.
```
