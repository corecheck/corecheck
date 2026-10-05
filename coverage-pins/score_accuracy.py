#!/usr/bin/env python3
"""Score corecheck coverage-report accuracy.

A highlighted line is accurate when its coverage type is a real diff outcome
(new code, deleted code, or a file moving in or out of the build). A
highlighted line is a miss when it is lost or gained baseline coverage: a
zero-versus-nonzero flip on a line the pull request did not have to edit.

Accuracy for one report is accurate highlighted lines / all highlighted lines.
A report with no highlighted lines scores 100. Re-run this against the same
cohort (the most recently updated pulls) after new coverage jobs finish.
"""

import argparse
import json
import statistics
import sys
import urllib.error
import urllib.request
from collections import Counter
from concurrent.futures import ThreadPoolExecutor, as_completed
from datetime import datetime, timezone

API = "https://api.corecheck.dev"
FALSE_TYPES = {"lost_baseline_coverage", "gained_baseline_coverage"}


def get_json(url):
    req = urllib.request.Request(url, headers={"User-Agent": "corecheck-accuracy-score"})
    with urllib.request.urlopen(req, timeout=60) as resp:
        return json.load(resp)


def list_pulls(pages):
    pulls = []
    for page in range(1, pages + 1):
        batch = get_json(f"{API}/pulls?page={page}")
        if not batch:
            break
        pulls.extend(batch)
    return pulls


def fetch_latest_report(number):
    url = f"{API}/pulls/{number}/report"
    try:
        report = get_json(url)
    except urllib.error.HTTPError as err:
        if err.code == 404:
            return number, None
        raise
    if report.get("status") != "success" or not report.get("coverage"):
        return number, None
    return number, report


def score_report(report):
    true = 0
    false = 0
    by_type = Counter()
    for coverage_type, files in (report.get("coverage") or {}).items():
        for hunks in files.values():
            for hunk in hunks:
                for line in hunk.get("lines") or []:
                    if not line.get("highlight"):
                        continue
                    by_type[coverage_type] += 1
                    if coverage_type in FALSE_TYPES:
                        false += 1
                    else:
                        true += 1
    total = true + false
    accuracy = 100.0 if total == 0 else 100.0 * true / total
    return {
        "pr": report.get("pr_number"),
        "report_id": report.get("id"),
        "generated_at": report.get("generated_at"),
        "true_lines": true,
        "false_lines": false,
        "accuracy": accuracy,
        "by_type": dict(by_type),
    }


def summarize(rows):
    accuracies = [row["accuracy"] for row in rows]
    false_lines = [row["false_lines"] for row in rows]
    true_lines = sum(row["true_lines"] for row in rows)
    false_total = sum(false_lines)
    highlighted = true_lines + false_total
    by_type = Counter()
    for row in rows:
        by_type.update(row["by_type"])
    generated = [row["generated_at"] for row in rows if row["generated_at"]]
    return {
        "reports": len(rows),
        "median_accuracy_pct": round(statistics.median(accuracies), 1) if accuracies else None,
        "macro_accuracy_pct": round(statistics.fmean(accuracies), 1) if accuracies else None,
        "micro_accuracy_pct": round(100.0 * true_lines / highlighted, 1) if highlighted else 100.0,
        "clean_reports_pct": round(100.0 * sum(1 for n in false_lines if n == 0) / len(rows), 1) if rows else None,
        "median_false_lines": statistics.median(false_lines) if false_lines else 0,
        "mean_false_lines": round(statistics.fmean(false_lines), 1) if false_lines else 0,
        "true_highlighted_lines": true_lines,
        "false_highlighted_lines": false_total,
        "highlighted_lines_by_type": dict(by_type),
        "oldest_generated_at": min(generated) if generated else None,
        "newest_generated_at": max(generated) if generated else None,
    }


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--pages", type=int, default=1, help="pages of recently updated pulls (100 per page)")
    parser.add_argument("--since", help="only score reports generated at or after this UTC timestamp (YYYY-MM-DD or full ISO)")
    parser.add_argument("--workers", type=int, default=8)
    parser.add_argument("--out", help="write the full JSON result to this path")
    args = parser.parse_args()

    pulls = list_pulls(args.pages)
    rows = []
    missing = 0
    older_than_since = 0
    with ThreadPoolExecutor(max_workers=args.workers) as pool:
        futures = [pool.submit(fetch_latest_report, pull["number"]) for pull in pulls]
        for future in as_completed(futures):
            number, report = future.result()
            if report is None:
                missing += 1
                continue
            row = score_report(report)
            if args.since and (row["generated_at"] or "") < args.since:
                older_than_since += 1
                continue
            rows.append(row)

    rows.sort(key=lambda row: row["pr"] or 0)
    summary = summarize(rows)
    cohort = f"latest successful report for each of the {len(pulls)} most recently updated pulls that has one"
    if args.since:
        cohort += f", generated at or after {args.since}"
    result = {
        "scored_at": datetime.now(timezone.utc).isoformat(),
        "cohort": cohort,
        "pulls_considered": len(pulls),
        "pulls_without_successful_report": missing,
        "reports_older_than_since": older_than_since,
        "definition": (
            "Accuracy is the share of highlighted lines that are not lost_baseline_coverage "
            "or gained_baseline_coverage. Median accuracy is the typical report. "
            "Macro accuracy is the mean of per-report accuracies "
            "(a report with no highlighted lines scores 100). Micro accuracy pools every highlighted line. "
            "Clean reports have zero lost or gained baseline highlights."
        ),
        "summary": summary,
        "reports": rows,
    }
    text = json.dumps(result, indent=2)
    if args.out:
        with open(args.out, "w", encoding="utf-8") as fh:
            fh.write(text)
            fh.write("\n")
    print(json.dumps({k: result[k] for k in ("scored_at", "cohort", "pulls_considered", "pulls_without_successful_report", "summary")}, indent=2))
    return 0


if __name__ == "__main__":
    sys.exit(main())
