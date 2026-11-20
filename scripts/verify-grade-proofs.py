#!/usr/bin/env python3
"""Reproduce the failed winning proposal and the correct grading behavior."""
import json
import pathlib
import subprocess

root = pathlib.Path(__file__).resolve().parents[1]
source = (root / 'backend/internal/arena/grade.go').read_text()
client = '''import http.client,json,socket,sys
class Unix(http.client.HTTPConnection):
 def connect(self):
  self.sock=socket.socket(socket.AF_UNIX,socket.SOCK_STREAM);self.sock.settimeout(210);self.sock.connect('/socket/validator.sock')
c=Unix('localhost');c.request('POST','/validate',sys.stdin.read(),{'Content-Type':'application/json'});r=c.getresponse();print(r.read().decode());sys.exit(0 if r.status==200 else 1)
'''
proofs = {}
for name, content, expected in [
    ('original_invalid_winner', source.replace('v.Cmp(big.NewInt(c[2])) == 0', 'v.Cmp(big.NewInt(c[2])) == 0 && c[2].Sign() == 1'), False),
    ('compilable_sign_regression', source.replace('v.Cmp(big.NewInt(c[2])) == 0', 'v.Cmp(big.NewInt(c[2])) == 0 && c[2] > 0'), False),
    ('correct_grading', source, True),
]:
    payload = json.dumps({'profile':'arena-grade-v1','source':content})
    proc = subprocess.run(['docker','compose','-f','compose.host.yml','exec','-T','validator','python3','-c',client], cwd=root, input=payload,text=True,capture_output=True,timeout=230,check=True)
    proof = json.loads(proc.stdout)
    assert proof['passed'] is expected, (name,proof)
    if name == 'original_invalid_winner': assert 'Sign undefined' in proof['output'], proof
    if name == 'compilable_sign_regression': assert 'FAIL: TestGradeAcceptance/signs_and_zero' in proof['output'], proof
    proofs[name] = proof
    print(name, 'exit_code='+str(proof['exit_code']), 'expected_result_confirmed',flush=True)
out = root/'artifacts'/'validation'
out.mkdir(parents=True,exist_ok=True)
(out/'grade-proofs.json').write_text(json.dumps(proofs,indent=2)+'\n')
