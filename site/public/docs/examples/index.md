---
title: "Examples"
description: "Copy-paste-ready manifest examples for deploying infrastructure with Planton across multiple cloud providers"
icon: "code"
order: 50
---

# Examples

This section provides ready-to-use manifest examples for deploying infrastructure with Planton. Each example is a complete, valid manifest that you can copy, customize, and deploy.

## How to Use These Examples

Every example follows the same workflow:

1. **Copy** the manifest YAML into a file (e.g., `my-resource.yaml`)
2. **Customize** the fields for your environment — resource names, regions, credentials, and configuration values
3. **Set provisioner labels** — choose `pulumi` or `tofu` as your provisioner and configure the corresponding state backend labels
4. **Deploy** with the Planton CLI:

```bash
# Preview what will be created
planton plan -f my-resource.yaml

# Deploy the resource
planton apply -f my-resource.yaml
```

## In This Section

- **[Manifest Gallery](./manifest-gallery)** — Curated manifest examples organized by cloud provider, covering storage, databases, networking, Kubernetes clusters, and workloads

## Related Resources

- **[Tutorials](/docs/tutorials)** — Guided walkthroughs that take you step-by-step through deploying and managing resources
- **[Catalog](/docs/catalog)** — Complete reference for all 360+ deployment components with full field documentation
- **[Manifests Guide](/docs/guides/manifests)** — How to write and structure Planton manifests
- **[CLI Reference](/docs/cli/cli-reference)** — All commands and flags for the Planton CLI
