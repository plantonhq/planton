# AWS CloudWatch Synthetics

Deploys CloudWatch Synthetics resources: a canary — a scheduled scripted probe that hits the health endpoint, walks the checkout flow, or screenshots the page from a Synthetics-managed Lambda, writing run artifacts to S3 — and/or owned groups, the console containers that aggregate canary results into a fleet view. The two arms deploy independently: a canary instance monitors an endpoint and joins groups by name; a groups-only instance owns shared groups for many canaries. The canary's name is `metadata.name` (lowercase letters, digits, hyphens, underscores), and renaming replaces the canary.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Synthetics Canary** — created only when the `canary` arm is set: the probe with its S3-staged code bundle, runtime, schedule (with retries), per-run sizing, optional VPC placement, artifact location and encryption, and retention windows. With `startCanary: true` the schedule starts after create and updates.
- **Synthetics Groups** — one per `groups` entry, each a name-and-tags container this instance owns. Groups shared by many canaries belong in one owning instance.
- **Group Associations** — one per `groupNames` entry, joining this canary to a group by name — owned groups from this spec or groups that exist elsewhere.
- **AWS Tags** — resource metadata tags applied automatically to the canary and groups (the association join is untaggable at AWS).

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with credentials for the target AWS account, carrying Synthetics, Lambda, S3, and IAM pass-role permissions. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.

### AWS Account

- **An execution role** — trusting `lambda.amazonaws.com`, with the canary permissions AWS documents: `s3:PutObject` on the artifact bucket, logs, and `cloudwatch:PutMetricData`. Reference an AwsIamRole's `role_arn` output.
- **An artifact bucket and a code bucket** — often the same bucket; reference an AwsS3Bucket. The code zip must already sit in S3 in the runtime's layout (see Key Configuration) — local-path uploads are deliberately not modeled; an AwsS3ObjectSet carries small bundles inline as base64.
- **Subnets and security groups** (only for VPC-placed canaries) — the private-endpoint probing shape.

## Deploy

### Console

Open the deployment store, find **AWS CloudWatch Synthetics**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and the canary and groups arms. Start from the **Heartbeat Canary** or **Shared Canary Groups** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsCloudwatchSynthetics
metadata:
  name: checkout-heartbeat
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  canary:
    artifactBucket:
      value: acme-canary-artifacts
    artifactPrefix: canary/checkout
    executionRoleArn:
      value: arn:aws:iam::123456789012:role/canary-exec
    handler: heartbeat.handler
    runtimeVersion: syn-nodejs-puppeteer-9.1
    code:
      s3Bucket:
        value: acme-canary-code
      s3Key: canaries/heartbeat.zip
    schedule:
      expression: rate(5 minutes)
      maxRetries: 1
    runConfig:
      memoryInMb: 960
      timeoutInSeconds: 60
      environmentVariables:
        TARGET_URL: https://checkout.example.com/health
    failureRetentionPeriod: 31
    successRetentionPeriod: 7
    startCanary: true
    deleteLambda: true
```

```shell
planton apply -f synthetics-canary.yaml
```

This creates a five-minute heartbeat canary with one retry, trimmed success retention, and a clean-teardown posture, running immediately. A Stack Job tracks the provisioning in real time.

### InfraChart

When the canary deploys alongside its role and buckets in one chart, wire the references via ValueFromRef:

```yaml
spec:
  region: us-east-1
  canary:
    artifactBucket:
      valueFrom:
        kind: AwsS3Bucket
        name: canary-artifacts
        fieldPath: status.outputs.bucket_id
    executionRoleArn:
      valueFrom:
        kind: AwsIamRole
        name: canary-exec
        fieldPath: status.outputs.role_arn
    handler: heartbeat.handler
    runtimeVersion: syn-nodejs-puppeteer-9.1
    code:
      s3Bucket:
        valueFrom:
          kind: AwsS3Bucket
          name: canary-code
          fieldPath: status.outputs.bucket_id
      s3Key: canaries/heartbeat.zip
    schedule:
      expression: rate(5 minutes)
    startCanary: false
