# GCP SCC Mute Config

A Security Command Center mute rule for a project, a folder, or the whole organization. Findings that match its filter -- accepted risks, known-noisy detectors, sandbox projects -- are muted, so triage, notifications that skip muted findings, and exports see only what needs action. Muted findings are not deleted; they stay queryable.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **API enablement** -- `securitycenter.googleapis.com` on a project rule's project (never disabled on destroy)
- **Mute config** -- one `scc_v2_{project,folder,organization}_mute_config`, chosen by the scope

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with Security Command Center mute-config admin permissions (`roles/securitycenter.muteConfigsEditor`) at the scope.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Security Command Center

Security Command Center must be activated on the scope.

### Optional Dependencies

- **`GcpProject`** / **`GcpFolder`** -- the scope, by reference.

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpSccMuteConfig
metadata:
  name: sandbox-public-buckets
spec:
  muteConfigId: sandbox-public-buckets
  filter: category = "PUBLIC_BUCKET_ACL" AND resource.project_display_name = "sandbox"
  type: DYNAMIC
```

```shell
planton apply -f scc-mute-config.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `muteConfigId` | `string` | The rule's ID: lowercase letters, digits, hyphens; starts with a letter; at most 63 characters. Immutable. |
| `filter` | `string` | Which findings are muted, e.g. `category = "PUBLIC_BUCKET_ACL"`. |
| `type` | `string` | `DYNAMIC` (existing and future findings, reversible) or `STATIC` (future findings, permanent). |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `scope` | `object` | provider project | At most one of `projectId` (`GcpProject` ref), `folderId` (`GcpFolder` ref), `organizationId` (numeric). |
| `description` | `string` | none | Why the rule exists. |
| `location` | `string` | `global` | Where the configuration is stored; a residency location (`eu`, `us`) only if data residency was set up at activation. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE` removes it, `PREVENT` fails destroy, `ABANDON` keeps it in Google. |

### Validation Rules

- `scope` names at most one of project, folder, organization; `organizationId` is numeric without the `organizations/` prefix.
- `location` is `global` or a residency location; `deletionPolicy` takes only Google's values.
- `muteConfigId` follows Google's ID rule; `type` is `DYNAMIC` or `STATIC`.

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | `{parent}/locations/{location}/muteConfigs/{muteConfigId}` |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Prefer DYNAMIC.** A dynamic rule mutes existing and future matches and lifts its mute when the rule changes or is deleted; a static rule permanently mutes future matches only.
- **Write the filter for the scope.** A filter naming project X on a rule scoped to project Y matches nothing.
- **Muting is not fixing.** Record the accepted risk in `description`.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Kinds

- **GcpSccNotificationConfig** -- streams findings, optionally skipping muted ones
- **GcpSccBigQueryExport** -- exports findings, optionally skipping muted ones

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
