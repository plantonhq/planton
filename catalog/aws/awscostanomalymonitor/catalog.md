# AWS Cost Anomaly Monitor

Deploys a Cost Explorer anomaly monitor: AWS's ML-driven watcher that learns your normal spend pattern and flags deviations, together with the alert subscriptions that decide who hears, how often, and above what impact. A monitor takes one of two shapes — DIMENSIONAL segments spend by one built-in dimension (the classic by-service monitor), CUSTOM watches exactly the slice a Cost Explorer expression selects (a team's tag, a member account, a cost category). Cost Explorer is account-global; the spec's region is only the provider's API endpoint.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Anomaly Monitor** — the ML watcher in its chosen shape. Both shape arms are create-only: changing the shape replaces the monitor, and only `monitorName` updates in place.
- **Alert Subscriptions** — one per `subscriptions` entry: a delivery frequency, its recipients, and an optional impact threshold, each getting its own ARN (echoed in the `subscription_arns` output).

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with Cost Explorer permissions. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- **Cost Explorer enabled** — a first visit to the Cost Explorer console enables it; data appears within 24 hours.
- **The DIMENSIONAL singleton check** — AWS permits exactly one services monitor per account and auto-creates it (`Default-Services-Monitor`) on accounts that enabled Cost Explorer on or after 2023-03-27. Check `aws ce get-anomaly-monitors` before deploying a DIMENSIONAL monitor: if it exists, import it rather than creating (a second create fails); CUSTOM monitors have no such limit (up to 500 per account).
- **Activated cost-allocation tags** (only for tag-sliced CUSTOM monitors) — an unactivated tag key matches nothing.

## Deploy

### Console

Open the deployment store, find **AWS Cost Anomaly Monitor**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields covering the monitor shape and subscriptions. Start from the **Custom Slice Monitor** preset in the [Presets](#presets) tab to watch one team's spend with real-time alerts.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsCostAnomalyMonitor
metadata:
  name: team-platform-monitor
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  monitorName: Team Platform Spend Monitor
  monitorType: CUSTOM
  monitorSpecification:
    And: null
    CostCategories: null
    Dimensions: null
    Not: null
    Or: null
    Tags:
      Key: user:team
      MatchOptions:
        - EQUALS
      Values:
        - platform
  subscriptions:
    - name: platform-daily-summary
      frequency: DAILY
      subscribers:
        - address:
            value: finops@acme-corp.com
          type: EMAIL
      thresholdExpression:
        dimension:
          key: ANOMALY_TOTAL_IMPACT_ABSOLUTE
          matchOptions:
            - GREATER_THAN_OR_EQUAL
          values:
            - "100"
```

```shell
planton apply -f cost-anomaly-monitor.yaml
```

This creates a CUSTOM monitor watching spend tagged `team: platform` (spelled `user:team` — Cost Explorer's canonical form for user-defined tag keys), with a daily email digest of anomalies whose absolute impact reaches 100 USD. An Infra Job tracks the provisioning in real time.

### InfraChart

When real-time alerts should publish to an SNS topic deployed in the same chart, wire the subscriber via ValueFromRef:

```yaml
spec:
  region: us-east-1
  monitorName: Team Platform Spend Monitor
  monitorType: CUSTOM
  monitorSpecification:
    And: null
    CostCategories: null
    Dimensions: null
    Not: null
    Or: null
    Tags:
      Key: user:team
      MatchOptions:
        - EQUALS
      Values:
        - platform
  subscriptions:
    - name: platform-alerts
      frequency: IMMEDIATE
      subscribers:
        - address:
            valueFrom:
              kind: AwsSnsTopic
              name: cost-alerts
              fieldPath: status.outputs.topic_arn
          type: SNS
```

The InfraPipeline resolves the dependency graph, deploys the topic first, then binds the monitor's immediate alerts to it.

## Key Configuration

These are the most important decisions when configuring an anomaly monitor. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Your account probably already has the services monitor** — AWS auto-creates `Default-Services-Monitor` (the DIMENSIONAL/SERVICE shape) on post-2023 Cost Explorer accounts, and permits exactly one per account: creating a second fails at apply with "Limit exceeded on dimensional spend monitor creation". Adopt the existing monitor through import (its ARN is the import ID), or reach for CUSTOM monitors, which allow up to 500.

**The shape is a one-way door** — `monitorType`, `monitorDimension`, and `monitorSpecification` are all create-only; changing any of them replaces the monitor and restarts its learning. Only the display name updates in place.

**Author CUSTOM expressions in canonical form** — the Expression document must carry every root member (`And`, `CostCategories`, `Dimensions`, `Not`, `Or`, `Tags`, unused ones explicitly `null`) and tag keys in Cost Explorer's prefixed spelling (`user:<key>` for your cost-allocation tags, `aws:<key>` for AWS-generated ones). A sparser or unprefixed document deploys fine — and then proposes a replacement on every subsequent plan, because the provider stores the re-marshaled form and CE echoes tag keys back prefixed.

**Frequency pairs with the channel** — IMMEDIATE subscriptions deliver individual alerts and every subscriber must be SNS; DAILY and WEEKLY summaries deliver via email only. The spec validates the pairing at manifest time instead of letting AWS reject it mid-apply.

**Impact thresholds are the noise dial** — without a `thresholdExpression`, every anomaly the monitor flags alerts. Compose `ANOMALY_TOTAL_IMPACT_ABSOLUTE` (dollars) with `ANOMALY_TOTAL_IMPACT_PERCENTAGE` (percent above normal) under `and` for "at least 100 USD AND at least 10% above normal" — the pairing that filters both trivial blips and proportionally-small noise on large bills.

**SNS topics need the right publish policy** — the topic must allow `costalerts.amazonaws.com` to publish, or immediate alerts are silently lost.

**Fresh monitors train silently** — roughly ten days of history before anything gets flagged. No alerts in the first days is expected, not broken.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsSnsTopic** | `subscriptions[].subscribers[].address` (type SNS) | `status.outputs.topic_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `monitor_arn` | The monitor's ARN (also the provider's import ID) | Externally managed subscriptions binding to this monitor; IAM statements scoping who may edit it |

`subscription_arns` is also echoed — each subscription's ARN keyed by its name. It exists for import addressing rather than as a composition input.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**The account baseline** — one DIMENSIONAL/SERVICE monitor per account (imported where AWS already created it), with a daily email digest thresholded to anomalies worth reading. Per-service segmentation means an EC2 spike and an S3 spike arrive as separate stories. Start from the **Service Spend Monitor** preset.

**A team's own anomaly stream** — a CUSTOM monitor over one cost-allocation tag with IMMEDIATE alerts into SNS (and from there, chat or incident tooling), noise-filtered by a composed absolute-and-percentage threshold. The slice's owner hears about their own anomalies without FinOps triaging in the middle. Start from the **Custom Slice Monitor** preset.

**Anomalies plus budgets, not either-or** — a budget catches "we spent more than we planned"; an anomaly monitor catches "this spend is abnormal for us" even while comfortably under budget. Accounts serious about cost run both against the same slices.

## Works With

- [**AWS SNS Topic**](/infra-catalog/aws-sns-topic) — the delivery channel for IMMEDIATE alerts, wired via the subscriber reference
- [**AWS Cost Category**](/infra-catalog/aws-cost-category) — team/project groupings a CUSTOM monitor's expression can slice by
- [**AWS Budget**](/infra-catalog/aws-budget) — the complementary fixed-threshold guardrail for planned spend ceilings
