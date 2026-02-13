---
title: "Publicip"
description: "Publicip deployment documentation"
icon: "package"
order: 100
componentName: "azurepublicip"
---

# AzurePublicIp: Research & Design Documentation

## 1. What Is an Azure Public IP Address?

An Azure Public IP Address is a static or dynamic IPv4/IPv6 address allocated from
Azure's public IP pool. It provides inbound internet connectivity to Azure resources
such as load balancers, application gateways, NAT gateways, VPN gateways, and
virtual machines.

Public IPs are one of Azure's foundational networking primitives. They sit at the
edge of the Azure network and are the entry point for all internet-facing traffic.

### Key Properties

- **SKU**: Standard (the only supported SKU since Basic was retired Sept 2025)
- **Allocation**: Static (Standard SKU always uses static allocation)
- **IP Version**: IPv4 (IPv6 is supported but niche)
- **Availability Zones**: Standard SKU supports zone-redundant and zonal deployments
- **DNS Integration**: Optional domain name label creates a stable FQDN
- **Idle Timeout**: Configurable TCP/UDP idle timeout (4-30 minutes)
- **Pricing**: ~$3.65/month per static Standard IPv4 address (as of 2025)

### Basic SKU Retirement

Microsoft retired the Basic SKU for Public IP Addresses on **September 30, 2025**.
All existing Basic IPs were automatically migrated to Standard. New deployments
cannot use Basic. This is why this OpenMCF component only supports Standard SKU.

