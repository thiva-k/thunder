# OID4VC End-to-End Testing Runbook

Bring up Thunder from scratch as both a **VC issuer (OpenID4VCI)** and a **verifier (OpenID4VP)**, issue a PID credential to a real EUDI wallet over `ngrok`, then verify it back. All identity resources (subject user, wallet clients, credential configuration, presentation definition) are created from a single declarative bundle via the import API. Written so a new user — or an agent — can run the whole thing unattended.

> Branch: `openid4vp-management`. All paths are relative to the repo root `/Users/thiva/Desktop/Repos/thunder3`.

Both **Heidi/Funke** and **Lissi** complete issuance **and** verification end-to-end against self-signed material over an ngrok HTTPS origin — see §11.

---

## 0. Architecture

```
            ┌──────────────── ngrok (one public HTTPS origin) ───────────────┐
 wallet ───▶│  https://<ngrok-host>                                          │
 (phone)    │     /openid4vci/*  /openid4vp/*  /oauth2/*  /gate/*  → :8090    │
            └────────────────────────────────────────────────────────────────┘
```

- **Backend** (Go) on `:8090` — issuer + verifier + OAuth + login, and it serves the **gate** login UI at `/gate/*` from `backend/cmd/server/apps/gate`. One origin, so the wallet only ever talks to the ngrok host.
- The wallet MUST reach Thunder over a single public HTTPS origin → `ngrok`.
- **ngrok request inspector** at `http://localhost:4040` is the primary debug tool.

Two flows are exercised:
1. **Issue (OID4VCI):** `GET /openid4vci/offer` → QR → wallet runs PAR → authorize → gate login → token (DPoP) → nonce → credential.
2. **Verify (OID4VP):** `POST /openid4vp/initiate` → QR → wallet posts an encrypted `vp_token` → `GET /openid4vp/status/{txn}` → `COMPLETED` + result token.

---

## 1. Prerequisites

- `ngrok` installed and authenticated (`ngrok config add-authtoken …`).
- A wallet on a phone. **Heidi/Funke** and **Lissi** both work end-to-end with the self-signed setup below. The **EUDI DE Sandbox** wallet needs a CA-chained verifier cert (see §11).
- Go toolchain + `pnpm` (for the one-time `make run` that bootstraps platform defaults and builds the gate UI).
- Python with `qrcode` (`pip install qrcode`) and `curl`, `jq`, `openssl`.

---

## 2. Start ngrok and capture the host

```bash
ngrok http https://localhost:8090 --host-header=localhost
```
```bash
export NGROK_HOST="abc123.ngrok-free.app"     # <-- from ngrok output, no scheme
export NGROK_URL="https://${NGROK_HOST}"
```
> Free ngrok hosts change on every restart. A new host requires re-doing §3 (the verifier cert SAN) and §4 (the config URLs + client_id).

---

## 3. Generate issuer + verifier signing certs

Both are EC P-256 (ES256), placed in `backend/cmd/server/repository/resources/security/`.

```bash
cd /Users/thiva/Desktop/Repos/thunder3/backend/cmd/server/repository/resources/security

# Issuer key (signs the SD-JWT VC). No SAN needed — issuance has no host binding.
openssl ecparam -name prime256v1 -genkey -noout -out vci-signing.key
openssl req -new -x509 -key vci-signing.key -out vci-signing.cert -days 365 \
  -subj "/CN=Thunder VCI Issuer"

# Verifier key (signs the OID4VP request object / JAR). The wallet checks the
# cert's DNS SAN matches the client_id host, so it MUST carry the ngrok host.
openssl ecparam -name prime256v1 -genkey -noout -out vp-verifier.key
openssl req -new -x509 -key vp-verifier.key -out vp-verifier.cert -days 365 \
  -subj "/CN=Thunder VP Verifier" \
  -addext "subjectAltName=DNS:${NGROK_HOST}"

# confirm the SAN:
openssl x509 -in vp-verifier.cert -noout -ext subjectAltName   # DNS:<NGROK_HOST>
```

