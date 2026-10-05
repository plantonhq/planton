# DigitalOcean Reserved IP

Built for 100% parity with the Terraform DigitalOcean provider's reserved-IP family at the pinned provider version: one kind covering `digitalocean_reserved_ip`, `digitalocean_reserved_ipv6`, and `digitalocean_reserved_ipv6_assignment` (the v4 assignment resource is deliberately never created -- v4 assigns through the reservation's own mutable argument).

## What this kind models

A static public IP address reserved in a region -- IPv4 or IPv6 -- optionally assigned to a droplet. The address survives the droplets behind it: re-pointing it is the classic manual-failover move.

- `ip_version` -- `ipv4` (the default) or `ipv6` (create-only)
- `region` -- where the address is reserved; assignment only works to droplets in the same region (create-only)
- `droplet` -- optional assignment target, wired by reference (or a literal numeric droplet id); assign/re-point/unassign all apply in place

## Quick start

```yaml
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanReservedIp
metadata:
  name: web-frontdoor-ip
spec:
  region: nyc3
  droplet:
    valueFrom:
      kind: DigitalOceanDroplet
      name: web-server
      fieldPath: status.outputs.droplet_id
```

Deploy with either provisioner; both produce identical resources and outputs.

## Outputs

| Output | Description |
|---|---|
| `reserved_ip_address` | The reserved address itself (the resource's API identity and import id) |
| `urn` | DigitalOcean uniform resource name (e.g. `do:reservedip:203.0.113.10`), as used by project membership |

## Behavior worth knowing

- **An UNASSIGNED reserved IPv4 bills monthly (prorated hourly); an assigned one is free.** IPv6 reservations are free either way. An idle, forgotten reservation is the expensive state.
- **The address is permanent for the reservation's lifetime and unrecoverable after destroy.** A recreated reservation gets a different address -- update DNS deliberately.
- **v4 and v6 assign differently under the hood** (v4 inline and mutable; v6 through a separate assignment resource because its inline argument is dead upstream) -- the modules absorb this; the spec is identical either way.

## Module layout

- `iac/tf/` -- OpenTofu/Terraform module (provider pinned `~> 2.99`)
- `iac/pulumi/` -- Pulumi module (Go, pulumi-digitalocean SDK)
- Both engines wire the same spec fields and export the same outputs; behavioral parity is the contract.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
