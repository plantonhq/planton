# Terraform Module: AWS Bedrock Inference Profile

Provisions an Amazon Bedrock application inference profile using
Terraform.

## Resources Created

- `aws_bedrock_inference_profile` — The application profile: name from
  `metadata.name`, optional description, the model source
  (`model_source.copy_from` from `spec.source_arn`), and the canonical
  identity tags. Every argument is create-time-immutable; a change
  replaces the profile and its ARN.

## Usage

The module is executed by the Planton platform. `variables.tf` is
GENERATED from the kind's spec (`planton tofu generate-variables
AwsBedrockInferenceProfile`) — never edit it by hand.
