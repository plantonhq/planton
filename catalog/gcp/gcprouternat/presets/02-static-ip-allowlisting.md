# Static-IP Allowlisted Egress

This preset creates a NAT gateway whose egress IPs are referenced `GcpAddress` reservations — stable addresses a partner, payment processor, or compliance regime can allowlist — scoped to the workload subnetwork with a raised per-VM port floor.

## When to Use

- Third parties (banks, payment APIs, partner networks) require a fixed list of source IPs
- Compliance regimes that require documented, stable egress addresses
- Connection-heavy services that exhaust the default 64 ports per VM

## Prerequisites

Each entry in `natIps` references an existing reservation:

1. A `GcpAddress` per egress IP, `addressType: EXTERNAL`, in the same region as the NAT

## Key Configuration Choices

- **Addresses by reference** — the reservations are first-class nodes with their own lifecycle; their literal IPs are their `address` outputs (that is the list you hand to the partner)
- **Zero-downtime rotation** — add a new `GcpAddress` to `natIps`, then move the old entry to `drainNatIps`: established connections finish, new connections use the remaining IPs
- **Subnetwork scoping** — only the listed workload subnetwork gets NAT; other subnets in the region stay dark
- **`minPortsPerVm: 128`** — double the default port floor for services that fan out many concurrent connections

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `my-gcp-project-123` | GCP project ID | GCP Console or `GcpProject` outputs |
| `my-prod-vpc` | Your `GcpVpcNetwork` resource name | Your VPC manifest |
| `egress-ip-a` / `egress-ip-b` | Your `GcpAddress` resource names | Your address manifests |
| `my-workload-subnet` | Your `GcpSubnetwork` resource name | Your subnetwork manifest |

## Related Presets

- **01-all-subnets-auto** — when egress IPs do not need to be stable
- **03-private-nat** — NAT between VPC networks (Network Connectivity Center spokes)

## Related Kinds

- [GcpAddress](/docs/catalog/gcp/gcpaddress) — the reserved egress IPs this preset references
- [GcpSubnetwork](/docs/catalog/gcp/gcpsubnetwork) — the scoped workload subnetwork
