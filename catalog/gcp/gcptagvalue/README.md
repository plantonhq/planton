# GCP Tag Value

Creates a Google Cloud Resource Manager tag value — the VALUE half of a tag: `prod` under the `environment` key, `pci` under `data-classification`. A value belongs to exactly one key (`GcpTagKey`) and is what gets bound to resources (`GcpTagBinding`) and what organization policies, IAM conditions, and firewall rules test for. Declare one value per allowed setting of a key.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Tag value** -- the `tags_tag_value` under its key, with its short name and description

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials that can administer tags at the key's owner.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Tag Key

- **The key must exist** — declare it with `GcpTagKey` and reference its `name` output, or pass the `tagKeys/{id}` literal.
- **IAM**: the deploying identity needs `roles/resourcemanager.tagAdmin` on the key's owner.
- **Short names are unique per key** and reserved for 30 days after deletion; with an allowed-values regex on the key, the short name must match it.

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpTagValue
metadata:
  name: prod
spec:
  tagKey:
    value: tagKeys/281475647562788
  description: Customer-facing workloads
```

```shell
planton apply -f tag-value.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `tagKey` | `StringValueOrRef` | The key, as a `GcpTagKey` reference or the `tagKeys/{id}` literal. Immutable. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `shortName` | `string` | `metadata.name` | 1-256 characters of any UTF-8 except `/`, `\`, `'`, `"`. Immutable. |
| `description` | `string` | — | At most 256 characters. Mutable. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE` (fails while bindings exist), `PREVENT` (destroy fails), or `ABANDON` (unmanaged, stays live). |

### Validation Rules

- **`tagKey`** literal matches `^tagKeys/[0-9]+$`.
- **`shortName`** contains none of `/`, `\`, `'`, `"`; at most 256 characters.
- **`deletionPolicy`**: `DELETE`, `PREVENT`, or `ABANDON`.

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | `tagValues/{id}` — what a `GcpTagBinding`'s `tagValue` references |
| `namespaced_name` | `string` | `{parent}/{keyShortName}/{shortName}` |
| `tag_value_id` | `string` | The bare numeric ID |
| `create_time` | `string` | RFC 3339 creation timestamp |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **The key and the short name are immutable** — rename by creating a new value and re-binding.
- **A value with bindings cannot be deleted** — destroy the `GcpTagBinding`s first; a chart that references the value from its bindings does this in order.
- **Create-time tags on folders and projects** take this value's `name` (`tagValues/{id}`) as the map value.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Components

- [GcpTagKey](/docs/catalog/gcp/gcptagkey) — the key this value belongs to
- [GcpTagBinding](/docs/catalog/gcp/gcptagbinding) — attaches this value to a resource
- [GcpOrgPolicy](/docs/catalog/gcp/gcporgpolicy) — rules conditioned on this value

## Additional Resources

- [Tags overview](https://cloud.google.com/resource-manager/docs/tags/tags-overview)
- [Creating and managing tags](https://cloud.google.com/resource-manager/docs/tags/tags-creating-and-managing)

## Support

For issues, questions, or contributions, please refer to the Planton documentation or open an issue in the repository.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
