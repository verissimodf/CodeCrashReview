# HTML Sandbox
FROM node:20-slim AS base

# Install HTML tools
RUN apt-get update && apt-get install -y --no-install-recommends \
    tidy \
    && rm -rf /var/lib/apt/lists/*

# Install Node.js tools for HTML
RUN npm install -g \
    htmlhint \
    prettier \
    @html-eslint/parser \
    @html-eslint/eslint-plugin

# Create non-root user
RUN useradd -m -u 1000 -s /bin/bash sandbox && \
    mkdir -p /workspace && \
    chown sandbox:sandbox /workspace

USER sandbox
WORKDIR /workspace

CMD ["sleep", "infinity"]