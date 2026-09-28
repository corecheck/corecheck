#!/usr/bin/env python3
"""Render the dashboards and start local Grafana against dev CloudWatch.

Edit the templates in deploy/terraform/monitoring/grafana-dashboard-templates.
This process re-renders them, and Grafana reloads the files within a few seconds.
Queries still run against the real dev log groups and metrics.
"""

from __future__ import annotations

import argparse
import os
import signal
import subprocess
import sys
import time
from pathlib import Path

from render import GENERATED, TEMPLATES, lookup_account, render_all

HERE = Path(__file__).resolve().parent
AWS_ENV = GENERATED / "aws.env"


def export_credentials(profile: str | None, region: str) -> None:
    cmd = ["aws", "configure", "export-credentials", "--format", "env-no-export"]
    if profile:
        cmd.extend(["--profile", profile])
    proc = subprocess.run(cmd, check=False, capture_output=True, text=True)
    if proc.returncode != 0:
        cmd[cmd.index("env-no-export")] = "env"
        proc = subprocess.run(cmd, check=False, capture_output=True, text=True)
    if proc.returncode != 0:
        detail = proc.stderr.strip() or proc.stdout.strip()
        raise SystemExit(f"could not export AWS credentials ({detail})")

    lines: list[str] = []
    for raw in proc.stdout.splitlines():
        line = raw.strip()
        if line.startswith("export "):
            line = line[len("export ") :]
        if not line or "=" not in line:
            continue
        key = line.split("=", 1)[0]
        if key in {"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY", "AWS_SESSION_TOKEN"}:
            lines.append(line)
    if not any(line.startswith("AWS_ACCESS_KEY_ID=") for line in lines):
        raise SystemExit("AWS credential export did not include AWS_ACCESS_KEY_ID")
    lines.append(f"AWS_REGION={region}")
    lines.append(f"AWS_DEFAULT_REGION={region}")
    AWS_ENV.parent.mkdir(parents=True, exist_ok=True)
    AWS_ENV.write_text("\n".join(lines) + "\n")
    AWS_ENV.chmod(0o600)


def compose(args: list[str]) -> int:
    return subprocess.call(["docker", "compose", *args], cwd=HERE)


def watch(workspace: str, region: str, account: str, proc: subprocess.Popen[bytes]) -> None:
    templates = sorted(TEMPLATES.glob("*.json.tftpl"))
    seen = {path: path.stat().st_mtime_ns for path in templates}
    while proc.poll() is None:
        time.sleep(1)
        changed = False
        for path in templates:
            mtime = path.stat().st_mtime_ns
            if seen.get(path) != mtime:
                seen[path] = mtime
                changed = True
        if not changed:
            continue
        try:
            written = render_all(workspace, region, account)
        except Exception as exc:
            print(f"render failed: {exc}", file=sys.stderr)
            continue
        if written:
            print("re-rendered dashboards")


def serve(args: argparse.Namespace) -> int:
    account = args.account or lookup_account(args.profile)
    render_all(args.workspace, args.region, account)
    export_credentials(args.profile, args.region)
    print("GitHub  http://localhost:3000/d/corecheck-github-overview")
    print("Tests   http://localhost:3000/d/corecheck-tests")
    print("Benches http://localhost:3000/d/corecheck-benchmarks")
    print("Jobs    http://localhost:3000/d/corecheck-jobs")
    proc = subprocess.Popen(["docker", "compose", "up", "--remove-orphans"], cwd=HERE)
    try:
        if args.no_watch:
            return proc.wait()
        watch(args.workspace, args.region, account, proc)
        return proc.wait()
    except KeyboardInterrupt:
        return 0
    finally:
        if proc.poll() is None:
            proc.send_signal(signal.SIGINT)
            try:
                proc.wait(timeout=30)
            except subprocess.TimeoutExpired:
                proc.kill()
                proc.wait()


def parse_args(argv: list[str]) -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--workspace", default="dev")
    parser.add_argument("--region", default="ap-south-1")
    parser.add_argument("--account", help="AWS account id. Looked up with sts when omitted")
    parser.add_argument("--profile", help="AWS CLI profile for credentials and the account lookup")
    parser.add_argument("--no-watch", action="store_true", help="Do not re-render when the templates change")
    parser.add_argument("command", nargs="?", choices=("up", "down"), default="up")
    return parser.parse_args(argv)


def main(argv: list[str] | None = None) -> None:
    args = parse_args(sys.argv[1:] if argv is None else argv)
    if args.command == "down":
        raise SystemExit(compose(["down"]))
    raise SystemExit(serve(args))


if __name__ == "__main__":
    os.chdir(HERE)
    main()
