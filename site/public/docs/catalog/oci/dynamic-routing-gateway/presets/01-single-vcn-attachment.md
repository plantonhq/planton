---
title: "Single VCN Attachment"
description: "This preset creates a Dynamic Routing Gateway with a single VCN attachment. It is the most common DRG starting point, enabling inter-VCN routing and serving as the prerequisite for adding..."
type: "preset"
rank: "01"
presetSlug: "01-single-vcn-attachment"
componentSlug: "dynamic-routing-gateway"
componentTitle: "Dynamic Routing Gateway"
provider: "oci"
icon: "package"
order: 1
---

# Single VCN Attachment

This preset creates a Dynamic Routing Gateway with a single VCN attachment. It is the most common DRG starting point, enabling inter-VCN routing and serving as the prerequisite for adding Site-to-Site VPN (IPSec), FastConnect, or remote peering connections later. No custom route tables or distributions are defined; OCI auto-generates default route tables per network type.

## When to Use

- Connecting a VCN to a DRG as a prerequisite for IPSec VPN or FastConnect
- Enabling future VCN peering by establishing the DRG hub
- Simple single-VCN environments that need a routing gateway for on-premises connectivity
- Starting point before evolving into a hub-and-spoke topology

## Key Configuration Choices

- **Single VCN attachment** (`attachments` with one entry) -- attaches the specified VCN to the DRG. OCI automatically creates a default route table for VCN attachments with routes imported from the VCN's CIDR blocks.
- **No custom route tables** -- the DRG uses OCI's auto-generated default route tables, which import routes from attached networks automatically. Custom tables are only needed for hub-and-spoke or transit routing.
- **No route distributions** -- OCI creates a default import distribution per attachment type and a default export distribution. For a single VCN attachment, these defaults provide correct bidirectional routing without additional configuration.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment where the DRG will be created | OCI Console > Identity > Compartments, or `OciCompartment` outputs |
| `<vcn-ocid>` | OCID of the VCN to attach to the DRG | OCI Console > Networking > Virtual Cloud Networks, or `OciVcn` outputs |

## Related Presets

- **02-hub-and-spoke** -- Use instead when connecting multiple VCNs through the DRG with custom route tables and route distributions for inter-VCN traffic control
