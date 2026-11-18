# TypeScript Sandbox (extends JavaScript)
FROM node:20-slim AS base

# Install tools
RUN apt-get update && apt-get install -y --no-install-recommends \
    && rm -rf /var/lib/apt/lists/*

# Install global npm packages for TypeScript
RUN npm install -g \
    typescript \
    ts-node \
    @types/node \
    eslint \
    @typescript-eslint/eslint-plugin \
    @typescript-eslint/parser \
    prettier \
    jest \
    ts-jest \
    @types/jest \
    semgrep

# Create non-root user
RUN useradd -m -u 1000 -s /bin/bash sandbox && \
    mkdir -p /workspace && \
    chown sandbox:sandbox /workspace

USER sandbox
WORKDIR /workspace

CMD ["sleep", "infinity"]