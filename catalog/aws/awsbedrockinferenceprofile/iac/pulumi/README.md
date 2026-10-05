# Pulumi Module: AWS Bedrock Inference Profile

Provisions an Amazon Bedrock application inference profile using Pulumi
(Go).

## Resources Created

- `bedrock.InferenceProfile` — The application profile: name from
  `metadata.name`, optional description, the model source, and the
  canonical identity tags. Behavioral parity with the Terraform module is
  the contract: identical send conditions and outputs.

## Usage

The module is executed by the Planton platform through the entrypoint in
`main.go`, which loads the `AwsBedrockInferenceProfileIacInput`.
