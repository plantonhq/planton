# AwsOpenSearchDomain

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsOpenSearchDomainSpec defines the desired configuration for an Amazon
OpenSearch Service domain (successor to Amazon Elasticsearch Service).

OpenSearch provides managed search, analytics, and observability capabilities.
A domain consists of data nodes (required), optional dedicated master nodes for
cluster management, optional coordinator node pools for request fan-out, optional
UltraWarm nodes for infrequently accessed data, and optional cold storage backed by S3.

Deployment models:
- **Public** (no vpc_options) — accessible over the internet, secured via access
  policies and optionally fine-grained access control (FGAC).
- **VPC** (vpc_options set) — deployed into VPC subnets, secured via security
  groups and access policies. VPC configuration is ForceNew — changing it
  destroys and recreates the domain.

Notes:
- `encrypt_at_rest` KMS key is ForceNew — choose the encryption key carefully upfront.
- `advanced_security_options.enabled` cannot be disabled once turned on (ForceNew).
- Domain name must match `^[a-z][0-9a-z\-]{2,27}$` (3-28 chars, lowercase, hyphens).
- Engine version format: "OpenSearch_X.Y" or "Elasticsearch_X.Y".
- Credentials, region, and deployment workflow live outside this spec in IaC inputs.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsOpenSearchDomain
metadata:
  name: awsos-demo
