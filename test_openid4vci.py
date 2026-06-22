#!/usr/bin/env python3
"""
OpenID4VCI credential-offer QR generator.

Requests an issuer-initiated credential offer from Thunder and renders the
openid-credential-offer:// deep link as a QR code for an EUDI wallet to scan.
The wallet then runs the authorization_code flow (login) and fetches the
credential from the issuer.

Usage:
  python test_openid4vci.py [options]

Options:
  --base-url URL    Thunder server URL (default: https://localhost:8090)
  --config-id ID    Credential configuration to offer (default: eudi-pid)
  --png PATH        Also write the QR to a PNG file (e.g. /tmp/offer.png)

Dependencies:
  pip install requests qrcode
"""

import argparse
import json
import sys

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


def fetch_offer(session: requests.Session, base_url: str, config_id: str) -> dict:
    resp = session.get(
        f"{base_url}/openid4vci/offer",
        params={"credential_configuration_id": config_id},
        timeout=10,
    )
    resp.raise_for_status()
    return resp.json()


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
    parser.add_argument("--config-id", default="eudi-pid",
                        help="Credential configuration ID to offer")
    parser.add_argument("--png", default="",
                        help="Optional path to also write the QR as a PNG")
    args = parser.parse_args()

    session = requests.Session()
    session.verify = False

    print(f"\n→ GET {args.base_url}/openid4vci/offer  (config={args.config_id})")
    try:
        data = fetch_offer(session, args.base_url, args.config_id)
    except requests.HTTPError as e:
        sys.exit(f"Offer failed: {e.response.status_code} {e.response.text}")
    except requests.ConnectionError:
        sys.exit(f"Cannot connect to {args.base_url} — is Thunder running?")

    offer = data.get("credential_offer", {})
    offer_uri = data.get("credential_offer_uri", "")
    if not offer_uri:
        sys.exit("No credential_offer_uri in response")

    print("\nCredential offer:")
    print(json.dumps(offer, indent=2))

    print("\nScan with the EUDI Wallet to receive the credential:\n")
    print_qr(offer_uri)
    print(f"\n  offer_uri : {offer_uri}\n")

    if args.png:
        qrcode.make(offer_uri).save(args.png)
        print(f"  PNG written: {args.png}\n")


if __name__ == "__main__":
    main()
