#!/usr/bin/env python3
"""
OpenID4VP REST API smoke-test.

Initiates a transaction, renders the wallet QR in the terminal, polls for
completion, then decodes and prints the result token payload.

Usage:
  python scripts/test_openid4vp.py [options]

Options:
  --base-url URL         Thunder server URL (default: https://localhost:8090)
  --definition-id ID     Presentation definition (default: eudi-pid)
  --rp-id ID             Relying party identifier (default: test-rp)
  --poll-interval SECS   Polling interval (default: 2.0)

Dependencies:
  pip install requests qrcode
"""

import argparse
import base64
import json
import sys
import time

import urllib3

try:
    import requests
except ImportError:
    sys.exit("Missing dependency — run:  pip install requests qrcode")

try:
    import qrcode
except ImportError:
    sys.exit("Missing dependency — run:  pip install requests qrcode")

urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)


def initiate(session: requests.Session, base_url: str, definition_id: str, rp_id: str) -> dict:
    resp = session.post(
        f"{base_url}/openid4vp/initiate",
        json={"definition_id": definition_id, "rp_id": rp_id},
        timeout=10,
    )
    resp.raise_for_status()
    return resp.json()


def poll_status(session: requests.Session, base_url: str, txn_id: str) -> dict:
    resp = session.get(
        f"{base_url}/openid4vp/status/{txn_id}",
        timeout=10,
    )
    resp.raise_for_status()
    return resp.json()


def decode_jwt_payload(token: str) -> dict:
    parts = token.split(".")
    if len(parts) != 3:
        raise ValueError("not a valid compact JWT")
    padded = parts[1] + "=" * (4 - len(parts[1]) % 4)
    return json.loads(base64.urlsafe_b64decode(padded))


def print_qr(data: str) -> None:
    qr = qrcode.QRCode(border=1)
    qr.add_data(data)
    qr.make(fit=True)
    qr.print_ascii(invert=True)


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__,
                                     formatter_class=argparse.RawDescriptionHelpFormatter)
    parser.add_argument("--base-url", default="https://localhost:8090",
                        help="Thunder server base URL")
    parser.add_argument("--definition-id", default="eudi-pid",
                        help="Presentation definition ID")
    parser.add_argument("--rp-id", default="test-rp",
                        help="Relying party identifier")
    parser.add_argument("--poll-interval", type=float, default=2.0,
                        help="Status poll interval in seconds")
    args = parser.parse_args()

    session = requests.Session()
    session.verify = False

    # ── 1. Initiate ──────────────────────────────────────────────────────────
    print(f"\n→ POST {args.base_url}/openid4vp/initiate  (definition={args.definition_id})")
    try:
        data = initiate(session, args.base_url, args.definition_id, args.rp_id)
    except requests.HTTPError as e:
        sys.exit(f"Initiate failed: {e.response.status_code} {e.response.text}")
    except requests.ConnectionError:
        sys.exit(f"Cannot connect to {args.base_url} — is Thunder running?")

    txn_id     = data["txn_id"]
    wallet_url = data["wallet_url"]
    status_url = data["status_url"]
    expires_at = data["expires_at"]

    print(f"  txn_id     : {txn_id}")
    print(f"  status_url : {status_url}")
    print(f"  expires_at : {expires_at}")

    # ── 2. QR code ───────────────────────────────────────────────────────────
    print("\nScan with the EUDI Wallet:\n")
    print_qr(wallet_url)
    print(f"\n  wallet_url : {wallet_url}\n")

    # ── 3. Poll ──────────────────────────────────────────────────────────────
    print("Polling for wallet response", end="", flush=True)
    while True:
        time.sleep(args.poll_interval)
        try:
            status_data = poll_status(session, args.base_url, txn_id)
        except requests.HTTPError as e:
            print()
            sys.exit(f"Status poll failed: {e.response.status_code} {e.response.text}")

        status = status_data.get("status", "UNKNOWN")

        if status == "PENDING":
            print(".", end="", flush=True)
            continue

        print()

        if status == "COMPLETED":
            print("\n✓ COMPLETED\n")
            result_token = status_data.get("result_token", "")
            if not result_token:
                print("  (no result_token in response)")
                break

            print("─" * 60)
            print("Result token (raw JWT):")
            print(f"  {result_token}")
            print()
            print("Result token — decoded header:")
            try:
                header = decode_jwt_payload(result_token.split(".")[0] + ".placeholder.sig")
            except Exception:
                pass
            try:
                raw_header = result_token.split(".")[0]
                padded = raw_header + "=" * (4 - len(raw_header) % 4)
                header = json.loads(base64.urlsafe_b64decode(padded))
                print(json.dumps(header, indent=2))
            except Exception as e:
                print(f"  (decode error: {e})")

            print()
            print("Result token — decoded payload:")
            try:
                payload = decode_jwt_payload(result_token)
                print(json.dumps(payload, indent=2))
            except Exception as e:
                print(f"  (decode error: {e})")
            print("─" * 60)
            break

        elif status == "FAILED":
            error = status_data.get("error", "unknown")
            print(f"\n✗ FAILED: {error}")
            sys.exit(1)

        elif status == "EXPIRED":
            print("\n✗ EXPIRED: transaction timed out before wallet responded")
            sys.exit(1)

        else:
            print(f"\n? Unexpected status: {status}")
            print(json.dumps(status_data, indent=2))
            sys.exit(1)


if __name__ == "__main__":
    main()
