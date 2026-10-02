#!/usr/bin/env python3
"""Render the public Grafana dashboard templates for local Grafana.

The substitutions mirror deploy/terraform/monitoring/dashboard_definitions.tf
and the namespace rule in deploy/terraform/monitoring/dashboard.tf. Terraform's
templatefile syntax used here is ${name} plus $${ to emit a literal ${.
"""

from __future__ import annotations

import argparse
import json
import re
import subprocess
import sys
import tempfile
from pathlib import Path

HERE = Path(__file__).resolve().parent
REPO = HERE.parents[1]
TEMPLATES = REPO / "deploy/terraform/monitoring/grafana-dashboard-templates"
GENERATED = HERE / "generated"

# Keys match local.public_dashboard_templates in dashboard_definitions.tf.
DASHBOARDS = {
    "github": {
        "template": "github-overview.json.tftpl",
        "route": "/",
        "title": "Corecheck GitHub Overview",
        "datadog_dashboard_id": "b5p-ekj-pxv",
        "datadog_dashboard_title": "Bitcoin Core GitHub Overview",
        "datadog_widget_count": 2,
        "public_dashboard_env": "PUBLIC_DASHBOARD_GITHUB_URL",
    },
    "tests": {
        "template": "tests.json.tftpl",
        "route": "/tests",
        "title": "Corecheck Tests",
        "datadog_dashboard_id": "7ck-zbu-au3",
        "datadog_dashboard_title": "Bitcoin Core tests",
        "datadog_widget_count": 5,
        "public_dashboard_env": "PUBLIC_DASHBOARD_TESTS_URL",
    },
    "benchmarks": {
        "template": "benchmarks.json.tftpl",
        "route": "/benchmarks",
        "title": "Corecheck Benchmarks",
        "datadog_dashboard_id": "qem-ga2-953",
        "datadog_dashboard_title": "Bitcoin Core benchmarks",
        "datadog_widget_count": 6,
        "public_dashboard_env": "PUBLIC_DASHBOARD_BENCHMARKS_URL",
    },
    "jobs": {
        "template": "jobs.json.tftpl",
        "route": "/jobs",
        "title": "Corecheck Jobs",
        "datadog_dashboard_id": "sxt-cxy-nsc",
        "datadog_dashboard_title": "Corecheck job executions",
        "datadog_widget_count": 35,
        "public_dashboard_env": "PUBLIC_DASHBOARD_JOBS_URL",
    },
}

IDENT = re.compile(r"[A-Za-z_][A-Za-z0-9_]*\Z")
LEFTOVER = re.compile(r"\$\{([^{}]+)\}")


def render_template(source: str, variables: dict[str, str]) -> str:
    """Apply the subset of Terraform template syntax these dashboards use."""
    out: list[str] = []
    i = 0
    while i < len(source):
        if source.startswith("$${", i):
            out.append("${")
            i += 3
            continue
        if source.startswith("${", i):
            end = source.find("}", i + 2)
            if end == -1:
                raise ValueError("unclosed ${ interpolation")
            key = source[i + 2 : end].strip()
            if not IDENT.match(key):
                raise ValueError(f"unsupported template expression: ${{{key}}}")
            if key not in variables:
                raise KeyError(key)
            out.append(variables[key])
            i = end + 1
            continue
        out.append(source[i])
        i += 1
    return "".join(out)


def context(workspace: str, region: str, account: str, dashboard: dict) -> dict[str, str]:
    namespace = "Corecheck/prod" if workspace == "default" else f"Corecheck/{workspace}"
    shared = {
        "telemetry_namespace": namespace,
        "compute_region": region,
        "github_events_log_group": f"/corecheck/github-events/{workspace}",
        "test_results_log_group": f"/corecheck/test-results/{workspace}",
        "benchmark_results_log_group": f"/corecheck/benchmark-results/{workspace}",
        "workflow_state_machine_arn": (
            f"arn:aws:states:{region}:{account}:stateMachine:start-jobs-{workspace}"
        ),
        "mutation_state_machine_arn": (
            f"arn:aws:states:{region}:{account}:stateMachine:start-mutation-jobs-{workspace}"
        ),
        "coverage_queue_name": f"coverage-queue-{workspace}",
        "sonar_queue_name": f"sonar-queue-{workspace}",
        "bench_queue_name": f"bench-queue-{workspace}",
        "mutation_queue_name": f"mutation-queue-{workspace}",
    }
    specific = {
        "route": dashboard["route"],
        "title": dashboard["title"],
        "datadog_dashboard_id": dashboard["datadog_dashboard_id"],
        "datadog_dashboard_title": dashboard["datadog_dashboard_title"],
        "datadog_widget_count": str(dashboard["datadog_widget_count"]),
        "public_dashboard_env": dashboard["public_dashboard_env"],
    }
    return {**shared, **specific}


