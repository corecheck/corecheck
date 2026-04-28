#!/usr/bin/env python3

import argparse
import json
from collections import defaultdict
from pathlib import Path


def load_line_counts(path: Path) -> dict[tuple[str, int], int]:
    with path.open() as f:
        data = json.load(f)

    counts: dict[tuple[str, int], int] = {}

    for file_entry in data.get("files", []):
        file_path = file_entry.get("file")
        if not file_path:
            continue

        for line in file_entry.get("lines", []):
            line_number = line.get("line_number")
            if line_number is None:
                continue

            counts[(file_path, int(line_number))] = int(line.get("count", 0) or 0)

    return counts


def coverage_state(count: int | None) -> str:
    if count is None:
        return "missing"
    if count == 0:
        return "zero"
    return "covered"


def format_count(count: int | None) -> str:
    if count is None:
        return "-"
    return str(count)


def build_diffs(
    base_counts: dict[tuple[str, int], int],
    other_counts: dict[tuple[str, int], int],
) -> list[tuple[str, int, int | None, int | None, str, str]]:
    diffs = []

    for key in sorted(set(base_counts) | set(other_counts)):
        base = base_counts.get(key)
        other = other_counts.get(key)
        base_state = coverage_state(base)
        other_state = coverage_state(other)
        if base_state != other_state:
            file_path, line_number = key
            diffs.append((file_path, line_number, base, other, base_state, other_state))

    return diffs


def main() -> None:
    parser = argparse.ArgumentParser(
        description=(
            "Diff per-line coverage state between two gcovr coverage JSON files. "
            "Reports missing/zero/covered transitions and ignores covered-to-covered count changes."
        )
    )
    parser.add_argument("base", type=Path, help="Baseline JSON file")
    parser.add_argument("other", type=Path, help="Comparison JSON file")
    parser.add_argument(
        "--file",
        dest="file_filters",
        action="append",
        default=[],
        help="Only include files whose path contains this substring (repeatable)",
    )
    parser.add_argument(
        "--sort",
        choices=["transition", "path"],
        default="transition",
        help="Sort detailed output by transition type or by file/line",
    )
    parser.add_argument(
        "--limit",
        type=int,
        default=200,
        help="Max changed lines to print in the detailed section",
    )
    parser.add_argument(
        "--all",
        action="store_true",
        help="Print all changed lines",
    )
    args = parser.parse_args()

    base_counts = load_line_counts(args.base)
    other_counts = load_line_counts(args.other)
    diffs = build_diffs(base_counts, other_counts)

    if args.file_filters:
        diffs = [
            row for row in diffs if any(file_filter in row[0] for file_filter in args.file_filters)
        ]

    if args.sort == "transition":
        diffs.sort(key=lambda row: (row[4], row[5], row[0], row[1]))
    else:
        diffs.sort(key=lambda row: (row[0], row[1]))

    per_file: dict[str, int] = defaultdict(int)
    per_transition: dict[tuple[str, str], int] = defaultdict(int)
    for file_path, _, _, _, base_state, other_state in diffs:
        per_file[file_path] += 1
        per_transition[(base_state, other_state)] += 1

    print(f"Changed lines: {len(diffs)}")
    print(f"Changed files: {len(per_file)}")
    print()

    if per_transition:
        print("Transitions:")
        for (base_state, other_state), count in sorted(
            per_transition.items(), key=lambda item: (-item[1], item[0][0], item[0][1])
        ):
            print(f"{count:6d}  {base_state:>7} -> {other_state}")
        print()

    if per_file:
        print("Top files by changed lines:")
        for file_path, count in sorted(per_file.items(), key=lambda x: (-x[1], x[0]))[:20]:
            print(f"{count:6d}  {file_path}")
        print()

    if not diffs:
        print("No coverage state differences found.")
        return

    rows = diffs if args.all else diffs[: args.limit]

    print("Detailed line diffs:")
    print("BASE      OTHER     BASE_COUNT  OTHER_COUNT  FILE:LINE")
    for file_path, line_number, base, other, base_state, other_state in rows:
        print(
            f"{base_state:>7} -> {other_state:<7}  "
            f"{format_count(base):>10}  {format_count(other):>11}  {file_path}:{line_number}"
        )

    if not args.all and len(diffs) > len(rows):
        print()
        print(f"... truncated to {len(rows)} rows; use --all or raise --limit")


if __name__ == "__main__":
    main()
