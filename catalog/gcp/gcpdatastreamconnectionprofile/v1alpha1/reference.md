# GcpDatastreamConnectionProfile

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpDatastreamConnectionProfileSpec defines a Datastream connection
profile (`google_datastream_connection_profile`) -- where one source
database or destination is and how Datastream signs in. Streams
reference a source profile and a destination profile; one profile
serves every stream that reads the same database or writes the same
destination.

Set exactly ONE profile type: bigquery_profile or gcs_profile
(destinations), or mysql_profile, postgresql_profile, oracle_profile,
sql_server_profile, or mongodb_profile (sources). Salesforce and Spanner
sources use profiles Google's GA provider cannot create yet; a stream
names such a profile by its full resource name.

Connectivity for a source: omit both options for public connectivity
(allowlist Datastream's regional IPs on the database), set
private_connection to go through a GcpDatastreamPrivateConnection, or
set forward_ssh_connectivity to tunnel through a bastion -- never both.

Immutable: project_id, location, connection_profile_id,
create_without_validation, certificates and keys. Hosts, users,
passwords, and connectivity update in place.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpDatastreamConnectionProfile
metadata:
  name: orders-postgres
spec:
  projectId:
    value: my-gcp-project
  location: us-central1
  displayName: Orders PostgreSQL
  postgresqlProfile:
    # A Cloud SQL instance's public IP; allowlist Datastream's regional IPs
    # in the instance's authorized networks.
    hostname:
      value: 203.0.113.10
    username: datastream
    secretManagerStoredPassword:
      value: projects/my-gcp-project/secrets/orders-datastream-password/versions/1
    database: orders
  labels:
    team: data
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.location` | `string` | yes |  |  |
| `spec.connectionProfileId` | `string` |  |  |  |
| `spec.displayName` | `string` |  |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.createWithoutValidation` | `bool` |  |  |  |
| `spec.bigqueryProfile` | `bool` |  |  |  |
| `spec.gcsProfile` | `GcpDatastreamConnectionProfileGcsProfile` |  |  |  |
| `spec.gcsProfile.bucket` | `string \| valueFrom` | yes |  | GcpGcsBucket (`status.outputs.bucket_name`) |
| `spec.gcsProfile.rootPath` | `string` |  |  |  |
| `spec.mysqlProfile` | `GcpDatastreamConnectionProfileMysqlProfile` |  |  |  |
| `spec.mysqlProfile.hostname` | `string \| valueFrom` | yes |  | GcpCloudSql (`status.outputs.public_ip`) |
| `spec.mysqlProfile.port` | `int32` |  |  |  |
| `spec.mysqlProfile.username` | `string` | yes |  |  |
| `spec.mysqlProfile.password` | `string` (sensitive) |  |  |  |
| `spec.mysqlProfile.secretManagerStoredPassword` | `string \| valueFrom` |  |  | GcpSecretManagerSecret (`status.outputs.latest_version_name`) |
| `spec.mysqlProfile.sslConfig` | `GcpDatastreamConnectionProfileMysqlSslConfig` |  |  |  |
| `spec.mysqlProfile.sslConfig.caCertificate` | `string` (sensitive) |  |  |  |
| `spec.mysqlProfile.sslConfig.clientCertificate` | `string` (sensitive) |  |  |  |
| `spec.mysqlProfile.sslConfig.clientKey` | `string` (sensitive) |  |  |  |
| `spec.postgresqlProfile` | `GcpDatastreamConnectionProfilePostgresqlProfile` |  |  |  |
| `spec.postgresqlProfile.hostname` | `string \| valueFrom` | yes |  | GcpCloudSql (`status.outputs.public_ip`) |
| `spec.postgresqlProfile.port` | `int32` |  |  |  |
| `spec.postgresqlProfile.username` | `string` | yes |  |  |
| `spec.postgresqlProfile.password` | `string` (sensitive) |  |  |  |
| `spec.postgresqlProfile.secretManagerStoredPassword` | `string \| valueFrom` |  |  | GcpSecretManagerSecret (`status.outputs.latest_version_name`) |
| `spec.postgresqlProfile.database` | `string` | yes |  |  |
| `spec.postgresqlProfile.sslConfig` | `GcpDatastreamConnectionProfilePostgresqlSslConfig` |  |  |  |
| `spec.postgresqlProfile.sslConfig.serverVerification` | `GcpDatastreamConnectionProfilePostgresqlServerVerification` |  |  |  |
| `spec.postgresqlProfile.sslConfig.serverVerification.caCertificate` | `string` (sensitive) | yes |  |  |
| `spec.postgresqlProfile.sslConfig.serverAndClientVerification` | `GcpDatastreamConnectionProfilePostgresqlServerAndClientVerification` |  |  |  |
| `spec.postgresqlProfile.sslConfig.serverAndClientVerification.caCertificate` | `string` (sensitive) | yes |  |  |
| `spec.postgresqlProfile.sslConfig.serverAndClientVerification.clientCertificate` | `string` (sensitive) | yes |  |  |
| `spec.postgresqlProfile.sslConfig.serverAndClientVerification.clientKey` | `string` (sensitive) | yes |  |  |
| `spec.oracleProfile` | `GcpDatastreamConnectionProfileOracleProfile` |  |  |  |
| `spec.oracleProfile.hostname` | `string` | yes |  |  |
| `spec.oracleProfile.port` | `int32` |  |  |  |
| `spec.oracleProfile.username` | `string` | yes |  |  |
| `spec.oracleProfile.password` | `string` (sensitive) |  |  |  |
| `spec.oracleProfile.secretManagerStoredPassword` | `string \| valueFrom` |  |  | GcpSecretManagerSecret (`status.outputs.latest_version_name`) |
| `spec.oracleProfile.databaseService` | `string` | yes |  |  |
| `spec.oracleProfile.connectionAttributes` | `map<string, string>` |  |  |  |
| `spec.sqlServerProfile` | `GcpDatastreamConnectionProfileSqlServerProfile` |  |  |  |
| `spec.sqlServerProfile.hostname` | `string \| valueFrom` | yes |  | GcpCloudSql (`status.outputs.public_ip`) |
| `spec.sqlServerProfile.port` | `int32` |  |  |  |
| `spec.sqlServerProfile.username` | `string` | yes |  |  |
| `spec.sqlServerProfile.password` | `string` (sensitive) |  |  |  |
| `spec.sqlServerProfile.secretManagerStoredPassword` | `string \| valueFrom` |  |  | GcpSecretManagerSecret (`status.outputs.latest_version_name`) |
| `spec.sqlServerProfile.database` | `string` | yes |  |  |
| `spec.mongodbProfile` | `GcpDatastreamConnectionProfileMongodbProfile` |  |  |  |
| `spec.mongodbProfile.hostAddresses` | `[]GcpDatastreamConnectionProfileMongodbHostAddress` | yes |  |  |
| `spec.mongodbProfile.hostAddresses[].hostname` | `string` | yes |  |  |
| `spec.mongodbProfile.hostAddresses[].port` | `int32` |  |  |  |
| `spec.mongodbProfile.username` | `string` | yes |  |  |
| `spec.mongodbProfile.password` | `string` (sensitive) |  |  |  |
| `spec.mongodbProfile.secretManagerStoredPassword` | `string \| valueFrom` |  |  | GcpSecretManagerSecret (`status.outputs.latest_version_name`) |
| `spec.mongodbProfile.replicaSet` | `string` |  |  |  |
| `spec.mongodbProfile.srvConnectionFormat` | `bool` |  |  |  |
| `spec.mongodbProfile.standardConnectionFormat` | `GcpDatastreamConnectionProfileMongodbStandardConnectionFormat` |  |  |  |
| `spec.mongodbProfile.standardConnectionFormat.directConnection` | `bool` |  |  |  |
| `spec.mongodbProfile.sslConfig` | `GcpDatastreamConnectionProfileMongodbSslConfig` |  |  |  |
| `spec.mongodbProfile.sslConfig.caCertificate` | `string` (sensitive) |  |  |  |
| `spec.mongodbProfile.sslConfig.clientCertificate` | `string` (sensitive) |  |  |  |
| `spec.mongodbProfile.sslConfig.clientKey` | `string` (sensitive) |  |  |  |
| `spec.mongodbProfile.sslConfig.secretManagerStoredClientKey` | `string \| valueFrom` |  |  | GcpSecretManagerSecret (`status.outputs.latest_version_name`) |
| `spec.mongodbProfile.additionalOptions` | `map<string, string>` |  |  |  |
| `spec.privateConnection` | `string \| valueFrom` |  |  | GcpDatastreamPrivateConnection (`status.outputs.name`) |
| `spec.forwardSshConnectivity` | `GcpDatastreamConnectionProfileForwardSshConnectivity` |  |  |  |
| `spec.forwardSshConnectivity.hostname` | `string` | yes |  |  |
| `spec.forwardSshConnectivity.port` | `int32` |  |  |  |
| `spec.forwardSshConnectivity.username` | `string` | yes |  |  |
| `spec.forwardSshConnectivity.password` | `string` (sensitive) |  |  |  |
| `spec.forwardSshConnectivity.privateKey` | `string` (sensitive) |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project the profile lives in: a literal project ID or a
GcpProject reference. If omitted, the provider's default project is
used. Immutable.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.location

`string` · required

The Datastream region, e.g. "us-central1". Streams using the profile
live in the same region. Immutable.

- rule: {"required":true,"string":{"pattern":"^[a-z]+-[a-z]+[0-9]+$"}}

### spec.connectionProfileId

`string`

The profile's ID. Defaults to metadata.name. Immutable.

### spec.displayName

`string`

The name shown in the console. Defaults to metadata.name.

### spec.labels

`map<string, string>`

Labels on the profile. The platform attribution labels are added on
top and win on a key conflict.

### spec.createWithoutValidation

`bool`

Create the profile without Google's connectivity test. Useful when the
database or its network is not reachable yet; a wrong host or
password then surfaces when a stream starts. Immutable.

### spec.bigqueryProfile

`bool`

A BigQuery destination. Google's block has no settings -- the stream
decides datasets -- so true declares it. Datastream's service agent
writes with BigQuery Data Editor on the target datasets.

### spec.gcsProfile

`GcpDatastreamConnectionProfileGcsProfile`

A Cloud Storage destination.

### spec.gcsProfile.bucket

`string | valueFrom` · required

The bucket name -- a GcpGcsBucket reference (its bucket_name output) or
a literal. Datastream's service agent needs object write access on it
(grant roles/storage.objectAdmin through the bucket's iam_members).

- references: GcpGcsBucket (`status.outputs.bucket_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpGcsBucket, name: <that resource's name>, fieldPath: status.outputs.bucket_name}} -- a bare string does not parse

### spec.gcsProfile.rootPath

`string`

The folder inside the bucket every stream using this profile writes
under, e.g. "/datastream". A stream's own path nests below it.

### spec.mysqlProfile

`GcpDatastreamConnectionProfileMysqlProfile`

A MySQL source.

- rule: password and secret_manager_stored_password are mutually exclusive

### spec.mysqlProfile.hostname

`string | valueFrom` · required

The server's address. Defaults to a GcpCloudSql reference's public_ip
-- Google's direct path to Cloud SQL, with Datastream's regional IPs
allowlisted in the instance's authorized networks. For a private
instance, point at a proxy VM in your VPC reached through a private
connection (a reference with another fieldPath, or a literal).

- references: GcpCloudSql (`status.outputs.public_ip`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpCloudSql, name: <that resource's name>, fieldPath: status.outputs.public_ip}} -- a bare string does not parse

### spec.mysqlProfile.port

`int32`

The port. Empty uses 3306.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"lte":65535,"gte":1}}

### spec.mysqlProfile.username

`string` · required

The replication user.

- rule: {"required":true}

### spec.mysqlProfile.password

`string` · sensitive

The user's password, stored by Google and never returned. Mutually
exclusive with secret_manager_stored_password.

### spec.mysqlProfile.secretManagerStoredPassword

`string | valueFrom`

The Secret Manager secret VERSION holding the password, as
projects/{project}/secrets/{secret}/versions/{version} -- a
GcpSecretManagerSecret reference (its latest_version_name output, set
when the secret declares an initial version) or a literal version
name. Datastream's service agent reads it
(roles/secretmanager.secretAccessor on the secret). Mutually exclusive
with password.

- references: GcpSecretManagerSecret (`status.outputs.latest_version_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpSecretManagerSecret, name: <that resource's name>, fieldPath: status.outputs.latest_version_name}} -- a bare string does not parse

### spec.mysqlProfile.sslConfig

`GcpDatastreamConnectionProfileMysqlSslConfig`

TLS for the session.

- rule: client_certificate and client_key go together, and both require ca_certificate

### spec.mysqlProfile.sslConfig.caCertificate

`string` · sensitive

PEM certificate of the CA that signed the server's certificate.

### spec.mysqlProfile.sslConfig.clientCertificate

`string` · sensitive

PEM certificate Datastream presents to the server. Google requires
client_key and ca_certificate with it.

### spec.mysqlProfile.sslConfig.clientKey

`string` · sensitive

PEM private key of client_certificate. Google requires
client_certificate and ca_certificate with it.

### spec.postgresqlProfile

`GcpDatastreamConnectionProfilePostgresqlProfile`

A PostgreSQL source.

- rule: password and secret_manager_stored_password are mutually exclusive

### spec.postgresqlProfile.hostname

`string | valueFrom` · required

The server's address. Defaults to a GcpCloudSql reference's public_ip
-- Google's direct path to Cloud SQL, with Datastream's regional IPs
allowlisted. AlloyDB: reference a GcpAlloydbInstance with fieldPath
status.outputs.ip_address, reached through a private connection and a
proxy VM. A literal takes any hostname or IP.

- references: GcpCloudSql (`status.outputs.public_ip`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpCloudSql, name: <that resource's name>, fieldPath: status.outputs.public_ip}} -- a bare string does not parse

### spec.postgresqlProfile.port

`int32`

The port. Empty uses 5432.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"lte":65535,"gte":1}}

### spec.postgresqlProfile.username

`string` · required

The replication user (it needs the REPLICATION attribute, or
cloudsqlsuperuser's replication role on Cloud SQL).

- rule: {"required":true}

### spec.postgresqlProfile.password

`string` · sensitive

The user's password, stored by Google and never returned. Mutually
exclusive with secret_manager_stored_password.

### spec.postgresqlProfile.secretManagerStoredPassword

`string | valueFrom`

The Secret Manager secret VERSION holding the password -- a
GcpSecretManagerSecret reference (its latest_version_name output) or a
literal projects/{project}/secrets/{secret}/versions/{version}.
Datastream's service agent reads it. Mutually exclusive with password.

- references: GcpSecretManagerSecret (`status.outputs.latest_version_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpSecretManagerSecret, name: <that resource's name>, fieldPath: status.outputs.latest_version_name}} -- a bare string does not parse

### spec.postgresqlProfile.database

`string` · required

The database to replicate from.

- rule: {"required":true}

### spec.postgresqlProfile.sslConfig

`GcpDatastreamConnectionProfilePostgresqlSslConfig`

TLS for the session.

- rule: set at most one of server_verification or server_and_client_verification

### spec.postgresqlProfile.sslConfig.serverVerification

`GcpDatastreamConnectionProfilePostgresqlServerVerification`

The server proves its identity.

### spec.postgresqlProfile.sslConfig.serverVerification.caCertificate

`string` · required · sensitive

PEM root CA certificate of the server.

- rule: {"required":true}

### spec.postgresqlProfile.sslConfig.serverAndClientVerification

`GcpDatastreamConnectionProfilePostgresqlServerAndClientVerification`

Both the server and Datastream prove their identity.

### spec.postgresqlProfile.sslConfig.serverAndClientVerification.caCertificate

`string` · required · sensitive

PEM root CA certificate of the server.

- rule: {"required":true}

### spec.postgresqlProfile.sslConfig.serverAndClientVerification.clientCertificate

`string` · required · sensitive

PEM certificate Datastream presents, signed by a CA the server trusts.

- rule: {"required":true}

### spec.postgresqlProfile.sslConfig.serverAndClientVerification.clientKey

`string` · required · sensitive

PEM private key of client_certificate.

- rule: {"required":true}

### spec.oracleProfile

`GcpDatastreamConnectionProfileOracleProfile`

An Oracle source.

- rule: password and secret_manager_stored_password are mutually exclusive

### spec.oracleProfile.hostname

`string` · required

The server's hostname or IP.

- rule: {"required":true}

### spec.oracleProfile.port

`int32`

The port. Empty uses 1521.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"lte":65535,"gte":1}}

### spec.oracleProfile.username

`string` · required

The database user.

- rule: {"required":true}

### spec.oracleProfile.password

`string` · sensitive

The user's password, stored by Google and never returned. Mutually
exclusive with secret_manager_stored_password.

### spec.oracleProfile.secretManagerStoredPassword

`string | valueFrom`

The Secret Manager secret VERSION holding the password -- a
GcpSecretManagerSecret reference (its latest_version_name output) or a
literal version name. Mutually exclusive with password.

- references: GcpSecretManagerSecret (`status.outputs.latest_version_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpSecretManagerSecret, name: <that resource's name>, fieldPath: status.outputs.latest_version_name}} -- a bare string does not parse

### spec.oracleProfile.databaseService

`string` · required

The database service (the Oracle service name) to connect to, e.g.
"ORCL".

- rule: {"required":true}

### spec.oracleProfile.connectionAttributes

`map<string, string>`

Extra Oracle connection-string attributes, as key/value pairs.

### spec.sqlServerProfile

`GcpDatastreamConnectionProfileSqlServerProfile`

A SQL Server source.

- rule: password and secret_manager_stored_password are mutually exclusive

### spec.sqlServerProfile.hostname

`string | valueFrom` · required

The server's address. Defaults to a GcpCloudSql reference's public_ip,
with Datastream's regional IPs allowlisted; a literal takes any
hostname or IP.

- references: GcpCloudSql (`status.outputs.public_ip`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpCloudSql, name: <that resource's name>, fieldPath: status.outputs.public_ip}} -- a bare string does not parse

### spec.sqlServerProfile.port

`int32`

The port. Empty uses 1433.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"lte":65535,"gte":1}}

### spec.sqlServerProfile.username

`string` · required

The database user.

- rule: {"required":true}

### spec.sqlServerProfile.password

`string` · sensitive

The user's password, stored by Google and never returned. Mutually
exclusive with secret_manager_stored_password.

### spec.sqlServerProfile.secretManagerStoredPassword

`string | valueFrom`

The Secret Manager secret VERSION holding the password -- a
GcpSecretManagerSecret reference (its latest_version_name output) or a
literal version name. Mutually exclusive with password.

- references: GcpSecretManagerSecret (`status.outputs.latest_version_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpSecretManagerSecret, name: <that resource's name>, fieldPath: status.outputs.latest_version_name}} -- a bare string does not parse

### spec.sqlServerProfile.database

`string` · required

The database to replicate from.

- rule: {"required":true}

### spec.mongodbProfile

`GcpDatastreamConnectionProfileMongodbProfile`

A MongoDB source.

- rule: set exactly one of srv_connection_format or standard_connection_format
- rule: replica_set must be empty with the SRV connection format
- rule: password and secret_manager_stored_password are mutually exclusive

### spec.mongodbProfile.hostAddresses

`[]GcpDatastreamConnectionProfileMongodbHostAddress` · required

The hosts to connect to -- every member of a replica set for the
standard format, the one SRV domain for the SRV format.

- rule: {"repeated":{"minItems":"1"}}

### spec.mongodbProfile.hostAddresses[].hostname

`string` · required

The host's name or IP. For the SRV format, the SRV record's domain.

- rule: {"required":true}

### spec.mongodbProfile.hostAddresses[].port

`int32`

The host's port. Leave empty for the SRV format (the record carries
it).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"lte":65535,"gte":1}}

### spec.mongodbProfile.username

`string` · required

The database user.

- rule: {"required":true}

### spec.mongodbProfile.password

`string` · sensitive

The user's password, stored by Google and never returned. Mutually
exclusive with secret_manager_stored_password.

### spec.mongodbProfile.secretManagerStoredPassword

`string | valueFrom`

The Secret Manager secret VERSION holding the password -- a
GcpSecretManagerSecret reference (its latest_version_name output) or a
literal version name. Mutually exclusive with password.

- references: GcpSecretManagerSecret (`status.outputs.latest_version_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpSecretManagerSecret, name: <that resource's name>, fieldPath: status.outputs.latest_version_name}} -- a bare string does not parse

### spec.mongodbProfile.replicaSet

`string`

The replica set's name -- needed for a self-hosted replica set with the
standard format; must be empty with the SRV format.

### spec.mongodbProfile.srvConnectionFormat

`bool`

Connect through a DNS SRV seed list (mongodb+srv://). Google's SRV
block has no settings: true declares it.

### spec.mongodbProfile.standardConnectionFormat

`GcpDatastreamConnectionProfileMongodbStandardConnectionFormat`

Connect through a standard mongodb:// URI to the listed hosts.

### spec.mongodbProfile.standardConnectionFormat.directConnection

`bool`

Connect straight to the one listed host instead of discovering the
replica set. Google now recommends additional_options
{"directConnection": "true"} instead.

### spec.mongodbProfile.sslConfig

`GcpDatastreamConnectionProfileMongodbSslConfig`

TLS for the session.

- rule: client_key and secret_manager_stored_client_key are mutually exclusive
- rule: a client identity needs client_certificate, ca_certificate, and one of client_key or secret_manager_stored_client_key

### spec.mongodbProfile.sslConfig.caCertificate

`string` · sensitive

PEM certificate of the CA that signed the server's certificate.

### spec.mongodbProfile.sslConfig.clientCertificate

`string` · sensitive

PEM certificate Datastream presents to the server.

### spec.mongodbProfile.sslConfig.clientKey

`string` · sensitive

PEM private key of client_certificate. Mutually exclusive with
secret_manager_stored_client_key.

### spec.mongodbProfile.sslConfig.secretManagerStoredClientKey

`string | valueFrom`

The Secret Manager secret VERSION holding the client key -- a
GcpSecretManagerSecret reference (its latest_version_name output) or a
literal version name. Mutually exclusive with client_key.

- references: GcpSecretManagerSecret (`status.outputs.latest_version_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpSecretManagerSecret, name: <that resource's name>, fieldPath: status.outputs.latest_version_name}} -- a bare string does not parse

### spec.mongodbProfile.additionalOptions

`map<string, string>`

Extra MongoDB connection-string options, keyed exactly as MongoDB
names them, e.g. {"serverSelectionTimeoutMS": "10000"}.

### spec.privateConnection

`string | valueFrom`

Reach the source through a private connection -- a
GcpDatastreamPrivateConnection reference (its name output) or a literal
projects/{project}/locations/{location}/privateConnections/{id} in the
profile's location. Mutually exclusive with forward_ssh_connectivity.

- references: GcpDatastreamPrivateConnection (`status.outputs.name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpDatastreamPrivateConnection, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.forwardSshConnectivity

`GcpDatastreamConnectionProfileForwardSshConnectivity`

Reach the source through an SSH bastion. Mutually exclusive with
private_connection.

- rule: password and private_key are mutually exclusive

### spec.forwardSshConnectivity.hostname

`string` · required

The bastion's hostname or IP.

- rule: {"required":true}

### spec.forwardSshConnectivity.port

`int32`

The bastion's SSH port. Empty uses 22.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"lte":65535,"gte":1}}

### spec.forwardSshConnectivity.username

`string` · required

The SSH user.

- rule: {"required":true}

### spec.forwardSshConnectivity.password

`string` · sensitive

The SSH password. Mutually exclusive with private_key. Immutable.

### spec.forwardSshConnectivity.privateKey

`string` · sensitive

The SSH private key, PEM. Mutually exclusive with password; the
stronger choice.

### spec.deletionPolicy

`string`

What happens to the profile when this resource is destroyed:
  "" / "DELETE" -- deleted (Google refuses while a stream uses it)
  "PREVENT"     -- destroy fails
  "ABANDON"     -- it leaves management and stays in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `spec.exactly_one_profile`: set exactly one of bigquery_profile, gcs_profile, mysql_profile, postgresql_profile, oracle_profile, sql_server_profile, mongodb_profile
- `spec.one_connectivity`: private_connection and forward_ssh_connectivity are mutually exclusive

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpDatastreamConnectionProfile, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | Full resource name -- projects/{project}/locations/{location}/connectionProfiles/{connection_profile_id}. The value a GcpDatastreamStream's source_connection_profile or destination_connection_profile takes. |
| `status.outputs.connection_profile_id` | `string` | The profile's ID. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.gcsProfile.bucket` | GcpGcsBucket | `status.outputs.bucket_name` |
| `spec.mysqlProfile.hostname` | GcpCloudSql | `status.outputs.public_ip` |
| `spec.mysqlProfile.secretManagerStoredPassword` | GcpSecretManagerSecret | `status.outputs.latest_version_name` |
| `spec.postgresqlProfile.hostname` | GcpCloudSql | `status.outputs.public_ip` |
| `spec.postgresqlProfile.secretManagerStoredPassword` | GcpSecretManagerSecret | `status.outputs.latest_version_name` |
| `spec.oracleProfile.secretManagerStoredPassword` | GcpSecretManagerSecret | `status.outputs.latest_version_name` |
| `spec.sqlServerProfile.hostname` | GcpCloudSql | `status.outputs.public_ip` |
| `spec.sqlServerProfile.secretManagerStoredPassword` | GcpSecretManagerSecret | `status.outputs.latest_version_name` |
| `spec.mongodbProfile.secretManagerStoredPassword` | GcpSecretManagerSecret | `status.outputs.latest_version_name` |
| `spec.mongodbProfile.sslConfig.secretManagerStoredClientKey` | GcpSecretManagerSecret | `status.outputs.latest_version_name` |
| `spec.privateConnection` | GcpDatastreamPrivateConnection | `status.outputs.name` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpDatastreamStream | `spec.sourceConfig.sourceConnectionProfile` | `status.outputs.name` |
| GcpDatastreamStream | `spec.destinationConfig.destinationConnectionProfile` | `status.outputs.name` |

## See Also

- [Overview](../README.md)
