# Pulumi Module: AWS Bedrock Guardrail

Provisions an Amazon Bedrock guardrail using Pulumi (Go).

## Resources Created

- `bedrock.Guardrail` — The guardrail's mutable DRAFT definition: the
  required blocked messagings, up to five policy families (content
  filters, denied topics, word filters, sensitive information, contextual
  grounding), optional cross-region profile and KMS key.
- `bedrock.GuardrailVersion` — One immutable published version per
  `spec.versions` entry (resource name `version-<entry name>`), ordered
  after the guardrail via an explicit dependency so a same-deploy draft
  edit is captured. Each AWS-assigned number is exported in the
  `version_numbers` output map.

## Notable Behavior

- Behavioral parity with the Terraform module is the contract: identical
  send conditions (send-when-set action/enabled arms, omitted tier =
  AWS default), identical constants (DENY topics, PROFANITY managed
  list), identical outputs.
- Version entries iterate name-sorted for deterministic previews.

## Usage

The module is executed by the Planton platform through the entrypoint in
`main.go`, which loads the `AwsBedrockGuardrailIacInput`.