---

## 4. Configure for the ngrok origin

Config is `backend/cmd/server/config/default.json` (base) deep-merged with `backend/cmd/server/deployment.yaml` (override). Keep the ngrok host **only** in `deployment.yaml` so the tracked base stays clean.

`default.json` ships both engines **disabled** by default (`signing_key_id: ""`); `deployment.yaml` enables them by pointing at the EC keys generated in §3:
```jsonc
"openid4vci": { "signing_key_id": "",  "enforce_scope": false }   // disabled until overridden
"openid4vp":  { "signing_key_id": "",  "enforce_trusted_issuer": false,
                "enforce_key_binding": true, "use_presentation_definition": false /* DCQL */ }
```

Edit **`deployment.yaml`** — set the public origin, register the two new keys, and override the verifier identity for the ngrok host:
```yaml
server:
  public_url: "https://<NGROK_HOST>"          # issuer/verifier derive their public URLs from this

gate_client:
  hostname: "<NGROK_HOST>"                     # generated gate links use the public origin
  port: 443
  scheme: "https"
  login_path: "gate/signin"
  error_path: "gate/error"

crypto:
  keys:
    - id: "default-key"
      cert_file: "config/certs/signing.cert"
      key_file: "config/certs/signing.key"
    - id: "vci-signing"
      cert_file: "repository/resources/security/vci-signing.cert"
      key_file: "repository/resources/security/vci-signing.key"
    - id: "vp-verifier"
      cert_file: "repository/resources/security/vp-verifier.cert"
      key_file: "repository/resources/security/vp-verifier.key"

# Partial override — only these two fields change; the rest of openid4vp comes from default.json.
openid4vp:
  client_id: "x509_san_dns:<NGROK_HOST>"
  signing_key_id: "vp-verifier"
```
> The verifier's per-transaction response-decryption key is ephemeral (generated per request); `ephemeral_key_id` is just a `kid` label and needs no registered key.

---

## 5. Bootstrap platform defaults (one time)

`make run` runs the backend from `backend/cmd/server`, bootstraps the default OU, default flows (incl. `default-basic-flow`), admin user, and the Console app, and builds the frontend. It uses the same `backend/cmd/server/database/*.db` the test session below uses.

```bash
cd /Users/thiva/Desktop/Repos/thunder3
make run
# wait until it prints "Servers running", then Ctrl+C
```

Make the backend serve the gate UI on the ngrok origin, and point the gate at the public origin:
```bash
cd /Users/thiva/Desktop/Repos/thunder3
mkdir -p backend/cmd/server/apps/gate
cp -r frontend/apps/gate/dist/* backend/cmd/server/apps/gate/
# gate's runtime config must target the public origin so the phone browser reaches the API:
printf 'window.runtimeConfig={server:{public_url:"%s"}};\n' "$NGROK_URL" \
  > backend/cmd/server/apps/gate/config.js
```

---

## 6. Run the backend for the test session

Run from `backend/cmd/server` with security disabled so the admin/import APIs are open locally. Same DB + server home as the bootstrap above, so it sees the default OU/flows and serves `/gate`.

```bash
cd /Users/thiva/Desktop/Repos/thunder3/backend/cmd/server
SKIP_SECURITY=true go run .
```
Liveness + correct origin:
```bash
curl -sk -o /dev/null -w '%{http_code}\n' https://localhost:8090/gate                                 # 200
curl -s "${NGROK_URL}/.well-known/openid-credential-issuer" | jq -r '.credential_issuer'              # == NGROK_URL
```
> Shut down with SIGTERM/Ctrl+C (not `kill -9`) so SQLite WAL checkpoints and data persists.

---

## 7. Import the declarative bundle

One bundle (`thunderid-config.yaml` + `thunderid.env`) creates everything identity-side: the `pid-holder` user type, the subject **`erika`**, the **Heidi** and **Lissi** wallet OAuth clients, the **credential configuration** (`eudi-pid`), and the **presentation definition** (`eudi-pid`). `credential_configuration` and `presentation_definition` are importable just like any other declarative resource.

