# AWS SES Family: Configuration Set and Email Identity Forged at Full SESv2 Depth

**Date**: July 8, 2026
**Type**: Feature
**Components**: API Definitions, Provider Framework, Terraform Modules, Pulumi Modules, E2E Harness, CLI

## Summary

The AWS catalog gains its email-sending surface: `AwsSesConfigurationSet` (enum 366) and `AwsSesEmailIdentity` (enum 367), both modeled at the full SESv2 provider surface with 100% dual-engine parity, first-ever SES E2E coverage (all four live lanes green with a zero-orphan sweep), and dense provider-honest documentation. Transactional email is near-universal in real product architectures; until now a Planton user simply could not model it.

## What Was Built

### AwsSesConfigurationSet (366, `awssescs`)

The sending-policy container — the named group of rules identities reference as their default and any SendEmail call can name explicitly:

- **Delivery options**: TLS posture (`REQUIRE`/`OPTIONAL`), retry window (`max_delivery_seconds`, 300–50400), and the dedicated IP pool arm (`sending_pool_name`).
- **Reputation metrics** (CloudWatch bounce/complaint rates), the **per-set sending kill switch** (tri-state `optional bool`, catalog default true), **suppression-list overrides** (absent inherits the account default — a genuine tri-state honored in both engines), **custom tracking domain** with HTTPS policy, and **VDM** engagement-metrics/optimized-shared-delivery dials.
- **Event destinations folded as per-name blocks** — each named destination is its own AWS sub-resource (many-per-set, set-scoped lifecycle, never FK-referenced: the established fold class), publishing a chosen slice of email events into exactly one of five arms (CEL-enforced): CloudWatch dimension configurations, an EventBridge bus (ref), a Kinesis Firehose stream + IAM role (refs), an SNS topic (ref), or a Pinpoint application (literal ARN). AWS defaults a destination's `enabled` to FALSE — a created-but-silent destination is a classic source of missing events — so the catalog defaults it to true and both engines always send the value explicitly.
- Outputs: `configuration_set_arn` + `configuration_set_name` (the output-backed join key).

### AwsSesEmailIdentity (367, `awssesid`)

The verified sender — the trust anchor nothing leaves SES without:

- The identity string is the AWS identifier (deliberately spec-derived, not `metadata.name` — it must be the exact DNS name mail is sent from; immutable).
- **DKIM**: Easy DKIM (`next_signing_key_length`, RSA_1024/RSA_2048) XOR BYODKIM (base64 PKCS #8 private key — `(sensitive)`, a managed secret — paired with the selector); the pairing, exclusivity, and domain-identity-only rules are spec-level CEL.
- **Folded satellites**: the custom MAIL FROM domain (with `behavior_on_mx_failure` honesty — `REJECT_MESSAGE` fails sends when the MX record is missing), bounce/complaint email forwarding (tri-state), and per-name authorization policies (cross-account sending grants, Struct → JSON in both engines).
- The optional configuration-set ref (`AwsSesConfigurationSet.status.outputs.configuration_set_name`) keeps the kind registry **dependency-free**: optional composition is declared per-scenario via the `planton.dev/e2e-prerequisites` annotation, never as a registry prerequisite.
- Outputs include the repeated **`dkim_tokens`** — the three Easy DKIM CNAME tokens that compose directly into `AwsRoute53DnsRecord` nodes, which is the point of the kind being a first-class node.

### E2E (first-ever SES coverage)

- sesv2 SDK verifiers for both kinds (plain NotFound semantics — SES has no lifecycle states and no verification waiter; a PENDING identity exists and never blocks create/destroy).
- Config set: an install profile (so it can serve as a scenario prerequisite) plus two scenarios — minimal, and an event-destinations scenario proving the folded satellites live (CloudWatch metrics + an SNS feedback destination composed from the shared topic fixture via the annotation).
- Identity: the domain-easy-dkim scenario riding the config-set prerequisite chain.
- **Live dual-engine E2E 4/4 green**: config set 23s–50s per scenario-lane; identity chain 41.7s (Pulumi) / 50.1s (Terraform). Zero-orphan sweep clean.

## Defects Found and Fixed Along the Way

- **`coalesce(x, "") != ""` errors in HCL when the field IS empty** — `coalesce` requires at least one non-null, non-empty argument, so the guard fails at plan time exactly in the case it was meant to handle. Generator-flattened `StringValueOrRef` fields are never null (they carry the contract default `""`), so both SES Terraform modules now use plain `!= ""` comparisons. The class is folded into the Terraform-module authoring rule.
- **SES identity policies cannot ship in a static committed manifest** — AWS validates `CreateEmailIdentityPolicy` strictly: the document's `Resource` must be the identity's own ARN (account + region embedded). The live rejection ("Invalid ARN: ARNs must start with 'arn:': *") is recorded in the scenario, profile, and research doc; the fold is proven by spec/CEL tests and the offline plan gate.
- **The variables.tf drift guard failed under the Bazel sandbox** (pre-existing — the sandbox has no repo checkout, so `go.mod` is unreachable): it now skips explicitly under Bazel with the reason, keeping the plain `go test` lane as the enforcement point.
- **`planton tofu generate-variables` wrote its output to stderr** via the builtin `println`, silently breaking `> variables.tf` capture — now `fmt.Println` (the same class as the earlier `load-tfvars` fix).

## Validation

- Offline gate all green: `make protos`, spec tests ×2 (go + Bazel), targeted Go + Pulumi builds, `make build-go`, kind-map regeneration, drift guard (both kinds enrolled, contracts byte-identical to the generator), outputs conformance (+2 cases, incl. the repeated `dkim_tokens` population proof), `tofu init`+`validate` ×2, offline `tofu plan` from both hack manifests, `validate-refs --check`, `secret-coverage --check`, `validate-outputs` ×2, all ten manifests CLI-validated, site catalog regenerated, scaffolding-leakage grep clean.
- Live gate: all four dual-engine lanes green with a zero-orphan account sweep (no configuration sets, identities, or fixture topics left).

## Deferred Surface (recorded with reasons in the kind research docs)

Dedicated IP pools/assignments (paid capacity surface; the `sending_pool_name` arm composes by name), contact lists (marketing data plane), tenants (new multi-tenant surface), account-level VDM/suppression attributes (account singletons), classic-SES receiving (a separate product surface), and `aws_ses_template` (classic-SES-only; SESv2 templates have no provider resource).
