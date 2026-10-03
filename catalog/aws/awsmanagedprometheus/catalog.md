# AWS Managed Prometheus

Deploys an Amazon Managed Prometheus (AMP) workspace — a PromQL-compatible metrics backend your clusters remote-write into — with retention controls, recording and alerting rule namespaces, a managed Alertmanager, query logging, a cross-account resource policy, and PromQL anomaly detectors. Scrapers are deliberately a separate kind (AwsManagedPrometheusScraper) that references this workspace's ARN, because a scraper can exist with zero workspaces. Two AWS contracts shape the lifecycle before you write a manifest: a workspace alias can never be unset once set (clearing it replaces the workspace), and the workspace configuration persists after destroy (AWS exposes no delete for it).

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Prometheus Workspace** — the AMP workspace with optional alias, optional customer-managed KMS encryption, and workspace event logging to CloudWatch when `logging` is set
- **Workspace Configuration** — created only when `configuration` is set; metric retention, out-of-order ingest window, rule query offset, and per-label-set active-series caps. AWS creates this object via update and offers no delete: removing the block later leaves the last-applied values in place
- **Alert Manager Definition** — created only when `alertManagerDefinition` is set; the workspace's single alertmanager.yml document (SNS is the supported receiver)
- **Rule Group Namespaces** — one per `ruleGroupNamespaces` entry; each holds a standard Prometheus rules file with recording and alerting rules
- **Query Logging Configuration** — created only when `queryLogging` is set; logs queries above a Query Samples Processed threshold to CloudWatch log groups
- **Resource Policy** — created only when `resourcePolicy` is set; the IAM document granting other principals or accounts remote-write and query access. The module stamps each statement's Resource with the workspace's own ARN
- **Anomaly Detectors** — one per `anomalyDetectors` entry; Random Cut Forest models watching the series a PromQL query selects

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with permissions to manage AMP workspaces, plus CloudWatch Logs when the logging arms are set and KMS when encrypting with a customer-managed key. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- **CloudWatch log groups** (only for logging) — `logging.logGroupArn` and each `queryLogging` destination reference an AwsCloudwatchLogGroup's `log_group_arn` output or a literal group ARN. AWS requires a `:*` suffix on these ARNs; both engines append it when absent, so wire the bare output and never hand-craft the suffix.
- **An SNS topic** (only for alerting) — AMP's managed Alertmanager delivers exclusively to SNS, so the `alertManagerDefinition` needs a topic ARN to route to.
- **A KMS key** (only for customer-managed encryption) — `kmsKeyArn` references an AwsKmsKey's `key_arn` output.

## Deploy

### Console

