---
title: "Standard Production VPC"
description: "This preset creates a production-ready Alibaba Cloud VPC with a /16 CIDR block providing 65,536 IP addresses. The address space accommodates dozens of VSwitches across multiple availability zones..."
type: "preset"
rank: "01"
presetSlug: "01-standard-production"
componentSlug: "vpc"
componentTitle: "VPC"
provider: "alicloud"
icon: "package"
order: 1
---

# Standard Production VPC

This preset creates a production-ready Alibaba Cloud VPC with a /16 CIDR block providing 65,536 IP addresses. The address space accommodates dozens of VSwitches across multiple availability zones while remaining compatible with VPC peering and CEN (Cloud Enterprise Network) multi-VPC connectivity. Tags are included for organizational governance and cost tracking.

## When to Use

- Production workloads requiring a well-sized, peering-compatible network foundation
- Environments where multiple VSwitches will span several availability zones
- Deployments that need organizational tagging for cost allocation and resource management
- Foundation VPC for downstream components: VSwitches, NAT gateways, security groups, ACK clusters, RDS instances

## Key Configuration Choices

- **/16 CIDR block** (`cidrBlock: 10.0.0.0/16`) -- 65,536 IPs; large enough for dozens of VSwitches across all AZs in a region, small enough to avoid address conflicts when peering VPCs or connecting via CEN. A /8 would consume the entire 10.x.x.x range, making multi-VPC architectures impossible.
- **10.x range** (`10.0.0.0/16`) -- The 10.x range is the conventional choice for production infrastructure across cloud providers. Use a different /16 within 10.x (e.g., `10.1.0.0/16`, `10.2.0.0/16`) if you run multiple VPCs in the same account.
- **Tags included** (`team`, `costCenter`) -- Production VPCs should carry organizational metadata for cost attribution and operational ownership. Replace placeholders with your team's values.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<alibaba-cloud-region>` | Alibaba Cloud region code (e.g., `cn-hangzhou`, `cn-shanghai`, `ap-southeast-1`) | Your deployment region strategy |
| `<your-vpc-name>` | VPC name (1-128 chars; cannot start with `http://` or `https://`) | Choose a name following your organization's naming convention (e.g., `prod-platform-vpc`) |
| `<your-team>` | Team or business unit that owns this VPC | Your organizational structure |
| `<your-cost-center>` | Cost center code for billing attribution | Your finance or cloud operations team |

## Related Presets

- **02-development** -- Use instead for development and testing environments where tagging and description are unnecessary
- **03-dual-stack-ipv6** -- Use instead when your workloads require IPv6 connectivity
