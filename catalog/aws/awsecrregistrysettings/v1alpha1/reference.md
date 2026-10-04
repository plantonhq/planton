# AwsEcrRegistrySettings

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsEcrRegistrySettingsSpec defines the REGISTRY-level ECR
configuration for one AWS region: the registry permissions policy,
scanning configuration, cross-region/cross-account replication,
pull-through cache rules, repository creation templates, account
settings, and pull-time update exclusions. Individual repositories
are the AwsEcrRepo kind; everything here governs the registry all
repositories in the region share.

This is a SETTINGS SINGLETON: AWS keeps exactly one private
registry per account+region, and this component manages its
registry-wide posture. Deploy at most one instance per region -
two instances targeting the same region fight over the same
registry objects. metadata.name never reaches AWS (identity tags
aside) - it is Planton-side identity only.

Destroy semantics DIFFER per arm (each taught on its arm below):
the registry policy and the keyed collections (cache rules,
creation templates, exclusions) genuinely delete; scanning and
replication RESET to their empty defaults; account settings
PERSIST at their last-applied values.

## Example

```yaml
# Canonical AwsEcrRegistrySettings example (hack/dev manifest and
# refgen Example source): enhanced scanning, a same-account
# cross-region replication rule, a credential-less pull-through cache
# of registry.k8s.io with its creation template, account settings,
# and one pull-time update exclusion for a CI role.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsEcrRegistrySettings
metadata:
  name: use-west-2-registry
  id: use-west-2-registry
  org: test-org
  env: dev
spec:
  region: us-west-2
  scanning:
    scanType: ENHANCED
    rules:
      - scanFrequency: CONTINUOUS_SCAN
        filters:
          - "prod-*"
      - scanFrequency: SCAN_ON_PUSH
        filters:
          - "*"
  replicationRules:
    - destinations:
        - region: us-east-1
          registryId: "123456789012"
      repositoryFilters:
        - prod-
  pullThroughCacheRules:
    - ecrRepositoryPrefix: k8s
      upstreamRegistryUrl: registry.k8s.io
  repositoryCreationTemplates:
    - prefix: k8s
      description: Repositories created by the registry.k8s.io cache
      appliedFor:
        - PULL_THROUGH_CACHE
      imageTagMutability: IMMUTABLE
  accountSettings:
    basicScanTypeVersion: AWS_NATIVE
    registryPolicyScope: V2
  pullTimeUpdateExclusions:
    - value: arn:aws:iam::123456789012:role/ci-push-role
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.registryPolicy` | `string` |  |  |  |
| `spec.scanning` | `AwsEcrRegistryScanning` |  |  |  |
| `spec.scanning.scanType` | `string` |  |  |  |
| `spec.scanning.rules` | `[]AwsEcrRegistryScanningRule` |  |  |  |
| `spec.scanning.rules[].scanFrequency` | `string` |  |  |  |
| `spec.scanning.rules[].filters` | `[]string` | yes |  |  |
| `spec.replicationRules` | `[]AwsEcrReplicationRule` |  |  |  |
| `spec.replicationRules[].destinations` | `[]AwsEcrReplicationDestination` | yes |  |  |
| `spec.replicationRules[].destinations[].region` | `string` | yes |  |  |
| `spec.replicationRules[].destinations[].registryId` | `string` |  |  |  |
| `spec.replicationRules[].repositoryFilters` | `[]string` |  |  |  |
| `spec.pullThroughCacheRules` | `[]AwsEcrPullThroughCacheRule` |  |  |  |
| `spec.pullThroughCacheRules[].ecrRepositoryPrefix` | `string` | yes |  |  |
| `spec.pullThroughCacheRules[].upstreamRegistryUrl` | `string` | yes |  |  |
| `spec.pullThroughCacheRules[].upstreamRepositoryPrefix` | `string` |  |  |  |
| `spec.pullThroughCacheRules[].credentialArn` | `string \| valueFrom` |  |  | AwsSecretsManagerSecret (`status.outputs.secret_arn`) |
| `spec.pullThroughCacheRules[].customRoleArn` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.repositoryCreationTemplates` | `[]AwsEcrRepositoryCreationTemplate` |  |  |  |
| `spec.repositoryCreationTemplates[].prefix` | `string` | yes |  |  |
| `spec.repositoryCreationTemplates[].description` | `string` |  |  |  |
| `spec.repositoryCreationTemplates[].appliedFor` | `[]string` | yes |  |  |
| `spec.repositoryCreationTemplates[].customRoleArn` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.repositoryCreationTemplates[].imageTagMutability` | `string` |  |  |  |
| `spec.repositoryCreationTemplates[].imageTagMutabilityExclusionFilters` | `[]string` |  |  |  |
| `spec.repositoryCreationTemplates[].encryption` | `AwsEcrTemplateEncryption` |  |  |  |
| `spec.repositoryCreationTemplates[].encryption.type` | `string` |  |  |  |
| `spec.repositoryCreationTemplates[].encryption.kmsKey` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.repositoryCreationTemplates[].lifecyclePolicy` | `string` |  |  |  |
| `spec.repositoryCreationTemplates[].repositoryPolicy` | `string` |  |  |  |
| `spec.repositoryCreationTemplates[].resourceTags` | `map<string, string>` |  |  |  |
| `spec.accountSettings` | `AwsEcrAccountSettings` |  |  |  |
| `spec.accountSettings.basicScanTypeVersion` | `string` |  |  |  |
| `spec.accountSettings.blobMounting` | `bool` |  |  |  |
| `spec.accountSettings.registryPolicyScope` | `string` |  |  |  |
| `spec.pullTimeUpdateExclusions` | `[]string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |

