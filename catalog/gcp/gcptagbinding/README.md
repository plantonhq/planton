# GCP Tag Binding

Attaches one Google Cloud Resource Manager tag value to one resource — the act of tagging. The value (`GcpTagValue`, e.g. `environment/prod`) already exists; the binding says "this project, this folder, this VM carries it", and from that moment every organization policy conditioned on the tag, every IAM condition, and every firewall policy rule targeting it applies to the resource — and, for projects and folders, to everything beneath them. A binding with no parent tags the project you are deploying into.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Tag binding** -- the `tags_tag_binding` between the value and the resource's full resource name; for a regional or zonal resource (set `location`), the `tags_location_tag_binding` served from that location instead

## Before You Deploy

### A Binding Is Replaced, Never Edited — Read This First

- **Every field is immutable.** Any change destroys the old binding and creates the new one. Deletion is immediate, with no soft-delete window.
- **One value per key per resource.** Binding `environment/staging` to a project that already carries `environment/prod` is rejected by Google, not swapped; destroy the old binding first.
- **Google requires a project's NUMBER.** A `GcpProject` reference resolves to it and a numeric literal is used as is; a project ID, or an empty parent (the provider's default project), is looked up once at apply time.

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials that can use the tag value and tag the target resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Resource

- **IAM**: the deploying identity needs `roles/resourcemanager.tagUser` on the tag value AND on the resource being tagged.
- **Regional and zonal resources need `location`** (a Compute instance's zone, a Cloud SQL instance's region); organizations, folders, and projects are global.

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpTagBinding
metadata:
  name: project-environment-prod
spec:
  tagValue:
    value: tagValues/281476102962987
```

```shell
planton apply -f tag-binding.yaml
```

With no `parent`, the value is bound to the project the credentials are configured for.

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `tagValue` | `StringValueOrRef` | The value: a `GcpTagValue` reference, or the `tagValues/{id}` or `{org}/{key}/{value}` literal. Immutable. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `parent` | `message` | provider default project | At most one of `projectId` (a `GcpProject` reference, or a literal ID or number), `folderId` (a `GcpFolder` reference), `organizationId`, `resourceName` (any other taggable resource's full resource name). Immutable. |
| `location` | `string` | — | Region or zone of a regional or zonal `resourceName`. Immutable. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE` (immediate), `PREVENT` (destroy fails), or `ABANDON` (unmanaged, stays tagged). |

### Validation Rules

- **`tagValue`** literal matches `^tagValues/[0-9]+$` or the three-segment namespaced form.
- **At most one parent arm**; `organizationId` is numeric; `resourceName` begins with `//{service}.googleapis.com/`.
- **`location`** only with `resourceName`, and shaped like a region (`us-central1`) or zone (`us-central1-a`).
- **`deletionPolicy`**: `DELETE`, `PREVENT`, or `ABANDON`.

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | `tagBindings/{encoded parent}/{tagValues/id}` |
| `parent` | `string` | The full resource name the tag is bound to, as sent |
| `tag_value` | `string` | `tagValues/{id}` |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Two provider resources, one kind** — `location` empty uses the global binding; set, the location-scoped one. Both engines make the same choice.
- **The project-number lookup is guarded** — it runs only when the number is not already known (a literal ID, or no parent), so a plan on a reference or a numeric literal performs no live read on either engine.
- **Tags inherit down the hierarchy** — a binding on a folder applies to every project and folder beneath it; a binding on a project to every resource in it.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Components

- [GcpTagValue](/docs/catalog/gcp/gcptagvalue) — the value being bound
- [GcpTagKey](/docs/catalog/gcp/gcptagkey) — the key the value belongs to
- [GcpFolder](/docs/catalog/gcp/gcpfolder) — a folder as the tagged resource
- [GcpProject](/docs/catalog/gcp/gcpproject) — a project as the tagged resource
- [GcpOrgPolicy](/docs/catalog/gcp/gcporgpolicy) — rules that fire on the bound tag

## Additional Resources

- [Attaching tags to resources](https://cloud.google.com/resource-manager/docs/tags/tags-creating-and-managing#attaching)
- [Full resource names](https://cloud.google.com/asset-inventory/docs/resource-name-format)
- [Tag Bindings API reference](https://cloud.google.com/resource-manager/reference/rest/v3/tagBindings)

## Support

For issues, questions, or contributions, please refer to the Planton documentation or open an issue in the repository.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