`thunderid.env` values are passed as the import `variables`; array-valued vars are parsed as JSON:
```bash
cd /Users/thiva/Desktop/Repos/thunder3
VARS=$(python3 - thunderid.env <<'PY'
import sys, json
pairs={}
for line in open(sys.argv[1]):
    line=line.rstrip()
    if '=' in line and not line.startswith('#'):
        k,_,v=line.partition('='); k=k.strip(); v=v.strip()
        try: pairs[k]=json.loads(v)
        except Exception: pairs[k]=v
print(json.dumps(pairs))
PY
)
CONTENT=$(jq -Rs . < thunderid-config.yaml)
curl -sk -X POST "${NGROK_URL}/import" -H 'Content-Type: application/json' \
  -d "{\"content\": ${CONTENT}, \"variables\": ${VARS}, \"options\": {\"upsert\": true}}" \
  | jq '.summary, (.results[] | {resourceType, status, message})'
# expect summary.failed == 0
```
Re-running the import is idempotent (resources upsert by their declared `id`).

> The credential configuration and presentation definition are **OU-bound**: each carries `ouHandle: default` in the bundle (resolved to the bootstrapped default OU). Creating either via the API/console requires an owning OU (`ouId`, or `ouHandle` on import); the handle stays globally unique per deployment, so the issuer metadata and offer/initiate endpoints resolve by handle unchanged.

Confirm the issuer advertises the credential and the verifier has the definition:
```bash
curl -s  "${NGROK_URL}/.well-known/openid-credential-issuer" | jq '.credential_configurations_supported | keys'   # ["eudi-pid"]
curl -sk "${NGROK_URL}/openid4vp/presentation-definitions" | jq '.[].handle'                                       # "eudi-pid"
```

> The bundle also includes an `eudi-user` type, `default-eudi-flow`, and an `eudi-demo-app` — those drive the separate "log in to an app with your EUDI wallet" flow and are not needed for the issue/verify round-trip below.

---

## 8. Issue a VC to the wallet

> **Console UI alternative:** the steps in §8 and §9 are also available from the Console without curl. On a **Verifiable Credential**, the QR (offer) action renders the issuance QR; on a **Verifiable Presentation**, the QR (Verify) action initiates a transaction, shows the scannable QR, and **polls the live status** — rendering the verified claims and holder key binding on completion. The curl flow below is the equivalent for headless/agent runs.

The issued claims come from the subject's user profile by name. The bundle's `erika` carries `given_name=Erika`, `family_name=Mustermann`, `birthdate=1986-03-22`, password `erika123`.

```bash
RESP=$(curl -sk "${NGROK_URL}/openid4vci/offer?credential_configuration_id=eudi-pid")
echo "$RESP" | jq
DEEPLINK=$(echo "$RESP" | jq -r '.credential_offer_uri')   # openid-credential-offer://...
python3 -c "import qrcode; qrcode.make('${DEEPLINK}').save('/tmp/qr_ISSUE.png')"; open /tmp/qr_ISSUE.png
```

Scan `/tmp/qr_ISSUE.png` with Heidi **or** Lissi. The offer is credential-specific, not wallet-specific — the same QR works for either wallet; each presents its own pre-registered `client_id`. The wallet runs PAR → `/oauth2/authorize` → gate → **log in as `erika` / `erika123`** → `/oauth2/token` (DPoP) → `/openid4vci/nonce` → `/openid4vci/credential`, and the PID appears in the wallet.

Confirm on the wire:
```bash
curl -s "http://localhost:4040/api/requests/http?limit=12" | python3 -c "
import sys,json
for r in sorted(json.load(sys.stdin)['requests'], key=lambda r:r.get('start',0)):
    u=r['request']['uri'].split('?')[0]
    if '/openid4vci/' in u or '/oauth2/' in u:
        print(r['response']['status_code'], r['request']['method'], u[-40:])"
# success ends with: 200 POST /openid4vci/credential
```

---

## 9. Verify the VC back to Thunder