## Field Details

### spec.region

`string` · required

The AWS region whose registry this instance manages. The region
IS the resource identity - one instance per region. Example:
"us-west-2".

- rule: {"string":{"minLen":"1"}}

### spec.registryPolicy

`string`

The registry permissions policy (JSON) - the IAM resource policy
that grants OTHER accounts registry-level actions: replication
into this registry (ecr:ReplicateImage) and pull-through cache
sharing. Repository-level grants belong on AwsEcrRepo's
repository policy instead. Destroying this arm deletes the
policy.

- rule: registry_policy must be a valid JSON policy document
- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.scanning

`AwsEcrRegistryScanning`

How images in this registry are scanned for vulnerabilities.
Destroying this arm RESETS the registry to BASIC scanning with
no rules (AWS has no delete - the modules put the empty default
back).

- rule: scan_frequency CONTINUOUS_SCAN requires scan_type ENHANCED - basic scanning only scans on push or manually

### spec.scanning.scanType

`string`

The scanning engine: "BASIC" (the built-in scanner, OS packages
only, free) or "ENHANCED" (Amazon Inspector - OS + language
packages, continuous re-scanning, Inspector pricing applies).

- rule: {"string":{"in":["BASIC","ENHANCED"]}}

### spec.scanning.rules

`[]AwsEcrRegistryScanningRule`

Which repositories are scanned, and when. No matching rule means
no automatic scanning for that repository. AWS caps the rules at
100.

- rule: {"repeated":{"maxItems":"100"}}

### spec.scanning.rules[].scanFrequency

`string`

When matched repositories are scanned: "SCAN_ON_PUSH" (each
pushed image once), "CONTINUOUS_SCAN" (re-scan as new CVEs
publish; ENHANCED only), or "MANUAL" (only on demand).

- rule: {"string":{"in":["SCAN_ON_PUSH","CONTINUOUS_SCAN","MANUAL"]}}

### spec.scanning.rules[].filters

`[]string` · required

Repository name patterns this rule matches - "*" for every
repository, "prod-*" for a prefix. (AWS wildcard filters:
lowercase letters, digits, dot, underscore, hyphen, slash, and
"*".)

- rule: {"repeated":{"minItems":"1","unique":true,"items":{"string":{"minLen":"1","maxLen":"256","pattern":"^[0-9a-z*](?:[0-9a-z_./*-]?[0-9a-z*]+)*$"}}}}

