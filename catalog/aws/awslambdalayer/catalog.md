# AWS Lambda Layer

Deploys a Lambda layer version — a shared code archive (libraries, custom runtimes, config files) that any number of Lambda functions attach by ARN instead of bundling into every deployment package, with cross-account and organization share grants managed in-line. A layer version is immutable at AWS: every spec field is fixed for life, so any change publishes a new version with a new ARN while functions keep the version they pinned. The layer name persists across versions and is wired from `metadata.name`.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Lambda layer version** — published from the S3 archive named in `code`, with its runtime and architecture compatibility metadata and license info. Lambda copies the archive at publish, so the S3 object only needs to exist during the deploy
- **Layer version permissions** — one per `permissions` entry: a statement in the version's resource policy granting `lambda:GetLayerVersion` (the only action AWS supports on layers) to a specific account, everyone, or everyone in one AWS Organization

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with Lambda permissions. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- **The layer content zip in S3** — in a bucket in the same region, laid out the way the runtimes expect (`python/`, `nodejs/node_modules/`, `java/lib/`, ...); Lambda unpacks the archive into `/opt`. Stage it with your build pipeline, an AwsS3Bucket, or an AwsS3ObjectSet.
- **For organization-wide sharing** — your AWS Organization's `o-...` id, from the AWS Organizations console.

## Deploy

### Console

Open the deployment store, find **AWS Lambda Layer**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, then the layer: the S3 archive location, compatibility metadata, and share grants. Start from the **Python Shared Utilities** preset in the [Presets](#presets) tab for the everyday team layer, or the **Organization-Shared Runtime Layer** preset when a platform team publishes for every member account.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsLambdaLayer
metadata:
  name: shared-python-utils
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  code:
    bucket:
      valueFrom:
        kind: AwsS3Bucket
        name: build-artifacts
        fieldPath: status.outputs.bucket_id
    key: layers/shared-python-utils.zip
  description: Shared Python utilities (logging, tracing, config)
  compatibleRuntimes:
    - python3.13
    - python3.12
  compatibleArchitectures:
    - x86_64
    - arm64
  licenseInfo: Apache-2.0
```

```shell
planton apply -f lambda-layer.yaml
```

This publishes version 1 of the `shared-python-utils` layer from the referenced artifact bucket, tagged compatible with both architectures so Graviton and x86 functions filter it correctly in the console. An Infra Job tracks the provisioning in real time.

### InfraChart

When the layer deploys alongside its artifact bucket in one chart, wire the bucket via ValueFromRef:

```yaml
spec:
  region: us-east-1
  code:
    bucket:
      valueFrom:
        kind: AwsS3Bucket
        name: build-artifacts
        fieldPath: status.outputs.bucket_id
    key: layers/shared-python-utils.zip
  compatibleRuntimes:
    - python3.13
```

The InfraPipeline resolves the dependency graph, deploys the bucket first, then publishes the layer from it.

## Key Configuration

These are the most important decisions when configuring a layer. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**A "change" is always a new version — plan rollout, not update** — Every spec field is fixed for life at AWS; applying any change publishes a new version with a new ARN. Consumers do not follow automatically: each function keeps the version it was configured with until its own next deploy points at the new ARN. Treat a layer change like a library release — publish, then roll consumers at their own pace.

**sourceCodeHash is your change detector, not AWS's** — AWS never reports the hash back; it exists to make content updates declarative. Set it from your build pipeline (`base64(sha256(zip))`): a new hash publishes a new version even when the S3 key stays the same. Without it, rewriting the object in place is invisible to the modules — only a changed key or object version triggers a publish.

**skipDestroy trades cleanup for continuity** — With `skipDestroy`, replaced versions stay available in AWS so functions pinning them keep working through a rollout, and dormant versions bill nothing — but nothing deletes them either. Sweep old versions with `aws lambda delete-layer-version` once consumers have moved on. The per-grant `skipDestroy` does the same for individual share statements.

**The compatibility lists are advisory** — `compatibleRuntimes` and `compatibleArchitectures` drive console filtering and warnings only; the API attaches any layer to any function. They are documentation for humans, not a compatibility wall — real compatibility is decided by what your archive contains.

**Keep share grants few and coarse** — Each `permissions` entry is one policy statement keyed by `statementId`. For organization-wide sharing, one wildcard principal scoped by `organizationId` beats a pile of per-account grants — not just for tidiness: the provider's permission importer reads only the first policy statement, so versions carrying several grants round-trip lossily on import.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsS3Bucket** | `code.bucket` | `status.outputs.bucket_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `layer_version_arn` | The published version's ARN (`...:layer:name:version`) | The `layers` list on AwsLambda functions — what consumers attach |
| `layer_arn` | The unversioned layer ARN — the identity persisting across versions | IAM policies and tooling that address the layer independent of version |

`version` (the published version number), `code_sha256` (the archive digest as Lambda stored it), and `permission_revision_ids` are also exported for audit and change verification rather than composition.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Team shared-utilities layer** — Common logging, tracing, and config-loading code published once and attached by every function, declared for both architectures. The trade against vendoring the code per function: one publish updates the library everywhere, but every consumer must redeploy to pick it up — version drift across functions is normal, not a bug. Start from the **Python Shared Utilities** preset.

**Organization-wide platform layer** — A platform team publishes for the whole AWS Organization: the wildcard principal scoped by `organizationId` lets every member account attach the layer, and `skipDestroy` keeps replaced versions alive so consumer functions never break mid-rollout. Start from the **Organization-Shared Runtime Layer** preset.

**Pipeline-driven content-addressed publishes** — The build pipeline stages the zip and stamps `sourceCodeHash`; re-applying with an unchanged hash is a no-op even when the S3 object was rewritten, and a changed hash publishes deterministically. The right shape once layer content changes more often than the manifest.

## Works With

- [**AWS Lambda**](/infra-catalog/aws-lambda) — the consumer: functions attach the `layer_version_arn` output in their layers list
- [**AWS S3 Bucket**](/infra-catalog/aws-s3-bucket) — the artifact bucket the layer content zip is published from, wired via `code.bucket`
- [**AWS S3 Object Set**](/infra-catalog/aws-s3-object-set) — stages the layer zip into the bucket when a build pipeline doesn't
