# AwsBedrockCustomModel — Kind Guide

Authored operational judgment for the Bedrock custom model kind.

## Design decisions

- **The kind IS the customization job.** The provider tracks the job (its
  identity is the job ARN; the model is derived from it), and so does this
  component: deploying starts the job, destroying deletes the model and
  the job record. Every field is create-time-immutable because
  CreateModelCustomizationJob has no update.
- **Single-member provider blocks flatten to leaves.** The provider nests
  each S3 location under a one-element config block
  (`training_data_config { s3_uri }`); the spec carries
  `training_data_s3_uri` directly. The parity manifest records each
  flattening.
- **`base_model_arn` says what the value is.** The provider calls it
  `base_model_identifier`; the value at the pin is a foundation-model ARN,
  and the spec's name plus its CEL pattern say so.
- **The job name defaults to metadata.name but stays overridable.** Job
  names are unique per account FOREVER (the name-history class) — the
  default serves the first run, and `job_name` exists precisely for
  re-runs. Taught on the field, here, and in the module comments.
- **IMPORTED is out of scope.** The provider's customization-type enum
  carries IMPORTED, but imported models arrive through a different AWS
  workflow (no customization job) — the spec's value domain deliberately
  omits it.

## Operational judgment

- **Validate the pipeline cheap.** One epoch, minimal dataset, smallest
  base model — confirm data format and role permissions before spending
  on the real run.
- **The role is the usual failure point.** It must trust
  `bedrock.amazonaws.com` AND reach all three S3 locations (plus KMS when
  buckets or the model use customer keys). The provider retries through
  "Could not assume provided IAM role" during propagation.
- **Hyperparameter keys are per base model** — epochCount/batchSize/
  learningRate are common, but consult the base model's documentation;
  AWS validates server-side.
- **Watch `job_status`, not apply success.** The deploy returns with the
  job InProgress; only Completed means a usable model. A Failed job's
  detail lives in the job record and the output S3 location.
- **Deleting while InProgress**: the provider stops the job (2-hour
  delete timeout) — prefer letting jobs finish or fail before destroy.

## Coverage decisions

- Every configurable argument of `aws_bedrock_custom_model` at the pinned
  provider is modeled, mapped, or excluded with a reason in
  `iac/provider-parity.yaml` (zero findings at forge time).
- The live E2E lane is a recorded deferral (real training spend +
  wall-clock; see e2e/profile.yaml for the unblock) — offline plans and
  previews prove both engines' renders arm-for-arm.
