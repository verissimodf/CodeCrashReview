# JavaScript/Node.js Sandbox
FROM node:20-slim AS base

# Install tools
RUN apt-get update && apt-get install -y --no-install-recommends \
    && rm -rf /var/lib/apt/lists/*

# Install global npm packages
RUN npm install -g \
    eslint \
    prettier \
    jest \
    @types/jest \
    typescript \
    ts-node \
    semgrep

# Create non-root user
RUN useradd -m -u 1000 -s /bin/bash sandbox && \
    mkdir -p /workspace && \
    chown sandbox:sandbox /workspace

USER sandbox
WORKDIR /workspace

CMD ["sleep", "infinity"]