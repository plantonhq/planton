# Development Subnet

This preset creates a minimal public subnet with no custom route rules. The subnet inherits the VCN's default route table, which is the simplest possible configuration. This is suitable for development, testing, proof-of-concept work, and ephemeral environments where simplicity and cost matter more than network segmentation.

## When to Use

- Development and testing environments where fine-grained routing is unnecessary
- Proof-of-concept deployments that need basic connectivity with minimal configuration
- Ephemeral environments spun up and torn down frequently
- Learning and experimentation with OCI networking

## Key Configuration Choices

- **Public IPs allowed** (`prohibitPublicIpOnVnic: false`) -- Instances can have public IPs for direct SSH access during development, eliminating the need for a bastion host.
- **Internet ingress allowed** (`prohibitInternetIngress: false`) -- Inbound traffic is permitted, subject to security list rules. Convenient for dev where you need to reach services directly.
- **No custom route rules** -- The subnet uses the VCN's default route table. When paired with the development VCN preset (which enables only an Internet Gateway), the default route table already routes internet traffic correctly.
- **CIDR block** (`cidrBlock: 10.0.0.0/24`) -- 256 addresses. For dev environments a single subnet is often sufficient.
- **DNS label** (`dnsLabel: dev1`) -- Enables VCN-internal DNS, keeping the experience consistent with production even in dev environments.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment where the subnet will be created | OCI Console > Identity > Compartments, or `OciCompartment` status outputs |
| `<vcn-ocid>` | OCID of the VCN this subnet belongs to | OCI Console > Networking > VCNs, or `OciVcn` status outputs (`vcnId`) |

## Related Presets

- **01-private** -- Use instead for production workloads that must not be publicly reachable
- **02-public** -- Use instead when you need explicit route rules for internet-facing production resources
