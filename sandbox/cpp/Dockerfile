# C++ Sandbox
FROM gcc:13-slim AS base

# Install security tools and utilities
RUN apt-get update && apt-get install -y --no-install-recommends \
    clang-format \
    cppcheck \
    valgrind \
    gdb \
    cmake \
    make \
    && rm -rf /var/lib/apt/lists/*

# Create non-root user
RUN useradd -m -u 1000 -s /bin/bash sandbox && \
    mkdir -p /workspace && \
    chown sandbox:sandbox /workspace

USER sandbox
WORKDIR /workspace

# Default command keeps container alive
CMD ["sleep", "infinity"]