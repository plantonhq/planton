# AliCloud EIP Address

Deploys an Alibaba Cloud Elastic IP Address (EIP). The component provisions a standalone public IPv4 address that persists independently of any cloud resource, allowing it to be associated with and disassociated from NAT gateways, load balancers, VPN gateways, and ECS instances without changing the address.

## What Gets Created

When you deploy an AliCloudEipAddress resource, OpenMCF provisions:

- **EIP** -- an `alicloud_eip_address` resource in the specified region with configurable bandwidth, ISP, and metering settings

## Prerequisites

- **Alibaba Cloud credentials** configured via environment variables or OpenMCF provider config

## Quick Start

Create a file `eip.yaml`:

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudEipAddress
metadata:
  name: my-eip
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AliCloudEipAddress.my-eip
spec:
  region: cn-hangzhou
  addressName: my-nat-eip
  description: EIP for NAT gateway outbound access
  bandwidth: 10
```

Deploy:

```shell
openmcf apply -f eip.yaml
```

This allocates a 10 Mbps EIP using BGP multi-line and PayByTraffic metering.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | Alibaba Cloud region for the EIP allocation (e.g., `cn-hangzhou`, `us-west-1`). | Required; non-empty |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `addressName` | `string` | `""` | EIP display name (1-128 chars). |
| `description` | `string` | `""` | Human-readable description. |
| `bandwidth` | `int32` | `5` | Maximum outbound bandwidth in Mbps (1-1000). |
| `internetChargeType` | `string` | `"PayByTraffic"` | Metering method. `PayByTraffic` bills per GB. `PayByBandwidth` bills for reserved bandwidth. Immutable after creation. |
| `isp` | `string` | `"BGP"` | ISP line type. `BGP` (default), `BGP_PRO`, `ChinaTelecom`, `ChinaUnicom`, `ChinaMobile`, and L2/FinanceCloud/International variants. Immutable after creation. |
| `resourceGroupId` | `string` | `""` | Resource group ID for organizational grouping. |
| `tags` | `map<string, string>` | `{}` | Tags applied to the EIP. Merged with standard OpenMCF tags. |

## Examples

### Minimal EIP

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudEipAddress
metadata:
  name: my-eip
spec:
  region: cn-hangzhou
```

### Named EIP for NAT Gateway

A named EIP intended for association with a NAT gateway. Uses default 5 Mbps bandwidth and PayByTraffic metering:

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudEipAddress
metadata:
  name: nat-eip
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AliCloudEipAddress.nat-eip
spec:
  region: cn-shanghai
  addressName: prod-nat-eip
  description: EIP for production NAT gateway outbound access
  tags:
    purpose: nat
    team: platform
```

### High-Bandwidth Production EIP

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudEipAddress
metadata:
  name: prod-lb-eip
  org: my-org
  env: production
spec:
  region: cn-hangzhou
  addressName: prod-alb-eip
  description: High-bandwidth EIP for production ALB
  bandwidth: 100
  internetChargeType: PayByBandwidth
  isp: BGP_PRO
  resourceGroupId: rg-prod-123
  tags:
    team: platform
    costCenter: engineering
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `eip_id` | `string` | The EIP allocation ID assigned by Alibaba Cloud |
| `ip_address` | `string` | The allocated public IPv4 address |

## Related Components

- [AliCloudNatGateway](/docs/catalog/alicloud/alicloudnatgateway) -- associate this EIP for SNAT outbound internet access
- [AliCloudApplicationLoadBalancer](/docs/catalog/alicloud/alicloudapplicationloadbalancer) -- use this EIP for internet-facing ALB
- [AliCloudNetworkLoadBalancer](/docs/catalog/alicloud/alicloudnetworkloadbalancer) -- use this EIP for internet-facing NLB
- [AliCloudVpnGateway](/docs/catalog/alicloud/alicloudvpngateway) -- use this EIP for VPN gateway public endpoint
