# Pulumi Module: AWS Bedrock Prompt

Provisions an Amazon Bedrock prompt (Prompt Management) using Pulumi (Go).

## Resources Created

- `bedrock.AgentPrompt` — The prompt's mutable DRAFT version with its
  variants: text or chat templates (AWS's template_type derived from the
  configured arm), model or agent-alias execution targets, tools, and
  inference settings.

## Notable Behavior

- Behavioral parity with the Terraform module is the contract: identical
  derived discriminators, identical constants (`default` cache point
  type), identical outputs.
- Only the DRAFT version is managed — AWS assigns a new version string on
  every update.
- The template tree mirrors the AwsBedrockFlow module's inline-prompt
  rendering arm-for-arm (upstream shares the same models between the two
  resources) — change them together.

## Usage

The module is executed by the Planton platform through the entrypoint in
`main.go`, which loads the `AwsBedrockPromptIacInput`.
