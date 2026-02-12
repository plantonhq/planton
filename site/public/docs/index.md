---
title: "Documentation"
description: "Comprehensive guides for OpenMCF - the open-source multi-cloud infrastructure framework"
icon: "📚"
order: 1
---

# Welcome to OpenMCF Documentation

OpenMCF is an open-source multi-cloud infrastructure framework that lets you author KRM-style YAML manifests once, validate them with Protobuf + Buf, and deploy with Pulumi or OpenTofu.

## Getting Started

New to OpenMCF? Start here:

- Install the CLI via Homebrew
- Validate your first manifest
- Deploy to your cloud provider or Kubernetes cluster

## CLI Reference

Master the OpenMCF command-line interface:

- **[Unified Commands](/docs/cli/unified-commands)** - kubectl-style `apply` and `destroy` commands
- **[Complete CLI Reference](/docs/cli/cli-reference)** - All commands, flags, and options
- **[Pulumi Commands](/docs/cli/pulumi-commands)** - Deploy with Pulumi (init, preview, up, refresh, destroy)
- **[OpenTofu Commands](/docs/cli/tofu-commands)** - Deploy with OpenTofu (init, plan, apply, refresh, destroy)

## Guides

In-depth guides for common scenarios:

- **[Manifest Structure](/docs/guides/manifests)** - Understanding and writing manifests
- **[Credentials Management](/docs/guides/credentials)** - Setting up cloud provider authentication
- **[Kustomize Integration](/docs/guides/kustomize)** - Multi-environment deployments
- **[Advanced Usage](/docs/guides/advanced-usage)** - Runtime overrides, URL manifests, and power techniques

## Deployment Components

Browse deployment components by cloud provider in the [Catalog](/docs/catalog):

