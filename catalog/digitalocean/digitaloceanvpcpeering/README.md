# DigitalOcean VPC Peering

Built for 100% parity with the Terraform DigitalOcean provider's `digitalocean_vpc_peering` resource at the pinned provider version.

## What this component models

A private-network peering connection between exactly two DigitalOcean VPCs, letting resources in both networks reach each other over DigitalOcean's private fabric -- no public internet path, no VPN.

- `peering_name` -- the connection's name (the ONLY field that updates in place)
- `vpc_1` / `vpc_2` -- the two VPCs, wired by reference (or literal VPC UUIDs); exactly two by construction, and changing either replaces the peering

## Quick start

```yaml
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanVpcPeering
metadata:
  name: app-to-data-peering
spec:
  peeringName: app-to-data
  vpc_1:
    valueFrom:
      kind: DigitalOceanVpc
      name: vpc-app
      fieldPath: status.outputs.vpc_id
  vpc_2:
    valueFrom:
      kind: DigitalOceanVpc
      name: vpc-data
      fieldPath: status.outputs.vpc_id
```

Deploy with either provisioner; both produce identical resources and outputs.

## Outputs

| Output | Description |
|---|---|
| `peering_id` | UUID of the peering connection (its API identity and import id) |

The peering's lifecycle status is not an output: both provisioners wait for ACTIVE before the apply succeeds, so a stored value could only ever say ACTIVE and would go stale once DigitalOcean moved the peering. Read it live from the control panel's Peering connections tab or `GET /v2/vpcs/peerings/{id}` (statuses are UPPERCASE: PROVISIONING, ACTIVE, DELETING).

## Behavior worth knowing

- **Peered VPCs must not overlap.** DigitalOcean rejects peerings between VPCs with overlapping IP ranges -- plan CIDRs before you need to peer.
- **The peering is symmetric and all-or-nothing.** Every resource in each VPC can reach every resource in the other; there is no route filtering on the peering itself (droplet firewalls do that per host).
- **Creates and deletes are settled, not instant.** The provider waits for ACTIVE on create and retries the delete through DigitalOcean's transient 403s while the peering settles (2-minute windows).

## Module layout

- `iac/tf/` -- OpenTofu/Terraform module (provider pinned `~> 2.99`)
- `iac/pulumi/` -- Pulumi module (Go, pulumi-digitalocean SDK)
- Both engines wire the same spec fields and export the same outputs; behavioral parity is the contract.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
