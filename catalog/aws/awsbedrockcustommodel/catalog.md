# AWS Bedrock Custom Model

Starts an Amazon Bedrock model-customization job — fine-tuning, continued pre-training, distillation, or reinforcement fine-tuning — that trains a foundation model on your own data and produces a custom model wired into the rest of your AI stack by reference. The deploy starts the job and returns; training runs asynchronously in AWS (minutes to many hours depending on data size and model) and bills per token processed. Every spec field is create-time-immutable: a customization job cannot be altered once started, so any change destroys the job record and the custom model and starts a new, separately billed run.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Model Customization Job** — the training run itself, on the base model named by `baseModelArn`, reading `trainingDataS3Uri` and writing metrics and logs to `outputDataS3Uri`; validation datasets and VPC-scoped data access are configured only when `validationDataS3Uris` / `vpcConfig` are set
- **Custom Model** — the trained model the job produces on completion, encrypted under your KMS key when `customModelKmsKeyArn` is set (a Bedrock-managed key otherwise), exported as `custom_model_arn`

The kind IS the job: destroying this component deletes the model and the job record, and a delete while the job is still running stops it first.

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with Bedrock model-customization permissions (`bedrock:CreateModelCustomizationJob` and its read/stop/delete siblings, plus `iam:PassRole` on the job role). Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- **A region that supports model customization** — a subset of Bedrock regions; us-east-1 has the widest base-model coverage. The base model must be eligible for the chosen customization type.
- **An IAM role trusting `bedrock.amazonaws.com`** — referenced by `roleArn`, with read access to the training and validation S3 locations, write access to the output location, and KMS permissions when the buckets or the model use customer keys. This role is the most common failure point.
- **Training data staged in S3** — in the format the customization type requires (JSONL prompt/completion pairs for fine-tuning; plain domain text for continued pre-training).

## Deploy

### Console

