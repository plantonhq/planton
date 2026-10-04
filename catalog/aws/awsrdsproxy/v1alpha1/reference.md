# AwsRdsProxy

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsRdsProxySpec defines one RDS Proxy - the managed connection pool
that sits between connection-hungry applications (Lambda above all)
and a database - with its connection-pool tuning, additional
endpoints, and database target managed in-line.

The proxy authenticates clients with IAM or native credentials and
signs in to the database with credentials read from Secrets Manager
(the auth entries below) under the IAM role you provide. The proxy
name is wired from metadata.name.

Fixed for life at the provider: engine_family, the proxy's subnets,
and both network-type dials - changing any of them replaces the
proxy (its endpoint DNS names change with it).

## Example

```yaml
# Canonical AwsRdsProxy example (hack/dev manifest and refgen Example
# source): a PostgreSQL proxy with IAM-authenticated clients, pool
# tuning, a read-only endpoint, and an RDS instance target.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsRdsProxy
metadata:
  name: orders-db-proxy
  id: orders-db-proxy
  org: test-org
  env: dev
spec:
  region: us-west-2
  engineFamily: POSTGRESQL
  roleArn:
    value: arn:aws:iam::123456789012:role/orders-proxy-secrets
  vpcSubnetIds:
    - value: subnet-0123456789abcdef0
    - value: subnet-0123456789abcdef1
  vpcSecurityGroupIds:
    - value: sg-0123456789abcdef0
  auth:
    - secretArn:
        value: arn:aws:secretsmanager:us-west-2:123456789012:secret:orders-db-creds-AbCdEf
      description: Application read-write credentials
      iamAuth: REQUIRED
      clientPasswordAuthType: POSTGRES_SCRAM_SHA_256
  requireTls: true
  idleClientTimeout: 900
  connectionPool:
    maxConnectionsPercent: 90
    maxIdleConnectionsPercent: 10
    connectionBorrowTimeout: 120
  endpoints:
    - name: readers
      targetRole: READ_ONLY
      vpcSubnetIds:
        - value: subnet-0123456789abcdef0
        - value: subnet-0123456789abcdef1
  target:
    dbInstanceIdentifier:
      value: orders-db
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.engineFamily` | `string` |  |  |  |
| `spec.roleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.vpcSubnetIds` | `[]string \| valueFrom` | yes |  | AwsSubnet (`status.outputs.subnet_id`) |
| `spec.vpcSecurityGroupIds` | `[]string \| valueFrom` |  |  | AwsSecurityGroup (`status.outputs.security_group_id`) |
| `spec.auth` | `[]AwsRdsProxyAuth` | yes |  |  |
| `spec.auth[].secretArn` | `string \| valueFrom` | yes |  | AwsSecretsManagerSecret (`status.outputs.secret_arn`) |
| `spec.auth[].description` | `string` |  |  |  |
| `spec.auth[].iamAuth` | `string` |  |  |  |
| `spec.auth[].clientPasswordAuthType` | `string` |  |  |  |
| `spec.auth[].username` | `string` |  |  |  |
| `spec.requireTls` | `bool` |  |  |  |
| `spec.idleClientTimeout` | `int64` |  |  |  |
| `spec.debugLogging` | `bool` |  |  |  |
| `spec.defaultAuthScheme` | `string` |  |  |  |
| `spec.endpointNetworkType` | `string` |  |  |  |
| `spec.targetConnectionNetworkType` | `string` |  |  |  |
| `spec.connectionPool` | `AwsRdsProxyConnectionPool` |  |  |  |
| `spec.connectionPool.connectionBorrowTimeout` | `int64` |  | `120` |  |
| `spec.connectionPool.initQuery` | `string` |  |  |  |
| `spec.connectionPool.maxConnectionsPercent` | `int64` |  | `100` |  |
| `spec.connectionPool.maxIdleConnectionsPercent` | `int64` |  | `50` |  |
| `spec.connectionPool.sessionPinningFilters` | `[]string` |  |  |  |
| `spec.endpoints` | `[]AwsRdsProxyEndpoint` |  |  |  |
| `spec.endpoints[].name` | `string` | yes |  |  |
| `spec.endpoints[].targetRole` | `string` |  |  |  |
| `spec.endpoints[].vpcSubnetIds` | `[]string \| valueFrom` | yes |  | AwsSubnet (`status.outputs.subnet_id`) |
| `spec.endpoints[].vpcSecurityGroupIds` | `[]string \| valueFrom` |  |  | AwsSecurityGroup (`status.outputs.security_group_id`) |
| `spec.target` | `AwsRdsProxyTarget` |  |  |  |
| `spec.target.dbInstanceIdentifier` | `string \| valueFrom` |  |  | AwsRdsInstance (`status.outputs.instance_identifier`) |
| `spec.target.dbClusterIdentifier` | `string \| valueFrom` |  |  | AwsRdsCluster (`status.outputs.cluster_identifier`) |

