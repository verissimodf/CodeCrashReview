#!/usr/bin/env python3
"""Build an operator-curated offline dependency image from lockfiles only.
No repository source, secrets, hooks or lifecycle scripts enter the build context.
Print the immutable image ID for VALIDATOR_*_IMAGE; never change host configuration.
"""
import argparse,json,pathlib,re,shutil,subprocess,tempfile
p=argparse.ArgumentParser();p.add_argument('profile',choices=['go','node','python']);p.add_argument('directory',type=pathlib.Path);p.add_argument('--base',required=True);a=p.parse_args()
if not re.fullmatch(r'[a-zA-Z0-9./:_-]+@sha256:[a-f0-9]{64}',a.base):p.error('base must be pinned by SHA-256 digest')
root=a.directory.resolve()
names={'go':['go.mod','go.sum'],'node':['package.json','package-lock.json'],'python':['requirements.lock']}[a.profile]
contents={}
for name in names:
 path=root/name
 if path.is_symlink() or not path.is_file() or path.stat().st_size>4*1024*1024:p.error('missing, oversized or symlinked lockfile: '+name)
 contents[name]=path.read_text()
if a.profile=='node':
 lock=json.loads(contents['package-lock.json'])
 if lock.get('lockfileVersion') not in (2,3):p.error('npm lockfile version 2 or 3 is required')
 for name,package in lock.get('packages',{}).items():
  if not name:continue
  resolved=package.get('resolved','')
  if package.get('link') or not resolved.startswith('https://registry.npmjs.org/') or not package.get('integrity'):p.error('only integrity-pinned npm registry packages are accepted')
 manifest=json.loads(contents['package.json']);manifest.pop('scripts',None);contents['package.json']=json.dumps(manifest)
 command='WORKDIR /opt/dependencies\nRUN npm ci --ignore-scripts --no-audit --no-fund --registry=https://registry.npmjs.org && ln -s /opt/dependencies/node_modules /node_modules'
elif a.profile=='go':
 if re.search(r'^\s*(replace|toolchain)\b',contents['go.mod'],re.M):p.error('replace/toolchain directives require manual operator review')
 command='ENV GOTOOLCHAIN=local GOPROXY=https://proxy.golang.org GOSUMDB=sum.golang.org GOMODCACHE=/opt/dependencies/gomodcache\nWORKDIR /opt/dependencies\nRUN go mod download && chmod -R a+rX /opt/dependencies/gomodcache'
else:
 logical=contents['requirements.lock'].replace('\\\n',' ')
 for line in logical.splitlines():
  text=line.split('#',1)[0].strip()
  if not text:continue
  words=text.split()
  if not re.fullmatch(r'[A-Za-z0-9_.-]+(?:\[[A-Za-z0-9_,.-]+\])?==[A-Za-z0-9_.+!-]+',words[0]):p.error('requirements must use exact versions; URLs and index options are forbidden')
  if len(words)<2 or any(not re.fullmatch(r'--hash=sha256:[a-f0-9]{64}',word) for word in words[1:]):p.error('each requirement needs SHA-256 hashes and no additional options')
 command='RUN python -m pip install --no-cache-dir --only-binary=:all: --require-hashes --index-url=https://pypi.org/simple -r /opt/dependencies/requirements.lock'
with tempfile.TemporaryDirectory(prefix='ccr-dependencies-') as work:
 directory=pathlib.Path(work)
 for name,content in contents.items():(directory/name).write_text(content)
 dockerfile='FROM '+a.base+'\nWORKDIR /opt/dependencies\nCOPY '+ ' '.join(names)+' ./\n'+command+'\n'
 (directory/'Dockerfile').write_text(dockerfile)
 imageid=directory/'image-id'
 subprocess.run(['docker','build','--iidfile',str(imageid),str(directory)],check=True)
 print('Validated dependency image: '+imageid.read_text().strip())
