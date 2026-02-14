---
title: "Load Balancer"
description: "Load Balancer deployment documentation"
icon: "package"
order: 100
componentName: "azureloadbalancer"
---

# Azure Load Balancer

Deploys an Azure Standard Load Balancer with configurable frontend (public or internal), backend address pools, health probes, and load balancing rules. The component bundles these sub-resources because a load balancer without them is non-functional.

## What Gets Created

When you deploy an AzureLoadBalancer resource, OpenMCF provisions:

- **Load Balancer** — a `lb.LoadBalancer` resource using Standard SKU in the specified region and resource group, with a single frontend IP configuration that is either public (using a public IP address) or internal (using a VNet subnet)
- **Backend Address Pools** — one `lb.BackendAddressPool` resource per entry in `backendPools`; actual pool membership (VMs, VMSS instances, NICs) is managed externally via AKS node pools, VMSS configurations, or NIC-to-pool associations
- **Health Probes** — one `lb.Probe` resource per entry in `healthProbes`, supporting TCP, HTTP, and HTTPS protocols with configurable intervals and failure thresholds
- **Load Balancing Rules** — one `lb.Rule` resource per entry in `rules`, mapping frontend port/protocol combinations to backend pools and health probes
- **Azure Tags** — resource metadata tags applied to the load balancer for tracking and governance

## Prerequisites

- **Azure credentials** configured via environment variables or OpenMCF provider config
- **An Azure Resource Group** where the load balancer will be created (can reference an AzureResourceGroup resource)
- **A Public IP or Subnet** — for a public load balancer, a Standard SKU public IP address; for an internal load balancer, a VNet subnet ID (can reference AzurePublicIp or AzureSubnet resources)
- **Backend infrastructure** — VMs, VMSS instances, or AKS node pools that will be associated with the backend pools after deployment

## Quick Start

Create a file `loadbalancer.yaml`:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureLoadBalancer
metadata:
  name: my-lb
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AzureLoadBalancer.my-lb
spec:
  region: eastus
  resourceGroup: my-rg
  name: my-lb
  publicIpId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Network/publicIPAddresses/my-pip
  backendPools:
    - name: default
  healthProbes:
    - name: http-probe
      protocol: Http
      port: 80
      requestPath: /health
  rules:
    - name: http-rule
      protocol: Tcp
      frontendPort: 80
      backendPort: 80
      backendPoolName: default
      probeName: http-probe
