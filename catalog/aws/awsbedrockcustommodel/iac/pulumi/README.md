# Pulumi Module: AWS Bedrock Custom Model

Provisions an Amazon Bedrock model-customization job (and its resulting
custom model) using Pulumi (Go).

## Resources Created

- `bedrock.CustomModel` — The customization job: base model, customization
  type, hyperparameters, job role, training/validation data and output
  locations, optional VPC placement and model KMS key. Behavioral parity
  with the Terraform module is the contract: identical send conditions,
  identical job-name defaulting from `metadata.name`, identical outputs.

## Usage

The module is executed by the Planton platform through the entrypoint in
`main.go`, which loads the `AwsBedrockCustomModelIacInput`.
