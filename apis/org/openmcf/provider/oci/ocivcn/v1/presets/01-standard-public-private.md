# Standard Public-Private VCN

This preset creates a production VCN with all three gateways enabled: Internet Gateway, NAT Gateway, and Service Gateway. This is the standard OCI networking topology for workloads that require both public-facing and private subnets, covering 80%+ of production deployments including OKE clusters, web applications, and hybrid architectures.

## When to Use

- Production workloads that need both public and private subnets
- OKE (Kubernetes) clusters where the API endpoint is public but worker nodes are private
- Web applications with public load balancers backed by private compute instances
- Any environment that needs outbound internet from private subnets (via NAT) and private access to OCI services (via Service Gateway)

## Key Configuration Choices

- **All three gateways enabled** -- Internet Gateway for public subnets, NAT Gateway for private outbound, Service Gateway for private OCI service access. This is the standard production topology that supports public/private subnet architectures.
- **Single /16 CIDR** (`cidrBlocks: ["10.0.0.0/16"]`) -- provides 65,536 addresses, sufficient for most deployments. OCI supports adding additional non-overlapping CIDRs later if needed.
- **DNS label set** (`dnsLabel: prodvcn`) -- enables VCN-internal DNS resolution so instances get hostnames like `<instance>.prodvcn.oraclevcn.com`. Must be alphanumeric, start with a letter, max 15 characters.
- **IPv6 disabled** (`isIpv6Enabled: false`) -- most OCI deployments use IPv4 only. IPv6 can be enabled later if needed but cannot be disabled once on.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment where the VCN will be created | OCI Console > Identity > Compartments, or `OciCompartment` outputs |

## Related Presets

- **02-private-only** -- Use instead when no resources should be directly internet-accessible (no Internet Gateway)
- **03-development** -- Use instead for dev/test environments where NAT and Service Gateway costs are unnecessary
