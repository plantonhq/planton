# Archive Storage

This preset creates an archive-tier Object Storage bucket with a 7-year retention rule for compliance data. Archive storage offers the lowest per-GB cost in OCI Object Storage, suitable for data that must be retained for regulatory reasons but is rarely if ever accessed. Objects require a restore request before they can be read (restore time: typically 1 hour).

## When to Use

- Regulatory compliance archives (financial records, healthcare data, audit logs) with multi-year retention requirements
- Database backups and snapshots that must be retained but are only accessed during disaster recovery
- Log archives where raw logs are preserved for forensic analysis but not actively queried
- Any data with a "write once, read rarely" access pattern where storage cost is the primary concern

## Key Configuration Choices

- **Archive storage tier** (`storageTier: archive`) -- lowest per-GB cost in OCI. Objects are stored offline and require a restore request (typically 1 hour) before they can be read. This tier is immutable after bucket creation.
- **7-year retention rule** (`retentionRules` with 7 years) -- objects cannot be deleted or overwritten before the retention period expires. This satisfies common regulatory requirements (SOX, HIPAA, PCI-DSS). Adjust the duration to match your compliance policy.
- **Versioning disabled** -- archive buckets typically store immutable data (write once). Versioning adds complexity and cost without benefit when objects are never overwritten.
- **Private access** (`accessType: no_public_access`) -- compliance data should never be publicly accessible.
- **No encryption key specified** -- uses Oracle-managed encryption by default. Add `kmsKeyId` if your compliance policy requires customer-managed keys.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment for the bucket | OCI Console > Identity > Compartments, or `OciCompartment` outputs |
| `<object-storage-namespace>` | Object Storage namespace for your tenancy | `oci os ns get` CLI command, or OCI Console > Object Storage |
| `<bucket-name>` | Globally unique bucket name within the namespace | Choose a name (e.g., `myorg-compliance-archive`) |

## Related Presets

- **01-private-versioned** -- Use instead for application data that is frequently accessed and needs versioning
- **03-public-read** -- Not applicable for compliance data; listed for completeness
