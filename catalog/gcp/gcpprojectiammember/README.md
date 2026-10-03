# GCP Project IAM Member

Deploys a single ADDITIVE project-level IAM grant (`google_project_iam_member`) — one role, to one member, on one project. This is the safe, composable unit of GCP access control: the grant merges into the project's IAM policy without touching any other member's bindings, and removal subtracts only this exact pair.

## What Gets Created

When you deploy a GcpProjectIamMember resource, Planton provisions:

- **Project IAM Member** — one (role, member[, condition]) entry merged into the target project's IAM policy

Nothing else in the policy is read as owned or modified — grants made by other charts, teams, or tools are never clobbered.

## Prerequisites

- **GCP credentials** configured via environment variables or Planton provider config
- **An existing GCP project** — referenced via `projectId`
- **IAM permissions** — see [`iac/permissions.yaml`](iac/permissions.yaml) for the least-privilege permission set the deploying principal needs
- **The role and the member must exist** — reference a GcpIamCustomRole and/or GcpServiceAccount, or use literal values

## Quick Start

Create a file `iam-grant.yaml`:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpProjectIamMember
metadata:
  name: app-logs-writer-grant
spec:
  projectId:
    value: my-gcp-project-123
  role:
    value: roles/logging.logWriter
  member:
    value: serviceAccount:my-app@my-gcp-project-123.iam.gserviceaccount.com
```

Deploy:

```shell
planton apply -f iam-grant.yaml
```

Or compose with the identity and role as first-class nodes:

```yaml
spec:
  role:
    valueFrom:
      kind: GcpIamCustomRole
      name: log-bucket-writer
      fieldPath: status.outputs.name
  member:
    valueFrom:
      kind: GcpServiceAccount
      name: my-app-identity
      fieldPath: status.outputs.member
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `role` | `StringValueOrRef` | The role to grant: a predefined role (`roles/storage.objectViewer`) or a custom role's fully-qualified name. References a GcpIamCustomRole's `name` output by default. Immutable. |
| `member` | `StringValueOrRef` | The identity receiving the grant, in IAM member format. References a GcpServiceAccount's `member` output by default. Immutable. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider default project | The project whose IAM policy receives the grant. Can reference a GcpProject. Immutable. |
| `condition` | `object` | — | IAM Condition (`title`, `expression`, optional `description`) restricting when the grant applies. Part of the grant's identity. Immutable. |

### Member Formats

| Format | Grants to |
|--------|-----------|
| `serviceAccount:<email>` | A service account (the most common in IaC) |
| `user:<email>` | A Google account |
| `group:<email>` | A Google group |
| `domain:<domain>` | Everyone in a Workspace/Cloud Identity domain |
| `principal://...` / `principalSet://...` | Workload identity federation principals |
| `allUsers` / `allAuthenticatedUsers` | Public access — grant with extreme care |

Grants to deleted principals (`deleted:...`) are rejected at deploy time.

## Additive vs Authoritative

GCP offers three write modes for project IAM: additive member (this component), authoritative per-role binding, and authoritative whole-policy. Only the additive member is safe for composition — the authoritative modes clobber every grant they do not list, so two independent tools managing the same role would silently remove each other's access. Planton deliberately models only the additive grant.

## Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `project_id` | `string` | The project whose policy received the grant (after reference resolution) |
| `role` | `string` | The granted role (after reference resolution) |
| `member` | `string` | The granted member (after reference resolution) |
| `etag` | `string` | The project IAM policy etag after the grant — useful for audit correlation |

## Deployment Methods

Planton supports two deployment methods:

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Everything is immutable**: IAM grants have no update — changing role, member, project, or condition replaces the grant atomically (destroy the old pair, create the new one), which mirrors the underlying API exactly.
- **Conditions are part of the grant's identity**: the same role granted with and without a condition are two independent grants that do not interfere.
- **Concurrent policy writes are serialized** per project by the provider, so many grants deploying in parallel converge safely.

## Related Kinds

- [GcpIamCustomRole](/docs/catalog/gcp/gcpiamcustomrole) — defines the custom role this grant can reference
- [GcpServiceAccount](/docs/catalog/gcp/gcpserviceaccount) — the identity most commonly granted (its `member` output feeds this component)
- [GcpProject](/docs/catalog/gcp/gcpproject) — provides the project whose policy receives the grant

## Additional Resources

- [IAM Overview](https://cloud.google.com/iam/docs/overview)
- [IAM Conditions](https://cloud.google.com/iam/docs/conditions-overview)
- [Principal Identifiers](https://cloud.google.com/iam/docs/principal-identifiers)

## Support

For issues, questions, or contributions, please refer to the Planton documentation or open an issue in the repository.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
