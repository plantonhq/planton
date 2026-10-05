# GcpGkeFleetScope — Pulumi Implementation

This directory contains the Pulumi implementation for a fleet team scope from the Planton spec: `gcp.projects.Service`, one `gcp.gkehub.Scope`, and the folded `gkehub.Namespace`, `gkehub.ScopeRbacRoleBinding`, and `gkehub.MembershipBinding` resources.

## File Organization

| File | Purpose |
|------|---------|
| `main.go` | Module entry point; invokes `Resources()` which wires locals, provider, and `scope` |
| `module/locals.go` | IaC input, attribution labels, the label merge |
| `module/scope.go` | API enablement, the scope and its children, the membership-name parser, the outputs |
| `module/outputs.go` | Output key constants (`name`, `scope_id`, `uid`) |

## Send Posture (parity with Terraform)

- **Logical names** -- `<name>-namespace-<id>`, `<name>-rbac-<id>`, `<name>-binding-<id>`, from the children's declared IDs.
- **Cluster binding** -- the scope's project; location and membership ID parsed from the membership's full name.
- **Labels** -- attribution labels on every resource's `Labels`, never on `NamespaceLabels`.

## Usage

```shell
planton pulumi up --manifest <manifest.yaml>
```
