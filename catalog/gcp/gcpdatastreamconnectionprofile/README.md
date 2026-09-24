# GCP Datastream Connection Profile

Where one source database or destination is, and how Datastream signs in. Exactly one profile type: `bigqueryProfile` or `gcsProfile` for destinations, or `mysqlProfile`, `postgresqlProfile`, `oracleProfile`, `sqlServerProfile`, or `mongodbProfile` for sources. Sources connect over public connectivity (Datastream's regional IPs allowlisted), a `GcpDatastreamPrivateConnection`, or an SSH tunnel. One profile serves every stream that reads the same database or writes the same destination.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `datastream.googleapis.com` on the project (never disabled on destroy)
- **Connection profile** -- a `datastream_connection_profile` with its one profile type

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with Datastream admin permissions (`roles/datastream.admin`) on the project.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Optional Dependencies

- **`GcpCloudSql`** -- a Cloud SQL source, by its public IP (`*Profile.hostname`).
- **`GcpSecretManagerSecret`** -- the password's secret version (`secretManagerStoredPassword`, its `latest_version_name`).
- **`GcpGcsBucket`** -- a Cloud Storage destination (`gcsProfile.bucket`).
- **`GcpDatastreamPrivateConnection`** -- private reachability (`privateConnection`).

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpDatastreamConnectionProfile
metadata:
  name: bigquery
spec:
  location: us-central1
  bigqueryProfile: true
```

```shell
planton apply -f datastream-connection-profile.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `location` | `string` | The Datastream region, e.g. `us-central1`. Immutable. |
| one profile type | -- | Exactly one of `bigqueryProfile`, `gcsProfile`, `mysqlProfile`, `postgresqlProfile`, `oracleProfile`, `sqlServerProfile`, `mongodbProfile`. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider project | The project. |
| `connectionProfileId` | `string` | `metadata.name` | Immutable. |
| `displayName` | `string` | `metadata.name` | Console name. |
| `bigqueryProfile` | `bool` | `false` | Declares a BigQuery destination. |
| `gcsProfile` | `object` | -- | `bucket` (`GcpGcsBucket` reference), `rootPath`. |
| `mysqlProfile` | `object` | -- | `hostname` (`GcpCloudSql` public IP by default), `port`, `username`, `password` or `secretManagerStoredPassword`, `sslConfig`. |
| `postgresqlProfile` | `object` | -- | `hostname`, `port`, `username`, password, `database`, `sslConfig` (server or server-and-client verification). |
| `oracleProfile` | `object` | -- | `hostname`, `port`, `username`, password, `databaseService`, `connectionAttributes`. |
| `sqlServerProfile` | `object` | -- | `hostname`, `port`, `username`, password, `database`. |
| `mongodbProfile` | `object` | -- | `hostAddresses`, `username`, password, `replicaSet`, SRV or standard format, `sslConfig`, `additionalOptions`. |
| `privateConnection` | `StringValueOrRef` | public connectivity | A `GcpDatastreamPrivateConnection`. |
| `forwardSshConnectivity` | `object` | -- | An SSH bastion: `hostname`, `port`, `username`, `password` or `privateKey`. |
| `createWithoutValidation` | `bool` | `false` | Skip Google's connectivity test. Immutable. |
| `labels` | `map<string,string>` | none | Merged with the platform attribution labels. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE`, `PREVENT`, or `ABANDON`. |

### Validation Rules

- Exactly one profile type; at most one of `privateConnection` and `forwardSshConnectivity`.
- A password and a Secret Manager password are mutually exclusive on every source; so are an SSH password and key.
- MongoDB: exactly one of SRV and standard format, and no `replicaSet` with SRV; at least one host.
- TLS: a MySQL or MongoDB client certificate needs its key and the CA certificate; PostgreSQL takes at most one verification mode.
- Ports are 1-65535; empty uses the engine's default.

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | `projects/{project}/locations/{location}/connectionProfiles/{id}` -- what streams reference |
| `connection_profile_id` | `string` | The profile's id |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Cloud SQL is reached at its public IP by default.** Allowlist Datastream's regional IPs in the instance's authorized networks; for a private instance use a private connection and a proxy VM.
- **Keep passwords in Secret Manager.** `secretManagerStoredPassword` names a secret version; grant Datastream's service agent (`service-{project-number}@gcp-sa-datastream.iam.gserviceaccount.com`) `roles/secretmanager.secretAccessor` through the secret's `iamMembers`. A literal `password` is sensitive and never returned.
- **Salesforce and Spanner profiles** exist only in Google's beta provider today; a stream names such a profile by its full resource name.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Components

- **GcpDatastreamStream** -- streams read and write through profiles
- **GcpDatastreamPrivateConnection** -- private reachability
- **GcpCloudSql** / **GcpSecretManagerSecret** / **GcpGcsBucket** -- what profiles point at

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
