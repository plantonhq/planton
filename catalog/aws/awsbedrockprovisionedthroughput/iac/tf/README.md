# Terraform Module: AWS Bedrock Provisioned Throughput

Provisions Amazon Bedrock Provisioned Throughput (a dedicated
model-capacity purchase) using Terraform.

## Resources Created

- `aws_bedrock_provisioned_model_throughput` — The capacity purchase:
  model reference, model units, and optional commitment term (omitted =
  no-commitment hourly billing). BILLS FROM CREATION; committed purchases
  cannot be destroyed until the term lapses. Everything except tags is
  create-time-immutable.

## Usage

The module is executed by the Planton platform. `variables.tf` is
GENERATED from the kind's spec (`planton tofu generate-variables
AwsBedrockProvisionedThroughput`) — never edit it by hand.
