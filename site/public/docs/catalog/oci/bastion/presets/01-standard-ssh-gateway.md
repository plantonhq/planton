---
title: "Standard SSH Gateway"
description: "This preset creates an OCI Bastion with a client CIDR allow list and 3-hour maximum session TTL. The bastion provides secure, time-limited SSH access to compute instances and other resources in a..."
type: "preset"
rank: "01"
presetSlug: "01-standard-ssh-gateway"
componentSlug: "bastion"
componentTitle: "Bastion"
provider: "oci"
icon: "package"
order: 1
---

# Standard SSH Gateway

This preset creates an OCI Bastion with a client CIDR allow list and 3-hour maximum session TTL. The bastion provides secure, time-limited SSH access to compute instances and other resources in a private subnet without requiring public IP addresses on the target resources. Sessions are ephemeral and created on-demand via the OCI Console or CLI.

## When to Use

- SSH access to compute instances in private subnets for administration, debugging, or maintenance
- Port forwarding to private databases (MySQL, PostgreSQL, Oracle) for local client connections
- Secure access to private OKE API servers or other private endpoints without a site-to-site VPN
- Teams that need occasional private-network access without maintaining always-on VPN infrastructure

## Key Configuration Choices

- **Target subnet** (`targetSubnetId`) -- the bastion creates a private endpoint in this subnet. Sessions can reach any resource accessible from this subnet (including resources in peered VCNs if route rules are configured). Choose a subnet that has routes to the resources your team needs to access.
- **Client CIDR allow list** (`clientCidrBlockAllowList`) -- only clients with IP addresses matching these CIDR blocks can create and connect to sessions. Use your corporate VPN range, office IP ranges, or CI/CD runner subnets. This is the primary access control mechanism -- keep it as narrow as possible. Multiple CIDRs can be listed.
- **3-hour max session TTL** (`maxSessionTtlInSeconds: 10800`) -- individual sessions cannot exceed this duration. The OCI default of 3 hours balances convenience (long enough for most administrative tasks) with security (sessions do not persist indefinitely). Reduce to 1800 (30 minutes) for high-security environments or increase to 28800 (8 hours) for extended maintenance windows.
- **Display name is immutable** -- the bastion's display name cannot be changed after creation. Choose a descriptive name that identifies the environment and purpose.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment where the bastion will be created | OCI Console > Identity > Compartments, or `OciCompartment` status outputs |
| `<private-subnet-ocid>` | OCID of the private subnet the bastion connects to | OCI Console > Networking > Subnets, or `OciSubnet` status outputs |
| `<allowed-cidr>` | CIDR block(s) allowed to connect (e.g., `10.0.0.0/16` for VPN, `203.0.113.5/32` for a single IP) | Your network team or VPN provider documentation |

## Related Presets

- **02-dns-proxy-enabled** -- Use instead when sessions need to target resources by FQDN rather than IP address
