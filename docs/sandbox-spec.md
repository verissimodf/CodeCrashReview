# CodeCrashReview - Sandbox Specification

## Overview

The sandbox system provides isolated, secure environments for code execution during matches. Each language runs in a dedicated Docker container with strict resource limits and security policies.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Sandbox Manager                          │
├─────────────────────────────────────────────────────────────┤
│  Language Pool (Pre-warmed containers)                      │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌───────┐ │
│  │  C++    │ │ Python  │ │  Java   │ │  JS/TS  │ │ Rust  │ │
│  │  Pool   │ │  Pool   │ │  Pool   │ │  Pool   │ │ Pool  │ │
│  └─────────┘ └─────────┘ └─────────┘ └─────────┘ └───────┘ │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐           │
│  │ Shell   │ │  HTML   │ │ Gherkin │ │  ...    │           │
│  │ Pool    │ │  Pool   │ │ Pool    │ │         │           │
│  └─────────┘ └─────────┘ └─────────┘ └─────────┘           │
└─────────────────────────────────────────────────────────────┘
```

## Container Specifications

### Base Images

| Language | Base Image | Size | Tools Included |
|----------|------------|------|----------------|
| C++ | `gcc:13-slim` | ~200MB | g++, clang-format, cppcheck, valgrind, gdb, cmake |
| Python | `python:3.12-slim` | ~150MB | bandit, pylint, black, isort, mypy, pytest, semgrep |
| Java | `eclipse-temurin:21-jdk-slim` | ~300MB | maven, gradle, spotbugs |
| JavaScript | `node:20-slim` | ~180MB | eslint, prettier, jest, typescript, semgrep |
| TypeScript | `node:20-slim` | ~180MB | tsc, ts-node, @types/node, jest, ts-jest |
| Rust | `rust:1.78-slim` | ~400MB | clippy, rustfmt, rust-analyzer, cargo-audit |
| Shell | `alpine:3.19` | ~50MB | bash, shellcheck, shfmt, bats |
| HTML | `node:20-slim` | ~180MB | tidy, htmlhint, prettier |
| Gherkin | `python:3.12-slim` | ~150MB | behave, pytest-bdd, gherkin-official |

### Resource Limits (Default)

| Resource | Limit | Configurable |
|----------|-------|--------------|
| CPU | 2 cores | `SANDBOX_CPU_LIMIT` |
| Memory | 2 GB | `SANDBOX_MEMORY_LIMIT` |
| PIDs | 100 | `SANDBOX_PIDS_LIMIT` |
| Timeout | 30s | Per-request |
| Network | Disabled | Always |
| Rootfs | Read-only | Where possible |

### Security Configuration

```yaml
# Docker HostConfig
NetworkMode: "codecrash-sandbox"  # Internal bridge, no external access
Resources:
  CPUQuota: 200000      # 2 cores (100000 = 1 core)
  CPUPeriod: 100000
  Memory: 2147483648    # 2 GB
  PidsLimit: 100
SecurityOpt:
  - "no-new-privileges:true"
  - "seccomp=/etc/docker/seccomp-default.json"  # Optional custom profile
ReadonlyRootfs: true
Tmpfs:
  /tmp: "rw,noexec,nosuid,size=100m"
  /workspace: "rw,noexec,nosuid,size=500m"
User: "1000:1000"  # Non-root user
CapDrop: ["ALL"]
```

## Execution Flow

### 1. Request Reception
```go
type ExecutionRequest struct {
    Files          []CodeFile
    Command        string
    Args           []string
    Stdin          string
    TimeoutSeconds int
    ResourceLimits ResourceLimits
}
```

### 2. Container Selection
- Check pool for available container of required language
- If pool empty, create new container (cold start ~3s)
- Reuse warm container (<500ms)

### 3. File Preparation
- Write files to `/workspace` in container
- Preserve directory structure
- Set appropriate permissions

### 4. Compilation (if needed)
| Language | Compile Command |
|----------|-----------------|
| C++ | `g++ -std=c++17 -O2 -pipe -static -s -o /tmp/main /workspace/main.cpp` |
| Java | `javac -d /tmp /workspace/Main.java` |
| TypeScript | `tsc /workspace/main.ts --outDir /tmp` |
| Rust | `rustc -O /workspace/main.rs -o /tmp/main` |
| Others | No compilation needed |

### 5. Execution
- Run compiled binary or interpreter
- Capture stdout, stderr
- Enforce timeout
- Monitor resource usage

### 6. Result Collection
```go
type ExecutionResult struct {
    ExitCode      int
    Stdout        string
    Stderr        string
    DurationMs    int64
    MemoryUsed    int64
    TimedOut      bool
    OOMKilled     bool
}
```

### 7. Cleanup
- Remove container (or return to pool)
- Clean temporary files
- Update pool metrics

## Language-Specific Details

### C++
```dockerfile
# Compile flags
CXXFLAGS="-std=c++17 -O2 -pipe -static -s -Wall -Wextra"
# Execution
/tmp/main
```

### Python
```bash
# Direct execution
python3 /workspace/main.py
# With coverage
coverage run -m pytest
coverage report
```

### Java
```bash
# Compile
javac -d /tmp /workspace/Main.java
# Run
java -cp /tmp Main
# With tests
mvn test  # or gradle test
```

### JavaScript/TypeScript
```bash
# JS
node /workspace/main.js
# TS
tsc /workspace/main.ts --outDir /tmp
node /tmp/main.js
# Tests
npm test  # or jest
```

### Rust
```bash
# Compile
rustc -O /workspace/main.rs -o /tmp/main
# Run
/tmp/main
# Tests
cargo test
```

### Shell
```bash
bash /workspace/main.sh
# Linting
shellcheck /workspace/main.sh
```

### HTML
```bash
# Validation
tidy -q /workspace/index.html
# Or serve
npx serve /workspace
```

### Gherkin
```bash
# Run BDD tests
behave /workspace/features/
# With pytest-bdd
pytest /workspace/tests/
```

## Pool Management

### Pre-warming Strategy
1. On startup, create `SANDBOX_POOL_SIZE_PER_LANGUAGE` containers per language
2. Keep containers running with `sleep infinity`
3. On request, grab container from pool
4. After execution, return to pool (reset state)

### Pool Configuration
```go
type PoolConfig struct {
    Language           string
    Image              string
    PoolSize           int        // Default: 3
    MaxPoolSize        int        // Default: 10
    ContainerTTL       time.Duration  // Default: 1 hour
    HealthCheckInterval time.Duration // Default: 30s
}
```

### Container Lifecycle
```
Create → Start → Warm (idle) → Busy (executing) → Warm (idle) → ... → Expire → Remove
```

### Health Checks
- Periodic `docker exec` to verify responsiveness
- Remove unhealthy containers
- Auto-replace failed containers

## API Interface

### Execute Code
```http
POST /api/sandbox/execute
Content-Type: application/json