spec:
  region: us-west-2
  engineVersion: "OpenSearch_2.17"
  clusterConfig:
    instanceType: r6g.large.search
    instanceCount: 3
    dedicatedMasterEnabled: true
    dedicatedMasterType: r6g.large.search
    dedicatedMasterCount: 3
    zoneAwarenessEnabled: true
    availabilityZoneCount: 3
  ebsOptions:
    ebsEnabled: true
    volumeType: gp3
    volumeSize: 100
  encryptAtRestEnabled: true
  nodeToNodeEncryptionEnabled: true
  vpcOptions:
    subnetIds:
      - value: subnet-0aaa1111bbb222333
      - value: subnet-0bbb2222ccc333444
      - value: subnet-0ccc3333ddd444555
    securityGroupIds:
      - value: sg-0abc1234def567890
  domainEndpointOptions:
    enforceHttps: true
    tlsSecurityPolicy: "Policy-Min-TLS-1-2-PFS-2023-10"
  advancedSecurityOptions:
    enabled: true
    internalUserDatabaseEnabled: true
    masterUserName: admin
    # An obvious throwaway that still satisfies FGAC's complexity rule
    # (upper/lower/digit/special) -- never a real-looking credential.
    masterUserPassword:
      value: "E2e-Replace-Me-0!"
  samlOptions:
    idpEntityId: https://idp.example.com/saml/metadata
    idpMetadataContent: |
      <EntityDescriptor xmlns="urn:oasis:names:tc:SAML:2.0:metadata" entityID="https://idp.example.com/saml/metadata"></EntityDescriptor>
    rolesKey: roles
    sessionTimeoutMinutes: 120
  authorizedVpcEndpointAccessAccounts:
    - "123456789012"
  autoTuneOptions:
    desiredState: ENABLED
    useOffPeakWindow: true
  offPeakWindowOptions:
    enabled: true
    windowStartHour: 23
    windowStartMinute: 30
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.engineVersion` | `string` | yes |  |  |
| `spec.clusterConfig` | `AwsOpenSearchDomainClusterConfig` | yes |  |  |
| `spec.clusterConfig.instanceType` | `string` | yes |  |  |
| `spec.clusterConfig.instanceCount` | `int32` |  | `1` |  |
| `spec.clusterConfig.dedicatedMasterEnabled` | `bool` |  |  |  |
| `spec.clusterConfig.dedicatedMasterType` | `string` |  |  |  |
| `spec.clusterConfig.dedicatedMasterCount` | `int32` |  |  |  |
| `spec.clusterConfig.nodeOptions` | `[]AwsOpenSearchDomainNodeOption` |  |  |  |
| `spec.clusterConfig.nodeOptions[].nodeType` | `string` | yes |  |  |
| `spec.clusterConfig.nodeOptions[].enabled` | `bool` |  |  |  |
| `spec.clusterConfig.nodeOptions[].instanceType` | `string` |  |  |  |
| `spec.clusterConfig.nodeOptions[].count` | `int32` |  |  |  |
| `spec.clusterConfig.zoneAwarenessEnabled` | `bool` |  |  |  |
| `spec.clusterConfig.availabilityZoneCount` | `int32` |  |  |  |
| `spec.clusterConfig.warmEnabled` | `bool` |  |  |  |
| `spec.clusterConfig.warmType` | `string` |  |  |  |
| `spec.clusterConfig.warmCount` | `int32` |  |  |  |
| `spec.clusterConfig.coldStorageEnabled` | `bool` |  |  |  |
| `spec.clusterConfig.multiAzWithStandbyEnabled` | `bool` |  |  |  |
| `spec.ebsOptions` | `AwsOpenSearchDomainEbsOptions` | yes |  |  |
| `spec.ebsOptions.ebsEnabled` | `bool` |  |  |  |
| `spec.ebsOptions.volumeType` | `string` |  |  |  |
| `spec.ebsOptions.volumeSize` | `int32` |  |  |  |
| `spec.ebsOptions.iops` | `int32` |  |  |  |
| `spec.ebsOptions.throughput` | `int32` |  |  |  |
| `spec.encryptAtRestEnabled` | `bool` |  | `true` |  |
| `spec.kmsKeyId` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.nodeToNodeEncryptionEnabled` | `bool` |  | `true` |  |
| `spec.vpcOptions` | `AwsOpenSearchDomainVpcOptions` |  |  |  |
| `spec.vpcOptions.subnetIds` | `[]string \| valueFrom` |  |  | AwsSubnet (`status.outputs.subnet_id`) |
| `spec.vpcOptions.securityGroupIds` | `[]string \| valueFrom` |  |  | AwsSecurityGroup (`status.outputs.security_group_id`) |
| `spec.domainEndpointOptions` | `AwsOpenSearchDomainEndpointOptions` |  |  |  |
| `spec.domainEndpointOptions.enforceHttps` | `bool` |  | `true` |  |
| `spec.domainEndpointOptions.tlsSecurityPolicy` | `string` |  |  |  |
| `spec.domainEndpointOptions.customEndpointEnabled` | `bool` |  |  |  |
| `spec.domainEndpointOptions.customEndpoint` | `string` |  |  |  |
| `spec.domainEndpointOptions.customEndpointCertificateArn` | `string \| valueFrom` |  |  | AwsCertManagerCert (`status.outputs.cert_arn`) |
| `spec.advancedSecurityOptions` | `AwsOpenSearchDomainAdvancedSecurityOptions` |  |  |  |
| `spec.advancedSecurityOptions.enabled` | `bool` |  |  |  |
| `spec.advancedSecurityOptions.internalUserDatabaseEnabled` | `bool` |  |  |  |
| `spec.advancedSecurityOptions.anonymousAuthEnabled` | `bool` |  |  |  |
| `spec.advancedSecurityOptions.masterUserArn` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.advancedSecurityOptions.masterUserName` | `string` |  |  |  |
| `spec.advancedSecurityOptions.masterUserPassword` | `string \| valueFrom` (sensitive) |  |  |  |
| `spec.advancedSecurityOptions.jwtOptions` | `AwsOpenSearchDomainJwtOptions` |  |  |  |
| `spec.advancedSecurityOptions.jwtOptions.enabled` | `bool` |  |  |  |
| `spec.advancedSecurityOptions.jwtOptions.jwksUrl` | `string` |  |  |  |
| `spec.advancedSecurityOptions.jwtOptions.publicKey` | `string` |  |  |  |
| `spec.advancedSecurityOptions.jwtOptions.rolesKey` | `string` |  |  |  |
| `spec.advancedSecurityOptions.jwtOptions.subjectKey` | `string` |  |  |  |
| `spec.cognitoOptions` | `AwsOpenSearchDomainCognitoOptions` |  |  |  |
| `spec.cognitoOptions.enabled` | `bool` |  |  |  |
| `spec.cognitoOptions.userPoolId` | `string \| valueFrom` |  |  | AwsCognitoUserPool (`status.outputs.user_pool_id`) |
| `spec.cognitoOptions.identityPoolId` | `string` |  |  |  |
| `spec.cognitoOptions.roleArn` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.logPublishingOptions` | `[]AwsOpenSearchDomainLogPublishingOption` |  |  |  |
| `spec.logPublishingOptions[].logType` | `string` | yes |  |  |
| `spec.logPublishingOptions[].cloudwatchLogGroupArn` | `string \| valueFrom` | yes |  | AwsCloudwatchLogGroup (`status.outputs.log_group_arn`) |
| `spec.logPublishingOptions[].enabled` | `bool` |  | `true` |  |
| `spec.accessPolicies` | `object` |  |  |  |
| `spec.autoTuneOptions` | `AwsOpenSearchDomainAutoTuneOptions` |  |  |  |
| `spec.autoTuneOptions.desiredState` | `string` | yes |  |  |
| `spec.autoTuneOptions.maintenanceSchedules` | `[]AwsOpenSearchDomainAutoTuneMaintenanceSchedule` |  |  |  |
| `spec.autoTuneOptions.maintenanceSchedules[].startAt` | `string` | yes |  |  |
| `spec.autoTuneOptions.maintenanceSchedules[].durationHours` | `int32` | yes |  |  |
| `spec.autoTuneOptions.maintenanceSchedules[].cronExpressionForRecurrence` | `string` | yes |  |  |
| `spec.autoTuneOptions.rollbackOnDisable` | `string` |  |  |  |
| `spec.autoTuneOptions.useOffPeakWindow` | `bool` |  |  |  |
| `spec.automatedSnapshotStartHour` | `int32` |  |  |  |
| `spec.offPeakWindowOptions` | `AwsOpenSearchDomainOffPeakWindowOptions` |  |  |  |
| `spec.offPeakWindowOptions.enabled` | `bool` |  | `true` |  |
| `spec.offPeakWindowOptions.windowStartHour` | `int32` |  |  |  |
| `spec.offPeakWindowOptions.windowStartMinute` | `int32` |  |  |  |
| `spec.autoSoftwareUpdateEnabled` | `bool` |  |  |  |
| `spec.deploymentStrategy` | `string` |  |  |  |
| `spec.ipAddressType` | `string` |  |  |  |
| `spec.advancedOptions` | `map<string, string>` |  |  |  |
| `spec.aimlOptions` | `AwsOpenSearchDomainAimlOptions` |  |  |  |
| `spec.aimlOptions.naturalLanguageQueryGenerationDesiredState` | `string` |  |  |  |
| `spec.aimlOptions.s3VectorsEngineEnabled` | `bool` |  |  |  |
| `spec.aimlOptions.serverlessVectorAccelerationEnabled` | `bool` |  |  |  |
| `spec.identityCenterOptions` | `AwsOpenSearchDomainIdentityCenterOptions` |  |  |  |
| `spec.identityCenterOptions.enabledApiAccess` | `bool` |  |  |  |
| `spec.identityCenterOptions.identityCenterInstanceArn` | `string` |  |  |  |
| `spec.identityCenterOptions.rolesKey` | `string` |  |  |  |
| `spec.identityCenterOptions.subjectKey` | `string` |  |  |  |
| `spec.samlOptions` | `AwsOpenSearchDomainSamlOptions` |  |  |  |
| `spec.samlOptions.idpEntityId` | `string` | yes |  |  |
| `spec.samlOptions.idpMetadataContent` | `string` | yes |  |  |
| `spec.samlOptions.masterBackendRole` | `string` |  |  |  |
| `spec.samlOptions.masterUserName` | `string` (sensitive) |  |  |  |
| `spec.samlOptions.rolesKey` | `string` |  |  |  |
| `spec.samlOptions.subjectKey` | `string` |  |  |  |
| `spec.samlOptions.sessionTimeoutMinutes` | `int32` |  |  |  |
| `spec.authorizedVpcEndpointAccessAccounts` | `[]string` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the resource will be created.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.engineVersion

