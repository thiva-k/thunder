# OID4VC (Issuer + Verifier) — Progress

_Branch: `openid4vp-management` · Last updated: 2026-06-10_

Goal: demonstrate **Thunder as both a Verifiable Credential issuer (OpenID4VCI) and verifier (OpenID4VP)**, end-to-end against a real EUDI wallet — issue a credential to the wallet, then verify it back to Thunder.

---

## 1. What we planned

### Issuer (OpenID4VCI) — minimal, config-driven
- **Flow:** `authorization_code` (reuse Thunder's existing `/oauth2/authorize` + `/oauth2/token` + login engine + PKCE/PAR/DPoP). Pre-authorized-code flow deferred.
- **One fixed credential type:** vct `urn:eudi:pid:de:1`, format `dc+sd-jwt`, claims `given_name` / `family_name` / `birthdate` sourced from the authenticated Thunder user's profile.
- **Holder key binding (`cnf`) required**; no revocation; no batch.
- **Endpoints:** `GET /.well-known/openid-credential-issuer`, `POST /openid4vci/nonce`, `POST /openid4vci/credential` (no VCI token endpoint — wallet uses Thunder's `/oauth2/token`).
- **Foundation:** new SD-JWT VC issuance primitive `sdjwt.Issue()` reusing the existing `sdjwt` package, feeding straight back into `sdjwt.Parse`/`Verify` for a self-test round-trip before any wallet.
- **Exit criteria:** in-process authorize → token → nonce → credential, issued SD-JWT accepted by Thunder's own verifier path.

### Verifier (OpenID4VP) — Phase 1 management layer
- **Verifier identity** (client_id, signing key, alg, base_url, enforce flags) stays in **deployment config**.
- **Presentation definitions** become **managed resources**: CRUD API + Console UI + configdb table (moved out of static config).
- **Global trusted issuers** at deployment-config level (empty list = trust all); no per-definition binding.
- **Request state moved from in-memory/cache → runtimedb** (encrypted ephemeral key), so verification survives restarts and scales.
- **Claim mapping dropped** for Phase 1.
- **Console UI:** dedicated "Verifiable Presentations" sidebar tab with presentation-definition CRUD, plus a flow-step selector.

### Demo target
- Issue PID to a wallet, then verify it back to Thunder over `ngrok` (single-origin reverse proxy: `/gate`→5190, `*`→8090).
- Wallets tried: **Lissi** (HAIP/EUDI), then **Heidi/Funke**.

---

## 2. What we have achieved ✅

### Issuer (OpenID4VCI) — built & working
- New package `backend/internal/openid4vci/`: `signer.go`, `store.go`, `proof.go`, `claims.go`, `metadata.go`, `credential.go`, `offer.go`, `service.go`, `handler.go`, `init.go`, `model.go`, `error_constants.go`.
- SD-JWT issuance primitive `backend/internal/system/jose/sdjwt/issue.go` (+ `issue_test.go`), incl. the critical **DER→P1363 ECDSA** signature conversion so issued credentials verify.
- Config: `OpenID4VCIConfig` + `CredentialConfig` in `system/config/config.go`; seeded in `default.json` (`eudi-pid`, scope `urn:eudi:pid:de:1`).
- Wired in `servicemanager.go`; public paths opened in `security/permissions.go` (`/openid4vci/**`).
- **Live result:** full issuance works end-to-end against **both Lissi and Heidi** over ngrok — PAR → authorize → Thunder login (gate) → token (DPoP) → nonce → credential → wallet stores the PID.

### Verifier (OpenID4VP) — Phase 1 built, tested, live-verified
- New files: `db_store.go` (runtimedb state, encrypted ephemeral key), `definition_model.go`, `definition_store.go` (configdb), `definition_service.go`, `definition_handler.go`, `definition_errors.go` (+ `definition_service_test.go`).
- Modified: `init.go` (shared trust store, seed + load definitions), `service.go`, `model.go`, `request.go` (client metadata, optional aud, DCQL/PEX switch), `dcql.go`, `response.go` (DCQL + PEX parsing), `store.go` (cache store removed).
- DB: `OPENID4VP_REQUEST_STATE` (runtimedb), `OPENID4VP_PRESENTATION_DEFINITION` (configdb) for both sqlite + postgres, with cleanup wiring.
- Console UI: `frontend/apps/console/src/features/verifiable-presentations/` (models, api, components, pages, constants) + App.tsx routes + sidebar entry + flow-step `PresentationDefinitionSelect`. Typecheck passes.
- **Live-verified:** management API CRUD (Test 1); verifier initiate/status + runtimedb encrypted state + signed JAR (Test 2).

### 🎯 End-to-end round-trip — ACHIEVED with Heidi
Issue PID to Heidi → verify back to Thunder → `status: COMPLETED` with a signed result token. Verified claims: `given_name=Erika`, `family_name=Mustermann`, `birthdate=1986-03-22`, holder key binding (`cnf.jwk`) **verified**, trusted-issuer + nonce + KB-JWT all checked.

---

## 3. Issues we are facing ⚠️

### Primary blocker: Lissi rejects the verifier ("Contact is not verified")
- **Symptom:** Lissi receives the request, shows the data, but on **Confirm** returns `access_denied` with "WSO2 — Contact is not verified."
- **Root cause (confirmed via ngrok inspector + Lissi docs):** Lissi enforces **X.509 certificate trust anchored in eIDAS/EU trusted lists** (EUDI WRPAC / RP Access Certificate model). Our verifier signs the JAR with a **self-signed `x509_san_dns` cert** → not an EV/qualified cert chaining to a regulated CA → Lissi labels it "not verified" and **hard-blocks**. Wallets fetch `/.well-known/openid-federation` and `/.well-known/trust-statement` on our server (both **404**), so there is no fallback trust path either.
- **Why Heidi works but Lissi doesn't:** absent trust metadata → Heidi treats the verifier as *unknown* and warns-but-proceeds; Lissi *fails closed*. Same cert/JAR/DCQL — only the wallet's handling of missing trust differs. (Issuance has no such gate, which is why issuance worked on both.)
- **Not a code bug:** no public issue/forum reproduces it; it's Lissi's policy working as designed. There is **no Lissi developer/test toggle** to accept untrusted verifiers, and **no self-serve path** in code.

### Resolution paths for Lissi (all require an external trust anchor — infrastructure, not code)
1. **SPRIND German Sandbox** RP registration → obtain an RP Access Certificate from a Lissi-recognized anchor (most practical for a demo).
2. **Direct Lissi onboarding** — `wallet@lissi.id` / Starter Program / Connector.
3. **Real EV/QWAC certificate** for the domain (SAN matching).
4. ~~Self-signed OpenID Federation entity config~~ — **won't work**; the blocking check is the cert policy, and a self-signed entity statement has no chain to a Lissi-known anchor.
5. Note: HAIP is **deprecating `x509_san_dns` → `x509_hash`**; newer Lissi builds may reject the scheme regardless.

### Secondary / interop notes
- Heidi sends the **`proofs` (plural, draft 15+ batch)** credential-request form, not `proof` (singular). Fixed: credential endpoint now accepts both.
- Lissi **rejected PEX** (`presentation_definition`); we run **DCQL** (`use_presentation_definition: false`). Both code paths exist and are switchable.
- DPoP EC verification bug (P1363/DER, ecdh vs ecdsa) and credential-endpoint DPoP scheme acceptance — both fixed during Lissi/Heidi bring-up.

---

## 4. Current state

| Capability | State |
|---|---|
| OID4VCI issuance (auth_code, SD-JWT, cnf) | ✅ Working live (Lissi + Heidi) |
| OID4VP verification (DCQL, encrypted response, KB-JWT) | ✅ Working live (Heidi round-trip COMPLETED) |
| Presentation-definition CRUD API + configdb table | ✅ Built + tested + live |
| Verifier state in runtimedb (encrypted) | ✅ Built + live-verified |
| Console "Verifiable Presentations" UI | ✅ Built, typecheck passes |
| End-to-end issue→verify round-trip | ✅ Achieved with **Heidi** |
| Same round-trip with **Lissi** | ⛔ Blocked on verifier trust list (needs registered RP cert) |

**Runtime config (demo):** verifier `client_id = x509_san_dns:<ngrok-host>`, signing key `vp-verifier` (EC, ngrok host as DNS SAN), `enforce_trusted_issuer + enforce_key_binding = true`, `use_presentation_definition = false` (DCQL). Issuer `credential_issuer = <ngrok>`, signing key `vci-signing` (EC). Trusted issuers: German PID issuer + our own `vci-signing` issuer. Server run with `SKIP_SECURITY=true`, exposed via ngrok with the Go single-origin proxy.

### Recommended next steps
1. **Lock in the working state** — commit OID4VCI package + `proofs` fix + Phase 1 OID4VP management on `openid4vp-management` (Heidi round-trip is the proof).
2. **Lissi:** pursue **SPRIND sandbox RP registration** (or Lissi onboarding) to get a trusted verifier cert — only if Lissi-specific trust is a hard requirement.
3. **Deferred:** declarative-resource export of presentation definitions; OID4VCI tests; claim mapping (later phase).

### Key reference
- Lissi trust model: https://lissi-id.medium.com/trust-in-the-digital-space-7762471351cf
- EUDI WRPAC validation: https://bmi.usercontent.opencode.de/eudi-wallet/eidas-2.0-architekturkonzept/content/ecosystem-architecture/trust/wallet-relying-party-authentication/
- SPRIND sandbox: https://www.lissi.id/blog/germanys-eudi-wallet-sandbox-is-coming-your-guide-to-getting-ready
- Plan: `.claude/plans/shiny-snuggling-horizon.md`