{
  "files": [
    {"path": "main.cpp", "language": "cpp", "content": "#include <iostream>\nint main() { std::cout << \"Hello\"; }"}
  ],
  "command": "./main",
  "args": [],
  "timeout_seconds": 30
}
```

Response:
```json
{
  "exit_code": 0,
  "stdout": "Hello",
  "stderr": "",
  "duration_ms": 45,
  "memory_used": 2048000,
  "timed_out": false,
  "oom_killed": false
}
```

### Supported Operations
| Operation | Endpoint | Description |
|-----------|----------|-------------|
| Execute | `POST /api/sandbox/execute` | Run code |
| Compile | `POST /api/sandbox/compile` | Compile only |
| Test | `POST /api/sandbox/test` | Run test suite |
| Lint | `POST /api/sandbox/lint` | Run linter |

## Monitoring & Metrics

### Key Metrics
| Metric | Description | Target |
|--------|-------------|--------|
| Cold Start Time | Time to create new container | <3s |
| Warm Execution | Time to execute in warm container | <500ms |
| Success Rate | Successful executions / total | >99.9% |
| Memory Usage | Peak memory per execution | <512MB |
| CPU Usage | CPU time per execution | <2s |
| Pool Utilization | Active / total containers | <80% |

### Prometheus Metrics
```prometheus
# Sandbox execution duration
codecrash_sandbox_execution_duration_seconds{language="python",status="success"}

# Container pool size
codecrash_sandbox_pool_size{language="python",state="warm"}

# Execution errors
codecrash_sandbox_errors_total{language="python",error="timeout"}

# Resource usage
codecrash_sandbox_memory_bytes{language="python"}
codecrash_sandbox_cpu_seconds{language="python"}
```

## Error Handling

### Common Errors
| Error | Cause | Resolution |
|-------|-------|------------|
| `TIMEOUT` | Execution exceeded limit | Increase timeout or optimize code |
| `OOM_KILLED` | Memory limit exceeded | Increase memory limit or optimize |
| `COMPILATION_ERROR` | Syntax/type errors | Fix code before execution |
| `CONTAINER_CRASH` | Container died unexpectedly | Check logs, restart pool |
| `POOL_EXHAUSTED` | No available containers | Increase pool size |

### Retry Logic
```go
func executeWithRetry(ctx context.Context, req *ExecutionRequest, maxRetries int) (*ExecutionResult, error) {
    for i := 0; i <= maxRetries; i++ {
        result, err := execute(ctx, req)
        if err == nil || !isRetryable(err) {
            return result, err
        }
        // Exponential backoff
        time.Sleep(time.Duration(i+1) * 500 * time.Millisecond)
    }
    return nil, fmt.Errorf("max retries exceeded")
}
```

## Customization

### Custom Seccomp Profile
```json
{
  "defaultAction": "SCMP_ACT_ERRNO",
  "architectures": ["SCMP_ARCH_X86_64", "SCMP_ARCH_X86"],
  "syscalls": [
    {"names": ["read", "write", "open", "close", "exit", "exit_group"], "action": "SCMP_ACT_ALLOW"},
    {"names": ["mmap", "mprotect", "munmap"], "action": "SCMP_ACT_ALLOW"},
    {"names": ["brk", "rt_sigaction", "rt_sigprocmask"], "action": "SCMP_ACT_ALLOW"}
  ]
}
```

### Custom Resource Limits
```yaml
# Per-language overrides in config
sandbox_limits:
  cpp:
    cpu_limit: 4.0
    memory_limit: "4g"
    timeout: 60
  python:
    cpu_limit: 2.0
    memory_limit: "2g"
    timeout: 30
  rust:
    cpu_limit: 2.0
    memory_limit: "3g"
    timeout: 60
```

## Testing

### Unit Tests
- Mock Docker client for fast tests
- Test resource limit enforcement
- Test timeout handling
- Test cleanup on errors

### Integration Tests
- Real container execution
- Multi-language test suite
- Concurrent execution tests
- Resource limit verification

### Benchmark Suite
```bash
# Run benchmarks
make sandbox-benchmark

# Outputs:
# - Cold start times per language
# - Warm execution times
# - Memory overhead
# - Concurrent execution capacity
```

## Future Enhancements

1. **Firecracker MicroVMs** - Stronger isolation
2. **gVisor** - Kernel-level sandboxing
3. **WASM Runtimes** - Faster cold starts for some languages
4. **Distributed Sandboxes** - Run on separate nodes
5. **GPU Support** - For ML/CUDA challenges
5. **Persistent Volumes** - For multi-file projects
6. **Live Debugging** - Attach debugger to running container