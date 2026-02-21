# Development Single-Zone VSwitch

This preset creates a minimal VSwitch in a single availability zone using a small /24 CIDR from the 192.168.x.x range. It omits tags and IPv6 configuration, keeping the setup simple and inexpensive for development and testing.

## When to Use

- Development or sandbox environments where high availability is not required
- Quick proof-of-concept deployments in a single AZ
- Lightweight test environments for a small number of ECS instances or containers

## Key Configuration Choices

- **Small CIDR** (`cidrBlock: "192.168.0.0/24"`) -- 256 addresses, sufficient for development workloads without wasting IP space
- **192.168.x.x range** -- Assumes a development VPC using 192.168.0.0/16, keeping it separate from production 10.x.x.x ranges
- **No tags** -- Development resources are typically short-lived and don't need cost-tracking metadata
- **No IPv6** -- Keeps the configuration minimal; enable via the 03-ipv6-enabled preset if needed

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<alibaba-cloud-region>` | Alibaba Cloud region code (e.g., `cn-hangzhou`, `ap-southeast-1`) | Your deployment region |
| `<your-vpc-id>` | VPC ID that this VSwitch belongs to | Alibaba Cloud VPC console or `AlicloudVpc` stack outputs |
| `<availability-zone>` | Availability zone within the region (e.g., `cn-hangzhou-a`) | Alibaba Cloud ECS console > Zones |
| `<your-dev-vswitch-name>` | VSwitch name (1-128 characters) | Choose a descriptive name |

## Related Presets

- **02-prod-app-tier** -- Use for production workloads with a larger CIDR and organizational tags
- **03-ipv6-enabled** -- Use when dual-stack networking is required
