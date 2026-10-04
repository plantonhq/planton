# Terraform Module: AWS Bedrock Model Access

Provisions Amazon Bedrock model access (a foundation-model marketplace
agreement, plus the optional account use-case form) using Terraform.

## Resources Created

- `aws_bedrock_foundation_model_agreement` — Accepts the model's PUBLIC
  offer; the offer token comes from the
  `aws_bedrock_foundation_model_agreement_offers` data source at deploy
  time. Create waits for AVAILABLE; destroy cancels the agreement.
- `aws_bedrock_use_case_for_model_access` — Only when `use_case_form` is
  declared: the account-global, write-once form (jsonencoded), ordered
  before the agreement.

## Notable Behavior

- The offers data source reads at plan time, so planning this module
  requires AWS credentials (the data-source-at-plan class).

## Usage

The module is executed by the Planton platform. `variables.tf` is
GENERATED from the kind's spec (`planton tofu generate-variables
AwsBedrockModelAccess`) — never edit it by hand.
