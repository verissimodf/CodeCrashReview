"""Operator-configured Python service delivery. Never accept caller commands or ports."""
import base64,json,os,pathlib,re,shutil,subprocess,time,uuid
import project
ROOT=pathlib.Path('/jobs/deployments')

def docker(*args,check=True):
 return subprocess.run(['docker',*args],capture_output=True,text=True,timeout=45,check=check)

def targets():
 path=os.environ.get('DEPLOY_TARGETS_FILE','/config/deploy-targets.json')
 try:return json.loads(pathlib.Path(path).read_text())
 except FileNotFoundError:return []

def health(name,target):
 # Probe executes trusted Python stdlib code, not any repository health script.
 code="import urllib.request,sys; r=urllib.request.urlopen('http://127.0.0.1:8080"+target['health_path']+"',timeout=2); body=r.read(65536).decode(); sys.exit(0 if r.status==200 and body=="+repr(target['expected_body'])+" else 1)"
 for _ in range(15):
  if docker('exec',name,'python','-I','-c',code,check=False).returncode==0:return True
  time.sleep(1)
 return False

def deploy(data):
 target=next((t for t in targets() if t['id']==data.get('target')),None)
 if not target:raise ValueError('deployment target is not configured')
 identifier=target['id']
 if not re.fullmatch(r'[a-z0-9-]{1,40}',identifier):raise ValueError('invalid target configuration')
 if target.get('profile')!='python-unittest-v1':raise ValueError('unsupported deployment runtime')
 if not re.fullmatch(r'/[A-Za-z0-9/_-]*',target['health_path']):raise ValueError('invalid health path')
 image=os.environ.get('VALIDATOR_PYTHON_IMAGE','')
 if data.get('image')!=image:raise ValueError('deployment runtime differs from tested runtime')
 if not isinstance(target.get('port'),int) or not 1024<=target['port']<=65535:raise ValueError('invalid target port')
 files={}
 for name,value in data.get('files',{}).items():
  project.safe_path(name);files[name]=base64.b64decode(value,validate=True)
 if not 0<len(files)<=1000 or sum(map(len,files.values()))>project.LIMIT:raise ValueError('invalid deployment size')
 if project.manifest_hash(files)!=data.get('source_sha256'):raise ValueError('deployment source identity mismatch')
 entry=str(project.safe_path(target['entrypoint']))
 if entry not in files:raise ValueError('configured service entrypoint missing')
 ROOT.mkdir(exist_ok=True,mode=0o755)
 statefile=ROOT/(identifier+'.json')
 state=json.loads(statefile.read_text()) if statefile.exists() else {}
 previous=state.get('current')
 name='ccr-deploy-'+identifier+'-'+uuid.uuid4().hex[:10]
 directory=ROOT/name;directory.mkdir(mode=0o755)
 for path,content in files.items():
  dest=directory/path;dest.parent.mkdir(parents=True,exist_ok=True,mode=0o755);dest.write_bytes(content);dest.chmod(0o444)
 network='ccr-deploy-'+identifier
 if docker('network','inspect',network,check=False).returncode:docker('network','create','--internal',network)
 host=os.environ['VALIDATOR_HOST_ROOT']+'/deployments/'+name
 try:
  if previous:docker('stop',previous)
  docker('run','-d','--pull=never','--name',name,'--network',network,'--read-only','--cap-drop=ALL','--security-opt=no-new-privileges','--user=65534:65534','--pids-limit=64','--memory=256m','--cpus=0.5','--tmpfs=/tmp:rw,nosuid,nodev,size=32m','--log-driver=json-file','--log-opt=max-size=1m','--log-opt=max-file=2','--publish','127.0.0.1:'+str(target['port'])+':8080','--mount','type=bind,src='+host+',dst=/project,readonly','--workdir=/project','--env=PYTHONDONTWRITEBYTECODE=1',image,'python','-B','/project/'+entry)
  if not health(name,target):raise ValueError('deployed service failed the configured functional probe')
 except Exception:
  docker('rm','-f',name,check=False);shutil.rmtree(directory)
  if previous:
   docker('start',previous)
   if not health(previous,target):raise ValueError('deployment failed and previous service could not be verified')
  raise ValueError('deployment failed; previous service restored' if previous else 'deployment failed; no previous service existed')
 old=state.get('previous')
 if old:
  docker('rm','-f',old,check=False)
  olddir=ROOT/old
  if old.startswith('ccr-deploy-'+identifier+'-') and olddir.is_dir():shutil.rmtree(olddir)
 state={'current':name,'previous':previous,'source_sha256':data['source_sha256'],'review_id':data['review_id'],'image':image,'target':identifier,'url':'http://127.0.0.1:'+str(target['port']),'functional_probe_passed':True}
 temporary=statefile.with_suffix('.tmp');temporary.write_text(json.dumps(state));temporary.replace(statefile)
 return state
