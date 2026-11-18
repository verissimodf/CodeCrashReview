# Python Sandbox
FROM python:3.12-slim AS base

# Install security tools and utilities
RUN apt-get update && apt-get install -y --no-install-recommends \
    bandit \
    pylint \
    black \
    isort \
    mypy \
    pytest \
    pytest-cov \
    hypothesis \
    && rm -rf /var/lib/apt/lists/*

# Install Python security packages
RUN pip install --no-cache-dir \
    semgrep \
    safety \
    pip-audit

# Create non-root user
RUN useradd -m -u 1000 -s /bin/bash sandbox && \
    mkdir -p /workspace && \
    chown sandbox:sandbox /workspace

USER sandbox
WORKDIR /workspace

CMD ["sleep", "infinity"]