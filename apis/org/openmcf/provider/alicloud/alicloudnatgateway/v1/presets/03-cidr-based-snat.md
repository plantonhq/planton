# CIDR-Based SNAT NAT Gateway

This preset creates a NAT Gateway with SNAT entries specified by CIDR blocks instead of VSwitch IDs. This provides fine-grained control over which IP ranges get outbound internet access.

## When to Use

- When you want to NAT only a subset of a VSwitch's address space
- Shared VSwitches where only specific IP ranges should have internet access
- Security-sensitive environments requiring explicit IP range whitelisting for outbound NAT

## Key Configuration Choices

- **sourceCidr instead of sourceVswitchId** -- allows partial VSwitch NAT or custom CIDR targeting. These two fields are mutually exclusive in the provider.
- **Named SNAT entries** -- clear names for each CIDR rule for operational visibility

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<alibaba-cloud-region>` | Alibaba Cloud region code (e.g., `cn-hangzhou`) | Your deployment region |
| `<your-vpc-id>` | VPC ID the NAT Gateway belongs to | Alibaba Cloud VPC console or `AliCloudVpc` stack outputs |
| `<your-nat-vswitch-id>` | VSwitch ID for NAT Gateway placement | Alibaba Cloud VPC console or `AliCloudVswitch` stack outputs |
| `<your-eip-id>` | EIP allocation ID to associate with the NAT Gateway | Alibaba Cloud EIP console or `AliCloudEipAddress` stack outputs |
| `<your-nat-name>` | NAT Gateway name (2-128 chars) | Choose a descriptive name |

Adjust the `sourceCidr` values (`10.0.1.0/24`, `10.0.2.0/24`) to match your actual VPC CIDR allocation.

## Related Presets

- **01-single-vswitch** -- Simpler setup using VSwitch ID for the common case
- **02-multi-az-production** -- Production multi-AZ with VSwitch-based SNAT
