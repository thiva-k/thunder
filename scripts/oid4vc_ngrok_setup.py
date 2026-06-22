#!/usr/bin/env python3
"""
OID4VC + ngrok setup helper.

Run from the Thunder repo root:

    python3 scripts/oid4vc_ngrok_setup.py

It will:
  1. Kill any backend on :8090 and delete the SQLite DBs in
     backend/cmd/server/database (fresh schema; pass --keep-db to skip).
  2. Start ngrok (or reuse a running tunnel) pointing at the backend on :8090.
  3. Generate the issuer (vci-signing) and verifier (vp-verifier) EC signing certs,
     with the verifier cert carrying the current ngrok host as a DNS SAN.
  4. Rewrite backend/cmd/server/deployment.yaml for the runbook flow (public_url,
     gate_client, CORS, crypto keys, openid4vp + openid4vci). Backs up to *.bak.
  5. Copy the gate onto the public origin, then run `SKIP_SECURITY=true make run`
     (which initializes the DB schema, bootstraps the default OU/flows/console/admin,
     and serves backend/console/gate) and import the bundle once it's up. The bundle
     resources depend on those default resources, hence make run before import.
  6. Keep make run + ngrok in the foreground (Ctrl+C stops both). Pass --no-load to
     only do certs/config/ngrok and not start make run.

Prerequisites: ngrok (authenticated), openssl, make, sqlite3 (via make), and the
toolchain make run needs (go, pnpm). Inspector: http://localhost:4040.
"""

import argparse
import json
import os
import shutil
import signal
import ssl
import subprocess
import sys
import time
import urllib.error
import urllib.request
from datetime import datetime, timezone

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
BACKEND_DIR = os.path.join(ROOT, "backend", "cmd", "server")
SECURITY_DIR = os.path.join(BACKEND_DIR, "repository", "resources", "security")
DEPLOYMENT_YAML = os.path.join(BACKEND_DIR, "deployment.yaml")
EUDI_ACCESS_CHAIN = os.path.join(SECURITY_DIR, "eudi-access-chain.crt")
EUDI_ACCESS_KEY = os.path.join(SECURITY_DIR, "eudi-access.key")
EUDI_REG_CERT = os.path.join(SECURITY_DIR, "registration-certificate-undefined.json")
DB_DIR = os.path.join(BACKEND_DIR, "database")
GATE_DIST = os.path.join(ROOT, "frontend", "apps", "gate", "dist")
GATE_SERVE_DIR = os.path.join(BACKEND_DIR, "apps", "gate")
BUNDLE_YAML = os.path.join(ROOT, "thunderid-config.yaml")
BUNDLE_ENV = os.path.join(ROOT, "thunderid.env")
MAKERUN_LOG = os.path.join(BACKEND_DIR, "setup-makerun.log")
NGROK_API = "http://127.0.0.1:4040/api/tunnels"
INSPECTOR_URL = "http://localhost:4040"
CONSOLE_URL = "https://localhost:5191/console"

SSL_CTX = ssl.create_default_context()
SSL_CTX.check_hostname = False
SSL_CTX.verify_mode = ssl.CERT_NONE


def fail(msg):
    print("ERROR: " + msg, file=sys.stderr)
    sys.exit(1)


def kill_port(port):
    """Terminate any process listening on the given TCP port."""
    if shutil.which("lsof") is None:
        return
    out = subprocess.run(["lsof", "-ti", "tcp:%d" % port, "-sTCP:LISTEN"], capture_output=True, text=True)
    pids = [p for p in out.stdout.split() if p.strip()]
    for pid in pids:
        print("  killing pid %s on :%d" % (pid, port))
        subprocess.run(["kill", pid])
    if pids:
        time.sleep(1)


def reset_databases():
    """Delete the SQLite DB files (and WAL/SHM sidecars); make run recreates the schema."""
    if not os.path.isdir(DB_DIR):
        print("  no database dir at %s" % os.path.relpath(DB_DIR, ROOT))
        return
    removed = 0
    for fn in os.listdir(DB_DIR):
        if fn.endswith(".db") or ".db-" in fn:
            try:
                os.remove(os.path.join(DB_DIR, fn))
                removed += 1
            except OSError:
                pass
    print("  removed %d db file(s) from %s" % (removed, os.path.relpath(DB_DIR, ROOT)))


