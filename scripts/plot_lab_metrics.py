"""Standalone SVG plots from sanitized lab JSON. No extra plotting dependencies.
CPU and memory use docker stats, not a claim of physical production capacity.
"""
import argparse,html,json,math,pathlib,re,xml.etree.ElementTree as ET
p=argparse.ArgumentParser();p.add_argument('directory',type=pathlib.Path);a=p.parse_args()
colors=['#2563eb','#d97706','#059669'];width,height=920,330;left,right,top,bottom=72,24,58,65

def chart(filename,title,ylabel,series,note='',ymax=None):
 all_points=[v for _,values in series for v in values]
 if not all_points:return
 xmax=max(x for x,y in all_points) or 1;ymax=ymax or max(y for x,y in all_points)*1.12 or 1
 def point(x,y):return left+x/xmax*(width-left-right),height-bottom-y/ymax*(height-top-bottom)
 svg=[f'<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 {width} {height}" role="img" aria-label="{html.escape(title)}">', '<rect width="100%" height="100%" fill="#fff"/>',f'<text x="{left}" y="22" font-family="sans-serif" font-size="16" fill="#111827">{html.escape(title)}</text>',f'<text x="{left}" y="42" font-family="sans-serif" font-size="11" fill="#475569">{html.escape(note)}</text>']
 for i in range(6):
  y=ymax*i/5;_,py=point(0,y);svg.extend([f'<line x1="{left}" y1="{py:.2f}" x2="{width-right}" y2="{py:.2f}" stroke="#e2e8f0"/>',f'<text x="{left-8}" y="{py+4:.2f}" text-anchor="end" font-family="sans-serif" font-size="11">{y:.2g}</text>'])
 for i in range(7):
  x=xmax*i/6;px,_=point(x,0);svg.append(f'<text x="{px:.2f}" y="{height-bottom+20}" text-anchor="middle" font-family="sans-serif" font-size="11">{x/60:.1f}</text>')
 for i,(name,values) in enumerate(series):
  points=' '.join(f'{x:.2f},{y:.2f}' for x,y in [point(x,y) for x,y in values]);svg.append(f'<polyline fill="none" stroke="{colors[i%len(colors)]}" stroke-width="1.8" points="{points}"/>');svg.append(f'<text x="{left+i*230}" y="{height-12}" font-family="sans-serif" font-size="12" fill="{colors[i%len(colors)]}">{html.escape(name)}</text>')
 svg.extend([f'<text x="{width/2}" y="{height-36}" text-anchor="middle" font-family="sans-serif" font-size="12">Elapsed minutes</text>',f'<text transform="translate(17 {height/2}) rotate(-90)" text-anchor="middle" font-family="sans-serif" font-size="12">{html.escape(ylabel)}</text>','</svg>']);output='\n'.join(svg);ET.fromstring(output);(a.directory/filename).write_text(output)

resources=[]
for line in (a.directory/'resources.jsonl').read_text().splitlines():
 try:resources.append(json.loads(line))
 except json.JSONDecodeError:pass
if resources:
 origin=resources[0]['time'];cpu=[];memory=[];wal=[];wal_origin=None
 for r in resources:
  elapsed=r['time']-origin;raw=r['cgroup'];cpu.append((elapsed,float(raw['CPUPerc'].rstrip('%'))))
  m=re.match(r'([\d.]+)([A-Za-z]+)',raw['MemUsage']);units={'B':1,'kB':1000,'KiB':1024,'MB':1e6,'MiB':1024**2,'GB':1e9,'GiB':1024**3}
  if m:memory.append((elapsed,float(m[1])*units[m[2]]/1024**2))
  for line in r['sql'].splitlines():
   if line.startswith('{'):
    v=json.loads(line);value=float(v['wal_bytes']);wal_origin=value if wal_origin is None else wal_origin;wal.append((elapsed,(value-wal_origin)/1024**2))
 chart('cpu.svg','Owned load cgroup CPU','% of one logical CPU',[('CPU usage',cpu)],'Hard cap: 2 CPUs = 200%; API and runner combined',200)
 chart('memory.svg','Owned load cgroup working set','MiB',[('Docker working set',memory)],'Hard cap: 1536 MiB; axis scaled to measured values')
 chart('wal.svg','Owned test database WAL growth','MiB delta',[('WAL growth',wal)],'This lab only; safety bound 512 MiB delta')
summary=a.directory/'summary.json'
if summary.exists():
 d=json.loads(summary.read_text());diagnostics=[(i%2,row) for p in d['phases'] for i,row in enumerate(p['diagnostics']) if row.get('SampleUnixMS')]
 if diagnostics:
  first=min(row['SampleUnixMS'] for _,row in diagnostics)
  for field,filename,title,unit,factor in [('HeapAlloc','heap.svg','Live Go heap per API','MiB',1/1024**2),('Goroutines','goroutines.svg','Goroutines per API','count',1),('FDs','fds.svg','File descriptors per API','count',1)]:
   series=[(f'API {api+1}',[((r['SampleUnixMS']-first)/1000,r[field]*factor) for n,r in diagnostics if n==api]) for api in [0,1]];chart(filename,title,unit,series,'Test-only diagnostics; full active interval followed by idle')
 windows=d.get('latency_windows',[])
 if windows:
  chart('latency-trend.svg','Observed HTTP latency by 60-second window','milliseconds',[(q,[(r['offset_seconds'],r[q]) for r in windows if r['phase']=='soak' and r.get('window_seconds',60)>=30]) for q in ['p95','p99']],'Complete >=30s windows; see CSV for n/partial tail. Descriptive, no SLA')
print('Validated standalone SVG plots:',', '.join(p.name for p in a.directory.glob('*.svg')))
