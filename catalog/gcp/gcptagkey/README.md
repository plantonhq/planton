# GCP Tag Key

Creates a Google Cloud Resource Manager tag key — the NAME half of a tag such as `environment` or `data-classification`. A key owns a set of values (`GcpTagValue`: `prod`, `staging`), and a value is bound to a resource (`GcpTagBinding`) so organization policies, IAM conditions, and firewall policies can key on it. Tags are governed metadata (created centrally, permissioned, immutable in name, evaluated by the platform's policy engines) where labels are free-form and enforce nothing. A key is owned by the organization or by one project; never by a folder.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Tag key** -- the `tags_tag_key` under the organization or project, with its short name, description, optional purpose, and optional allowed-values regex

## Before You Deploy

### Almost Everything About a Key Is Immutable — Read This First

- **Owner, short name, purpose, and purpose data recreate the key when changed** — and Google refuses to delete a key that still has values. Only the description and the allowed-values regex update in place. Rename by creating a new key.
- **A folder cannot own a key.** Choose the organization (values bindable everywhere) or one project (values bindable only inside it).
- **Short names are reserved for 30 days after deletion** under the same owner.

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials that can administer tags at the owner.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Owner

- **IAM**: the deploying identity needs `roles/resourcemanager.tagAdmin` on the owner — on the project for a project-owned key (no organization-level grant needed), on the organization otherwise.

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpTagKey
metadata:
  name: environment
spec:
  parent:
    organizationId: "123456789012"
  description: The environment a resource belongs to
```

```shell
planton apply -f tag-key.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `parent` | `message` | Exactly one of `organizationId` (numeric) or `projectId` (a `GcpProject` reference or literal). Immutable. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `shortName` | `string` | `metadata.name` | 1-256 characters of any UTF-8 except `/`, `\`, `'`, `"`. Immutable. |
| `description` | `string` | — | At most 256 characters. Mutable. |
| `purpose` | `string` | — | `GCE_FIREWALL` (secure tags for network firewall policies; needs `purposeData.network`) or `DATA_GOVERNANCE`. Immutable. |
| `purposeData` | `map<string,string>` | `{}` | Per-purpose data; for `GCE_FIREWALL`, `network: {project}/{vpc}`. Immutable; only with a purpose. |
| `allowedValuesRegex` | `string` | — | RE2 every value's short name must match; also makes the key dynamic. Mutable. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE` (fails while values exist), `PREVENT` (destroy fails), or `ABANDON` (unmanaged, stays live). |

### Validation Rules

- **Exactly one owner arm**: `organizationId` or `projectId`.
- **`shortName`** contains none of `/`, `\`, `'`, `"`; at most 256 characters.
- **`purpose`**: empty, `GCE_FIREWALL`, or `DATA_GOVERNANCE`; **`purposeData`** only with a purpose.
- **`deletionPolicy`**: `DELETE`, `PREVENT`, or `ABANDON`.

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | `tagKeys/{id}` — what a `GcpTagValue`'s `tagKey` references |
| `namespaced_name` | `string` | `{org_or_project}/{shortName}` — the form `resource.matchTag` tests |
| `tag_key_id` | `string` | The bare numeric ID |
| `create_time` | `string` | RFC 3339 creation timestamp |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Tags versus labels** — a tag is governed and policy-evaluated; a label is free-form metadata. Use labels for cost allocation and search, tags for anything a policy or a condition must trust.
- **A dynamic key** (one with `allowedValuesRegex`) accepts bindings whose value was never declared, as long as it matches — the pattern for high-cardinality tags such as ticket numbers.
- **Delete order** — bindings, then values, then the key. A chart that declares all three by reference does this automatically.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Components

- [GcpTagValue](/docs/catalog/gcp/gcptagvalue) — the values under this key
- [GcpTagBinding](/docs/catalog/gcp/gcptagbinding) — attaches a value to a resource
- [GcpOrgPolicy](/docs/catalog/gcp/gcporgpolicy) — rules conditioned on the tag
- [GcpProject](/docs/catalog/gcp/gcpproject) — the owner of a project-scoped key

## Additional Resources

- [Tags overview](https://cloud.google.com/resource-manager/docs/tags/tags-overview)
- [Creating and managing tags](https://cloud.google.com/resource-manager/docs/tags/tags-creating-and-managing)
- [Tags for firewalls](https://cloud.google.com/firewall/docs/tags-firewalls-overview)

## Support

For issues, questions, or contributions, please refer to the Planton documentation or open an issue in the repository.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