```

Deploy:

```shell
openmcf apply -f loadbalancer.yaml
```

This creates a public-facing Standard Load Balancer with one backend pool, an HTTP health probe on port 80, and a TCP rule forwarding port 80 traffic to the backend pool.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | Azure region for the Load Balancer (e.g., `eastus`, `westeurope`). Must match the region of backend resources. | Required, minimum length 1 |
| `resourceGroup` | `StringValueOrRef` | Azure Resource Group name. Can reference an AzureResourceGroup resource via `valueFrom`. | Required |
| `name` | `string` | Name of the Load Balancer. Must be unique within the resource group. | Required, 1-80 characters |
| `publicIpId` | `StringValueOrRef` | Public IP address resource ID for a public (internet-facing) load balancer. Mutually exclusive with `subnetId` -- exactly one must be set. Can reference an AzurePublicIp resource via `valueFrom`. | Conditionally required |
| `subnetId` | `StringValueOrRef` | Subnet resource ID for an internal (private VNet) load balancer. Mutually exclusive with `publicIpId` -- exactly one must be set. Can reference an AzureSubnet resource via `valueFrom`. | Conditionally required |
| `backendPools` | `AzureBackendPool[]` | Backend address pools that receive load-balanced traffic. | Required, minimum 1 item |
| `backendPools[].name` | `string` | Name of the backend pool. Must be unique within the load balancer. | Required, 1-80 characters |
| `healthProbes` | `AzureHealthProbe[]` | Health probes that check backend instance availability. | Required, minimum 1 item |
| `healthProbes[].name` | `string` | Name of the health probe. Must be unique within the load balancer. | Required, 1-80 characters |
| `healthProbes[].protocol` | `string` | Protocol for the health probe. Values: `Tcp`, `Http`, `Https`. | Required |
| `healthProbes[].port` | `int` | Port number to probe on backend instances. | Required, 1-65535 |
| `rules` | `AzureLoadBalancingRule[]` | Load balancing rules that define traffic routing from frontend to backend pools. | Required, minimum 1 item |
| `rules[].name` | `string` | Name of the load balancing rule. Must be unique within the load balancer. | Required, 1-80 characters |
| `rules[].protocol` | `string` | Transport protocol. Values: `Tcp`, `Udp`, `All` (HA ports). | Required |
| `rules[].frontendPort` | `int` | Port on the frontend that receives traffic. Use 0 for HA ports (protocol `All`). | 0-65534 |
| `rules[].backendPort` | `int` | Port on backend instances that receives forwarded traffic. Use 0 for HA ports (protocol `All`). | 0-65535 |
| `rules[].backendPoolName` | `string` | Name of the backend pool to route traffic to. Must match a name in `backendPools`. | Required |
| `rules[].probeName` | `string` | Name of the health probe for backend health checks. Must match a name in `healthProbes`. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `privateIpAddress` | `string` | _(dynamic)_ | Static private IP address for an internal load balancer. Only valid when `subnetId` is set. Must fall within the subnet's address range. If omitted, Azure dynamically allocates a private IP. |
| `healthProbes[].requestPath` | `string` | | URI path for HTTP/HTTPS probes. Required when protocol is `Http` or `Https`, ignored for `Tcp`. The probe sends a GET request and expects HTTP 200. |
| `healthProbes[].intervalInSeconds` | `int` | `15` | Interval between probe attempts, in seconds. Lower values detect failures faster. Minimum: 5. |
| `healthProbes[].numberOfProbes` | `int` | `2` | Consecutive probe failures before marking a backend unhealthy. Minimum: 1. |
| `rules[].idleTimeoutInMinutes` | `int` | `4` | TCP idle timeout in minutes. Connections idle longer than this are closed. Range: 4-100. Higher values suit long-lived connections (WebSocket, database pools). |
| `rules[].enableFloatingIp` | `bool` | `false` | Enable floating IP (Direct Server Return). When enabled, backends receive the original frontend IP as destination. Required for SQL AlwaysOn availability groups and some HA clustering scenarios. |
| `rules[].disableOutboundSnat` | `bool` | `false` | Disable outbound SNAT for this rule's backend pool. Enable when using explicit outbound rules or a NAT Gateway for outbound connectivity to avoid SNAT port exhaustion. |

## Examples

### Public Load Balancer with HTTP

A simple internet-facing load balancer forwarding HTTP traffic to a single backend pool:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureLoadBalancer
metadata:
  name: web-lb
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzureLoadBalancer.web-lb
spec:
  region: eastus
  resourceGroup: prod-rg
  name: web-lb
  publicIpId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.Network/publicIPAddresses/web-pip
  backendPools:
    - name: default
  healthProbes:
    - name: http-probe
      protocol: Http
      port: 80
      requestPath: /health
  rules:
    - name: http-rule
      protocol: Tcp
      frontendPort: 80
      backendPort: 80
      backendPoolName: default
      probeName: http-probe
```

### Internal Load Balancer with Static IP

A private VNet load balancer with a static IP address for stable DNS resolution and firewall rules:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureLoadBalancer
metadata:
  name: internal-lb
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzureLoadBalancer.internal-lb
spec:
  region: westeurope
  resourceGroup: prod-rg
  name: internal-lb
  subnetId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.Network/virtualNetworks/prod-vnet/subnets/app
  privateIpAddress: "10.0.1.100"
  backendPools:
    - name: api-pool
  healthProbes:
    - name: tcp-probe
      protocol: Tcp
      port: 8080
  rules:
    - name: api-rule
      protocol: Tcp
      frontendPort: 8080
      backendPort: 8080
      backendPoolName: api-pool
      probeName: tcp-probe
      idleTimeoutInMinutes: 30
