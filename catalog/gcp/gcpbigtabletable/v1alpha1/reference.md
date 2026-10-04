# GcpBigtableTable

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

GcpBigtableTableSpec defines a table inside a Cloud Bigtable instance —
the schema-bearing, infrastructure-owned unit: column families with
their retention (GC) policies, pre-splits for load distribution, change
streams, automated backups, and deletion protection.

Tables are many-per-instance with independent lifecycles: teams add and
remove tables without touching the instance, and a table's GC policy
changes never disturb its neighbors. Data (rows and cells) is
application territory; this resource owns the structure applications
write into.

Important behavioral notes:

  - table_name, instance, and split_keys are immutable after creation.
    Changing split_keys REPLACES the table — data loss unless backed up.
  - Column families and their GC policies are mutable: add families as
    the application grows, tighten or loosen retention in place.
  - deletion_protection defaults to PROTECTED: destroying the table
    fails until it is explicitly set UNPROTECTED — a safety default for
    a data-bearing resource, applied identically by both IaC engines.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpBigtableTable
metadata:
  name: events
spec:
  # GCP project owning the instance. Replace with your project ID.
  projectId:
    value: my-gcp-project-123

  # The parent Bigtable instance's short name.
  instance:
    value: my-bigtable-instance

  # Column families with retention. Without a GC policy every cell
  # version accumulates forever.
  columnFamilies:
    - family: measurements
      gcPolicy:
        maxAge: 720h # 30 days
    - family: metadata
      gcPolicy:
        maxVersions: 3

  # Keep destroy possible for the hack manifest without a two-step
  # unprotect.
  deletionProtection: UNPROTECTED

  # Explicit DELETE (the provider default) proves the round-trip on both
  # the table and its per-family GC policies.
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.instance` | `string \| valueFrom` | yes |  | GcpBigtableInstance (`status.outputs.instance_name`) |
| `spec.tableName` | `string` |  |  |  |
| `spec.columnFamilies` | `[]GcpBigtableTableColumnFamily` |  |  |  |
| `spec.columnFamilies[].family` | `string` | yes |  |  |
| `spec.columnFamilies[].type` | `string` |  |  |  |
| `spec.columnFamilies[].gcPolicy` | `GcpBigtableTableGcPolicy` |  |  |  |
| `spec.columnFamilies[].gcPolicy.mode` | `string` |  |  |  |
| `spec.columnFamilies[].gcPolicy.maxAge` | `string` |  |  |  |
| `spec.columnFamilies[].gcPolicy.maxVersions` | `int32` |  |  |  |
| `spec.columnFamilies[].gcPolicy.gcRules` | `string` |  |  |  |
| `spec.columnFamilies[].gcPolicy.ignoreWarnings` | `bool` |  |  |  |
| `spec.splitKeys` | `[]string` |  |  |  |
| `spec.changeStreamRetention` | `string` |  |  |  |
| `spec.automatedBackupPolicy` | `GcpBigtableTableAutomatedBackupPolicy` |  |  |  |
| `spec.automatedBackupPolicy.retentionPeriod` | `string` | yes |  |  |
| `spec.automatedBackupPolicy.frequency` | `string` | yes |  |  |
| `spec.automatedBackupPolicy.locations` | `[]string` |  |  |  |
| `spec.deletionProtection` | `string` |  | `PROTECTED` |  |
| `spec.rowKeySchema` | `string` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

GCP project owning the Bigtable instance. Can be a literal project
ID or a reference to a GcpProject resource. If omitted, the
provider's default project is used.
Immutable: changing the project destroys and recreates the table.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.instance

`string | valueFrom` · required

The Bigtable instance this table lives in — the instance's short
name (a GcpBigtableInstance reference resolves to it). Immutable
after creation.

- references: GcpBigtableInstance (`status.outputs.instance_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpBigtableInstance, name: <that resource's name>, fieldPath: status.outputs.instance_name}} -- a bare string does not parse

### spec.tableName

`string`

Name for the table (1-50 characters). If not specified, defaults to
metadata.name. Immutable after creation.

- rule: table_name must be 1-50 characters of letters, numbers, underscores, hyphens, or periods

### spec.columnFamilies

`[]GcpBigtableTableColumnFamily`

Column families with their GC policies. Mutable: append families as
the application grows. At least one family is the practical minimum —
a table without families cannot store data — but the API allows an
empty table, so this is not enforced.

### spec.columnFamilies[].family

`string` · required

Column family name (unique within the table).

- rule: {"required":true,"string":{"minLen":"1","maxLen":"64"}}

### spec.columnFamilies[].type

`string`

Value type for aggregate families. The shortcuts intsum, intmin,
intmax, and inthll declare server-side aggregate cells (e.g. a
counter incremented atomically at write time); a raw JSON Type
object expresses anything else. Leave empty for regular families.

### spec.columnFamilies[].gcPolicy

`GcpBigtableTableGcPolicy`

Garbage-collection policy for this family. Strongly recommended:
without one, every cell version accumulates forever. Managed as the
API's own per-family GC policy object, so policy changes never touch
the table or its data.