## Field Details

### spec.region

`string` · required

The AWS region the proxy lives in. Example: "us-west-2".

- rule: {"string":{"minLen":"1"}}

### spec.engineFamily

`string`

The database engine family the proxy speaks. Must match the
target database's engine. Fixed for life.

- rule: {"string":{"in":["MYSQL","POSTGRESQL","SQLSERVER"]}}

### spec.roleArn

`string | valueFrom` · required

The IAM role the proxy assumes to read database credentials from
Secrets Manager. Its policy needs secretsmanager:GetSecretValue
on every secret in auth (plus kms:Decrypt when the secrets use a
customer-managed key), and its trust policy must allow
rds.amazonaws.com to assume it. Reference an AwsIamRole role_arn
output or pass a literal ARN.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.vpcSubnetIds

`[]string | valueFrom` · required

The subnets the proxy places its network interfaces in. AWS
requires at least two, in different availability zones. Fixed for
life. Reference AwsSubnet subnet_id outputs or pass literal
subnet-... ids.

- references: AwsSubnet (`status.outputs.subnet_id`)
- rule: {"repeated":{"minItems":"2"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.vpcSecurityGroupIds

`[]string | valueFrom`

The security groups on the proxy's network interfaces - they must
allow the applications in and the database out. Empty uses the
VPC's default security group. Reference AwsSecurityGroup
security_group_id outputs or pass literal sg-... ids.

- references: AwsSecurityGroup (`status.outputs.security_group_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecurityGroup, name: <that resource's name>, fieldPath: status.outputs.security_group_id}} -- a bare string does not parse

### spec.auth

`[]AwsRdsProxyAuth` · required

How the proxy signs in to the database: one entry per Secrets
Manager secret holding database credentials (multiple entries
let different applications use different database users through
one proxy).

- rule: {"repeated":{"minItems":"1"}}

### spec.auth[].secretArn

`string | valueFrom` · required

The Secrets Manager secret holding the database credentials
(the standard RDS secret shape: username + password). Reference
an AwsSecretsManagerSecret secret_arn output or pass a literal
ARN.

- references: AwsSecretsManagerSecret (`status.outputs.secret_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecretsManagerSecret, name: <that resource's name>, fieldPath: status.outputs.secret_arn}} -- a bare string does not parse

### spec.auth[].description

`string`

What this sign-in is for.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"1024"}}

### spec.auth[].iamAuth

`string`

Whether clients using this sign-in authenticate to the PROXY with
IAM: "DISABLED" (native database credentials - the default),
"REQUIRED" (IAM tokens only), or "ENABLED" (either). IAM auth
requires require_tls.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["DISABLED","REQUIRED","ENABLED"]}}

### spec.auth[].clientPasswordAuthType

`string`

The password hashing scheme the database uses for this user.
Unset leaves AWS's engine-family default
(MYSQL_NATIVE_PASSWORD / POSTGRES_SCRAM_SHA_256 /
SQL_SERVER_AUTHENTICATION).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["MYSQL_NATIVE_PASSWORD","MYSQL_CACHING_SHA2_PASSWORD","POSTGRES_SCRAM_SHA_256","POSTGRES_MD5","SQL_SERVER_AUTHENTICATION"]}}

