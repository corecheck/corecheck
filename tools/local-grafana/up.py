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
AWS_DIR = GENERATED
AWS_CREDENTIALS = AWS_DIR / "aws-credentials.json"
AWS_CONFIG = AWS_DIR / "config"


def write_aws_config(region: str, profile: str | None) -> None:
    """Grafana reads this on every CloudWatch request and reloads it when the token expires.

    The aws-credentials container rewrites aws-credentials.json from the host aws login
    session, which stays valid for up to 12 hours. A one-shot env file expires in 15 minutes.
    """
    AWS_DIR.mkdir(parents=True, exist_ok=True)
    AWS_CONFIG.write_text(
        "[default]\n"
        f"region = {region}\n"
        "credential_process = /bin/cat /etc/grafana/aws/aws-credentials.json\n"
    )
    (AWS_DIR / "empty-credentials").write_text("")
    cmd = ["aws", "configure", "export-credentials", "--format", "process"]
    if profile:
        cmd.extend(["--profile", profile])
    proc = subprocess.run(cmd, check=False, capture_output=True, text=True)
    if proc.returncode != 0:
        detail = proc.stderr.strip() or proc.stdout.strip()
        raise SystemExit(f"could not export AWS credentials ({detail})")
    AWS_CREDENTIALS.write_text(proc.stdout)
    AWS_CREDENTIALS.chmod(0o644)


def compose_env() -> dict[str, str]:
    # Compose's attached view uses cursor control codes. Fish in iTerm2
    # leaves the screen corrupted, so keep every compose command line-based.
    env = os.environ.copy()
    env["COMPOSE_ANSI"] = "never"
    return env


def compose(args: list[str]) -> int:
    return subprocess.call(["docker", "compose", *args], cwd=HERE, env=compose_env())


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
            print(f"render failed: {exc}", file=sys.stderr, flush=True)
            continue
        if written:
            print("re-rendered dashboards", flush=True)


def serve(args: argparse.Namespace) -> int:
    account = args.account or lookup_account(args.profile)
    render_all(args.workspace, args.region, account)
    write_aws_config(args.region, args.profile)
    print("GitHub  http://localhost:3000/d/corecheck-github-overview", flush=True)
    print("Tests   http://localhost:3000/d/corecheck-tests", flush=True)
    print("Benches http://localhost:3000/d/corecheck-benchmarks", flush=True)
    print("Jobs    http://localhost:3000/d/corecheck-jobs", flush=True)
    started = compose(["up", "-d", "--remove-orphans", "--quiet-pull"])
    if started != 0:
        return started
    print("Grafana is running. Ctrl-C stops it.", flush=True)
    logs = subprocess.Popen(
        ["docker", "compose", "logs", "-f", "--no-color", "--tail", "40"],
        cwd=HERE,
        env=compose_env(),
    )
    try:
        if args.no_watch:
            return logs.wait()
        watch(args.workspace, args.region, account, logs)
        return logs.wait()
    except KeyboardInterrupt:
        return 0
    finally:
        if logs.poll() is None:
            logs.send_signal(signal.SIGINT)
            try:
                logs.wait(timeout=5)
            except subprocess.TimeoutExpired:
                logs.kill()
                logs.wait()
        compose(["stop"])


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
