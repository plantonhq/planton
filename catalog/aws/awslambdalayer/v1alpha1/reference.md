# AwsLambdaLayer

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsLambdaLayerSpec defines one Lambda layer version - a shared code
archive (libraries, custom runtimes, config files) that many Lambda
functions attach by ARN - with its cross-account/organization share
grants managed in-line.

A layer VERSION is immutable at AWS: every argument below is fixed
for life, so changing any of them publishes a NEW version (a
replacement in IaC terms). Functions pin the exact version ARN they
were configured with and keep running through a replacement; roll
the consumers to the new version ARN at their own pace. The layer
NAME persists across versions - it is wired from metadata.name.

The archive itself lives in S3 (upload it there first - an
AwsS3Bucket plus your build pipeline, or an AwsS3ObjectSet). Lambda
copies the archive at publish time, so the S3 object only needs to
exist while this resource is being created.

## Example

```yaml
# Canonical AwsLambdaLayer example (hack/dev manifest and refgen
# Example source): a shared Python utilities layer published from an
# S3 archive, shared with one sibling account.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsLambdaLayer
metadata:
  name: shared-python-utils
  id: shared-python-utils
  org: test-org
  env: dev
spec:
  region: us-west-2
  code:
    bucket:
      value: layer-artifacts
    key: layers/shared-python-utils.zip
  sourceCodeHash: 47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=
  description: Shared Python utilities (logging, tracing, config)
  compatibleRuntimes:
    - python3.13
    - python3.12
  compatibleArchitectures:
    - x86_64
    - arm64
  licenseInfo: Apache-2.0
  permissions:
    - statementId: share-tooling-account
      principal: "222233334444"
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.code` | `AwsLambdaLayerS3Code` | yes |  |  |
| `spec.code.bucket` | `string \| valueFrom` | yes |  | AwsS3Bucket (`status.outputs.bucket_id`) |
| `spec.code.key` | `string` | yes |  |  |
| `spec.code.version` | `string` |  |  |  |
| `spec.sourceCodeHash` | `string` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.compatibleRuntimes` | `[]string` |  |  |  |
| `spec.compatibleArchitectures` | `[]string` |  |  |  |
| `spec.licenseInfo` | `string` |  |  |  |
| `spec.skipDestroy` | `bool` |  |  |  |
| `spec.permissions` | `[]AwsLambdaLayerPermission` |  |  |  |
| `spec.permissions[].statementId` | `string` | yes |  |  |
| `spec.permissions[].principal` | `string` |  |  |  |
| `spec.permissions[].organizationId` | `string` |  |  |  |
| `spec.permissions[].skipDestroy` | `bool` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region the layer version is published in. Example:
"us-west-2".

- rule: {"string":{"minLen":"1"}}

### spec.code

`AwsLambdaLayerS3Code` · required

The zip archive holding the layer content, in S3. Layout the
runtimes expect: libraries under python/, nodejs/node_modules/,
java/lib/, etc. (Lambda unpacks the archive into /opt).

- rule: {"required":true}

### spec.code.bucket

`string | valueFrom` · required

The S3 bucket holding the archive. Must be in the layer's region.
Reference an AwsS3Bucket bucket_id output or pass a literal
bucket name.

- references: AwsS3Bucket (`status.outputs.bucket_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsS3Bucket, name: <that resource's name>, fieldPath: status.outputs.bucket_id}} -- a bare string does not parse

### spec.code.key

`string` · required

The object key (path) of the layer content zip.

- rule: {"string":{"minLen":"1"}}

### spec.code.version

`string`

Pin a specific object version (versioned buckets). Empty
publishes from the current version; pair with source_code_hash
for fully deterministic publishes.

### spec.sourceCodeHash

`string`

Base64-encoded SHA256 of the archive. Set it (usually from your
build pipeline) to make content updates declarative: a new hash
publishes a new version, an unchanged hash is a no-op even when
the S3 object is rewritten in place. Leave empty to publish only
when the S3 key or object version changes. (AWS never reports
this value back - it is a local change detector.)

### spec.description

`string`

What this layer provides. AWS caps the description at 256
characters.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"256"}}

### spec.compatibleRuntimes

`[]string`

The runtimes this layer is built for, e.g. "python3.13",
"nodejs22.x", "java21", "provided.al2023". Purely advisory
metadata AWS uses for console filtering and compatibility
warnings - a function with a runtime outside this list can still
attach the layer via the API. AWS caps the list at 15.

- rule: {"repeated":{"maxItems":"15","unique":true}}

### spec.compatibleArchitectures

`[]string`

The instruction-set architectures the layer's binaries support.
Leave empty when the content is architecture-neutral (pure
Python/JS). AWS caps the list at 2 (both architectures).

- rule: {"repeated":{"maxItems":"2","unique":true,"items":{"string":{"in":["x86_64","arm64"]}}}}

### spec.licenseInfo

`string`

The layer content's software license: an SPDX identifier
("MIT", "Apache-2.0"), a URL, or the full text (up to 512
characters).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"512"}}

### spec.skipDestroy

`bool`

Keep the published version in AWS when this resource is destroyed
or replaced (a replacement then leaves the OLD version available
to the functions still pinning it). AWS keeps billing nothing for
dormant layer versions; the trade-off is manual cleanup via
DeleteLayerVersion when versions accumulate.

### spec.permissions

`[]AwsLambdaLayerPermission`

Who else may use this layer version, keyed by statement_id. Each
grant is one statement in the version's resource policy giving
lambda:GetLayerVersion (the only action AWS supports on layers -
the modules pin it).

- rule: organization_id requires principal "*" - AWS narrows the wildcard to the organization; with a specific account id the organization filter has nothing to narrow

### spec.permissions[].statementId

`string` · required

The statement's identity within the version's policy - the
for_each key on both engines and the key in the revision_ids
output map.

- rule: {"string":{"minLen":"1","maxLen":"100","pattern":"^[a-zA-Z0-9-_]+$"}}

### spec.permissions[].principal

`string`

Who may fetch the layer: a 12-digit AWS account id, or "*" for
everyone (scope a wildcard with organization_id to keep it
org-internal).

- rule: principal must be a 12-digit AWS account id or "*"

### spec.permissions[].organizationId

`string`

Restrict a wildcard principal to one AWS Organization
(o-... id). Every account in the organization may then use the
layer; accounts outside it may not.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^o-[a-z0-9]{10,32}$"}}

### spec.permissions[].skipDestroy

`bool`

Keep the grant in AWS when this entry is removed or the resource
is destroyed (the layer-version analog of the spec-level
skip_destroy).

## Validation Rules

- `spec.permission_statement_ids_unique`: permission statement_ids must be unique - each grant is one policy statement keyed by its statement_id

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsLambdaLayer, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.layer_arn` | `string` | The layer's unversioned ARN (arn:aws:lambda:region:account:layer:name) - the identity that persists across versions. |
| `status.outputs.layer_version_arn` | `string` | The published version's ARN (arn:aws:lambda:region:account:layer:name:version) - what functions attach. The chart-ready join key for AwsLambda's layers list. |
| `status.outputs.version` | `string` | The published version number ("1", "2", ...). |
| `status.outputs.code_sha256` | `string` | Base64-encoded SHA256 of the archive as Lambda stored it - the deployed content's own digest. |
| `status.outputs.permission_revision_ids` | `map<string, string>` | Policy revision ids keyed by each grant's statement_id. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.code.bucket` | AwsS3Bucket | `status.outputs.bucket_id` |

## See Also

- [Overview](../README.md)
