# AwsKmsKey

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsKmsKeySpec defines a customer-managed AWS KMS key: its
cryptographic shape (key spec and usage), the key policy that
governs access, automatic rotation, multi-region designation, and
friendly aliases.

KMS keys have no name in AWS -- identity is the generated key ID/ARN
-- so consumers compose with this key by referencing its key_arn
output (encryption-at-rest fields across databases, queues, buckets,
and functions all take it), or through an alias for humans and SDK
callers. metadata.name drives the Name identity tag.

Key-shape guidance: the default SYMMETRIC_DEFAULT encrypt/decrypt
key is what every AWS service-integration (S3, RDS, DynamoDB, MSK,
Lambda environment encryption, ...) expects -- choose an asymmetric
RSA/ECC spec only for external signing or public-key encryption
workflows, and an HMAC spec only for token authentication. The
shape is create-time immutable: changing key_spec or key_usage
replaces the key (old ciphertext stays decryptable only by the old
key -- plan key migration deliberately).

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsKmsKey
metadata:
  name: app-encryption-key-demo
spec:
  region: us-west-2
  description: "Demo encryption key for application data"
  aliases:
    - alias/demo/app-data
  deletionWindowDays: 30
  # Grants give principals scoped, revocable use of the key without editing
  # the key policy -- the pattern for wiring "this workload may encrypt and
  # decrypt under this key".
  grants:
    - name: orders-worker-data
      granteePrincipal:
        value: arn:aws:iam::123456789012:role/orders-worker
      operations:
        - Encrypt
        - Decrypt
        - GenerateDataKey
        - DescribeKey
      encryptionContextSubset:
        app: orders
      retireOnDelete: true
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.keySpec` | `string` |  |  |  |
| `spec.keyUsage` | `string` |  |  |  |
| `spec.policy` | `string` |  |  |  |
| `spec.bypassPolicyLockoutSafetyCheck` | `bool` |  |  |  |
| `spec.disabled` | `bool` |  |  |  |
| `spec.enableKeyRotation` | `bool` |  |  |  |
| `spec.rotationPeriodInDays` | `int32` |  |  |  |
| `spec.multiRegion` | `bool` |  |  |  |
| `spec.deletionWindowDays` | `int32` |  | `30` |  |
| `spec.aliases` | `[]string` |  |  |  |
| `spec.customKeyStoreId` | `string` |  |  |  |
| `spec.xksKeyId` | `string` |  |  |  |
| `spec.grants` | `[]AwsKmsKeyGrant` |  |  |  |
| `spec.grants[].name` | `string` |  |  |  |
| `spec.grants[].granteePrincipal` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.grants[].operations` | `[]string` | yes |  |  |
| `spec.grants[].retiringPrincipal` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.grants[].encryptionContextEquals` | `map<string, string>` |  |  |  |
| `spec.grants[].encryptionContextSubset` | `map<string, string>` |  |  |  |
| `spec.grants[].retireOnDelete` | `bool` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region the key lives in. Ciphertext is regional: data
encrypted under this key is decrypted in this region (use
multi_region + replica keys for cross-region decryption with the
same key material).
Example: "us-west-2", "eu-west-1".

- rule: {"string":{"minLen":"1"}}

### spec.description

`string`

Free-form description shown in the AWS Console. Up to 8192
characters.

- rule: {"string":{"maxLen":"8192"}}

### spec.keySpec

`string`

The cryptographic configuration, create-time immutable:
"SYMMETRIC_DEFAULT" (AES-256-GCM -- the default and what AWS
service integrations require), "RSA_2048"/"RSA_3072"/"RSA_4096"
(asymmetric encrypt/decrypt or sign/verify),
"ECC_NIST_P256"/"ECC_NIST_P384"/"ECC_NIST_P521" (asymmetric
sign/verify or ECDH key agreement),
"ECC_NIST_EDWARDS25519" (Ed25519 sign/verify only),
"ECC_SECG_P256K1" (sign/verify only -- the blockchain curve),
"ML_DSA_44"/"ML_DSA_65"/"ML_DSA_87" (post-quantum ML-DSA
sign/verify only), "HMAC_224"/"HMAC_256"/"HMAC_384"/"HMAC_512"
(MAC generation), or "SM2" (China regions only). Empty keeps the
AWS default (SYMMETRIC_DEFAULT).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["SYMMETRIC_DEFAULT","RSA_2048","RSA_3072","RSA_4096","ECC_NIST_P256","ECC_NIST_P384","ECC_NIST_P521","ECC_NIST_EDWARDS25519","ECC_SECG_P256K1","ML_DSA_44","ML_DSA_65","ML_DSA_87","HMAC_224","HMAC_256","HMAC_384","HMAC_512","SM2"]}}

### spec.keyUsage

`string`

What the key is used for, create-time immutable:
"ENCRYPT_DECRYPT" (the default -- required for SYMMETRIC_DEFAULT
and the only usage AWS service integrations support),
"SIGN_VERIFY" (asymmetric RSA/ECC/ML-DSA/SM2 signing keys),
"KEY_AGREEMENT" (ECDH shared-secret derivation -- NIST ECC curves
and SM2 only), or "GENERATE_VERIFY_MAC" (HMAC keys). Empty keeps
the AWS default (ENCRYPT_DECRYPT). The cross-field rules below
mirror AWS's own key-spec / key-usage compatibility matrix.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["ENCRYPT_DECRYPT","SIGN_VERIFY","KEY_AGREEMENT","GENERATE_VERIFY_MAC"]}}

### spec.policy

`string`

The key policy as a JSON document -- the resource-based policy
that is the root of access control on the key (IAM policies only
work when the key policy delegates to them). Empty keeps the AWS
default policy, which grants the account's root user full access
and enables IAM-policy delegation -- the right choice for most
keys. Set a custom policy for cross-account grants or to restrict
administration; keep the account root (or your admin role) as an
administrator so the key cannot become unmanageable.

### spec.bypassPolicyLockoutSafetyCheck

`bool`

Skip AWS's check that the key policy leaves the calling principal
able to manage the key. Setting this with a policy that locks out
every administrator makes the key PERMANENTLY unmanageable (only
AWS Support can recover it) -- leave false unless you are
deliberately constructing a lockout and understand the blast
radius.

### spec.disabled

`bool`

Create the key in (or flip a live key to) the disabled state --
every cryptographic operation under it fails until re-enabled. An
operational pause switch, gentler than deletion. False (the
default) keeps the key enabled.

### spec.enableKeyRotation

`bool`

Rotate the key material automatically. AWS supports automatic
rotation only for SYMMETRIC_DEFAULT keys with KMS-generated
material; rotation is transparent to callers (old material is
retained to decrypt old ciphertext). Recommended for long-lived
encryption keys. False (the AWS default) never rotates.

### spec.rotationPeriodInDays

`int32`

How often (in days) the material rotates, 90-2560. 0 keeps the
AWS default (365 days). Only meaningful with enable_key_rotation.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"lte":2560,"gte":90}}

### spec.multiRegion

`bool`

Make this a multi-Region PRIMARY key, create-time immutable.
Replica keys in other regions then share its key material --
ciphertext encrypted in one region decrypts in another. This kind
always creates the primary; replicas are created from it and
reference its key_arn.

### spec.deletionWindowDays

`int32`

Waiting period (days) between scheduling deletion and the key
being destroyed, 7-30 -- the recovery window against accidental
deletion (ciphertext under a destroyed key is unrecoverable).
0 keeps the AWS default (30 days).

- default: `30`
- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"lte":30,"gte":7}}

### spec.aliases

`[]string`

Friendly names for the key, each beginning "alias/" (e.g.
"alias/orders-db"). Aliases are how humans and SDK callers
address the key without its generated ID; many aliases may point
at one key, and each materializes as its own alias resource so
list edits add/remove in place. The "alias/aws/" prefix is
reserved for AWS-managed keys.

- rule: {"repeated":{"unique":true,"items":{"string":{"pattern":"^alias/[0-9A-Za-z_/-]+$"}}}}

### spec.customKeyStoreId

`string`

The custom key store to create the key in: a CloudHSM key store
(backed by your CloudHSM cluster) or an external key store (backed
by a key manager outside AWS). Supply a literal key-store id
(cks-...); custom key stores are account-level infrastructure the
catalog does not provision. Create-time immutable. Custom key
store keys must be symmetric encryption keys, never rotate
automatically, and cannot be multi-Region (AWS's contract,
enforced by the rules below). Empty (the default) keeps the key
material in standard AWS KMS.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"22"}}

### spec.xksKeyId

`string`

The id of an existing key in the external key manager, for keys
created in an EXTERNAL key store -- KMS forwards cryptographic
operations under this key to that external key. Requires
custom_key_store_id (pointing at an external key store).
Create-time immutable. Up to 128 characters.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"128"}}

### spec.grants

`[]AwsKmsKeyGrant`

KMS grants on this key: scoped, revocable permissions that let a
principal use the key for specific operations without editing the
key policy -- the mechanism for wiring "this workload role may
encrypt/decrypt under this key" as a first-class dependency, and
for cross-account key usage. Each entry materializes as its own
grant resource; grants are create-time immutable (any change
replaces the grant, which is safe -- grants carry no state).

- rule: encryption_context_equals and encryption_context_subset are mutually exclusive -- a grant takes at most one constraint
- rule: grantee_principal must be an IAM principal ARN (arn:...) -- AWS's CreateGrant rejects bare service principals on this parameter
- rule: retiring_principal must be an IAM principal ARN (arn:...) -- AWS's RetireGrant contract takes IAM principals only

### spec.grants[].name

`string`

Friendly name for the grant, shown by ListGrants next to the
generated grant id. Optional. Up to 256 characters: letters,
digits, and _:/- .

- rule: name may use letters, digits, and _:/- (max 256 characters)

### spec.grants[].granteePrincipal

`string | valueFrom` · required

The IAM principal the grant permits to use the key, in ARN form: an
IAM role (the mainstream pattern -- reference an AwsIamRole's
role_arn output), an IAM user, an account root, or a federated/
assumed-role principal. Cross-account delegation works by naming a
principal in another account. NOT a service principal: AWS's
CreateGrant takes those through a separate GranteeServicePrincipal
parameter (paired with a mandatory SourceArn constraint) that the
Terraform provider does not expose -- a bare service principal here
is rejected live with "InvalidArnException: GranteePrincipal does
not refer to a valid principal" (live-verified 2026-08-11).

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.grants[].operations

`[]string` · required

The KMS operations the grant allows, at least one. The domain is
AWS's own GrantOperation set; which entries make sense depends on
the key's shape (e.g. Sign/Verify need a signing key, GenerateMac
an HMAC key) -- AWS validates that pairing at grant creation.

- rule: {"repeated":{"minItems":"1","unique":true,"items":{"string":{"in":["Decrypt","Encrypt","GenerateDataKey","GenerateDataKeyWithoutPlaintext","ReEncryptFrom","ReEncryptTo","Sign","Verify","GetPublicKey","CreateGrant","RetireGrant","DescribeKey","GenerateDataKeyPair","GenerateDataKeyPairWithoutPlaintext","GenerateMac","VerifyMac","DeriveSharedSecret"]}}}}

### spec.grants[].retiringPrincipal

`string | valueFrom`

The IAM principal (ARN form) allowed to retire the grant when it is
no longer needed, in addition to the key administrators. Reference
an AwsIamRole's role_arn output or pass a literal role/user ARN.
Service principals are not accepted here for the same reason as
grantee_principal (AWS's RetiringServicePrincipal parameter is not
in the provider surface).

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.grants[].encryptionContextEquals

`map<string, string>`

Constrain the grant to requests whose encryption context EQUALS
exactly these key-value pairs. Only valid for operations that take
an encryption context (symmetric keys). Mutually exclusive with
encryption_context_subset.

### spec.grants[].encryptionContextSubset

`map<string, string>`

Constrain the grant to requests whose encryption context CONTAINS
these key-value pairs (a subset match -- the request may carry
more). Mutually exclusive with encryption_context_equals.

### spec.grants[].retireOnDelete

`bool`

How teardown releases the grant: false (the default) REVOKES it --
the hard stop that denies all further use immediately; true RETIRES
it -- the graceful path AWS recommends when the grant's work is
done. Both remove the grant; they differ in intent and in the API
permission they exercise (RevokeGrant vs RetireGrant).

## Validation Rules

- `rotation_only_for_symmetric`: automatic rotation is only supported for SYMMETRIC_DEFAULT keys -- asymmetric and HMAC key material never rotates automatically
- `rotation_period_requires_rotation`: rotation_period_in_days only applies when enable_key_rotation is true
- `symmetric_keys_encrypt_decrypt`: SYMMETRIC_DEFAULT keys only support key_usage ENCRYPT_DECRYPT (the default -- leave key_usage empty)
- `hmac_keys_generate_verify_mac`: HMAC key specs require key_usage GENERATE_VERIFY_MAC, and GENERATE_VERIFY_MAC requires an HMAC key spec
- `ecc_nist_curves_usage`: ECC_NIST_P256/P384/P521 keys require key_usage SIGN_VERIFY or KEY_AGREEMENT
- `ecc_sign_only_curves_usage`: ECC_NIST_EDWARDS25519 and ECC_SECG_P256K1 keys are signing keys -- set key_usage SIGN_VERIFY
- `ml_dsa_keys_sign_verify`: ML-DSA post-quantum keys are signing keys -- set key_usage SIGN_VERIFY
- `key_agreement_key_specs`: key_usage KEY_AGREEMENT is only supported for ECC_NIST_P256/P384/P521 and SM2 keys
- `custom_key_store_symmetric_only`: custom key store keys must be symmetric encryption keys -- leave key_spec and key_usage empty (or SYMMETRIC_DEFAULT / ENCRYPT_DECRYPT)
- `custom_key_store_no_rotation`: automatic rotation is not supported for keys in a custom key store
- `custom_key_store_no_multi_region`: multi-Region keys cannot be created in a custom key store
- `xks_key_requires_custom_key_store`: xks_key_id requires custom_key_store_id (the external key store the key lives in)
- `no_reserved_aws_aliases`: the alias/aws/ prefix is reserved for AWS-managed keys -- choose a different alias name

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsKmsKey, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.key_id` | `string` | The generated key ID (UUID; "mrk-..." for multi-Region keys). |
| `status.outputs.key_arn` | `string` | The key ARN -- the join key encryption-at-rest fields across the catalog reference (databases, queues, buckets, functions, ...). |
| `status.outputs.alias_names` | `[]string` | The alias names attached to the key (each "alias/..."), in spec order -- the human-friendly addresses SDK callers may use instead of the key ID. |
| `status.outputs.grant_ids` | `map<string, string>` | The AWS-generated grant IDs, keyed exactly as the module keys each grant: by the entry's position in spec.grants ("0", "1", ...). Grant identity lives in these IDs, not in the optional friendly name -- RetireGrant/RevokeGrant and state import take them. (The one-time grant TOKEN from creation stays deliberately unexposed: it is a sensitive eventual-consistency bridge, not an identifier.) |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.grants[].granteePrincipal` | AwsIamRole | `status.outputs.role_arn` |
| `spec.grants[].retiringPrincipal` | AwsIamRole | `status.outputs.role_arn` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsAppRunnerService | `spec.kmsKeyArn` | `status.outputs.key_arn` |
| AwsAthenaWorkgroup | `spec.resultConfiguration.kmsKeyArn` | `status.outputs.key_arn` |
| AwsAthenaWorkgroup | `spec.managedQueryResults.kmsKey` | `status.outputs.key_arn` |
| AwsAthenaWorkgroup | `spec.customerContentEncryptionKmsKey` | `status.outputs.key_arn` |
| AwsAthenaWorkgroup | `spec.monitoring.managedLogging.kmsKey` | `status.outputs.key_arn` |
| AwsAthenaWorkgroup | `spec.monitoring.s3Logging.kmsKey` | `status.outputs.key_arn` |
| AwsAuroraDsql | `spec.kmsEncryptionKey` | `status.outputs.key_arn` |
| AwsBackupVault | `spec.standard.kmsKeyArn` | `status.outputs.key_arn` |
| AwsBackupVault | `spec.airGapped.encryptionKeyArn` | `status.outputs.key_arn` |
| AwsBedrockAgent | `spec.customerEncryptionKeyArn` | `status.outputs.key_arn` |
| AwsBedrockAgentCoreEvaluation | `spec.evaluators[].kmsKeyArn` | `status.outputs.key_arn` |
| AwsBedrockAgentCoreGateway | `spec.kmsKeyArn` | `status.outputs.key_arn` |
| AwsBedrockAgentCoreIdentity | `spec.policyEngine.encryptionKeyArn` | `status.outputs.key_arn` |
| AwsBedrockAgentCoreMemory | `spec.encryptionKeyArn` | `status.outputs.key_arn` |
| AwsBedrockAgentCoreTokenVault | `spec.kmsKeyArn` | `status.outputs.key_arn` |
| AwsBedrockCustomModel | `spec.customModelKmsKeyArn` | `status.outputs.key_arn` |
| AwsBedrockFlow | `spec.customerEncryptionKeyArn` | `status.outputs.key_arn` |
| AwsBedrockGuardrail | `spec.kmsKeyArn` | `status.outputs.key_arn` |
| AwsBedrockKnowledgeBase | `spec.managed.kmsKeyArn` | `status.outputs.key_arn` |
| AwsBedrockKnowledgeBase | `spec.dataSources[].kmsKeyArn` | `status.outputs.key_arn` |
| AwsBedrockPrompt | `spec.customerEncryptionKeyArn` | `status.outputs.key_arn` |
| AwsCloudTrail | `spec.kmsKeyId` | `status.outputs.key_arn` |
| AwsCloudTrailEventDataStore | `spec.kmsKeyId` | `status.outputs.key_arn` |
| AwsCloudwatchLogAnomalyDetector | `spec.kmsKeyId` | `status.outputs.key_arn` |
| AwsCloudwatchLogGroup | `spec.kmsKeyId` | `status.outputs.key_arn` |
| AwsCloudwatchSynthetics | `spec.canary.artifactEncryptionKmsKeyArn` | `status.outputs.key_arn` |
| AwsCodeBuildProject | `spec.encryptionKey` | `status.outputs.key_arn` |
| AwsCodePipeline | `spec.artifactStores[].encryptionKeyId` | `status.outputs.key_arn` |
| AwsCognitoUserPool | `spec.lambdaConfig.kmsKeyId` | `status.outputs.key_arn` |
| AwsConfigRecorder | `spec.deliveryChannel.s3KmsKeyArn` | `status.outputs.key_arn` |
| AwsDlmLifecyclePolicy | `spec.customPolicy.schedules[].crossRegionCopyRules[].cmkArn` | `status.outputs.key_arn` |
| AwsDlmLifecyclePolicy | `spec.customPolicy.action.crossRegionCopies[].cmkArn` | `status.outputs.key_arn` |
| AwsDocumentDb | `spec.instances[].performanceInsightsKmsKeyId` | `status.outputs.key_arn` |
| AwsDocumentDb | `spec.kmsKeyId` | `status.outputs.key_arn` |
| AwsDynamodb | `spec.serverSideEncryption.kmsKeyArn` | `status.outputs.key_arn` |
| AwsDynamodb | `spec.replicas[].kmsKeyArn` | `status.outputs.key_arn` |
| AwsEbsSnapshot | `spec.copyFrom.kmsKeyId` | `status.outputs.key_arn` |
| AwsEbsSnapshot | `spec.importFrom.kmsKeyId` | `status.outputs.key_arn` |
| AwsEbsVolume | `spec.kmsKeyId` | `status.outputs.key_arn` |
| AwsEc2Instance | `spec.rootBlockDevice.kmsKeyId` | `status.outputs.key_arn` |
| AwsEc2Instance | `spec.ebsBlockDevices[].kmsKeyId` | `status.outputs.key_arn` |
| AwsEcrRegistrySettings | `spec.repositoryCreationTemplates[].encryption.kmsKey` | `status.outputs.key_arn` |
| AwsEcrRepo | `spec.kmsKeyId` | `status.outputs.key_arn` |
| AwsEcsCluster | `spec.executeCommandConfiguration.kmsKeyId` | `status.outputs.key_arn` |
| AwsEcsCluster | `spec.managedStorageConfiguration.fargateEphemeralStorageKmsKeyId` | `status.outputs.key_arn` |
| AwsEcsCluster | `spec.managedStorageConfiguration.kmsKeyId` | `status.outputs.key_arn` |
| AwsEcsService | `spec.serviceConnect.services[].tls.kmsKey` | `status.outputs.key_arn` |
| AwsEcsService | `spec.volumeConfiguration.managedEbsVolume.kmsKeyId` | `status.outputs.key_arn` |
| AwsEksCluster | `spec.kmsKeyArn` | `status.outputs.key_arn` |
| AwsElasticFileSystem | `spec.kmsKeyId` | `status.outputs.key_arn` |
| AwsElasticFileSystem | `spec.replication.destinationKmsKeyId` | `status.outputs.key_arn` |
| AwsEventBridgeApiDestination | `spec.connection.kmsKeyIdentifier` | `status.outputs.key_arn` |
| AwsEventBridgeBus | `spec.kmsKeyIdentifier` | `status.outputs.key_arn` |
| AwsEventBridgeBus | `spec.archives[].kmsKeyIdentifier` | `status.outputs.key_arn` |
| AwsEventBridgePipe | `spec.kmsKeyIdentifier` | `status.outputs.key_arn` |
| AwsEventBridgeScheduler | `spec.kmsKeyArn` | `status.outputs.key_arn` |
| AwsFsxLustreFileSystem | `spec.kmsKeyId` | `status.outputs.key_arn` |
| AwsFsxOntapFileSystem | `spec.kmsKeyId` | `status.outputs.key_arn` |
| AwsFsxOpenzfsFileSystem | `spec.kmsKeyId` | `status.outputs.key_arn` |
| AwsFsxWindowsFileSystem | `spec.kmsKeyId` | `status.outputs.key_arn` |
| AwsGuardDuty | `spec.publishingDestination.kmsKeyArn` | `status.outputs.key_arn` |
| AwsKinesisFirehose | `spec.sseKmsKeyArn` | `status.outputs.key_arn` |
| AwsKinesisFirehose | `spec.extendedS3.kmsKeyArn` | `status.outputs.key_arn` |
| AwsKinesisFirehose | `spec.extendedS3.s3Backup.kmsKeyArn` | `status.outputs.key_arn` |
| AwsKinesisFirehose | `spec.opensearch.s3Config.kmsKeyArn` | `status.outputs.key_arn` |
| AwsKinesisFirehose | `spec.opensearchServerless.s3Config.kmsKeyArn` | `status.outputs.key_arn` |
| AwsKinesisFirehose | `spec.httpEndpoint.s3Config.kmsKeyArn` | `status.outputs.key_arn` |
| AwsKinesisFirehose | `spec.redshift.s3Config.kmsKeyArn` | `status.outputs.key_arn` |
| AwsKinesisFirehose | `spec.redshift.s3Backup.kmsKeyArn` | `status.outputs.key_arn` |
| AwsKinesisFirehose | `spec.splunk.s3Config.kmsKeyArn` | `status.outputs.key_arn` |
| AwsKinesisFirehose | `spec.snowflake.s3Config.kmsKeyArn` | `status.outputs.key_arn` |
| AwsKinesisFirehose | `spec.iceberg.s3Config.kmsKeyArn` | `status.outputs.key_arn` |
| AwsKinesisStream | `spec.kmsKeyId` | `status.outputs.key_arn` |
| AwsLambda | `spec.sourceKmsKeyArn` | `status.outputs.key_arn` |
| AwsLambda | `spec.kmsKeyArn` | `status.outputs.key_arn` |
| AwsLambdaEventSourceMapping | `spec.kmsKeyArn` | `status.outputs.key_arn` |
| AwsLaunchTemplate | `spec.blockDeviceMappings[].ebs.kmsKeyId` | `status.outputs.key_arn` |
| AwsManagedPrometheus | `spec.kmsKeyArn` | `status.outputs.key_arn` |
| AwsMemorydbCluster | `spec.kmsKeyArn` | `status.outputs.key_arn` |
| AwsMskCluster | `spec.kmsKeyArn` | `status.outputs.key_arn` |
| AwsMwaaEnvironment | `spec.kmsKeyArn` | `status.outputs.key_arn` |
| AwsNeptuneCluster | `spec.kmsKeyId` | `status.outputs.key_arn` |
| AwsOpenSearchDomain | `spec.kmsKeyId` | `status.outputs.key_arn` |
| AwsOpenSearchServerlessCollection | `spec.encryption.kmsKeyArn` | `status.outputs.key_arn` |
| AwsRdsCluster | `spec.instances[].performanceInsightsKmsKeyId` | `status.outputs.key_arn` |
| AwsRdsCluster | `spec.masterUserSecretKmsKeyId` | `status.outputs.key_arn` |
| AwsRdsCluster | `spec.kmsKeyId` | `status.outputs.key_arn` |
| AwsRdsCluster | `spec.performanceInsightsKmsKeyId` | `status.outputs.key_arn` |
| AwsRdsCluster | `spec.activityStream.kmsKeyId` | `status.outputs.key_arn` |
| AwsRdsInstance | `spec.kmsKeyId` | `status.outputs.key_arn` |
| AwsRdsInstance | `spec.masterUserSecretKmsKeyId` | `status.outputs.key_arn` |
| AwsRdsInstance | `spec.performanceInsightsKmsKeyId` | `status.outputs.key_arn` |
| AwsRedisElasticache | `spec.kmsKeyId` | `status.outputs.key_arn` |
| AwsRedshiftCluster | `spec.masterPasswordSecretKmsKeyId` | `status.outputs.key_arn` |
| AwsRedshiftCluster | `spec.kmsKeyId` | `status.outputs.key_arn` |
| AwsRedshiftServerlessNamespace | `spec.adminPasswordSecretKmsKeyId` | `status.outputs.key_arn` |
| AwsRedshiftServerlessNamespace | `spec.kmsKeyId` | `status.outputs.key_arn` |
| AwsRoute53Zone | `spec.dnssec.kmsKeyArn` | `status.outputs.key_arn` |
| AwsS3Bucket | `spec.encryption.kmsKeyId` | `status.outputs.key_arn` |
| AwsS3Bucket | `spec.replication.rules[].destination.replicaKmsKeyId` | `status.outputs.key_arn` |
| AwsS3Bucket | `spec.inventoryConfigurations[].destination.sseKmsKeyId` | `status.outputs.key_arn` |
| AwsS3Bucket | `spec.metadataConfiguration.inventoryTableEncryption.kmsKeyArn` | `status.outputs.key_arn` |
| AwsS3Bucket | `spec.metadataConfiguration.journalTableEncryption.kmsKeyArn` | `status.outputs.key_arn` |
| AwsS3ObjectSet | `spec.objects[].kmsKey` | `status.outputs.key_arn` |
| AwsS3TableBucket | `spec.encryption.kmsKeyArn` | `status.outputs.key_arn` |
| AwsS3TableBucket | `spec.namespaces[].tables[].encryption.kmsKeyArn` | `status.outputs.key_arn` |
| AwsS3VectorBucket | `spec.encryption.kmsKeyArn` | `status.outputs.key_arn` |
| AwsS3VectorBucket | `spec.indexes[].encryption.kmsKeyArn` | `status.outputs.key_arn` |
| AwsSagemakerDomain | `spec.kmsKeyId` | `status.outputs.key_arn` |
| AwsSagemakerDomain | `spec.defaultUserSettings.canvasAppSettings.workspaceSettings.s3KmsKeyId` | `status.outputs.key_arn` |
| AwsSagemakerDomain | `spec.defaultUserSettings.sharingSettings.s3KmsKeyId` | `status.outputs.key_arn` |
| AwsSagemakerDomain | `spec.userProfiles[].userSettings.canvasAppSettings.workspaceSettings.s3KmsKeyId` | `status.outputs.key_arn` |
| AwsSagemakerDomain | `spec.userProfiles[].userSettings.sharingSettings.s3KmsKeyId` | `status.outputs.key_arn` |
| AwsSagemakerEndpoint | `spec.productionVariants[].coreDump.kmsKeyArn` | `status.outputs.key_arn` |
| AwsSagemakerEndpoint | `spec.shadowVariants[].coreDump.kmsKeyArn` | `status.outputs.key_arn` |
| AwsSagemakerEndpoint | `spec.kmsKeyArn` | `status.outputs.key_arn` |
| AwsSagemakerEndpoint | `spec.asyncInference.kmsKeyArn` | `status.outputs.key_arn` |
| AwsSagemakerEndpoint | `spec.dataCapture.kmsKeyArn` | `status.outputs.key_arn` |
| AwsSagemakerFeatureGroup | `spec.onlineStore.kmsKeyArn` | `status.outputs.key_arn` |
| AwsSagemakerFeatureGroup | `spec.offlineStore.kmsKeyArn` | `status.outputs.key_arn` |
| AwsSagemakerNotebookInstance | `spec.kmsKeyArn` | `status.outputs.key_arn` |
| AwsSecretsManagerSecret | `spec.kmsKeyId` | `status.outputs.key_arn` |
| AwsSecretsManagerSecret | `spec.replicaRegions[].kmsKeyId` | `status.outputs.key_arn` |
| AwsServerlessElasticache | `spec.kmsKeyId` | `status.outputs.key_arn` |
| AwsSnsTopic | `spec.kmsKeyId` | `status.outputs.key_arn` |
| AwsSqsQueue | `spec.kmsKeyId` | `status.outputs.key_arn` |
| AwsSsmParameter | `spec.keyId` | `status.outputs.key_arn` |
| AwsStepFunction | `spec.encryption.kmsKeyId` | `status.outputs.key_arn` |
| KubernetesLangfuse | `spec.blobStorage.eventUpload.sse.kmsKeyId` | `status.outputs.key_arn` |
| KubernetesLangfuse | `spec.blobStorage.mediaUpload.sse.kmsKeyId` | `status.outputs.key_arn` |
| KubernetesLangfuse | `spec.blobStorage.batchExport.sse.kmsKeyId` | `status.outputs.key_arn` |

## See Also

- [Overview](../README.md)
