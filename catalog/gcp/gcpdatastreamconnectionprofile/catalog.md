# GCP Datastream Connection Profile

Tells Datastream where a database or destination is and how to sign in -- MySQL, PostgreSQL, Oracle, SQL Server, or MongoDB sources, BigQuery or Cloud Storage destinations -- once, so every stream that reads that database or writes that destination reuses it.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `datastream.googleapis.com` on the project (never disabled on destroy)
- **Connection profile** -- a `datastream_connection_profile` with its one profile type

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with Datastream admin permissions (`roles/datastream.admin`) on the project. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Optional Dependencies

- **`GcpCloudSql`** -- a Cloud SQL source, by its public IP.
- **`GcpSecretManagerSecret`** -- the password's secret version.
- **`GcpGcsBucket`** -- a Cloud Storage destination.
- **`GcpDatastreamPrivateConnection`** -- private reachability.

## Deploy

### Console

Open the deployment store, find **GCP Datastream Connection Profile**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Cloud SQL PostgreSQL Source** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpDatastreamConnectionProfile
metadata:
  name: bigquery
  org: acme-corp
  env: prod
spec:
  location: us-central1
  bigqueryProfile: true
```

```shell
planton apply -f datastream-connection-profile.yaml
```

This creates the BigQuery destination profile every stream in `us-central1` can share. A Stack Job tracks the provisioning in real time.

### InfraChart

Reference a `GcpCloudSql` from `postgresqlProfile.hostname` and a `GcpSecretManagerSecret` from `secretManagerStoredPassword`; a `GcpDatastreamStream` in the same chart references this profile.

## Key Configuration

These are the most important decisions when configuring this component. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The profile type** -- a destination (BigQuery, Cloud Storage) or a source engine; pick exactly one.

**Connectivity** -- public (allowlist Datastream's IPs), a private connection, or an SSH bastion.

**The credential** -- prefer a Secret Manager version over a literal password, and grant Datastream's service agent access to it.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpCloudSql** | `mysqlProfile.hostname`, `postgresqlProfile.hostname`, `sqlServerProfile.hostname` | `status.outputs.public_ip` |
| **GcpSecretManagerSecret** | `*.secretManagerStoredPassword`, `mongodbProfile.sslConfig.secretManagerStoredClientKey` | `status.outputs.latest_version_name` |
| **GcpGcsBucket** | `gcsProfile.bucket` | `status.outputs.bucket_name` |
| **GcpDatastreamPrivateConnection** | `privateConnection` | `status.outputs.name` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|-----------------------|
| `name` | The profile's full resource name | `GcpDatastreamStream` source and destination profiles |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Cloud SQL PostgreSQL source** -- A Cloud SQL database over its public IP with the password in Secret Manager. Start from the **Cloud SQL PostgreSQL Source** preset.

**BigQuery destination** -- The one destination every stream shares. Start from the **BigQuery Destination** preset.

**Private MySQL** -- A MySQL server with no public address, over a private connection with TLS. Start from the **MySQL over a Private Connection** preset.

## Works With

- [**GCP Datastream Stream**](/cloud-catalog/gcp-datastream-stream) -- streams through profiles
- [**GCP Datastream Private Connection**](/cloud-catalog/gcp-datastream-private-connection) -- private reachability
- [**GCP Cloud SQL**](/cloud-catalog/gcp-cloud-sql) -- a source database
- [**GCP Secret Manager Secret**](/cloud-catalog/gcp-secret-manager-secret) -- the credential
