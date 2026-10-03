# GcpGkeFleetMembership — Pulumi Implementation

This directory contains the Pulumi implementation for an explicit fleet membership from the Planton spec: `gcp.projects.Service` and one `gcp.gkehub.Membership`.

## File Organization

| File | Purpose |
|------|---------|
| `main.go` | Module entry point; invokes `Resources()` which wires locals, provider, and `membership` |
| `module/locals.go` | IaC input, attribution labels, the label merge |
| `module/membership.go` | API enablement, the membership, the resource-link helper, the outputs |
| `module/outputs.go` | Output key constants (`name`, `membership_id`, `location`) |

## Send Posture (parity with Terraform)

- **`MembershipId`** / **`Location`** -- `metadata.name` and `global` when unset.
- **`Endpoint.GkeCluster.ResourceLink`** -- the `//container.googleapis.com/` form, composed by `gkeClusterResourceLink`.
- **`Authority.Issuer`** -- sent only when set.

## Usage

```shell
planton pulumi up --manifest <manifest.yaml>
```
