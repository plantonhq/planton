# AWS CloudWatch Logs Account Policy

Deploys a CloudWatch Logs account-level policy — one rule AWS applies to every log group in the region: mask PII at ingest, forward everything to one destination, index the fields your queries filter on, transform events as they arrive, or extract metrics from log fields. The account-wide blast radius is the point of the resource: four of the five policy types offer no narrowing at all, and only subscription-filter policies can carve out exceptions. The policy's identity is its (name, type) pair — changing either replaces it.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **CloudWatch Logs Account Policy** — one policy object per (`policyName`, `policyType`) pair, carrying the type's own policy document. The provider's scope argument is pinned to `ALL` (its only legal value); `selectionCriteria` is sent only when set, and AWS accepts it only on subscription-filter policies.

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with credentials for the target AWS account, carrying `logs:PutAccountPolicy`, `logs:DescribeAccountPolicies`, and `logs:DeleteAccountPolicy`. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- **Weigh the account-wide effect per type** — the policy applies to every log group in the region, including ones created later. Masking and forwarding change what readers and destinations see; field indexing and metric extraction only add derived data.
- **A forwarding destination** (only for SUBSCRIPTION_FILTER_POLICY) — the Kinesis stream, Firehose, or Lambda the policy document's destination ARN points at, with the delivery permissions that destination type requires.

## Deploy

### Console

Open the deployment store, find **AWS CloudWatch Logs Account Policy**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and the spec fields — region, policy name, policy type, and the type's document. Start from the **Account-Wide PII Masking** or **Field Indexing for App Logs** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsCloudwatchLogAccountPolicy
metadata:
  name: app-field-index
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  policyName: app-field-index
  policyType: FIELD_INDEX_POLICY
  policyDocument:
    Fields:
      - requestId
      - customerId
      - sourceIp
```

```shell
planton apply -f log-account-policy.yaml
```

This creates a field-index policy that indexes `requestId`, `customerId`, and `sourceIp` across every log group in us-east-1 — Logs Insights queries filtering on indexed fields scan less. An Infra Job tracks the provisioning in real time.

## Key Configuration

These are the most important decisions when configuring an account policy. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The policy type is the capability** — `policyType` picks one of five account-wide behaviors: DATA_PROTECTION_POLICY masks sensitive data, SUBSCRIPTION_FILTER_POLICY forwards matching events, FIELD_INDEX_POLICY declares indexed fields, TRANSFORMER_POLICY rewrites events at ingest, METRIC_EXTRACTION_POLICY derives metrics from log fields. Changing the type replaces the policy.

**Each type carries its own document schema** — `policyDocument` is not one grammar: data protection carries data-identifier statements, subscription filter a destination-plus-filter-pattern object, field index a `Fields` list, transformer a processor pipeline, metric extraction a metric mapping. AWS validates the document server-side at Put time — a wrong-shaped document fails the apply, never the plan.

**Narrowing exists only for subscription filters** — AWS accepts `selectionCriteria` on SUBSCRIPTION_FILTER_POLICY alone, and its one supported grammar is an exact-name exclusion list: `LogGroupName NOT IN ["noisy-group"]` — no prefix form, no IN form. Every other type applies account-wide, period; PutAccountPolicy rejects any criteria string on them with "Invalid selection criteria provided". If you need a subtree-scoped transformer or index, use the per-log-group resource instead.

**One policy per type is the practical quota** — AWS bounds account policies per type (one for most types). Treat each (name, type) pair as a singleton per capability: a second FIELD_INDEX_POLICY under a different name is rejected server-side where the type is single-instance.

**Account transformer vs per-group transformer** — a TRANSFORMER_POLICY here applies to every log group; the per-log-group transformer (on AwsCloudwatchLogGroup) wins where both exist. Prefer per-group transformers for service-specific parsing and the account policy for org-wide normalization.

**What updates in place** — document edits apply in place; `policyName`, `policyType`, and `selectionCriteria` all replace the policy. The policy applies to matching log groups immediately, including groups created after it.

## Outputs and Dependencies

### What This Kind Consumes

This kind has no foreign key dependencies. A subscription-filter policy's destination ARN travels inside `policyDocument` as a plain string, not as a typed reference.

### What This Kind Provides

After provisioning, `status.outputs` echoes the policy's identity back: `policy_name` and `policy_type`, which together form the provider's import ID (`policy_name:policy_type`). Both are input echoes — nothing downstream composes on an account policy via ValueFromRef.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Account-wide PII masking** — a DATA_PROTECTION_POLICY auditing and masking emails and card numbers at ingest, so a leaked debug line never stores plaintext PII. It changes what every log reader sees — roll it out deliberately. Start from the **Account-Wide PII Masking** preset.

**Field indexing for query cost** — a FIELD_INDEX_POLICY declaring the fields Logs Insights queries filter on (`requestId`, `customerId`, `sourceIp`). Indexing adds derived data without changing log content, which makes it the safest type to adopt first. Start from the **Field Indexing for App Logs** preset.

**Account-wide forwarding with carve-outs** — a SUBSCRIPTION_FILTER_POLICY sending every log group's matching events to one destination, with `selectionCriteria` excluding the noisy groups by exact name. The one type where narrowing exists — use it when a handful of groups would drown the pipeline.

## Works With

- [**AWS CloudWatch Log Group**](/infra-catalog/aws-cloudwatch-log-group) — the groups the policy governs; per-group transformers and indexes on the group win over the account policy for service-specific needs
- [**AWS Kinesis Data Stream**](/infra-catalog/aws-kinesis-stream) — a common destination for subscription-filter policies, carried in the policy document by ARN
- [**AWS CloudWatch Logs Delivery**](/infra-catalog/aws-cloudwatch-log-delivery) — the cross-account destination other accounts' subscription policies can target