```

The InfraPipeline resolves the dependency graph, deploys the buckets and role first, then provisions the canary against them.

## Key Configuration

These are the most important decisions when configuring a Synthetics deployment. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The zip layout is the runtime's contract** — Node.js runtimes require `nodejs/node_modules/<fileName>.js` inside the zip with `handler` set to `<fileName>.handler`; Python runtimes use `python/<fileName>.py`. A wrong layout creates a canary that lands in CREATE_FAILED — and AWS's only repair is delete-and-recreate, which the provider performs automatically.

**startCanary is the cost lever** — a READY canary costs nothing; runs are what bill, so cost scales linearly with schedule frequency. Keep `startCanary: false` in preproduction manifests and flip it in place when monitoring should begin — the provider calls StartCanary and StopCanary around updates.

**Never put secrets in environment variables** — `runConfig.environmentVariables` land in the Synthetics-managed Lambda, AWS never returns them on reads (write-only), and they surface in the Lambda console. Scripts that need credentials should read Secrets Manager or Parameter Store at run time under the execution role.

**Retention trims artifact cost** — every run writes screenshots, HAR files, and logs to S3, and the 31-day default on success artifacts accumulates fast on frequent schedules. `successRetentionPeriod: 7` with failures kept at 31 is the common production posture.

**New script versions ship through S3** — upload a new zip (or object version) and update `code.s3Key` or `s3Version`; the provider stops and starts the canary around the in-place update. Runtime upgrades (`runtimeVersion`) are also in-place — AWS deprecates runtimes on a published schedule, so plan periodic bumps.

**VPC placement is for private endpoints** — set `vpcConfig` only when the probe target is unreachable from the public internet; the canary's Lambda then attaches to your subnets and security groups, and probing public endpoints from inside a VPC needs an egress path.

**Teardown leaves a Lambda layer behind** — `deleteLambda: true` removes the Synthetics-managed Lambda function on destroy, but AWS's DeleteCanary never cleans up the `cwsyn-*` Lambda layer versions it built from the script. A dormant layer version costs nothing; delete the versions by hand if you audit the account after decommissioning canaries.

**Group joins are name-based on purpose** — the association joins a canary ARN to a group name, so shared groups live in one owning instance (or pre-exist) and every canary instance joins by literal name. No instance ever fights another over the group object.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsS3Bucket** | `canary.artifactBucket` | `status.outputs.bucket_id` |
| **AwsS3Bucket** | `canary.code.s3Bucket` | `status.outputs.bucket_id` |
| **AwsIamRole** | `canary.executionRoleArn` | `status.outputs.role_arn` |
| **AwsKmsKey** (SSE_KMS artifacts) | `canary.artifactEncryptionKmsKeyArn` | `status.outputs.key_arn` |
| **AwsSubnet** (VPC placement) | `canary.vpcConfig.subnetIds[]` | `status.outputs.subnet_id` |
| **AwsSecurityGroup** (VPC placement) | `canary.vpcConfig.securityGroupIds[]` | `status.outputs.security_group_id` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `canary_name` | The canary's name (the provider's import ID) | The `CanaryName` dimension in CloudWatch alarms on the canary's SuccessPercent and Duration metrics |
| `canary_arn` | The canary's ARN | IAM policy scoping |
| `engine_arn` | ARN of the Synthetics-managed Lambda behind the canary | Locating the run logs and function when debugging |

The remaining outputs are records rather than composition inputs: `source_location_arn` points at the staged code, `canary_status` echoes the lifecycle state after apply, and `group_arns` / `group_ids` identify the owned groups (joins use group names, never these).

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Heartbeat canary** — a five-minute health probe with one retry: the script reads `TARGET_URL` from its environment and fails the run when the endpoint does. Trimmed success retention keeps artifact cost down; `deleteLambda` keeps teardown clean. Start from the **Heartbeat Canary** preset.

**Shared canary groups** — the groups-only shape: one instance owns the fleet's console groups, and every canary instance joins them by name (`groupNames: [prod-critical-flows]`). The CloudWatch console then shows the fleet's run results on one screen. Start from the **Shared Canary Groups** preset.

**Alarm on the canary, not in it** — the canary emits SuccessPercent and Duration metrics per run; pair it with an AwsCloudwatchAlarm on those metrics so failures page the on-call. The script's job is to probe honestly; the alarm's job is to decide when it matters.

## Works With

- [**AWS S3 Bucket**](/cloud-catalog/aws-s3-bucket) — holds the code zip and receives every run's artifacts
- [**AWS S3 Object Set**](/cloud-catalog/aws-s3-object-set) — stages small canary code bundles inline as base64
- [**AWS IAM Role**](/cloud-catalog/aws-iam-role) — the execution role the canary's Lambda runs under
- [**AWS CloudWatch Alarm**](/cloud-catalog/aws-cloudwatch-alarm) — pages on the canary's SuccessPercent and Duration metrics
- [**AWS KMS Key**](/cloud-catalog/aws-kms-key) — customer-managed encryption for run artifacts (SSE_KMS)
- [**AWS Subnet**](/cloud-catalog/aws-subnet) / [**AWS Security Group**](/cloud-catalog/aws-security-group) — network placement for canaries probing private endpoints
