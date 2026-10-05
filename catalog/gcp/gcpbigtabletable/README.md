# GCP Bigtable Table

Deploys a Cloud Bigtable table (`google_bigtable_table`) with its column families and per-family garbage-collection policies (`google_bigtable_gc_policy`) — the schema-bearing, infrastructure-owned unit inside a Bigtable instance: families with retention, pre-splits for load distribution, change streams, automated backups, and deletion protection.

## What Gets Created

One table plus one GC-policy object per column family that declares one (the API's own one-per-family granularity — folded into this kind because a GC policy has no independent life apart from its family). Data (rows and cells) is application territory; this resource owns the structure applications write into.

## Prerequisites

- **GCP credentials** configured via environment variables or Planton provider config
- **A Bigtable instance** — referenced via `instance` (a `GcpBigtableInstance` resource or a literal instance name)
- **IAM permissions** — see [`iac/permissions.yaml`](iac/permissions.yaml) for the least-privilege permission set the deploying principal needs

## Quick Start

Create a file `table.yaml`:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpBigtableTable
metadata:
  name: sensor-readings
spec:
  instance:
    valueFrom:
      kind: GcpBigtableInstance
      name: prod-bigtable
      fieldPath: status.outputs.instance_name
  columnFamilies:
    - family: measurements
      gcPolicy:
        maxAge: 2160h # 90 days
```

Deploy:

```shell
planton apply -f table.yaml
```

## Configuration Reference

### Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider default project | Project owning the instance. |
| `instance` | `StringValueOrRef` | — (required) | Parent instance's short name. Immutable. |
| `tableName` | `string` | `metadata.name` | Table name (1-50 chars). Immutable. |
| `columnFamilies[].family` | `string` | — | Family name. Mutable list — append as the app grows. |
| `columnFamilies[].type` | `string` | — | Aggregate shortcuts (`intsum`/`intmin`/`intmax`/`inthll`) or a raw JSON Type. |
| `columnFamilies[].gcPolicy` | message | — | `maxAge`, `maxVersions`, `mode` (UNION/INTERSECTION when combining), or raw `gcRules` JSON; `ignoreWarnings` for replicated-instance expansions. |
| `splitKeys` | `string[]` | — | Row keys to pre-split at. Immutable — changing REPLACES the table. |
| `changeStreamRetention` | `string` | off | CDC feed retention, 24h-168h; `0` disables. |
| `automatedBackupPolicy` | message | off | `retentionPeriod` + `frequency` (duration strings). |
| `deletionProtection` | `string` | `PROTECTED` | API-side guard — deletion by ANY client fails until set `UNPROTECTED`. |
| `rowKeySchema` | `string` | — | Structured row-key schema as Type JSON. In-place update unsupported: clear, apply, then set. |

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| `table_id` | `string` | Fully qualified table resource path |
| `table_name` | `string` | Short table name (what clients open) |
| `instance_name` | `string` | The parent instance |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md).

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md).

## Important Notes

- **GC is opt-in and load-bearing**: Bigtable never deletes old cell versions without a GC policy — an unbounded family accumulates every write forever, which is the most common source of surprise Bigtable bills. Give every family a policy.
- **`deletionProtection` defaults PROTECTED**: the API-side guard blocks deletion by any client (not just IaC) until explicitly set `UNPROTECTED` — the safety default for a data-bearing resource, sent explicitly by both engines.
- **`splitKeys` is a creation-time decision**: changing it replaces the table and its data. Set it right at creation or manage splits operationally.
- **GC policy changes on replicated instances**: expanding what is eligible for collection needs `ignoreWarnings: true` — Bigtable otherwise rejects it as a data-loss safety measure.
- **No labels surface**: the Bigtable table resource has no labels (instance-level labels only) — both engines skip labels identically.
- **`deletionPolicy` covers both objects**: one field drives the destroy behavior of the table AND its per-family GC policies — `DELETE` (default), `PREVENT` (destroy fails), or `ABANDON` (drop from management, keep the table; also the escape hatch when a GC-policy delete is rejected on a replicated instance).
- **`automatedBackupPolicy.locations` needs ENTERPRISE_PLUS**: restricting backup placement to specific zones (`projects/{project}/locations/{zone}`) is only accepted for tables on ENTERPRISE_PLUS instances; empty means all zones of the instance.

### Deliberately not modeled (recorded reasons)

- **App profiles, authorized/logical/materialized views, schema bundles** — separate provider resources with real but second-order demand; Tier-2 candidates on concrete pull.
- **Table IAM trio** — resource-scoped IAM stays unmodeled catalog-wide (additive project grants compose instead).

## Related Kinds

- [GcpBigtableInstance](/docs/catalog/gcp/gcpbigtableinstance) — the parent instance
- [GcpProject](/docs/catalog/gcp/gcpproject) — provides the GCP project

## Additional Resources

- [Bigtable garbage collection](https://cloud.google.com/bigtable/docs/garbage-collection)
- [Bigtable schema design](https://cloud.google.com/bigtable/docs/schema-design)

## Support

For issues, questions, or contributions, please refer to the Planton documentation or open an issue in the repository.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
