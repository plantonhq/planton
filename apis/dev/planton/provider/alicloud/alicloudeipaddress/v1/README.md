# AliCloudEipAddress

Manages an Alibaba Cloud Elastic IP Address (EIP).

## Overview

An EIP is a static, public IPv4 address that exists independently of any cloud resource. Unlike the auto-assigned public IP on an ECS instance (which is released when the instance stops), an EIP persists until explicitly released. It can be associated with and disassociated from ECS instances, NAT gateways, ALB/NLB load balancers, and VPN gateways without changing the address.

### What Gets Created

- **EIP** -- a standalone Elastic IP Address allocated in the specified region

Association with a target resource (e.g., NAT gateway, load balancer) is handled by the downstream component, not by this component.

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `region` | string | Alibaba Cloud region (e.g., `cn-hangzhou`, `cn-shanghai`, `us-west-1`) |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `addressName` | string | `""` | EIP display name (1-128 chars) |
| `description` | string | `""` | Human-readable description |
| `bandwidth` | int32 | `5` | Maximum outbound bandwidth in Mbps (1-1000) |
| `internetChargeType` | string | `"PayByTraffic"` | Metering method: `PayByTraffic` or `PayByBandwidth` |
| `isp` | string | `"BGP"` | ISP line type (see below) |
| `resourceGroupId` | string | `""` | Resource group for organizational grouping |
| `tags` | map | `{}` | Key-value tags applied to the EIP |

### ISP Values

| Value | Description |
|-------|-------------|
| `BGP` | Multi-line BGP (default, available in all regions) |
| `BGP_PRO` | Premium BGP with optimized routing for China mainland |
| `ChinaTelecom` | China Telecom single-carrier |
| `ChinaUnicom` | China Unicom single-carrier |
| `ChinaMobile` | China Mobile single-carrier |
| `ChinaTelecom_L2` | China Telecom L2 single-carrier |
| `ChinaUnicom_L2` | China Unicom L2 single-carrier |
| `ChinaMobile_L2` | China Mobile L2 single-carrier |
| `BGP_FinanceCloud` | BGP for Chinese finance cloud regions |
| `BGP_International` | International BGP (outside mainland China) |

### Bandwidth and Charging

- **PayByTraffic**: `bandwidth` acts as a ceiling. You pay per GB of outbound data. Best for bursty or unpredictable workloads.
- **PayByBandwidth**: `bandwidth` is the guaranteed allocation. You pay for the full bandwidth regardless of usage. Best for steady, high-throughput workloads.

Both `internetChargeType` and `isp` are immutable after creation (changing them requires replacing the EIP).

## Stack Outputs

| Output | Description |
|--------|-------------|
| `eip_id` | The EIP allocation ID, referenced by downstream components via StringValueOrRef |
| `ip_address` | The allocated public IPv4 address |

## Related Components

- **AliCloudNatGateway** -- associates an EIP for SNAT outbound internet access
- **AliCloudApplicationLoadBalancer** -- uses EIPs for internet-facing load balancers
- **AliCloudNetworkLoadBalancer** -- uses EIPs for internet-facing L4 load balancers
- **AliCloudVpnGateway** -- uses an EIP for the VPN gateway public endpoint