Source: [Azure Basic SKU retirement announcement](https://azure.microsoft.com/en-us/updates/upgrade-to-standard-sku-public-ip-addresses-in-azure-by-30-september-2025-basic-sku-will-be-retired/)

## 2. Deployment Landscape

### How People Deploy Public IPs Today

#### Level 0: Azure Portal (Click-Ops)

The Azure Portal provides a GUI for creating Public IPs. Users select the SKU,
allocation method, region, and optional DNS label. This is fine for learning but
creates undocumented infrastructure.

#### Level 1: Azure CLI

```bash
az network public-ip create \
  --name my-pip \
  --resource-group my-rg \
  --location eastus \
  --sku Standard \
  --allocation-method Static \
  --dns-name myapp \
  --zone 1 2 3
```

Simple and scriptable, but lacks state management and drift detection.

#### Level 2: ARM Templates / Bicep

```bicep
resource publicIp 'Microsoft.Network/publicIPAddresses@2023-09-01' = {
  name: 'my-pip'
  location: 'eastus'
  sku: {
    name: 'Standard'
    tier: 'Regional'
  }
  properties: {
    publicIPAllocationMethod: 'Static'
    dnsSettings: {
      domainNameLabel: 'myapp'
    }
  }
  zones: ['1', '2', '3']
}
```

Azure-native IaC with full lifecycle management. Verbose but complete.

#### Level 3: Terraform

```hcl
resource "azurerm_public_ip" "main" {
  name                = "my-pip"
  location            = "eastus"
  resource_group_name = "my-rg"
  allocation_method   = "Static"
  sku                 = "Standard"
  domain_name_label   = "myapp"
  zones               = ["1", "2", "3"]
}
```

The most popular IaC approach for multi-cloud teams. Clean, readable, and
well-supported by the Azure Terraform provider.

#### Level 4: Pulumi

```go
publicIp, _ := network.NewPublicIp(ctx, "my-pip", &network.PublicIpArgs{
    Name:              pulumi.String("my-pip"),
    Location:          pulumi.String("eastus"),
    ResourceGroupName: pulumi.String("my-rg"),
    AllocationMethod:  pulumi.String("Static"),
    Sku:               pulumi.String("Standard"),
    DomainNameLabel:   pulumi.String("myapp"),
    Zones:             pulumi.StringArray{pulumi.String("1"), pulumi.String("2"), pulumi.String("3")},
})
```

Programmatic IaC with type safety and testability.

#### Level 5: OpenMCF (This Component)

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzurePublicIp
metadata:
  name: my-pip
spec:
  region: eastus
  resource_group: my-rg
  name: my-pip
  domain_name_label: myapp
  zones: ["1", "2", "3"]
```

Declarative, Kubernetes-style API that abstracts Pulumi/Terraform behind a consistent
multi-cloud interface. Enables infra chart composition where Public IPs are referenced
by downstream resources via `StringValueOrRef`.

## 3. 80/20 Analysis: What We Include and What We Skip

### Included (80% of Use Cases)

| Feature | Rationale |
|---------|-----------|
| Static allocation | Standard SKU requires it; all production use cases need it |
| Standard SKU | Only supported SKU since Sept 2025 |
| DNS label | Common for creating stable FQDNs; used frequently in production |
| Availability zones | Essential for production resilience |
| Idle timeout | Important production tunable for long-lived connections |
| Tags | Automatic via OpenMCF metadata |

### Excluded (20% Niche / Advanced)

| Feature | Rationale |
|---------|-----------|
| Basic SKU | Retired by Azure; not supported |
| Dynamic allocation | Incompatible with Standard SKU |
| IPv6 | Niche; <5% of Azure deployments use IPv6 public IPs |
| Global SKU tier | Only for cross-region load balancing; very niche |
| IP prefix | For CGNAT/NAT Gateway scenarios; covered by AzureNatGateway |
| DDoS protection plan | Governance concern; separate from IP provisioning |
| IP tags | Azure-internal routing hints; extremely niche |
| Edge zones | Azure Stack Edge; specialized hardware deployments |
| Reverse FQDN | Email server scenarios; very niche |

## 4. Downstream Consumers

Public IP addresses are consumed by three Azure resources in the OpenMCF plan:

### AzureApplicationGateway (R10)
- Uses `public_ip_id` for the frontend IP configuration
- Application Gateway requires a dedicated Standard SKU Public IP

### AzureLoadBalancer (R09)
- Uses `public_ip_id` for the frontend IP of public load balancers
- Internal load balancers don't need a Public IP

### AzureNatGateway (Existing)
- Already exists in OpenMCF; currently creates its own inline Public IP
- Could be refactored to reference an AzurePublicIp for more flexibility

### DNS Records
- `ip_address` output can be used to create A records in AzureDnsRecord
- `fqdn` output can be used for CNAME record targets

## 5. Infra Chart Integration

### Enterprise Network Foundation

In the `enterprise-network-foundation` infra chart, Public IPs are Layer 1 resources
created after the resource group:

```
AzureResourceGroup (Layer 0)
├── AzureVpc (Layer 1)
│   └── AzureSubnet (Layer 2)
├── AzurePublicIp [gateway] (Layer 1)  <-- THIS RESOURCE
│   └── AzureApplicationGateway (Layer 2) -- references public_ip_id
├── AzurePublicIp [lb] (Layer 1)
│   └── AzureLoadBalancer (Layer 2) -- references public_ip_id
├── AzurePublicIp [nat] (Layer 1)
│   └── AzureNatGateway (Layer 2) -- references public_ip_id
└── AzureLogAnalyticsWorkspace (Layer 1)
```

Each downstream resource references the Public IP via `StringValueOrRef`:

```yaml
public_ip_id:
  valueFrom:
    kind: AzurePublicIp
    name: gateway-pip
    fieldPath: status.outputs.public_ip_id
```

## 6. Design Decisions

### Why No SKU Field

Azure retired the Basic SKU on September 30, 2025. Since this platform is being
built in 2026, there is zero reason to support a deprecated SKU. Including it would:

1. Add a proto enum type for two values where only one is valid
2. Require validation to prevent selecting Basic
3. Confuse users with an option that doesn't work
4. Create maintenance burden for a dead code path

Hardcoding Standard in the IaC module eliminates all of this complexity. If Azure
introduces a new SKU tier in the future, we can add a `sku` field then.

### Why No Allocation Method Field

Standard SKU requires static allocation. Azure's API rejects Dynamic + Standard.
Including an `allocation_method` field would add a proto enum type with two values
where only one is valid. Same rationale as the SKU decision.

### Why Include Idle Timeout

The default idle timeout of 4 minutes is too short for many enterprise workloads.
WebSocket connections, gRPC streams, and database connections through a Public IP
will be terminated after 4 minutes of inactivity. A configurable timeout (up to 30
minutes) is a meaningful production lever that costs nothing to include in the spec.

### Why Include Zones

Availability zones provide resilience against datacenter failures. A zone-redundant
Public IP (`zones: ["1","2","3"]`) survives the loss of an entire availability zone.
This is table stakes for production infrastructure and easy to configure.

## 7. Scope Boundaries

### What This Component Does

- Creates a Standard SKU, Static allocation Azure Public IP Address
- Optionally configures a DNS domain name label for FQDN creation
- Optionally pins to specific availability zones (zonal or zone-redundant)
- Configures idle timeout for TCP/UDP connection lifecycle
- Tags the resource with OpenMCF metadata
- Exports the IP ID, address, FQDN, and name for downstream consumption

### What This Component Does NOT Do

- **NAT Gateway association** -- handled by AzureNatGateway
- **Load Balancer association** -- handled by AzureLoadBalancer
- **Application Gateway association** -- handled by AzureApplicationGateway
- **DDoS protection** -- governance concern at subscription level
- **IPv6 addresses** -- niche; future iteration if demand exists
- **IP prefix management** -- handled by AzureNatGateway for scale scenarios
