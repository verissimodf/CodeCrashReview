# Shell Script Sandbox
FROM alpine:3.19 AS base

# Install shell tools
RUN apk add --no-cache \
    bash \
    shellcheck \
    shfmt \
    bats \
    coreutils \
    findutils \
    grep \
    sed \
    awk

# Create non-root user
RUN adduser -D -u 1000 -s /bin/bash sandbox && \
    mkdir -p /workspace && \
    chown sandbox:sandbox /workspace

USER sandbox
WORKDIR /workspace

CMD ["sleep", "infinity"]