`string` · required

OpenSearch or Elasticsearch engine version. Format: "OpenSearch_X.Y" (e.g.,
"OpenSearch_2.11") or "Elasticsearch_X.Y" (e.g., "Elasticsearch_7.10").
Upgrades are applied in place; a version DOWNGRADE (or a change to an
incompatible version) forces domain recreation.

- rule: {"required":true}

### spec.clusterConfig

`AwsOpenSearchDomainClusterConfig` · required

Cluster topology: data nodes, dedicated masters, coordinator node pools,
zone awareness, warm/cold storage.

- rule: {"required":true}
- rule: dedicated_master_type requires dedicated_master_enabled to be true
- rule: dedicated_master_count requires dedicated_master_enabled to be true
- rule: availability_zone_count requires zone_awareness_enabled to be true
- rule: availability_zone_count must be 2 or 3 when set
- rule: warm_type requires warm_enabled to be true
- rule: warm_count requires warm_enabled to be true
- rule: warm_count must be between 2 and 150 when set
- rule: warm_enabled requires warm_type and warm_count
- rule: cold_storage_enabled requires warm_enabled to be true (cold storage depends on UltraWarm)

### spec.clusterConfig.instanceType

`string` · required

Instance type for data nodes. Uses the `.search` suffix.
Examples: "t3.small.search" (dev), "r6g.large.search" (production),
"r6g.2xlarge.search" (high-memory workloads).

- rule: {"required":true}

### spec.clusterConfig.instanceCount

`int32` · optional (explicit presence)

Number of data node instances. Default: 1.
For zone-aware deployments, use a multiple of the availability zone count.

- default: `1`

### spec.clusterConfig.dedicatedMasterEnabled

`bool`

Enable dedicated master nodes for cluster stability. Dedicated masters handle
cluster management tasks (shard allocation, index state management) without
competing with data node workloads. Recommended for production.

### spec.clusterConfig.dedicatedMasterType

`string`

Instance type for dedicated master nodes. Does not need EBS storage.
Example: "r6g.large.search". Only used when `dedicated_master_enabled` is true.

### spec.clusterConfig.dedicatedMasterCount

`int32`

Number of dedicated master nodes. AWS recommends 3 for production (provides
quorum for split-brain protection). Only used when `dedicated_master_enabled` is true.

### spec.clusterConfig.nodeOptions

`[]AwsOpenSearchDomainNodeOption`

Additional node pools beyond data and master nodes. Today AWS supports one
pool type: "coordinator" nodes, which take over request routing, query
fan-out, and response aggregation so data nodes spend their capacity on
indexing and searching. Valuable for high-concurrency dashboards and
aggregation-heavy workloads.

- rule: an enabled node pool requires instance_type and count (at least 1)

### spec.clusterConfig.nodeOptions[].nodeType

`string` · required

The pool type. AWS currently supports "coordinator".

- rule: {"required":true,"string":{"in":["coordinator"]}}

### spec.clusterConfig.nodeOptions[].enabled

`bool`

Whether the pool is active. Set false to keep the pool definition while
removing its nodes.

### spec.clusterConfig.nodeOptions[].instanceType

`string`

Instance type for nodes in this pool (`.search` suffix).
Example: "m7g.large.search". Required when the pool is enabled.

### spec.clusterConfig.nodeOptions[].count

