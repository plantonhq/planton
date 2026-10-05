<p align="center">
  <img src="logo.svg" alt="AWS Config Rule" width="80"/>
</p>

# AWS Config Rule

Manage one [AWS Config rule](https://docs.aws.amazon.com/config/latest/developerguide/evaluate-config.html)
— a compliance check evaluated against the resource configurations the
region's recorder captures — plus its optional auto-remediation.

`metadata.name` is the rule name (AWS caps account-scoped rules at 128
characters; organization-scoped ones at 64).

## What Gets Managed

- **One evaluation source** (exactly one of three):
  - `managed` — an AWS-authored rule by identifier (e.g.
    `S3_BUCKET_VERSIONING_ENABLED`); the zero-code path.
  - `custom_lambda` — your Lambda function evaluates resources (AWS
    Config must be allowed to invoke it).
  - `custom_policy` — a CloudFormation-Guard policy; custom logic
    without running any compute.
- **Scope**: which resources the rule evaluates (types, one pinned
  resource, or a tag).
- **Evaluation modes**: DETECTIVE (after deployment) and/or PROACTIVE
  (before provisioning) — account-scoped rules only.
- **Organization deployment**: presence of `organization` deploys the
  rule across every account in the AWS Organization (run from the
  management or delegated-admin account).
- **Remediation**: the SSM document AWS Config runs against
  non-compliant resources, manually or automatically with a retry
  contract.

The region needs a running recorder
([AwsConfigRecorder](../awsconfigrecorder)) or every evaluation
reports nothing — AWS rejects rule creation without one. Conformance
packs (template bundles that create their own rules) are deliberately
NOT part of this kind; they own their own lifecycle and ship as
their own kind.

Destroying this component deletes the rule and its remediation
configuration.

See [v1alpha1/reference.md](v1alpha1/reference.md) for the full field
reference generated from the spec proto.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