- rule: gc_rules is mutually exclusive with mode, max_age, and max_versions — express the policy one way
- rule: mode combines max_age and max_versions — set both when using UNION or INTERSECTION
- rule: set mode (UNION or INTERSECTION) when combining max_age and max_versions

### spec.columnFamilies[].gcPolicy.mode

`string`

Combination mode when both max_age and max_versions are set.
UNION: collect when either condition is met (most common).
INTERSECTION: collect only when both conditions are met.

- rule: mode must be UNION or INTERSECTION

### spec.columnFamilies[].gcPolicy.maxAge

`string`

Drop cells older than this duration (e.g. "720h" for 30 days,
"8760h" for a year). Duration string in Go format.

- rule: max_age must be a duration string such as 720h or 24h30m

### spec.columnFamilies[].gcPolicy.maxVersions

`int32`

Keep only the newest N versions of each cell. 0 means unset.

- rule: {"int32":{"gte":0}}

### spec.columnFamilies[].gcPolicy.gcRules

`string`

Raw JSON GC rule tree for nested policies the typed fields cannot
express (e.g. a union of an intersection and an age rule). Mutually
exclusive with mode/max_age/max_versions. See the Bigtable Admin
API's GcRule JSON format.

### spec.columnFamilies[].gcPolicy.ignoreWarnings

`bool`

Allow a policy change that EXPANDS what is eligible for collection
on a replicated (multi-cluster) instance — Bigtable otherwise
rejects it as a safety measure against surprise data loss.

### spec.splitKeys

`[]string`

Row keys to pre-split the table at, so initial load distributes
across tablets instead of hammering one server (e.g. user prefixes
"user1", "user5", "user9"). Immutable: changing this REPLACES the
table and its data — set it right at creation, or manage splits
operationally.

### spec.changeStreamRetention

`string`

Retain change stream data (a CDC feed consumable by Dataflow) for
this duration, between 1 and 7 days (e.g. "24h0m0s"). Empty disables
change streams; setting "0" on an existing table disables them.

- rule: change_stream_retention must be a duration string between 24h and 168h, or 0 to disable

### spec.automatedBackupPolicy

`GcpBigtableTableAutomatedBackupPolicy`

Built-in automated backups: how often to take them and how long to
keep them. Omit to leave automated backups off.

### spec.automatedBackupPolicy.retentionPeriod

`string` · required

How long automated backups are retained (e.g. "72h" for 3 days).
Duration string in Go format.

- rule: {"required":true,"string":{"pattern":"^([0-9]+(\\.[0-9]+)?(h|m|s|ms))+$"}}

### spec.automatedBackupPolicy.frequency

`string` · required

How often automated backups are taken (e.g. "24h" for daily).
Duration string in Go format.

- rule: {"required":true,"string":{"pattern":"^([0-9]+(\\.[0-9]+)?(h|m|s|ms))+$"}}

### spec.automatedBackupPolicy.locations

`[]string`

Cloud Bigtable zones where automated backups are ALLOWED to be
created, each in the format "projects/{project}/locations/{zone}".
Empty means backups may be created in all zones of the instance.
Can only be set for tables in ENTERPRISE_PLUS instances (see the
GcpBigtableInstance edition field).

### spec.deletionProtection

`string` · optional (explicit presence)

API-side deletion guard. PROTECTED (the default): the table cannot
be deleted by any client until this is set UNPROTECTED first — the
safety default for a data-bearing resource, sent explicitly by both
IaC engines so destroy behavior never depends on the engine.

- default: `PROTECTED`
- rule: deletion_protection must be PROTECTED or UNPROTECTED

### spec.rowKeySchema

`string`

Structured row key schema as the API's Type JSON (declares how row
keys decompose into typed fields for SQL queries and change
streams). In-place update is not supported by the API: to change an
existing schema, clear this field, apply, then set the new schema
and apply again. Byte delimiters must be base64-encoded.

### spec.deletionPolicy

`string`

Deletion policy — what a destroy does to BOTH objects this component
manages: the table and its per-family GC policies. Applies only once
deletion_protection (above) is UNPROTECTED:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the table (and its data) is deleted; GC policies go
               with it
  "PREVENT" -- destroy FAILS; a second wall for data-bearing tables
  "ABANDON" -- the table is removed from management but left intact
               in Bigtable — also the escape hatch when a GC-policy
               delete is rejected on a replicated instance

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpBigtableTable, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.table_id` | `string` | Fully qualified table resource path (projects/{project}/instances/{instance}/tables/{table}). The canonical identifier for Admin API calls and IAM bindings. |
| `status.outputs.table_name` | `string` | Short table name — what Bigtable client libraries open, together with the project and instance. |
| `status.outputs.instance_name` | `string` | Short name of the instance the table lives in — confirms the parent without chasing the reference chain. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.instance` | GcpBigtableInstance | `status.outputs.instance_name` |

## See Also

- [Overview](../README.md)
