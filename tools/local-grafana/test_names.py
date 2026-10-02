#!/usr/bin/env python3
"""List test names for the Grafana dropdowns, biggest slowdown first.

Grafana's CloudWatch datasource cannot run a Logs Insights query to fill a
template variable. This serves that list in the same order as the Biggest
slowdowns tables: largest percent increase in median duration from the first
half of the selected range to the second half. A test is ranked only when
both halves have at least 3 runs and the earlier median is at least 10
seconds. Every other test is still listed, after the ranked ones.
"""

from __future__ import annotations

import argparse
import json
import os
import subprocess
import threading
import time
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer

LOG_GROUP = os.environ.get("TEST_RESULTS_LOG_GROUP", "/corecheck/test-results/dev")
REGION = os.environ.get("AWS_REGION", os.environ.get("AWS_DEFAULT_REGION", "ap-south-1"))
PORT = int(os.environ.get("PORT", "8080"))
CACHE_SECONDS = 120
QUERY_TIMEOUT_SECONDS = 55

QUERY = """
fields test_name, duration_s,
  if(toMillis(@timestamp) < ({start_ms} + {end_ms}) / 2, duration_s, null) as early_d,
  if(toMillis(@timestamp) >= ({start_ms} + {end_ms}) / 2, duration_s, null) as late_d
| filter test_type = "{test_type}" and duration_s > 0
| stats pct(early_d, 50) as baseline_s, pct(late_d, 50) as current_s, count(early_d) as n_early, count(late_d) as n_late by test_name
| limit 10000
"""

_cache: dict[tuple[str, int, int], tuple[float, list[str]]] = {}
_cache_lock = threading.Lock()


def _aws_json(args: list[str]) -> dict:
    proc = subprocess.run(
        ["aws", *args, "--region", REGION, "--output", "json"],
        check=False,
        capture_output=True,
        text=True,
        timeout=QUERY_TIMEOUT_SECONDS,
    )
    if proc.returncode != 0:
        detail = (proc.stderr or proc.stdout or "aws logs failed").strip().splitlines()
        raise RuntimeError(detail[-1][:300] if detail else "aws logs failed")
    return json.loads(proc.stdout)


def _rows(test_type: str, start_ms: int, end_ms: int) -> list[dict[str, str]]:
    started = _aws_json(
        [
            "logs",
            "start-query",
            "--log-group-name",
            LOG_GROUP,
            "--start-time",
            str(start_ms // 1000),
            "--end-time",
            str(end_ms // 1000 + 1),
            "--query-string",
            QUERY.format(test_type=test_type, start_ms=start_ms, end_ms=end_ms),
        ]
    )
    query_id = started["queryId"]
    deadline = time.monotonic() + QUERY_TIMEOUT_SECONDS
    while True:
        payload = _aws_json(["logs", "get-query-results", "--query-id", query_id])
        status = payload.get("status")
        if status == "Complete":
            rows: list[dict[str, str]] = []
            for raw in payload.get("results") or []:
                rows.append({cell["field"]: cell.get("value", "") for cell in raw})
            return rows
        if status in {"Failed", "Cancelled", "Timeout"}:
            raise RuntimeError(f"logs query {status}")
        if time.monotonic() > deadline:
            raise RuntimeError("logs query timed out")
        time.sleep(0.5)


def _rank(row: dict[str, str]) -> tuple[int, float, str]:
    name = row.get("test_name", "")
    try:
        n_early = float(row["n_early"])
        n_late = float(row["n_late"])
        baseline = float(row["baseline_s"])
        current = float(row["current_s"])
    except (KeyError, TypeError, ValueError):
        return (2, 0.0, name)
    if n_early >= 3 and n_late >= 3 and baseline >= 10:
        change = (current - baseline) / baseline
        if change > 0:
            return (0, -change, name)
        return (1, change, name)
    return (2, 0.0, name)


def test_names(test_type: str, start_ms: int, end_ms: int) -> list[str]:
    key = (test_type, start_ms, end_ms)
    now = time.monotonic()
    with _cache_lock:
        cached = _cache.get(key)
        if cached and now - cached[0] < CACHE_SECONDS:
            return cached[1]
    names = [row.get("test_name", "") for row in sorted(_rows(test_type, start_ms, end_ms), key=_rank)]
    names = [name for name in names if name]
    with _cache_lock:
        _cache[key] = (time.monotonic(), names)
    return names


def _parse_window(query: dict[str, list[str]]) -> tuple[str, int, int]:
    test_type = (query.get("test_type") or [""])[0]
    if test_type not in {"functional", "unit"}:
        raise ValueError("test_type must be functional or unit")
    try:
        start_ms = int((query.get("from") or [""])[0])
        end_ms = int((query.get("to") or [""])[0])
    except ValueError as exc:
        raise ValueError("from and to must be epoch milliseconds") from exc
    if end_ms <= start_ms:
        raise ValueError("to must be after from")
    if end_ms - start_ms > 400 * 24 * 60 * 60 * 1000:
        raise ValueError("time range is longer than 400 days")
    return test_type, start_ms, end_ms


class Handler(BaseHTTPRequestHandler):
    def do_GET(self) -> None:  # noqa: N802
        path, _, raw_query = self.path.partition("?")
        if path == "/health":
            self._send(200, {"ok": True})
            return
        if path != "/names":
            self._send(404, {"error": "not found"})
            return
        from urllib.parse import parse_qs

        try:
            test_type, start_ms, end_ms = _parse_window(parse_qs(raw_query))
            names = test_names(test_type, start_ms, end_ms)
        except ValueError as exc:
            self._send(400, {"error": str(exc)})
            return
        except Exception as exc:
            print(f"test name query failed: {exc}", flush=True)
            self._send(502, {"error": "test name query failed"})
            return
        self._send(200, [{"name": name} for name in names])

    def log_message(self, fmt: str, *args: object) -> None:
        print("%s - %s" % (self.address_string(), fmt % args), flush=True)

    def _send(self, status: int, payload: object) -> None:
        body = json.dumps(payload).encode()
        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)


def serve() -> None:
    server = ThreadingHTTPServer(("0.0.0.0", PORT), Handler)
    print(f"test names listening on {PORT} for {LOG_GROUP} in {REGION}", flush=True)
    server.serve_forever()


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--once", action="store_true", help="Print the first names for a range, then exit")
    parser.add_argument("--test-type", default="functional", choices=("functional", "unit"))
    parser.add_argument("--days", type=int, default=90)
    args = parser.parse_args()
    if not args.once:
        serve()
        return
    end_ms = int(time.time() * 1000)
    start_ms = end_ms - args.days * 24 * 60 * 60 * 1000
    names = test_names(args.test_type, start_ms, end_ms)
    print(f"{len(names)} {args.test_type} tests")
    for name in names[:8]:
        print(name)


if __name__ == "__main__":
    main()
