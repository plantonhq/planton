# OCI Autonomous Database — Design Notes

## Why One Component for All Workload Types

OCI Autonomous Database supports five workload types: OLTP (ATP), Data Warehouse (ADW), JSON Database (AJD), APEX, and Lakehouse (LH). Despite different optimizer profiles and feature availability, the underlying OCI API resource is the same — `oci_database_autonomous_database`. The workload type is a single field (`db_workload`) that determines internal behavior.

Splitting these into separate OpenMCF components (e.g., OciAutonomousTransactionProcessing, OciAutonomousDataWarehouse) would mean:

- Five near-identical proto definitions differing only in enum default.
- Five near-identical Pulumi modules with duplicated Go code.
- Five separate catalog entries that users would need to discover and choose between.

A single OciAutonomousDatabase component with a `dbWorkload` selector keeps the API surface small. Users pick the workload type in their manifest; everything else — compute, storage, networking, encryption, backups — is shared configuration.

The trade-off is that some fields are only meaningful for certain workload types (e.g., `licenseModel` is ignored for AJD/APEX workloads because they always use LICENSE_INCLUDED). This is documented in field descriptions rather than enforced by separate types.

## Compute Model: ECPU vs OCPU

Oracle introduced ECPUs as a replacement for OCPUs. The component supports both via the `computeModel` enum. The Go module passes the value to the Pulumi resource as-is (uppercased). No conversion or normalization is applied.

The spec uses `optional float compute_count` (proto3 optional) so that omitting the field is distinguishable from setting it to zero. The Go module only sets `ComputeCount` on the Pulumi args when the proto field is non-nil.

## Storage Size Mutual Exclusivity

The OCI API accepts storage in either terabytes (`data_storage_size_in_tbs`) or gigabytes (`data_storage_size_in_gb`), but not both. The proto definition enforces this with a CEL validation rule:

```
!(this.data_storage_size_in_tbs > 0 && this.data_storage_size_in_gb > 0)
```

Terabytes are typical for serverless deployments. Gigabytes provide finer granularity for dedicated Exadata infrastructure where databases share a pool of storage.

## Credential Handling

Two mutually exclusive options exist for the database admin password:

1. **`adminPassword`** — plaintext in the manifest. Acceptable for development/testing but not recommended for production.
2. **`secretId`** + optional `secretVersionNumber` — references an OCI Vault secret. The Pulumi resource reads the secret value at deploy time.

A CEL validation rule ensures only one is set:

```
!(this.admin_password != '' && has(this.secret_id))
```

The Go module checks each independently and sets the appropriate Pulumi arg. There is no fallback or default — if neither is set, the OCI API will reject the request.

## Foreign Key References

Six fields accept `StringValueOrRef`, enabling `valueFrom` references to other OpenMCF-managed resources:

| Field | Default Referenced Kind | Referenced Field Path |
|-------|------------------------|----------------------|
| `compartmentId` | OciCompartment | `status.outputs.compartmentId` |
| `subnetId` | OciSubnet | `status.outputs.subnetId` |
| `nsgIds` | OciSecurityGroup | `status.outputs.networkSecurityGroupId` |
| `secretId` | — (no default kind) | — |
| `kmsKeyId` | — (no default kind) | — |
| `vaultId` | — (no default kind) | — |
| `autonomousContainerDatabaseId` | — (no default kind) | — |

Fields without a default kind require explicit `kind` and `fieldPath` in the `valueFrom` block, or a literal `value`.

## Tagging Strategy

The Go module in `locals.go` builds freeform tags from:

1. Static tags: `resource: "true"`, `resource_kind: "OciAutonomousDatabase"`, `resource_id: <metadata.id>`
2. Conditional tags: `organization` (from `metadata.org`), `environment` (from `metadata.env`)
3. User labels: all entries from `metadata.labels` are merged in

These are applied as `freeformTags` on the Autonomous Database resource. OCI defined tags are not used — they require pre-existing tag namespaces which adds operational overhead.

## Display Name Fallback

If `spec.displayName` is empty, the Go module falls back to `metadata.name`. This is implemented in `initializeLocals()` and the resulting `locals.DisplayName` is used both as the Pulumi resource name and the OCI display name.

## What Is Deferred

The following capabilities are not yet implemented:

- **Clone from existing ADB** — the OCI API supports cloning from a source database or backup. This would add `source`, `source_id`, `clone_type`, and `timestamp` fields. Deferred to avoid proto complexity until there is user demand.
- **Refreshable clones** — a read-only clone that periodically refreshes from a source. Requires additional fields for refresh mode and interval.
- **Cross-region Data Guard** — the current implementation only supports local (same-region) Data Guard via `isLocalDataGuardEnabled`. Cross-region requires a separate standby resource and peer configuration.
- **Scheduled operations** — the OCI API supports scheduled start/stop and scaling operations. This would require a repeated message for schedules.
- **Resource pools** — the ability to share resources across multiple ADBs using resource pools. Requires pool OCID and pool leader references.
- **Long-term backups** — configuring long-term backup retention beyond the standard automatic backup window. Requires additional fields for retention type and schedule.
- **Connection wallet management** — downloading and managing the connection wallet is handled outside the Pulumi resource and is left to operators.