### spec.auth[].username

`string`

Pin this sign-in to one database username. Empty uses the
username stored in the secret.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"128"}}

### spec.requireTls

`bool`

Require TLS on client connections to the proxy. Strongly
recommended with IAM authentication (IAM auth REQUIRES TLS at
connect time).

### spec.idleClientTimeout

`int64`

Seconds a client connection may sit idle before the proxy closes
it (1-28800). Unset leaves AWS's default (1800 = 30 minutes).

- rule: idle_client_timeout must be between 1 and 28800 seconds (AWS defaults to 1800 when unset)

### spec.debugLogging

`bool`

Log the SQL statements the proxy forwards, for debugging. The
logs can contain sensitive query text - leave off outside active
troubleshooting.

### spec.defaultAuthScheme

`string`

The default authentication scheme for the proxy's endpoints:
"IAM_AUTH" (clients present IAM tokens) or "NONE" (native
database credentials). Unset leaves AWS's default. Per-secret
iam_auth entries refine this.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["IAM_AUTH","NONE"]}}

### spec.endpointNetworkType

`string`

The IP protocol of the proxy's DEFAULT endpoint: "IPV4" (the
default), "IPV6", or "DUAL". Fixed for life.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["IPV4","IPV6","DUAL"]}}

### spec.targetConnectionNetworkType

`string`

The IP protocol between the proxy and the database: "IPV4" (the
default) or "IPV6". Fixed for life.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["IPV4","IPV6"]}}

### spec.connectionPool

`AwsRdsProxyConnectionPool`

Connection-pool tuning for the default target group. Omit for
AWS's defaults. (The target group is part of the proxy - it has
no delete of its own; destroying the proxy takes it along.)

### spec.connectionPool.connectionBorrowTimeout

`int64` · optional (explicit presence)

Seconds a client waits for a database connection from the pool
before timing out (0-3600; 0 means fail immediately). Unset
leaves AWS's default (120).

- default: `120`
- rule: connection_borrow_timeout must be between 0 and 3600 seconds

### spec.connectionPool.initQuery

`string`

SQL to run on every new database connection before handing it to
a client - session presets like "SET x=1, y=2" (one statement;
engines differ on multi-statement support).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"4096"}}

### spec.connectionPool.maxConnectionsPercent

`int64` · optional (explicit presence)

Ceiling on database connections, as a percent of the database's
max_connections (1-100). Unset leaves AWS's default (100).

- default: `100`
- rule: max_connections_percent must be between 1 and 100

### spec.connectionPool.maxIdleConnectionsPercent

`int64` · optional (explicit presence)

Ceiling on IDLE database connections the pool keeps warm, as a
percent of max_connections (0-100; must not exceed
max_connections_percent). Unset leaves AWS's default (50).

- default: `50`
- rule: max_idle_connections_percent must be between 0 and 100

### spec.connectionPool.sessionPinningFilters

`[]string`

Session states that EXCLUDE a connection from reuse by other
clients (session pinning). The only value AWS supports today is
"EXCLUDE_VARIABLE_SETS" - session variable changes no longer pin
the connection.

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EXCLUDE_VARIABLE_SETS"]}}}}

### spec.endpoints

`[]AwsRdsProxyEndpoint`

Additional proxy endpoints, keyed by name - read-only endpoints
for Aurora reader farms, or endpoints in other network
configurations.

### spec.endpoints[].name

`string` · required

The endpoint's name - the for_each key on both engines, the key
in the endpoint output maps, and part of the endpoint's DNS name.
Lowercase letters, digits, and single hyphens; must start with a
letter and must not end with a hyphen.

- rule: endpoint name must not contain consecutive hyphens or end with a hyphen
- rule: {"string":{"minLen":"1","maxLen":"63","pattern":"^[a-z][0-9a-z-]*$"}}

### spec.endpoints[].targetRole

`string`

