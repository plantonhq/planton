# OciBastion

## Overview

OciBastion is an Planton component that deploys an OCI Bastion service instance. It provides a single declarative manifest to create a managed SSH gateway with configurable client CIDR restrictions, session TTL limits, and optional DNS proxy support.

## Purpose

Private subnets in OCI do not have internet-reachable IPs. Operators need a way to SSH into compute instances, database systems, or other resources in private subnets for troubleshooting and maintenance. The OCI Bastion service provides a managed, auditable SSH gateway that eliminates the need to maintain a self-hosted jump box. This component provisions the bastion infrastructure; sessions are created separately via the OCI CLI or Console.

## Key Features

- **Managed SSH gateway** — OCI handles the bastion lifecycle, patching, and availability.
- **CIDR-based access control** — restrict which client IP ranges can establish sessions.
- **Session TTL limits** — enforce a maximum session duration (default 3 hours, configurable up to 3 hours).
- **DNS proxy support** — optional FQDN resolution and SOCKS5 proxy for sessions that target hosts by DNS name.
- **Foreign key references** — `compartmentId` and `targetSubnetId` support `valueFrom` to reference Planton-managed resources.

## Constraints

- `targetSubnetId` is immutable after creation — changing it forces recreation.
- `displayName` is immutable after creation.
- `isDnsProxyEnabled` is immutable after creation.
- `bastionType` is hardcoded to `STANDARD` (the only publicly documented type).
- Sessions (`oci_bastion_session`) are NOT managed by this component — they are ephemeral operational artifacts.

## Use Cases

| Scenario | Configuration |
|----------|---------------|
| Development access to private instances | Minimal bastion with no CIDR restrictions |
| Corporate network access only | `clientCidrBlockAllowList` restricted to corporate CIDRs |
| Extended maintenance windows | `maxSessionTtlInSeconds` set to 28800 (8 hours) |
| FQDN-based target resolution | `isDnsProxyEnabled: true` for DNS name access |

## Production Features

- **Freeform tags** — automatically populated from `metadata.labels`, including `resource_kind`, `resource_id`, `organization`, and `environment`.
- **Audit trail** — OCI Bastion sessions are logged in the OCI Audit service for compliance.
- **Network isolation** — the bastion endpoint lives inside the target subnet; no public IP is exposed.
