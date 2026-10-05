# Terraform Module: AWS Bedrock Guardrail

Provisions an Amazon Bedrock guardrail using Terraform.

## Resources Created

- `aws_bedrock_guardrail` — The guardrail's mutable DRAFT definition: the
  required blocked messagings, up to five policy families (content
  filters, denied topics, word filters, sensitive information, contextual
  grounding), optional cross-region profile and KMS key.
- `aws_bedrock_guardrail_version` — One immutable published version per
  `spec.versions` entry (`for_each` keyed by the entry's `name`), ordered
  after the guardrail so a same-deploy draft edit is captured. Each
  AWS-assigned number is exported in the `version_numbers` output map.

## Notable Behavior

- Topic type (`DENY`) and the managed word list type (`PROFANITY`) are
  module constants — AWS defines no other values at the pin.
- Per-direction `*_action`/`*_enabled` arms are sent only when set
  (explicit false included), so one-direction disablement is expressible.
- `keep_on_delete` maps to the provider's `skip_destroy`: the published
  version survives removal from management.

## Usage

The module is executed by the Planton platform. `variables.tf` is
GENERATED from the kind's spec (`planton tofu generate-variables
AwsBedrockGuardrail`) — never edit it by hand.