Whether the endpoint fronts writers or readers: "READ_WRITE"
(the default) or "READ_ONLY" (Aurora reader farms - read-only
endpoints distribute across up to the cluster's readers). Fixed
for life.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["READ_WRITE","READ_ONLY"]}}

### spec.endpoints[].vpcSubnetIds

`[]string | valueFrom` · required

The subnets this endpoint places its network interfaces in (at
least two, different availability zones). Fixed for life.

- references: AwsSubnet (`status.outputs.subnet_id`)
- rule: {"repeated":{"minItems":"2"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.endpoints[].vpcSecurityGroupIds

`[]string | valueFrom`

The security groups on this endpoint's network interfaces. Empty
inherits the VPC default.

- references: AwsSecurityGroup (`status.outputs.security_group_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecurityGroup, name: <that resource's name>, fieldPath: status.outputs.security_group_id}} -- a bare string does not parse

### spec.target

`AwsRdsProxyTarget`

The database the proxy fronts. Omit to create the proxy first and
register the target later (the proxy is usable only once a
target is registered).

- rule: set exactly one of db_instance_identifier and db_cluster_identifier - a proxy fronts one RDS instance or one Aurora cluster

### spec.target.dbInstanceIdentifier

`string | valueFrom`

The RDS instance the proxy fronts. Reference an AwsRdsInstance
instance_identifier output or pass a literal identifier.

- references: AwsRdsInstance (`status.outputs.instance_identifier`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsRdsInstance, name: <that resource's name>, fieldPath: status.outputs.instance_identifier}} -- a bare string does not parse

### spec.target.dbClusterIdentifier

`string | valueFrom`

The Aurora cluster the proxy fronts (the proxy tracks the
cluster's writer/reader topology). Reference an AwsRdsCluster
cluster_identifier output or pass a literal identifier.

- references: AwsRdsCluster (`status.outputs.cluster_identifier`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsRdsCluster, name: <that resource's name>, fieldPath: status.outputs.cluster_identifier}} -- a bare string does not parse

## Validation Rules

- `spec.endpoint_names_unique`: endpoint names must be unique within the proxy

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsRdsProxy, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.proxy_name` | `string` | The proxy's name (wired from metadata.name) - the provider's import ID and the join key for AWS CLI/API lookups. |
| `status.outputs.proxy_arn` | `string` | The proxy's ARN. |
| `status.outputs.endpoint` | `string` | The proxy's DEFAULT endpoint DNS name - what applications connect to. The chart-ready join key for application database hosts. |
| `status.outputs.default_target_group_arn` | `string` | The default target group's ARN. |
| `status.outputs.default_target_group_name` | `string` | The default target group's name (always "default" at AWS). |
| `status.outputs.endpoint_addresses` | `map<string, string>` | Additional endpoints' DNS names keyed by endpoint name. |
| `status.outputs.endpoint_arns` | `map<string, string>` | Additional endpoints' ARNs keyed by endpoint name. |
| `status.outputs.target_type` | `string` | The registered target's type as AWS reports it (RDS_INSTANCE or TRACKED_CLUSTER) - part of the target's import ID. |
| `status.outputs.target_rds_resource_id` | `string` | The registered target's RDS resource id - part of the target's import ID. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.roleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.vpcSubnetIds` | AwsSubnet | `status.outputs.subnet_id` |
| `spec.vpcSecurityGroupIds` | AwsSecurityGroup | `status.outputs.security_group_id` |
| `spec.auth[].secretArn` | AwsSecretsManagerSecret | `status.outputs.secret_arn` |
| `spec.endpoints[].vpcSubnetIds` | AwsSubnet | `status.outputs.subnet_id` |
| `spec.endpoints[].vpcSecurityGroupIds` | AwsSecurityGroup | `status.outputs.security_group_id` |
| `spec.target.dbInstanceIdentifier` | AwsRdsInstance | `status.outputs.instance_identifier` |
| `spec.target.dbClusterIdentifier` | AwsRdsCluster | `status.outputs.cluster_identifier` |

## See Also

- [Overview](../README.md)
