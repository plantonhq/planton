# AWS CloudWatch Dashboard

Deploys a CloudWatch dashboard — a named canvas of metric graphs, Logs Insights queries, alarm status tiles, and markdown panels your team opens first in an incident. The dashboard's identity is its name; every change is an in-place PutDashboard upsert (AWS has no separate create and update calls), so widget edits apply without replacing anything. Dashboards are untaggable at AWS — no tags argument exists on the resource — so the catalog's usual metadata-derived tags deliberately do not apply here.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **CloudWatch Dashboard** — named by `dashboardName`, carrying the full widget layout from `dashboardBody`. Any widget type CloudWatch supports lands on the 24-column grid: metric graphs, Logs Insights queries, alarm status tiles, text/markdown panels.

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with credentials for the target AWS account, carrying `cloudwatch:PutDashboard`, `cloudwatch:GetDashboard`, and `cloudwatch:DeleteDashboards`. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- Nothing needs to pre-exist. Metric widgets charting a metric that does not exist yet render an empty graph, never an error — the dashboard can ship ahead of the services it observes, and graphs populate as traffic arrives.

## Deploy

### Console

Open the deployment store, find **AWS CloudWatch Dashboard**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and the spec fields — the region, the dashboard name, and the widget body. Start from the **Service Health Dashboard** or **Alarm Overview Dashboard** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsCloudwatchDashboard
metadata:
  name: checkout-health
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  dashboardName: CheckoutHealth
  dashboardBody:
    widgets:
      - type: text
        x: 0
        "y": 0
        width: 24
        height: 2
        properties:
          markdown: "# Checkout service health"
      - type: metric
        x: 0
        "y": 2
        width: 12
        height: 6
        properties:
          metrics:
            - ["AWS/Lambda", "Errors", "FunctionName", "checkout"]
          period: 300
          stat: Sum
          region: us-east-1
          title: Lambda errors
```

```shell
planton apply -f cloudwatch-dashboard.yaml
```

This creates a two-widget dashboard named `CheckoutHealth` in us-east-1 — a markdown header and a Lambda error graph. An Infra Job tracks the provisioning in real time.

## Key Configuration

These are the most important decisions when configuring a dashboard. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Widget keys write naturally** — Manifests parse under YAML 1.2 rules, so the widget position key `y` (like `yes`, `no`, `on`, `off`) is an ordinary string with or without quotes; only `true` and `false` are booleans. Pasting the console's JSON verbatim is always safe: JSON keys are quoted by definition.

**Prototype in the console, own it in the manifest** — The fastest authoring loop is building the dashboard visually in the CloudWatch console, opening Actions → View/edit source, and pasting the document into `dashboardBody`. From then on the manifest is the source of truth; every apply is an idempotent upsert, so console-side edits show as drift and the next apply restores the manifest's layout.

**The name is the identity** — `dashboardName` is a separate field because dashboard names allow uppercase letters `metadata.name` cannot (`ServiceHealth`). Changing the name replaces the dashboard; everything else — the entire widget body — updates in place.

**The body diffs semantically** — AWS normalizes the document server-side and both engines diff it as JSON, so key order and whitespace never show as drift. A plan showing a body change you did not make means someone edited the dashboard in the console.

**Region scoping** — Dashboards are region-scoped objects (each region's console lists its own set), but their widgets may chart metrics from any region: each metric widget carries its own `region` property. One dashboard in your home region can be the cross-region single pane.

**Dashboard count is the cost driver** — AWS bills per dashboard above the account's free allotment, so the lever is how many dashboards exist, not how many widgets they carry. Prefer one dense dashboard per service over many sparse ones; short-lived test dashboards cost effectively nothing.

## Outputs and Dependencies

### What This Kind Consumes

This kind has no foreign key dependencies. Widgets reference metrics by namespace and dimension names and alarms by ARN, all inside the `dashboardBody` document — these travel as plain strings, not typed references.

### What This Kind Provides

After provisioning, `status.outputs` records the dashboard's identity: `dashboard_name` (the provider's import ID) and `dashboard_arn`. Nothing downstream composes on a dashboard — these outputs exist for auditing and import, not for ValueFromRef wiring.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**One dashboard per service** — a markdown header naming the service and its runbook, error-count and p95-latency graphs side by side, rows added as the service grows. Every edit applies in place, so the dashboard evolves with the service. Start from the **Service Health Dashboard** preset.

**The on-call landing page** — an alarm-status strip across the top wired to the service's AwsCloudwatchAlarm ARNs, with a Logs Insights error tail beneath it. One glance answers "what is firing and what is it logging." Start from the **Alarm Overview Dashboard** preset.

**Ship the dashboard with the stack** — because metric widgets tolerate nonexistent metrics, the dashboard can deploy in the same change as the service it charts. Graphs stay empty until traffic arrives, then populate — no ordering dependency to manage.

## Works With

- [**AWS CloudWatch Alarm**](/infra-catalog/aws-cloudwatch-alarm) — alarm status widgets reference alarm ARNs; the dashboard is where their state is read at a glance
- [**AWS CloudWatch Log Group**](/infra-catalog/aws-cloudwatch-log-group) — Logs Insights query widgets tail and aggregate a log group's events on the same canvas
- [**AWS CloudWatch Synthetics**](/infra-catalog/aws-cloudwatch-synthetics) — canary success-rate and latency metrics are natural widgets on a service health dashboard