`int32`

Number of nodes in the pool (at least 1). Required when the pool is
enabled.

### spec.clusterConfig.zoneAwarenessEnabled

`bool`

Enable zone awareness to distribute data nodes and replicas across multiple
Availability Zones for resilience against AZ-level failures.

### spec.clusterConfig.availabilityZoneCount

`int32`

Number of Availability Zones. Must be 2 or 3.
Only used when `zone_awareness_enabled` is true.

### spec.clusterConfig.warmEnabled

`bool`

Enable UltraWarm storage tier for infrequently accessed, read-only data.
UltraWarm uses S3-backed storage at lower cost per GB than hot storage.
Requires warm_type and warm_count.

### spec.clusterConfig.warmType

`string`

Instance type for UltraWarm nodes: "ultrawarm1.medium.search",
"ultrawarm1.large.search", or "ultrawarm1.xlarge.search". Required when
`warm_enabled` is true.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["ultrawarm1.medium.search","ultrawarm1.large.search","ultrawarm1.xlarge.search"]}}

### spec.clusterConfig.warmCount

`int32`

Number of UltraWarm nodes. Range: 2-150. Required when `warm_enabled`
is true.

### spec.clusterConfig.coldStorageEnabled

`bool`

Enable cold storage backed by S3. Requires UltraWarm to be enabled.
Cold storage provides the lowest-cost tier for data that is rarely queried.

### spec.clusterConfig.multiAzWithStandbyEnabled

`bool`

Enable Multi-AZ with Standby for 99.99% availability SLA. Deploys standby
nodes in a different AZ that take over automatically during AZ failures.
Requires 3 AZs and at least 3 data nodes.

### spec.ebsOptions

`AwsOpenSearchDomainEbsOptions` · required

EBS volume configuration for data node storage. Required for most instance types
(all except certain storage-optimized types that use instance storage).

- rule: {"required":true}
- rule: volume_type, volume_size, iops, and throughput require ebs_enabled to be true
- rule: volume_type must be 'gp3', 'gp2', 'io1', or 'standard' when set
- rule: iops is only valid for 'gp3' or 'io1' volume types
- rule: throughput is only valid for 'gp3' volume type
- rule: throughput must be at least 125 MiB/s when set

### spec.ebsOptions.ebsEnabled

`bool`

Whether EBS volumes are attached to data nodes. Required for most instance
types. Only storage-optimized instances (e.g., i3) use instance storage.

### spec.ebsOptions.volumeType

`string`

EBS volume type. "gp3" (recommended), "gp2", "io1", or "standard".
gp3 provides predictable performance with configurable IOPS and throughput.

### spec.ebsOptions.volumeSize

`int32`

Size of each EBS volume in GB. The total storage is volume_size * instance_count.

### spec.ebsOptions.iops

`int32`

Provisioned IOPS for the volume. Only valid for "gp3" and "io1" volume types.
gp3 baseline: 3000 IOPS. io1: specify based on workload.

### spec.ebsOptions.throughput

`int32`

Provisioned throughput in MiB/s. Only valid for "gp3" volume type.
Minimum: 125 MiB/s. gp3 baseline: 125 MiB/s.

### spec.encryptAtRestEnabled

`bool` · optional (explicit presence)

Enable encryption at rest for indices and automated snapshots. Uses the
AWS-managed `aws/es` key unless `kms_key_id` is provided. Defaults to
TRUE (secure by default); set an explicit false only for legacy engine
versions that cannot encrypt.
One-way: encryption cannot be disabled once enabled (disabling forces
domain recreation), and enabling it on very old Elasticsearch versions
(< 6.7) also forces recreation.

- default: `true`

### spec.kmsKeyId

`string | valueFrom`

Customer-managed KMS key ARN or ID for at-rest encryption. ForceNew — the
KMS key cannot be changed after domain creation.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.nodeToNodeEncryptionEnabled

`bool` · optional (explicit presence)

Enable TLS encryption for all traffic between nodes in the cluster.
Defaults to TRUE (secure by default). One-way: cannot be disabled once
enabled (disabling forces domain recreation).

- default: `true`

### spec.vpcOptions

`AwsOpenSearchDomainVpcOptions`

VPC placement configuration. When provided, the domain is deployed into VPC
subnets and is not publicly accessible. ForceNew — adding or removing VPC
options destroys and recreates the domain.

- rule: vpc_options requires at least one subnet_id

### spec.vpcOptions.subnetIds

`[]string | valueFrom`

Subnet IDs where OpenSearch deploys ENIs. For zone-aware domains, provide
subnets in 2 or 3 AZs matching the cluster's availability_zone_count.