<div class="grid grid-cols-1 md:grid-cols-2 gap-4 my-6">
  <a href="/docs/catalog/aws" class="flex items-center gap-3 p-4 rounded-lg border border-purple-900/30 bg-slate-900/30 hover:bg-slate-800/50 transition-colors">
    <img src="/images/providers/aws.svg" alt="AWS" class="w-8 h-8 object-contain" />
    <div>
      <div class="font-semibold text-white">AWS</div>
      <div class="text-sm text-slate-400">22 components</div>
    </div>
  </a>
  <a href="/docs/catalog/gcp" class="flex items-center gap-3 p-4 rounded-lg border border-purple-900/30 bg-slate-900/30 hover:bg-slate-800/50 transition-colors">
    <img src="/images/providers/gcp.svg" alt="GCP" class="w-8 h-8 object-contain" />
    <div>
      <div class="font-semibold text-white">GCP</div>
      <div class="text-sm text-slate-400">17 components</div>
    </div>
  </a>
  <a href="/docs/catalog/azure" class="flex items-center gap-3 p-4 rounded-lg border border-purple-900/30 bg-slate-900/30 hover:bg-slate-800/50 transition-colors">
    <img src="/images/providers/azure.svg" alt="Azure" class="w-8 h-8 object-contain" />
    <div>
      <div class="font-semibold text-white">Azure</div>
      <div class="text-sm text-slate-400">7 components</div>
    </div>
  </a>
  <a href="/docs/catalog/cloudflare" class="flex items-center gap-3 p-4 rounded-lg border border-purple-900/30 bg-slate-900/30 hover:bg-slate-800/50 transition-colors">
    <img src="/images/providers/cloudflare.svg" alt="Cloudflare" class="w-8 h-8 object-contain" />
    <div>
      <div class="font-semibold text-white">Cloudflare</div>
      <div class="text-sm text-slate-400">7 components</div>
    </div>
  </a>
  <a href="/docs/catalog/civo" class="flex items-center gap-3 p-4 rounded-lg border border-purple-900/30 bg-slate-900/30 hover:bg-slate-800/50 transition-colors">
    <img src="/images/providers/civo.svg" alt="Civo" class="w-8 h-8 object-contain" />
    <div>
      <div class="font-semibold text-white">Civo</div>
      <div class="text-sm text-slate-400">12 components</div>
    </div>
  </a>
  <a href="/docs/catalog/digitalocean" class="flex items-center gap-3 p-4 rounded-lg border border-purple-900/30 bg-slate-900/30 hover:bg-slate-800/50 transition-colors">
    <img src="/images/providers/digital-ocean.svg" alt="DigitalOcean" class="w-8 h-8 object-contain" />
    <div>
      <div class="font-semibold text-white">DigitalOcean</div>
      <div class="text-sm text-slate-400">14 components</div>
    </div>
  </a>
  <a href="/docs/catalog/atlas" class="flex items-center gap-3 p-4 rounded-lg border border-purple-900/30 bg-slate-900/30 hover:bg-slate-800/50 transition-colors">
    <img src="/images/providers/mongodb-atlas.svg" alt="MongoDB Atlas" class="w-8 h-8 object-contain" />
    <div>
      <div class="font-semibold text-white">Atlas</div>
      <div class="text-sm text-slate-400">1 component</div>
    </div>
  </a>
  <a href="/docs/catalog/confluent" class="flex items-center gap-3 p-4 rounded-lg border border-purple-900/30 bg-slate-900/30 hover:bg-slate-800/50 transition-colors">
    <img src="/images/providers/confluent.svg" alt="Confluent" class="w-8 h-8 object-contain" />
    <div>
      <div class="font-semibold text-white">Confluent</div>
      <div class="text-sm text-slate-400">1 component</div>
    </div>
  </a>
  <a href="/docs/catalog/kubernetes" class="flex items-center gap-3 p-4 rounded-lg border border-purple-900/30 bg-slate-900/30 hover:bg-slate-800/50 transition-colors">
    <img src="/images/providers/kubernetes.svg" alt="Kubernetes" class="w-8 h-8 object-contain" />
    <div>
      <div class="font-semibold text-white">Kubernetes</div>
      <div class="text-sm text-slate-400">36 components</div>
    </div>
  </a>
  <a href="/docs/catalog/snowflake" class="flex items-center gap-3 p-4 rounded-lg border border-purple-900/30 bg-slate-900/30 hover:bg-slate-800/50 transition-colors">
    <img src="/images/providers/snowflake.svg" alt="Snowflake" class="w-8 h-8 object-contain" />
    <div>
      <div class="font-semibold text-white">Snowflake</div>
      <div class="text-sm text-slate-400">1 component</div>
    </div>
  </a>
</div>

## Key Features

- **One Model, Many Clouds**: Single API structure across AWS, GCP, Azure, and Kubernetes
- **Validation First**: Buf ProtoValidate catches errors before deployment
- **Battle-Tested Modules**: Curated Pulumi and OpenTofu modules
- **CLI-First Workflow**: Developer-grade CLI for all operations
- **Security & Governance**: Provider credentials as stack inputs, consistent labeling

## Quick Example

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: RedisKubernetes
metadata:
  name: my-redis
  labels:
    openmcf.org/provisioner: pulumi
spec:
  replicas: 3
  resources:
    limits:
      memory: 2Gi
      cpu: 1000m
```

```bash
# Simple kubectl-style workflow
openmcf validate redis.yaml
openmcf apply -f redis.yaml

# Or use provisioner-specific commands
openmcf pulumi up -f redis.yaml --stack myorg/project/dev
```

## Troubleshooting

Having issues? Common solutions:

- **Manifest validation errors** — Check field names and types against the component's protobuf schema
- **Authentication and credentials** — See the [Credentials Management](/docs/guides/credentials) guide
- **Pulumi/OpenTofu-specific issues** — Ensure the provisioner CLI is installed and your backend is configured
- **Network and connectivity** — The CLI clones modules from GitHub; verify network access

## Resources

- [GitHub Repository](https://github.com/plantonhq/openmcf)
- [Buf Schema Registry](https://buf.build/openmcf/apis)
- [Issue Tracker](https://github.com/plantonhq/openmcf/issues)

