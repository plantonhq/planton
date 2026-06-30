# =============================================================================
# IaC Runner - Terraform Variant
# =============================================================================
# A lightweight base image for running OpenTofu/Terraform programs.
# No Go or Pulumi - just the essentials for Terraform-based IaC.
#
# Usage:
#   FROM ghcr.io/plantonhq/planton/base-images/iac-runner:terraform-v0.0.1
#
# Build: Triggered by .github/workflows/iac-runner-terraform.yml
# =============================================================================

FROM debian:bullseye

ARG OPENTOFU_VERSION=1.9.1

# Set environment
ENV DEBIAN_FRONTEND=noninteractive

# OCI labels
LABEL org.opencontainers.image.source="https://github.com/plantonhq/planton"
LABEL org.opencontainers.image.description="IaC Runner Terraform variant with OpenTofu"
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

# ---- OpenTofu -------------------------------------------------------------
RUN mkdir -p $HOME/.terraform.d/plugin-cache && \
    wget -q "https://github.com/opentofu/opentofu/releases/download/v${OPENTOFU_VERSION}/tofu_${OPENTOFU_VERSION}_linux_amd64.tar.gz" && \
    tar -xf "tofu_${OPENTOFU_VERSION}_linux_amd64.tar.gz" && \
    mv tofu /usr/local/bin/ && \
    echo 'plugin_cache_dir = "$HOME/.terraform.d/plugin-cache"' > $HOME/.tofurc && \
    rm "tofu_${OPENTOFU_VERSION}_linux_amd64.tar.gz"

# Verify installation
RUN tofu version && \
    echo "" && \
    echo "IaC Runner Terraform variant ready"

