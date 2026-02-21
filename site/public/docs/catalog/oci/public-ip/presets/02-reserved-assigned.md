---
title: "Reserved Assigned Public IP"
description: "This preset allocates a reserved public IP and immediately assigns it to an existing private IP on a VNIC. The IP persists across instance reboots and can be reassigned to a different private IP..."
type: "preset"
rank: "02"
presetSlug: "02-reserved-assigned"
componentSlug: "public-ip"
componentTitle: "Public IP"
provider: "oci"
icon: "package"
order: 2
---

# Reserved Assigned Public IP

This preset allocates a reserved public IP and immediately assigns it to an existing private IP on a VNIC. The IP persists across instance reboots and can be reassigned to a different private IP later without releasing the address. Use this when you need a stable public address bound to a specific compute instance or network interface at creation time.

## When to Use

- Assigning a stable public IP to a compute instance that serves as a bastion host or VPN endpoint
- Binding a known IP to a NAT instance or network virtual appliance
- Replacing an ephemeral public IP on an existing instance with a reserved one for long-term stability
- Any resource that needs a public IP immediately upon provisioning

## Key Configuration Choices

- **RESERVED lifetime** (`lifetime: RESERVED`) -- the IP is persistent and region-scoped. Even if the target instance is terminated, the IP remains in the compartment and can be reassigned.
- **Assigned to a private IP** (`privateIpId`) -- the public IP is mapped to the specified private IP at creation time. The private IP must be a primary private IP on a VNIC in the same region.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment where the public IP will be created | OCI Console > Identity > Compartments, or `OciCompartment` outputs |
| `<private-ip-ocid>` | OCID of the private IP to assign the public IP to | OCI Console > Compute > Instance > Attached VNICs > Primary VNIC > IP Addresses, or via `oci network private-ip list` CLI |

## Related Presets

- **01-reserved-unassigned** -- Use instead when the target resource does not exist yet and the IP should be allocated ahead of time