def get_running_tunnel():
    """Return the https public_url of a running ngrok tunnel, or None."""
    try:
        with urllib.request.urlopen(NGROK_API, timeout=2) as resp:
            data = json.loads(resp.read().decode())
    except Exception:
        return None
    for tunnel in data.get("tunnels", []):
        url = tunnel.get("public_url", "")
        if url.startswith("https://"):
            return url
    return None


def start_ngrok(port):
    """Start ngrok pointing at the backend and wait for the tunnel URL."""
    if shutil.which("ngrok") is None:
        fail("ngrok not found on PATH. Install it and run `ngrok config add-authtoken <token>`.")
    print("Starting ngrok -> https://localhost:%d ..." % port)
    proc = subprocess.Popen(
        ["ngrok", "http", "https://localhost:%d" % port, "--host-header=localhost"],
        stdout=subprocess.DEVNULL,
        stderr=subprocess.DEVNULL,
    )
    for _ in range(40):
        time.sleep(0.5)
        url = get_running_tunnel()
        if url:
            return url, proc
        if proc.poll() is not None:
            fail("ngrok exited unexpectedly. Try `ngrok http https://localhost:%d` manually." % port)
    proc.terminate()
    fail("Timed out waiting for the ngrok tunnel. Is ngrok authenticated?")


def run_openssl(args):
    result = subprocess.run(["openssl"] + args, capture_output=True, text=True)
    if result.returncode != 0:
        fail("openssl %s failed:\n%s" % (" ".join(args), result.stderr.strip()))


def gen_cert(name, subject, host, force):
    """Generate an EC P-256 key + self-signed cert. host adds a DNS SAN when set."""
    key = os.path.join(SECURITY_DIR, name + ".key")
    cert = os.path.join(SECURITY_DIR, name + ".cert")
    if os.path.exists(cert) and os.path.exists(key) and not force:
        print("  %s: present, keeping" % name)
        return
    run_openssl(["ecparam", "-name", "prime256v1", "-genkey", "-noout", "-out", key])
    req = ["req", "-new", "-x509", "-key", key, "-out", cert, "-days", "365", "-subj", subject]
    if host:
        req += ["-addext", "subjectAltName=DNS:" + host]
    run_openssl(req)
    print("  %s: generated%s" % (name, (" (SAN=DNS:%s)" % host) if host else ""))


def compute_x509_hash(cert_pem_path):
    """Return the base64url SHA-256 of the DER leaf cert (no padding) for x509_hash client_id."""
    import hashlib
    import base64
    result = subprocess.run(
        ["openssl", "x509", "-outform", "DER", "-in", cert_pem_path],
        capture_output=True,
    )
    if result.returncode != 0:
        fail("openssl x509 failed: %s" % result.stderr.decode().strip())
    digest = hashlib.sha256(result.stdout).digest()
    return base64.urlsafe_b64encode(digest).rstrip(b"=").decode()


DEPLOYMENT_TEMPLATE = """\
# Generated by scripts/oid4vc_ngrok_setup.py on __GENERATED_AT__
# Tailored for the OID4VC end-to-end runbook (docs/working/oid4vc-testing.md).
server:
  hostname: "localhost"
  port: __PORT__
  public_url: "__URL__"

gate_client:
  hostname: "__HOST__"
  port: 443
  scheme: "https"
  login_path: "gate/signin"
  error_path: "gate/error"

tls:
  min_version: "1.3"
  cert_file: "config/certs/server.cert"
  key_file: "config/certs/server.key"

database:
  config:
    type: "sqlite"
    sqlite:
      path: "database/configdb.db"
      options: "_journal_mode=WAL&_busy_timeout=5000&_pragma=foreign_keys(1)"
      max_open_conns: 500
      max_idle_conns: 100
      conn_max_lifetime: 3600
  runtime:
    type: "sqlite"
    sqlite:
      path: "database/runtimedb.db"
      options: "_journal_mode=WAL&_busy_timeout=5000&_pragma=foreign_keys(1)"
      max_open_conns: 500
      max_idle_conns: 100
      conn_max_lifetime: 3600
  user:
    type: "sqlite"
    sqlite:
      path: "database/userdb.db"
      options: "_journal_mode=WAL&_busy_timeout=5000&_pragma=foreign_keys(1)"
      max_open_conns: 500
      max_idle_conns: 100
      conn_max_lifetime: 3600

crypto:
  encryption:
    key: "file://config/certs/crypto.key"
  password_hashing:
    algorithm: "PBKDF2"
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
__EUDI_KEY_ENTRY__
jwt:
  preferred_key_id: "default-key"

cors:
  allowed_origins:
    - "__URL__"
    - "https://localhost:3000"
    - "https://localhost:5190"
    - "https://localhost:5191"

passkey:
  allowed_origins:
    - "https://localhost:8090"

consent:
  enabled: true
  base_url: "http://localhost:9090/api/v1"

email:
  smtp:
    host: "127.0.0.1"
    port: 2525
    username: "dev"
    password: "dev"
    from_address: "noreply@thunderid.dev"
    enable_start_tls: false
    enable_authentication: true

# Issuer signs the SD-JWT VC with this key (EC/ES256). Disabled in default.json;
# enabled here so the issuer engine starts.
openid4vci:
  signing_key_id: "vci-signing"

__OPENID4VP_BLOCK__"""


