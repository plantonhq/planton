---
title: "Public Subnet"
description: "This preset creates a public subnet with an inline route rule that sends all traffic through an Internet Gateway. VNICs in this subnet can be assigned public IP addresses and receive inbound internet..."
type: "preset"
rank: "02"
presetSlug: "02-public"
componentSlug: "subnet"
componentTitle: "Subnet"
provider: "oci"
icon: "package"
order: 2
---

# Public Subnet

This preset creates a public subnet with an inline route rule that sends all traffic through an Internet Gateway. VNICs in this subnet can be assigned public IP addresses and receive inbound internet traffic. This is the standard configuration for internet-facing resources such as load balancers, bastion hosts, and API gateways.

## When to Use

- Load balancer subnets that receive inbound traffic from the internet
- Bastion host subnets for SSH/RDP access into private resources
- API gateway subnets that serve external clients
- Any resource that needs a public IP address and direct internet connectivity

## Key Configuration Choices

- **Public IPs allowed** (`prohibitPublicIpOnVnic: false`) -- VNICs in this subnet can be assigned public IP addresses, making resources directly reachable from the internet.
- **Internet ingress allowed** (`prohibitInternetIngress: false`) -- Inbound internet traffic is permitted, subject to security lists and NSG rules. Use NSGs to control which ports and protocols are open.
- **Internet Gateway route** (`destination: 0.0.0.0/0` via Internet Gateway) -- All non-local traffic routes through the VCN's Internet Gateway for both inbound and outbound connectivity.
- **CIDR block** (`cidrBlock: 10.0.0.0/24`) -- 256 addresses in the first /24 block of a /16 VCN. Public subnets typically need fewer addresses than private subnets since they host fewer resources (load balancers, bastions).
- **DNS label** (`dnsLabel: pub1`) -- Enables VCN-internal DNS so resources get hostnames like `<instance>.pub1.<vcn-dns-label>.oraclevcn.com`.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment where the subnet will be created | OCI Console > Identity > Compartments, or `OciCompartment` status outputs |
| `<vcn-ocid>` | OCID of the VCN this subnet belongs to | OCI Console > Networking > VCNs, or `OciVcn` status outputs (`vcnId`) |
| `<internet-gateway-ocid>` | OCID of the Internet Gateway attached to the VCN | `OciVcn` status outputs (`internetGatewayId`) |

## Related Presets

- **01-private** -- Use instead for subnets hosting backend workloads that must not be publicly reachable
- **03-development** -- Use instead for dev/test environments where custom route rules are unnecessary
