# AWS Cost Category

Deploys a Cost Explorer cost category: a dimension you define that groups every line of AWS spend into exactly one named value ("Cost Center" = platform/data/mobile), visible everywhere Cost Explorer reports. Rules evaluate in order with first match winning — REGULAR rules assign a value you name to the spend an expression selects, INHERITED_VALUE rules take the value from a dimension itself (the account name, a tag's value). Optional split-charge rules re-allocate a shared value's costs across targets for full chargeback. Cost Explorer is account-global; the spec's region is only the provider's API endpoint.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Cost Category Definition** — the named category with its ordered rules, optional default value for unmatched spend, and optional effective start month. The provider's rule version is module-pinned to `CostCategoryExpression.v1`, the only value the API accepts.
- **Split-Charge Rules** — configured only when `splitChargeRules` is set: re-allocations of one value's costs across target values, proportionally, evenly, or by fixed percentages.

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with Cost Explorer permissions. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- **Activated cost-allocation tags** (only for tag-based rules) — a tag expression or an inherited TAG rule sees nothing until the tag key is activated in the Billing console.

## Deploy

### Console

Open the deployment store, find **AWS Cost Category**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and the ordered rule list. Start from the **Team Chargeback** preset in the [Presets](#presets) tab for the one-rule category most organizations need first.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsCostCategory
metadata:
  name: team-chargeback
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  categoryName: Team
  defaultValue: unallocated
  rules:
    - type: INHERITED_VALUE
      inheritedValue:
        dimensionName: TAG
        dimensionKey: team
```

```shell
planton apply -f cost-category.yaml
```

This creates a "Team" category with one inherited-value rule fanning out to a value per distinct `team` tag value — new teams appear automatically — and untagged spend landing visibly in `unallocated`. An Infra Job tracks the provisioning in real time.

## Key Configuration

These are the most important decisions when configuring a cost category. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Order is the algorithm** — rules evaluate in spec order and the first match wins, so a broad early rule swallows spend later rules were meant to catch. Put narrow, specific rules first (a shared account routed wholesale) and broad fan-outs (the tag inheritance) after.

**One inherited rule beats N regular rules** — an INHERITED_VALUE rule over a tag fans out to a value per distinct tag value with no rule edits as teams appear. Its blind spot: it only sees ACTIVATED cost-allocation tags, and un-activated keys inherit nothing — silently.

**Cost category rules take a restricted dimension set** — only `BILLING_ENTITY`, `LINKED_ACCOUNT`, `LINKED_ACCOUNT_NAME`, `RECORD_TYPE`, `REGION`, `SERVICE_CODE`, and `USAGE_TYPE`, not the full Cost Explorer vocabulary budgets and anomaly monitors accept. By-service rules use `SERVICE_CODE` with service codes (`AmazonS3`, `AmazonEC2`), never the SERVICE display-name dimension.

**Make the default value your tagging-gap report** — `defaultValue` names where unmatched spend lands (`unallocated` reads better in front of finance than AWS's literal `Uncategorized`), and watching that bucket shrink is the honest measure of tagging coverage.

**Split-charge rules dissolve the shared bucket** — PROPORTIONAL splits by each target's own costs, EVEN splits equally, FIXED takes one percentage per target summing to 100. Two constraints to plan around: a split source cannot be the target of another split rule (AWS rejects circular allocation), and the source value must exist before it can be split.

**Changes are retroactive to the month** — AWS re-categorizes the current and previous month on every rule change, and `effectiveStart` accepts month boundaries only, up to twelve months back. Reports for closed months further back never change.

**Renaming replaces; deleting end-dates** — `categoryName` is create-only, so a rename resets the category's history. A deleted category keeps answering reads until month end — AWS end-dates it (`effective_end`) rather than erasing it, keeping in-flight reports coherent.

## Outputs and Dependencies

### What This Kind Consumes

This kind has no foreign key dependencies — rules reference accounts, tags, and other categories by name inside Cost Explorer expressions, which travel as plain values.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `category_name` | The category's name — the key other Cost Explorer expressions reference | Budget filter expressions and anomaly-monitor slices grouping by this category; other categories building on it |
| `category_arn` | The category's ARN (also the provider's import ID) | IAM policy statements scoping who may edit the categorization rules |

`effective_start` and `effective_end` are also echoed — the AWS-normalized window the rules apply in. They are observability echoes, not composition inputs.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Chargeback by tag** — one inherited-value rule over the `team` (or `product`) tag, untagged spend surfaced in an `unallocated` default bucket. The bucket is the point: it turns invisible tagging gaps into a line item somebody owns. Start from the **Team Chargeback** preset.

**Shared costs dissolved into products** — product values from their tags, a platform value composed with or-logic (the platform tag OR the Support service), and a PROPORTIONAL split-charge rule re-allocating the platform value across products by their own spend. Full chargeback with no orphan "shared" line — the trade is that product owners now see platform costs they don't directly control. Start from the **Shared Platform Split** preset.

**Categories built on categories** — a rule's `costCategory` leaf matches another category's values, letting a coarse "Business Unit" category build on a finer "Team" one. Know the failure mode: deleting a referenced category breaks the referencing rule at AWS, not at plan time.

## Works With

- [**AWS Budget**](/infra-catalog/aws-budget) — budgets scoped to a category's values through their filter expressions
- [**AWS Cost Anomaly Monitor**](/infra-catalog/aws-cost-anomaly-monitor) — CUSTOM monitors watching one category value's spend as its own anomaly stream
