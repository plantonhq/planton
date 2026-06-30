# Private Subnet

This preset creates a private subnet with inline route rules that direct internet-bound traffic through a NAT Gateway and OCI service traffic through a Service Gateway. No VNICs in this subnet can have public IP addresses, and inbound internet traffic is blocked. This is the standard subnet configuration for production workloads -- worker nodes, databases, application backends, and any resource that should not be directly internet-accessible.

## When to Use

- Production workloads where resources must not be publicly reachable (OKE worker nodes, databases, internal APIs)
- Any subnet backing private compute instances that still need outbound internet access (OS patching, container image pulls, external API calls)
- Regulated environments that mandate zero public IP exposure
- Backend tiers in multi-tier architectures (the private half of a public/private VCN topology)

## Key Configuration Choices

- **Public IP prohibited** (`prohibitPublicIpOnVnic: true`) -- VNICs in this subnet cannot be assigned public IPs, even if explicitly requested. This is the strongest enforcement of private-only networking.
- **Internet ingress blocked** (`prohibitInternetIngress: true`) -- Blocks all inbound internet traffic to VNICs regardless of security rules or NSG configuration. Defense in depth on top of the public IP prohibition.
- **NAT Gateway route** (`destination: 0.0.0.0/0` via NAT Gateway) -- Allows private instances to initiate outbound internet connections without being publicly addressable. Required for package updates, image pulls, and external API calls.
- **Service Gateway route** (`destination: all-iad-services-in-oracle-services-network` via Service Gateway) -- Routes traffic to OCI services (Object Storage, Container Registry, APM, etc.) over the Oracle backbone network instead of the internet. Improves latency, reduces data transfer costs, and avoids internet exposure.
- **CIDR block** (`cidrBlock: 10.0.1.0/24`) -- 256 addresses in the second /24 block of a /16 VCN. Leaves 10.0.0.0/24 available for a public subnet. Adjust the CIDR to fit your VCN's address plan.
- **DNS label** (`dnsLabel: priv1`) -- Enables VCN-internal DNS so instances get hostnames like `<instance>.priv1.<vcn-dns-label>.oraclevcn.com`.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment where the subnet will be created | OCI Console > Identity > Compartments, or `OciCompartment` status outputs |
| `<vcn-ocid>` | OCID of the VCN this subnet belongs to | OCI Console > Networking > VCNs, or `OciVcn` status outputs (`vcnId`) |
| `<nat-gateway-ocid>` | OCID of the NAT Gateway attached to the VCN | `OciVcn` status outputs (`natGatewayId`) |
| `<service-gateway-ocid>` | OCID of the Service Gateway attached to the VCN | `OciVcn` status outputs (`serviceGatewayId`) |

## Related Presets

- **02-public** -- Use instead for subnets that host internet-facing resources (load balancers, bastion hosts)
- **03-development** -- Use instead for dev/test environments where custom route rules are unnecessary
