# Pulumi Module: AWS Bedrock Flow

Provisions an Amazon Bedrock flow using Pulumi (Go).

## Resources Created

- `bedrock.AgentFlow` — The flow's mutable DRAFT definition: the node
  graph (nodes with per-class configuration union members, connections
  with data/conditional arms) plus the execution role and optional KMS
  key.

## Notable Behavior

- Behavioral parity with the Terraform module is the contract: identical
  derived union members (structural node classes render EMPTY members,
  the Loop family renders none — an upstream gap at the pin), identical
  constants (Python_3 inline-code language, `default` cache point type,
  S3 retrieval/storage services), identical outputs.
- Graph topology is validated by AWS server-side at create/update.
- The inline prompt tree mirrors the AwsBedrockPrompt module arm-for-arm
  (upstream shares the same models between the two resources) — change
  them together.

## Usage

The module is executed by the Planton platform through the entrypoint in
`main.go`, which loads the `AwsBedrockFlowIacInput`.
