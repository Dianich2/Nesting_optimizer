"""Create a unique disposable PostGIS lab, never reuse another database/volume.
Example: python3 scripts/create_stages17_19_lab.py --metadata /secure/path/lab.json
The selected metadata file MUST NOT already exist. Keep it private; it contains
synthetic credentials. Requires the existing local postgis/postgis:16-3.4 image.
No cleanup of other containers, volumes, databases or networks is performed.
"""
import argparse,json,os,pathlib,secrets,subprocess,time,uuid
parser=argparse.ArgumentParser();parser.add_argument('--metadata',type=pathlib.Path,required=True);args=parser.parse_args()
assert not args.metadata.exists(), 'refusing to overwrite lab metadata'
def docker(*argv,**kwargs):
 return subprocess.check_output(['docker',*argv],text=True,**kwargs).strip()
docker('image','inspect','postgis/postgis:16-3.4')
name='nesting-1719-'+uuid.uuid4().hex[:12]
metadata={'name':name,'network_name':name+'-bridge','volume':name+'-data','db_container':name+'-db','user':'nesting_lab','password':secrets.token_urlsafe(40),'access_secret':secrets.token_urlsafe(48),'refresh_secret':secrets.token_urlsafe(48)}
with os.fdopen(os.open(args.metadata,os.O_WRONLY|os.O_CREAT|os.O_EXCL,0o600),'w') as f:json.dump(metadata,f)
def save():args.metadata.write_text(json.dumps(metadata,indent=2))
metadata['network']=docker('network','create','--label','nesting.lab='+name,metadata['network_name']);save()
docker('volume','create','--label','nesting.lab='+name,metadata['volume']);save()
env=os.environ.copy();env['POSTGRES_PASSWORD']=metadata['password']
metadata['container']=docker('run','-d','--name',metadata['db_container'],'--label','nesting.lab='+name,'--network',metadata['network_name'],'--cpus','1','--memory','1g','--memory-swap','1g','--pids-limit','128','-p','127.0.0.1::5432','-v',metadata['volume']+':/var/lib/postgresql/data','-e','POSTGRES_PASSWORD','-e','POSTGRES_USER='+metadata['user'],'-e','POSTGRES_DB=nesting_test','postgis/postgis:16-3.4',env=env);save()
for attempt in range(80):
 ready=subprocess.run(['docker','exec',metadata['container'],'pg_isready','-U',metadata['user'],'-d','nesting_test'],stdout=subprocess.DEVNULL,stderr=subprocess.DEVNULL)
 if ready.returncode==0:break
 time.sleep(.25)
else:raise SystemExit('Own lab readiness failed; IDs preserved for diagnosis')
gate=docker('exec',metadata['container'],'psql','-X','-v','ON_ERROR_STOP=1','-U',metadata['user'],'-d','nesting_test','-Atc','BEGIN READ ONLY;SELECT current_database(),current_user,postgis_version();ROLLBACK;')
assert 'nesting_test|nesting_lab|3.4' in gate
print('Created own bounded lab:',name,'; private metadata:',args.metadata)
