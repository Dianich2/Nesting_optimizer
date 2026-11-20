#!/usr/bin/env python3
"""Validate three fresh growth logs; emit aggregate-only CSV (no payload or IDs)."""
import csv
import json
import sys
from pathlib import Path


def main():
    if len(sys.argv) != 5:
        raise SystemExit("usage: summarize_operational_growth.py LOG1 LOG2 LOG3 OUT.csv")
    rows = []
    for run, name in enumerate(sys.argv[1:4], 1):
        source = Path(name).read_text()
        if not source.rstrip().endswith("PASS") or "--- FAIL:" in source:
            raise SystemExit(f"failed/incomplete run: {name}")
        snapshots = [json.loads(line.split("GROWTH_ROW ", 1)[1])
                     for line in source.splitlines() if "GROWTH_ROW " in line]
        cleanup = [json.loads(line.split("CLEANUP_ROW ", 1)[1])
                   for line in source.splitlines() if "CLEANUP_ROW " in line]
        if [s["checkpoint_per_kind"] for s in snapshots] != [100, 1000, 5000, 5000]:
            raise SystemExit(f"missing checkpoints: {name}")
        if len(cleanup) != 1 or any(cleanup[0][k] != v for k, v in {
                "deleted_payloads": 5010, "actual_native_optimize_calls": 9500,
                "preview_tombstones": 7600, "remaining_running": 0}.items()):
            raise SystemExit(f"cleanup/count mismatch: {name}")
        for position, snapshot in enumerate(snapshots):
            n = snapshot["checkpoint_per_kind"]
            preview_failed = 2600 if position == 3 else n // 50
            expected = {"nesting_operations": {"completed": n},
                        "nesting_preview_operations": {"completed": n, "failed": preview_failed}}
            if snapshot["ready_count"] != (0 if position == 3 else 10):
                raise SystemExit(f"ready quota/count mismatch: {name}")
            common = {k: v for k, v in snapshot.items() if k != "operations"}
            for table, data in snapshot["operations"].items():
                counts = {s["State"]: s["Count"] for s in data["states"]}
                if counts != expected[table]:
                    raise SystemExit(f"terminal state mismatch: {name}/{table}")
                row = {"run": run, "phase": "cleaned" if position == 3 else "growth",
                       "table": table, **common, **{k: v for k, v in data.items() if k != "states"},
                       "completed": counts.get("completed", 0), "failed": counts.get("failed", 0)}
                rows.append(row)
        print(f"run {run}: 10000 usecase operations, cleanup 5010, tombstones 7600: PASS")
    with Path(sys.argv[4]).open("w", newline="") as out:
        writer = csv.DictWriter(out, fieldnames=list(rows[0]))
        writer.writeheader()
        writer.writerows(rows)
    print(f"wrote {len(rows)} aggregate table snapshots")


if __name__ == "__main__":
    main()
