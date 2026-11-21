"""Narrow Docker broker. No caller-controlled image, command, tests or mount paths."""
import hashlib
import http.server
import json
import os
import pathlib
import shutil
import socketserver
import subprocess
import tempfile
import uuid
import threading
import project
import deployment

ROOT = pathlib.Path('/jobs')
HOST_ROOT = os.environ['VALIDATOR_HOST_ROOT']
IMAGE = os.environ['VALIDATOR_IMAGE']
SOCKET = '/socket/validator.sock'
TESTS = pathlib.Path('/validator/grade_test.go').read_bytes()


class Handler(http.server.BaseHTTPRequestHandler):
    def do_POST(self):
        if self.path not in ('/validate','/deploy'):
            self.send_error(404)
            return
        try:
            size = int(self.headers.get('Content-Length', '0'))
            if not 0 < size <= 30*1024*1024:
                raise ValueError('invalid request size')
            data = json.loads(self.rfile.read(size))
            if not isinstance(data,dict):
                raise ValueError('request must be an object')
            if self.path == '/deploy':
                self.reply(deployment.deploy(data))
                return
            if data.get('profile') in project.PROFILES:
                job = pathlib.Path(tempfile.mkdtemp(prefix='project-',dir=ROOT))
                job.chmod(0o755)
                try:
                    result = project.validate(data,job,HOST_ROOT+'/'+job.name)
                    self.reply(result)
                finally:
                    shutil.rmtree(job)
                return
            if data.get('profile') != 'arena-grade-v1' or not isinstance(data.get('source'), str):
                raise ValueError('unsupported validation profile')
            source = data['source'].encode()
            if len(source) > 65536:
                raise ValueError('source too large')
        except (ValueError, UnicodeError) as error:
            self.send_error(400, str(error))
            return
        directory = pathlib.Path(tempfile.mkdtemp(prefix='job-', dir=ROOT))
        name = 'ccr-proof-' + uuid.uuid4().hex
        try:
            directory.chmod(0o755)
            (directory / 'grade.go').write_bytes(source)
            (directory / 'grade_test.go').write_bytes(TESTS)
            for path in directory.iterdir():
                path.chmod(0o444)
            command = ['docker', 'run', '--rm', '--pull=never', '--log-driver=none', '--name', name,
                       '--network=none', '--read-only', '--cap-drop=ALL',
                       '--security-opt=no-new-privileges', '--user=65534:65534',
                       '--pids-limit=128', '--memory=512m', '--cpus=1',
                       '--tmpfs=/tmp:rw,exec,nosuid,nodev,size=384m',
                       '--mount', f'type=bind,src={HOST_ROOT}/{directory.name},dst=/proof,readonly',
                       '--env=GOCACHE=/tmp/cache', '--env=GOPATH=/tmp/go',
                       '--env=GOPROXY=off', '--env=GOTOOLCHAIN=local', '--env=CGO_ENABLED=0',
                       IMAGE, 'go', 'test', '-count=1', '-timeout=30s', '-v',
                       '/proof/grade.go', '/proof/grade_test.go']
            captured = bytearray()
            process = subprocess.Popen(command, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
            def drain():
                while True:
                    block = process.stdout.read(4096)
                    if not block:
                        return
                    captured.extend(block[:max(0, 16384-len(captured))])
            reader = threading.Thread(target=drain, daemon=True)
            reader.start()
            try:
                code = process.wait(timeout=180)
            except subprocess.TimeoutExpired:
                subprocess.run(['docker', 'rm', '-f', name], capture_output=True, timeout=15)
                process.kill()
                process.wait()
                code = 124
            reader.join(timeout=5)
            log = captured.decode(errors='replace')
            result = {'profile': 'arena-grade-v1', 'passed': code == 0,
                      'exit_code': code, 'source_sha256': hashlib.sha256(source).hexdigest(),
                      'tests_sha256': hashlib.sha256(TESTS).hexdigest(), 'image': IMAGE,
                      'output': log}
            self.reply(result)
        finally:
            subprocess.run(['docker', 'rm', '-f', name], capture_output=True, timeout=15)
            shutil.rmtree(directory)

    def log_message(self, fmt, *args):
        print(fmt % args, flush=True)

    def reply(self,result):
        payload=json.dumps(result).encode()
        self.send_response(200)
        self.send_header('Content-Type','application/json')
        self.send_header('Content-Length',str(len(payload)))
        self.end_headers()
        self.wfile.write(payload)


class Server(socketserver.UnixStreamServer):
    allow_reuse_address = True


ROOT.mkdir(exist_ok=True)
pathlib.Path('/socket').mkdir(exist_ok=True)
pathlib.Path(SOCKET).unlink(missing_ok=True)
with Server(SOCKET, Handler) as server:
    os.chmod(SOCKET, 0o666)
    server.serve_forever()
