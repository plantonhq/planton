# GCP IAM Custom Role

Deploys a project-scoped IAM custom role (`google_project_iam_custom_role`) — a named, reusable bundle of permissions tailored to exactly what a workload or team needs, instead of granting one of Google's broad predefined roles.

## What Gets Created

When you deploy a GcpIamCustomRole resource, Planton provisions:

- **IAM Custom Role** — a `google_project_iam_custom_role` in the specified project, grantable as `projects/<project>/roles/<role_id>`

No additional supporting resources are created. The role definition is pure metadata — it grants nothing until an IAM grant (e.g. a GcpProjectIamMember) binds it to an identity.

## Prerequisites

- **GCP credentials** configured via environment variables or Planton provider config
- **An existing GCP project** — referenced via `projectId`
- **IAM permissions** — the deploying principal's permissions are listed in [`iac/permissions.yaml`](iac/permissions.yaml)

## Quick Start

Create a file `custom-role.yaml`:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpIamCustomRole
metadata:
  name: log-bucket-writer
spec:
  projectId:
    value: my-gcp-project-123
  roleId: logBucketWriter
  title: Log Bucket Writer
  description: Grants exactly the permissions needed to write log objects
  permissions:
    - storage.objects.create
    - storage.objects.get
```

Deploy:

```shell
planton apply -f custom-role.yaml
```

This creates the role `projects/my-gcp-project-123/roles/logBucketWriter`, ready to be granted with a GcpProjectIamMember.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `roleId` | `string` | Role identifier, unique within the project. Forms the full name `projects/<project>/roles/<roleId>`. | 3-64 chars; letters, digits, underscores, periods; NO hyphens. Immutable. |
| `title` | `string` | Human-readable title shown in the GCP console and IAM policy pickers. | Max 100 chars. Mutable. |
| `permissions` | `list(string)` | The IAM permissions the role grants, e.g. `storage.objects.get`. | At least one required. Mutable. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider default project | GCP project that owns this role. Can reference a GcpProject resource. Immutable. |
| `description` | `string` | `""` | What this role is for and who should hold it. Max 256 chars. Mutable. |
| `stage` | `string` | `GA` | Launch stage label: `ALPHA`, `BETA`, `GA`, `DEPRECATED`, `DISABLED`, `EAP`. `DISABLED` keeps the role defined while rejecting all of its grants. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE` soft-deletes the role on destroy (recoverable 7 days); `PREVENT` fails the destroy; `ABANDON` leaves the role active with every binding working. |

## Why Custom Roles

Predefined roles bundle far more permissions than most workloads need — `roles/storage.admin` on a service account that only writes log objects is an audit finding waiting to happen. A custom role is the least-privilege building block: define the exact permission set once, then grant it wherever needed. Editing the role's permissions updates every existing grant immediately — that is the point of defining the bundle in one place.

Discover valid permission strings with:

```shell
gcloud iam list-testable-permissions //cloudresourcemanager.googleapis.com/projects/my-gcp-project-123
```

## Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | Fully-qualified role name (`projects/<project>/roles/<roleId>`) — the grantable handle; feed it directly into a GcpProjectIamMember's `role` field |
| `role_id` | `string` | The bare role ID within the project |
| `deleted` | `bool` | Whether the role is currently soft-deleted |

## Deployment Methods

Planton supports two deployment methods:

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Immutability**: `roleId` and `projectId` are ForceNew — changing either destroys and recreates the role, breaking every grant that references the old role name. `title`, `description`, `stage`, and `permissions` update in place.
- **Soft delete**: GCP retains deleted custom roles for up to 14 days, during which the `roleId` stays reserved and grants of the role are rejected. Re-creating a role with a soft-deleted ID undeletes and updates it in place — the modules handle this natively, so a destroy/recreate cycle within the window converges rather than failing.
- **Unsupported permissions**: some permissions cannot be used in custom roles (marked `NOT_SUPPORTED` in `list-testable-permissions` output); the API rejects them at deploy time.
- **Org-scoped roles**: this kind models project-scoped roles. Organization-scoped custom roles are a separate concern with a different parent resource.

## Related Kinds

- [GcpProjectIamMember](/docs/catalog/gcp/gcpprojectiammember) — grants this role to an identity (its `role` field references this role's `name` output)
- [GcpServiceAccount](/docs/catalog/gcp/gcpserviceaccount) — the identity most commonly granted custom roles
- [GcpProject](/docs/catalog/gcp/gcpproject) — provides the GCP project that owns the role

## Additional Resources

- [Creating and Managing Custom Roles](https://cloud.google.com/iam/docs/creating-custom-roles)
- [Understanding IAM Custom Roles](https://cloud.google.com/iam/docs/understanding-custom-roles)
- [Permissions Reference](https://cloud.google.com/iam/docs/permissions-reference)

## Support

For issues, questions, or contributions, please refer to the Planton documentation or open an issue in the repository.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
