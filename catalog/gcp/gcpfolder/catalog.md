# GCP Folder

Creates a Google Cloud Resource Manager folder: a node in the resource hierarchy that groups projects and other folders so IAM policy, organization policies, and billing views apply to the whole group at once. A landing zone is folders inside folders -- `environments/production`, `teams/payments` -- each carrying the guardrails and grants every project beneath it inherits. Declare the parent as a reference to another `GcpFolder` and the chart builds the tree in order; place a project inside with `GcpProject`'s `folderId`.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Folder** -- the `folder` under the organization or inside another folder, with its display name, its destroy guard (`deletionProtection`, on by default), and any create-time tags

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials that can create folders in the target organization. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.

### GCP Organization

- **A Google Cloud Organization** -- folders exist only inside an organization; accounts without one cannot create folders.
- **IAM**: the deploying identity needs `roles/resourcemanager.folderAdmin` (or `folderCreator` plus `folderEditor`) on the parent organization or folder. Folder-level roles are granted at the organization or on the parent folder, never on a project.
- **Display names are unique among siblings** -- two folders under the same parent cannot share a display name, and a deleted folder's name stays reserved for 30 days.

## Deploy

### Console

Open the deployment store, find **GCP Folder**, and click **Deploy**. The creation wizard walks you through the parent (organization or folder), the display name, and the destroy guard. Start from the **Environment Folder** preset in the [Presets](#presets) tab for the common shape.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpFolder
metadata:
  name: production
  org: acme-corp
  env: prod
spec:
  parent:
    organizationId: "123456789012"
  displayName: production
  deletionProtection: true
  deletionPolicy: PREVENT
```

```shell
planton apply -f folder.yaml
```

This creates a top-level folder named `production` directly under the organization, protected against accidental destroy twice over (the guard and the `PREVENT` policy). A Stack Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, a nested folder references its parent folder via ValueFromRef, and projects, policies, and tag bindings reference the folder's `folder_id`:

```yaml
spec:
  parent:
    folderId:
      valueFrom:
        kind: GcpFolder
        name: environments
        fieldPath: status.outputs.folder_id
  displayName: production
```

The InfraPipeline deploys the parent folder first, then this one, then every `GcpProject`, `GcpOrgPolicy`, or `GcpTagBinding` that points at `status.outputs.folder_id`.

## Key Configuration

These are the most important decisions when configuring a folder. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Parent** -- exactly one of `organizationId` (a top-level folder) or `folderId` (nested; a reference to another `GcpFolder`). Changing the parent MOVES the folder in place with everything inside it -- nothing is recreated, but every IAM grant and organization policy inherited from the old parent stops applying and the new parent's start at once.

**Destroy guard** -- `deletionProtection` defaults to true: a destroy fails until it is set to false and applied first. Deleting a folder is a hierarchy-wide act; the guard keeps a just-emptied folder from disappearing on one accidental destroy. `ABANDON` bypasses the guard; `PREVENT` overrides both.

**Create-time tags** -- `tags` (`tagKeys/{id}` -> `tagValues/{id}`) bind at creation and RECREATE the folder when changed, which Google refuses for a non-empty folder. Use them only when an organization policy conditioned on the tag must govern the folder from its first second; otherwise bind tags after creation with `GcpTagBinding`.

**Deletion policy** -- `DELETE` (default) soft-deletes the folder for 30 days (Google refuses while it still holds projects or folders); `PREVENT` fails the destroy -- the guard for the folders a landing zone is built on; `ABANDON` leaves the folder and everything in it unmanaged.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpFolder** (optional) | `parent.folderId` | `status.outputs.folder_id` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `folder_id` | The folder's numeric ID | A nested `GcpFolder`'s `parent.folderId`, a `GcpProject`'s `folderId`, a `GcpOrgPolicy`'s `scope.folderId`, a `GcpTagBinding`'s `parent.folderId` |
| `name` | `folders/{folder_id}` | Addressing the folder in IAM and Google's APIs |
| `lifecycle_state` | `ACTIVE`, or `DELETE_REQUESTED` during the soft-delete window | Health checks |
| `create_time` | RFC 3339 creation timestamp | Auditing |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Environment folder** -- a top-level folder per environment under the organization, `PREVENT` on destroy. Start from the **Environment Folder** preset.

**Nested team folder** -- a folder inside another folder by reference, building the `environments/production/payments` tree. Start from the **Nested Team Folder** preset.

**Tagged at create** -- a folder that carries a tag from its first second so a tag-conditioned organization policy governs it immediately. Start from the **Tagged At Create** preset.

## Works With

- [**GCP Project**](/cloud-catalog/gcp-project) -- projects placed inside the folder by reference (`folderId`)
- [**GCP Organization Policy**](/cloud-catalog/gcp-org-policy) -- guardrails scoped to the folder and inherited by everything beneath it
- [**GCP Tag Binding**](/cloud-catalog/gcp-tag-binding) -- tags on the folder after creation
