# Cloud SQL PostgreSQL Source

## Use Case

Point Datastream at a Cloud SQL for PostgreSQL database over its public IP -- Google's direct path -- with the replication user's password kept in Secret Manager instead of the manifest.

## When to Use

- Streaming an operational Cloud SQL database into BigQuery or Cloud Storage
- Teams that keep database credentials in Secret Manager

## What This Creates

- A PostgreSQL source profile in `us-central1` for the `orders` database on the `orders-db` instance

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `postgresqlProfile.hostname` | `GcpCloudSql` public IP | For a private instance, use a private connection and a proxy VM's address instead. |
| `postgresqlProfile.secretManagerStoredPassword` | `GcpSecretManagerSecret` version | Grant Datastream's service agent `roles/secretmanager.secretAccessor` on the secret (its `iamMembers`). |
| `postgresqlProfile.database` | `orders` | The database whose publication and replication slot the stream uses. |

Before a stream starts, allowlist Datastream's regional IPs in the instance's authorized networks and turn on `cloudsql.logical_decoding`.
