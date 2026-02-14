# DigitalOcean VPC

Deploys a Virtual Private Cloud (VPC) on DigitalOcean, providing a private, isolated network for Droplets, Kubernetes clusters, managed databases, and other resources within a single region. The component supports both explicit CIDR allocation and DigitalOcean's automatic IP range generation.

## What Gets Created

When you deploy a DigitalOceanVpc resource, OpenMCF provisions:

- **VPC** — a `digitalocean_vpc` resource in the specified region, with an optional user-defined CIDR block or an auto-generated `/20` range when no IP range is specified

## Prerequisites

- **DigitalOcean credentials** configured via environment variables or OpenMCF provider config
- **A target region** selected from DigitalOcean's available datacenter regions
- **CIDR planning** (optional) — if you need a specific IP range, ensure it does not overlap with existing VPCs or DigitalOcean's reserved ranges (`10.244.0.0/16`, `10.245.0.0/16`, `10.246.0.0/24`, `10.229.0.0/16`)

## Quick Start

Create a file `vpc.yaml`:

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanVpc
metadata:
  name: my-vpc
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.DigitalOceanVpc.my-vpc
spec:
  region: nyc3
```

Deploy:

```shell
openmcf apply -f vpc.yaml
```

This creates a VPC in the NYC3 region with a DigitalOcean auto-generated `/20` CIDR block (4,096 IPs).

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `enum` | DigitalOcean region where the VPC will be created. Valid values: `nyc3`, `sfo3`, `fra1`, `sgp1`, `lon1`, `tor1`, `blr1`, `ams3`. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `description` | `string` | `""` | A human-readable description for the VPC. Maximum 100 characters. |
| `ipRangeCidr` | `string` | `""` (auto-generated `/20`) | The IP range for the VPC in CIDR notation. Only `/16`, `/20`, or `/24` blocks are supported. When omitted, DigitalOcean auto-generates a non-conflicting `/20` block. |
| `isDefaultForRegion` | `bool` | `false` | Whether this VPC should be set as the default for the specified region. Only one VPC can be the default per region. |

## Examples

### Dev VPC with Auto-Generated IP Range

A minimal VPC for development, letting DigitalOcean auto-assign a `/20` CIDR block:

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanVpc
metadata:
  name: dev-vpc
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.DigitalOceanVpc.dev-vpc
spec:
  region: sfo3
  description: "Development environment VPC"
```

### Staging VPC with Explicit /20 CIDR

A staging VPC with a specific IP range to avoid conflicts when peering with other VPCs:

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanVpc
metadata:
  name: staging-vpc
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.DigitalOceanVpc.staging-vpc
spec:
  region: fra1
  description: "Staging environment VPC"
  ipRangeCidr: "10.100.16.0/20"
```

### Production VPC with Large /16 CIDR

A production VPC with the maximum `/16` block (65,536 IPs) for workloads expected to scale, such as VPC-native Kubernetes clusters and managed databases:

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanVpc
metadata:
  name: prod-vpc
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.DigitalOceanVpc.prod-vpc
spec:
  region: nyc3
  description: "Production VPC for all services"
  ipRangeCidr: "10.101.0.0/16"
  isDefaultForRegion: true
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `vpcId` | `string` | The unique identifier (UUID) of the created DigitalOcean VPC |

## Related Components

- [DigitalOceanKubernetesCluster](/docs/catalog/digitalocean/digitaloceankubernetescluster) — deploys a managed Kubernetes cluster into the VPC
- [DigitalOceanDatabaseCluster](/docs/catalog/digitalocean/digitaloceandatabasecluster) — provisions managed databases with private VPC connectivity
- [DigitalOceanDroplet](/docs/catalog/digitalocean/digitaloceandroplet) — creates Droplets placed within the VPC
- [DigitalOceanLoadBalancer](/docs/catalog/digitalocean/digitaloceanloadbalancer) — provisions load balancers that route traffic to VPC resources
- [DigitalOceanFirewall](/docs/catalog/digitalocean/digitaloceanfirewall) — controls network access to resources within the VPC
