# Skyline Lounge — Sky Pass verifier

A standalone **OpenID4VP verifier** kiosk that completes the verifiable-credentials
story: Wayfinder *issues* the Wayfinder Sky Pass into a guest's wallet, and the
Skyline Lounge — a **separate relying party** — *verifies* it at the door and
grants or denies entry based on the verified loyalty tier.

The wallet is the only thing that travels between the two apps; the lounge never
talks to Wayfinder. It is a thin client of Thunder's OpenID4VP API, so it needs
no OAuth client registration — the request object is signed by Thunder's verifier
key (`x509_san_dns`), which the wallet trusts.

## How it works

1. Guest taps **Present Sky Pass** → the lounge calls Thunder
   `POST /openid4vp/initiate` (`definition_id=wayfinder-skypass`) and shows the
   returned wallet URL as a QR.
2. Guest scans it and approves the disclosure in their wallet.
3. The lounge polls `GET /openid4vp/status/{txn_id}`; on `COMPLETED` it decodes
   the signed result token, then shows **"Welcome, <name> — <tier> member"** and
   **access granted** if the verified `tier` is in `ALLOWED_TIERS` (Gold/Platinum
   by default), otherwise **access denied**.

## Run

Prereqs: a running Thunder with the `wayfinder-skypass` presentation definition,
reachable at a public URL (the tunnel from `scripts/oid4vc_ngrok_setup.py`), and a
guest who has already been issued the Sky Pass in their wallet.

```bash
cp .env.example .env          # set THUNDER_BASE_URL to your tunnel URL
node --env-file=.env server.js
# or: THUNDER_BASE_URL=https://<tunnel> npm start
```

Open http://localhost:8795 and present a Sky Pass.

## Configuration (`.env`)

| Variable | Default | Purpose |
|---|---|---|
| `THUNDER_BASE_URL` | — | Thunder's public URL (the tunnel). Required. |
| `PORT` | `8795` | Kiosk port. |
| `SKYPASS_DEFINITION_ID` | `wayfinder-skypass` | Presentation definition to verify. |
| `ALLOWED_TIERS` | `Gold,Platinum` | Tiers granted lounge access. |
