"""Run isolated browser acceptance against this task's dedicated test DB.

No .env loading. Credentials below are exclusively for the disposable fixture DB.
Only the API PID spawned here is stopped; Playwright owns its Vite/browser.
"""
import argparse
import json
import os
import secrets
import pathlib
import subprocess
import tempfile
import time
import urllib.request

parser = argparse.ArgumentParser()
parser.add_argument('--label', required=True)
parser.add_argument('--api-binary', default='/tmp/nesting-stage16-api.test')
parser.add_argument('--lab-metadata')
args, extra = parser.parse_known_args()
root = pathlib.Path(__file__).resolve().parents[1]
work = pathlib.Path(tempfile.mkdtemp(prefix=f'nesting-stage16-{args.label}-'))
ready = work / 'ready.json'
lab=None
if args.lab_metadata:
    private=pathlib.Path(args.lab_metadata)
    assert private.stat().st_mode & 0o077 == 0
    lab=json.loads(private.read_text())
    inspected=json.loads(subprocess.check_output(['docker','inspect',lab['container']],text=True))[0]
    assert inspected['Id']==lab['container'] and inspected['NetworkSettings']['Networks'][lab['network_name']]['NetworkID']==lab['network']
    assert any(v.get('Name')==lab['volume'] for v in inspected['Mounts'])
    for field in ['access_secret','refresh_secret']:
        lab.setdefault(field,secrets.token_urlsafe(48))
    private.write_text(json.dumps(lab))
db_container=lab['db_container'] if lab else 'nesting-stage16-db'
db_user=lab['user'] if lab else 'postgres'
db_password=lab['password'] if lab else 'stage16_test_db_only'
port = subprocess.check_output(['docker', 'port', db_container, '5432/tcp'], text=True).strip().rsplit(':', 1)[1]
assert port.isdigit()
for _ in range(60):
    gate = subprocess.run(['docker', 'exec', db_container, 'psql', '-X', '-v', 'ON_ERROR_STOP=1', '-U', db_user, '-d', 'nesting_test', '-Atc', 'BEGIN READ ONLY; SELECT current_database(),current_user,postgis_version(); ROLLBACK;'], capture_output=True, text=True)
    if gate.returncode == 0 and 'nesting_test|'+db_user+'|3.4' in gate.stdout:
        (work / 'db-gate.log').write_text(gate.stdout)
        break
    time.sleep(.25)
else:
    raise SystemExit('dedicated test DB gate failed')
env = os.environ.copy()
env.update(NESTING_TEST_DSN=f'postgres://{db_user}:{db_password}@127.0.0.1:{port}/nesting_test?sslmode=disable&connect_timeout=5', NESTING_BROWSER_API='1', NESTING_E2E_READY_FILE=str(ready), NESTING_E2E_ORIGIN='http://localhost:5188', GOMAXPROCS='2', GOMEMLIMIT='512MiB')
with (work / 'api.log').open('w') as log:
    container_id=None
    api_args=[args.api_binary,'-test.run','^TestBrowserAcceptanceAPI$','-test.count=1','-test.v','-test.timeout=25m']
    if lab:
        env.update(NESTING_EXPECTED_TEST_USER=db_user,NESTING_LAB_ACCESS_SECRET=lab['access_secret'],NESTING_LAB_REFRESH_SECRET=lab['refresh_secret'])
        create=['docker','create','--name',work.name+'-api','--network','host','--cpus','1','--memory','768m','--memory-swap','768m','--pids-limit','256','--read-only','--user',f'{os.getuid()}:{os.getgid()}','--tmpfs','/tmp:rw,exec,mode=1777,size=128m','--tmpfs','/var/lib/postgresql/data:rw,size=1m','-v',str(root/'backend')+':'+str(root/'backend')+':ro','-v',str(work)+':'+str(work)+':rw','-v',str(pathlib.Path(args.api_binary).resolve())+':/tests/api.test:ro','-w',str(root/'backend/internal/integration')]
        for key in ['NESTING_TEST_DSN','NESTING_EXPECTED_TEST_USER','NESTING_BROWSER_API','NESTING_E2E_READY_FILE','NESTING_E2E_ORIGIN','NESTING_LAB_ACCESS_SECRET','NESTING_LAB_REFRESH_SECRET','GOMAXPROCS','GOMEMLIMIT']:
            create+=['-e',key]
        if 'race' in args.api_binary:
            create+=['-v','/usr/lib/x86_64-linux-gnu:/hostlib:ro','--entrypoint','/hostlib/ld-linux-x86-64.so.2','postgis/postgis:16-3.4','--library-path','/hostlib']
        else:
            create+=['--entrypoint','/tests/api.test','postgis/postgis:16-3.4']
        tail=api_args[1:]
        if 'race' in args.api_binary:tail=['/tests/api.test']+tail
        container_id=subprocess.check_output(create+tail,env=env,text=True).strip()
        api_args=['docker','start','-a',container_id]
    api = subprocess.Popen(api_args,cwd=root/'backend/internal/integration',env=env,stdout=log,stderr=subprocess.STDOUT)
    try:
        for _ in range(150):
            if api.poll() is not None:
                raise SystemExit(f'test API exited; diagnostics {work}')
            if ready.exists():
                break
            time.sleep(.2)
        else:
            raise SystemExit('test API readiness timeout')
        data = json.loads(ready.read_text())
        assert data['database'] == 'nesting_test' and data['schema'].startswith('test_') and data['pid'] == (1 if container_id else api.pid)
        assert data['api'].startswith('http://127.0.0.1:') and data['stats'].startswith('http://127.0.0.1:')
        with urllib.request.urlopen(data['stats'], timeout=2) as response:
            assert json.load(response)['optimize_calls'] == 0
        if container_id:
            data['container_id']=container_id
            data['host_pid']=json.loads(subprocess.check_output(['docker','inspect',container_id],text=True))[0]['State']['Pid']
            ready.write_text(json.dumps(data))
        env.update(PLAYWRIGHT_JSON_OUTPUT_FILE=str(work / 'results.json'), NESTING_API_TARGET=data['api'], NESTING_E2E_PORT='5188', NESTING_E2E_STATS=data['stats'], E2E_CHROME='/usr/bin/google-chrome')
        print('Isolated topology:', json.dumps(data), 'artifacts:', str(work), flush=True)
        cmd = ['npm', 'run', 'test:e2e', '--', '--workers=1', '--retries=0', '--trace=off', '--reporter=list,json', '--output', str(work / 'browser')] + extra
        result = subprocess.run(cmd, cwd=root / 'frontend', env=env, timeout=1200)
        raise SystemExit(result.returncode)
    finally:
        if container_id:
            subprocess.run(['docker','stop','--timeout=5',container_id],stdout=subprocess.DEVNULL,check=True)
        if api.poll() is None:
            if not container_id: api.terminate()
            try:
                api.wait(timeout=8)
            except subprocess.TimeoutExpired:
                api.kill()
                api.wait(timeout=3)
