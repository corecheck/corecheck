import json
import hashlib
from collections import defaultdict

MAX_SIGNED_32 = 2**31 - 1
WRAPAROUND_FALLBACK = 1337
WORKSPACE_PREFIX = "/tmp/bitcoin/"
BUILD_WORKSPACE_PREFIX = "/tmp/bitcoin/build/"
ALLOWED_EXTENSIONS = (".cpp", ".h", ".c")
EXCLUDED_PREFIXES = (
    "src/test",
    "src/qt/test",
    "src/wallet/test",
    "test",
    "src/bench",
)

def normalize_count(count):
    # LLVM coverage bug can wrap around and yield huge unsigned values; clamp to a stable fallback.
    if count > MAX_SIGNED_32:
        return WRAPAROUND_FALLBACK
    return count

def md5_stub(filename, line):
    # gcovr hashes source context; we can't reproduce it
    # so we make a stable placeholder
    return hashlib.md5(f"{filename}:{line}".encode()).hexdigest()

def normalize_filename(filename):
    if filename.startswith(BUILD_WORKSPACE_PREFIX):
        return filename[len(BUILD_WORKSPACE_PREFIX):]
    if filename.startswith(WORKSPACE_PREFIX):
        return filename[len(WORKSPACE_PREFIX):]
    if filename.startswith("build/src/"):
        return filename[len("build/"):]
    return filename


def should_include_filename(filename):
    if not filename.startswith("src/"):
        return False

    if filename.startswith(EXCLUDED_PREFIXES):
        return False

    return filename.endswith(ALLOWED_EXTENSIONS)

def lcov_to_gcovr_json(lcov_path):
    files = {}

    current_file = None

    with open(lcov_path, "r", encoding="utf-8") as f:
        for raw in f:
            line = raw.strip()

            if line.startswith("SF:"):
                current_file = normalize_filename(line[3:])
                if not should_include_filename(current_file):
                    current_file = None
                    continue

                files.setdefault(current_file, {
                    "file": current_file,
                    "lines": defaultdict(lambda: {
                        "branches": []
                    }),
                    "functions": {}
                })

            elif line.startswith("DA:") and current_file:
                lineno, count = line[3:].split(",")
                lineno = int(lineno)
                count = normalize_count(int(count))

                files[current_file]["lines"][lineno].update({
                    "line_number": lineno,
                    "function_name": None,
                    "count": count,
                    "gcovr/md5": md5_stub(current_file, lineno),
                })

            elif line.startswith("BRDA:") and current_file:
                parts = line[5:].split(",")
                lineno = int(parts[0])
                count = parts[3]

                if count == "-":
                    count = 0
                else:
                    count = normalize_count(int(count))

                branchno = len(files[current_file]["lines"][lineno]["branches"])

                files[current_file]["lines"][lineno]["branches"].append({
                    "branchno": branchno,
                    "count": count,
                    "fallthrough": False,
                    "throw": False,
                    "source_block_id": 0
                })

            elif line == "end_of_record":
                current_file = None

    # Final assembly
    out = {
        "gcovr/format_version": "0.14",
        "files": []
    }

    for f in files.values():
        out["files"].append({
            "file": f["file"],
            "lines": [
                {
                    **line,
                    "branches": line["branches"]
                }
                for _, line in sorted(f["lines"].items())
            ],
            "functions": []  # cannot be reconstructed from LCOV
        })

    return out


if __name__ == "__main__":
    import sys

    if len(sys.argv) != 3:
        print("usage: lcov_to_gcovr.py coverage.info output.json")
        sys.exit(1)

    data = lcov_to_gcovr_json(sys.argv[1])

    with open(sys.argv[2], "w", encoding="utf-8") as f:
        json.dump(data, f, indent=2)

    print("Wrote gcovr-style JSON")
