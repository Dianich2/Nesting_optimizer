"""Opt-in bounded real-HTTP load. Requires mode0600 metadata from the OWN lab.
Build: cd backend && CGO_ENABLED=0 go test -c -o /tmp/nesting17bin/load.test ./internal/integration
No arbitrary URL, no .env, no secrets printed; only recorded container is stopped.
"""
import argparse,json,os,pathlib,subprocess,time,uuid
p=argparse.ArgumentParser();p.add_argument('--metadata',required=True);p.add_argument('--binary',required=True);p.add_argument('--test',choices=['TestStage18MixedLoad','TestStage18GeometryProfiles','TestStage18RetainedLoadAudit','TestStage18IndependentOwners','TestStage18HeterogeneousOccupied','TestRealProcessLifecycle','TestStage19PGRecovery','TestStage19PendingIdentityChanges','TestStage19StatusTransportFaults'],required=True);p.add_argument('--output',required=True);p.add_argument('--soak',action='store_true');p.add_argument('--profile',action='store_true');p.add_argument('--compare',action='store_true');p.add_argument('--smoke',action='store_true');p.add_argument('--resume-high',type=pathlib.Path);p.add_argument('--audit-schema');a=p.parse_args()
private=pathlib.Path(a.metadata);assert private.stat().st_mode&0o077==0
m=json.loads(private.read_text());assert m['name'].startswith('nesting-1719-') and m['db_container'].startswith(m['name'])
def docker(*args,**kw):return subprocess.check_output(['docker',*args],text=True,**kw).strip()
x=json.loads(docker('inspect',m['container']))[0]
assert x['Id']==m['container'] and x['NetworkSettings']['Networks'][m['network_name']]['NetworkID']==m['network']
assert any(v.get('Name')==m['volume'] for v in x['Mounts'])
assert all(v['HostIp']=='127.0.0.1' for v in x['NetworkSettings']['Ports']['5432/tcp'])
gate=docker('exec',m['container'],'psql','-X','-v','ON_ERROR_STOP=1','-U',m['user'],'-d','nesting_test','-Atc','BEGIN READ ONLY; SELECT current_database(),current_user,postgis_version(); ROLLBACK;')
assert 'nesting_test|'+m['user']+'|3.4' in gate
root=pathlib.Path(__file__).resolve().parents[1];out=pathlib.Path(a.output);out.mkdir(parents=True,exist_ok=True)
env=os.environ.copy();env.update(NESTING_TEST_DSN=f"postgres://{m['user']}:{m['password']}@{m['db_container']}:5432/nesting_test?sslmode=disable&connect_timeout=5",NESTING_EXPECTED_TEST_USER=m['user'],NESTING_STAGE18='1',NESTING_TEST_PROCESS_BINARY='/tests/load.test',GOMAXPROCS='2',GOMEMLIMIT='192MiB')
if a.soak:env['NESTING_STAGE18_SOAK']='1'
if a.soak or a.profile:env['NESTING_LAB_PROFILE_DIR']='/diagnostics'
if a.compare:env['NESTING_STAGE18_COMPARE']='1'
if a.smoke:env['NESTING_STAGE18_SMOKE']='1'
if a.resume_high:
 previous=(a.resume_high/'test.log').read_text()
 for phase in ['smoke','baseline','step4','step8']:assert '--- PASS: TestStage18MixedLoad/'+phase+' ' in previous
 for line in previous.splitlines():
  if 'LOAD_PHASE_JSON ' in line:
   phase=json.loads(line.split('LOAD_PHASE_JSON ',1)[1])
   assert not any(r['Error'] or r['Status']==401 or r['Status']>=500 and r['Status']!=503 for r in phase['Samples'])
 env['NESTING_STAGE18_RESUME_HIGH']='1'
if a.audit_schema:
 assert a.test=='TestStage18RetainedLoadAudit' and a.audit_schema.startswith('test_') and a.audit_schema[5:].isdigit()
 env['NESTING_STAGE18_AUDIT_SCHEMA']=a.audit_schema
if a.test=='TestStage19PGRecovery':env['NESTING_STAGE19_PG_RESTART']='1'
create=['create','--name',m['name']+'-load-'+uuid.uuid4().hex[:8],'--network',m['network_name'],'--cpus','2','--memory','1536m','--memory-swap','1536m','--pids-limit','256','--user',f'{os.getuid()}:{os.getgid()}','--read-only','--tmpfs','/tmp:rw,exec,mode=1777,size=128m','--tmpfs','/var/lib/postgresql/data:rw,noexec,size=1m','-v',str(root/'backend')+':/work:ro','-v',str(pathlib.Path(a.binary).resolve())+':/tests/load.test:ro','-w','/work/internal/integration']
if a.soak or a.profile:create+=['-v',str(out.resolve())+':/diagnostics:rw']
for k in ['NESTING_TEST_DSN','NESTING_EXPECTED_TEST_USER','NESTING_STAGE18','NESTING_STAGE18_SOAK','NESTING_TEST_PROCESS_BINARY','GOMAXPROCS','GOMEMLIMIT','NESTING_STAGE19_PG_RESTART','NESTING_STAGE18_COMPARE','NESTING_STAGE18_AUDIT_SCHEMA','NESTING_STAGE18_SMOKE','NESTING_STAGE18_RESUME_HIGH','NESTING_LAB_PROFILE_DIR']:
 if k in env:create+=['-e',k]
