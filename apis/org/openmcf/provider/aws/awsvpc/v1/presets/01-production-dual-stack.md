# Production Dual-Stack VPC

This preset creates a production-ready VPC with a `/16` IPv4 CIDR (65,536
addresses) and an Amazon-provided IPv6 `/56`, with DNS support and DNS hostnames
enabled. It is the standard foundation for a production AWS network: deploy
`AwsSubnet`, `AwsInternetGateway`, and `AwsNatGateway` components that reference
this VPC to build out the topology.

## When to Use

- Production workloads that want dual-stack (IPv4 + IPv6) networking
- The foundation for EKS clusters, ECS services, RDS instances, and other AWS
  resources, composed via subnets that reference this VPC
- Any deployment that needs Route 53 private hosted zones or service discovery
  (which require DNS hostnames)

## Key Configuration Choices

- **/16 IPv4 CIDR** (`cidrBlock: 10.0.0.0/16`) -- 65,536 IPs; standard production
  sizing that leaves room for many subnets
- **Amazon-provided IPv6** (`assignGeneratedIpv6CidrBlock: true`) -- a /56 IPv6
  block for dual-stack subnets
- **DNS hostnames** (`enableDnsHostnames: true`) -- required for Route 53 private
  hosted zones, service discovery, and VPC endpoints
- **DNS support** (`enableDnsSupport: true`) -- Amazon-provided DNS resolution
  within the VPC (on by default; set here for explicitness)

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<aws-region>` | AWS region code (e.g., `us-east-1`, `eu-west-1`) | Your deployment region |

## Related Presets

- **02-development** -- a minimal single-CIDR VPC for development environments
