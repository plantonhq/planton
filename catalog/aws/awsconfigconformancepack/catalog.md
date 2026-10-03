# AWS Config Conformance Pack

Deploys an AWS Config conformance pack -- a template bundle that turns a ruleset (CIS, PCI, or your own) into one unit that deploys, scores, and unwinds together, per account or across the whole AWS Organization. The template CREATES its own Config rules, prefixed with the pack name and managed by the pack's service-linked role -- a pack references no standalone rules. Deployment requires a running Config recorder in the region; AWS rejects the pack without one.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Conformance Pack** -- an account-scoped pack, or (when `organizationScope: true`) an organization pack deployed into every member account minus exclusions; the pack's name is `metadata.name`
- **Pack-Owned Config Rules** -- created by AWS from the template, not by the module directly: every rule the template declares materializes prefixed with the pack name, owned and managed by the pack's service-linked role

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with AWS Config permissions; for organization scope, its credentials must belong to the management account or a delegated Config administrator. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- A RUNNING Config recorder in the region -- AWS rejects pack deployment without one, and at organization scope every member account needs its own recorder for the pack's rules to evaluate.
- For `templateS3Uri`: the deploying principal must be able to read the S3 object.
- For organization-scope delivery: the results bucket must be named with the `awsconfigconforms` prefix -- an AWS naming contract enforced at deploy.

## Deploy

### Console

Open the deployment store, find **AWS Config Conformance Pack**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and the spec: the template (inline or S3), scope, parameters, and delivery. Start from the **S3 Best Practices Pack** preset in the [Presets](#presets) tab for a pack small enough to read end to end.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsConfigConformancePack
metadata:
  name: s3-best-practices
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  templateBody: |
    Resources:
      S3BucketPublicReadProhibited:
        Type: AWS::Config::ConfigRule
        Properties:
          ConfigRuleName: s3-bucket-public-read-prohibited
          Source:
            Owner: AWS
            SourceIdentifier: S3_BUCKET_PUBLIC_READ_PROHIBITED
      S3BucketPublicWriteProhibited:
        Type: AWS::Config::ConfigRule
        Properties:
          ConfigRuleName: s3-bucket-public-write-prohibited
          Source:
            Owner: AWS
            SourceIdentifier: S3_BUCKET_PUBLIC_WRITE_PROHIBITED
      S3BucketVersioningEnabled:
        Type: AWS::Config::ConfigRule
        Properties:
          ConfigRuleName: s3-bucket-versioning-enabled
          Source:
            Owner: AWS
            SourceIdentifier: S3_BUCKET_VERSIONING_ENABLED
```

```shell
planton apply -f aws-config-conformance-pack.yaml
```

This deploys a three-rule S3 hygiene pack -- no public reads, no public writes, versioning on -- scored as one compliance number in this account, with all three rules created and owned by the pack. An Infra Job tracks the provisioning in real time.

### InfraChart

When the pack deploys alongside its results bucket in one chart, wire the delivery bucket via ValueFromRef:

```yaml
spec:
  region: us-east-1
  templateS3Uri: s3://acme-conformance-templates/security-baseline.yaml
  deliveryS3Bucket:
    valueFrom:
      kind: AwsS3Bucket
      name: awsconfigconforms-acme-results
      fieldPath: status.outputs.bucket_id
```

The InfraPipeline resolves the dependency graph, creates the bucket first, then the pack delivering results into it.

## Key Configuration

These are the most important decisions when configuring a conformance pack. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Pack rules are pack-owned** -- they appear in the region's rules list prefixed with the pack name, but editing or deleting them outside the pack fights the service-linked role, which reasserts the template. Every rule change goes through the template; the standalone AWS Config Rule kind is for rules that live outside any pack.

**Account or organization scope is a real fork** -- `organizationScope: true` deploys the pack into every member account (minus `excludedAccounts`, which only mean something at this scope), requires the management or delegated-administrator account, drops the name cap from 256 to 128 characters, and unwinds slowly on delete -- member-account teardown takes minutes. Exempt sandbox accounts explicitly instead of silently.

**Two template forms with an asymmetric contract** -- account packs accept `templateBody` and `templateS3Uri` together (AWS prefers the S3 one); organization packs accept exactly one. An S3-hosted template is one artifact of record, versioned like code -- the right form for org baselines. Either way AWS never reports the template back, so template drift is undetectable by design and imports re-assert it on the first apply.

**Evaluations bill like ordinary rules** -- a pack of 20 rules evaluates like 20 standalone rules. The cost lever is the AWS Config Recorder's recording group: scope it to the types the pack's rules actually inspect.

**Delivery is optional, with a naming contract** -- unset, results stay queryable through the Config APIs and console only. When `deliveryS3Bucket` is set on an organization pack, the bucket name must start with `awsconfigconforms` -- AWS enforces it at deploy, not at validation.

**Destroy deletes the pack and everything it created** -- all pack-owned rules go with it, which is the point: compliance as one unit includes teardown as one unit.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsS3Bucket** | `deliveryS3Bucket` | `status.outputs.bucket_id` |

### What This Kind Provides

`status.outputs` echoes the pack's identity -- `pack_name` (the import ID at either scope), `pack_arn` (the account-scope or organization-scope ARN, whichever deployed), and `region` (packs are addressed by region plus name, so verifiers need both). These are audit and import values, not composition inputs: compliance scores are consumed through the Config console and APIs, not by other Infra Components.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Small readable pack first** -- three managed S3 rules deployed and scored as one unit: baseline hygiene with one conformance number instead of three scattered rules, and one-step teardown. The right first pack in any account -- small enough to read end to end before trusting the mechanism with a CIS-sized ruleset. Start from the **S3 Best Practices Pack** preset.

**Organization security baseline** -- one S3-hosted ruleset deployed to every member account from the management or delegated-admin account, with sandbox accounts excluded by ID. New member accounts are picked up automatically, and each account scores individually in the aggregated Config view. Start from the **Organization Security Baseline** preset.

**Pack for the baseline, standalone rules for the exceptions** -- keep the org-wide ruleset in a pack (one template of record) and express account-specific checks as standalone AWS Config Rule resources; the two coexist in the same region without fighting because pack rules are name-prefixed and role-owned.

## Works With

- [**AWS Config Recorder**](/infra-catalog/aws-config-recorder) -- the hard prerequisite; deploy it first in every region (and account) the pack covers
- [**AWS S3 Bucket**](/infra-catalog/aws-s3-bucket) -- the optional results-delivery bucket (org scope requires the `awsconfigconforms` name prefix), and the host for S3-based templates
- [**AWS Config Rule**](/infra-catalog/aws-config-rule) -- the standalone alternative for single checks and account-specific exceptions outside any pack
- [**AWS Config Aggregator**](/infra-catalog/aws-config-aggregator) -- rolls per-account pack scores into the organization-wide compliance view
