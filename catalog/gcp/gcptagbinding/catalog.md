# GCP Tag Binding

Attaches one Google Cloud Resource Manager tag value to one resource -- the act of tagging. The value (`GcpTagValue`, e.g. `environment/prod`) already exists; the binding says "this project, this folder, this VM carries it", and from that moment every organization policy conditioned on the tag, every IAM condition, and every firewall policy rule targeting it applies to the resource -- and, for projects and folders, to everything beneath them. A binding with no parent tags the project you are deploying into.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Tag binding** -- the `tags_tag_binding` between the value and the resource's full resource name; for a regional or zonal resource (set `location`), the `tags_location_tag_binding` served from that location instead

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials that can use the tag value and tag the target resource. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.

### GCP Resource

- **IAM**: the deploying identity needs `roles/resourcemanager.tagUser` on the tag value AND on the resource being tagged (for a project binding, on the project).
- **One value per key per resource** -- Google rejects a second value of the same key on a resource; destroy the old binding first.
- **Regional and zonal resources need `location`** -- a Compute instance, a Cloud SQL instance, a GKE cluster; organizations, folders, and projects are global.

## Deploy

### Console

Open the deployment store, find **GCP Tag Binding**, and click **Deploy**. The creation wizard walks you through the tag value and the resource. Start from the **Bind Project** preset in the [Presets](#presets) tab for the common shape.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpTagBinding
metadata:
  name: project-environment-prod
  org: acme-corp
  env: prod
spec:
  tagValue:
    value: tagValues/281476102962987
```

```shell
planton apply -f tag-binding.yaml
```

This binds `environment/prod` to the project the credentials are configured for -- no parent needed. A Stack Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, the binding references its `GcpTagValue` and the resource it tags (a `GcpProject` or `GcpFolder`) via ValueFromRef:

```yaml
spec:
  tagValue:
    valueFrom:
      kind: GcpTagValue
      name: environment-prod
      fieldPath: status.outputs.name
  parent:
    folderId:
      valueFrom:
        kind: GcpFolder
        name: production
        fieldPath: status.outputs.folder_id
```

The InfraPipeline deploys the key, the value, and the folder first, then the binding -- and destroys them in the reverse order, which is the order Google requires.

## Key Configuration

These are the most important decisions when configuring a tag binding. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Tag value** -- a reference to a `GcpTagValue` (its `name` output), or the `tagValues/{id}` or namespaced (`{org}/{key}/{value}`) literal. Immutable.

**Parent** -- at most one of `projectId` (a `GcpProject` reference, or a literal ID or number), `folderId`, `organizationId`, or `resourceName` (any other taggable resource by full resource name, `//compute.googleapis.com/projects/p/zones/z/instances/i`). Empty means the provider's default project. Google requires a project's NUMBER; a reference resolves to it, and a literal ID is looked up once at apply time. Immutable.

**Location** -- the region or zone of a regional or zonal `resourceName`; must stay empty for organizations, folders, and projects. Immutable.

**Deletion policy** -- `DELETE` (default) removes the binding at once (no recovery window); `PREVENT` fails the destroy -- the guard for a tag a policy depends on; `ABANDON` leaves the resource tagged.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpTagValue** | `tagValue` | `status.outputs.name` |
| **GcpProject** (optional) | `parent.projectId` | `status.outputs.project_number` |
| **GcpFolder** (optional) | `parent.folderId` | `status.outputs.folder_id` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `name` | `tagBindings/{encoded parent}/{tagValues/id}` | Auditing; addressing the binding in tooling |
| `parent` | The full resource name the tag is bound to, as sent | Confirming which project number a literal ID resolved to |
| `tag_value` | `tagValues/{id}` | Cross-checking the bound value |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Bind project** -- tag the project by reference to a `GcpProject`. Start from the **Bind Project** preset.

**Bind folder** -- tag a folder so everything beneath inherits the tag. Start from the **Bind Folder** preset.

**Bind regional resource** -- tag a Cloud SQL instance by full resource name with its region. Start from the **Bind Regional Resource** preset.

## Works With

- [**GCP Tag Value**](/cloud-catalog/gcp-tag-value) -- the value being bound
- [**GCP Tag Key**](/cloud-catalog/gcp-tag-key) -- the key the value belongs to
- [**GCP Folder**](/cloud-catalog/gcp-folder) -- a folder as the tagged resource
- [**GCP Project**](/cloud-catalog/gcp-project) -- a project as the tagged resource
- [**GCP Organization Policy**](/cloud-catalog/gcp-org-policy) -- rules that fire on the bound tag
