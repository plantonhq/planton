# GCP Tag Value

Creates a Google Cloud Resource Manager tag value: the VALUE half of a tag -- `prod` under the `environment` key, `pci` under `data-classification`. A value belongs to exactly one key (`GcpTagKey`) and is what gets bound to resources (`GcpTagBinding`) and what organization policies, IAM conditions, and firewall rules test for. Declare one value per allowed setting of a key; a key's vocabulary is the set of values under it.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Tag value** -- the `tags_tag_value` under its key, with its short name and description

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials that can administer tags at the key's owner. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.

### GCP Tag Key

- **The key must exist** -- declare it with `GcpTagKey` and reference its `name` output, or pass the `tagKeys/{id}` literal.
- **IAM**: the deploying identity needs `roles/resourcemanager.tagAdmin` on the key's owner (organization or project).
- **Short names are unique per key** and reserved for 30 days after deletion; when the key has an allowed-values regex, the short name must match it.

## Deploy

### Console

Open the deployment store, find **GCP Tag Value**, and click **Deploy**. The creation wizard walks you through the key and the short name. Start from the **Environment Prod** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpTagValue
metadata:
  name: prod
  org: acme-corp
  env: prod
spec:
  tagKey:
    value: tagKeys/281475647562788
  shortName: prod
  description: Customer-facing workloads
```

```shell
planton apply -f tag-value.yaml
```

This declares `environment/prod` under the existing `environment` key. A Stack Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, the value references its `GcpTagKey` via ValueFromRef, and every `GcpTagBinding` references the value's `name`:

```yaml
spec:
  tagKey:
    valueFrom:
      kind: GcpTagKey
      name: environment
      fieldPath: status.outputs.name
  shortName: prod
```

The InfraPipeline deploys the key first, then the value, then every binding that points at `status.outputs.name`.

## Key Configuration

These are the most important decisions when configuring a tag value. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Key** -- the `GcpTagKey` the value belongs to (its `name` output, `tagKeys/{id}`). Immutable.

**Short name** -- the value as written in conditions (`resource.matchTag('{org}/environment', 'prod')`); defaults to the manifest name. Any UTF-8 except `/`, `\`, `'`, `"`. Immutable: rename by creating a new value.

**Deletion policy** -- `DELETE` (default) fails while any binding still uses the value -- destroy the bindings first (a chart's dependency order does this when they reference the value); `PREVENT` fails the destroy; `ABANDON` leaves the value and its bindings live.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpTagKey** | `tagKey` | `status.outputs.name` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `name` | `tagValues/{id}` | A `GcpTagBinding`'s `tagValue`; a `GcpFolder`'s or `GcpProject`'s create-time `tags` value |
| `namespaced_name` | `{parent}/{keyShortName}/{shortName}` | Human-readable tooling |
| `tag_value_id` | The bare numeric ID | Tooling |
| `create_time` | RFC 3339 creation timestamp | Auditing |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Environment prod** -- the `prod` value of the `environment` key. Start from the **Environment Prod** preset.

**Environment nonprod** -- the sibling value, showing one key with several values. Start from the **Environment Nonprod** preset.

## Works With

- [**GCP Tag Key**](/cloud-catalog/gcp-tag-key) -- the key this value belongs to
- [**GCP Tag Binding**](/cloud-catalog/gcp-tag-binding) -- attaches this value to a resource
- [**GCP Organization Policy**](/cloud-catalog/gcp-org-policy) -- rules conditioned on this value
