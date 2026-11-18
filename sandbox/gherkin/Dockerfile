# Gherkin/BDD Sandbox
FROM python:3.12-slim AS base

# Install Gherkin tools
RUN apt-get update && apt-get install -y --no-install-recommends \
    && rm -rf /var/lib/apt/lists/*

# Install Python BDD tools
RUN pip install --no-cache-dir \
    behave \
    pytest-bdd \
    gherkin-official \
    pytest \
    pytest-cov

# Create non-root user
RUN useradd -m -u 1000 -s /bin/bash sandbox && \
    mkdir -p /workspace && \
    chown sandbox:sandbox /workspace

USER sandbox
WORKDIR /workspace

CMD ["sleep", "infinity"]