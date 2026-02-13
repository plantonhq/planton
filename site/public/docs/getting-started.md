---
title: "Getting Started"
description: "Install OpenMCF CLI and deploy your first resource"
icon: "rocket"
order: 2
---

# Getting Started

This guide will help you install the OpenMCF CLI and deploy your first infrastructure resource.

## Prerequisites

Before you begin, ensure you have the following installed:

- **Git** - Required for cloning modules
- **Pulumi CLI** - For Pulumi deployments (`brew install pulumi`)
- **Terraform/OpenTofu CLI** - For Terraform deployments (`brew install opentofu`)
- **Cloud provider credentials** - AWS, GCP, Azure, etc. configured locally

## Installation

Install the OpenMCF CLI using Homebrew:

```bash
brew install plantonhq/tap/openmcf
```

Verify the installation:

```bash
openmcf version
```

## Your First Deployment

Let's deploy a PostgreSQL database to a local Kubernetes cluster.

### Step 1: Create a Local Kubernetes Cluster

If you don't have a Kubernetes cluster, create one using Kind:

```bash
brew install kind
kind create cluster
```

### Step 2: Create Your Manifest

Create a file named `postgres.yaml`:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesPostgres
metadata:
  name: dev-database
  labels:
    openmcf.org/provisioner: pulumi
spec:
  namespace:
    value: default
  container:
    replicas: 1
    resources:
      limits:
        cpu: 500m
        memory: 1Gi
    diskSize: 1Gi
```

### Step 3: Validate (Optional)

Validate your manifest before deployment:

```bash
openmcf validate -f postgres.yaml
```

This runs validation rules defined in the protobuf, catching errors like:
- Invalid CPU format
- Replica count out of range
- Missing required fields

### Step 4: Deploy

Set up a local Pulumi backend:

```bash
pulumi login --local
```

Deploy the resource using the unified kubectl-style command:

```bash
# Simple unified command (automatically detects provisioner from label)
openmcf apply -f postgres.yaml

# Or use the traditional Pulumi-specific command
openmcf pulumi up -f postgres.yaml --stack org/dev/local
```

### Step 5: Verify

Check that the PostgreSQL instance is running:

```bash
kubectl get pods
# You should see: dev-database-postgresql-0
```

## What Happened?

Behind the scenes, the CLI:

1. Read and validated your manifest
2. Identified the `KubernetesPostgres` deployment component
3. Cloned the corresponding Pulumi module from GitHub
4. Set up the environment with your manifest as input
5. Delegated to Pulumi to deploy the resources
6. PostgreSQL is now running in your cluster!

## Next Steps

- **Explore Components**: Browse [deployment components](/docs/catalog) across 14 cloud providers
- **Learn Concepts**: Understand the [architecture](/docs/concepts/architecture) and [key concepts](/docs/concepts)
- **Follow Tutorials**: Try the [AWS S3 bucket tutorial](/docs/tutorials/first-aws-resource) or [Kubernetes Postgres tutorial](/docs/tutorials/first-kubernetes-resource)
- **Deploy to Cloud**: Set up [AWS](/docs/guides/aws-provider-setup), [GCP](/docs/guides/gcp-provider-setup), or [Azure](/docs/guides/azure-provider-setup) credentials

## Common Commands

```bash
# Validate a manifest
openmcf validate -f config.yaml

# Unified kubectl-style commands (provisioner auto-detected from manifest)
openmcf apply -f config.yaml
openmcf destroy -f config.yaml

# Provisioner-specific commands (still supported)
openmcf pulumi up -f config.yaml --stack org/project/env
openmcf tofu apply -f config.yaml

# Override specific values
openmcf apply -f config.yaml --set spec.container.cpu=500m
```

## Troubleshooting

### "Module not found"

The CLI clones modules from GitHub. Ensure you have:
- Git installed and configured
- Network connectivity to GitHub

### "Validation failed"

Check the error message for specific field validation failures. Common issues:
- Invalid resource units (e.g., CPU should be "500m" not "500")
- Missing required fields
- Values outside allowed ranges

### "Pulumi/Terraform not found"

Install the required IaC tool:

```bash
brew install pulumi    # For Pulumi deployments
brew install opentofu  # For Terraform deployments
```

## Get Help

- **GitHub Issues**: [Report bugs or request features](https://github.com/plantonhq/openmcf/issues)
- **Documentation**: Browse the full [documentation](/docs)
- **Troubleshooting**: Check the [Troubleshooting Guide](/docs/troubleshooting) for common issues
- **Examples**: Check the repository for example manifests

