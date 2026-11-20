#!/usr/bin/env python3
"""Summarize six fresh, paired test logs; refuse incomplete/failed runs."""
import argparse
import csv
import json
import statistics
from collections import Counter, defaultdict
from pathlib import Path

parser = argparse.ArgumentParser()
parser.add_argument('--input', type=Path, required=True)
parser.add_argument('--output', type=Path, required=True)
args = parser.parse_args()
rows = []
for variant in ('before', 'after'):
    for repeat in range(1, 4):
        path = args.input / f'stage6-{variant}-{repeat}.log'
        text = path.read_text()
        if not text.rstrip().endswith('PASS'):
            raise SystemExit(f'Incomplete or failed experiment: {path.name}')
        parsed = [json.loads(line.split('GEOMETRY_ROW ', 1)[1])
                  for line in text.splitlines() if 'GEOMETRY_ROW ' in line]
        if len(parsed) != 47:
            raise SystemExit(f'Expected 47 fixtures in {path.name}, got {len(parsed)}')
        for row in parsed:
            if row['placed'] + row['unplaced'] != row['quantity']:
                raise SystemExit(f'Count inconsistency in {path.name}')
            if row['status'] == 'completed' and not row['valid']:
                raise SystemExit(f'Invalid completed layout in {path.name}')
            if row['status'] == 'timeout' and row['placed'] != 0:
                raise SystemExit(f'Partial timeout result in {path.name}')
        rows.extend(dict(row, variant=variant, repeat=repeat) for row in parsed)
args.output.mkdir(parents=True, exist_ok=True)
with (args.output / 'geometry-raw.csv').open('w') as target:
    writer = csv.DictWriter(target, fieldnames=sorted(rows[0]))
    writer.writeheader()
    writer.writerows(rows)
key_fields = ('fixture', 'algorithm', 'quantity', 'vertices', 'holes', 'angles')
groups = defaultdict(lambda: defaultdict(list))
for row in rows:
    groups[tuple(row[field] for field in key_fields)][row['variant']].append(row)
summary = ['| Fixture / algorithm / q / v / h / angles | Before wall ms median [min,max] | After wall ms median [min,max] | After CPU ms median | Alloc MiB median per attempt | Sampled goroutines max | Process cumulative RSS KiB max | Placed | Outcomes |',
           '|---|---:|---:|---:|---:|---:|---:|---|---|']
for key, variants in sorted(groups.items()):
    if any(len(variants[label]) != 3 for label in ('before', 'after')):
        raise SystemExit(f'Unpaired fixture: {key}')
    before, after = variants['before'], variants['after']
    def wall(values):
        sample = [row['wall_ms'] for row in values]
        return f'{statistics.median(sample):.3f} [{min(sample):.3f},{max(sample):.3f}]'
    summary.append('| ' + ' / '.join(map(str, key)) + ' | ' + wall(before) + ' | ' + wall(after) +
                   f" | {statistics.median(row['cpu_ms'] for row in after):.3f}" +
                   f" | {statistics.median(row['bytes'] for row in after)/2**20:.2f}" +
                   f" | {max(row['peak_goroutines_sampled'] for row in after)}" +
                   f" | {max(row['maxrss_kib_process_cumulative'] for row in after)}" +
                   ' | ' + ','.join(map(str, sorted(set(row['placed'] for row in after)))) +
                   ' | ' + ', '.join(f'{name}:{count}' for name, count in Counter(row['status'] for row in after).items()) + ' |')
(args.output / 'geometry-summary.md').write_text('\n'.join(summary) + '\n')
print(f'Validated {len(rows)} rows, {len(groups)} paired fixtures, six fresh PASS processes')
