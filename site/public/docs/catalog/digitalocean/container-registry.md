---
title: "Container Registry"
description: "Container Registry deployment documentation"
icon: "package"
order: 100
componentName: "digitaloceancontainerregistry"
---

# DigitalOcean Container Registry

Deploys a private, OCI-compliant container registry on DigitalOcean for storing Docker images and Helm charts. The component configures the registry name, subscription tier, and region, then exposes the server URL and Docker credentials as stack outputs for use by downstream workloads.

## What Gets Created

When you deploy a DigitalOceanContainerRegistry resource, OpenMCF provisions:

- **Container Registry** — a `digitalocean_container_registry` resource with the specified name, subscription tier, and region
- **Docker Credentials** (Terraform only) — a `digitalocean_container_registry_docker_credentials` resource that generates write-enabled credentials for pushing and pulling images

DigitalOcean restricts each account to a single container registry. Deploying a second DigitalOceanContainerRegistry resource on the same account will fail.

## Prerequisites

- **DigitalOcean credentials** configured via environment variables or OpenMCF provider config
- **No existing container registry** on the target DigitalOcean account (one registry per account)

## Quick Start

Create a file `registry.yaml`:

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanContainerRegistry
metadata:
  name: my-registry
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.DigitalOceanContainerRegistry.my-registry
spec:
  name: my-registry
  subscriptionTier: starter
  region: nyc3
```

Deploy:

```shell
openmcf apply -f registry.yaml
```

This creates a container registry named `my-registry` on the free starter tier in the NYC3 region.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `name` | `string` | Registry name, unique within the DigitalOcean account. | Required, 1–63 characters, lowercase letters/numbers/hyphens, must start and end with an alphanumeric character. Pattern: `^[a-z0-9]([-a-z0-9]*[a-z0-9])?$` |
| `subscriptionTier` | `enum` | Storage and pricing tier. Valid values: `starter` (free, limited storage), `basic` (paid, moderate storage), `professional` (paid, highest storage, production ready). | Required |
| `region` | `enum` | DigitalOcean region where registry data is stored. Valid values: `nyc3`, `sfo3`, `fra1`, `sgp1`, `lon1`, `tor1`, `blr1`, `ams3`. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `garbageCollectionEnabled` | `bool` | `false` | Enable automatic garbage collection of untagged images. Note: the Pulumi provisioner logs a warning and ignores this field because the upstream DigitalOcean provider does not yet support it. The Terraform provisioner handles GC via a custom controller. |

## Examples

### Starter Registry for Development

A free-tier registry for personal or development use:

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanContainerRegistry
metadata:
  name: dev-registry
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.DigitalOceanContainerRegistry.dev-registry
spec:
  name: dev-registry
  subscriptionTier: starter
  region: sfo3
```

### Basic Registry in Europe

A paid-tier registry in Frankfurt for teams that need more storage than the starter tier provides:

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanContainerRegistry
metadata:
  name: team-registry
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.DigitalOceanContainerRegistry.team-registry
spec:
  name: team-registry
  subscriptionTier: basic
  region: fra1
```

### Professional Registry for Production

A production-grade registry with the highest storage allocation and garbage collection enabled:

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanContainerRegistry
metadata:
  name: prod-registry
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.DigitalOceanContainerRegistry.prod-registry
spec:
  name: prod-registry
  subscriptionTier: professional
  region: nyc3
  garbageCollectionEnabled: true
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `registryName` | `string` | Name of the created container registry |
| `serverUrl` | `string` | Full registry URL for Docker login (e.g., `registry.digitalocean.com/prod-registry`) |
| `region` | `string` | Region slug where the registry is hosted |

## Related Components

- [DigitalOceanKubernetesCluster](/docs/catalog/digitalocean/kubernetes-cluster) — integrates with the container registry to pull images without additional credentials
- [DigitalOceanAppPlatformService](/docs/catalog/digitalocean/app-platform-service) — can deploy containers directly from the registry