Open the deployment store, find **AWS Managed Prometheus**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and the spec: region and alias first, then the optional configuration, rules, Alertmanager, logging, resource-policy, and anomaly-detector sections. Start from the **Platform Metrics Workspace** preset in the [Presets](#presets) tab for the EKS-fleet landing-zone shape.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsManagedPrometheus
metadata:
  name: platform-metrics
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  alias: platform-metrics
  configuration:
    retentionPeriodInDays: 90
  ruleGroupNamespaces:
    - name: platform-slos
      data: |
        groups:
          - name: request-rates
            rules:
              - record: job:http_requests:rate5m
                expr: sum(rate(http_requests_total[5m])) by (job)
  alertManagerDefinition: |
    alertmanager_config: |
      route:
        receiver: platform-alerts
      receivers:
        - name: platform-alerts
          sns_configs:
            - topic_arn: arn:aws:sns:us-east-1:123456789012:platform-alerts
              sigv4:
                region: us-east-1
```

```shell
planton apply -f aws-managed-prometheus.yaml
```

This creates a workspace with 90-day retention, one recording-rules namespace, and Alertmanager routing to the named SNS topic. An Infra Job tracks the provisioning in real time.

### InfraChart

When the workspace deploys alongside its log group in one chart, wire the logging reference via ValueFromRef:

```yaml
spec:
  region: us-east-1
  alias: platform-metrics
  logging:
    logGroupArn:
      valueFrom:
        kind: AwsCloudwatchLogGroup
        name: amp-workspace-events
        fieldPath: status.outputs.log_group_arn
```

The InfraPipeline resolves the dependency graph, creates the log group first, then provisions the workspace with event logging wired into it.

## Key Configuration

These are the most important decisions when configuring a workspace. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Never set an alias you might want to remove** — AWS offers no un-alias: once set, `alias` can change but never clear. Emptying the field replaces the workspace — new workspace ID, all stored metrics gone. Renaming in place is safe; removal is destruction.

**The configuration outlives your manifest** — the workspace configuration (retention, label-set limits) is created via update and has no delete API. Removing the `configuration` block is a no-op at AWS: the last-applied values persist. Set the values you want instead of removing the block and expecting defaults back. Left entirely unset, retention stays at AWS's 150-day default — and retention is the storage-cost lever, so 90 days on a busy fleet is a deliberate saving, not an oversight.

**Encryption is decided at create** — changing `kmsKeyArn` replaces the workspace. Choose between AWS-owned and customer-managed encryption before the first metric lands, not after a compliance review finds the gap.

**Alertmanager speaks SNS only** — a pasted alertmanager.yml with webhook receivers validates at AWS but never fires. Route through an SNS topic and fan out to email, Slack, or PagerDuty from there.

**Label-set limits are the noisy-neighbor control** — `limitsPerLabelSet` caps active series per label population (a per-team cap, for example). A `maxSeries` of 0 blocks ingestion for that set entirely — a kill switch for a misbehaving emitter, applied in place with no workspace disruption.

**One rule namespace, one owner** — each namespace's `data` replaces the whole rules file on every apply. Treat a namespace as one owned file (per team, per service); hand-merging rules from several owners into one namespace guarantees someone's rules get clobbered. Renaming a namespace replaces it.

**Resource policy statements carry no Resource member** — AMP requires every statement's Resource to be exactly this workspace's own ARN, so both engines compose it after the workspace exists; an authored Resource is replaced, never honored. Write a single action as a string (`"Action": "aps:RemoteWrite"`), not a one-element list — AMP's stored form collapses single-element arrays, and a list-authored policy diffs forever against that echo.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsKmsKey** | `kmsKeyArn` | `status.outputs.key_arn` |
| **AwsCloudwatchLogGroup** | `logging.logGroupArn` | `status.outputs.log_group_arn` |
| **AwsCloudwatchLogGroup** | `queryLogging.destinations[].logGroupArn` | `status.outputs.log_group_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `workspace_arn` | The workspace's ARN | The `destination` of an AwsManagedPrometheusScraper; IAM policies granting remote-write or query |
| `prometheus_endpoint` | The Prometheus-compatible query/remote-write endpoint URL | Remote-write agent configuration and Grafana data sources |
| `workspace_id` | The workspace's ID | Addressing the workspace in AWS API calls and CLI operations |

`rule_group_namespace_arns`, `anomaly_detector_ids`, and `anomaly_detector_arns` are also present — alias-keyed maps echoed for audit and import addressing, not composition inputs.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**EKS fleet landing zone** — one workspace per environment with 90-day retention, a recording-plus-alerting rules namespace, and Alertmanager routing to SNS. Pair it with an AwsManagedPrometheusScraper per cluster, or point in-cluster remote-write agents at `prometheus_endpoint`. Start from the **Platform Metrics Workspace** preset.

**Governed multi-team workspace** — a shared workspace with per-team active-series caps, expensive-query logging above a QSP threshold, and an anomaly detector watching the ingest rate itself, with missing data marked anomalous — silence from the fleet IS the incident. Start from the **Governed Multi-Team Workspace** preset.

**Cross-account ingest hub** — a central workspace whose `resourcePolicy` grants sibling accounts `aps:RemoteWrite`, collapsing per-account observability stacks into one query surface. Trades blast-radius isolation for a single pane of glass; the label-set caps above are what keep one account from starving the rest.

## Works With

- [**AWS Managed Prometheus Scraper**](/infra-catalog/aws-managed-prometheus-scraper) — the agentless collector that scrapes EKS or VPC targets into this workspace, wired via `workspace_arn`
- [**AWS CloudWatch Log Group**](/infra-catalog/aws-cloudwatch-log-group) — the destination for workspace event logging and query logging, wired via the `logGroupArn` references
- [**AWS KMS Key**](/infra-catalog/aws-kms-key) — customer-managed encryption for metric data via `kmsKeyArn`
- [**AWS SNS Topic**](/infra-catalog/aws-sns-topic) — the one receiver the managed Alertmanager delivers to; fan out to on-call tooling from there
- [**AWS EKS Cluster**](/infra-catalog/aws-eks-cluster) — the typical metrics source, scraped in-VPC or remote-writing through in-cluster agents
