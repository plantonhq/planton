# Development VCN

This preset creates a minimal-cost VCN with only an Internet Gateway. NAT Gateway and Service Gateway are omitted to avoid their hourly charges. This is the simplest VCN configuration, suitable for development, testing, proof-of-concept work, and ephemeral environments where cost matters more than network segmentation.

## When to Use

- Development and testing environments with budget constraints
- Proof-of-concept deployments that need basic internet connectivity
- Ephemeral environments spun up and torn down frequently
- Learning and experimentation with OCI networking

## Key Configuration Choices

- **Internet Gateway only** (`isInternetGatewayEnabled: true`) -- provides basic inbound and outbound internet connectivity via public subnets. Instances with public IPs can be reached directly.
- **No NAT Gateway** (`isNatGatewayEnabled: false`) -- private subnets cannot reach the internet. If outbound access is needed, use public subnets instead. NAT Gateways incur hourly charges that are unnecessary for dev environments.
- **No Service Gateway** (`isServiceGatewayEnabled: false`) -- OCI service traffic (Object Storage, Container Registry) will traverse the internet instead of the Oracle backbone. For dev workloads, this is acceptable. Service Gateways incur hourly charges.
- **Single /16 CIDR** (`cidrBlocks: ["10.0.0.0/16"]`) -- standard address space, same as production. Using the same CIDR range across environments simplifies subnet design and makes it easier to promote configurations from dev to production.
- **DNS label set** (`dnsLabel: devvcn`) -- enables VCN-internal DNS even in dev environments, keeping the experience consistent with production.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment where the VCN will be created | OCI Console > Identity > Compartments, or `OciCompartment` outputs |

## Related Presets

- **01-standard-public-private** -- Use instead for production workloads that need NAT and Service Gateway
- **02-private-only** -- Use instead for security-hardened environments with no public internet exposure
