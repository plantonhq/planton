# GCP Firestore Database

Deploys a Cloud Firestore database with configurable type, edition, PITR, CMEK encryption, and delete protection.

## What Gets Created

When you deploy a GcpFirestoreDatabase resource, Planton provisions:

- **Firestore Database** -- a `google_firestore_database` resource with the chosen type (Native or Datastore mode), location, edition, and optional CMEK encryption

## Firestore Database Architecture

```
GCP Project
└── Firestore Database (type + location + edition)
    ├── Collections
    │   ├── Documents
    │   └── Subcollections
    └── Indexes (managed by application, not IaC)
```

A Firestore database is a serverless, fully managed NoSQL document store. Unlike relational databases, there is no compute instance to manage -- GCP handles scaling, replication, and availability transparently.

### Multi-Region vs Single-Region

| Location Type | Examples | Availability | Latency | Cost |
|---|---|---|---|---|
| Multi-region | `nam5` (US), `eur3` (Europe) | Higher (cross-region replication) | Slightly higher writes | Higher |
| Single-region | `us-east1`, `europe-west1` | Standard (single-region) | Lower | Lower |

Multi-region locations replicate data across multiple GCP regions automatically, providing higher availability at the cost of slightly higher write latency and storage costs.

## Database Type Deep Dive

### FIRESTORE_NATIVE

The modern Firestore experience:
- Real-time listeners for live data synchronization
- Offline support in mobile and web SDKs
- Hierarchical data model with collections and subcollections
- Strong consistency for all reads
- Supports both STANDARD and ENTERPRISE editions
- Default concurrency mode: OPTIMISTIC

### DATASTORE_MODE

Legacy compatibility mode:
- Datastore client library API
- Entity-group transactions
- No real-time listeners or offline support
- Only supports STANDARD edition
- Default concurrency mode: PESSIMISTIC
- Available concurrency modes: PESSIMISTIC, OPTIMISTIC_WITH_ENTITY_GROUPS

Most new applications should use FIRESTORE_NATIVE. DATASTORE_MODE is for existing Datastore applications that need to continue using the Datastore API.

## Database Edition

| Edition | Features | Requirements |
|---|---|---|
| STANDARD | Standard Firestore SLA and features | Any database type |
| ENTERPRISE | Enhanced SLA, advanced security, additional data access modes | FIRESTORE_NATIVE only |

ENTERPRISE edition is immutable after creation. It enables additional capabilities (MongoDB-compatible API, enhanced security) that are auto-enabled with sensible defaults.

## 80/20 Scoping Rationale

### What We Include

| Feature | Included | Rationale |
|---|---|---|
| Database creation with type | Yes | Core lifecycle event |
| Location selection | Yes | Data residency and latency |
| Database edition | Yes | Major capability/pricing lever |
| CMEK encryption | Yes | Compliance requirement |
| PITR (point-in-time recovery) | Yes | Disaster recovery essential |
| Delete protection | Yes | Production safety guard |
| Concurrency mode | Yes | Transaction behavior control |

### What We Exclude

| Feature | Excluded | Rationale |
|---|---|---|
| App Engine integration mode | Yes | Legacy feature; modern deployments don't use it |
| Resource Manager tags | Yes | Advanced organizational feature; not infrastructure config |
| Firestore data access mode | Yes | Enterprise-only, auto-enabled with defaults; v2 candidate |
| MongoDB-compatible mode | Yes | Enterprise-only, auto-enabled with defaults; v2 candidate |
| Realtime updates mode | Yes | Enterprise-only, auto-enabled with defaults; v2 candidate |
| Deletion policy | Yes | IaC-internal concern; hardcoded to DELETE for lifecycle management |
| Indexes | Yes | Application-level concern managed by Firestore SDKs or firebase.json |
| Security rules | Yes | Application-level concern, typically in version control alongside app code |
| Backup schedules | Yes | Operational concern managed separately |

### Deliberate Design Choices

**`deletion_policy` hardcoded to DELETE:** The GCP API / Terraform has a `deletion_policy` field with "DELETE" or "ABANDON". We hardcode it to "DELETE" in both the Pulumi and Terraform modules so that IaC tools manage the full lifecycle. Users who want protection against accidental deletion should use `delete_protection_state = DELETE_PROTECTION_ENABLED`, which is a GCP API-level guard that works across all interfaces.

**No App Engine integration mode:** The `app_engine_integration_mode` field controls legacy App Engine integration that 99% of modern Firestore deployments don't use. Including it would confuse users and add a field with no practical value for new applications. GCP defaults it sensibly.

**CMEK via `kms_key_name` (not `cmek_config` sub-message):** We flatten the CMEK configuration to a single `kms_key_name` field for consistency with other Planton components (GcpSpannerDatabase, GcpBigQueryDataset, GcpBigtableInstance). The Pulumi/Terraform modules wrap it in the appropriate `cmek_config`/`CmekConfig` structure internally.

**No GCP labels:** Firestore databases do not support GCP labels. This is a GCP platform limitation. Resource Manager tags are available but are an advanced organizational feature that operates differently from labels (immutable, requires tag key IDs). Excluded from v1.

## Encryption: Google-Managed vs CMEK

| Feature | Google-Managed (default) | CMEK |
|---|---|---|
| Key management | Automatic | User manages key in KMS |
| Compliance | Meets most requirements | Required for HIPAA, PCI-DSS, FedRAMP |
| Key rotation | Automatic | User-controlled |
| Location constraint | None | Key must be in same location as database |
| Cost | Free | KMS key charges apply |
| Mutability | N/A | Immutable (changing key requires recreation) |

Location mapping for CMEK with multi-region databases:
- `nam5` (Firestore) → `us` (Cloud KMS multi-region)
- `eur3` (Firestore) → `europe` (Cloud KMS multi-region)

## Infra Chart Composition Patterns

### Pattern 1: Standalone Database (most common)
```
GcpProject (Layer 0)
└── GcpFirestoreDatabase (Layer 1, references project via valueFrom)
```

### Pattern 2: Database + KMS (enterprise)
```
GcpProject (Layer 0)
├── GcpKmsKeyRing (Layer 1)
│   └── GcpKmsKey (Layer 2)
│       └── GcpFirestoreDatabase (Layer 3, references key via valueFrom)
└── GcpFirestoreDatabase (Layer 1, no CMEK)
```

### Pattern 3: Multi-database project
```
GcpProject (Layer 0)
├── GcpFirestoreDatabase "(default)" (Layer 1)
├── GcpFirestoreDatabase "orders" (Layer 1)
└── GcpFirestoreDatabase "analytics" (Layer 1)
```

## Provider Version Requirements

- **Terraform**: `~> 6.0` (required for `cmek_config` and `database_edition`)
- **Pulumi**: `v9` (pulumi-gcp SDK)

## References

- [Cloud Firestore Documentation](https://cloud.google.com/firestore/docs)
- [Terraform google_firestore_database](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/firestore_database)
- [Pulumi gcp.firestore.Database](https://www.pulumi.com/registry/packages/gcp/api-docs/firestore/database/)
- [Firestore Locations](https://cloud.google.com/firestore/docs/locations)
- [Firestore CMEK](https://cloud.google.com/firestore/docs/cmek)
