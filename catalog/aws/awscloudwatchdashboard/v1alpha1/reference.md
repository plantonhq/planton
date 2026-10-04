# AwsCloudwatchDashboard

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsCloudwatchDashboardSpec defines one CloudWatch dashboard: a named
canvas of widgets (metric graphs, logs insights queries, alarms,
text) laid out by the dashboard body document.

Identity is the dashboard name; every change is an in-place
PutDashboard upsert (AWS has no separate create/update). Dashboards
are untaggable at AWS - there is no tags argument on the resource,
so the catalog's usual metadata-derived tags deliberately do not
apply here.

## Example

```yaml
# Canonical AwsCloudwatchDashboard example (hack/dev manifest and
# refgen Example source): a two-widget service dashboard (a markdown
# header and a Lambda-errors metric graph). The widget position key
# "y" is quoted for readability only - manifests parse under YAML 1.2
# rules, where a bare y is an ordinary string.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsCloudwatchDashboard
metadata:
  name: service-health
  id: service-health
  org: test-org
  env: dev
spec:
  region: us-west-2
  dashboardName: ServiceHealth
  dashboardBody:
    widgets:
      - type: text
        x: 0
        "y": 0
        width: 24
        height: 2
        properties:
          markdown: "# Service health"
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
          region: us-west-2
          title: Lambda errors
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.dashboardName` | `string` | yes |  |  |
| `spec.dashboardBody` | `object` | yes |  |  |

## Field Details

### spec.region

`string` · required

The AWS region used by the provider while managing the dashboard.
Dashboards are region-scoped objects (each region's console shows
its own set), though their widgets may chart metrics from any
region. Example: "us-east-1".

- rule: {"string":{"minLen":"1"}}

### spec.dashboardName

`string` · required

The dashboard's name in AWS - an explicit field because dashboard
names allow uppercase letters metadata.name cannot ("ServiceHealth").
Letters, digits, hyphens, and underscores, up to 255. Changing the
name replaces the dashboard.

- rule: {"string":{"minLen":"1","maxLen":"255","pattern":"^[0-9A-Za-z_-]+$"}}

### spec.dashboardBody

`object` · required

The dashboard body: the widget layout document with a "widgets"
array, exactly as the CloudWatch console's Actions -> View/edit
source shows it. Each widget carries its type (metric / log /
alarm / text), position (x, y, width, height), and properties.
AWS normalizes the JSON server-side; both engines diff it
semantically, so key order and whitespace never cause drift.

YAML authors: manifests parse under YAML 1.2 rules, so the widget
position key y (like yes, no, on, off) is an ordinary string with
or without quotes - only true and false are booleans. Pasting the
console's JSON body verbatim is always safe (JSON keys are quoted
by definition).

- rule: {"required":true}

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsCloudwatchDashboard, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.dashboard_name` | `string` | The dashboard's name (the provider's import ID). |
| `status.outputs.dashboard_arn` | `string` | The dashboard's ARN. |

## See Also

- [Overview](../README.md)
