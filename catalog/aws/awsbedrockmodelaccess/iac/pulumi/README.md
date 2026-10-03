# Pulumi Module: AWS Bedrock Model Access

Provisions Amazon Bedrock model access (a foundation-model marketplace
agreement, plus the optional account use-case form) using Pulumi (Go).

## Resources Created

- `bedrockfoundation.ModelAgreement` — Accepts the model's PUBLIC offer;
  the offer token comes from the `bedrockfoundation.GetModelAgreementOffers`
  invoke at deploy time.
- `bedrock.UseCaseForModelAccess` — Only when `use_case_form` is
  declared: the account-global, write-once form (jsonencoded), ordered
  before the agreement via an explicit dependency.

## Notable Behavior

- The offers invoke runs at preview time, so previewing this module
  requires AWS credentials (the data-source-at-plan class). Behavioral
  parity with the Terraform module is the contract.

## Usage

The module is executed by the Planton platform through the entrypoint in
`main.go`, which loads the `AwsBedrockModelAccessIacInput`.