def write_deployment(host, url, port, force_san_dns=False):
    if shutil.which("openssl") is None:
        fail("openssl not found on PATH.")
    if os.path.exists(DEPLOYMENT_YAML):
        backup = DEPLOYMENT_YAML + ".bak"
        shutil.copy2(DEPLOYMENT_YAML, backup)
        print("Backed up existing deployment.yaml -> %s" % os.path.relpath(backup, ROOT))

    eudi_present = (
        not force_san_dns
        and os.path.exists(EUDI_ACCESS_CHAIN) and os.path.exists(EUDI_ACCESS_KEY)
    )
    if eudi_present:
        x509_hash = compute_x509_hash(EUDI_ACCESS_CHAIN)
        eudi_key_entry = (
            '    - id: "eudi-access-key"\n'
            '      cert_file: "repository/resources/security/eudi-access-chain.crt"\n'
            '      key_file: "repository/resources/security/eudi-access.key"\n'
        )
        reg_cert_line = ""
        if os.path.exists(EUDI_REG_CERT):
            reg_cert_name = os.path.basename(EUDI_REG_CERT)
            reg_cert_line = (
                '\n  registration_cert_file: "repository/resources/security/%s"' % reg_cert_name
            )
        openid4vp_block = (
            "# Verifier signs its request object (JAR) with the EUDI sandbox access key.\n"
            "# client_id uses x509_hash of the leaf cert — wallet validates via German Registrar CA chain.\n"
            "openid4vp:\n"
            '  client_id: "x509_hash:%s"\n'
            '  signing_key_id: "eudi-access-key"%s\n'
        ) % (x509_hash, reg_cert_line)
        print("  EUDI access cert detected — using x509_hash client_id (%s)" % x509_hash)
    else:
        eudi_key_entry = ""
        openid4vp_block = (
            "# Verifier signs its request object (JAR) with vp-verifier; the wallet validates\n"
            "# the cert's DNS SAN against this client_id host.\n"
            "openid4vp:\n"
            '  client_id: "x509_san_dns:%s"\n'
            '  signing_key_id: "vp-verifier"\n'
        ) % host
        if force_san_dns:
            print("  --heidi mode — using x509_san_dns client_id (ignoring EUDI cert)")
        else:
            print("  No EUDI access cert found — falling back to x509_san_dns client_id")

    content = (
        DEPLOYMENT_TEMPLATE.replace("__GENERATED_AT__", datetime.now(timezone.utc).isoformat())
        .replace("__PORT__", str(port))
        .replace("__URL__", url)
        .replace("__HOST__", host)
        .replace("__EUDI_KEY_ENTRY__", eudi_key_entry)
        .replace("__OPENID4VP_BLOCK__", openid4vp_block)
    )
    with open(DEPLOYMENT_YAML, "w") as f:
        f.write(content)
    print("Wrote %s" % os.path.relpath(DEPLOYMENT_YAML, ROOT))


GATE_CONFIG_TEMPLATE = """\
window.__THUNDERID_RUNTIME_CONFIG__ = {
  brand: {
    product_name: 'ThunderID',
    favicon: {
      light: 'assets/images/favicon.ico',
      dark: 'assets/images/favicon-inverted.ico',
    },
  },
  client: {
    base: '/gate',
  },
  server: {
    hostname: '__HOST__',
    port: 443,
    http_only: false,
  },
};
"""


