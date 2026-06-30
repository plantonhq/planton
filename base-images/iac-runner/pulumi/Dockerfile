# =============================================================================
# Pulumi Runner - Pre-Built Binaries Variant
# =============================================================================
# A base image for running Pulumi programs with pre-built binaries for all
# platform deployment components. Includes Go toolchain for organization
# (custom) modules that require compilation.
#
# Usage:
#   FROM ghcr.io/plantonhq/planton/base-images/pulumi-runner:v0.0.100
#
# Build: Triggered by .github/workflows/pulumi-runner.yml
#
# Binary Structure:
#   /opt/planton/pulumi-binaries/{provider}/{component}
#   e.g., /opt/planton/pulumi-binaries/aws/awsekscluster
#         /opt/planton/pulumi-binaries/gcp/gcpgkecluster
#         /opt/planton/pulumi-binaries/kubernetes/kubernetesargocd
# =============================================================================

FROM debian:bullseye

ARG PULUMI_VERSION=v3.202.0
ARG GOLANG_VERSION=1.25.0

# Set environment
ENV DEBIAN_FRONTEND=noninteractive \
    PATH="/usr/local/go/bin:/root/go/bin:${PATH}" \
    GOMODCACHE=/var/cache/go-mod \
    GOCACHE=/var/cache/go-build \
    PLANTON_BINARIES_DIR=/opt/planton/pulumi-binaries

# OCI labels
LABEL org.opencontainers.image.source="https://github.com/plantonhq/planton"
LABEL org.opencontainers.image.description="Pulumi Runner with pre-built binaries for all deployment components"
LABEL org.opencontainers.image.licenses="Apache-2.0"

EXPOSE 8080

# ---- Base packages --------------------------------------------------------
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        ca-certificates \
        debian-archive-keyring \
        git \
        wget && \
    rm -rf /var/lib/apt/lists/*

# ---- Go -------------------------------------------------------------------
# Required for organization (custom) modules that need compilation
RUN wget -q "https://go.dev/dl/go${GOLANG_VERSION}.linux-amd64.tar.gz" -O go.tar.gz && \
    tar -xf go.tar.gz -C /usr/local && \
    rm go.tar.gz && \
    mkdir -p ${GOMODCACHE} ${GOCACHE}

# ---- Pulumi ---------------------------------------------------------------
RUN wget -q "https://get.pulumi.com/releases/sdk/pulumi-${PULUMI_VERSION}-linux-x64.tar.gz" && \
    tar -xf "pulumi-${PULUMI_VERSION}-linux-x64.tar.gz" && \
    mv pulumi/* /usr/local/bin/ && \
    rm -rf pulumi "pulumi-${PULUMI_VERSION}-linux-x64.tar.gz"

# =============================================================================
# PRE-BUILT DEPLOYMENT COMPONENT BINARIES
# =============================================================================
# These binaries are compiled during the GitHub Actions workflow from the
# source code of each deployment component. This eliminates:
#   - Go module download time at runtime
#   - Go compilation time at runtime
#   - Need for pre-warmed Go caches
#
# The IaC Runner checks for a binary at ${PLANTON_BINARIES_DIR}/{provider}/{component}
# before falling back to Git-based source execution.

# Create binaries directory
RUN mkdir -p ${PLANTON_BINARIES_DIR}

# Copy pre-built binaries from GitHub Actions workflow
# Structure: binaries/{provider}/{component}
COPY binaries/ ${PLANTON_BINARIES_DIR}/

# Verify binaries and display summary
RUN echo "=== Pre-Built Pulumi Binaries ===" && \
    echo "Total binaries: $(find ${PLANTON_BINARIES_DIR} -type f -executable 2>/dev/null | wc -l)" && \
    echo "" && \
    echo "Per-provider counts:" && \
    for provider in ${PLANTON_BINARIES_DIR}/*/; do \
        if [ -d "$provider" ]; then \
            count=$(find "$provider" -type f -executable 2>/dev/null | wc -l); \
            echo "  $(basename "$provider"): $count"; \
        fi; \
    done && \
    echo "" && \
    echo "Total size: $(du -sh ${PLANTON_BINARIES_DIR} 2>/dev/null | cut -f1)" && \
    echo "" && \
    echo "Go version: $(go version)" && \
    echo "Pulumi version: $(pulumi version)" && \
    echo "" && \
    echo "Pulumi Runner ready with pre-built binaries"
