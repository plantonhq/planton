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
- **IAM**: see [`iac/permissions.yaml`](iac/permissions.yaml) for the least-privilege permission set the deploying identity needs.
- **The metrics the SLI filters reference must exist** (or begin existing) for the SLO to measure anything — filters are validated syntactically at create, not against live data.

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpMonitoringSlo
metadata:
  name: checkout-availability
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

## The three service arms

- **`serviceId`** -- measure an EXISTING Monitoring service (GCP auto-detects services for App Engine, Istio canonical services, and friends).
- **`customService`** -- create a blank-slate service container; the SLI's metric filters define what "the service" means. The right arm for anything GCP does not auto-detect.
- **`basicService`** -- create a service from a well-known type + labels (e.g. `CLOUD_RUN` + service_name/location); GCP wires the telemetry association.

## Outputs

| Output | Description |
|--------|-------------|
| `slo_name` | `projects/{p}/services/{s}/serviceLevelObjectives/{id}` — the burn-rate alert handle |
| `service_name` | `projects/{p}/services/{s}` — the measured service |

## Works With

- **GcpMonitoringAlertPolicy** -- burn-rate alerts filter on `select_slo_burn_rate("{slo_name}", ...)` — the pairing that turns an SLO into pages
- **GcpLogMetric** -- log-based metrics feed request-based SLIs for services whose truth lives in logs
- **GcpProject** -- provides the GCP project where the SLO is created

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