def copy_gate(host):
    """Serve the gate on the public origin (the backend serves /gate from its home)."""
    if not os.path.isdir(GATE_DIST):
        print("  gate dist not found (%s); issuance scans need it. Run once more after the first"
              " make run builds the frontend." % os.path.relpath(GATE_DIST, ROOT))
        return
    if os.path.isdir(GATE_SERVE_DIR):
        shutil.rmtree(GATE_SERVE_DIR)
    shutil.copytree(GATE_DIST, GATE_SERVE_DIR)
    with open(os.path.join(GATE_SERVE_DIR, "config.js"), "w") as f:
        f.write(GATE_CONFIG_TEMPLATE.replace("__HOST__", host))
    print("  gate copied to %s (server -> %s:443)" % (os.path.relpath(GATE_SERVE_DIR, ROOT), host))


def start_make_run(port):
    """Start `SKIP_SECURITY=true make run` and wait until the servers are up."""
    if shutil.which("make") is None:
        fail("make not found on PATH.")
    env = os.environ.copy()
    env["SKIP_SECURITY"] = "true"  # restored from env by build.sh, so the final servers stay open
    log = open(MAKERUN_LOG, "w")
    print("  running `SKIP_SECURITY=true make run` (log: %s)" % os.path.relpath(MAKERUN_LOG, ROOT))
    proc = subprocess.Popen(
        ["make", "run"], cwd=ROOT, env=env, stdout=log, stderr=subprocess.STDOUT, start_new_session=True
    )
    print("  building + bootstrapping (this can take a few minutes) ...")
    deadline = time.time() + 900
    while time.time() < deadline:
        if proc.poll() is not None:
            fail("make run exited early; see %s" % os.path.relpath(MAKERUN_LOG, ROOT))
        try:
            with open(MAKERUN_LOG) as f:
                if "Servers running" in f.read():
                    break
        except OSError:
            pass
        time.sleep(2)
    else:
        stop_make_run(proc, port)
        fail("make run did not report 'Servers running'; see %s" % os.path.relpath(MAKERUN_LOG, ROOT))
    # Confirm the backend is actually reachable.
    health = "https://localhost:%d/.well-known/openid-credential-issuer" % port
    for _ in range(60):
        try:
            urllib.request.urlopen(health, context=SSL_CTX, timeout=2)
            break
        except urllib.error.HTTPError:
            break
        except Exception:
            time.sleep(1)
    print("  servers are up")
    return proc


def stop_make_run(proc, port):
    """Terminate the make run process group and free the server ports."""
    print("Stopping make run ...")
    try:
        os.killpg(os.getpgid(proc.pid), signal.SIGTERM)
    except (ProcessLookupError, PermissionError):
        proc.terminate()
    try:
        proc.wait(timeout=10)
    except subprocess.TimeoutExpired:
        try:
            os.killpg(os.getpgid(proc.pid), signal.SIGKILL)
        except (ProcessLookupError, PermissionError):
            proc.kill()
    for p in (port, 5190, 5191):
        kill_port(p)


def parse_env_vars():
    """Parse thunderid.env into a template-variable map (array values decoded as JSON)."""
    variables = {}
    if not os.path.exists(BUNDLE_ENV):
        return variables
    for line in open(BUNDLE_ENV):
        line = line.rstrip()
        if "=" not in line or line.startswith("#"):
            continue
        key, _, value = line.partition("=")
        key, value = key.strip(), value.strip()
        try:
            variables[key] = json.loads(value)
        except ValueError:
            variables[key] = value
    return variables


def update_bundle_env(url):
    """Write NGROK_URL into thunderid.env so the bundle template resolves the resource server identifier."""
    lines = []
    replaced = False
    if os.path.exists(BUNDLE_ENV):
        for line in open(BUNDLE_ENV):
            if line.startswith("NGROK_URL="):
                lines.append("NGROK_URL=%s\n" % url)
                replaced = True
            else:
                lines.append(line)
    if not replaced:
        lines.insert(0, "NGROK_URL=%s\n" % url)
    with open(BUNDLE_ENV, "w") as f:
        f.writelines(lines)
    print("  wrote NGROK_URL=%s to %s" % (url, os.path.relpath(BUNDLE_ENV, ROOT)))


