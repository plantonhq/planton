# Private-Only VCN

This preset creates a security-hardened VCN with no Internet Gateway. NAT Gateway and Service Gateway provide outbound connectivity, but no resources in this VCN are directly reachable from the public internet. This is the standard topology for regulated industries, backend processing, and environments that require zero public exposure.

## When to Use

- Regulated environments (finance, healthcare) where no resource should have a public IP
- Private OKE clusters accessed via bastion hosts, VPN, or FastConnect
- Database VCNs that only need outbound access for patching and OCI service calls
- Backend processing workloads that pull data from external sources but never serve inbound traffic

## Key Configuration Choices

- **No Internet Gateway** (`isInternetGatewayEnabled: false`) -- zero resources in this VCN can receive direct inbound internet traffic. Access to resources is via bastion, VPN, DRG peering, or FastConnect.
- **NAT Gateway enabled** (`isNatGatewayEnabled: true`) -- private subnets can initiate outbound internet connections (OS patching, pulling container images, calling external APIs) without being publicly addressable.
- **Service Gateway enabled** (`isServiceGatewayEnabled: true`) -- traffic to OCI services (Object Storage, Container Registry, APM, etc.) stays on the Oracle backbone network and never traverses the internet.
- **Single /16 CIDR** (`cidrBlocks: ["10.0.0.0/16"]`) -- standard address space. All subnets in this VCN will be private.
- **DNS label set** (`dnsLabel: privvcn`) -- enables VCN-internal DNS resolution for service discovery between private resources.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment where the VCN will be created | OCI Console > Identity > Compartments, or `OciCompartment` outputs |

## Related Presets

- **01-standard-public-private** -- Use instead when you need both public and private subnets (e.g., public load balancers fronting private compute)
- **03-development** -- Use instead for low-cost dev/test environments where gateway charges are a concern
