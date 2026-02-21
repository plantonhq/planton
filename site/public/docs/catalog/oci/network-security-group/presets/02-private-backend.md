---
title: "Private Backend NSG"
description: "This preset creates a Network Security Group for resources that should only be reachable from within the VCN. All protocols and ports are allowed from the VCN CIDR block, while traffic from outside..."
type: "preset"
rank: "02"
presetSlug: "02-private-backend"
componentSlug: "network-security-group"
componentTitle: "Network Security Group"
provider: "oci"
icon: "package"
order: 2
---

# Private Backend NSG

This preset creates a Network Security Group for resources that should only be reachable from within the VCN. All protocols and ports are allowed from the VCN CIDR block, while traffic from outside the VCN is denied. This is the standard configuration for databases, application servers, internal microservices, and any backend resource that sits behind a load balancer or API gateway.

## When to Use

- Database instances (Autonomous Database, MySQL, PostgreSQL) that should only accept connections from application subnets within the VCN
- Application server instances that receive traffic from a load balancer NSG but should never be directly internet-accessible
- Internal microservices that communicate with each other within the VCN
- Worker nodes in an OKE cluster that receive traffic from the API endpoint and load balancer subnets
- Any resource where zero public inbound exposure is required

## Key Configuration Choices

- **VCN-wide ingress** (`protocol: all` from `10.0.0.0/16`) -- Allows all traffic from any resource within the VCN. This is intentionally broad within the VCN boundary: fine-grained port restrictions between internal services add operational complexity without meaningful security benefit in most architectures, since the VCN perimeter is the trust boundary. Adjust the CIDR to match your VCN's address range if it differs from the default `10.0.0.0/16`.
- **No public ingress** -- There are no rules with `0.0.0.0/0` as the source. Resources attached to this NSG cannot receive traffic from the public internet, regardless of whether they have a public IP.
- **All outbound** (`protocol: all` to `0.0.0.0/0`) -- Backend resources need outbound connectivity for OS patching, container image pulls, OCI service API calls, and DNS resolution. If your security posture requires restricted egress, replace this with specific rules for your outbound targets.
- **Stateful rules** (`stateless` not set, defaults to `false`) -- All rules are stateful, meaning return traffic is automatically allowed. This is the standard for backend workloads.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment where the NSG will be created | OCI Console > Identity > Compartments, or `OciCompartment` status outputs |
| `<vcn-ocid>` | OCID of the VCN this NSG belongs to | OCI Console > Networking > VCNs, or `OciVcn` status outputs (`vcnId`) |

## Related Presets

- **01-web-tier** -- Use instead for internet-facing resources that need HTTP/HTTPS inbound from anywhere
- **03-development** -- Use instead for dev/test environments where all traffic should be permitted including from outside the VCN
