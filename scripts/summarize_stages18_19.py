"""Summarize sanitized bounded lab logs; nearest-rank percentiles, no SLA claims."""
import argparse,collections,csv,json,math,pathlib,re
p=argparse.ArgumentParser();p.add_argument('directory',type=pathlib.Path);a=p.parse_args()
phases=[];profiles=[]
log_text=(a.directory/'test.log').read_text()
for line in log_text.splitlines():
 for marker,dest in [('LOAD_PHASE_JSON ',phases),('GEOMETRY_PROFILE_JSON ',profiles)]:
  if marker in line:dest.append(json.loads(line.split(marker,1)[1]))
def quantile(x,p):return sorted(x)[max(0,math.ceil(len(x)*p)-1)] if x else None
summary=[];endpoints=[];windows=[];algorithms=[]
for phase in phases:
 samples=phase.pop('Samples') or [];groups=collections.defaultdict(list)
 for row in samples:
  path=re.sub(r'[0-9a-f]{8}-(?:[0-9a-f]{4}-){3}[0-9a-f]{12}',':uuid',row['Endpoint']);path=re.sub(r'/\d+(?=/|$)','/:id',path)
  groups[(path,row['Status'],row['Error'])].append(row['Milliseconds'])
 for (endpoint,status,error),latency in groups.items():endpoints.append({'phase':phase['Name'],'endpoint':endpoint,'status':status,'network_error':error,'n':len(latency),'rps':len(latency)/phase['Seconds'],**{k:quantile(latency,v) for k,v in [('p50',.5),('p90',.9),('p95',.95),('p99',.99),('max',1)]}})
 by_window=collections.defaultdict(list);by_algorithm=collections.defaultdict(list)
 start=phase.get('StartedUnixMS',0)
 for row in samples:
  if start and row.get('StartedUnixMS'):by_window[max(0,(row['StartedUnixMS']-start)//60000)].append(row)
  if row.get('Algorithm'):by_algorithm[row['Algorithm']].append(row)
 for window,items in sorted(by_window.items()):
  values=[r['Milliseconds'] for r in items];duration=min(60,max(.001,phase['Seconds']-window*60))
  windows.append({'phase':phase['Name'],'offset_seconds':window*60,'window_seconds':duration,'n':len(values),'rps':len(values)/duration,'network_errors':sum(r['Error'] for r in items),'unexpected_5xx':sum(r['Status']>=500 and r['Status']!=503 for r in items),'p99_small_sample':len(values)<1000,**{k:quantile(values,v) for k,v in [('p50',.5),('p90',.9),('p95',.95),('p99',.99),('max',1)]}})
 for algorithm,items in sorted(by_algorithm.items()):
  values=[r['Milliseconds'] for r in items]
  algorithms.append({'phase':phase['Name'],'algorithm':algorithm,'http_requests_including_replays':len(items),'statuses':dict(collections.Counter(str(r['Status']) for r in items)),**{k:quantile(values,v) for k,v in [('p50',.5),('p95',.95),('p99',.99),('max',1)]}})
 latency=[r['Milliseconds'] for r in samples]
 summary.append({'phase':phase['Name'],'test_passed':bool(re.search(r'--- PASS: TestStage18MixedLoad/'+re.escape(phase['Name'])+r'\s',log_text)),'clients':phase['Clients'],'apis':phase['APIs'],'seconds':phase['Seconds'],'requests':len(samples),'rps':len(samples)/phase['Seconds'],'statuses':dict(collections.Counter(str(r['Status']) for r in samples)),'network_errors':sum(r['Error'] for r in samples),**{k:quantile(latency,v) for k,v in [('p50',.5),('p90',.9),('p95',.95),('p99',.99),('max',1)]},'diagnostics':phase['Diagnostics']})
(a.directory/'summary.json').write_text(json.dumps({'phases':summary,'profiles':profiles,'latency_windows':windows,'http_by_algorithm':algorithms},indent=2))
if endpoints:
 with (a.directory/'endpoint-outcomes.csv').open('w') as out:
  writer=csv.DictWriter(out,fieldnames=list(endpoints[0]));writer.writeheader();writer.writerows(endpoints)
if windows:
 with (a.directory/'latency-windows.csv').open('w') as out:
  writer=csv.DictWriter(out,fieldnames=list(windows[0]));writer.writeheader();writer.writerows(windows)
if algorithms:(a.directory/'http-by-algorithm.json').write_text(json.dumps(algorithms,indent=2))
for row in summary:print(row['phase'],'n=',row['requests'],'rps=',round(row['rps'],2),'p95=',round(row['p95'],3) if row['p95'] is not None else None,'p99=',row['p99'],'statuses=',row['statuses'])
for row in profiles:print(row['Profile'],row['Algorithm'],row['Result']['Status'],row['Result']['Milliseconds'])