```bash
RESP=$(curl -sk -X POST "${NGROK_URL}/openid4vp/initiate" -H 'Content-Type: application/json' \
  -d '{"definition_id":"eudi-pid","rp_id":"demo-rp"}')
echo "$RESP" | jq
TXN=$(echo "$RESP"  | jq -r '.txn_id')
WURL=$(echo "$RESP" | jq -r '.wallet_url')     # openid4vp://?client_id=...&request_uri=...
python3 -c "import qrcode; qrcode.make('${WURL}').save('/tmp/qr_VERIFY.png')"; open /tmp/qr_VERIFY.png
echo "TXN=$TXN"
```

Scan `/tmp/qr_VERIFY.png`. The wallet does `GET /openid4vp/request` (the signed JAR — its cert must carry the ngrok SAN from §3), shows the requested claims, and on Share posts an encrypted `vp_token` to `POST /openid4vp/response`. Each transaction is single-use; re-run `initiate` for a fresh QR (e.g. to try the other wallet).

Poll + decode:
```bash
curl -sk "${NGROK_URL}/openid4vp/status/${TXN}" | jq    # { "status": "COMPLETED", "result_token": "<JWT>" }

TOKEN=$(curl -sk "${NGROK_URL}/openid4vp/status/${TXN}" | jq -r '.result_token')
python3 -c "import sys,json,base64; p=sys.argv[1].split('.')[1]; print(json.dumps(json.loads(base64.urlsafe_b64decode(p+'==')),indent=2))" "$TOKEN"
# verified_claims: given_name / family_name / birthdate, plus cnf.jwk (holder key binding verified)
```

> The PD requests `given_name + family_name + birthdate` (optional and mandatory claims are concatenated into one DCQL `claims` list, no `claim_sets`, so all are requested). A wallet's consent screen may *summarize* the request by the credential's display attribute (`given_name`) — the disclosed set is governed by the DCQL, and all requested claims come back in `verified_claims`.

---

## 10. Debugging

- **ngrok inspector:** `http://localhost:4040` — replays every wallet request. The dumps in §8/§9 are the fastest triage.
- **Backend log:** start with `... go run . > /tmp/thunder.log 2>&1` and grep it. `LOG_LEVEL=debug` surfaces the real reason behind generic wallet-facing errors (e.g. DPoP).

| Symptom | Cause | Fix |
|---|---|---|
| `unable to determine resource type` on import | missing/typo'd `# resource_type:` comment | each doc needs `# resource_type: <type>` |
| import `failed` with "already exists" | re-import with a different `id` than the stored one | keep the declared `id` stable; upsert keys on it |
| gate 404 / blank after redirect | gate not served on the ngrok origin | redo §5 copy + `config.js` |
| `invalid_dpop_proof` | DPoP defaults / clock skew | check `oauth.dpop` in `default.json`; `LOG_LEVEL=debug` for the real reason |
| `no presentation definition registered` | bundle not imported | run §7 |
| wallet "issuer/verifier not trusted" | wallet trust policy (see §11) | use Heidi/Lissi over ngrok, or a CA-chained cert |
| stale QR scanned | macOS Preview cached old image | unique filenames + `killall Preview` |

---

## 11. Wallet trust notes

- **Heidi/Funke** and **Lissi** — both complete **issuance and verification** end-to-end with the self-signed `vci-signing` (issuer) and `vp-verifier` (verifier, DNS-SAN-matched) certs over the ngrok HTTPS origin. No trust-list registration required for either leg.
- **EUDI DE Sandbox wallet** — requires an **Access Certificate** from the German Sandbox Registrar plus a **Registration Certificate**; switch `client_id` to `x509_hash:<sha256-of-access-cert-DER>`, point `signing_key_id` at the CA-chained key, set `registration_cert_file`, and add `A256GCM` to `response_enc_values`. This is infrastructure, not code.

For local/demo round-trips, **use Heidi or Lissi**. See [oid4vc-progress.md](oid4vc-progress.md) for the full trust analysis.
```