- references: AwsSubnet (`status.outputs.subnet_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.vpcOptions.securityGroupIds

`[]string | valueFrom`

Security group IDs controlling inbound/outbound traffic to the domain.
Must allow HTTPS (port 443) from clients that need to access OpenSearch.

- references: AwsSecurityGroup (`status.outputs.security_group_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecurityGroup, name: <that resource's name>, fieldPath: status.outputs.security_group_id}} -- a bare string does not parse

### spec.domainEndpointOptions

`AwsOpenSearchDomainEndpointOptions`

HTTPS enforcement, TLS policy, and custom endpoint configuration.

- rule: custom_endpoint requires custom_endpoint_enabled to be true
- rule: custom_endpoint_certificate_arn requires custom_endpoint_enabled to be true
- rule: custom_endpoint_certificate_arn literal value must be an ACM certificate ARN (arn:...)

### spec.domainEndpointOptions.enforceHttps

`bool` · optional (explicit presence)

Require HTTPS for all traffic to the domain endpoint. Default: true.
Strongly recommended for all environments.

- default: `true`

### spec.domainEndpointOptions.tlsSecurityPolicy

`string`

TLS security policy for the HTTPS endpoint. Controls the minimum TLS version
and cipher suites. "Policy-Min-TLS-1-2-PFS-2023-10" is the recommended
baseline; the FIPS policy serves regulated workloads. Leave empty for the
provider default.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["Policy-Min-TLS-1-0-2019-07","Policy-Min-TLS-1-2-2019-07","Policy-Min-TLS-1-2-PFS-2023-10","Policy-Min-TLS-1-2-RFC9151-FIPS-2024-08"]}}

### spec.domainEndpointOptions.customEndpointEnabled

`bool`

Enable a custom domain endpoint (e.g., "search.example.com") instead of the
AWS-generated endpoint.

### spec.domainEndpointOptions.customEndpoint

`string`

The fully qualified domain name for the custom endpoint.
Only used when `custom_endpoint_enabled` is true.

### spec.domainEndpointOptions.customEndpointCertificateArn

`string | valueFrom`

ACM certificate ARN for the custom endpoint. Must be a valid certificate
covering the custom_endpoint FQDN.
Only used when `custom_endpoint_enabled` is true.

