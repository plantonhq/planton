---
title: "Guides"
description: "In-depth guides for using OpenMCF - manifests, credentials, kustomize, and advanced techniques"
icon: "book"
order: 20
---

# OpenMCF Guides

Comprehensive guides to help you master OpenMCF.

---

## Core Guides

### [Manifest Structure](/docs/guides/manifests)

Learn everything about writing OpenMCF manifests.

**Topics covered**:
- What manifests are and why they matter
- Anatomy of a manifest (apiVersion, kind, metadata, spec, status)
- Validation and error detection
- Default values system
- Best practices for writing maintainable manifests

**Read this if**: You're new to OpenMCF or want to understand manifests deeply.

---

### [Credentials Management](/docs/guides/credentials)

Complete guide to providing cloud provider credentials securely.

**Topics covered**:
- Three ways to provide credentials (environment variables, files, embedded)
- Provider-specific guides (AWS, GCP, Azure, Cloudflare, Kubernetes, etc.)
- Security best practices
- CI/CD credential injection
- Troubleshooting authentication failures

**Read this if**: You're setting up OpenMCF for the first time or deploying to a new cloud provider.

---

### [Kustomize Integration](/docs/guides/kustomize)

Use Kustomize for managing multi-environment deployments.

**Topics covered**:
- What Kustomize is and why use it
- Directory structure and overlays
- Creating base manifests and environment patches
- Common patterns (environment-specific resources, labels, images)
- Deploying with `--kustomize-dir` and `--overlay` flags

**Read this if**: You're managing multiple environments (dev/staging/prod) with similar infrastructure.

---

### [State Backends](/docs/guides/state-backends)

Configure state storage for Pulumi, OpenTofu, and Terraform.

**Topics covered**:
- What state backends are and why they matter
- Pulumi state management (stack labels, Pulumi Cloud, local)
- OpenTofu/Terraform backends via manifest labels
- Backend types (S3, GCS, Azure Storage, local)
- Provisioner-specific labels (`terraform.*` vs `tofu.*`)
- Best practices for production state management

**Read this if**: You're deploying to production and need to configure shared state storage.

---

### [Advanced Usage](/docs/guides/advanced-usage)

Master advanced OpenMCF techniques.

**Topics covered**:
- Runtime value overrides with `--set`
- Loading manifests from URLs
- The `validate` and `load-manifest` commands
- Module directory overrides for custom modules
- Combining techniques (kustomize + --set, URL + overrides)
- Power user workflows and scripting

**Read this if**: You want to unlock advanced features and build sophisticated deployment workflows.

---

## Learning Path

### For Beginners

1. Start with [Getting Started](/docs/getting-started) - Install CLI and deploy your first resource
2. Read [Manifest Structure](/docs/guides/manifests) - Understand how to write manifests
3. Follow [Credentials Guide](/docs/guides/credentials) - Set up cloud credentials
4. Try [Pulumi Commands](/docs/cli/pulumi-commands) or [OpenTofu Commands](/docs/cli/tofu-commands) - Learn deployment commands

### For Intermediate Users

1. Review [Kustomize Integration](/docs/guides/kustomize) - Set up multi-environment deployments
2. Configure [State Backends](/docs/guides/state-backends) - Set up shared state storage
3. Explore [Advanced Usage](/docs/guides/advanced-usage) - Learn --set flags and URL manifests
4. Browse [Catalog](/docs/catalog) - Discover available deployment components

### For Advanced Users

1. Study component-specific documentation in [Catalog](/docs/catalog)
2. Fork and customize IaC modules
3. Build automation scripts and CI/CD pipelines
4. Contribute back to the project

---

## Quick Reference

### Essential Commands

```bash
# Validate manifest
openmcf validate -f resource.yaml

# Deploy with Pulumi
openmcf pulumi up -f resource.yaml

# Deploy with OpenTofu
openmcf tofu apply -f resource.yaml

# View effective manifest (with defaults)
openmcf load-manifest -f resource.yaml
```

### Common Flags

```bash
-f <file>                    # Manifest file or URL
--kustomize-dir <dir> --overlay <env> # Use kustomize
--set key=value                      # Override values
--module-dir <path>                  # Custom module location
--yes / --auto-approve               # Skip confirmation
```

---

## Related Documentation

- [CLI Reference](/docs/cli/cli-reference) - Complete CLI command reference
- [State Backends](/docs/guides/state-backends) - Configure state storage
- [Deployment Component Catalog](/docs/catalog) - Browse available components

---

## Getting Help

**Questions?** [GitHub Discussions](https://github.com/plantonhq/openmcf/discussions)

**Found a bug?** [Open an issue](https://github.com/plantonhq/openmcf/issues)

**Want to contribute?** Pull requests welcome!

