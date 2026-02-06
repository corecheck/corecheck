import json
import hashlib
from collections import defaultdict

def md5_stub(filename, line):
    # gcovr hashes source context; we can't reproduce it
    # so we make a stable placeholder
    return hashlib.md5(f"{filename}:{line}".encode()).hexdigest()

def lcov_to_gcovr_json(lcov_path):
    files = {}

    current_file = None

    with open(lcov_path, "r", encoding="utf-8") as f:
        for raw in f:
            line = raw.strip()

            if line.startswith("SF:"):
                current_file = line[3:]
                files[current_file] = {
                    "file": current_file,
                    "lines": defaultdict(lambda: {
                        "branches": []
                    }),
                    "functions": {}
                }

            elif line.startswith("DA:") and current_file:
                lineno, count = line[3:].split(",")
                lineno = int(lineno)

                files[current_file]["lines"][lineno].update({
                    "line_number": lineno,
                    "function_name": None,
                    "count": int(count),
                    "gcovr/md5": md5_stub(current_file, lineno),
                })

            elif line.startswith("BRDA:") and current_file:
                parts = line[5:].split(",")
                lineno = int(parts[0])
                count = parts[3]

                if count == "-":
                    count = 0
                else:
                    count = int(count)

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