- references: AwsCertManagerCert (`status.outputs.cert_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsCertManagerCert, name: <that resource's name>, fieldPath: status.outputs.cert_arn}} -- a bare string does not parse

### spec.advancedSecurityOptions

`AwsOpenSearchDomainAdvancedSecurityOptions`

Advanced security options enable fine-grained access control: internal user
database, IAM-based authentication, JWT bearer authentication, and role-based
index-level permissions. Once enabled, FGAC cannot be disabled (ForceNew).

- rule: internal_user_database_enabled requires advanced security (enabled) to be true
- rule: anonymous_auth_enabled requires advanced security (enabled) to be true
- rule: master_user_arn and master_user_name/master_user_password are mutually exclusive; use IAM-based OR internal user database authentication, not both
- rule: when advanced security is enabled, provide either master_user_arn (IAM) or master_user_name + master_user_password (internal user database)
- rule: master_user_password requires master_user_name to be set
- rule: master_user_name requires advanced security (enabled) to be true
- rule: jwt_options requires advanced security (enabled) to be true
- rule: master_user_arn literal value must be an IAM user or role ARN (arn:...)

### spec.advancedSecurityOptions.enabled

`bool`

Enable fine-grained access control. ForceNew if disabling (cannot disable
once enabled without domain recreation).

### spec.advancedSecurityOptions.internalUserDatabaseEnabled

`bool`

Enable the internal user database. When true, you can create users and roles
directly in OpenSearch Dashboards. When false, use IAM or SAML for authentication.

### spec.advancedSecurityOptions.anonymousAuthEnabled

`bool`

Allow anonymous (unauthenticated) requests while FGAC is enabled -- roles
mapped to the anonymous backend decide what such requests may do. Can only
be enabled at the moment FGAC itself is first enabled on the domain.

### spec.advancedSecurityOptions.masterUserArn

`string | valueFrom`

IAM entity ARN (user or role) designated as the master user. The master user
has full access to the cluster, indices, and OpenSearch Dashboards.
Mutually exclusive with master_user_name/master_user_password.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.advancedSecurityOptions.masterUserName

`string`

Username for the internal user database master user.
Mutually exclusive with master_user_arn.

### spec.advancedSecurityOptions.masterUserPassword

`string | valueFrom` · sensitive

Password for the internal user database master user. Must be at least 8
characters with uppercase, lowercase, digit, and special character.
Mutually exclusive with master_user_arn.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.advancedSecurityOptions.jwtOptions

`AwsOpenSearchDomainJwtOptions`

JWT bearer-token authentication: clients present tokens signed by an
external identity provider (validated against jwks_url or public_key).
Requires OpenSearch 2.11+.

- rule: provide jwks_url or public_key when jwt_options is configured

### spec.advancedSecurityOptions.jwtOptions.enabled

`bool`

Whether JWT authentication is active.

### spec.advancedSecurityOptions.jwtOptions.jwksUrl

`string`

URL of the identity provider's JSON Web Key Set used to validate token
signatures (e.g., "https://idp.example.com/.well-known/jwks.json").
Provide jwks_url or public_key (or both).

- rule: {"string":{"maxLen":"2048"}}

### spec.advancedSecurityOptions.jwtOptions.publicKey

`string`

PEM-encoded public key used to validate token signatures. An alternative to
jwks_url for providers without a JWKS endpoint. The key material is public
by definition -- it verifies signatures, it cannot create them.

### spec.advancedSecurityOptions.jwtOptions.rolesKey

`string`

The token claim that carries the user's roles/groups. Example: "roles".
Up to 64 characters.

- rule: {"string":{"maxLen":"64"}}

### spec.advancedSecurityOptions.jwtOptions.subjectKey

`string`

The token claim that identifies the user. Example: "sub", "email".
Up to 64 characters.

- rule: {"string":{"maxLen":"64"}}

### spec.cognitoOptions

`AwsOpenSearchDomainCognitoOptions`

Amazon Cognito authentication for OpenSearch Dashboards: users sign in
through a Cognito user pool and are authorized through a Cognito identity
pool. An alternative to FGAC's internal user database for the Dashboards
UI (the two can also be combined -- Cognito authenticates, FGAC authorizes).

- rule: user_pool_id, identity_pool_id, and role_arn are required when Cognito authentication is enabled

### spec.cognitoOptions.enabled

`bool`

Whether Cognito authentication for Dashboards is active.

### spec.cognitoOptions.userPoolId

`string | valueFrom`

The Cognito user pool users sign in through.

- references: AwsCognitoUserPool (`status.outputs.user_pool_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsCognitoUserPool, name: <that resource's name>, fieldPath: status.outputs.user_pool_id}} -- a bare string does not parse

### spec.cognitoOptions.identityPoolId

`string`

The Cognito identity pool that exchanges user-pool sign-ins for AWS
credentials. Identity pools have no Planton kind; provide the raw ID
(format: "<region>:<uuid>").

### spec.cognitoOptions.roleArn

`string | valueFrom`

IAM role that allows OpenSearch Service to configure the user and identity
pools (typically carrying the AmazonOpenSearchServiceCognitoAccess policy).

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.logPublishingOptions

`[]AwsOpenSearchDomainLogPublishingOption`

Publish domain logs to CloudWatch Logs for monitoring and troubleshooting.
Up to 4 configurations — one per log type (INDEX_SLOW_LOGS, SEARCH_SLOW_LOGS,
ES_APPLICATION_LOGS, AUDIT_LOGS).

- rule: cloudwatch_log_group_arn literal value must be a CloudWatch log group ARN (arn:...)
- rule: log_type must be 'INDEX_SLOW_LOGS', 'SEARCH_SLOW_LOGS', 'ES_APPLICATION_LOGS', or 'AUDIT_LOGS'

### spec.logPublishingOptions[].logType

`string` · required

Type of log to publish. Values:
- "INDEX_SLOW_LOGS" — indexing operations exceeding the slow log threshold
- "SEARCH_SLOW_LOGS" — search queries exceeding the slow log threshold
- "ES_APPLICATION_LOGS" — OpenSearch application and error logs
- "AUDIT_LOGS" — fine-grained access control audit trail (requires FGAC enabled)

- rule: {"required":true}

### spec.logPublishingOptions[].cloudwatchLogGroupArn

`string | valueFrom` · required

CloudWatch Logs log group ARN where logs are published.

- references: AwsCloudwatchLogGroup (`status.outputs.log_group_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsCloudwatchLogGroup, name: <that resource's name>, fieldPath: status.outputs.log_group_arn}} -- a bare string does not parse

### spec.logPublishingOptions[].enabled

`bool` · optional (explicit presence)

Whether this log publishing option is active. Default: true.
Set to false to temporarily disable publishing without removing the configuration.

- default: `true`

### spec.accessPolicies

`object`

IAM-based access policy for the domain. Serialized to JSON by the IaC modules.
Controls who can perform actions on the domain and its indices.
For VPC domains, this works in conjunction with security groups.
For public domains, this is the primary access control mechanism (unless FGAC is enabled).

### spec.autoTuneOptions

`AwsOpenSearchDomainAutoTuneOptions`

AWS Auto-Tune automatically optimizes JVM heap size, disk I/O, and other
performance settings based on cluster metrics. Configure maintenance windows
for changes that require a blue/green deployment. Not supported on t2/t3
(burstable) instance types.

- rule: maintenance_schedules and use_off_peak_window are mutually exclusive scheduling mechanisms

### spec.autoTuneOptions.desiredState

`string` · required

Auto-Tune state. "ENABLED" or "DISABLED".

- rule: {"required":true,"string":{"in":["ENABLED","DISABLED"]}}

### spec.autoTuneOptions.maintenanceSchedules

`[]AwsOpenSearchDomainAutoTuneMaintenanceSchedule`

Recurring windows during which Auto-Tune may apply blue/green optimizations.
Ignored when use_off_peak_window is true (the off-peak window is used
instead).

### spec.autoTuneOptions.maintenanceSchedules[].startAt

`string` · required

When the first window opens, as an RFC3339 timestamp.
Example: "2026-08-01T03:00:00Z".

- rule: {"required":true}

### spec.autoTuneOptions.maintenanceSchedules[].durationHours

`int32` · required

Length of each window, in hours (AWS's only supported duration unit).

- rule: {"required":true,"int32":{"gte":1}}

### spec.autoTuneOptions.maintenanceSchedules[].cronExpressionForRecurrence

`string` · required

Recurrence as a cron expression. Example: "cron(0 3 ? * SUN *)" for
Sundays at 03:00 UTC.

- rule: {"required":true}

### spec.autoTuneOptions.rollbackOnDisable

`string`

What happens to Auto-Tune's applied changes when Auto-Tune is later
disabled: "NO_ROLLBACK" (keep the tuned settings) or "DEFAULT_ROLLBACK"
(revert to defaults -- requires a maintenance schedule to perform the
rollback in).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["NO_ROLLBACK","DEFAULT_ROLLBACK"]}}

### spec.autoTuneOptions.useOffPeakWindow

`bool`

Schedule blue/green optimizations inside the domain's off-peak window
instead of explicit maintenance schedules.

### spec.automatedSnapshotStartHour

`int32` · optional (explicit presence)

Hour of day (0-23, UTC) when the service takes an automated daily snapshot
of the domain's indices. Only relevant for domains running Elasticsearch
versions below 5.3 — newer versions snapshot hourly regardless — but the
setting remains configurable on all domains.

- rule: {"int32":{"lte":23,"gte":0}}

### spec.offPeakWindowOptions

`AwsOpenSearchDomainOffPeakWindowOptions`

Daily 10-hour low-traffic window during which AWS schedules service software
updates and Auto-Tune blue/green optimizations. AWS defaults the window to
10:00 PM local time when not configured.

### spec.offPeakWindowOptions.enabled

`bool` · optional (explicit presence)

Whether the off-peak window feature is active. Defaults to TRUE -- AWS
enables the window on every new domain and requires it enabled at create
time; false is only meaningful as a later update (AWS may still reject
disabling it while features depend on the window).

- default: `true`

### spec.offPeakWindowOptions.windowStartHour

`int32` · optional (explicit presence)

Hour (0-23, local time) when the 10-hour window opens. AWS defaults to
22 (10:00 PM) when not configured.

- rule: {"int32":{"lte":23,"gte":0}}

### spec.offPeakWindowOptions.windowStartMinute

`int32` · optional (explicit presence)

Minute (0-59) past the hour when the window opens.

- rule: {"int32":{"lte":59,"gte":0}}

### spec.autoSoftwareUpdateEnabled

`bool`

Enable automatic service software updates. When true, AWS applies mandatory
and optional service software updates during the off-peak window. Both
engines always send this setting explicitly, so false actively turns
auto-updates OFF (including on domains where they were previously on).

### spec.deploymentStrategy

`string`

Blue/green deployment strategy for configuration changes that require one.
"Default": AWS picks the standard strategy. "CapacityOptimized": AWS
provisions additional capacity more conservatively to reduce the performance
impact of the migration on capacity-sensitive clusters.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["Default","CapacityOptimized"]}}

### spec.ipAddressType

`string`

IP address type for the domain. "ipv4" (default) or "dualstack" (IPv4 + IPv6).
One-way: changing from "dualstack" back to "ipv4" forces domain recreation.

### spec.advancedOptions

`map<string, string>`

Low-level key-value configuration options. Common options:
- "rest.action.multi.allow_explicit_index": "true" (default)
- "indices.fielddata.cache.size": percentage of heap
- "indices.query.bool.max_clause_count": max boolean clauses
Values must be strings.

### spec.aimlOptions

`AwsOpenSearchDomainAimlOptions`

AI/ML capabilities on the domain: natural-language query generation in
Dashboards, the S3 vectors engine, and GPU-accelerated vector search.

### spec.aimlOptions.naturalLanguageQueryGenerationDesiredState

`string`

Natural-language query generation in OpenSearch Dashboards ("ENABLED" or
"DISABLED"): users describe a query in plain language and Dashboards
generates the DSL. Requires OpenSearch 2.13+.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["ENABLED","DISABLED"]}}

### spec.aimlOptions.s3VectorsEngineEnabled

`bool`

Enable the S3 vectors engine: vector indexes stored on S3-backed storage
for large, cost-efficient vector search corpora.

### spec.aimlOptions.serverlessVectorAccelerationEnabled

`bool`

Enable GPU-accelerated (serverless) vector search acceleration on the
domain.

### spec.identityCenterOptions

`AwsOpenSearchDomainIdentityCenterOptions`

AWS IAM Identity Center (successor to AWS SSO) integration for OpenSearch
Dashboards and API access — workforce users sign in with their Identity
Center identity instead of IAM credentials.

- rule: identity_center_instance_arn is required when enabled_api_access is true

### spec.identityCenterOptions.enabledApiAccess

`bool`

Whether Identity Center API access is active.

### spec.identityCenterOptions.identityCenterInstanceArn

`string`

ARN of the IAM Identity Center instance to integrate with.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^arn:"}}

### spec.identityCenterOptions.rolesKey

`string`

The Identity Center attribute that carries a user's groups for role
mapping: "GroupName" or "GroupId".

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["GroupName","GroupId"]}}

### spec.identityCenterOptions.subjectKey

`string`

The Identity Center attribute that identifies a user: "UserName", "UserId",
or "Email".

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["UserName","UserId","Email"]}}

### spec.samlOptions

`AwsOpenSearchDomainSamlOptions`

SAML authentication for OpenSearch Dashboards: users sign in through an
external SAML 2.0 identity provider (Okta, Entra ID, ...). Managed as its
own provider resource alongside the domain; removing the block disables
SAML. Requires fine-grained access control (advanced_security_options).

- rule: master_backend_role and master_user_name are mutually exclusive

### spec.samlOptions.idpEntityId

`string` · required

The identity provider's entity ID (from the IdP metadata).
Example: "https://idp.example.com/saml/metadata".

- rule: {"string":{"minLen":"1"}}

### spec.samlOptions.idpMetadataContent

`string` · required

The IdP's SAML metadata document, as an XML string (typically downloaded
from the IdP's metadata endpoint).

- rule: {"string":{"minLen":"1"}}

### spec.samlOptions.masterBackendRole

`string`

The SAML backend role granted master (admin) access in Dashboards --
mutually exclusive with master_user_name. Empty grants no SAML
principal master access (map roles inside Dashboards instead).

### spec.samlOptions.masterUserName

`string` · sensitive

The SAML username granted master (admin) access in Dashboards --
mutually exclusive with master_backend_role.

### spec.samlOptions.rolesKey

`string`

The SAML assertion attribute that carries the user's roles/groups.
Empty uses the IdP's roles attribute conventions.

### spec.samlOptions.subjectKey

`string`

The SAML assertion attribute that identifies the user. Empty uses
NameID.

### spec.samlOptions.sessionTimeoutMinutes

`int32`

Dashboards session lifetime in minutes, 1-1440. 0 keeps the AWS
default (60).

- rule: {"int32":{"lte":1440,"gte":0}}

### spec.authorizedVpcEndpointAccessAccounts

`[]string`

AWS account IDs authorized to create OpenSearch-managed VPC endpoints
against this domain (the grantor side of cross-account private access).
Each entry is managed as its own provider resource keyed by account ID.

- rule: {"repeated":{"unique":true}}

## Validation Rules

- `engine_version_format`: engine_version must match 'OpenSearch_X.Y' or 'Elasticsearch_X.Y' format (e.g., 'OpenSearch_2.11', 'Elasticsearch_7.10')
- `kms_requires_encryption`: kms_key_id requires encrypt_at_rest_enabled to be true
- `ip_address_type_valid`: ip_address_type must be 'ipv4' or 'dualstack' when set
- `log_options_max_four`: at most 4 log publishing options are allowed (one per log_type)
- `audit_logs_require_fgac`: AUDIT_LOGS log type requires advanced_security_options to be enabled
- `saml_requires_fgac`: saml_options requires advanced_security_options to be enabled (SAML roles map through fine-grained access control)
- `jwt_requires_opensearch_2_11`: advanced_security_options.jwt_options requires engine_version OpenSearch_2.11 or newer (not available on Elasticsearch engines)
- `authorized_accounts_format`: authorized_vpc_endpoint_access_accounts entries must be 12-digit AWS account IDs

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsOpenSearchDomain, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.domain_id` | `string` | The unique identifier assigned to the domain by AWS. |
| `status.outputs.domain_name` | `string` | The name of the domain. Matches the name derived from metadata. |
| `status.outputs.domain_arn` | `string` | The Amazon Resource Name of the domain. Used in IAM policies, cross-service permissions, and as a reference in other AWS resources. |
| `status.outputs.endpoint` | `string` | The domain-specific endpoint for submitting index, search, and data upload requests. For VPC domains, this is a VPC endpoint. For public domains, this is an internet-accessible endpoint. Format: "search-{domain-name}-{id}.{region}.es.amazonaws.com" (no https://). |
| `status.outputs.dashboard_endpoint` | `string` | The endpoint for OpenSearch Dashboards (the visualization and management UI). Format: endpoint + "/_dashboards". |
| `status.outputs.endpoint_v2` | `string` | The dual-stack (IPv4 + IPv6) V2 domain endpoint that works with both ip_address_type settings. Populated for domains created or migrated onto the V2 endpoint format. |
| `status.outputs.dashboard_endpoint_v2` | `string` | The OpenSearch Dashboards endpoint on the dual-stack V2 domain endpoint. |
| `status.outputs.domain_endpoint_v2_hosted_zone_id` | `string` | The Route 53 hosted zone ID to alias when pointing DNS records at the V2 domain endpoint. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.kmsKeyId` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.vpcOptions.subnetIds` | AwsSubnet | `status.outputs.subnet_id` |
| `spec.vpcOptions.securityGroupIds` | AwsSecurityGroup | `status.outputs.security_group_id` |
| `spec.domainEndpointOptions.customEndpointCertificateArn` | AwsCertManagerCert | `status.outputs.cert_arn` |
| `spec.advancedSecurityOptions.masterUserArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.cognitoOptions.userPoolId` | AwsCognitoUserPool | `status.outputs.user_pool_id` |
| `spec.cognitoOptions.roleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.logPublishingOptions[].cloudwatchLogGroupArn` | AwsCloudwatchLogGroup | `status.outputs.log_group_arn` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsAppSyncApi | `spec.datasources[].opensearch.endpoint` | `status.outputs.endpoint` |
| AwsBedrockKnowledgeBase | `spec.storage.opensearchManaged.domainArn` | `status.outputs.domain_arn` |
| AwsKinesisFirehose | `spec.opensearch.domainArn` | `status.outputs.domain_arn` |

## See Also

- [Overview](../README.md)
