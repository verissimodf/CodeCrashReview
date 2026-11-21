#!/usr/bin/env python3
"""Real isolated baseline/candidate acceptance checks for each project runtime."""
import base64,json,pathlib,subprocess
ROOT=pathlib.Path(__file__).resolve().parents[1]
CLIENT="""import http.client,socket,sys
class Unix(http.client.HTTPConnection):
 def connect(self):
  self.sock=socket.socket(socket.AF_UNIX,socket.SOCK_STREAM);self.sock.settimeout(650);self.sock.connect('/socket/validator.sock')
c=Unix('localhost');c.request('POST','/validate',sys.stdin.read(),{'Content-Type':'application/json'});r=c.getresponse();print(r.read().decode());sys.exit(0 if r.status==200 else 1)
"""
fixtures={
'python-unittest-v1':('calc.py',{'calc.py':'def add(a,b): return a-b\n','test_calc.py':'import unittest\nfrom calc import add\nclass TestAdd(unittest.TestCase):\n def test_add(self): self.assertEqual(add(2,3),5)\n'}),
'node-test-v1':('calc.js',{'calc.js':'exports.add=(a,b)=>a-b;\n','calc.test.js':"const {test}=require('node:test'); const assert=require('node:assert/strict'); const {add}=require('./calc'); test('adds',()=>assert.equal(add(2,3),5));\n"}),
'go-test-v1':('calc.go',{'go.mod':'module proof\ngo 1.24\n','calc.go':'package proof\nfunc Add(a,b int) int { return a-b }\n','calc_test.go':'package proof\nimport "testing"\nfunc TestAdd(t *testing.T) { if Add(2,3)!=5 { t.Fatal("addition failed") } }\n'})}
results={}
for profile,(target,files) in fixtures.items():
 payload={'profile':profile,'directory':'.','files':{k:base64.b64encode(v.encode()).decode() for k,v in files.items()},'edits':[{'file':target,'old':'a-b','new':'a+b'}]}
 run=subprocess.run(['docker','compose','-f','compose.host.yml','exec','-T','validator','python3','-c',CLIENT],input=json.dumps(payload),text=True,capture_output=True,cwd=ROOT,timeout=660,check=True)
 proof=json.loads(run.stdout)
 assert not proof['baseline']['passed'] and proof['baseline']['exit_code']!=0,proof
 assert proof['passed'] and proof['test_count']>0 and proof['tests_preserved'],proof
 assert proof['source_sha256']!=proof['baseline_sha256'],proof
 results[profile]=proof
 print(profile,'baseline failed; candidate passed; original tests preserved',flush=True)
out=ROOT/'artifacts/validation/project-proofs.json';out.parent.mkdir(parents=True,exist_ok=True);out.write_text(json.dumps(results,indent=2)+'\n')
