"""Fixed project test profiles. Inputs never choose Docker flags or commands."""
import base64
import hashlib
import json
import os
import pathlib
import re
import subprocess
import threading
import uuid

LIMIT = 20 * 1024 * 1024
PROFILES = {
    'go-test-v1': ('VALIDATOR_IMAGE', ['go', 'test', '-json', '-count=1', '-timeout=120s', './...']),
    'python-unittest-v1': ('VALIDATOR_PYTHON_IMAGE', ['python', '-B', '-m', 'unittest', 'discover', '-s', '.', '-p', 'test*.py', '-v']),
    'node-test-v1': ('VALIDATOR_NODE_IMAGE', ['node', '--test']),
}


def safe_path(value):
    if not isinstance(value, str) or not value or '\\' in value or '\x00' in value:
        raise ValueError('invalid repository path')
    path = pathlib.PurePosixPath(value)
    if path.is_absolute() or '..' in path.parts or '.git' in path.parts or str(path) != value:
        raise ValueError('unsafe repository path')
    return path


def is_test(path):
    name = pathlib.PurePosixPath(path).name
    return name.endswith('_test.go') or (name.startswith('test') and name.endswith('.py')) or bool(re.search(r'\.(test|spec)\.[cm]?js$', name)) or any(p in ('test','tests','__tests__') for p in pathlib.PurePosixPath(path).parts)


def manifest_hash(files):
    digest = hashlib.sha256()
    for name in sorted(files):
        digest.update(name.encode()+b'\x00'+hashlib.sha256(files[name]).digest())
    return digest.hexdigest()


def prepare(data):
    profile = data.get('profile')
    if profile not in PROFILES:
        raise ValueError('unknown project profile')
    directory = data.get('directory', '.')
    if directory != '.': safe_path(directory)
    encoded = data.get('files')
    if not isinstance(encoded, dict) or not 0 < len(encoded) <= 1000:
        raise ValueError('project must contain 1..1000 files')
    original = {}
    size = 0
    for name, value in encoded.items():
        safe_path(name)
        if not isinstance(value,str): raise ValueError('invalid file encoding')
        content = base64.b64decode(value, validate=True)
        size += len(content)
        if size > LIMIT or len(content)>1024*1024:
            raise ValueError('project exceeds validation size limit')
        original[name] = content
    tests = {name: content for name,content in original.items() if is_test(name) and (directory=='.' or name.startswith(directory+'/'))}
    if not tests:
        raise ValueError('no original acceptance tests found in selected directory')
    candidate = dict(original)
    edits = data.get('edits')
    if not isinstance(edits,list) or not 1 <= len(edits) <= 5:
        raise ValueError('expected one to five exact edits')
    for edit in edits:
        name = edit.get('file')
        safe_path(name)
        if name not in candidate or is_test(name):
            raise ValueError('original tests and absent files cannot be modified')
        old,new = edit.get('old'),edit.get('new')
        if not isinstance(old,str) or not isinstance(new,str) or not old or len(new.encode())>65536:
            raise ValueError('invalid exact edit')
        old,new = old.encode(),new.encode()
        if candidate[name].count(old)!=1: raise ValueError('old text must occur exactly once')
        candidate[name] = candidate[name].replace(old,new,1)
    return profile,directory,original,candidate,manifest_hash(tests)


def execute(root, host_root, profile, directory, files):
    image = os.environ.get(PROFILES[profile][0], '')
    if not re.fullmatch(r'(?:[A-Za-z0-9./:_-]+@)?sha256:[a-f0-9]{64}', image): raise ValueError('profile image is not configured with immutable digest')
    root.mkdir(mode=0o755)
    for name,content in files.items():
        target = root/name
        target.parent.mkdir(parents=True,exist_ok=True,mode=0o755)
        target.write_bytes(content)
        target.chmod(0o444)
    if not (root/directory).is_dir(): raise ValueError('validation directory not found')
    container = 'ccr-project-'+uuid.uuid4().hex
    command = ['docker','run','--rm','--pull=never','--log-driver=none','--name',container,
               '--network=none','--read-only','--cap-drop=ALL','--security-opt=no-new-privileges',
               '--user=65534:65534','--pids-limit=128','--memory=768m','--cpus=1',
               '--tmpfs=/tmp:rw,exec,nosuid,nodev,size=512m',
               '--mount',f'type=bind,src={host_root},dst=/project,readonly',
               '--workdir','/project'+('' if directory=='.' else '/'+directory),
               '--env=HOME=/tmp','--env=GOCACHE=/tmp/cache','--env=GOPATH=/tmp/go',
               '--env=GOMODCACHE=/opt/dependencies/gomodcache','--env=GOPROXY=off','--env=GOTOOLCHAIN=local','--env=CGO_ENABLED=0',
               '--env=PYTHONDONTWRITEBYTECODE=1',image,*PROFILES[profile][1]]
    captured = bytearray()
    process = subprocess.Popen(command,stdout=subprocess.PIPE,stderr=subprocess.STDOUT)
    def drain():
        while True:
            block=process.stdout.read(4096)
            if not block: return
            captured.extend(block[:max(0,24000-len(captured))])
    reader=threading.Thread(target=drain,daemon=True);reader.start()
    try:
        try: code=process.wait(timeout=300)
        except subprocess.TimeoutExpired:
            subprocess.run(['docker','rm','-f',container],capture_output=True,timeout=15)
            process.kill();process.wait();code=124
        reader.join(timeout=5)
    finally:
        subprocess.run(['docker','rm','-f',container],capture_output=True,timeout=15)
    output=captured.decode(errors='replace')
    count=0
    if profile=='go-test-v1':
        for line in output.splitlines():
            try:
                event=json.loads(line)
                if event.get('Action')=='pass' and event.get('Test'): count+=1
            except (ValueError,AttributeError): pass
    else:
        pattern=r'Ran (\d+) tests?' if profile=='python-unittest-v1' else r'# tests (\d+)'
        found=re.search(pattern,output)
        if found: count=int(found.group(1))
    return {'passed':code==0 and count>0,'exit_code':code,'test_count':count,'output':output,'image':image}


def validate(data, job, host_job):
    profile,directory,original,candidate,tests_hash=prepare(data)
    baseline=execute(job/'baseline',host_job+'/baseline',profile,directory,original)
    revised=execute(job/'candidate',host_job+'/candidate',profile,directory,candidate)
    return {'profile':profile, **revised, 'baseline':baseline,
            'source_sha256':manifest_hash(candidate), 'tests_sha256':tests_hash,
            'baseline_sha256':manifest_hash(original), 'tests_preserved':True}
