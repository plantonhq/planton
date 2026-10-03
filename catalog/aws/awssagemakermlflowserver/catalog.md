# AWS SageMaker MLflow Server

Deploys an Amazon SageMaker MLflow tracking server — the classic managed MLflow deployment for experiments, runs, and model tracking, sized `Small`, `Medium`, or `Large` and billed hourly from the moment it reaches Created. Artifacts land in your S3 bucket through the server's IAM role; size resizes in place, while the role and the MLflow version pin are replace-on-change. Lifecycle operations are slow by design — creation and deletion each run roughly 17–25 minutes — so plan replacements as half-hour-plus events. For per-use billing with no idle charge, the serverless successor is the AWS SageMaker MLflow App.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **SageMaker MLflow Tracking Server** — a dedicated-capacity MLflow deployment named from `metadata.name`, wired to your S3 artifact store (`artifactStoreUri`) through the server's IAM role, with the chosen size, optional `mlflowVersion` pin, automatic model registration into the SageMaker Model Registry, and a weekly maintenance window

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with SageMaker control-plane permissions (`sagemaker:CreateMlflowTrackingServer` and its siblings). Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- An S3 bucket (or prefix) for the artifact store behind `artifactStoreUri`.
- An IAM role trusting `sagemaker.amazonaws.com` with read/write access to that bucket, wired via `roleArn`.

## Deploy

### Console

Open the deployment store, find **AWS SageMaker MLflow Server**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, the region, artifact store, role, and size. Start from the **Team Experiment Tracking** preset in the [Presets](#presets) tab for a single team's `Small` server, or the **Regulated Experiment Tracking** preset when every variable must be pinned. Budget roughly 25 minutes for the server to reach Created.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsSagemakerMlflowServer
metadata:
  name: team-mlflow
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  artifactStoreUri: s3://acme-mlflow/artifacts
  roleArn:
    valueFrom:
      kind: AwsIamRole
      name: mlflow-server-role
      fieldPath: status.outputs.role_arn
  size: Small
```

```shell
planton apply -f mlflow-server.yaml
```

This creates a `Small` tracking server (up to ~25 users) storing artifacts under the given S3 prefix — hourly billing starts when it reaches Created. An Infra Job tracks the provisioning in real time.

### InfraChart

When the server deploys alongside its artifact-store role in one chart, wire the role reference via ValueFromRef:

```yaml
spec:
  region: us-east-1
  artifactStoreUri: s3://acme-mlflow/artifacts
  roleArn:
    valueFrom:
      kind: AwsIamRole
      name: mlflow-server-role
      fieldPath: status.outputs.role_arn
  size: Small
```

The InfraPipeline resolves the dependency graph, creates the role first, then the server that assumes it.

## Key Configuration

These are the most important decisions when configuring a tracking server. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Server or serverless app** — This server bills hourly around the clock whether anyone logs a run or not, and every lifecycle operation takes ~17–25 minutes. If the team tracks intermittently, the serverless AwsSagemakerMlflowApp — per-use billing, nothing while idle — is the better fit. Choose the server for sustained heavy tracking where dedicated capacity earns its meter.

**Start Small; resize in place** — `size` sets the hourly rate (`Small` ~25 users, `Medium` ~50, `Large` ~100) and upgrades are a maintenance-window style in-place operation, not a replacement. There is no penalty for starting at Small and growing.

**Two replace-on-change fields, both half-hour events** — `roleArn` and `mlflowVersion` replace the server when changed, and a replacement is a delete plus a create at ~17–25 minutes each. Get the role right up front, and pin the version only when governance demands it.

**Pin `major.minor`, never a patch** — AWS normalizes `mlflowVersion` to `major.minor`, so a `3.0.1` pin would drift forever; the spec rejects patch-level values at manifest time. Omitted means AWS picks the latest.

**Treat auto-registration as one-way** — The provider cannot turn `automaticModelRegistration` back OFF: a true-to-false change is silently not transmitted. Disabling means replacing the server (~50 minutes of lifecycle) or an out-of-band API call. Enable it only once the Model Registry is truly the system of record.

**Set the maintenance window deliberately** — `weeklyMaintenanceWindowStart` (UTC `Ddd:HH:MM`) puts resizes and patching in your quiet hours; omitted lets AWS pick. The day token is mixed-case — AWS accepts `Tue:03:30` and rejects `TUE:03:30` — and the spec validates the exact API regex so a manifest never learns this from a failed deploy.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsIamRole** | `roleArn` | `status.outputs.role_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `tracking_server_url` | URL of the MLflow UI — also the tracking URI training code points at | `MLFLOW_TRACKING_URI` in training jobs and pipelines |
| `tracking_server_arn` | Amazon Resource Name of the server | IAM policies scoping MLflow access |
| `tracking_server_name` | The server's AWS identity | CLI and API operations addressing the server |

## Common Patterns

**Single-team tracker** — one `Small` server per ML team with its own artifact prefix and role. The size resizes in place as the team grows, so start at the bottom of the ladder. Start from the **Team Experiment Tracking** preset.

**Governed ML platform** — every variable pinned: an explicit `mlflowVersion`, auto-registration into the Model Registry, and a deliberate maintenance window in quiet hours. The shape for platforms where a surprise MLflow upgrade or an unregistered model is an audit finding. Start from the **Regulated Experiment Tracking** preset.

**Delete when idle** — because the meter runs from Created onward, ephemeral research efforts should delete the server between engagements rather than letting it idle. Remember deletion is itself a ~17–25 minute operation, and the S3 artifacts survive it.

## Works With

- [**AWS IAM Role**](/infra-catalog/aws-iam-role) — the artifact-store access role, wired via `roleArn`
- [**AWS S3 Bucket**](/infra-catalog/aws-s3-bucket) — the artifact store behind `artifactStoreUri`
- [**AWS SageMaker MLflow App**](/infra-catalog/aws-sagemaker-mlflow-app) — the serverless successor for intermittent tracking with no idle charge
- [**AWS SageMaker Model Registry**](/infra-catalog/aws-sagemaker-model-registry) — where auto-registered models land when `automaticModelRegistration` is on
