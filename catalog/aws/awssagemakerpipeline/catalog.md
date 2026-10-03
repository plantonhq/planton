# AWS SageMaker Pipeline

Deploys an Amazon SageMaker pipeline — the ML workflow DAG of processing, training, evaluation, and registration steps that executions run against, declared in SageMaker's pipeline-definition JSON and validated server-side at create. The definition comes from exactly one place: inline as structured YAML in the manifest, or an S3 object (optionally pinned to a version). Creating a pipeline is free — only executions bill — and everything except the name updates in place, so iterate on the DAG freely.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **SageMaker Pipeline** — named from `metadata.name`, with a Studio display name (the modules reuse the pipeline name when `displayName` is omitted, since the provider requires one), the execution role, the definition from its single declared source, and an optional default cap on parallel step execution (`parallelismMaxSteps`)

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with SageMaker control-plane permissions (`sagemaker:CreatePipeline` and its siblings). Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- An IAM role trusting `sagemaker.amazonaws.com` that can run the pipeline's steps — training jobs, processing jobs, model registration — wired via `roleArn`.
- For the S3 arm: the bucket and object holding the definition JSON (only for `definitionS3Location`).

## Deploy

### Console

Open the deployment store, find **AWS SageMaker Pipeline**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, the region and execution role, and the definition source. Start from the **Inline Training Pipeline** preset in the [Presets](#presets) tab to carry the definition in the manifest, or the **S3 Definition Pipeline** preset when definitions are published to S3.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsSagemakerPipeline
metadata:
  name: churn-training
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  description: Nightly churn-model training and registration
  roleArn:
    valueFrom:
      kind: AwsIamRole
      name: pipeline-execution-role
      fieldPath: status.outputs.role_arn
  definition:
    Version: "2020-12-01"
    Metadata: {}
    Parameters: []
    Steps:
      - Name: ReplaceMe
        Type: Fail
        Arguments:
          ErrorMessage: placeholder definition - replace with pipeline.definition() output
```

```shell
planton apply -f sagemaker-pipeline.yaml
```

This creates the pipeline shell with a placeholder single-step DAG (a Fail step is a legal one-node graph) — paste your SageMaker Python SDK `pipeline.definition()` output over it before starting real executions. An Infra Job tracks the provisioning in real time.

### InfraChart

When the pipeline deploys alongside its execution role and definition bucket in one chart, wire the references via ValueFromRef:

```yaml
spec:
  region: us-east-1
  roleArn:
    valueFrom:
      kind: AwsIamRole
      name: pipeline-execution-role
      fieldPath: status.outputs.role_arn
  definitionS3Location:
    bucket:
      valueFrom:
        kind: AwsS3Bucket
        name: ml-definitions
        fieldPath: status.outputs.bucket_id
    objectKey: pipelines/churn-training.json
```

The InfraPipeline resolves the dependency graph, creates the role and bucket first, then the pipeline reading its definition from them.

## Key Configuration

These are the most important decisions when configuring a pipeline. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Generate definitions; don't hand-write them** — The SageMaker Python SDK's `pipeline.definition()` produces the definition JSON. Commit its output into the manifest (or upload it to S3) rather than authoring the schema by hand — the schema is an SDK output format, not an authoring surface.

**Inline beats S3 when the definition fits** — Inline, the definition lives in the manifest as structured, diffable YAML and drifts visibly. On the S3 arm, AWS's describe API returns only the RESOLVED definition, never the S3 location — so a changed S3 object is invisible drift that refresh cannot see. If you must use S3, pin `versionId` and treat every definition change as a manifest change.

**A green apply is the validity claim** — AWS validates the step graph server-side at create and on every definition update: a malformed DAG fails the apply, not the first 2 AM execution.

**The declarative boundary stops at executions** — This resource owns the DAG; executions start against `pipeline_name` imperatively — from schedules, event rules, or SDK calls. Nothing here schedules runs, by design.

**The role runs everything** — Pipeline executions assume `roleArn` for every step: training jobs, processing jobs, model registration. Underscoped, executions fail mid-DAG; overscoped, every step inherits the excess. Scope it to exactly the step types the definition contains.

**Cap parallelism deliberately** — `parallelismMaxSteps` is a default cap across this pipeline's executions. Set it when parallel steps contend for instance quotas — a fan-out of training steps can exhaust a per-type quota and fail the run partway.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsIamRole** | `roleArn` | `status.outputs.role_arn` |
| **AwsS3Bucket** | `definitionS3Location.bucket` | `status.outputs.bucket_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `pipeline_name` | The pipeline's AWS identity | StartPipelineExecution calls from schedules, event rules, and SDK code |
| `pipeline_arn` | Amazon Resource Name of the pipeline | EventBridge rule targets; IAM policies scoping who may start executions |

## Common Patterns

**Inline-definition pipeline** — the `pipeline.definition()` output committed straight into the manifest, so the DAG and its infrastructure review together and drift is impossible. The right default whenever the definition fits comfortably in YAML. Start from the **Inline Training Pipeline** preset.

**S3-published definitions** — a CI job uploads the definition object and the manifest pins its `versionId`; useful when definitions are large or produced by a build system. The version pin is what keeps the S3 blind spot honest. Start from the **S3 Definition Pipeline** preset.

**Train-and-register workflow** — a DAG ending in a model-registration step that lands versions in a SageMaker Model Registry group, from which deployments promote approved packages. The pipeline is the producer half of registry-gated deployment.

## Works With

- [**AWS IAM Role**](/infra-catalog/aws-iam-role) — the execution role every step runs as, wired via `roleArn`
- [**AWS S3 Bucket**](/infra-catalog/aws-s3-bucket) — holds the definition object on the S3 arm, wired via `definitionS3Location.bucket`
- [**AWS SageMaker Model Registry**](/infra-catalog/aws-sagemaker-model-registry) — where registration steps land versioned model packages
- [**AWS SageMaker Feature Group**](/infra-catalog/aws-sagemaker-feature-group) — feature stores pipelines ingest into and build training datasets from
