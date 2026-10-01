# =============================================================================
# Common CI Toolchain Image
# =============================================================================
# Prebuilt image that bakes the entire toolchain the CI toolchain jobs need
# (e2e-aws, e2e-auth0, e2e-kubernetes, lint.terraform-modules) so those jobs
# stop reinstalling tools at runtime.
#
# Published to: ghcr.io/plantonhq/planton/ci-toolchain
# Built by:     .github/workflows/build.ci-image.yaml (rebuilt on tools/ci/docker changes)
#
# Version policy: tools are installed at their LATEST at build time (no pins),
# mirroring the previous per-run behavior of the workflows this replaces. The Go
# toolchain is the one exception - it is pinned to the EXACT base image tag equal
# to the `go` directive in go.mod; bump it here when go.mod bumps (that Dockerfile
# edit is what triggers a rebuild). A minor-only tag (golang:1.26-bookworm) froze
# at whatever patch was current at the last build and fell behind go.mod, and a
# GOTOOLCHAIN=local job then refuses to run.
#
# Note on preinstalled tools: a `container:` job does NOT inherit the tools that
# GitHub's ubuntu-latest runner ships with (aws, gh, jq, zip, git, docker-cli),
# so they are all installed explicitly below.
# =============================================================================

FROM golang:1.26.6-bookworm

ENV PATH=/root/.pulumi/bin:${PATH}

# -----------------------------------------------------------------------------
# Base OS packages + gh CLI (from GitHub's apt repo)
# -----------------------------------------------------------------------------
RUN set -eux; \
    apt-get update; \
    apt-get install -y --no-install-recommends \
        ca-certificates \
        curl \
        git \
        gnupg \
        jq \
        unzip \
        zip \
        docker.io; \
    mkdir -p -m 755 /etc/apt/keyrings; \
    curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg \
        | gpg --dearmor -o /etc/apt/keyrings/githubcli-archive-keyring.gpg; \
    chmod go+r /etc/apt/keyrings/githubcli-archive-keyring.gpg; \
    echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" \
        > /etc/apt/sources.list.d/github-cli.list; \
    apt-get update; \
    apt-get install -y --no-install-recommends gh; \
    rm -rf /var/lib/apt/lists/*

# -----------------------------------------------------------------------------
# AWS CLI v2 (latest) - used for Cloudflare R2 uploads / AWS OIDC identity checks.
# The unversioned installer URL always resolves to the current v2 release.
# -----------------------------------------------------------------------------
RUN set -eux; \
    curl -fsSL "https://awscli.amazonaws.com/awscli-exe-linux-$(uname -m).zip" -o /tmp/awscliv2.zip; \
    unzip -q /tmp/awscliv2.zip -d /tmp; \
    /tmp/aws/install; \
    rm -rf /tmp/aws /tmp/awscliv2.zip; \
    aws --version

# -----------------------------------------------------------------------------
# kubectl (latest stable) - stable.txt is a plain CDN pointer, not rate-limited.
# -----------------------------------------------------------------------------
RUN set -eux; \
    arch="$(dpkg --print-architecture)"; \
    kver="$(curl -fsSL https://dl.k8s.io/release/stable.txt)"; \
    curl -fsSLo /usr/local/bin/kubectl "https://dl.k8s.io/release/${kver}/bin/linux/${arch}/kubectl"; \
    chmod +x /usr/local/bin/kubectl; \
    kubectl version --client

# -----------------------------------------------------------------------------
# gotestsum (latest) + kind (pinned).
# kind stays pinned because its version is coupled to the Kubernetes node-image
# versions the e2e framework boots; this matches the original workflow's @v0.27.0.
# -----------------------------------------------------------------------------
ENV KIND_VERSION=v0.27.0
RUN set -eux; \
    go install gotest.tools/gotestsum@latest; \
    go install "sigs.k8s.io/kind@${KIND_VERSION}"; \
    cp "$(go env GOPATH)/bin/gotestsum" /usr/local/bin/gotestsum; \
    cp "$(go env GOPATH)/bin/kind" /usr/local/bin/kind; \
    gotestsum --version; \
    kind --version

# -----------------------------------------------------------------------------
# Pulumi (latest, installed to /root/.pulumi/bin, already on PATH above)
# -----------------------------------------------------------------------------
RUN set -eux; \
    curl -fsSL https://get.pulumi.com | sh; \
    pulumi version

# -----------------------------------------------------------------------------
# OpenTofu (latest, standalone install to /usr/local/bin).
# The installer's own "latest" lookup hits the rate-limited api.github.com REST
# API, which fails under an exhausted anonymous quota. We instead resolve the
# latest version from the github.com releases/latest redirect (plain web, not the
# metered API) and pass it explicitly - still "always latest", but reliable.
# -----------------------------------------------------------------------------
RUN set -eux; \
    latest_url="$(curl -fsSL -o /dev/null -w '%{url_effective}' https://github.com/opentofu/opentofu/releases/latest)"; \
    tofu_version="${latest_url##*/tag/v}"; \
    echo "Latest OpenTofu version: ${tofu_version}"; \
    curl -fsSL https://get.opentofu.org/install-opentofu.sh -o /tmp/install-opentofu.sh; \
    chmod +x /tmp/install-opentofu.sh; \
    /tmp/install-opentofu.sh --install-method standalone --skip-verify --opentofu-version "${tofu_version}"; \
    rm /tmp/install-opentofu.sh; \
    tofu version

WORKDIR /workspace
