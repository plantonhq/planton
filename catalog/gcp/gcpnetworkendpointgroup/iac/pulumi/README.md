# GcpNetworkEndpointGroup — Pulumi Implementation

This directory contains the Pulumi implementation for provisioning a
network endpoint group from the Planton spec -- zonal when `spec.zone` is
set, global when it is empty -- together with its membership.

## File Organization

| File | Purpose |
|------|---------|
| `main.go` | Module entry point; invokes `Resources()` which wires locals, provider, and the arm `spec.zone` selects |
| `module/locals.go` | Ambient project fallback, group name (spec or metadata.name fallback), the `IsZonal` scope selector |
| `module/zonal_network_endpoint_group.go` | The ZONAL arm: `gcp.compute.NetworkEndpointGroup` plus the bulk `gcp.compute.NetworkEndpointList` (created only when the list is non-empty) |
| `module/global_network_endpoint_group.go` | The GLOBAL arm: `gcp.compute.GlobalNetworkEndpointGroup` plus one `gcp.compute.GlobalNetworkEndpoint` per member, named by list position |
| `module/outputs.go` | Output key constants (`self_link`, `neg_name`, `neg_id`, `zone`, `size`) |

## Send Posture (parity with Terraform)

- **`NetworkEndpointType`** -- left nil on the zonal arm when the spec is
  empty (Google defaults it to `GCE_VM_IP_PORT`); required and always
  sent on the global arm.
- **`DefaultPort`, endpoint `Port`** -- sent only when set; a global
  endpoint's port falls back to `default_port`.
- **`Instance`, `IpAddress`, `Fqdn`** -- sent only when non-empty; the
  spec has already matched each to the group's type.
- **`DeletionPolicy`** -- DELETE (default), PREVENT, or ABANDON, fanned out
  to the endpoint resources; sent only when set on both engines.
- **`neg_id`** -- the zonal group's `GeneratedId` as a string; empty on
  the global arm, which exposes none. **`size`** is the declared count on
  both arms, the shape the Terraform module exports.
