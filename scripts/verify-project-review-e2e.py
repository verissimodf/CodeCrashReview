#!/usr/bin/env python3
"""Real project review with preserved Python acceptance tests.

Creates a labelled local Git fixture and a test account. Never prints credentials.
The persisted report contains the snapshot, proposals, proofs and judge scores.
"""
import base64,hashlib,hmac
import sys
import json
import pathlib
import secrets
import subprocess
import time
import urllib.request

root = pathlib.Path(__file__).resolve().parents[1]
output = root/'artifacts'/'validation'
output.mkdir(parents=True, exist_ok=True)
stamp = secrets.token_hex(4)
fixture = pathlib.Path('/opt/laytech-ai/repositories')/('python-proof-'+stamp)
fixture.mkdir(parents=True)
(fixture/'calc.py').write_text('def add(a, b):\n    return a - b\n')
(fixture/'test_calc.py').write_text('import unittest\nfrom calc import add\nclass TestAdd(unittest.TestCase):\n    def test_add(self): self.assertEqual(add(2, 3), 5)\n    def test_negative(self): self.assertEqual(add(-2, -3), -5)\n')
(fixture/'server.py').write_text('from http.server import BaseHTTPRequestHandler, HTTPServer\nimport json\nfrom calc import add\nclass Handler(BaseHTTPRequestHandler):\n    def do_GET(self):\n        body=(json.dumps({"sum":add(2,3)})+"\\n").encode()\n        self.send_response(200)\n        self.send_header("Content-Type","application/json")\n        self.send_header("Content-Length",str(len(body)))\n        self.end_headers()\n        self.wfile.write(body)\nHTTPServer(("0.0.0.0",8080),Handler).serve_forever()\n')
for args in [['init','-b','main'],['add','.'],['-c','user.name=CodeCrashReview validation','-c','user.email=validation@example.invalid','commit','-m','test: reproduce invalid winning proposal for correction']]:
    subprocess.run(['git','-C',str(fixture),*args],check=True,capture_output=True)

publish_enabled = '--publish' in sys.argv
source_branch = 'main'
if publish_enabled:
    source_branch = 'review/validation-source-'+stamp
    for args in [['branch','-m',source_branch],['remote','add','origin','https://github.com/verissimodf/CodeCrashReview.git'],['push','origin',source_branch]]:
        subprocess.run(['git','-C',str(fixture),*args],check=True,capture_output=True)
    print('source_fixture_branch='+source_branch,flush=True)

def request(path, data=None, token=None):
    headers = {'Host':'codecrashreview.laytech.com.br'}
    if data is not None: headers['Content-Type']='application/json'
    if token: headers['Authorization']='Bearer '+token
    req = urllib.request.Request('http://127.0.0.1/api'+path, data=None if data is None else json.dumps(data).encode(),headers=headers)
    with urllib.request.urlopen(req, timeout=240) as response: return json.load(response)

account = {'username':'proof_'+stamp,'email':'proof_'+stamp+'@example.invalid','password':'Proof!'+secrets.token_urlsafe(24)}
operator_enabled = '--operator' in sys.argv
if operator_enabled:
    # Host-admin integration test of the explicitly designated operator identity.
    # This is not a password-login test; no credential is printed or committed.
    inspected=json.loads(subprocess.run(['docker','inspect','codecrashreview-backend-1'],capture_output=True,text=True,check=True).stdout)[0]
    env=dict(value.split('=',1) for value in inspected['Config']['Env'])
    q="SELECT id::text,username FROM users WHERE email='contato@laytech.com.br';"
    row=subprocess.run(['docker','compose','-f','compose.host.yml','exec','-T','postgres','psql','-U','codecrash','-d','codecrash','-At','-c',q],capture_output=True,text=True,check=True).stdout.strip().split('|')
    assert len(row)==2 and row[0] in env['REVIEW_OPERATOR_IDS'].split(',')
    def encoded(value): return base64.urlsafe_b64encode(json.dumps(value,separators=(',',':')).encode()).rstrip(b'=')
    now=int(time.time()); message=encoded({'alg':'HS256','typ':'JWT'})+b'.'+encoded({'user_id':row[0],'username':row[1],'type':'access','sub':row[0],'iat':now,'nbf':now,'exp':now+3600})
    token=(message+b'.'+base64.urlsafe_b64encode(hmac.new(env['JWT_SECRET'].encode(),message,hashlib.sha256).digest()).rstrip(b'=')).decode()
    del env
