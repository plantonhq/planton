# AwsSecretsManagerSecret

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

AwsSecretsManagerSecretSpec defines the desired configuration for an AWS
Secrets Manager secret - a named, versioned, KMS-encrypted container for
credential material (database passwords, API keys, tokens, key/value JSON
documents) with optional automatic rotation and cross-region replication.

The secret's AWS name defaults to `metadata.name` and can be overridden
with `secret_name` when the name needs shapes `metadata.name` cannot
carry: AWS allows alphanumeric plus /_+=.@-, so hierarchical names like
"prod/payments/db" and service-required prefixes like
"ecr-pullthroughcache/dockerhub" (ECR pull-through-cache credentials)
are expressed through `secret_name`. Either way the name is ForceNew -
changing it (including setting `secret_name` on a secret already
deployed under `metadata.name`) destroys and recreates the secret.

The secret VALUE is set through exactly one of `string_value` (text or
JSON key/value document - the common case) or `binary_value`
(base64-encoded binary). Both are sensitive fields: the platform stores
them as managed-secret references and resolves them just-in-time at
deploy, so plaintext never lives in the control plane. Omitting both
creates the secret shell with no value - useful when an application or
rotation function writes the first version itself.

Deletion is soft by default: AWS schedules the secret for deletion after
`recovery_window_in_days` (default 30) during which it can be restored;
0 forces immediate, unrecoverable deletion. A deleted-but-recoverable
secret still holds its name - recreating a same-named secret during the
window requires waiting or force deletion (the modules retry create
through AWS's "scheduled for deletion" window).

Create-time-immutable (ForceNew) fields: the name (secret_name, else
metadata.name) and `type`. Everything else - value, KMS key, policy,
replicas, rotation - updates in place.

Credentials, region, and deployment workflow live outside this spec in
stack inputs.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsSecretsManagerSecret
metadata:
  name: test-app-credentials
  id: test-app-credentials
  org: test-org
  env: dev
spec:
  region: us-west-2
  description: Full-surface Secrets Manager hack manifest
  kmsKeyId:
    value: arn:aws:kms:us-west-2:123456789012:key/abc-123
  stringValue: '{"username":"app","password":"example-password"}'
  versionStages:
    - bluegreen-active
  policy:
    Version: "2012-10-17"
    Statement:
      - Sid: AllowAccountRead
        Effect: Allow
        Principal:
          AWS: arn:aws:iam::123456789012:root
        Action: secretsmanager:GetSecretValue
        Resource: "*"
  blockPublicPolicy: true
  replicaRegions:
    - region: us-east-1
    - region: eu-west-1
      kmsKeyId:
        value: arn:aws:kms:eu-west-1:123456789012:key/def-456
  forceOverwriteReplicaSecret: false
  recoveryWindowInDays: 7
  rotation:
    rotationLambdaArn:
      value: arn:aws:lambda:us-west-2:123456789012:function:rotate-app-credentials
    automaticallyAfterDays: 30
    rotateImmediately: true
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.secretName` | `string` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.kmsKeyId` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.stringValue` | `string` (sensitive) |  |  |  |
| `spec.binaryValue` | `string` (sensitive) |  |  |  |
| `spec.versionStages` | `[]string` |  |  |  |
| `spec.policy` | `object` |  |  |  |
| `spec.blockPublicPolicy` | `bool` |  | `true` |  |
| `spec.replicaRegions` | `[]AwsSecretsManagerSecretReplica` |  |  |  |
| `spec.replicaRegions[].region` | `string` | yes |  |  |
| `spec.replicaRegions[].kmsKeyId` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.forceOverwriteReplicaSecret` | `bool` |  |  |  |
| `spec.recoveryWindowInDays` | `int32` |  | `30` |  |
| `spec.type` | `string` |  |  |  |
| `spec.rotation` | `AwsSecretsManagerSecretRotation` |  |  |  |
| `spec.rotation.rotationLambdaArn` | `string \| valueFrom` |  |  | AwsLambda (`status.outputs.function_arn`) |
| `spec.rotation.externalRotationRoleArn` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.rotation.externalRotationMetadata` | `[]AwsSecretsManagerSecretRotationMetadata` |  |  |  |
| `spec.rotation.externalRotationMetadata[].key` | `string` | yes |  |  |
| `spec.rotation.externalRotationMetadata[].value` | `string` | yes |  |  |
| `spec.rotation.automaticallyAfterDays` | `int32` |  |  |  |
| `spec.rotation.scheduleExpression` | `string` |  |  |  |
| `spec.rotation.duration` | `string` |  |  |  |
| `spec.rotation.rotateImmediately` | `bool` |  | `true` |  |

## Field Details

### spec.region

`string` · required

The AWS region where the secret will be created.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.secretName

`string`

The explicit AWS-side secret name. Empty (the common case) means the
secret is named after `metadata.name`. Set it when the AWS name needs
shapes `metadata.name` cannot carry - hierarchical paths
("prod/payments/db") or service-required prefixes
("ecr-pullthroughcache/dockerhub"). The catalog's convention:
explicit-name fields are REQUIRED only on kinds whose AWS names can
never come from `metadata.name` (hierarchy-dominant or
charset-incompatible ones); where `metadata.name` is a legal default,
the field is optional so the common case stays terse. ForceNew:
setting or changing it replaces the secret.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"512","pattern":"^[A-Za-z0-9/_+=.@-]+$"}}

### spec.description

`string`

Human-readable description shown in the AWS console and
ListSecrets/DescribeSecret responses. Updates in place.

### spec.kmsKeyId

`string | valueFrom`

Customer-managed KMS key for encrypting the secret value - accepts a key
ARN, key ID, or alias. Without it, AWS uses the account's AWS-managed key
`aws/secretsmanager`. Required in practice when other AWS accounts must
read the secret (the AWS-managed key cannot be granted cross-account).
Updates in place - new versions encrypt under the new key; existing
versions remain readable (Secrets Manager tracks the key per version).

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.stringValue

`string` · sensitive

The secret value as text. Commonly a JSON key/value document (the shape
the Secrets Manager console and rotation functions expect, e.g.
'{"username":"app","password":"..."}') but any string up to 65536 bytes
is legal. Mutually exclusive with `binary_value`. Sensitive: supply a
managed-secret reference; the platform resolves it just-in-time at
deploy. Updating the value publishes a new version staged AWSCURRENT
(the previous version moves to AWSPREVIOUS).

### spec.binaryValue

`string` · sensitive

The secret value as base64-encoded binary (up to 65536 bytes decoded).
For non-text material such as raw keys or certificates in DER form.
Mutually exclusive with `string_value`. Sensitive: supply a
managed-secret reference resolved just-in-time at deploy.

### spec.versionStages

`[]string`

Additional staging labels to attach to the managed version beyond the
AWSCURRENT label AWS assigns automatically. Staging labels partition a
secret's versions (rotation moves AWSCURRENT/AWSPENDING/AWSPREVIOUS);
custom labels support blue/green-style consumption. Each label is 1-256
characters; a label can only be attached to one version at a time.
Only meaningful when a value arm is set.

- rule: {"repeated":{"items":{"string":{"minLen":"1","maxLen":"256"}}}}

### spec.policy

`object`

Resource-based IAM policy controlling who can read or manage this
secret - the mechanism for cross-account access and for restricting
reads to specific principals. Standard IAM policy document as a typed
object (Version/Statement). Removing the field deletes the policy.
Updates in place.

### spec.blockPublicPolicy

`bool` · optional (explicit presence)

Reject the policy if AWS determines it grants public (anonymous)
access - Secrets Manager's PutResourcePolicy validation gate, the
secure posture for any policy that is not deliberately public. Only
meaningful together with `policy`.

- default: `true`

### spec.replicaRegions

`[]AwsSecretsManagerSecretReplica`

Regions to replicate the secret to. Each replica is a read-only copy
(same name, regional ARN) kept in sync by Secrets Manager - consumers
in that region read locally with no cross-region call. Add or remove
entries in place; removed regions have their replica deleted - but
note that AWS performs that deletion ASYNCHRONOUSLY (seconds to
minutes, live-verified 2026-08-13): destroying a replicated secret
with recovery_window_in_days 0 can outrun it and strand a replica as
a live standalone secret in its region. Prefer a recovery window
(the primary's tombstone lets the async deletion complete), and if a
replica ever strands, recover with StopReplicationToReplica in the
replica's region followed by DeleteSecret - a stranded replica
rejects a direct delete ("Operation not permitted on a replica
secret") even after its primary is gone.

### spec.replicaRegions[].region

`string` · required

Region to replicate the secret into. Example: "us-east-1". Must differ
from the secret's own region (AWS rejects self-replication at the API).

- rule: {"required":true}

### spec.replicaRegions[].kmsKeyId

`string | valueFrom`

KMS key for encrypting the replica in its region - accepts a key ARN,
key ID, or alias. Must be a key that LIVES in the replica region (keys
are regional; a multi-Region key's replica in that region also works).
Without it, the replica uses that region's AWS-managed
`aws/secretsmanager` key.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.forceOverwriteReplicaSecret

`bool`

Overwrite an existing secret of the same name in a replica region
instead of failing replication. Default false - failing loudly on a
name collision is the safe posture; enable it deliberately when
re-pointing replication at regions that hold stale copies. It does
NOT clear a same-name secret that is itself a stranded ex-replica of
a deleted secret: replication then fails with "currently replicated
to <region> with a different arn" (live-verified 2026-08-13), and the
failure is SILENT at apply on both engines - neither waits on
replication status. Check ReplicationStatus (or the console) after
enabling replication into a region with name history.

### spec.recoveryWindowInDays

`int32` · optional (explicit presence)

Days AWS retains the secret in a recoverable soft-deleted state after
destroy: 7-30, or 0 to delete immediately and permanently. Default 30.
Consumed only at delete time - it never affects the running secret.
During the window the secret's NAME stays reserved; 0 is the right
choice for ephemeral/test secrets that must be recreatable immediately.
For secrets with replica_regions, prefer a non-zero window: 0
force-deletes the primary immediately after replica removal is
requested, which can outrun AWS's asynchronous replica deletion and
strand live replica secrets (see replica_regions).

- default: `30`

### spec.type

`string`

Partner identifier for an AWS "managed external secret" - a secret whose
authoritative value lives with a SaaS partner and is rotated through the
partner's integration (see the external arm on `rotation`). The exact
string comes from the partner's onboarding documentation. Leave empty
for ordinary self-managed secrets. ForceNew - changing it destroys and
recreates the secret.

- rule: {"string":{"maxLen":"256"}}

### spec.rotation

`AwsSecretsManagerSecretRotation`

Automatic rotation configuration. When set, Secrets Manager rotates the
secret on the configured cadence through exactly one of two mechanisms:
a rotation Lambda function you own (`rotation_lambda_arn` - the classic
path; the function receives createSecret/setSecret/testSecret/
finishSecret steps), or a partner-managed external rotation
(`external_rotation_role_arn` - pairs with `type`). Removing the block
cancels rotation (the secret and its versions are untouched).

- rule: set exactly one of automatically_after_days or schedule_expression
- rule: set exactly one of rotation_lambda_arn (self-managed rotation function) or external_rotation_role_arn (partner-managed external rotation)
- rule: external_rotation_metadata is only used with external_rotation_role_arn

### spec.rotation.rotationLambdaArn

`string | valueFrom`

ARN of the Lambda function that performs rotation - the self-managed
mechanism. The function must grant Secrets Manager invoke permission
(lambda permission with principal secretsmanager.amazonaws.com); AWS
rejects RotateSecret otherwise (the modules retry through IAM
propagation, but a missing permission fails the deploy). Mutually
exclusive with `external_rotation_role_arn`.

- references: AwsLambda (`status.outputs.function_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsLambda, name: <that resource's name>, fieldPath: status.outputs.function_arn}} -- a bare string does not parse

### spec.rotation.externalRotationRoleArn

`string | valueFrom`

IAM role Secrets Manager assumes to rotate a managed EXTERNAL secret
through the partner integration (pairs with the spec's `type` partner
identifier). Mutually exclusive with `rotation_lambda_arn`.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.rotation.externalRotationMetadata

`[]AwsSecretsManagerSecretRotationMetadata`

Key/value metadata passed to the partner's external rotation
integration (partner-defined keys - consult the partner's onboarding
documentation). Only meaningful with `external_rotation_role_arn`.

### spec.rotation.externalRotationMetadata[].key

`string` · required

Metadata key (partner-defined).

- rule: {"required":true}

### spec.rotation.externalRotationMetadata[].value

`string` · required

Metadata value.

- rule: {"required":true}

### spec.rotation.automaticallyAfterDays

`int32` · optional (explicit presence)

Rotate every N days (1-1000). AWS derives a schedule window from it.
Exactly one of `automatically_after_days` or `schedule_expression`
must be set.

- rule: {"int32":{"lte":1000,"gte":1}}

### spec.rotation.scheduleExpression

`string`

Rotation schedule as a cron or rate expression - the precise-control
alternative to `automatically_after_days` (exactly one of the two).
Examples: "rate(10 days)", "cron(0 16 1,15 * ? *)". Secrets Manager
cron expressions are UTC and must resolve to at most one rotation per
day.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^(rate|cron)\\(.+\\)$"}}

### spec.rotation.duration

`string`

Length of the rotation window in hours, e.g. "3h" (1h-24h). Only valid
together with a schedule; without it AWS ends the window at the end of
the schedule's day. Format: a number followed by "h".

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[0-9]{1,2}h$"}}

### spec.rotation.rotateImmediately

`bool` · optional (explicit presence)

Rotate once immediately when rotation is configured (default true).
Set false to test the rotation configuration without touching the
current value - AWS then only tests the setup at the next scheduled
window. Client-side flag consumed at configure time; it is not echoed
back by AWS and re-applies on every rotation-config update.

- default: `true`

## Validation Rules

- `value_arms_mutual_exclusion`: string_value and binary_value are mutually exclusive; a secret version holds one or the other
- `version_stages_require_value`: version_stages only apply to the version this manifest manages; set string_value or binary_value
- `binary_value_base64`: binary_value must be base64-encoded (A-Za-z0-9+/ with optional = padding)
- `recovery_window_valid_range`: recovery_window_in_days must be 0 (immediate permanent deletion) or between 7 and 30

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsSecretsManagerSecret, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.secret_arn` | `string` | The Amazon Resource Name of the secret - the canonical join key for IAM policies, ECS/Lambda secret injection, and cross-service references. Note AWS appends a random 6-character suffix to secret ARNs (arn:...:secret:name-AbCdEf), so the ARN is not derivable from the name. |
| `status.outputs.secret_name` | `string` | The secret's AWS name (spec.secret_name, else metadata.name). Consumers that resolve secrets by name (SDK GetSecretValue calls) join on this. |
| `status.outputs.version_id` | `string` | The version ID of the secret version this deployment manages (the AWSCURRENT version when a value arm is set; empty for a shell secret with no managed value). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.kmsKeyId` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.replicaRegions[].kmsKeyId` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.rotation.rotationLambdaArn` | AwsLambda | `status.outputs.function_arn` |
| `spec.rotation.externalRotationRoleArn` | AwsIamRole | `status.outputs.role_arn` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsAppSyncApi | `spec.datasources[].relationalDatabase.awsSecretStoreArn` | `status.outputs.secret_arn` |
| AwsBedrockAgentCoreEvaluation | `spec.harnesses[].model.gemini.apiKeyArn` | `status.outputs.secret_arn` |
| AwsBedrockAgentCoreEvaluation | `spec.harnesses[].model.openai.apiKeyArn` | `status.outputs.secret_arn` |
| AwsBedrockAgentCoreTools | `spec.browsers[].certificates[].secretArn` | `status.outputs.secret_arn` |
| AwsBedrockAgentCoreTools | `spec.codeInterpreters[].certificates[].secretArn` | `status.outputs.secret_arn` |
| AwsBedrockKnowledgeBase | `spec.sql.provisioned.auth.usernamePasswordSecretArn` | `status.outputs.secret_arn` |
| AwsBedrockKnowledgeBase | `spec.sql.serverless.auth.usernamePasswordSecretArn` | `status.outputs.secret_arn` |
| AwsBedrockKnowledgeBase | `spec.storage.rds.credentialsSecretArn` | `status.outputs.secret_arn` |
| AwsBedrockKnowledgeBase | `spec.storage.pinecone.credentialsSecretArn` | `status.outputs.secret_arn` |
| AwsBedrockKnowledgeBase | `spec.storage.mongodbAtlas.credentialsSecretArn` | `status.outputs.secret_arn` |
| AwsBedrockKnowledgeBase | `spec.storage.redisEnterpriseCloud.credentialsSecretArn` | `status.outputs.secret_arn` |
| AwsBedrockKnowledgeBase | `spec.dataSources[].confluence.credentialsSecretArn` | `status.outputs.secret_arn` |
| AwsBedrockKnowledgeBase | `spec.dataSources[].salesforce.credentialsSecretArn` | `status.outputs.secret_arn` |
| AwsBedrockKnowledgeBase | `spec.dataSources[].sharepoint.credentialsSecretArn` | `status.outputs.secret_arn` |
| AwsEcrRegistrySettings | `spec.pullThroughCacheRules[].credentialArn` | `status.outputs.secret_arn` |
| AwsFsxWindowsFileSystem | `spec.selfManagedActiveDirectory.domainJoinServiceAccountSecretArn` | `status.outputs.secret_arn` |
| AwsRdsProxy | `spec.auth[].secretArn` | `status.outputs.secret_arn` |

## See Also

- [Overview](../README.md)
