#!/usr/bin/env python3
"""Host-admin fault injection: failed candidate must restore the serving revision."""
import base64,hashlib,json,pathlib,subprocess,urllib.request
root=pathlib.Path(__file__).resolve().parents[1]
statepath=pathlib.Path.home()/'.local/share/codecrashreview/validator-jobs/deployments/python-validation.json'
state=json.loads(statepath.read_text());directory=statepath.parent/state['current']
files={str(p.relative_to(directory)):p.read_bytes() for p in directory.rglob('*') if p.is_file()}
assert b'return a + b' in files['calc.py']
files['calc.py']=files['calc.py'].replace(b'return a + b',b'return a - b')
h=hashlib.sha256()
for name in sorted(files):h.update(name.encode()+b'\0'+hashlib.sha256(files[name]).digest())
payload={**state,'files':{k:base64.b64encode(v).decode() for k,v in files.items()},'source_sha256':h.hexdigest()}
client="""import http.client,socket,sys
class Unix(http.client.HTTPConnection):
 def connect(self):
  self.sock=socket.socket(socket.AF_UNIX,socket.SOCK_STREAM);self.sock.settimeout(180);self.sock.connect('/socket/validator.sock')
c=Unix('localhost');c.request('POST','/deploy',sys.stdin.read(),{'Content-Type':'application/json'});r=c.getresponse();print(r.status);sys.exit(0 if r.status==400 else 1)
"""
subprocess.run(['docker','compose','-f','compose.host.yml','exec','-T','validator','python3','-c',client],cwd=root,input=json.dumps(payload),text=True,check=True,capture_output=True,timeout=200)
after=json.loads(statepath.read_text());assert after['current']==state['current']
with urllib.request.urlopen(state['url']+'/sum',timeout=10) as response:assert json.load(response)=={'sum':5}
report={'target':state['target'],'previous_container':state['current'],'restored_container':after['current'],'failed_candidate_rejected':True,'functional_response_after_rollback':{'sum':5}}
(root/'artifacts/validation/rollback-e2e.json').write_text(json.dumps(report,indent=2)+'\n')
print('failed candidate rejected; previous service restored and functionally verified')
