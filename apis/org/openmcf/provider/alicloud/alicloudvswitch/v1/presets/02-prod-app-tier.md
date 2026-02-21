# Production Application Tier VSwitch

This preset creates a VSwitch sized for production application workloads using a /20 CIDR from the 10.x.x.x range. The larger address space (4,096 IPs) accommodates Kubernetes node pools, ECS fleets, or other workloads that scale horizontally. Tags identify the VSwitch's role in the network topology.

## When to Use

- Production application tier hosting Kubernetes worker nodes (ACK) or ECS instances
- Workloads that need room to scale horizontally within a single availability zone
- Environments where cost-tracking and organizational tags are required

## Key Configuration Choices

- **Large CIDR** (`cidrBlock: "10.1.0.0/20"`) -- 4,096 addresses, sized for production workloads that may scale to hundreds of instances. Adjust the third octet to avoid overlap with other VSwitches in the same VPC.
- **10.x.x.x range** -- Standard production VPC range with ample room for multi-tier subnetting
- **Organizational tags** (`tier: application`, `team: platform`) -- Enables cost allocation and resource filtering. Replace with your own tag taxonomy.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<alibaba-cloud-region>` | Alibaba Cloud region code (e.g., `cn-hangzhou`, `cn-shanghai`) | Your deployment region |
| `<your-vpc-id>` | VPC ID that this VSwitch belongs to | Alibaba Cloud VPC console or `AliCloudVpc` stack outputs |
| `<availability-zone>` | Availability zone within the region (e.g., `cn-hangzhou-b`) | Alibaba Cloud ECS console > Zones |
| `<your-prod-vswitch-name>` | VSwitch name (1-128 characters) | Choose a descriptive name |

## Related Presets

- **01-dev-single-zone** -- Use for development environments with smaller address space
- **03-ipv6-enabled** -- Use when dual-stack networking is required in production
