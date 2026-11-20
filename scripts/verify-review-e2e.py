#!/usr/bin/env python3
"""Real two-agent replay of the formerly selected, uncompilable patch.

Creates a labelled local Git fixture and a test account. Never prints credentials.
The persisted report contains the snapshot, proposals, proofs and judge scores.
"""
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
fixture = pathlib.Path('/opt/laytech-ai/repositories')/('grade-replay-'+stamp)
(fixture/'backend/internal/arena').mkdir(parents=True)
source = (root/'backend/internal/arena/grade.go').read_text()
assert source.count('v.Cmp(big.NewInt(c[2])) == 0') == 1
source = source.replace('v.Cmp(big.NewInt(c[2])) == 0','v.Cmp(big.NewInt(c[2])) == 0 && c[2].Sign() == 1')
(fixture/'backend/internal/arena/grade.go').write_text(source)
for args in [['init','-b','main'],['add','.'],['-c','user.name=CodeCrashReview validation','-c','user.email=validation@example.invalid','commit','-m','test: reproduce invalid winning proposal for correction']]:
    subprocess.run(['git','-C',str(fixture),*args],check=True,capture_output=True)

def request(path, data=None, token=None):
    headers = {'Host':'codecrashreview.laytech.com.br'}
    if data is not None: headers['Content-Type']='application/json'
    if token: headers['Authorization']='Bearer '+token
    req = urllib.request.Request('http://127.0.0.1/api'+path, data=None if data is None else json.dumps(data).encode(),headers=headers)
    with urllib.request.urlopen(req, timeout=30) as response: return json.load(response)

account = {'username':'proof_'+stamp,'email':'proof_'+stamp+'@example.invalid','password':'Proof!'+secrets.token_urlsafe(24)}
session = request('/auth/register',account)
token = session['tokens']['access_token']
private = output/'.session.json'
private.touch(mode=0o600)
private.write_text(json.dumps({'account':account,'session':session}))
review = request('/reviews', {
    'source_type':'local','source':'/repositories/'+fixture.name,'branch':'main',
    'focus_paths':['backend/internal/arena/grade.go'], 'agent_a':'guardian','agent_b':'refactor',
    'review_focus':'reliability','evaluator_mode':'ai',
    'objective':'Correct this reproduced invalid winning proposal. In grade(), c[2] is int64 and has no Sign method. Remove the erroneous sign restriction; zero and negative expected results must receive full credit. The original big.Int comparison is sufficient and already avoids intermediate overflow. Produce a minimal exact edit. Independent compilation and functional tests will be executed by the system; do not claim you ran them.'
},token)
print('review_id='+review['id'],flush=True)
(output/'active-review.json').write_text(json.dumps(review))
last = None
while True:
    detail = request('/reviews/'+review['id'],token=token)
    if detail['status'] != last:
        last = detail['status']; print('status='+last,flush=True)
    if last in ('completed','failed','awaiting_human'):
        (output/'review-e2e.json').write_text(json.dumps(detail,indent=2)+'\n')
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
