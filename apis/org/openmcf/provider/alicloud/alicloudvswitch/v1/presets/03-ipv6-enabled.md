# IPv6-Enabled Dual-Stack VSwitch

This preset creates a VSwitch with both IPv4 and IPv6 addressing enabled. The parent VPC must have IPv6 enabled (`enableIpv6: true` in AlicloudVpc) for IPv6 allocation to succeed. A /24 IPv4 CIDR is paired with an automatically assigned /64 IPv6 block selected by `ipv6CidrBlockMask`.

## When to Use

- Applications that need to serve IPv6 clients directly (mobile apps, IoT devices)
- Environments migrating toward dual-stack networking
- Regulatory or compliance scenarios requiring IPv6 support

## Key Configuration Choices

- **IPv6 enabled** (`enableIpv6: true`) -- Allocates an IPv6 /64 block from the parent VPC's /56 allocation. Requires the parent VPC to have IPv6 enabled.
- **IPv6 CIDR block mask** (`ipv6CidrBlockMask: 42`) -- Selects a specific /64 segment from the VPC's IPv6 allocation. Valid range: 0-255. Adjust to avoid overlap with other dual-stack VSwitches in the same VPC.
- **172.16.x.x range** (`cidrBlock: "172.16.0.0/24"`) -- Uses the middle private range. Adjust to match your VPC's CIDR allocation.
- **Tagged as dual-stack** (`networkType: dual-stack`) -- Makes it easy to filter dual-stack VSwitches from IPv4-only ones.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<alibaba-cloud-region>` | Alibaba Cloud region code (e.g., `cn-hangzhou`, `cn-shanghai`) | Your deployment region |
| `<your-vpc-id>` | VPC ID with IPv6 enabled | Alibaba Cloud VPC console or `AlicloudVpc` stack outputs |
| `<availability-zone>` | Availability zone within the region (e.g., `cn-hangzhou-a`) | Alibaba Cloud ECS console > Zones |
| `<your-ipv6-vswitch-name>` | VSwitch name (1-128 characters) | Choose a descriptive name |

## Related Presets

- **01-dev-single-zone** -- Use for IPv4-only development environments
- **02-prod-app-tier** -- Use for IPv4-only production workloads with larger address space
