# AWS SageMaker MLflow App

Deploys an Amazon SageMaker serverless MLflow app — the MLflow 3.x successor to the hourly-billed tracking server, with nothing to size and per-use billing that costs nothing while idle. Artifacts land in your S3 bucket through the app's IAM role; associating SageMaker domains makes the app the default MLflow for every Studio user in them, and one app per account can hold the account-wide default. Everything updates in place except the role, which is the single replace-on-change field.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **SageMaker MLflow App** — a serverless MLflow 3.x deployment named from `metadata.name`, wired to your S3 artifact store (`artifactStoreUri`) through the app's IAM role, with optional domain associations, account-default status, automatic model registration into the SageMaker Model Registry, and a weekly maintenance window

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with SageMaker control-plane permissions (`sagemaker:CreateMlflowApp` and its siblings). Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- An S3 bucket (or prefix) for the artifact store behind `artifactStoreUri`.
- An IAM role trusting `sagemaker.amazonaws.com` with read/write access to that bucket, wired via `roleArn`.
- For domain associations: the SageMaker domains behind `defaultDomainIds` (only when associating domains).

## Deploy

### Console

Open the deployment store, find **AWS SageMaker MLflow App**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, the region, artifact store, and role, and the optional domain associations. Start from the **Serverless Experiment Tracking** preset in the [Presets](#presets) tab for the standalone shape, or the **Studio Default MLflow** preset to associate a domain.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsSagemakerMlflowApp
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
      name: mlflow-app-role
      fieldPath: status.outputs.role_arn
```

```shell
planton apply -f mlflow-app.yaml
```

This creates a serverless MLflow app storing artifacts under the given S3 prefix through the referenced role — no capacity to size, no idle charge. An Infra Job tracks the provisioning in real time.

### InfraChart

When the app deploys alongside its role and domain in one chart, wire both references via ValueFromRef:

```yaml
spec:
  region: us-east-1
  artifactStoreUri: s3://acme-mlflow/artifacts
  roleArn:
    valueFrom:
      kind: AwsIamRole
      name: mlflow-app-role
      fieldPath: status.outputs.role_arn
  defaultDomainIds:
    - valueFrom:
        kind: AwsSagemakerDomain
        name: data-science-domain
        fieldPath: status.outputs.domain_id
```

The InfraPipeline resolves the dependency graph, creates the role and domain first, then the app associated with them.

## Key Configuration

These are the most important decisions when configuring an MLflow app. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**App or tracking server** — This kind is the serverless MLflow 3.x deployment: billed per use, no idle charge, nothing to size. The classic dedicated-capacity server (AwsSagemakerMlflowServer) bills hourly with roughly 25-minute lifecycle operations. Default to the app for new deployments — per-use billing beats an always-on meter for all but sustained heavy tracking, and there is no capacity to outgrow. The app does NOT attach to a tracking server; it stands alone.

**Get the role right the first time** — `roleArn` is the ONE change that replaces the app; the name, artifact store, domains, registration mode, and maintenance window all update in place. A replacement means a new app ARN, which invalidates anything keyed on the old one.

**Associate domains instead of circulating URIs** — Studio users in `defaultDomainIds` domains track to the app automatically; that is the intended distribution path. Handing tracking URIs around by hand is what domain association exists to replace.

**One account default, ever** — `accountDefaultStatus: ENABLED` is account-global, and only one app per account can hold it. Decide which app owns the default before two manifests fight over it — the second apply will strip the first.

**Auto-registration is a pipeline decision** — `modelRegistrationMode: AutoModelRegistrationEnabled` registers every model logged to MLflow into the SageMaker Model Registry. Convenient for teams standardized on the registry; noisy if experiments log throwaway models — leave it disabled until the registry is the system of record.

**The maintenance window's day token is mixed-case** — `weeklyMaintenanceWindowStart` takes UTC `Ddd:HH:MM` with the day exactly as `Mon`…`Sun`; AWS rejects `SUN:03:00` server-side. The spec enforces the correct shape at manifest time.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsIamRole** | `roleArn` | `status.outputs.role_arn` |
| **AwsSagemakerDomain** | `defaultDomainIds[]` | `status.outputs.domain_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `app_arn` | Amazon Resource Name of the app — the identity every MLflow API operation keys on | Client tracking configuration; IAM policies scoping MLflow access |

`app_name` is also exported, but it updates in place and the APIs key on the ARN — treat the name as display metadata, not a composition input.

## Common Patterns

**Standalone team tracker** — one app per team with its own artifact prefix and role, no domain association: data scientists point their MLflow clients at the app directly. The zero-idle-cost shape for teams that track from anywhere, not just Studio. Start from the **Serverless Experiment Tracking** preset.

**Studio-default MLflow** — the app associated with a SageMaker domain via `defaultDomainIds`, so every Studio user in the domain tracks to it without configuration. Pair with `accountDefaultStatus: ENABLED` on exactly one app when the whole account should share a tracker. Start from the **Studio Default MLflow** preset.

**Registry-integrated tracking** — auto-registration on, so models logged from experiments flow straight into the SageMaker Model Registry for approval workflows. Worth it once the registry gates deployments; premature while experiments still log throwaways.

## Works With

- [**AWS IAM Role**](/infra-catalog/aws-iam-role) — the artifact-store access role, wired via `roleArn`
- [**AWS S3 Bucket**](/infra-catalog/aws-s3-bucket) — the artifact store behind `artifactStoreUri`
- [**AWS SageMaker Domain**](/infra-catalog/aws-sagemaker-domain) — domains whose Studio users default to this app, wired via `defaultDomainIds`
- [**AWS SageMaker Model Registry**](/infra-catalog/aws-sagemaker-model-registry) — where auto-registered models land when `modelRegistrationMode` is enabled
