# GcpMonitoringSlo

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

GcpMonitoringSloSpec defines a Cloud Monitoring service-level objective
(SLO) — the formal reliability target ("99.9% of requests succeed,
measured over a rolling 30 days") that error budgets, burn-rate alerts,
and SRE dashboards are built on.

An SLO is built from three parts:
  1. service -- WHAT is being measured. SLOs always live under a
     Monitoring "service": an existing one (GCP auto-detects services for
     App Engine, Istio, and friends), a custom service this kind creates,
     or a basic service this kind creates from a type + labels.
  2. sli     -- HOW good events are counted (the service-level
     indicator): a request success ratio, a latency distribution cut, or
     windowed criteria.
  3. goal + period -- the TARGET: the fraction of good events required,
     over a calendar period or a rolling window.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpMonitoringSlo
metadata:
  name: my-sample-slo
spec:
  # The Monitoring service the SLO measures — exactly one arm:
  # serviceId (an existing/auto-detected service), customService (create
  # a blank-slate container — the right arm for anything GCP does not
  # auto-detect), or basicService (create from a well-known type +
  # labels). The created service's id defaults to metadata.name.
  service:
    customService:
      displayName: Checkout

  # The target fraction of good service: three nines. Greater than 0, at
  # most 0.9999 (the GCP API refuses five nines).
  goal: 0.999

  # Measured over a rolling 30 days (the classic SRE form). The
  # alternative is calendarPeriod: DAY | WEEK | FORTNIGHT | MONTH —
  # exactly one of the two.
  rollingPeriodDays: 30

  # How good service is counted — exactly one SLI family. The good/total
  # ratio takes exactly TWO of good/bad/total filters; GCP derives the
  # third. Every filter must pin resource.type alongside metric.type:
  # GCP validates at create that a filter parses to exactly ONE monitored
  # resource type and rejects anything broader.
  sli:
    requestBasedSli:
      goodTotalRatio:
        goodServiceFilter: metric.type="serviceruntime.googleapis.com/api/request_count" resource.type="consumed_api" metric.labels.response_code_class="2xx"
        totalServiceFilter: metric.type="serviceruntime.googleapis.com/api/request_count" resource.type="consumed_api"

  # What a destroy does: DELETE (default — the error-budget history dies
  # with the SLO), PREVENT (the posture once burn-rate alerts reference
  # it), or ABANDON. Also covers any service this kind created.
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.service` | `GcpMonitoringSloService` | yes |  |  |
| `spec.service.serviceId` | `string` |  |  |  |
| `spec.service.customService` | `GcpMonitoringSloCustomService` |  |  |  |
| `spec.service.customService.serviceId` | `string` |  |  |  |
| `spec.service.customService.displayName` | `string` |  |  |  |
| `spec.service.customService.telemetryResourceName` | `string` |  |  |  |
| `spec.service.basicService` | `GcpMonitoringSloBasicService` |  |  |  |
| `spec.service.basicService.serviceId` | `string` |  |  |  |
| `spec.service.basicService.serviceType` | `string` |  |  |  |
| `spec.service.basicService.serviceLabels` | `map<string, string>` |  |  |  |
| `spec.sloId` | `string` |  |  |  |
| `spec.displayName` | `string` |  |  |  |
| `spec.goal` | `double` | yes |  |  |
| `spec.calendarPeriod` | `string` |  |  |  |
| `spec.rollingPeriodDays` | `int32` |  |  |  |
| `spec.sli` | `GcpMonitoringSloSli` | yes |  |  |
| `spec.sli.basicSli` | `GcpMonitoringSloBasicSli` |  |  |  |
| `spec.sli.basicSli.location` | `[]string` |  |  |  |
| `spec.sli.basicSli.method` | `[]string` |  |  |  |
| `spec.sli.basicSli.version` | `[]string` |  |  |  |
| `spec.sli.basicSli.availability` | `GcpMonitoringSloAvailability` |  |  |  |
| `spec.sli.basicSli.availability.enabled` | `bool` |  | `true` |  |
| `spec.sli.basicSli.latency` | `GcpMonitoringSloLatency` |  |  |  |
| `spec.sli.basicSli.latency.threshold` | `string` | yes |  |  |
| `spec.sli.requestBasedSli` | `GcpMonitoringSloRequestBasedSli` |  |  |  |
| `spec.sli.requestBasedSli.distributionCut` | `GcpMonitoringSloDistributionCut` |  |  |  |
| `spec.sli.requestBasedSli.distributionCut.distributionFilter` | `string` | yes |  |  |
| `spec.sli.requestBasedSli.distributionCut.range` | `GcpMonitoringSloRange` |  |  |  |
| `spec.sli.requestBasedSli.distributionCut.range.min` | `double` |  |  |  |
| `spec.sli.requestBasedSli.distributionCut.range.max` | `double` |  |  |  |
| `spec.sli.requestBasedSli.goodTotalRatio` | `GcpMonitoringSloGoodTotalRatio` |  |  |  |
| `spec.sli.requestBasedSli.goodTotalRatio.goodServiceFilter` | `string` |  |  |  |
| `spec.sli.requestBasedSli.goodTotalRatio.badServiceFilter` | `string` |  |  |  |
| `spec.sli.requestBasedSli.goodTotalRatio.totalServiceFilter` | `string` |  |  |  |
| `spec.sli.windowsBasedSli` | `GcpMonitoringSloWindowsBasedSli` |  |  |  |
| `spec.sli.windowsBasedSli.windowPeriod` | `string` |  |  |  |
| `spec.sli.windowsBasedSli.goodBadMetricFilter` | `string` |  |  |  |
| `spec.sli.windowsBasedSli.goodTotalRatioThreshold` | `GcpMonitoringSloGoodTotalRatioThreshold` |  |  |  |
| `spec.sli.windowsBasedSli.goodTotalRatioThreshold.threshold` | `double` |  |  |  |
| `spec.sli.windowsBasedSli.goodTotalRatioThreshold.basicSliPerformance` | `GcpMonitoringSloBasicSli` |  |  |  |
| `spec.sli.windowsBasedSli.goodTotalRatioThreshold.basicSliPerformance.location` | `[]string` |  |  |  |
| `spec.sli.windowsBasedSli.goodTotalRatioThreshold.basicSliPerformance.method` | `[]string` |  |  |  |
| `spec.sli.windowsBasedSli.goodTotalRatioThreshold.basicSliPerformance.version` | `[]string` |  |  |  |
| `spec.sli.windowsBasedSli.goodTotalRatioThreshold.basicSliPerformance.availability` | `GcpMonitoringSloAvailability` |  |  |  |
| `spec.sli.windowsBasedSli.goodTotalRatioThreshold.basicSliPerformance.availability.enabled` | `bool` |  | `true` |  |
| `spec.sli.windowsBasedSli.goodTotalRatioThreshold.basicSliPerformance.latency` | `GcpMonitoringSloLatency` |  |  |  |
| `spec.sli.windowsBasedSli.goodTotalRatioThreshold.basicSliPerformance.latency.threshold` | `string` | yes |  |  |
| `spec.sli.windowsBasedSli.goodTotalRatioThreshold.performance` | `GcpMonitoringSloPerformance` |  |  |  |
| `spec.sli.windowsBasedSli.goodTotalRatioThreshold.performance.distributionCut` | `GcpMonitoringSloDistributionCut` |  |  |  |
| `spec.sli.windowsBasedSli.goodTotalRatioThreshold.performance.distributionCut.distributionFilter` | `string` | yes |  |  |
| `spec.sli.windowsBasedSli.goodTotalRatioThreshold.performance.distributionCut.range` | `GcpMonitoringSloRange` |  |  |  |
| `spec.sli.windowsBasedSli.goodTotalRatioThreshold.performance.distributionCut.range.min` | `double` |  |  |  |
| `spec.sli.windowsBasedSli.goodTotalRatioThreshold.performance.distributionCut.range.max` | `double` |  |  |  |
| `spec.sli.windowsBasedSli.goodTotalRatioThreshold.performance.goodTotalRatio` | `GcpMonitoringSloGoodTotalRatio` |  |  |  |
| `spec.sli.windowsBasedSli.goodTotalRatioThreshold.performance.goodTotalRatio.goodServiceFilter` | `string` |  |  |  |
| `spec.sli.windowsBasedSli.goodTotalRatioThreshold.performance.goodTotalRatio.badServiceFilter` | `string` |  |  |  |
| `spec.sli.windowsBasedSli.goodTotalRatioThreshold.performance.goodTotalRatio.totalServiceFilter` | `string` |  |  |  |
| `spec.sli.windowsBasedSli.metricMeanInRange` | `GcpMonitoringSloMetricRange` |  |  |  |
| `spec.sli.windowsBasedSli.metricMeanInRange.timeSeries` | `string` | yes |  |  |
| `spec.sli.windowsBasedSli.metricMeanInRange.range` | `GcpMonitoringSloRange` |  |  |  |
| `spec.sli.windowsBasedSli.metricMeanInRange.range.min` | `double` |  |  |  |
| `spec.sli.windowsBasedSli.metricMeanInRange.range.max` | `double` |  |  |  |
| `spec.sli.windowsBasedSli.metricSumInRange` | `GcpMonitoringSloMetricRange` |  |  |  |
| `spec.sli.windowsBasedSli.metricSumInRange.timeSeries` | `string` | yes |  |  |
| `spec.sli.windowsBasedSli.metricSumInRange.range` | `GcpMonitoringSloRange` |  |  |  |
| `spec.sli.windowsBasedSli.metricSumInRange.range.min` | `double` |  |  |  |
| `spec.sli.windowsBasedSli.metricSumInRange.range.max` | `double` |  |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project that owns the SLO (and any service this kind creates).
Can be a literal project ID or a reference to a GcpProject resource.
If omitted, the provider's default project is used.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.service

`GcpMonitoringSloService` · required

The Monitoring service this SLO measures — exactly one arm. Changing
the resolved service REPLACES the SLO (the GCP API binds an SLO to its
service for life).

- rule: {"required":true}
- rule: set exactly one of: service_id (an existing service), custom_service, or basic_service

### spec.service.serviceId

`string`

Measure an EXISTING Monitoring service — one GCP auto-detected (App
Engine, Istio canonical services, and friends) or one created outside
this kind. The service ID is the last segment of the service resource
name (projects/{project}/services/{service_id}).

### spec.service.customService

`GcpMonitoringSloCustomService`

CREATE a custom service and measure it. The blank-slate form: a custom
service is just a named container for SLOs — the SLIs point at
whatever metrics define the service.

### spec.service.customService.serviceId

`string`

The service ID (the last segment of the service resource name).
Defaults to metadata.name. Changing it REPLACES the service and the
SLO under it.

### spec.service.customService.displayName

`string`

Display name shown in the Monitoring services list. Defaults to
metadata.name.

### spec.service.customService.telemetryResourceName

`string`

The full resource name of the underlying workload this service
represents (e.g. //run.googleapis.com/projects/{p}/locations/{l}/services/{s})
— how the console links the service to its resource. Optional.

### spec.service.basicService

`GcpMonitoringSloBasicService`

CREATE a basic service from a service type + labels (e.g. a
CLOUD_RUN service named by its service_name and location) and measure
it. GCP wires the telemetry association from the labels.

### spec.service.basicService.serviceId

`string`

The service ID (the last segment of the service resource name).
Defaults to metadata.name. Changing it REPLACES the service and the
SLO under it.

### spec.service.basicService.serviceType

`string`

The service type, e.g. APP_ENGINE, CLOUD_ENDPOINTS, CLUSTER_ISTIO,
ISTIO_CANONICAL_SERVICE, CLOUD_RUN. The GCP API validates the type
and its required labels server-side (the provider carries no
client-side list — new types appear with new products).

### spec.service.basicService.serviceLabels

`map<string, string>`

The labels that identify the concrete service instance for the chosen
type — e.g. for CLOUD_RUN: {"service_name": "checkout",
"location": "us-central1"}. Which keys are required is defined by the
service type (server-validated). Changing labels REPLACES the service.

### spec.sloId

`string`

The SLO identifier within the service (the last path segment of the
SLO resource name). Omit to let GCP assign one. Changing it REPLACES
the SLO.

- rule: slo_id may contain only letters, numerals, and the characters - _ : .

### spec.displayName

`string`

Human-friendly name shown in the SLO list and on dashboards. Defaults
to metadata.name when left empty.

### spec.goal

`double` · required

The target fraction of good service the SLO demands, e.g. 0.999 for
"three nines". Must be greater than 0 and at most 0.9999 (the GCP
API's own bound — five nines and beyond are not accepted).

- rule: goal must be greater than 0 and at most 0.9999
- rule: {"required":true}

### spec.calendarPeriod

`string`

Measure over a calendar period: DAY, WEEK, FORTNIGHT, or MONTH. The
error budget resets at each period boundary. Set exactly one of this
and rolling_period_days.

- rule: calendar_period must be one of: DAY, WEEK, FORTNIGHT, MONTH

### spec.rollingPeriodDays

`int32`

Measure over a rolling window of this many days (1 to 30 — the GCP
API's bounds). The classic SRE form is 28 or 30. Set exactly one of
this and calendar_period.

- rule: rolling_period_days must be between 1 and 30

### spec.sli

`GcpMonitoringSloSli` · required

The service-level indicator — how good service is counted. Exactly one
SLI family.

- rule: {"required":true}
- rule: set exactly one of: basic_sli, request_based_sli, or windows_based_sli

### spec.sli.basicSli

`GcpMonitoringSloBasicSli`

A basic SLI: GCP derives availability or latency from the service's
own telemetry — no filters to write. Only works for service types
whose telemetry GCP understands natively (App Engine, Cloud
Endpoints, Istio); custom services need request_based_sli instead.

- rule: set exactly one of availability or latency

### spec.sli.basicSli.location

`[]string`

Narrow the SLI to these locations (e.g. specific regions). Empty means
all locations.

### spec.sli.basicSli.method

`[]string`

Narrow the SLI to these RPC methods. Empty means all methods.

### spec.sli.basicSli.version

`[]string`

Narrow the SLI to these API versions. Empty means all versions.

### spec.sli.basicSli.availability

`GcpMonitoringSloAvailability`

Count availability (successful vs total requests) as good service.

### spec.sli.basicSli.availability.enabled

`bool` · optional (explicit presence)

Whether the availability SLI is enabled (the GCP API expects true —
the field exists for API-shape fidelity). Both IaC engines send the
value explicitly so behavior is identical regardless of engine.

- default: `true`

### spec.sli.basicSli.latency

`GcpMonitoringSloLatency`

Count requests faster than a threshold as good service.

### spec.sli.basicSli.latency.threshold

`string` · required

The latency threshold below which a response counts as good, as a
duration string (e.g. "1s", "0.5s").

- rule: {"required":true}

### spec.sli.requestBasedSli

`GcpMonitoringSloRequestBasedSli`

A request-based SLI: good service counted from metric filters — a
good/total ratio or a latency-distribution cut. The workhorse form
for custom services.

- rule: set exactly one of distribution_cut or good_total_ratio

### spec.sli.requestBasedSli.distributionCut

`GcpMonitoringSloDistributionCut`

Good service = values of a distribution metric falling inside a range
(the latency-SLO form: requests whose latency lands in [0, 500ms]).

### spec.sli.requestBasedSli.distributionCut.distributionFilter

`string` · required

A monitoring filter (https://cloud.google.com/monitoring/api/v3/filters)
selecting a DISTRIBUTION-valued metric, e.g. request latencies.
The filter MUST pin resource.type alongside metric.type: GCP validates
at create that every SLO filter parses to exactly ONE monitored
resource type and rejects anything broader (Error 400 "parses to N
resource types and must parse to 1").

- rule: {"required":true}

### spec.sli.requestBasedSli.distributionCut.range

`GcpMonitoringSloRange`

The range of values counted as good (e.g. min 0, max 500 for "under
500ms" when the metric is in milliseconds).

### spec.sli.requestBasedSli.distributionCut.range.min

`double` · optional (explicit presence)

The lower bound of the range. Unset means unbounded below.

### spec.sli.requestBasedSli.distributionCut.range.max

`double` · optional (explicit presence)

The upper bound of the range. Unset means unbounded above.

### spec.sli.requestBasedSli.goodTotalRatio

`GcpMonitoringSloGoodTotalRatio`

Good service = ratio of a "good" counter to a "total" counter (or
good vs bad).

- rule: set exactly two of: good_service_filter, bad_service_filter, total_service_filter (GCP derives the third)

### spec.sli.requestBasedSli.goodTotalRatio.goodServiceFilter

`string`

A monitoring filter counting GOOD events (a DELTA metric of int64 or
double type).

### spec.sli.requestBasedSli.goodTotalRatio.badServiceFilter

`string`

A monitoring filter counting BAD events.

### spec.sli.requestBasedSli.goodTotalRatio.totalServiceFilter

`string`

A monitoring filter counting TOTAL events.

### spec.sli.windowsBasedSli

`GcpMonitoringSloWindowsBasedSli`

A windows-based SLI: time is divided into windows and each window is
judged good or bad as a whole (e.g. "a window is good when p95
latency stays under 500ms"). For "no bad minutes" style objectives.

- rule: set exactly one window criterion: good_bad_metric_filter, good_total_ratio_threshold, metric_mean_in_range, or metric_sum_in_range

### spec.sli.windowsBasedSli.windowPeriod

`string`

The window length, as a duration string between "60s" and "604800s"
(one minute to one week). Empty defers to the GCP API default (60s).

### spec.sli.windowsBasedSli.goodBadMetricFilter

`string`

A window is good when this filter's BOOLEAN metric is true throughout
the window. Like every SLO filter, it must pin resource.type
alongside metric.type (GCP requires exactly one monitored resource
type per filter).

### spec.sli.windowsBasedSli.goodTotalRatioThreshold

`GcpMonitoringSloGoodTotalRatioThreshold`

A window is good when a request-based criterion (a good/total ratio
or a basic SLI) meets a threshold within the window.

- rule: set exactly one of basic_sli_performance or performance

### spec.sli.windowsBasedSli.goodTotalRatioThreshold.threshold

`double`

The window is good when the criterion's ratio meets or exceeds this
threshold (0 to 1).

- rule: threshold must be between 0 and 1

### spec.sli.windowsBasedSli.goodTotalRatioThreshold.basicSliPerformance

`GcpMonitoringSloBasicSli`

Judge each window with a basic SLI (availability or latency derived
from service telemetry).

- rule: set exactly one of availability or latency

### spec.sli.windowsBasedSli.goodTotalRatioThreshold.basicSliPerformance.location

`[]string`

Narrow the SLI to these locations (e.g. specific regions). Empty means
all locations.

### spec.sli.windowsBasedSli.goodTotalRatioThreshold.basicSliPerformance.method

`[]string`

Narrow the SLI to these RPC methods. Empty means all methods.

### spec.sli.windowsBasedSli.goodTotalRatioThreshold.basicSliPerformance.version

`[]string`

Narrow the SLI to these API versions. Empty means all versions.

### spec.sli.windowsBasedSli.goodTotalRatioThreshold.basicSliPerformance.availability

`GcpMonitoringSloAvailability`

Count availability (successful vs total requests) as good service.

### spec.sli.windowsBasedSli.goodTotalRatioThreshold.basicSliPerformance.availability.enabled

`bool` · optional (explicit presence)

Whether the availability SLI is enabled (the GCP API expects true —
the field exists for API-shape fidelity). Both IaC engines send the
value explicitly so behavior is identical regardless of engine.

- default: `true`

### spec.sli.windowsBasedSli.goodTotalRatioThreshold.basicSliPerformance.latency

`GcpMonitoringSloLatency`

Count requests faster than a threshold as good service.

### spec.sli.windowsBasedSli.goodTotalRatioThreshold.basicSliPerformance.latency.threshold

`string` · required

The latency threshold below which a response counts as good, as a
duration string (e.g. "1s", "0.5s").

- rule: {"required":true}

### spec.sli.windowsBasedSli.goodTotalRatioThreshold.performance

`GcpMonitoringSloPerformance`

Judge each window with a request-based criterion (distribution cut or
good/total ratio).

- rule: set exactly one of distribution_cut or good_total_ratio

### spec.sli.windowsBasedSli.goodTotalRatioThreshold.performance.distributionCut

`GcpMonitoringSloDistributionCut`

Distribution-cut criterion for the window.

### spec.sli.windowsBasedSli.goodTotalRatioThreshold.performance.distributionCut.distributionFilter

`string` · required

A monitoring filter (https://cloud.google.com/monitoring/api/v3/filters)
selecting a DISTRIBUTION-valued metric, e.g. request latencies.
The filter MUST pin resource.type alongside metric.type: GCP validates
at create that every SLO filter parses to exactly ONE monitored
resource type and rejects anything broader (Error 400 "parses to N
resource types and must parse to 1").

- rule: {"required":true}

### spec.sli.windowsBasedSli.goodTotalRatioThreshold.performance.distributionCut.range

`GcpMonitoringSloRange`

The range of values counted as good (e.g. min 0, max 500 for "under
500ms" when the metric is in milliseconds).

### spec.sli.windowsBasedSli.goodTotalRatioThreshold.performance.distributionCut.range.min

`double` · optional (explicit presence)

The lower bound of the range. Unset means unbounded below.

### spec.sli.windowsBasedSli.goodTotalRatioThreshold.performance.distributionCut.range.max

`double` · optional (explicit presence)

The upper bound of the range. Unset means unbounded above.

### spec.sli.windowsBasedSli.goodTotalRatioThreshold.performance.goodTotalRatio

`GcpMonitoringSloGoodTotalRatio`

Good/total ratio criterion for the window.

- rule: set exactly two of: good_service_filter, bad_service_filter, total_service_filter (GCP derives the third)

### spec.sli.windowsBasedSli.goodTotalRatioThreshold.performance.goodTotalRatio.goodServiceFilter

`string`

A monitoring filter counting GOOD events (a DELTA metric of int64 or
double type).

### spec.sli.windowsBasedSli.goodTotalRatioThreshold.performance.goodTotalRatio.badServiceFilter

`string`

A monitoring filter counting BAD events.

### spec.sli.windowsBasedSli.goodTotalRatioThreshold.performance.goodTotalRatio.totalServiceFilter

`string`

A monitoring filter counting TOTAL events.

### spec.sli.windowsBasedSli.metricMeanInRange

`GcpMonitoringSloMetricRange`

A window is good when the MEAN of a metric stays inside a range for
the window.

### spec.sli.windowsBasedSli.metricMeanInRange.timeSeries

`string` · required

A monitoring filter selecting the time series whose windowed mean (or
sum) is judged against `range`.

- rule: {"required":true}

### spec.sli.windowsBasedSli.metricMeanInRange.range

`GcpMonitoringSloRange`

The range the windowed value must stay inside for the window to count
as good.

### spec.sli.windowsBasedSli.metricMeanInRange.range.min

`double` · optional (explicit presence)

The lower bound of the range. Unset means unbounded below.

### spec.sli.windowsBasedSli.metricMeanInRange.range.max

`double` · optional (explicit presence)

The upper bound of the range. Unset means unbounded above.

### spec.sli.windowsBasedSli.metricSumInRange

`GcpMonitoringSloMetricRange`

A window is good when the SUM of a metric stays inside a range for
the window.

### spec.sli.windowsBasedSli.metricSumInRange.timeSeries

`string` · required

A monitoring filter selecting the time series whose windowed mean (or
sum) is judged against `range`.

- rule: {"required":true}

### spec.sli.windowsBasedSli.metricSumInRange.range

`GcpMonitoringSloRange`

The range the windowed value must stay inside for the window to count
as good.

### spec.sli.windowsBasedSli.metricSumInRange.range.min

`double` · optional (explicit presence)

The lower bound of the range. Unset means unbounded below.

### spec.sli.windowsBasedSli.metricSumInRange.range.max

`double` · optional (explicit presence)

The upper bound of the range. Unset means unbounded above.

### spec.labels

`map<string, string>`

User labels attached to the SLO for organizing and identifying it
(maps to the provider's user_labels), merged with Planton's platform
labels (which win on key conflicts). Also applied to any Monitoring
service this kind creates. Keys and values may contain only lowercase
letters, numerals, underscores, and dashes; keys must begin with a
letter.

### spec.deletionPolicy

`string`

Deletion policy — what happens when this resource is destroyed:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the SLO (and any service this kind created) is deleted;
               burn-rate alerts referencing it stop evaluating
  "PREVENT" -- destroy FAILS; protects the reliability contract from
               accidental teardown
  "ABANDON" -- the SLO is removed from management but keeps existing
               in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `exactly_one_period`: set exactly one of calendar_period or rolling_period_days

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpMonitoringSlo, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.slo_name` | `string` | The server-assigned resource name of the SLO. Format: projects/{project}/services/{service_id}/serviceLevelObjectives/{slo_id} The handle for burn-rate alert conditions (select_slo_burn_rate("{slo_name}", ...)) and the Monitoring API. |
| `status.outputs.service_name` | `string` | The resource name of the Monitoring service the SLO measures — existing or created by this kind. Format: projects/{project}/services/{service_id} |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |

## See Also

- [Overview](../README.md)