### spec.replicationRules

`[]AwsEcrReplicationRule`

Where this registry replicates images, evaluated in order.
Cross-account destinations also need the DESTINATION registry's
policy to allow ecr:ReplicateImage from this account. AWS caps
the rules at 10. Destroying this arm RESETS replication to none
(AWS has no delete - the modules put the empty rule set back).

- rule: {"repeated":{"maxItems":"10"}}

### spec.replicationRules[].destinations

`[]AwsEcrReplicationDestination` · required

Where matched images are copied. AWS caps a rule at 25
destinations.

- rule: {"repeated":{"minItems":"1","maxItems":"25"}}

### spec.replicationRules[].destinations[].region

`string` · required

The destination region ("us-east-1"). Same-account cross-region,
or cross-account with registry_id.

- rule: {"string":{"minLen":"1","pattern":"^[a-z]{2}(-[a-z]+)+-\\d$"}}

### spec.replicationRules[].destinations[].registryId

`string`

The destination registry's AWS account id. Your own account id
for cross-region replication; another account for cross-account
(that registry's policy must allow it).

- rule: {"string":{"pattern":"^[0-9]{12}$"}}

### spec.replicationRules[].repositoryFilters

`[]string`

Replicate only repositories whose names start with these
prefixes. Empty replicates EVERY repository in the registry. AWS
caps the filters at 100.

- rule: {"repeated":{"maxItems":"100","unique":true,"items":{"string":{"minLen":"1","maxLen":"256"}}}}

### spec.pullThroughCacheRules

`[]AwsEcrPullThroughCacheRule`

Pull-through cache rules, keyed by ecr_repository_prefix: pulls
of "{prefix}/..." transparently fetch from the upstream registry
and cache the image here. Destroying an entry deletes the rule
(already-cached repositories and images remain).

- rule: set at most one of credential_arn (external upstreams - Docker Hub, GitHub, Kubernetes, ...) and custom_role_arn (cross-account ECR upstreams)

### spec.pullThroughCacheRules[].ecrRepositoryPrefix

`string` · required

The prefix pulls are cached under - "docker.io/library/nginx"
pulled as "{prefix}/library/nginx". "ROOT" caches at the
registry root (no prefix). The for_each key on both engines and
the rule's import ID. Fixed for life.

- rule: {"string":{"minLen":"2","maxLen":"30","pattern":"^((?:[a-z0-9]+(?:[._-][a-z0-9]+)*/)*[a-z0-9]+(?:[._-][a-z0-9]+)*/?|ROOT)$"}}

### spec.pullThroughCacheRules[].upstreamRegistryUrl

`string` · required

The upstream registry's URL, e.g.
"registry-1.docker.io" (Docker Hub), "public.ecr.aws",
"ghcr.io", "registry.k8s.io", "quay.io", or another account's
"<account>.dkr.ecr.<region>.amazonaws.com". Fixed for life.

- rule: {"string":{"minLen":"1"}}

### spec.pullThroughCacheRules[].upstreamRepositoryPrefix

`string`

Re-root the upstream namespace: cache
"{upstream_repository_prefix}/..." instead of mirroring the
upstream's full paths. AWS requires 2-30 characters. Fixed for
life. (No min_len here - a zero-length floor keeps the field
OPTIONAL for the variables generator; AWS enforces the 2-char
minimum server-side.)

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"30","pattern":"^((?:[a-z0-9]+(?:[._-][a-z0-9]+)*/)*[a-z0-9]+(?:[._-][a-z0-9]+)*/?|ROOT)$"}}

### spec.pullThroughCacheRules[].credentialArn

`string | valueFrom`

The Secrets Manager secret holding the upstream's credentials
(for authenticated upstreams like Docker Hub; the secret name
must start with "ecr-pullthroughcache/"). Reference an
AwsSecretsManagerSecret secret_arn output or pass a literal ARN.
NOTE: once set, clearing this back to empty is not propagated by
the provider - replace the rule to drop credentials.