else:
    session = request('/auth/register',account)
    token = session['tokens']['access_token']
    private = output/'.project-session.json'
    private.touch(mode=0o600)
    private.write_text(json.dumps({'account':account,'session':session}))
review = request('/reviews', {
    'source_type':'github' if publish_enabled else 'local','source':'https://github.com/verissimodf/CodeCrashReview' if publish_enabled else '/repositories/'+fixture.name,'branch':source_branch,
    'focus_paths':['calc.py','test_calc.py'], 'validation_profile':'python-unittest-v1', 'validation_directory':'.', 'agent_a':'guardian','agent_b':'refactor',
    'review_focus':'reliability','evaluator_mode':'ai',
    'objective':'Correct add(a, b) in calc.py so it returns the sum, including negative operands. The current subtraction is wrong. Change only calc.py with one exact edit. Preserve all original tests. Tests run independently after your response; do not claim you executed them.'
},token)
print('review_id='+review['id'],flush=True)
(output/'active-project-review.json').write_text(json.dumps(review))
last = None
while True:
    detail = request('/reviews/'+review['id'],token=token)
    if detail['status'] != last:
        last = detail['status']; print('status='+last,flush=True)
    if last in ('completed','failed','awaiting_human'):
        (output/'project-review-e2e.json').write_text(json.dumps(detail,indent=2)+'\n')
        assert last == 'completed', detail.get('error_message')
        assert len(detail['proposals']) == 2
        winner = detail['verdict']['winner']
        selected = detail['proposals'] if winner == 'TIE' else [p for p in detail['proposals'] if p['slot']==winner]
        assert selected and all(p['verification']['proof']['passed'] for p in selected)
        assert detail['verdict']['scores']
        assert detail['evaluation_attempts'] and detail['evaluation_attempts'][-1]['error'] == ''
        print('functional_proof_and_persisted_verdict_confirmed winner='+winner,flush=True)
        break
    time.sleep(5)

if publish_enabled:
    github_token = subprocess.run(['gh','auth','token'],check=True,capture_output=True,text=True).stdout.strip()
    selected_slot = winner if winner != 'TIE' else 'A'
    receipt = request('/reviews/'+review['id']+'/publish', {'token':github_token,'slot':selected_slot,'branch':'review/validated-'+stamp,'tag':'validation-'+stamp,'release':False},token)
    github_token = ''
    assert len(receipt['commit_sha'])==40,receipt
    remote = subprocess.run(['git','ls-remote','https://github.com/verissimodf/CodeCrashReview.git','refs/heads/'+receipt['branch']],check=True,capture_output=True,text=True).stdout.split()[0]
    assert remote==receipt['commit_sha'],receipt
    saved = request('/reviews/'+review['id']+'/publications',token=token)
    assert saved[0]['status']=='completed' and saved[0]['receipt']['commit_sha']==remote,saved
    (output/'publication-e2e.json').write_text(json.dumps({'receipt':receipt,'remote_sha':remote,'persisted':True},indent=2)+'\n')
    print('publication_remote_and_database_verified commit='+remote,flush=True)

if operator_enabled:
    targets=request('/reviews/deployment-targets',token=token)
    assert targets['authorized'] and any(t['id']=='python-validation' for t in targets['targets'])
    deployed=request('/reviews/'+review['id']+'/deploy',{'target':'python-validation','slot':winner if winner!='TIE' else 'A'},token)
    assert deployed['functional_probe_passed'] and deployed['review_id']==review['id']
    with urllib.request.urlopen(deployed['url']+'/sum',timeout=10) as response: assert json.load(response)=={'sum':5}
    (output/'deployment-e2e.json').write_text(json.dumps(deployed,indent=2)+'\n')
    print('deployment_functional_response_verified',flush=True)
