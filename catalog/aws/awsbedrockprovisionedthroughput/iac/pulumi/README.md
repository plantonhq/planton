# Pulumi Module: AWS Bedrock Provisioned Throughput

Provisions Amazon Bedrock Provisioned Throughput (a dedicated
model-capacity purchase) using Pulumi (Go).

## Resources Created

- `bedrock.ProvisionedModelThroughput` — The capacity purchase: model
  reference, model units, and optional commitment term (omitted =
  no-commitment hourly billing). BILLS FROM CREATION. Behavioral parity
  with the Terraform module is the contract: identical send conditions
  and outputs.

## Usage

The module is executed by the Planton platform through the entrypoint in
`main.go`, which loads the `AwsBedrockProvisionedThroughputIacInput`.
