#!/usr/bin/env python3
"""Validate paired stage9 experiments and emit aggregate CSV/median ranges."""
import argparse
import csv
import json
import statistics
from pathlib import Path

parser = argparse.ArgumentParser()
parser.add_argument('--input', type=Path, required=True)
parser.add_argument('--output', type=Path, required=True)
args = parser.parse_args()
rows = []
for repeat in range(1, 4):
    for variant in ('before', 'after'):
        path = args.input / f'stage9-{variant}-{repeat}.log'
        source = path.read_text()
        if not source.rstrip().endswith('PASS') or '--- FAIL:' in source:
            raise SystemExit(f'Failed/incomplete experiment: {path.name}')
        parsed = [json.loads(line.split('GEOMETRY_ROW ', 1)[1])
                  for line in source.splitlines() if 'GEOMETRY_ROW ' in line]
        if len(parsed) != 10 or len({(r['fixture'], r['algorithm']) for r in parsed}) != 10:
            raise SystemExit(f'Expected 10 distinct fixtures: {path.name}')
        for r in parsed:
            if r['status'] != 'completed' or not r['valid'] or r['placed'] != r['quantity'] or r['unplaced']:
                raise SystemExit(f'Quality failure: {path.name}/{r["fixture"]}')
            rows.append(dict(r, variant=variant, repeat=repeat))
keys = sorted({(r['fixture'], r['algorithm']) for r in rows})
lines = ['|Fixture/algorithm|Before wall ms median [min,max]|After wall ms median [min,max]|Before/after allocated MiB median|Before/after allocations median|Placed|',
         '|---|---:|---:|---:|---:|---:|']
for fixture, algorithm in keys:
    selected = [r for r in rows if r['fixture'] == fixture and r['algorithm'] == algorithm]
    groups = {v: [r for r in selected if r['variant'] == v] for v in ('before', 'after')}
    if any(len(values) != 3 for values in groups.values()):
        raise SystemExit(f'Unpaired case: {fixture}/{algorithm}')
    # Stable deterministic algorithms must preserve layout metrics and NFP work.
    for field in ('placed', 'unplaced', 'utilization', 'nfp_calls', 'cache_hits', 'cache_misses'):
        if len({r[field] for r in selected}) != 1:
            raise SystemExit(f'Changed quality/work: {fixture}/{algorithm}/{field}')
    def wall(values):
        xs = [r['wall_ms'] for r in values]
        return f'{statistics.median(xs):.3f} [{min(xs):.3f},{max(xs):.3f}]'
    def med(field, divisor=1):
        return '/'.join(f'{statistics.median(r[field] for r in groups[v])/divisor:.2f}' for v in ('before', 'after'))
    lines.append(f'|{fixture}/{algorithm}|{wall(groups["before"])}|{wall(groups["after"])}|{med("bytes", 2**20)}|{med("allocs")}|{selected[0]["placed"]}|')
args.output.mkdir(parents=True, exist_ok=True)
with (args.output / 'targeted-performance.csv').open('w', newline='') as out:
    writer = csv.DictWriter(out, fieldnames=sorted(rows[0]))
    writer.writeheader()
    writer.writerows(rows)
(args.output / 'targeted-performance.md').write_text('\n'.join(lines) + '\n')
print(f'PASS: {len(rows)} rows, {len(keys)} paired cases, unchanged quality/NFP work')
