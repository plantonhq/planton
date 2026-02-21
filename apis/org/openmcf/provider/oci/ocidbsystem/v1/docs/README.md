# OciDbSystem — Design Documentation

## Design Rationale

OciDbSystem wraps the `oci_database_db_system` resource into a single OpenMCF declarative manifest. The OCI Database service models a DB System as three inseparable layers created together: the system (compute + storage), the DB Home (Oracle software installation), and the initial database instance. This component mirrors that structure by nesting `DbHome` and `Database` messages inside `OciDbSystemSpec` rather than exposing them as separate OpenMCF resources.

The decision to keep everything in one resource matches the OCI API contract — you cannot create a DB System without a DB Home and initial database. Splitting them into separate resources would require artificial ordering, cross-resource state coordination, and would not reflect the underlying API semantics.

## Architecture Decisions

### Single resource for system + home + database

The OCI `LaunchDbSystem` API requires all three layers in a single call. Modeling them as three separate OpenMCF resources would create a false impression that they can be independently lifecycle-managed. At creation time they cannot. Post-creation operations on DB Homes and databases (adding additional DB Homes, creating additional databases within a Home) are different API calls that may be modeled as separate resources in the future.

### Fresh creation only (source=NONE)

Clone and restore operations (`source=DB_BACKUP`, `DATABASE`, `DB_SYSTEM`) require a fundamentally different field set — backup OCIDs, source database OCIDs, point-in-time timestamps — and represent a different operational workflow. Including them would overload the spec with mutually exclusive field groups that complicate validation. They are deferred to a future version or a separate resource kind (e.g., `OciDbSystemRestore`).

### Enum string mapping

Proto enums use lowercase snake_case values (e.g., `enterprise_edition`, `bring_your_own_license`). The Pulumi module converts these to uppercase strings (`ENTERPRISE_EDITION`, `BRING_YOUR_OWN_LICENSE`) via `strings.ToUpper()` before passing them to the OCI provider. This keeps the manifest readable while matching the OCI API's expected format.

### Deprecated and excluded fields

- **`dbWorkload`** — deprecated by Oracle since November 2023. Omitted entirely rather than carrying forward a deprecated field.
- **`computeModel` / `computeCount`** — Exadata Cloud@Customer specific. VM and BM shapes use `cpuCoreCount` instead. Including these would add confusion for the 95%+ of users not running Exadata.
- **`backupDestinationDetails`** — advanced backup routing to non-default destinations. The OCI Database Backup and Recovery Service (DBRS) default covers most cases. This can be added later without breaking changes.

### StringValueOrRef for cross-resource composition

Fields like `compartmentId`, `subnetId`, `nsgIds`, `backupSubnetId`, `backupNetworkNsgIds`, `kmsKeyId`, and database-level `kmsKeyId`/`vaultId` accept either a literal `value` or a `valueFrom` reference pointing to another OpenMCF resource's stack output. This enables composability without requiring users to manually copy OCIDs between resources.

Each `StringValueOrRef` field has a `default_kind` and `default_kind_field_path` annotation that tells the OpenMCF resolver which resource kind and output field to look up when the user provides just a `name` in the `valueFrom` block.

### Display name fallback

When `displayName` is empty, the Pulumi module falls back to `metadata.name`. This is implemented in `locals.go` and keeps manifests concise — users only set `displayName` when they want a different OCI Console label.

### Freeform tags

The module automatically applies freeform tags from metadata:
- `resource: true`
- `resource_kind: OciDbSystem`
- `resource_id: <metadata.id>`
- `organization: <metadata.org>` (when set)
- `environment: <metadata.env>` (when set)
- All custom labels from `metadata.labels`

These tags provide cost attribution and resource discovery in the OCI Console without requiring users to configure tagging manually.

## Trade-offs

| Decision | Benefit | Cost |
|----------|---------|------|
| Single resource for system + home + database | Matches OCI API semantics; simpler manifest | Cannot independently manage additional DB Homes or databases post-creation via this resource |
| Fresh creation only | Cleaner validation; focused field set | Users who need clone/restore must use OCI Console or a future resource kind |
| Excluded deprecated fields | Smaller API surface; no confusion from deprecated options | Users who need `dbWorkload` must set it outside OpenMCF |
| Enum as proto enum (not string) | Compile-time validation; auto-generated docs | Requires uppercase conversion in the Pulumi module |
| Nested messages (DataCollectionOptions, DbSystemOptions, etc.) | Groups related fields logically | Deeper YAML nesting in manifests |

## What's Deferred

1. **Clone/restore scenarios** — `source=DB_BACKUP`, `DATABASE`, `DB_SYSTEM` workflows with their distinct field requirements.
2. **Additional DB Home management** — creating additional DB Homes within an existing DB System post-creation.
3. **Additional database management** — creating additional databases within an existing DB Home.
4. **Exadata Cloud@Customer fields** — `computeModel`, `computeCount`, and other Exadata-specific configuration.
5. **Advanced backup destinations** — `backupDestinationDetails` for routing backups to non-default locations.
6. **Data Guard configuration** — standby database association and switchover/failover management.
7. **Database upgrade management** — orchestrating version upgrades across DB Home and database layers.
8. **Patching operations** — applying specific patches beyond the maintenance window auto-patching.

## Module Structure

```
v1/
├── api.proto                  # Top-level OciDbSystem message
├── spec.proto                 # OciDbSystemSpec with all nested messages and enums
├── stack_outputs.proto        # OciDbSystemStackOutputs (4 outputs)
└── iac/pulumi/module/
    ├── main.go                # Entry point: initializes locals, provider, calls dbSystem()
    ├── locals.go              # Locals struct: display name fallback, freeform tag assembly
    ├── db_system.go           # DB System creation + all builder functions
    └── outputs.go             # Output key constants
```

The Pulumi module follows a consistent pattern across OpenMCF OCI components:
- `main.go` orchestrates resource creation
- `locals.go` computes derived values from the manifest
- A resource-specific file (here `db_system.go`) contains the actual resource creation and builder functions
- `outputs.go` defines output key constants to avoid string literals in export calls
