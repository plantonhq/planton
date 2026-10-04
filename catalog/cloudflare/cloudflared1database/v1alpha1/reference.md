# CloudflareD1Database

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `cloudflare.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

CloudflareD1DatabaseSpec provisions a Cloudflare D1 database: a serverless
SQLite database that a Worker queries via a `d1` binding. Placement is fixed at
creation by an optional region hint or a data-residency jurisdiction; schema
(tables, indexes) is managed by the application via Wrangler migrations, not at
this layer.

## Example

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareD1Database
metadata:
  name: d1-hack-database
spec:
  databaseName: planton-d1-hack
  accountId: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
  region: wnam
  readReplication:
    mode: disabled
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.accountId` | `string` | yes |  |  |
| `spec.databaseName` | `string` | yes |  |  |
| `spec.region` | `enum` |  |  |  |
| `spec.readReplication` | `CloudflareD1ReadReplication` |  |  |  |
| `spec.readReplication.mode` | `enum` | yes |  |  |
| `spec.jurisdiction` | `string` |  |  |  |

## Field Details

### spec.accountId

`string` · required

The Cloudflare account ID in which to create the database.

- rule: {"required":true,"string":{"len":"32","pattern":"^[0-9a-fA-F]{32}$"}}

### spec.databaseName

`string` · required

The unique name for the D1 database, unique within the account.

- rule: {"required":true,"string":{"maxLen":"64"}}

### spec.region

`enum`

Region hint for the database's primary instance (maps to
primary_location_hint). Fixed at creation, and Cloudflare never returns
it on read, so post-create edits are deliberately INERT in both engines
(live-verified 2026-08-26: without the ignore, the replace-forcing hint
would destroy the database -- data included -- to move a hint). To
change placement, recreate the database deliberately. Leave unspecified
to let Cloudflare choose, or use jurisdiction for a data-residency
constraint.

Allowed values (use exactly as shown):

- `cloudflare_d1_region_unspecified` -- Unspecified region (Cloudflare selects a default).
- `weur` -- Western Europe.
- `eeur` -- Eastern Europe.
- `apac` -- Asia-Pacific.
- `oc` -- Oceania.
- `wnam` -- Western North America.
- `enam` -- Eastern North America.

### spec.readReplication

`CloudflareD1ReadReplication`

Configures D1 Read Replication: read-only replicas in multiple regions for
lower global read latency. WARNING: enabling replication requires the Worker
to use the D1 Sessions API for consistency; omit to disable. Omitting is
exact, not approximate: Cloudflare always reports the mode ("disabled"
when never configured), and both engines send that server default
explicitly so an omitted block converges (live-verified 2026-08-26;
`auto` is available on the free plan).

### spec.readReplication.mode

`enum` · required

Replication mode. `auto` enables cross-region read replicas; `disabled`
keeps a single primary.

- rule: read_replication.mode must be "auto" or "disabled"
- rule: {"required":true,"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `read_replication_mode_unspecified` -- Unspecified (treated as the API default).
- `auto` -- Automatic read replication across regions.
- `disabled` -- Replication disabled (single primary).

### spec.jurisdiction

`string`

Data-residency jurisdiction, fixed at creation. One of "eu" (European Union)
or "fedramp" (US FedRAMP). Constrains where the database is placed; mutually
exclusive with region. Leave empty for no residency constraint.

- rule: jurisdiction must be one of "eu", "fedramp"

## Validation Rules

- `spec.region_xor_jurisdiction`: set at most one of region or jurisdiction (jurisdiction already constrains placement)

## Outputs

Reference an output from another manifest as `valueFrom: {kind: CloudflareD1Database, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.database_id` | `string` | The unique identifier of the created D1 database. A Worker's `d1` binding references this value. |
| `status.outputs.database_name` | `string` | The name of the database (same as the input name). |
| `status.outputs.version` | `string` | The D1 engine version Cloudflare reports for the database (server-assigned, e.g. "production"). Informational; not used for wiring. |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| CloudflarePagesProject | `spec.deploymentConfigs.preview.d1Databases[].databaseId` | `status.outputs.database_id` |
| CloudflarePagesProject | `spec.deploymentConfigs.production.d1Databases[].databaseId` | `status.outputs.database_id` |
| CloudflareWorker | `spec.d1Databases[].databaseId` | `status.outputs.database_id` |

## See Also

- [Overview](../README.md)
