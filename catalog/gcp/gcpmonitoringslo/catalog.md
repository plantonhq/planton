# GCP Monitoring SLO

Creates a Cloud Monitoring service-level objective — the formal reliability target ("99.9% of requests succeed, measured over a rolling 30 days") that error budgets, burn-rate alerts, and SRE dashboards are built on. One kind covers the whole story: it can measure an existing Monitoring service, or create the custom/basic service it measures in the same apply.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **SLO** -- a `monitoring.Slo` with the configured goal, period, and service-level indicator
- **Monitoring service** (optional, count-gated) -- a `monitoring.CustomService` or `monitoring.GenericService` when the spec's service arm asks for one
- **Monitoring API enablement** -- `monitoring.googleapis.com` enabled in the target project (never disabled on destroy)
- **GCP Labels** -- resource metadata labels applied automatically as `user_labels` on the SLO and any created service

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project.
- **Planton Runner** -- required when using Runner-based credential delivery.

### GCP Project

- **A GCP project** where the SLO is created (directly or via a GcpProject reference).
- **IAM**: the deploying identity needs `roles/monitoring.editor` or broader.
- **The metrics the SLI filters reference must exist** (or begin existing) for the SLO to measure anything.

## Deploy

### Console

Open the deployment store, find **GCP Monitoring SLO**, and click **Deploy**. The creation wizard walks you through the target project, the service arm, the goal and measurement period, and the service-level indicator. Start from the **Availability SLO** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpMonitoringSlo
metadata:
  name: checkout-availability
  org: acme-corp
  env: prod
spec:
  service:
    customService:
      displayName: Checkout
  goal: 0.999
  rollingPeriodDays: 30
  sli:
    requestBasedSli:
      goodTotalRatio:
        goodServiceFilter: metric.type="logging.googleapis.com/user/checkout/success" resource.type="global"
        totalServiceFilter: metric.type="logging.googleapis.com/user/checkout/requests" resource.type="global"
```

```shell
planton apply -f slo.yaml
```

Three nines of good checkouts, measured over a rolling 30 days, on a custom service created in the same apply. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, the SLO references its project via ValueFromRef:

```yaml
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: observability-project
      fieldPath: status.outputs.project_id
  service:
    customService:
      displayName: Checkout
  goal: 0.999
  rollingPeriodDays: 30
```

The InfraPipeline deploys the project first, then creates the service and SLO in it.

## Key Configuration

These are the most important decisions when configuring an SLO. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**service** -- exactly one arm: `serviceId` (an existing/auto-detected service), `customService` (create a blank-slate container — the right arm for anything GCP does not auto-detect), or `basicService` (create from a well-known type + labels, e.g. `CLOUD_RUN`).

**goal + period** -- the target fraction (greater than 0, at most 0.9999 — the API refuses five nines) over `calendarPeriod` (DAY/WEEK/FORTNIGHT/MONTH, budget resets at boundaries) or `rollingPeriodDays` (1–30, the classic SRE form).

**sli** -- exactly one family: `basicSli` (GCP derives availability/latency from service telemetry — auto-detected service types only), `requestBasedSli` (good/total ratio or latency-distribution cut from metric filters — the workhorse), or `windowsBasedSli` (time windows judged good/bad as a whole — "no bad minutes" objectives).

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** (optional) | `projectId` | `status.outputs.project_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `slo_name` | `projects/{p}/services/{s}/serviceLevelObjectives/{id}` | Burn-rate alert conditions (`select_slo_burn_rate`) |
| `service_name` | `projects/{p}/services/{s}` | Service-scoped dashboards and SLO siblings |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Availability SLO** -- good/total request ratio on a custom service over a rolling 30 days. Start from the **Availability SLO** preset.

**Latency SLO** -- the fraction of requests under a latency bound, from a distribution metric, over a calendar month. Start from the **Latency SLO** preset.

## Works With

- [**GCP Monitoring Alert Policy**](/infra-catalog/gcp-monitoring-alert-policy) -- burn-rate alerts are how an SLO pages someone
- [**GCP Log Metric**](/infra-catalog/gcp-log-metric) -- log-based metrics feed request-based SLIs
- [**GCP Project**](/infra-catalog/gcp-project) -- provides the GCP project where the SLO is created
