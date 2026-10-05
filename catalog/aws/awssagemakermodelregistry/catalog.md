# AWS SageMaker Model Registry

Deploys an Amazon SageMaker model package group — the model registry's unit of organization, the named shell your training pipelines register versioned model packages into and your deployments promote approved packages out of. The group itself is deliberately thin: versions register into it imperatively (pipelines, SDK calls), never through this resource, and everything except the optional cross-account resource policy is create-time only — even a description edit replaces the group. Creating a group costs nothing.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **SageMaker Model Package Group** — named from `metadata.name`, with an optional description that is part of the group's create-time identity
- **Model Package Group Policy** — created only when `resourcePolicy` is set; the IAM resource policy granting other accounts access to the group (cross-account model sharing). Removing the block from the spec deletes the policy, closing the group to its own account.

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with SageMaker control-plane permissions (`sagemaker:CreateModelPackageGroup` and its siblings). Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- Nothing beyond the connection — the group is free-standing.
- For cross-account sharing: the account IDs you will name as principals in `resourcePolicy` (only for that feature).

## Deploy

### Console

Open the deployment store, find **AWS SageMaker Model Registry**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, the region, and the description. Start from the **Team Model Registry** preset in the [Presets](#presets) tab for a single team's group, or the **Shared Model Registry** preset for the cross-account shape.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsSagemakerModelRegistry
metadata:
  name: churn-models
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  description: Versioned churn-prediction models registered by the training pipeline
```

```shell
planton apply -f model-registry.yaml
```

This creates the model package group; training pipelines register versioned packages into it by name from the moment it exists. An Infra Job tracks the provisioning in real time.

## Key Configuration

These are the most important decisions when configuring a model registry group. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Write the description once, well** — The provider marks `description` replace-on-change: an edit replaces the group, severing it from every version registered so far. Treat the description as part of the group's identity and finalize it before the first model version lands.

**The shelf, not the books** — Model package versions have no fields here by design: registering versions is an imperative, high-frequency act belonging to training pipelines and SDK calls, not declarative infrastructure. This resource guarantees the shelf exists with the right name, description, and sharing posture.

**Iterate on sharing through the policy** — `resourcePolicy` is the one arm that updates in place. Add and remove cross-account principals freely without touching the group; removing the block entirely deletes the policy and closes the group to its own account.

**Grant the minimum verbs** — Read-side consumers need `sagemaker:DescribeModelPackage` and `sagemaker:ListModelPackages`; only accounts that register models INTO your group need `sagemaker:CreateModelPackage`. A policy that grants create to a consumer account invites version pollution from outside your pipeline.

**Organize generously** — Groups are free, so give each model family its own group. One group per family keeps approval workflows and lineage legible; a catch-all group makes every listing a search problem.

## Outputs and Dependencies

### What This Kind Consumes

This kind has no foreign key dependencies — the spec's only inputs are the region, a description, and a policy document; cross-account principals travel as account IDs inside the policy JSON.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `model_package_group_name` | The group's AWS identity | Training pipelines name it as the registration target for new model packages |
| `model_package_group_arn` | Amazon Resource Name of the group | Cross-account policy statements; model package ARNs deployed via a SageMaker model's `modelPackageArn` live under it |

## Common Patterns

**One group per model family** — a private group whose description names the model's purpose, created before the training pipeline first runs. The pipeline registers versions; humans or automation approve them; deployments reference approved packages. Start from the **Team Model Registry** preset.

**Cross-account promotion registry** — the training account owns the group and grants the production account describe/list through `resourcePolicy`; production deploys approved packages by ARN without ever holding write access. Start from the **Shared Model Registry** preset.

**Registry-gated deployment** — pair the group with SageMaker models that deploy via `modelPackageArn` instead of raw images, so nothing reaches an endpoint that didn't pass the registry's approval status. The registry becomes the promotion gate, not just a catalog.

## Works With

- [**AWS SageMaker Model**](/infra-catalog/aws-sagemaker-model) — deploys versioned packages registered in this group via `modelPackageArn`
- [**AWS SageMaker Pipeline**](/infra-catalog/aws-sagemaker-pipeline) — the training pipelines that register model package versions into the group
- [**AWS SageMaker MLflow App**](/infra-catalog/aws-sagemaker-mlflow-app) — auto-registers models logged to MLflow into the registry when enabled
