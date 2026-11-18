# Java Sandbox
FROM eclipse-temurin:21-jdk-slim AS base

# Install tools
RUN apt-get update && apt-get install -y --no-install-recommends \
    maven \
    gradle \
    && rm -rf /var/lib/apt/lists/*

# Install security tools
RUN mkdir -p /opt/tools && \
    cd /opt/tools && \
    wget -q https://github.com/spotbugs/spotbugs/releases/download/4.8.6/spotbugs-4.8.6.tgz && \
    tar -xzf spotbugs-4.8.6.tgz && \
    rm spotbugs-4.8.6.tgz

ENV PATH="/opt/tools/spotbugs-4.8.6/bin:${PATH}"

# Create non-root user
RUN useradd -m -u 1000 -s /bin/bash sandbox && \
    mkdir -p /workspace && \
    chown sandbox:sandbox /workspace

USER sandbox
WORKDIR /workspace

CMD ["sleep", "infinity"]