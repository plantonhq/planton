# AWS SageMaker Model

Deploys an Amazon SageMaker model — the immutable serving definition (container image or model package, S3 artifacts, execution role, networking) that endpoints, batch transform jobs, and inference components deploy. A model is either a single `primaryContainer` or an inference pipeline of 2–15 `containers` run serially or addressed by hostname — exactly one form. Every field is create-time only: any change replaces the model, which is AWS's own contract — roll a new model and repoint the endpoint. The model itself costs nothing to keep; billing starts when something deploys it.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **SageMaker Model** — named from `metadata.name`, carrying the container definition(s) with per-container artifact wiring (compressed `modelDataUrl`, or uncompressed `modelDataSource` with gated-model EULA acceptance), adapter channels, MultiModel serving mode, private-registry image configuration, and optional VPC attachment with full network isolation

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with SageMaker control-plane permissions (`sagemaker:CreateModel` and its siblings). Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- An IAM role trusting `sagemaker.amazonaws.com` with ECR pull on the image and S3 read on the artifacts, wired via `executionRoleArn`. SageMaker assumes it at deploy time, so a missing grant surfaces as an endpoint failure, not a model-create failure.
- Model artifacts staged in S3: a `.tar.gz` for the compressed form, a prefix for uncompressed or MultiModel serving.
- For VPC-attached models: the subnets and security groups behind `vpcConfig` (only for private serving).

## Deploy

### Console

Open the deployment store, find **AWS SageMaker Model**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, the region and execution role, and the container definition. Start from the **Prebuilt Framework Model** preset in the [Presets](#presets) tab to serve an artifact on AWS's own framework image, or the **Inference Pipeline** preset for the chained-container shape.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsSagemakerModel
metadata:
  name: churn-model
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  executionRoleArn:
    valueFrom:
      kind: AwsIamRole
      name: sagemaker-execution-role
      fieldPath: status.outputs.role_arn
  primaryContainer:
    image: 246618743249.dkr.ecr.us-west-2.amazonaws.com/sagemaker-scikit-learn:1.2-1
    modelDataUrl: s3://acme-models/churn/model.tar.gz
    environment:
      SAGEMAKER_PROGRAM: inference.py
```

```shell
planton apply -f sagemaker-model.yaml
```

This creates a single-container model serving a scikit-learn artifact on AWS's prebuilt framework image — deployable by any endpoint variant that references it. An Infra Job tracks the provisioning in real time.

### InfraChart

When the model deploys alongside its execution role in one chart, wire the role reference via ValueFromRef:

```yaml
spec:
  region: us-west-2
  executionRoleArn:
    valueFrom:
      kind: AwsIamRole
      name: sagemaker-execution-role
      fieldPath: status.outputs.role_arn
  primaryContainer:
    image: 246618743249.dkr.ecr.us-west-2.amazonaws.com/sagemaker-scikit-learn:1.2-1
    modelDataUrl: s3://acme-models/churn/model.tar.gz
```

The InfraPipeline resolves the dependency graph, creates the role first, then the model that assumes it.

## Key Configuration

These are the most important decisions when configuring a model. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Treat every change as a replacement** — All fields are create-time only; any spec change replaces the model. That is AWS's intended flow: roll a new model, repoint the endpoint, and delete the old one once nothing references it. Since models cost nothing to keep, retain prior versions as instant rollback targets.

**One container or a pipeline** — `primaryContainer` is the common form; `containers` (2–15) builds an inference pipeline where `inferenceExecutionMode: Serial` feeds each container's output to the next, or `Direct` lets callers address a container by hostname. The spec enforces exactly one form at manifest time — AWS's CreateModel contract is not schema-enforced upstream.

**Compressed or uncompressed artifacts** — `modelDataUrl` (a `.tar.gz`) is the classic form; `modelDataSource` with `compressionType: None` on an `S3Prefix` skips the tarball extraction that dominates load time for multi-GB models, and it is the only form that accepts a gated model's EULA (`acceptEula: true`). At most one of the two per container.

**Prebuilt framework images live in per-region ACCOUNTS** — The scikit-learn/XGBoost registry account for us-west-2 is 246618743249; us-west-1's is 746614075791. Using another region's account fails CreateModel with a misleading 400 claiming the repository "does not grant … to sagemaker.amazonaws.com service principal" — that error means wrong account or nonexistent repo, not a policy you can fix. Derive the account from the sagemaker-python-sdk image URI configuration for your region.

**Network isolation is absolute** — `enableNetworkIsolation` blocks ALL inbound and outbound calls from the container, even to AWS services; artifacts and images still load normally. Combine with `vpcConfig` for private serving, and expect containers that phone home (license checks, telemetry) to break under it.

**MultiModel for fleets of small models** — `mode: MultiModel` lets one endpoint serve many models loaded on demand from the artifact prefix, trading cold-load latency on first invocation for a fraction of the instance cost. `multiModelCache` tunes the trade and exists only in this mode.

**A model must have something to load** — A container with an image but no artifacts passes CreateModel yet can never answer an endpoint's health checks; the endpoint parks at Failed. Wire `modelDataUrl` or `modelDataSource` unless the image truly embeds its weights.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsIamRole** | `executionRoleArn` | `status.outputs.role_arn` |
| **AwsSubnet** | `vpcConfig.subnetIds[]` | `status.outputs.subnet_id` |
| **AwsSecurityGroup** | `vpcConfig.securityGroupIds[]` | `status.outputs.security_group_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `model_name` | The model's AWS identity | Endpoint production variants wire their `model` reference to it |
| `model_arn` | Amazon Resource Name of the model | IAM policies scoping model access |

## Common Patterns

**Prebuilt framework serving** — a trained artifact on AWS's own framework image (scikit-learn, XGBoost, PyTorch), with `SAGEMAKER_PROGRAM` naming your inference script. No container build, no registry to run — the fastest path from artifact to endpoint. Start from the **Prebuilt Framework Model** preset.

**Serial inference pipeline** — 2–15 containers chained so preprocessing, prediction, and postprocessing each live in their own image, deployed and scaled as one unit. The trade: one slow stage gates the whole chain, and every stage replaces together. Start from the **Inference Pipeline** preset.

**Registry-governed deployment** — `modelPackageArn` instead of a raw image deploys a versioned, approved package from the SageMaker Model Registry: the shape for teams whose promotion flow runs through registry approvals rather than image tags.

**Large-model uncompressed loading** — `modelDataSource` on an `S3Prefix` with `compressionType: None`, plus `additionalModelDataSources` channels for adapters or speculative-decoding draft models mounted under their own paths.

## Works With

- [**AWS SageMaker Endpoint**](/infra-catalog/aws-sagemaker-endpoint) — deploys the model behind real-time variants, wired via `model_name`
- [**AWS IAM Role**](/infra-catalog/aws-iam-role) — the execution role SageMaker assumes, wired via `executionRoleArn`
- [**AWS SageMaker Model Registry**](/infra-catalog/aws-sagemaker-model-registry) — the source of versioned model packages deployed via `modelPackageArn`
- [**AWS S3 Bucket**](/infra-catalog/aws-s3-bucket) — where model artifacts live, referenced by URI
- [**AWS ECR Repository**](/infra-catalog/aws-ecr-repo) — where custom inference images live, referenced by registry path in `image`
- [**AWS Subnet**](/infra-catalog/aws-subnet) and [**AWS Security Group**](/infra-catalog/aws-security-group) — VPC attachment for private serving, wired via `vpcConfig`