cid=docker(*create,'--entrypoint','/tests/load.test','postgis/postgis:16-3.4','-test.run=^'+a.test+'$','-test.count=1','-test.v','-test.timeout=25m',env=env)
(out/'identity.json').write_text(json.dumps({'container':cid,'db':m['container'],'network':m['network'],'volume':m['volume'],'gate':gate,'cpu':2,'memory_bytes':1536*1024**2,'pids':256},indent=2))
failed=False
restarted=False
initial_wal=None
try:
 with (out/'test.log').open('w') as log,(out/'resources.jsonl').open('w') as metrics:
  proc=subprocess.Popen(['docker','start','-a',cid],stdout=log,stderr=subprocess.STDOUT);deadline=time.monotonic()+1500
  while proc.poll() is None:
   if time.monotonic()>deadline:raise RuntimeError('bounded wall deadline exceeded')
   if a.test=='TestStage19PGRecovery' and not restarted and 'OWN_PG_RESTART_READY' in (out/'test.log').read_text():
    # Identity was verified above; stop/start this exact OWN container only.
    restarted=True
    docker('stop','--time','5',m['container']);time.sleep(2);docker('start',m['container'])
    for attempt in range(50):
     check=subprocess.run(['docker','exec',m['container'],'pg_isready','-U',m['user'],'-d','nesting_test'],stdout=subprocess.DEVNULL,stderr=subprocess.DEVNULL)
     if check.returncode==0:break
     time.sleep(.2)
   stats=json.loads(docker('stats','--no-stream','--format','{{json .}}',cid));state=json.loads(docker('inspect',cid))[0]['State']
   sql=docker('exec',m['container'],'psql','-X','-v','ON_ERROR_STOP=1','-U',m['user'],'-d','nesting_test','-Atc',"BEGIN READ ONLY; SELECT json_build_object('db_bytes',pg_database_size(current_database()),'deadlocks',deadlocks,'temp_bytes',temp_bytes,'wal_bytes',(SELECT wal_bytes FROM pg_stat_wal),'active',(SELECT count(*) FROM pg_stat_activity WHERE datname=current_database()),'waiting',(SELECT count(*) FROM pg_locks WHERE NOT granted),'idle_tx',(SELECT count(*) FROM pg_stat_activity WHERE datname=current_database() AND state='idle in transaction'),'query_max_ms',(SELECT COALESCE(max(extract(epoch from clock_timestamp()-query_start)*1000),0) FROM pg_stat_activity WHERE datname=current_database() AND state='active' AND pid<>pg_backend_pid()),'tx_max_ms',(SELECT COALESCE(max(extract(epoch from clock_timestamp()-xact_start)*1000),0) FROM pg_stat_activity WHERE datname=current_database() AND pid<>pg_backend_pid()),'long_idle_tx',(SELECT count(*) FROM pg_stat_activity WHERE datname=current_database() AND state='idle in transaction' AND clock_timestamp()-state_change>interval '5 seconds')) FROM pg_stat_database WHERE datname=current_database(); ROLLBACK;")
   peaks=''
   if state['Running']:
    try:peaks=docker('exec',cid,'/bin/sh','-c','cat /sys/fs/cgroup/memory.peak /sys/fs/cgroup/memory.events /sys/fs/cgroup/cpu.stat /sys/fs/cgroup/pids.peak',stderr=subprocess.DEVNULL)
    except subprocess.CalledProcessError:
     latest=json.loads(docker('inspect',cid))[0]['State']
     if latest['Running']:raise
     # Normal process exit may race the read-only metric probe. Exit status is
     # still checked by docker-start and final-state below; no active error hidden.
     peaks='unavailable_after_container_exit'
   metrics.write(json.dumps({'time':time.time(),'cgroup':stats,'peaks':peaks,'oom':state['OOMKilled'],'sql':sql})+'\n');metrics.flush()
   if state['OOMKilled']:raise RuntimeError('STOP OOM')
   for line in sql.splitlines():
    if line.startswith('{'):
     row=json.loads(line)
     if initial_wal is None:initial_wal=float(row['wal_bytes'])
     if float(row['wal_bytes'])-initial_wal>512*1024**2:raise RuntimeError('STOP WAL growth bound')
     if row['db_bytes']>2*1024**3 or row['long_idle_tx']>0 or row['waiting']>50:raise RuntimeError('STOP database safety bound')
   time.sleep(3)
  failed=proc.returncode!=0
finally:
 subprocess.run(['docker','stop','--time','5',cid],stdout=subprocess.DEVNULL,check=False)
 state=json.loads(docker('inspect',cid))[0]['State'];(out/'final-state.json').write_text(json.dumps(state,indent=2))
print('Bounded lab finished:',out,'FAIL' if failed or state['OOMKilled'] else 'PASS')
raise SystemExit(1 if failed or state['OOMKilled'] else 0)
