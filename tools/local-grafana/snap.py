#!/usr/bin/env python3
"""Save PNG screenshots of the local Grafana panels.

Grafana renders each panel through the image-renderer service started by
docker compose. The files land in generated/snaps/, which is gitignored.
"""

from __future__ import annotations

import argparse
import json
import re
import sys
import urllib.error
import urllib.parse
import urllib.request
from pathlib import Path

HERE = Path(__file__).resolve().parent
SNAPS = HERE / "generated" / "snaps"
BASE = "http://localhost:3000"

DASHBOARDS = {
    "github": "corecheck-github-overview",
    "tests": "corecheck-tests",
    "benchmarks": "corecheck-benchmarks",
    "jobs": "corecheck-jobs",
}


def slug(text: str) -> str:
    cleaned = re.sub(r"[^a-z0-9]+", "-", text.lower()).strip("-")
    return cleaned[:60] or "panel"


def dashboard_json(uid: str) -> dict:
    with urllib.request.urlopen(f"{BASE}/api/dashboards/uid/{uid}", timeout=30) as response:
        return json.load(response)["dashboard"]


def panel_list(dashboard: dict, only: set[int] | None) -> list[dict]:
    panels = []
    for panel in dashboard.get("panels", []):
        panel_id = panel.get("id")
        if panel.get("type") == "row" or panel_id is None:
            continue
        if only and panel_id not in only:
            continue
        panels.append(panel)
    return panels


def render_panel(uid: str, panel_id: int, width: int, height: int, time_from: str, time_to: str) -> bytes:
    query = urllib.parse.urlencode(
        {
            "orgId": 1,
            "panelId": panel_id,
            "width": width,
            "height": height,
            "from": time_from,
            "to": time_to,
            "theme": "dark",
            "timeout": 90,
        }
    )
    url = f"{BASE}/render/d-solo/{uid}/{uid}?{query}"
    request = urllib.request.Request(url)
    try:
        with urllib.request.urlopen(request, timeout=120) as response:
            body = response.read()
            content_type = response.headers.get("Content-Type", "")
    except urllib.error.HTTPError as exc:
        detail = exc.read().decode("utf-8", "replace")[:500]
        raise SystemExit(f"render failed for panel {panel_id}: HTTP {exc.code}\n{detail}") from exc
    if "image" not in content_type:
        raise SystemExit(f"render for panel {panel_id} returned {content_type or 'no content type'}")
    return body


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("dashboard", choices=sorted(DASHBOARDS))
    parser.add_argument("--panel", type=int, action="append", dest="panels", help="Panel id to render. Repeatable. Default is every panel.")
    parser.add_argument("--width", type=int, default=1400)
    parser.add_argument("--height", type=int, default=800)
    parser.add_argument("--from", dest="time_from", default="now-90d")
    parser.add_argument("--to", dest="time_to", default="now")
    args = parser.parse_args()

    uid = DASHBOARDS[args.dashboard]
    only = set(args.panels) if args.panels else None
    panels = panel_list(dashboard_json(uid), only)
    if not panels:
        raise SystemExit("no matching panels")

    out_dir = SNAPS / args.dashboard
    out_dir.mkdir(parents=True, exist_ok=True)
    for panel in panels:
        dest = out_dir / f"{panel['id']}-{slug(panel.get('title') or 'panel')}.png"
        dest.write_bytes(render_panel(uid, panel["id"], args.width, args.height, args.time_from, args.time_to))
        print(dest.relative_to(HERE))


if __name__ == "__main__":
    try:
        main()
    except urllib.error.HTTPError as exc:
        detail = exc.read().decode("utf-8", "replace")[:500]
        raise SystemExit(f"Grafana returned HTTP {exc.code} for {exc.url}\n{detail}") from exc
    except urllib.error.URLError as exc:
        raise SystemExit(f"local Grafana is not reachable at {BASE} ({exc.reason})") from exc
    except KeyboardInterrupt:
        sys.exit(1)
