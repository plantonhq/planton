---
title: "Public IP"
description: "Public IP deployment documentation"
icon: "package"
order: 100
componentName: "azurepublicip"
---

# Azure Public IP

Deploys an Azure Public IP Address with Standard SKU and static allocation in a specified region and resource group. The component supports optional DNS label assignment, availability zone configuration, and idle timeout tuning. Public IPs created by this component are referenced by downstream resources such as load balancers, application gateways, and NAT gateways via their resource ID.

## What Gets Created

When you deploy an AzurePublicIp resource, OpenMCF provisions:

- **Public IP Address** — a `network.PublicIp` resource in the specified region and resource group, configured with Standard SKU and static allocation (Basic SKU was retired September 2025; Standard SKU requires static allocation)
- **DNS A Record** — when `domainNameLabel` is set, Azure creates an A record at `{label}.{region}.cloudapp.azure.com` pointing to the allocated IP
- **Azure Tags** — resource metadata tags applied to the Public IP for tracking and governance

## Prerequisites

- **Azure credentials** configured via environment variables or OpenMCF provider config
- **An Azure Resource Group** where the Public IP will be created (can reference an AzureResourceGroup resource)
- **Region selection** — the Public IP must be in the same region as the resource it will be attached to (load balancer, application gateway, NAT gateway, etc.)

## Quick Start

Create a file `publicip.yaml`:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzurePublicIp
metadata:
  name: my-pip
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AzurePublicIp.my-pip
spec:
  region: eastus
  resourceGroup: my-rg
  name: my-pip
```

Deploy:

```shell
openmcf apply -f publicip.yaml
```

This creates a Standard SKU Public IP with static allocation, a 4-minute idle timeout, and no DNS label or zone preference.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | Azure region for the Public IP (e.g., `eastus`, `westeurope`). Must match the region of the resource it will be attached to. | Required, minimum length 1 |
| `resourceGroup` | `StringValueOrRef` | Azure Resource Group name. Can reference an AzureResourceGroup resource via `valueFrom`. | Required |
| `name` | `string` | Name of the Public IP resource. Must be unique within the resource group. Allowed characters: alphanumeric, underscores, hyphens, and periods. Must start with an alphanumeric character. | Required, 1–80 characters |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `domainNameLabel` | `string` | `""` | DNS label that creates an A record at `{label}.{region}.cloudapp.azure.com`. Must start with a lowercase letter, end with a letter or digit, contain only lowercase letters, digits, and hyphens, and be 3–63 characters. Must be unique within the Azure region. |
| `zones` | `string[]` | `[]` | Availability zones for the Public IP. Valid values: `"1"`, `"2"`, `"3"`. Use `["1", "2", "3"]` for zone-redundant (recommended for production), `["1"]` for zonal, or omit for no zone preference. Zone support depends on the Azure region. |
| `idleTimeoutInMinutes` | `int` | `4` | Idle timeout in minutes for TCP/UDP connections. Higher values suit long-lived connections (WebSocket, gRPC streaming, database connections). Lower values free resources faster for short-lived traffic. Range: 4–30. |

## Examples

### Basic Public IP

A minimal Public IP for development or testing:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzurePublicIp
metadata:
  name: dev-pip
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AzurePublicIp.dev-pip
spec:
  region: eastus
  resourceGroup: dev-rg
  name: dev-pip
```

### Public IP with DNS Label

A Public IP with a DNS label for a stable domain name, useful when external clients need a human-readable endpoint:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzurePublicIp
metadata:
  name: api-pip
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.AzurePublicIp.api-pip
spec:
  region: westeurope
  resourceGroup: staging-rg
  name: api-pip
  domainNameLabel: my-api-staging
```

After deployment, the Public IP is reachable at `my-api-staging.westeurope.cloudapp.azure.com`.

### Zone-Redundant Production Public IP

A production Public IP spread across all three availability zones with an extended idle timeout for long-lived connections:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzurePublicIp
metadata:
  name: prod-lb-pip
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzurePublicIp.prod-lb-pip
spec:
  region: eastus
  resourceGroup: prod-rg
  name: prod-lb-pip
  domainNameLabel: prod-api
  zones:
    - "1"
    - "2"
    - "3"
  idleTimeoutInMinutes: 15
```

### Using Foreign Key References

Reference an OpenMCF-managed resource group instead of hardcoding the name:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzurePublicIp
metadata:
  name: ref-pip
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzurePublicIp.ref-pip
spec:
  region: eastus
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: my-rg
      field: status.outputs.resource_group_name
  name: ref-pip
  zones:
    - "1"
    - "2"
    - "3"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `public_ip_id` | `string` | Azure Resource Manager ID of the Public IP. This is the primary output referenced by downstream resources (AzureApplicationGateway, AzureLoadBalancer, AzureNatGateway) via `valueFrom`. |
| `ip_address` | `string` | The allocated static IPv4 address. Persistent for the lifetime of the resource. |
| `fqdn` | `string` | Fully qualified domain name associated with this Public IP. Only populated when `domainNameLabel` is set. Format: `{domainNameLabel}.{region}.cloudapp.azure.com`. |
| `public_ip_name` | `string` | Name of the Public IP resource. |

## Related Components

- [AzureResourceGroup](/docs/catalog/azure/resource-group) — provides the resource group for Public IP placement
- [AzureLoadBalancer](/docs/catalog/azure/load-balancer) — attaches a Public IP as a frontend IP configuration
- [AzureApplicationGateway](/docs/catalog/azure/application-gateway) — attaches a Public IP for HTTP/HTTPS ingress
- [AzureNatGateway](/docs/catalog/azure/nat-gateway) — attaches a Public IP for outbound NAT
