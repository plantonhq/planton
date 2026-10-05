# Terraform Module: AWS Bedrock Custom Model

Provisions an Amazon Bedrock model-customization job (and its resulting
custom model) using Terraform.

## Resources Created

- `aws_bedrock_custom_model` — The customization job: base model,
  customization type, hyperparameters, job role, training/validation data
  and output locations, optional VPC placement and model KMS key. Every
  argument is create-time-immutable (the job cannot be altered once
  started); the job name defaults to `metadata.name` and MUST be
  overridden via `spec.job_name` for re-runs (AWS never reuses job names).

## Usage

The module is executed by the Planton platform. `variables.tf` is
GENERATED from the kind's spec (`planton tofu generate-variables
AwsBedrockCustomModel`) — never edit it by hand.