def import_bundle(port):
    """Import thunderid-config.yaml (user type, subject, wallet apps, credential config, PD)."""
    if not os.path.exists(BUNDLE_YAML):
        print("  bundle %s not found; skipping import" % os.path.relpath(BUNDLE_YAML, ROOT))
        return
    payload = {"content": open(BUNDLE_YAML).read(), "variables": parse_env_vars(), "options": {"upsert": True}}
    req = urllib.request.Request(
        "https://localhost:%d/import" % port,
        data=json.dumps(payload).encode(),
        headers={"Content-Type": "application/json"},
        method="POST",
    )
    try:
        with urllib.request.urlopen(req, context=SSL_CTX, timeout=60) as resp:
            body = json.loads(resp.read().decode())
    except urllib.error.HTTPError as e:
        print("  import failed (HTTP %s): %s" % (e.code, e.read().decode()[:300]))
        return
    summary = body.get("summary", {})
    print("  imported %d, failed %d" % (summary.get("imported", 0), summary.get("failed", 0)))
    for r in body.get("results", []):
        if r.get("status") != "success":
            print("    ! %s %s: %s" % (r.get("resourceType"), r.get("status"), r.get("message", "")))


def main():
    parser = argparse.ArgumentParser(description="Set up certs + deployment.yaml + ngrok, then make run + import.")
    parser.add_argument("--port", type=int, default=8090, help="Backend HTTPS port (default 8090).")
    parser.add_argument("--regen-issuer", action="store_true", help="Force-regenerate the vci-signing cert too.")
    parser.add_argument("--keep-db", action="store_true", help="Keep the existing SQLite databases.")
    parser.add_argument("--no-load", action="store_true", help="Only certs/config/ngrok; don't run make run or import.")
    wallet = parser.add_mutually_exclusive_group()
    wallet.add_argument(
        "--heidi", action="store_true",
        help="VP verifier: x509_san_dns + self-signed cert (works with Heidi/Lissi; ignores EUDI access cert).",
    )
    wallet.add_argument(
        "--eudi", action="store_true",
        help="VP verifier: x509_hash + EUDI access cert (works with IDGo/Sprind). Default when cert files are present.",
    )
    args = parser.parse_args()

    os.makedirs(SECURITY_DIR, exist_ok=True)

    print("Resetting local environment ...")
    kill_port(args.port)
    if args.keep_db:
        print("  keeping existing databases (--keep-db)")
    else:
        reset_databases()

    started = None
    url = get_running_tunnel()
    if url:
        print("Reusing running ngrok tunnel: %s" % url)
    else:
        url, started = start_ngrok(args.port)
    host = url.split("://", 1)[1].rstrip("/")
    print("ngrok host: %s" % host)

    print("Updating bundle env ...")
    update_bundle_env(url)

    print("Generating signing certs in %s" % os.path.relpath(SECURITY_DIR, ROOT))
    gen_cert("vci-signing", "/CN=Thunder VCI Issuer", None, args.regen_issuer)
    gen_cert("vp-verifier", "/CN=Thunder VP Verifier", host, True)

    write_deployment(host, url, args.port, force_san_dns=args.heidi)

    make_run = None
    if not args.no_load:
        print("Serving the gate on the public origin ...")
        copy_gate(host)
        make_run = start_make_run(args.port)
        print("Importing the bundle ...")
        import_bundle(args.port)

    print("")
    print("=" * 64)
    print("  Public URL : %s" % url)
    print("  Inspector  : %s" % INSPECTOR_URL)
    if make_run is not None:
        print("  Console    : %s  (login admin/admin)" % CONSOLE_URL)
    print("=" * 64)

    if make_run is not None:
        print("Servers + ngrok running. In the Console use the QR actions:")
        print("    Verifiable Credentials  -> QR (offer)  = issue")
        print("    Verifiable Presentations -> Verify     = present + live status")
        print("Press Ctrl+C to stop everything.")
        try:
            make_run.wait()
        except KeyboardInterrupt:
            print("")
            stop_make_run(make_run, args.port)
            if started is not None:
                started.terminate()
        return

    print("Skipped make run (--no-load). Start it yourself with `make run`.")
    if started is not None:
        print("ngrok is running here; Ctrl+C to stop.")
        try:
            started.wait()
        except KeyboardInterrupt:
            started.terminate()


if __name__ == "__main__":
    main()
