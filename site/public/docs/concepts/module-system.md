---
title: "Module System"
description: "How OpenMCF resolves, downloads, caches, and versions IaC modules -- from the staging repository to workspace isolation and version pinning"
icon: "package"
order: 45
---

# Module System

When you run a deployment command, the CLI needs to find the correct IaC module for the component you are deploying. OpenMCF's module system handles this automatically -- cloning the module repository, caching it locally, creating isolated workspace copies for each deployment, and supporting version pinning for reproducible builds.

## Module Resolution Chain

The CLI resolves module paths through a priority chain. The first method that succeeds is used:

### 1. Direct Path (--module-dir)

If you provide `--module-dir` pointing to a directory that contains a valid module (a `Pulumi.yaml` for Pulumi, or `.tf` files for Terraform), the CLI uses that directory directly. No cloning, no caching.

```bash
openmcf pulumi up -f postgres.yaml \
  --module-dir ./my-custom-module \
  --stack my-org/my-project/prod
```

This is primarily used for local module development -- when you are editing a module and want to test your changes without publishing a new version.

### 2. Local Module (--local-module)

The `--local-module` flag tells the CLI to use the local OpenMCF repository checkout to derive the module directory. This is useful when developing within the OpenMCF repository itself.

### 3. Pre-Built Binary (Pulumi only)

For Pulumi modules, the CLI attempts to download a pre-built binary for the component. These binaries are published as GitHub releases with version tags. This is the fastest path for Pulumi deployments since it avoids cloning the entire repository.

### 4. Zip Download (OpenTofu/Terraform only)

For Terraform modules, the CLI attempts to download a zip archive of the module from GitHub releases. Like the binary approach, this avoids a full repository clone.

### 5. Staging Repository (Fallback)

If none of the above methods work, the CLI falls back to the staging system.

## The Staging System

The staging repository is a local clone of the OpenMCF repository at `~/.openmcf/staging/`. It serves as the source for IaC modules when pre-built artifacts are not available.

### How Staging Works

1. **First run**: The CLI clones the OpenMCF repository to `~/.openmcf/staging/`
2. **Subsequent runs**: The CLI verifies the staging repo exists and is at the correct version
3. **Version mismatch**: If the CLI version does not match the staging version, the CLI fetches and checks out the correct version

The staging version is tracked in `~/.openmcf/staging/.version`.

### Workspace Isolation

The staging repository is never used directly for deployments. Instead, the CLI copies the staging repo into an isolated workspace directory for each deployment:

- **Pulumi workspaces**: `~/.openmcf/pulumi/staging-workspaces/{stack-fqdn}/`
- **OpenTofu workspaces**: `~/.openmcf/tofu/`

Each deployment operates on its own copy of the module, which means concurrent deployments of different components (or different versions of the same component) never interfere with each other.

### Module Path Convention

Within the staging workspace, the CLI locates modules using a deterministic path:

```
{workspace}/apis/org/openmcf/provider/{provider}/{kind}/v1/iac/{engine}/
```

For example, the Pulumi module for `KubernetesPostgres`:

```
~/.openmcf/pulumi/staging-workspaces/my-org-my-project-prod/
  apis/org/openmcf/provider/kubernetes/kubernetespostgres/v1/iac/pulumi/
```

The provider name is derived from the `CloudResourceKind` enum metadata. The kind name is the lowercase version of the enum entry. The engine is either `pulumi` or `tf`.

## Version Management

### Checking the Current Version

```bash
openmcf modules-version
```

This shows the version of the IaC modules currently checked out in the staging area.

### Updating to Latest

```bash
openmcf pull
```

This fetches the latest changes from the upstream OpenMCF repository and pulls them into the staging area. If the staging repo was on a specific version tag, it first checks out the `main` branch, pulls, and then optionally restores the previous version.

### Switching Versions

```bash
openmcf checkout v0.3.9
```

This checks out a specific version tag, branch name, or commit SHA in the staging area. The special value `latest` checks out the most recent release tag:

```bash
openmcf checkout latest
```

### Per-Deployment Version Override

You can override the module version for a specific deployment without changing the staging area:

```bash
openmcf pulumi up -f postgres.yaml \
  --module-version v0.3.8 \
  --stack my-org/my-project/prod
```

The `--module-version` flag checks out the specified version in the workspace copy only. The staging area remains untouched, so other deployments continue using the staging version.

## Workspace Cleanup

By default, the CLI cleans up workspace copies after each deployment. If you want to inspect the workspace after a run (for debugging), use the `--no-cleanup` flag:

```bash
openmcf pulumi up -f postgres.yaml \
  --no-cleanup \
  --stack my-org/my-project/prod
```

The workspace copy will remain at its path after the command completes.

## Local Module Development

When developing or customizing a module, use `--module-dir` to point at your local copy:

```bash
# Edit the module locally
cd ~/my-openmcf-fork/apis/org/openmcf/provider/kubernetes/kubernetespostgres/v1/iac/pulumi/

# Test with the local module
openmcf pulumi up -f postgres.yaml \
  --module-dir . \
  --stack my-org/my-project/dev
```

This bypasses the entire resolution chain and uses the local directory directly. Changes you make to the module files are immediately reflected in the next deployment.

## What's Next

- **[Dual IaC Engines](dual-iac-engines)** -- The Pulumi and OpenTofu/Terraform modules that the module system resolves
- **[State Management](state-management)** -- How deployment state is stored and managed
- **[Deployment Components](deployment-components)** -- The anatomy of a component including its IaC modules
