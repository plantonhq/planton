# GCP Spanner Application

Provisions a production-ready Cloud Spanner deployment with a Spanner instance, database, dedicated service account, and optional VPC networking for the application environment.

Cloud Spanner is a fully managed, globally distributed, strongly consistent relational database. This chart sets up the database tier along with the IAM identity your application needs to connect, so you can focus on building your application rather than wiring infrastructure.

## Architecture

```
                    ┌──────────────────┐
                    │  GcpServiceAccount│
                    │  (spanner-app-sa) │
                    │  roles/spanner.   │
                    │  databaseUser     │
                    └──────────────────┘

┌──────────────────────────────────────────┐
│              Spanner                      │
│                                          │
│  ┌──────────────────┐                    │
│  │ GcpSpannerInstance│                    │
│  │  (1+ nodes)      │                    │
│  └────────┬─────────┘                    │
│           │                              │
│           ▼                              │
│  ┌──────────────────┐                    │
│  │ GcpSpannerDatabase│                    │
│  │  (GoogleSQL/PG)  │                    │
│  └──────────────────┘                    │
└──────────────────────────────────────────┘

┌──────────────────────────────────────────┐
│      Optional: Networking                 │
│                                          │
│  ┌──────────────────┐                    │
│  │    GcpVpc        │                    │
│  └────────┬─────────┘                    │
│           │                              │
│           ▼                              │
│  ┌──────────────────┐                    │
│  │ GcpFirewallRule   │                    │
│  │  (allow-internal) │                    │
│  └──────────────────┘                    │
└──────────────────────────────────────────┘
```

## Dependency Graph

```
Layer 0 (parallel):  GcpVpc, GcpServiceAccount, GcpSpannerInstance
Layer 1 (depends):   GcpFirewallRule (← VPC), GcpSpannerDatabase (← SpannerInstance)
```

## Included Cloud Resources

| Resource | Kind | Group | Purpose |
|----------|------|-------|---------|
| VPC Network | `GcpVpc` | network | Application networking (optional) |
| Firewall Rule | `GcpFirewallRule` | network | Allow internal traffic (optional) |
| Service Account | `GcpServiceAccount` | identity | Application identity with Spanner access |
| Spanner Instance | `GcpSpannerInstance` | database | Compute capacity for Spanner databases |
| Spanner Database | `GcpSpannerDatabase` | database | The application database |

## Parameters

| Parameter | Description | Default | Required |
|-----------|-------------|---------|----------|
| `gcp_project_id` | GCP project ID | `my-gcp-project` | Yes |
| `instance_name` | Spanner instance name (6-30 chars) | `my-spanner-instance` | Yes |
| `display_name` | Human-readable instance name | `My Spanner Instance` | Yes |
| `instance_config` | Geographic placement (e.g., `regional-us-central1`) | `regional-us-central1` | Yes |
| `num_nodes` | Node count (each ~10K QPS reads) | `1` | Yes |
| `edition` | STANDARD, ENTERPRISE, or ENTERPRISE_PLUS | (empty = default) | No |
| `database_name` | Database name (2-30 chars) | `my-database` | Yes |
| `database_dialect` | GOOGLE_STANDARD_SQL or POSTGRESQL | `GOOGLE_STANDARD_SQL` | Yes |
| `enable_drop_protection` | API-level deletion protection | `false` | No |
| `service_account_id` | Service account for the application | `spanner-app-sa` | Yes |
| `networkingEnabled` | Create VPC and firewall | `true` | No |
| `vpc_name` | VPC network name | `spanner-app-vpc` | No |

## What Gets Created

### Always Created
- **Spanner Instance**: Allocated compute capacity in the specified region/multi-region
- **Spanner Database**: Application database with chosen SQL dialect
- **Service Account**: Granted `roles/spanner.databaseUser` for read/write access to all databases in the project

### Conditionally Created (networkingEnabled: true)
- **VPC Network**: Custom-mode VPC for the application environment
- **Firewall Rule**: Allows all internal traffic (10.0.0.0/8) for application-to-application communication

## Usage Notes

### Connecting Your Application

The service account created by this chart has `roles/spanner.databaseUser`, which provides read/write access to Spanner databases. Your application should authenticate using this service account:

```go
// Go example
client, err := spanner.NewClient(ctx, "projects/my-project/instances/my-instance/databases/my-db")
```

### Spanner Networking

Cloud Spanner is a fully managed service — it does not require VPC peering or private networking to function. Applications connect via the Spanner client libraries using IAM authentication. The optional VPC in this chart provides networking for your **application compute** (GCE instances, GKE pods, Cloud Run services), not for Spanner itself.

For private connectivity to Spanner (avoiding public internet), enable [Private Google Access](https://cloud.google.com/vpc/docs/private-google-access) on the subnet where your application runs.

### Scaling

- **1 node** = ~10,000 QPS reads or ~2,000 QPS writes
- For production workloads, consider using `autoscaling_config` instead of fixed `num_nodes` — modify the Spanner Instance resource directly after chart deployment

### Adding CMEK Encryption

To encrypt the database with a Customer-Managed Encryption Key, deploy a [GcpKmsKeyRing](https://github.com/plantonhq/planton) and [GcpKmsKey](https://github.com/plantonhq/planton) in the same region as the Spanner instance, then set `kmsKeyName` on the Spanner Database resource.

## Important Notes

- `instance_name`, `instance_config`, `database_name`, and `database_dialect` are **immutable** after creation. Changing them requires recreating the resource.
- The service account receives `roles/spanner.databaseUser` at the **project** level, giving it access to all Spanner databases in the project. For finer-grained access, modify the service account IAM bindings after deployment.
- Automatic backups are enabled by default on the instance (`defaultBackupScheduleType: AUTOMATIC`).
