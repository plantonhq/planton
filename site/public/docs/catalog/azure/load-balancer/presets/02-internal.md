---
title: "Internal Load Balancer"
description: "This preset creates an internal (private VNet) Azure Load Balancer with Standard SKU, using a subnet frontend instead of a public IP. Traffic is distributed across backend instances using a private..."
type: "preset"
rank: "02"
presetSlug: "02-internal"
componentSlug: "load-balancer"
componentTitle: "Load Balancer"
provider: "azure"
icon: "package"
order: 2
---

# Internal Load Balancer

This preset creates an internal (private VNet) Azure Load Balancer with Standard SKU, using a subnet frontend instead of a public IP. Traffic is distributed across backend instances using a private IP address within the VNet. This is the standard configuration for internal services that should not be exposed to the internet -- such as internal APIs, microservice tiers, or database connection pools.

## When to Use

- Internal microservices or APIs that receive traffic only from within the VNet or peered networks
- Database or middleware tiers in a multi-tier architecture where the load balancer must stay private
- AKS internal ingress or internal service endpoints that need a stable private IP
- Workloads that require a predictable private frontend IP for DNS or firewall rules

## Key Configuration Choices

- **Internal frontend** (`subnetId`) -- Uses a private IP from the specified subnet instead of a public IP. No internet exposure
- **Optional static IP** (`privateIpAddress`) -- Set a specific IP within the subnet range for predictable addressing. Omit for dynamic allocation
- **TCP health probe** (`healthProbes: Tcp on port 8080`) -- Simple TCP connectivity check every 15 seconds; marks backend unhealthy after 2 failures
- **TCP rule on port 8080** (`rules: Tcp 8080→8080`) -- Routes internal traffic on port 8080 to backend port 8080

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<azure-region>` | Azure region (must match the subnet's VNet region) | Your regional deployment strategy |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |
| `<your-lb-name>` | Name for the load balancer (unique within resource group) | Your naming convention |
| `<subnet-resource-id>` | Full ARM resource ID of the target subnet | Azure portal or `AzureSubnet` status outputs |
| `<private-ip-address>` | Static private IP within the subnet range (or remove line for dynamic) | Your network IP plan |

## Related Presets

- **01-public** -- Use instead for internet-facing load balancing with a public IP frontend
