# GCP Folder

Creates a Google Cloud Resource Manager folder — a node in the resource hierarchy that groups projects and other folders so IAM policy, organization policies, and billing views apply to the whole group at once. A landing zone is folders inside folders (`environments/production`, `teams/payments`), each carrying the guardrails and grants every project beneath it inherits. Declare the parent as a reference to another `GcpFolder` and the chart builds the tree in dependency order; place a project inside with `GcpProject`'s `folderId`.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Folder** -- the `folder` under the organization or inside another folder, with its display name, its destroy guard (`deletionProtection`, on by default), and any create-time tags

## Before You Deploy

### A Folder Is a Hierarchy Node — Read This First

- **The destroy guard is on by default.** With `deletionProtection` unset or true, a destroy fails until the field is set to false and applied first. Deleting a folder is a hierarchy-wide act; the guard keeps a just-emptied folder from disappearing on one accidental destroy. `deletionPolicy: ABANDON` bypasses the guard; `PREVENT` overrides both.
- **A non-empty folder cannot be deleted.** Google refuses while any project or folder is inside it. A chart that declares the children by reference destroys them first.
- **Changing the parent moves the folder; changing the tags recreates it.** A move carries everything inside along and re-scopes inherited IAM and policies at once. Create-time `tags` are immutable — for tags on an existing folder use `GcpTagBinding`.

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials that can create folders in the target organization.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Organization

- **A Google Cloud Organization** -- folders exist only inside an organization.
- **IAM**: the deploying identity needs `roles/resourcemanager.folderAdmin` (or `folderCreator` plus `folderEditor`) on the parent organization or folder.

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpFolder
metadata:
  name: production
spec:
  parent:
    organizationId: "123456789012"
  deletionPolicy: PREVENT
```

```shell
planton apply -f folder.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `parent` | `message` | Exactly one of `organizationId` (numeric, a top-level folder) or `folderId` (a `GcpFolder` reference or numeric literal, a nested folder). |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `displayName` | `string` | `metadata.name` | 3-30 characters of letters, digits, spaces, hyphens, underscores; starts and ends alphanumeric; unique among siblings. Mutable. |
| `deletionProtection` | `bool` | `true` | Client-side destroy guard. Always sent explicitly by both engines. |
| `tags` | `map<string,string>` | `{}` | Create-time tags, `tagKeys/{id}` -> `tagValues/{id}`. Immutable — changing them recreates the folder. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE` (soft-delete, 30-day recovery), `PREVENT` (destroy fails), or `ABANDON` (unmanaged, stays live). |

### Validation Rules

- **Exactly one parent arm**: `organizationId` or `folderId`, never both or neither.
- **`organizationId`** is numeric, without the `organizations/` prefix.
- **`displayName`**: `^[A-Za-z0-9][A-Za-z0-9 _-]{1,28}[A-Za-z0-9]$` when set.
- **`tags`**: keys match `^tagKeys/[0-9]+$`, values match `^tagValues/[0-9]+$`.
- **`deletionPolicy`**: `DELETE`, `PREVENT`, or `ABANDON`.

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `folder_id` | `string` | The folder's numeric ID — what every child references |
| `name` | `string` | `folders/{folder_id}` |
| `lifecycle_state` | `string` | `ACTIVE`, or `DELETE_REQUESTED` during the soft-delete window |
| `create_time` | `string` | RFC 3339 creation timestamp |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Display names are reserved for 30 days after deletion** — a deleted folder is soft-deleted and recoverable, and a fresh folder with the same name under the same parent fails until the window closes.
- **Ten levels deep at most** — Google's hierarchy limit; a landing zone rarely needs more than three.
- **IAM on the folder is a separate concern** — grants belong on their own additive resources, as `GcpProjectIamMember` is for projects; organization policies scoped to the folder are `GcpOrgPolicy` resources referencing `folder_id`.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Components

- [GcpProject](/docs/catalog/gcp/gcpproject) — projects placed inside the folder by reference
- [GcpOrgPolicy](/docs/catalog/gcp/gcporgpolicy) — guardrails scoped to the folder
- [GcpTagBinding](/docs/catalog/gcp/gcptagbinding) — tags on the folder after creation

## Additional Resources

- [Creating and managing folders](https://cloud.google.com/resource-manager/docs/creating-managing-folders)
- [Resource hierarchy overview](https://cloud.google.com/resource-manager/docs/cloud-platform-resource-hierarchy)
- [Folders API reference](https://cloud.google.com/resource-manager/reference/rest/v3/folders)

## Support

For issues, questions, or contributions, please refer to the Planton documentation or open an issue in the repository.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
