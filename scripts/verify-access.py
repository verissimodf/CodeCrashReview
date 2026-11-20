#!/usr/bin/env python3
"""Exercise authentication, ownership and invalid repository requests on the host."""
import json
import pathlib
import secrets
import urllib.request
import urllib.error

root = pathlib.Path(__file__).resolve().parents[1]
out = root/'artifacts/validation'
session = json.loads((out/'.session.json').read_text())
review = json.loads((out/'review-e2e.json').read_text())

def request(path, data=None, token=None):
    headers={'Host':'codecrashreview.laytech.com.br','Content-Type':'application/json'}
    if token: headers['Authorization']='Bearer '+token
    req=urllib.request.Request('http://127.0.0.1/api'+path, data=None if data is None else json.dumps(data).encode(),headers=headers)
    try:
        with urllib.request.urlopen(req,timeout=15) as response: return response.status,json.load(response)
    except urllib.error.HTTPError as error: return error.code,json.load(error)

results={}
def check(name,actual,expected):
    assert actual==expected,(name,actual,expected)
    results[name]={'actual':actual,'expected':expected}

check('anonymous_reviews',request('/reviews')[0],401)
stamp=secrets.token_hex(4)
status,other=request('/auth/register',{'username':'isolation_'+stamp,'email':stamp+'@example.invalid','password':'Proof!'+secrets.token_urlsafe(20)})
assert status in (200,201),other
token=other['tokens']['access_token']
check('other_owner_cannot_read',request('/reviews/'+review['id'],token=token)[0],404)
check('other_owner_cannot_decide',request('/reviews/'+review['id']+'/verdict',{'winner':'A','rationale':'Unauthorized attempt must fail.'},token)[0],409)
check('other_owner_list_empty',len(request('/reviews',token=token)[1]),0)
base={'source_type':'github','source':'https://github.com/verissimodf/CodeCrashReview','branch':'main','agent_a':'guardian','agent_b':'refactor','review_focus':'reliability','evaluator_mode':'ai'}
for name,change in [('absolute_focus',{'focus_paths':['/etc/passwd']}),('traversal_focus',{'focus_paths':['../secret']}),('incompatible_specialty',{'review_focus':'security'}),('duplicate_agents',{'agent_b':'guardian'}),('non_github_source',{'source':'http://127.0.0.1/private'})]:
    check(name,request('/reviews',base|change,token)[0],400)
(out/'access-checks.json').write_text(json.dumps(results,indent=2)+'\n')
print(str(len(results))+' access-control checks passed')
