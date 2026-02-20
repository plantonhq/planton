# Reserved Unassigned Public IP

This preset allocates a reserved public IP without assigning it to any resource. The IP persists independently of any compute instance or load balancer, making it suitable for pre-provisioning stable addresses for DNS records, firewall allowlists, or partner integrations before the target infrastructure exists.

## When to Use

- Pre-allocating a static IP for DNS A-records before deploying the target resource
- Reserving IPs for firewall allowlists or partner API integrations that require known addresses
- Creating a pool of unassigned IPs for manual or automated assignment to future resources
- Any scenario where the IP must outlive the resource it is attached to

## Key Configuration Choices

- **RESERVED lifetime** (`lifetime: RESERVED`) -- the IP persists until explicitly deleted, surviving instance termination and VNIC detachment. This is the correct choice for any IP that needs to be stable across resource lifecycle events.
- **No private IP assignment** -- the IP is created unassigned. It can be attached to a private IP later via the OCI Console, API, or by updating this manifest with `privateIpId`.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment where the public IP will be created | OCI Console > Identity > Compartments, or `OciCompartment` outputs |

## Related Presets

- **02-reserved-assigned** -- Use instead when the target private IP is already known and the IP should be bound at creation time