- references: AwsSecretsManagerSecret (`status.outputs.secret_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecretsManagerSecret, name: <that resource's name>, fieldPath: status.outputs.secret_arn}} -- a bare string does not parse

### spec.pullThroughCacheRules[].customRoleArn

`string | valueFrom`

The IAM role this registry assumes to pull from a cross-account
ECR upstream. Reference an AwsIamRole role_arn output or pass a
literal ARN. The same clearing caveat as credential_arn applies.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.repositoryCreationTemplates

`[]AwsEcrRepositoryCreationTemplate`

Templates applied to repositories the registry creates on your
behalf (replication, pull-through cache, or create-on-push),
keyed by prefix. Without a matching template such repositories
are created with bare defaults. Destroying an entry deletes the
template (repositories it already stamped keep their settings).

- rule: image_tag_mutability_exclusion_filters requires image_tag_mutability IMMUTABLE_WITH_EXCLUSION or MUTABLE_WITH_EXCLUSION
- rule: an image_tag_mutability exclusion filter can contain at most 2 wildcards (*)

### spec.repositoryCreationTemplates[].prefix

`string` · required

The repository-name prefix this template stamps - "ROOT" applies
to every auto-created repository without a more specific match.
The for_each key on both engines and the template's import ID.
Fixed for life.

- rule: {"string":{"minLen":"2","maxLen":"256","pattern":"^(ROOT|(?:[a-z0-9]+(?:[._-][a-z0-9]+)*/)*[a-z0-9]+(?:[._-][a-z0-9]+)*)$"}}

### spec.repositoryCreationTemplates[].description

`string`

What this template is for.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"255"}}

### spec.repositoryCreationTemplates[].appliedFor

`[]string` · required

Which auto-creation paths use this template: "REPLICATION"
(repositories created to receive replicated images),
"PULL_THROUGH_CACHE" (repositories created by cache pulls), or
"CREATE_ON_PUSH" (repositories created by a first push).

- rule: {"repeated":{"minItems":"1","unique":true,"items":{"string":{"in":["REPLICATION","PULL_THROUGH_CACHE","CREATE_ON_PUSH"]}}}}

### spec.repositoryCreationTemplates[].customRoleArn

`string | valueFrom`

The IAM role ECR assumes when creating repositories from this
template. Reference an AwsIamRole role_arn output or pass a
literal ARN.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.repositoryCreationTemplates[].imageTagMutability

`string`

Tag mutability for stamped repositories: "MUTABLE" (AWS's
default), "IMMUTABLE", or the *_WITH_EXCLUSION modes that carry
per-filter exceptions.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["MUTABLE","IMMUTABLE","IMMUTABLE_WITH_EXCLUSION","MUTABLE_WITH_EXCLUSION"]}}

### spec.repositoryCreationTemplates[].imageTagMutabilityExclusionFilters

`[]string`

Tag patterns exempted from the mutability mode (e.g. "latest",
"dev-*"). Wildcard patterns, at most 2 "*" each, AWS caps the
list at 5.

- rule: {"repeated":{"maxItems":"5","unique":true,"items":{"string":{"minLen":"1","maxLen":"128","pattern":"^[a-zA-Z0-9._*-]+$"}}}}

### spec.repositoryCreationTemplates[].encryption

`AwsEcrTemplateEncryption`

How stamped repositories encrypt images at rest. Omit for AES256
(the AWS default).

- rule: kms_key applies only when type is KMS or KMS_DSSE - AES256 uses an Amazon-managed key

### spec.repositoryCreationTemplates[].encryption.type

`string`

The encryption mode: "AES256" (Amazon-managed, the default),
"KMS" (your key, single-layer), or "KMS_DSSE" (your key,
dual-layer).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["AES256","KMS","KMS_DSSE"]}}

### spec.repositoryCreationTemplates[].encryption.kmsKey

`string | valueFrom`

The KMS key for KMS/KMS_DSSE modes. Reference an AwsKmsKey
key_arn output or pass a literal key ARN. Empty with a KMS mode
uses the AWS-managed aws/ecr key.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.repositoryCreationTemplates[].lifecyclePolicy

`string`

The lifecycle policy (JSON) stamped onto created repositories -
image expiration rules.

- rule: lifecycle_policy must be a valid JSON lifecycle policy document
- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.repositoryCreationTemplates[].repositoryPolicy

`string`

The repository permissions policy (JSON) stamped onto created
repositories.

- rule: repository_policy must be a valid JSON policy document
- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.repositoryCreationTemplates[].resourceTags

`map<string, string>`

AWS tags stamped onto created repositories (these are the
repositories' tags, not this template's).

- rule: {"map":{"keys":{"string":{"minLen":"1","maxLen":"128"}},"values":{"string":{"maxLen":"256"}}}}

### spec.accountSettings

`AwsEcrAccountSettings`

Account-level ECR feature toggles. These PERSIST after destroy at
their last-applied values (AWS has no delete or reset for them) -
to change a value back, apply the desired value before
destroying.

### spec.accountSettings.basicScanTypeVersion

`string`

Which engine BASIC scanning uses: "AWS_NATIVE" (the current
Amazon scanner) or "CLAIR" (the legacy scanner, on its way out).
Unset leaves the account's current setting.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["AWS_NATIVE","CLAIR"]}}

### spec.accountSettings.blobMounting

`bool` · optional (explicit presence)

Whether layer blob mounting between repositories is enabled
(faster pushes of images sharing base layers). Unset leaves the
account's current setting.

### spec.accountSettings.registryPolicyScope

`string`

The registry policy permission scope: "V2" (current semantics -
registry policies grant only registry-level actions) or "V1"
(legacy). Unset leaves the account's current setting.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["V1","V2"]}}

### spec.pullTimeUpdateExclusions

`[]string | valueFrom`

IAM principals whose image pushes do NOT refresh pull-time
metrics, keyed by principal ARN - keeps automation (CI/CD roles,
replication roles) from masking real image usage in lifecycle
policies that expire by days-since-last-pull. Reference
AwsIamRole role_arn outputs or pass literal principal ARNs.
Destroying an entry deregisters the exclusion.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

## Validation Rules

- `spec.at_least_one_arm`: configure at least one arm (registry_policy / scanning / replication_rules / pull_through_cache_rules / repository_creation_templates / account_settings / pull_time_update_exclusions) - an instance managing nothing is dead configuration
- `spec.cache_rule_prefixes_unique`: pull_through_cache_rules prefixes must be unique - each rule owns one repository prefix in the registry
- `spec.creation_template_prefixes_unique`: repository_creation_templates prefixes must be unique - each template owns one repository prefix in the registry

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsEcrRegistrySettings, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.registry_id` | `string` | The registry id - the account's 12-digit id, and the import ID for the policy/scanning/replication singletons. |
| `status.outputs.registry_url` | `string` | The registry's pull URL base, "{account}.dkr.ecr.{region}.amazonaws.com" - the chart-ready prefix cached upstream images pull through. |
| `status.outputs.pull_through_cache_rule_registry_ids` | `map<string, string>` | Cache-rule registry ids keyed by ecr_repository_prefix (each rule's import ID is its prefix; the map's keys enumerate the rules that exist). |
| `status.outputs.repository_creation_template_registry_ids` | `map<string, string>` | Creation-template registry ids keyed by prefix (each template's import ID is its prefix). |
| `status.outputs.pull_time_update_exclusion_arns` | `map<string, string>` | Pull-time update exclusions keyed by the RESOLVED principal ARN (the exclusion's import ID is the ARN itself). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.pullThroughCacheRules[].credentialArn` | AwsSecretsManagerSecret | `status.outputs.secret_arn` |
| `spec.pullThroughCacheRules[].customRoleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.repositoryCreationTemplates[].customRoleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.repositoryCreationTemplates[].encryption.kmsKey` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.pullTimeUpdateExclusions` | AwsIamRole | `status.outputs.role_arn` |

## See Also

- [Overview](../README.md)
