"""Run isolated browser acceptance against this task's dedicated test DB.

No .env loading. Credentials below are exclusively for the disposable fixture DB.
Only the API PID spawned here is stopped; Playwright owns its Vite/browser.
"""
import argparse
import json
import os
import pathlib
import subprocess
import tempfile
import time
import urllib.request

parser = argparse.ArgumentParser()
parser.add_argument('--label', required=True)
parser.add_argument('--api-binary', default='/tmp/nesting-stage16-api.test')
args, extra = parser.parse_known_args()
root = pathlib.Path(__file__).resolve().parents[1]
work = pathlib.Path(tempfile.mkdtemp(prefix=f'nesting-stage16-{args.label}-'))
ready = work / 'ready.json'
port = subprocess.check_output(['docker', 'port', 'nesting-stage16-db', '5432/tcp'], text=True).strip().rsplit(':', 1)[1]
assert port.isdigit()
for _ in range(60):
    gate = subprocess.run(['docker', 'exec', 'nesting-stage16-db', 'psql', '-X', '-v', 'ON_ERROR_STOP=1', '-U', 'postgres', '-d', 'nesting_test', '-Atc', 'BEGIN READ ONLY; SELECT current_database(),current_user,postgis_version(); ROLLBACK;'], capture_output=True, text=True)
    if gate.returncode == 0 and 'nesting_test|postgres|3.4' in gate.stdout:
        (work / 'db-gate.log').write_text(gate.stdout)
        break
    time.sleep(.25)
else:
    raise SystemExit('dedicated test DB gate failed')
env = os.environ.copy()
env.update(NESTING_TEST_DSN=f'postgres://postgres:stage16_test_db_only@127.0.0.1:{port}/nesting_test?sslmode=disable&connect_timeout=5', NESTING_BROWSER_API='1', NESTING_E2E_READY_FILE=str(ready), NESTING_E2E_ORIGIN='http://localhost:5188', GOMAXPROCS='2', GOMEMLIMIT='512MiB')
with (work / 'api.log').open('w') as log:
    api = subprocess.Popen([args.api_binary, '-test.run', '^TestBrowserAcceptanceAPI$', '-test.count=1', '-test.v', '-test.timeout=25m'], cwd=root / 'backend/internal/integration', env=env, stdout=log, stderr=subprocess.STDOUT)
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
        assert data['database'] == 'nesting_test' and data['schema'].startswith('test_') and data['pid'] == api.pid
        assert data['api'].startswith('http://127.0.0.1:') and data['stats'].startswith('http://127.0.0.1:')
        with urllib.request.urlopen(data['stats'], timeout=2) as response:
            assert json.load(response)['optimize_calls'] == 0
        env.update(PLAYWRIGHT_JSON_OUTPUT_FILE=str(work / 'results.json'), NESTING_API_TARGET=data['api'], NESTING_E2E_PORT='5188', NESTING_E2E_STATS=data['stats'], E2E_CHROME='/usr/bin/google-chrome')
        print('Isolated topology:', json.dumps(data), 'artifacts:', str(work), flush=True)
        cmd = ['npm', 'run', 'test:e2e', '--', '--workers=1', '--retries=0', '--trace=off', '--reporter=list,json', '--output', str(work / 'browser')] + extra
        result = subprocess.run(cmd, cwd=root / 'frontend', env=env, timeout=1200)
        raise SystemExit(result.returncode)
    finally:
        if api.poll() is None:
            api.terminate()
            try:
                api.wait(timeout=8)
            except subprocess.TimeoutExpired:
                api.kill()
                api.wait(timeout=3)