```

### Multi-Rule Load Balancer with HTTPS and TCP

A public load balancer serving both HTTPS web traffic and a TCP database proxy, each with dedicated backend pools and health probes:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureLoadBalancer
metadata:
  name: multi-lb
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzureLoadBalancer.multi-lb
spec:
  region: eastus
  resourceGroup: prod-rg
  name: multi-lb
  publicIpId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.Network/publicIPAddresses/multi-pip
  backendPools:
    - name: web-servers
    - name: db-proxies
  healthProbes:
    - name: https-probe
      protocol: Https
      port: 443
      requestPath: /healthz
      intervalInSeconds: 10
      numberOfProbes: 3
    - name: tcp-3306-probe
      protocol: Tcp
      port: 3306
      intervalInSeconds: 10
  rules:
    - name: https-rule
      protocol: Tcp
      frontendPort: 443
      backendPort: 443
      backendPoolName: web-servers
      probeName: https-probe
    - name: mysql-rule
      protocol: Tcp
      frontendPort: 3306
      backendPort: 3306
      backendPoolName: db-proxies
      probeName: tcp-3306-probe
      idleTimeoutInMinutes: 30
      disableOutboundSnat: true
```

### SQL AlwaysOn with Floating IP

An internal load balancer for SQL Server AlwaysOn availability groups, using floating IP (Direct Server Return) so backends receive the original frontend IP:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureLoadBalancer
metadata:
  name: sql-lb
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzureLoadBalancer.sql-lb
spec:
  region: eastus
  resourceGroup: prod-rg
  name: sql-lb
  subnetId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.Network/virtualNetworks/prod-vnet/subnets/data
  privateIpAddress: "10.0.2.50"
  backendPools:
    - name: sql-nodes
  healthProbes:
    - name: sql-probe
      protocol: Tcp
      port: 59999
      intervalInSeconds: 5
      numberOfProbes: 2
  rules:
    - name: sql-rule
      protocol: Tcp
      frontendPort: 1433
      backendPort: 1433
      backendPoolName: sql-nodes
      probeName: sql-probe
      idleTimeoutInMinutes: 4
      enableFloatingIp: true
      disableOutboundSnat: true
```

### Using Foreign Key References

Reference OpenMCF-managed resources instead of hardcoding Azure resource IDs:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureLoadBalancer
metadata:
  name: ref-lb
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzureLoadBalancer.ref-lb
spec:
  region: eastus
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: my-rg
      field: status.outputs.resource_group_name
  name: ref-lb
  publicIpId:
    valueFrom:
      kind: AzurePublicIp
      name: my-pip
      field: status.outputs.public_ip_id
  backendPools:
    - name: default
  healthProbes:
    - name: http-probe
      protocol: Http
      port: 80
      requestPath: /health
  rules:
    - name: http-rule
      protocol: Tcp
      frontendPort: 80
      backendPort: 80
      backendPoolName: default
      probeName: http-probe
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `lb_id` | `string` | Azure Resource Manager ID of the Load Balancer |
| `lb_name` | `string` | Name of the Load Balancer |
| `frontend_ip_address` | `string` | Frontend IP address of the Load Balancer. For public LBs, this is the public IP address. For internal LBs, this is the private IP address from the subnet. |
| `frontend_ip_configuration_id` | `string` | Azure Resource Manager ID of the frontend IP configuration. Useful for creating NAT rules or additional routing configurations. |
| `backend_pool_id` | `string` | Azure Resource Manager ID of the first (default) backend address pool. Use this to associate VMSS instances, AKS node pools, or NICs with the pool. |

## Related Components

- [AzureResourceGroup](/docs/catalog/azure/resource-group) -- provides the resource group for load balancer placement
- [AzurePublicIp](/docs/catalog/azure/public-ip) -- provides a Standard SKU public IP address for public load balancers
- [AzureSubnet](/docs/catalog/azure/subnet) -- provides a VNet subnet for internal load balancers
- [AzureVpc](/docs/catalog/azure/vpc-virtual-network) -- provides the virtual network that contains subnets used by internal load balancers
- [AzureDnsRecord](/docs/catalog/azure/dns-record) -- creates DNS A-records pointing to the load balancer frontend IP