Open the deployment store, find **AWS Bedrock Custom Model**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and the spec: region and base model, the job role, hyperparameters, and the S3 data locations. Start from the **Fine-Tune Minimal** preset in the [Presets](#presets) tab — one epoch on a small dataset to prove the pipeline before the real run.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsBedrockCustomModel
metadata:
  name: support-titan-ft
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  baseModelArn: arn:aws:bedrock:us-east-1::foundation-model/amazon.titan-text-lite-v1
  hyperparameters:
    epochCount: "1"
  roleArn:
    valueFrom:
      kind: AwsIamRole
      name: bedrock-customization
      fieldPath: status.outputs.role_arn
  trainingDataS3Uri: s3://acme-training-data/support/train.jsonl
  outputDataS3Uri: s3://acme-training-data/support/output/
```

```shell
planton apply -f custom-model.yaml
```

This starts a one-epoch fine-tuning job on Titan Text Lite; the deploy returns while training continues and `job_status` reports its progress. An Infra Job tracks the provisioning in real time.

### InfraChart

When the customization job deploys alongside its role in one chart, wire the reference via ValueFromRef:

```yaml
spec:
  region: us-east-1
  baseModelArn: arn:aws:bedrock:us-east-1::foundation-model/amazon.titan-text-lite-v1
  hyperparameters:
    epochCount: "1"
  roleArn:
    valueFrom:
      kind: AwsIamRole
      name: bedrock-customization
      fieldPath: status.outputs.role_arn
  trainingDataS3Uri: s3://acme-training-data/support/train.jsonl
  outputDataS3Uri: s3://acme-training-data/support/output/
```

The InfraPipeline resolves the dependency graph, deploys the role first, then starts the customization job with it.

## Key Configuration

These are the most important decisions when configuring a custom model. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Every field is a one-way door** — CreateModelCustomizationJob has no update: changing anything replaces the job and the model, and the replacement is a fresh, fully billed training run. Get the data format, role, and hyperparameters right on a cheap validation run first — one epoch, minimal dataset, smallest eligible base model.

**Job names never come back** — AWS reserves job names per account forever, even after deletion. `jobName` defaults to `metadata.name`, which serves the first run; any destroy/recreate needs an explicitly fresh `jobName` or AWS rejects the create.

**The customization type dictates the data** — FINE_TUNING (the default) trains on labeled prompt/completion JSONL; CONTINUED_PRE_TRAINING takes unlabeled domain text; DISTILLATION transfers capability from a larger teacher; REINFORCEMENT_FINE_TUNING trains against a reward signal. The base model must be eligible for the chosen type, and the training file format follows the type — a format mismatch fails after the job starts, not at apply.

**Hyperparameters are per-base-model strings** — `epochCount`, `batchSize`, and `learningRate` are the common keys, but the legal set and ranges belong to each base model, and AWS validates them server-side. `epochCount` is also the primary cost lever: cost scales with tokens processed times epochs.

**The role must reach all three S3 locations** — training, validation, and output, plus KMS when any of them (or the model) uses customer keys. Role misconfiguration is the dominant failure mode; the provider retries through IAM propagation delays, but a genuinely missing grant fails the job.

**Watch `job_status`, not apply success** — the deploy returns with the job InProgress; only Completed means a usable model. A Failed job's detail lives in the job record and the output S3 location. Deleting while InProgress stops the job (with a 2-hour delete timeout) — prefer letting jobs finish or fail before destroying.

**Private data needs `vpcConfig`** — when training data is reachable only through private networking, set both `subnetIds` and `securityGroupIds` (required together) to run the job's data access inside your VPC. Omitted, the job uses Bedrock's managed networking.

**A finished model still cannot serve** — custom models serve through Provisioned Throughput (or on-demand custom-model deployment where AWS supports it). Plan the AwsBedrockProvisionedThroughput purchase referencing `custom_model_arn` as part of the same rollout.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsIamRole** | `roleArn` | `status.outputs.role_arn` |
| **AwsKmsKey** | `customModelKmsKeyArn` | `status.outputs.key_arn` |
| **AwsSubnet** | `vpcConfig.subnetIds[]` | `status.outputs.subnet_id` |
| **AwsSecurityGroup** | `vpcConfig.securityGroupIds[]` | `status.outputs.security_group_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `custom_model_arn` | ARN of the resulting custom model | An AwsBedrockProvisionedThroughput purchase's `modelArn` — the step that makes the model servable |
| `job_status` | Job status at the end of the deploy (InProgress, Completed, Failed, Stopping, Stopped) | Deployment pipelines that gate the throughput purchase on Completed |

`custom_model_name` echoes `metadata.name`, and `job_arn` is the job record's identifier — both are audit trails rather than composition inputs.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Validate cheap, then train for real** — a one-epoch run on a minimal dataset against the smallest eligible base model proves the data format and role permissions before the expensive run. Because the spec is immutable, the real run is a second component (or a recreate with a fresh `jobName`), not an edit. Start from the **Fine-Tune Minimal** preset.

**Fine-tune on private data** — the same fine-tuning shape with `vpcConfig` pinning data access to your subnets and security groups, for training sets that never traverse public networking. The trade: you now own the network path — the subnets need a route to S3 (a gateway endpoint is the usual answer). Start from the **Fine-Tune in a Private VPC** preset.

**Fine-tune, buy capacity, serve** — this component produces the model, an AwsBedrockProvisionedThroughput referencing `custom_model_arn` makes it servable, and agents or applications consume the throughput's ARN. In a chart, the reference wiring orders all three.

## Works With

- [**AWS IAM Role**](/infra-catalog/aws-iam-role) — the job role Bedrock assumes, wired via `roleArn`
- [**AWS Bedrock Provisioned Throughput**](/infra-catalog/aws-bedrock-provisioned-throughput) — buys serving capacity for the finished model via `custom_model_arn`
- [**AWS S3 Bucket**](/infra-catalog/aws-s3-bucket) — holds the training, validation, and output data the S3 URIs point at
- [**AWS KMS Key**](/infra-catalog/aws-kms-key) — customer-managed encryption of the resulting model via `customModelKmsKeyArn`
- [**AWS Subnet**](/infra-catalog/aws-subnet) — VPC placement for private-data jobs via `vpcConfig.subnetIds`
- [**AWS Security Group**](/infra-catalog/aws-security-group) — network rules on those interfaces via `vpcConfig.securityGroupIds`