def datasource_yaml(region: str) -> str:
    return f"""apiVersion: 1
datasources:
  - name: Corecheck CloudWatch
    uid: corecheck-cloudwatch
    type: cloudwatch
    access: proxy
    editable: true
    jsonData:
      authType: default
      defaultRegion: {region}
  - name: Corecheck test names
    uid: corecheck-test-names
    type: yesoreyeram-infinity-datasource
    access: proxy
    editable: false
    jsonData:
      allowedHosts:
        - http://test-names:8080
"""


def write_if_changed(path: Path, content: str) -> bool:
    if path.exists() and path.read_text() == content:
        return False
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(content)
    return True


def leftover_expressions(rendered: str) -> list[str]:
    """${...} sequences that are not Grafana data links left by $${ escapes."""
    return [expr for expr in LEFTOVER.findall(rendered) if not expr.startswith("__")]


def render_all(workspace: str, region: str, account: str, out_dir: Path = GENERATED) -> list[Path]:
    written: list[Path] = []
    dashboards = out_dir / "dashboards"
    for key, dashboard in DASHBOARDS.items():
        source_path = TEMPLATES / dashboard["template"]
        rendered = render_template(source_path.read_text(), context(workspace, region, account, dashboard))
        leftovers = leftover_expressions(rendered)
        if leftovers:
            raise ValueError(f"{dashboard['template']} has unsubstituted expressions: {leftovers}")
        json.loads(rendered)
        dest = dashboards / f"{key}.json"
        if write_if_changed(dest, rendered):
            written.append(dest)
    datasource = out_dir / "provisioning/datasources/cloudwatch.yaml"
    if write_if_changed(datasource, datasource_yaml(region)):
        written.append(datasource)
    return written


def lookup_account(profile: str | None) -> str:
    cmd = ["aws", "sts", "get-caller-identity", "--query", "Account", "--output", "text"]
    if profile:
        cmd.extend(["--profile", profile])
    proc = subprocess.run(cmd, check=False, capture_output=True, text=True)
    if proc.returncode != 0:
        detail = proc.stderr.strip() or proc.stdout.strip()
        raise SystemExit(f"could not resolve AWS account id ({detail})")
    account = proc.stdout.strip()
    if not account.isdigit() or len(account) != 12:
        raise SystemExit(f"unexpected account id from sts: {account!r}")
    return account


def check() -> None:
    sample = render_template("pre $${__data.fields.PR} ${title}", {"title": "T"})
    if sample != "pre ${__data.fields.PR} T":
        raise SystemExit(f"escape render failed: {sample!r}")
    with tempfile.TemporaryDirectory() as tmp:
        out = Path(tmp)
        render_all("dev", "ap-south-1", "000000000000", out)
        github = json.loads((out / "dashboards/github.json").read_text())
        if github["uid"] != "corecheck-github-overview":
            raise SystemExit("github dashboard uid mismatch")
        if "/corecheck/github-events/dev" not in json.dumps(github):
            raise SystemExit("github log group was not substituted")
        if "Corecheck/dev" not in github["description"]:
            raise SystemExit("telemetry namespace was not substituted")
        if "${__data.fields.PR}" not in (out / "dashboards/github.json").read_text():
            raise SystemExit("grafana data-link escape was lost")
        jobs = json.loads((out / "dashboards/jobs.json").read_text())
        encoded = json.dumps(jobs)
        if "arn:aws:states:ap-south-1:000000000000:stateMachine:start-jobs-dev" not in encoded:
            raise SystemExit("jobs state machine arn was not substituted")
    print("rendered dashboards ok")


def parse_args(argv: list[str]) -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--workspace", default="dev", help="Terraform workspace, used in log group and queue names")
    parser.add_argument("--region", default="ap-south-1", help="Region the dashboards query")
    parser.add_argument("--account", help="AWS account id. Looked up with sts when omitted")
    parser.add_argument("--profile", help="AWS CLI profile for the account lookup")
    parser.add_argument("--check", action="store_true", help="Render with a fake account and validate the JSON")
    return parser.parse_args(argv)


def main(argv: list[str] | None = None) -> None:
    args = parse_args(sys.argv[1:] if argv is None else argv)
    if args.check:
        check()
        return
    account = args.account or lookup_account(args.profile)
    written = render_all(args.workspace, args.region, account)
    if written:
        for path in written:
            print(path.relative_to(HERE))
    else:
        print("dashboards already up to date")


if __name__ == "__main__":
    main()
