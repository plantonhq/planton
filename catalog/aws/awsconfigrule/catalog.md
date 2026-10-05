# AWS Config Rule

Creates one AWS Config compliance check over the region's recorded configurations -- an AWS-managed rule, a custom Lambda evaluator, or a CloudFormation-Guard policy -- with optional organization-wide deployment and SSM auto-remediation. Exactly one of the three source arms is set, the rule name comes from `metadata.name`, and the region must already have a running configuration recorder or every evaluation reports nothing.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Config Rule** -- the compliance check itself: an account-scoped rule, or (when `organization` is set) an organization rule deployed to every member account, with the source arm deciding which provider resource renders
- **Remediation Configuration** -- created only when `remediation` is set (account-scoped rules only); the SSM document AWS Config runs against non-compliant resources, manual or automatic with a retry contract

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with AWS Config permissions; for organization rules, its credentials must belong to the management account or the Config delegated administrator. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- A RUNNING configuration recorder in the region -- AWS rejects rule creation without one, and a rule under a stopped recorder evaluates nothing silently.
- For `customLambda`: the function must already carry a `lambda:InvokeFunction` permission for `config.amazonaws.com` -- AWS validates the grant at rule creation.
- For `remediation`: the SSM document (AWS-authored like `AWS-ConfigureS3BucketVersioning`, or your own) and any role it assumes.

## Deploy

### Console

Open the deployment store, find **AWS Config Rule**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and the spec: the source arm, scope, evaluation modes, and remediation. Start from the **Managed Rule with Remediation** preset in the [Presets](#presets) tab for the zero-code detect-and-repair shape.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsConfigRule
metadata:
  name: s3-bucket-versioning-enabled
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  description: Checks that versioning is enabled on every recorded S3 bucket
  managed:
    ruleIdentifier: S3_BUCKET_VERSIONING_ENABLED
  scope:
    complianceResourceTypes:
      - AWS::S3::Bucket
```

```shell
planton apply -f aws-config-rule.yaml
```

This creates an account-scoped rule running AWS's maintained versioning check against every recorded S3 bucket in the region. An Infra Job tracks the provisioning in real time.

### InfraChart

When a custom-Lambda rule deploys alongside its evaluator in one chart, wire the function via ValueFromRef:

```yaml
spec:
  region: us-east-1
  description: Custom evaluation of EC2 instance compliance
  customLambda:
    functionArn:
      valueFrom:
        kind: AwsLambda
        name: config-evaluator
        fieldPath: status.outputs.function_arn
  scope:
    complianceResourceTypes:
      - AWS::EC2::Instance
```

The InfraPipeline resolves the dependency graph, deploys the function first, then creates the rule that invokes it.

## Key Configuration

These are the most important decisions when configuring a Config rule. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Pick the source arm by how much logic you own** -- `managed` references one of AWS's hundreds of maintained rules by identifier: zero code, AWS keeps the logic current. `customPolicy` is the cheap custom path: your logic as a CloudFormation-Guard policy, evaluated inside AWS Config with no compute of your own to deploy, patch, or pay for -- but change-triggered only. `customLambda` is full control for checks that need to call other APIs, at the price of operating a function. Exactly one arm is allowed; validation fails a manifest with zero or two before AWS sees it.

**No recorder, no rule** -- AWS rejects rule creation in a region without a configuration recorder, and the quieter failure is worse: a rule under a STOPPED recorder, or one whose recording group misses the rule's types, evaluates nothing without erroring. Keep the AWS Config Recorder's scope covering every type your rules evaluate.

**Scope is what keeps evaluations relevant** -- unset, the rule evaluates every recorded resource its logic applies to. `complianceResourceTypes` narrows to types, `tagKey`/`tagValue` to tagged resources, and `complianceResourceId` pins one specific resource (its type required alongside). Unscoped Guard rules in particular evaluate everything the recorder captures.

**Organization scope changes the rules of the game** -- setting `organization` deploys the check to every member account (with per-account exclusions), but must run from the management or delegated-administrator account, drops the name cap from 128 to 64 characters, and gives up remediation and proactive evaluation -- AWS has neither for organization rules. Custom organization rules declare `triggerTypes` on the organization message; managed ones derive their own. Member-account rollout is asynchronous.

**Automatic remediation is a loop -- bound it** -- `remediation` with `automatic: true` requires the retry contract (`maximumAutomaticAttempts` within `retryAttemptSeconds`), and the `errorPercentage` circuit breaker is what stops a broken SSM document from remediating the fleet into an outage. Start with `automatic: false` -- the fix stays one click away in the Config console -- and graduate to automatic only after the document has proven itself.

**Periodic versus change-triggered** -- `maximumExecutionFrequency` matters only for periodic rules; change-triggered rules evaluate as recorded resources change and ignore it. Guard-policy rules are change-triggered by design, and AWS rejects scheduled triggers on organization Guard rules -- the spec encodes that rejection at validation time.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsLambda** | `customLambda.functionArn` | `status.outputs.function_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `rule_name` | The rule's name (`metadata.name` echoed) | The key remediation attachments and aggregator compliance queries address rules by |

The remaining outputs -- `rule_arn` (the organization rule ARN for org-scoped rules), `rule_id` (empty for organization rules), and `remediation_arn` (set only when remediation is configured) -- are record values for audit and import, not composition inputs.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Managed rule with a one-click fix** -- an AWS-authored check (S3 bucket versioning) with an SSM document wired as human-triggered remediation, throttled by concurrency and circuit-broken on failures. The zero-code detect-and-repair shape; flip to automatic only after the manual fix has proven itself. Start from the **Managed Rule with Remediation** preset.

**Guard policy for house rules** -- organization-specific checks no managed rule covers (tagging standards, naming conventions, property combinations) written in the Guard DSL, with the engine running inside AWS Config. Keep `scope` matched to the types the policy addresses. Start from the **Guard Policy Rule** preset.

**Organization-wide baseline** -- the same managed check deployed to every member account via the `organization` arm, run from the delegated-administrator account with exclusions for accounts that genuinely differ. Pair with an AWS Config Aggregator so the compliance picture rolls up to one place.

## Works With

- [**AWS Config Recorder**](/infra-catalog/aws-config-recorder) -- the regional prerequisite; rules only evaluate what the recorder captures
- [**AWS Lambda**](/infra-catalog/aws-lambda) -- the evaluator behind `customLambda` rules, wired via `functionArn`
- [**AWS Config Aggregator**](/infra-catalog/aws-config-aggregator) -- rolls this rule's compliance results up across accounts and regions
- [**AWS Config Conformance Pack**](/infra-catalog/aws-config-conformance-pack) -- the packaged alternative when a whole set of related rules ships as one unit
