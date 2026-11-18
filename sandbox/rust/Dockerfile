# Rust Sandbox
FROM rust:1.78-slim AS base

# Install tools
RUN apt-get update && apt-get install -y --no-install-recommends \
    clang \
    lldb \
    && rm -rf /var/lib/apt/lists/*

# Install Rust tools
RUN rustup component add clippy rustfmt rust-analyzer && \
    cargo install cargo-audit cargo-outdated cargo-tree

# Create non-root user
RUN useradd -m -u 1000 -s /bin/bash sandbox && \
    mkdir -p /workspace && \
    chown sandbox:sandbox /workspace

USER sandbox
WORKDIR /workspace

CMD ["sleep", "infinity"]